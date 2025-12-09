package wikisp

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnames/gn"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnuuid"
	"github.com/sfborg/harvester/internal/sources/wikisp/wsparser"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/harvester/pkg/errcode"
	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/sfborg/sflib/pkg/sfga"
	"golang.org/x/sync/errgroup"
)

func (w *wikisp) ToSfga(sfga sfga.Archive) error {
	w.sfga = sfga

	// Insert metadata
	meta := coldp.Meta{
		Title:          "Wikispecies",
		Description:    "A central, extensive database for taxonomy - an open, extensive database for scientists and the public to reflect upon the diversity of life on Earth",
		URL:            "https://species.wikimedia.org/",
		License:        "CC0",
		TaxonomicScope: "All life",
		Keywords:       []string{"taxonomy", "biodiversity", "species", "nomenclature"},
	}
	if err := w.sfga.InsertMeta(&meta); err != nil {
		return fmt.Errorf("failed to insert metadata: %w", err)
	}

	chIn := make(chan string)
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		defer close(chIn)
		err := w.readXML(ctx, chIn)
		return err
	})

	g.Go(func() error {
		err := w.parsePages(ctx, chIn)
		return err
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (w *wikisp) parsePages(_ context.Context, chIn <-chan string) error {
	slog.Info("starting WikiSpecies parsing")

	// PASS 1: Parse and categorize all pages
	for pageStr := range chIn {
		w.parsePage(pageStr)
		w.stats.TotalPages++
	}

	slog.Info("pass 1 complete", "total_pages", w.stats.TotalPages,
		"taxon_pages", w.stats.TaxonPages)
	gn.Info(
		"Pass 1 complete: %d pages, %d taxa",
		w.stats.TotalPages, w.stats.TaxonPages,
	)

	// PASS 2: Add redirects to synonym map
	for from, to := range w.storage.redirects {
		acceptedID := w.storage.taxonIDs[to]
		if acceptedID == "" {
			w.stats.RedirectTargetNotFound++
			w.stats.MissingRedirectTargets[to] = append(
				w.stats.MissingRedirectTargets[to], from)
			continue
		}

		w.addSynonymFromRedirect(from, acceptedID)
	}

	slog.Info("pass 2 complete", "synonyms", len(w.synonymMap))
	gn.Info("Pass 2 complete: %d synonyms", len(w.synonymMap))

	// PASS 3: Create NameUsage entries
	nameUsages, vernaculars := w.createNameUsages()

	// Process synonyms
	for _, syn := range w.synonymMap {
		nu := createSynonymNameUsage(syn, w.gnp)
		nameUsages = append(nameUsages, nu)
		w.stats.SynonymsTotal++
	}

	slog.Info("pass 3 complete", "name_usages", len(nameUsages),
		"vernaculars", len(vernaculars))
	gn.Info(
		"Pass 3 complete: %d name_usages, %d vernaculars",
		len(nameUsages), len(vernaculars),
	)

	// Insert to SFGA
	if err := w.sfga.InsertNameUsages(nameUsages); err != nil {
		return fmt.Errorf("failed to insert name usages: %w", err)
	}

	if len(vernaculars) > 0 {
		if err := w.sfga.InsertVernaculars(vernaculars); err != nil {
			return fmt.Errorf("failed to insert vernacular names: %w", err)
		}
	}

	// Log final statistics
	logStats(w.stats)

	return nil
}

func (w *wikisp) createNameUsages() ([]coldp.NameUsage, []coldp.Vernacular) {
	var nameUsages []coldp.NameUsage
	var vernaculars []coldp.Vernacular

	// Process accepted taxa
	for _, pd := range w.taxonPages {
		nu, ok := w.createNameUsageWithValidation(pd)
		if ok {
			nameUsages = append(nameUsages, nu)
			w.stats.NamesAccepted++
		} else {
			w.stats.NamesRejected++
		}

		// Vernacular names
		for lang, name := range pd.VernacularNames {
			vn := coldp.Vernacular{
				TaxonID:  pd.ID,
				Name:     name,
				Language: lang,
			}
			vernaculars = append(vernaculars, vn)
		}
	}
	return nameUsages, vernaculars
}

func (w *wikisp) readXML(_ context.Context, chIn chan<- string) error {
	elements, err := os.ReadDir(w.cfg.ExtractDir)
	if err != nil {
		return err
	}

	var path string
	for _, v := range elements {
		if strings.HasSuffix(v.Name(), ".xml") {
			path = filepath.Join(w.cfg.ExtractDir, v.Name())
			break
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var inPage bool
	var page []string
	for scanner.Scan() {
		line := scanner.Text()
		if !inPage && pageStart.MatchString(line) {
			inPage = true
		}
		if inPage {
			page = append(page, line)
			if pageEnd.MatchString(line) {
				chIn <- strings.Join(page, "\n")
				page = nil
				inPage = false
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (w *wikisp) parsePage(pageStr string) {
	// Parse XML
	var page PageXML
	if err := xml.Unmarshal([]byte(pageStr), &page); err != nil {
		w.stats.SkippedInvalidXML++
		slog.Warn("invalid XML", "error", err)
		return
	}

	// Categorize page
	switch {
	case isRedirect(&page):
		w.stats.SkippedRedirects++
		if from, to, ok := extractRedirect(&page); ok {
			w.storage.redirects[from] = to
		}
	case isTemplate(&page):
		w.stats.SkippedTemplates++
		// Extract template ID for parent resolution
		templateName := strings.TrimPrefix(page.Title, "Template:")
		w.storage.templateIDs[templateName] = fmt.Sprintf("%d", page.ID)
	default:
		// Try to parse as taxon page
		pd, err := extractPageData(&page, w.wsp)
		if err != nil {
			if !errcode.Is(err, errcode.WikispSkipPage) {
				w.stats.TaxonPagesFailed++
			}
			return
		}

		w.stats.TaxonPages++
		w.storage.taxonIDs[pd.Title] = pd.ID
		w.taxonPages = append(w.taxonPages, pd)

		// Collect synonyms from this page
		for _, synName := range pd.Synonyms {
			w.addSynonymFromSection(synName, pd.ID)
		}
	}
	// Progress logging
	if w.stats.TotalPages%10000 == 0 {
		slog.Info("processing pages", "total", w.stats.TotalPages,
			"taxa", w.stats.TaxonPages)
	}
}

// cleanWikiText removes wiki markup and HTML from text.
func cleanWikiText(text string) string {
	// Remove wiki links: [[Link|Display]] → Display, [[Link]] → Link
	text = wikiLink.ReplaceAllString(text, "$2")

	// Remove bold/italic markup
	text = boldItalic.ReplaceAllString(text, "")

	// Replace templates: {{template|param1|param2}} → param2 (keep last part)
	// This preserves author names from {{a|full|short}} templates
	for {
		before := text
		text = templateLink.ReplaceAllStringFunc(text, func(match string) string {
			// Extract content between {{ and }}
			content := match[2 : len(match)-2]
			// Split by |
			parts := strings.Split(content, "|")
			if len(parts) > 1 {
				// Return last non-empty part
				for i := len(parts) - 1; i >= 0; i-- {
					if strings.TrimSpace(parts[i]) != "" {
						return strings.TrimSpace(parts[i])
					}
				}
			}
			// If no parts or all empty, return empty
			return ""
		})
		// Keep replacing until no more templates
		if text == before {
			break
		}
	}

	// Remove HTML tags
	text = htmlTag.ReplaceAllString(text, "")

	// Remove extinct markers
	text = strings.ReplaceAll(text, "†", "")
	text = strings.ReplaceAll(text, "&dagger;", "")

	// Clean HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")

	// Remove colons followed by numbers/URLs (like ": 101" or ": http://...")
	colonNumRe := regexp.MustCompile(`:\s*\d+.*$`)
	text = colonNumRe.ReplaceAllString(text, "")

	// Normalize whitespace
	text = whitespace.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// parseSections parses wikitext into sections based on ==Header== markers.
// A section is stored in in sections map where the key is section header and
// the value is section header, level and section's lines.
func parseSections(text string) map[string]*Section {
	sections := make(map[string]*Section)
	var currentSection *Section

	for line := range strings.SplitSeq(text, "\n") {
		if matches := sectionHeader.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			header := strings.ToLower(strings.TrimSpace(matches[2]))
			currentSection = &Section{
				Header: header,
				Level:  level,
				Lines:  []string{},
			}
			sections[header] = currentSection
		} else if currentSection != nil && strings.TrimSpace(line) != "" {
			currentSection.Lines = append(currentSection.Lines, line)
		}
	}

	return sections
}

// makeWikiURL creates a WikiSpecies URL for a given title.
func makeWikiURL(title string) string {
	escaped := url.QueryEscape(strings.ReplaceAll(title, " ", "_"))
	return "https://species.wikimedia.org/wiki/" + escaped
}

// isRedirect checks if a page is a redirect.
func isRedirect(page *PageXML) bool {
	return strings.Contains(page.Revision.Text.Content, "#REDIRECT") ||
		strings.Contains(page.Revision.Text.Content, "#redirect") ||
		strings.Contains(page.Revision.Text.Content, "<redirect")
}

// isTemplate checks if a page is a template.
func isTemplate(page *PageXML) bool {
	return strings.HasPrefix(page.Title, "Template:")
}

// isTaxonPage checks if a page is a taxon page by looking for required
// sections.
func isTaxonPage(page *PageXML, sections map[string]*Section) bool {
	// Skip namespace pages
	notTaxon := []string{
		"Type", "Catalog", "WS", "Topic", "Module",
		"Help", "Wikispecies", "MediaWiki", "Translations",
		"Category", "Template",
	}
	for _, v := range notTaxon {
		if strings.HasPrefix(page.Title, v+":") {
			return false
		}
	}

	// Must have both Name and Taxonavigation sections
	_, hasName := sections["{{int:name}}"]
	_, hasTaxonav := sections["{{int:taxonavigation}}"]

	return hasName && hasTaxonav
}

// extractRedirect extracts redirect target from a page.
func extractRedirect(page *PageXML) (from, to string, ok bool) {
	text := page.Revision.Text.Content

	// Try #REDIRECT [[Target]]
	if matches := redirectLink.FindStringSubmatch(text); matches != nil {
		return page.Title, matches[1], true
	}

	// Try <redirect title="Target" />
	if matches := redirectTitle.FindStringSubmatch(text); matches != nil {
		return page.Title, matches[1], true
	}

	return "", "", false
}

// extractVernacularNames extracts vernacular names from VN template.
func extractVernacularNames(section *Section) map[string]string {
	if section == nil {
		return nil
	}

	vnames := make(map[string]string)
	inVN := false

	for _, line := range section.Lines {
		// Start of VN template
		if strings.Contains(line, "{{VN") || strings.Contains(line, "{{vn") {
			inVN = true
			continue
		}

		// End of template
		if inVN && strings.Contains(line, "}}") {
			break
		}

		// Parse language=name pairs
		if inVN && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				lang := strings.TrimSpace(strings.TrimPrefix(parts[0], "|"))
				name := strings.TrimSpace(parts[1])

				// Validate: language code should be short
				if len(lang) > 0 && len(lang) <= 4 && len(name) > 0 {
					vnames[lang] = name
				}
			}
		}
	}

	return vnames
}

// extractSynonyms extracts synonym names from the synonyms section.
func extractSynonyms(sections map[string]*Section) []string {
	section, ok := sections["{{int:synonyms}}"]
	if !ok {
		// Try alternative spellings
		if s, ok := sections["{{int:synonymy}}"]; ok {
			section = s
		} else {
			return nil
		}
	}

	var synonyms []string

	for _, line := range section.Lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and markers
		if line == "" || strings.HasPrefix(line, "{{") {
			continue
		}

		// Remove list markers
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimPrefix(line, "**")
		line = strings.TrimSpace(line)

		// Clean wiki markup
		cleaned := cleanWikiText(line)

		if cleaned != "" {
			synonyms = append(synonyms, cleaned)
		}
	}

	return synonyms
}

// extractScientificName extracts the scientific name from the Name section.
// Uses wsparser with gnparser fallback.
// Returns the wsparser.ParsedName result.
func extractScientificName(
	section *Section,
	wsp *wsparser.WSParser,
) wsparser.ParsedName {
	resEmpty := wsparser.ParsedName{}

	if section == nil || len(section.Lines) == 0 {
		return resEmpty
	}

	// Get first non-empty, non-marker line
	for _, line := range section.Lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "{{") ||
			strings.HasPrefix(line, "===") {
			continue
		}

		// Skip list markers
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		// Parse with wsparser (includes gnparser fallback)
		parsed, err := wsp.Parse(line)
		if err != nil {
			slog.Debug("parse error",
				"line", line, "error", err)
			fmt.Printf("name-0 => %s\n", line)
			return resEmpty
		}

		slog.Debug("parsed scientific name",
			"input", line,
			"canonical", parsed.Canonical,
			"authorship", parsed.Authorship,
			"quality", parsed.Quality)

		return parsed
	}

	return resEmpty
}

// extractParentTemplate extracts parent template from Taxonavigation section.
func extractParentTemplate(section *Section) string {
	if section == nil {
		return ""
	}

	// Look for template links like {{Anthozoa}}
	for _, line := range section.Lines {
		matches := templateLink.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			templateContent := strings.TrimSpace(match[1])

			// Skip known non-parent templates
			if strings.HasPrefix(templateContent, "int:") ||
				strings.HasPrefix(templateContent, "Image") ||
				strings.HasPrefix(templateContent, "DISPLAYTITLE") {
				continue
			}

			// Extract template name (before any |)
			parts := strings.Split(templateContent, "|")
			templateName := strings.TrimSpace(parts[0])

			if templateName != "" {
				return templateName
			}
		}
	}

	return ""
}

// extractPageData parses a taxon page and extracts data.
func extractPageData(
	page *PageXML,
	wsp *wsparser.WSParser,
) (*PageData, error) {
	// Extract BASEPAGENAME from title (remove namespace prefix)
	baseName := page.Title
	if idx := strings.Index(page.Title, ":"); idx != -1 {
		baseName = page.Title[idx+1:]
	}

	// Substitute {{BASEPAGENAME}} in content
	content := strings.ReplaceAll(
		page.Revision.Text.Content,
		"{{BASEPAGENAME}}",
		baseName,
	)

	// Parse into sections
	sections := parseSections(content)

	// Validate it's a taxon page
	if !isTaxonPage(page, sections) {
		return nil, &gn.Error{Code: errcode.WikispSkipPage}
	}

	// Extract data
	pd := &PageData{
		ID:    fmt.Sprintf("%d", page.ID),
		Title: page.Title,
	}

	// Scientific name
	if nameSection, ok := sections["{{int:name}}"]; ok {
		parsed := extractScientificName(nameSection, wsp)
		pd.ScientificName = parsed.Canonical
		pd.Authorship = parsed.Authorship
		pd.ParseQuality = parsed.Quality
	}

	// Parent template
	if taxonavSection, ok := sections["{{int:taxonavigation}}"]; ok {
		pd.ParentTemplate = extractParentTemplate(taxonavSection)
	}

	// Vernacular names
	if vnSection, ok := sections["{{int:vernacular names}}"]; ok {
		pd.VernacularNames = extractVernacularNames(vnSection)
	}

	// Synonyms
	pd.Synonyms = extractSynonyms(sections)

	// Must have at least a scientific name
	if pd.ScientificName == "" {
		return nil, fmt.Errorf("no scientific name found")
	}

	return pd, nil
}

// addSynonymFromSection adds a synonym from a synonym section, using wsparser
// with gnparser fallback, then gnparser for normalization.
func (w *wikisp) addSynonymFromSection(
	synName string,
	acceptedID string,
) {
	// Parse with wsparser (includes gnparser fallback)
	wsParsed, err := w.wsp.Parse(synName)
	if err != nil || wsParsed.Canonical == "" {
		slog.Debug("parse failed for synonym",
			"name", synName, "error", err)
		w.stats.SynonymsParseFailed++
		return
	}

	canonicalName := wsParsed.Canonical
	authorship := wsParsed.Authorship
	quality := wsParsed.Quality

	if existing, ok := w.synonymMap[canonicalName]; ok {
		// Already exists - just mark it
		existing.HasSynonymSection = true
		w.stats.SynonymDuplicates++
	} else {
		// New synonym
		w.synonymMap[canonicalName] = &synonym{
			CanonicalName:     canonicalName,
			Authorship:        authorship,
			Quality:           quality,
			AcceptedID:        acceptedID,
			HasSynonymSection: true,
			HasRedirect:       false,
		}
	}
}

// addSynonymFromRedirect adds a synonym from a redirect, using wsparser with
// gnparser fallback, then gnparser for normalization.
func (w *wikisp) addSynonymFromRedirect(
	from string,
	acceptedID string,
) {
	// Parse with wsparser (includes gnparser fallback)
	wsParsed, err := w.wsp.Parse(from)
	if err != nil || wsParsed.Canonical == "" {
		slog.Debug("parse failed for redirect synonym",
			"name", from, "error", err)
		return
	}

	canonicalName := wsParsed.Canonical
	authorship := wsParsed.Authorship
	quality := wsParsed.Quality

	if existing, ok := w.synonymMap[canonicalName]; ok {
		// Already exists from synonym section - merge!
		existing.HasRedirect = true

		// Redirect is authoritative for acceptedID
		if existing.AcceptedID != acceptedID {
			slog.Warn("synonym accepted ID conflict, using redirect",
				"synonym", canonicalName,
				"section_id", existing.AcceptedID,
				"redirect_id", acceptedID)
		}
		existing.AcceptedID = acceptedID

		w.stats.SynonymDuplicates++
	} else {
		// New synonym from redirect only
		w.synonymMap[canonicalName] = &synonym{
			CanonicalName: canonicalName,
			Authorship:    authorship,
			Quality:       quality,
			AcceptedID:    acceptedID,
			HasRedirect:   true,
		}
	}
}

// createNameUsageWithValidation creates a NameUsage from PageData with
// validation.
func (w *wikisp) createNameUsageWithValidation(
	pd *PageData,
) (coldp.NameUsage, bool) {

	// Build full scientific name string (canonical + authorship)
	scientificNameString := pd.ScientificName
	if pd.Authorship != "" {
		scientificNameString = pd.ScientificName + " " + pd.Authorship
	}

	nu := coldp.NameUsage{
		ID:                   pd.ID,
		ScientificName:       pd.ScientificName,    // Canonical name
		ScientificNameString: scientificNameString, // Full name with authorship
		CanonicalFull:        pd.ScientificName,    // Just canonical
		Authorship:           pd.Authorship,        // Just authorship
		Code:                 nomcode.Unknown,
		TaxonomicStatus:      coldp.AcceptedTS,
		Link:                 makeWikiURL(pd.Title),
	}

	data.AddParsedData(w.gnp, &nu)

	// Warn about low quality parse
	if nu.ParseQuality.Valid && nu.ParseQuality.Int64 > 2 {
		slog.Warn("low quality parse",
			"name", pd.ScientificName,
			"quality", nu.ParseQuality.Int64,
			"title", pd.Title)
	}

	// Resolve parent
	if pd.ParentTemplate != "" {
		if parentID, found := resolveParentID(pd.ParentTemplate, w.storage); found {
			nu.ParentID = parentID
			w.stats.ParentResolved++
		} else {
			w.stats.ParentNotFound++
			w.stats.MissingParents[pd.ParentTemplate] = append(
				w.stats.MissingParents[pd.ParentTemplate], pd.Title)
			slog.Debug("parent template not found",
				"template", pd.ParentTemplate,
				"taxon", pd.Title)
		}
	}

	// Validate minimum requirements
	if nu.ID == "" || nu.ScientificNameString == "" {
		return nu, false
	}

	return nu, true
}

// resolveParentID attempts to resolve a parent template to a taxon ID.
func resolveParentID(
	parentTemplate string,
	storage *tempStorage,
) (string, bool) {
	if parentTemplate == "" {
		return "", false
	}

	// Try taxon IDs first (preferred - these are actual taxon pages)
	if id, ok := storage.taxonIDs[parentTemplate]; ok {
		return id, true
	}

	// Try cleaning the template name
	cleaned := cleanWikiText(parentTemplate)
	if id, ok := storage.taxonIDs[cleaned]; ok {
		return id, true
	}

	// Fall back to template IDs only if no taxon page exists
	if id, ok := storage.templateIDs[parentTemplate]; ok {
		return id, true
	}

	return "", false
}

// createSynonymNameUsage creates a NameUsage for a synonym.
func createSynonymNameUsage(syn *synonym, gnp gnparser.GNparser) coldp.NameUsage {
	// Extract canonical from syn.Name (format: "syn-X => canonical")
	canonical := syn.CanonicalName

	// Build full scientific name string (canonical + authorship)
	scientificNameString := canonical
	if syn.Authorship != "" {
		scientificNameString = canonical + " " + syn.Authorship
	}

	nu := coldp.NameUsage{
		ID:                   gnuuid.New(syn.CanonicalName).String(),
		ScientificName:       canonical,            // Canonical name
		ScientificNameString: scientificNameString, // Full name with authorship
		CanonicalFull:        canonical,            // Just canonical
		Authorship:           syn.Authorship,       // Just authorship
		ParentID:             syn.AcceptedID,
		TaxonomicStatus:      coldp.SynonymTS,
		Code:                 nomcode.Unknown,
	}

	data.AddParsedData(gnp, &nu)

	return nu
}

// logStats logs final parsing statistics.
func logStats(stats *parseStats) {
	slog.Info("WikiSpecies parsing complete",
		"total_pages", stats.TotalPages,
		"taxa_processed", stats.TaxonPages,
		"taxa_failed", stats.TaxonPagesFailed,
		"names_accepted", stats.NamesAccepted,
		"names_rejected", stats.NamesRejected,
		"synonyms", stats.SynonymsTotal,
		"synonym_duplicates", stats.SynonymDuplicates,
		"parents_resolved", stats.ParentResolved,
		"parents_not_found", stats.ParentNotFound,
		"redirects", stats.SkippedRedirects,
		"redirect_target_not_found", stats.RedirectTargetNotFound,
	)
	gn.Info("WikiSpecies parsing complete")

	// Log missing parent templates
	if len(stats.MissingParents) > 0 {
		slog.Warn("missing parent templates", "count", len(stats.MissingParents))
		gn.Info("Missing %d parent templates", len(stats.MissingParents))
		for template, taxa := range stats.MissingParents {
			slog.Info("missing parent template",
				"template", template,
				"needed_by", taxa)
		}
	}

	// Log missing redirect targets
	if len(stats.MissingRedirectTargets) > 0 {
		slog.Warn(
			"missing redirect targets", "count", len(stats.MissingRedirectTargets),
		)
		gn.Info("Missing %d redirect targets", len(stats.MissingRedirectTargets))
		for target, redirects := range stats.MissingRedirectTargets {
			slog.Info("missing redirect target",
				"target", target,
				"redirects_from", redirects)
		}
	}

	// Warn if high failure rate
	if stats.TaxonPages > 0 {
		failureRate := float64(stats.TaxonPagesFailed) /
			float64(stats.TaxonPages+stats.TaxonPagesFailed)
		if failureRate > 0.1 {
			slog.Warn("high taxon parsing failure rate",
				"rate", fmt.Sprintf("%.1f%%", failureRate*100))
		}
	}
}
