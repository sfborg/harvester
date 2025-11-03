package worldplants

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/google/uuid"
	"github.com/sfborg/sflib/pkg/coldp"
)

// synonymContext holds context for processing synonyms of an accepted name.
type synonymContext struct {
	acceptedNode   hNode
	acceptedNameID string
	namespace      uuid.UUID
	parser         *worldplants
}

// basionymLookup maps basionym keys to their NameUsage records.
type basionymLookup map[string]coldp.NameUsage

// processSynonyms processes the synonym list for an accepted name.
// Reference: original lines 998-1135
func (wp *worldplants) processSynonyms(
	node hNode,
	acceptedNameID string,
	referenceLookup map[string]citation,
) ([]*coldp.NameUsage, basionymLookup, error) {
	if node.verbatimSynonyms == "" {
		return nil, nil, nil
	}

	ctx := synonymContext{
		acceptedNode:   node,
		acceptedNameID: acceptedNameID,
		namespace:      wp.namespace,
		parser:         wp,
	}

	synonymList := parseSynonymList(node.verbatimSynonyms)

	var synonymUsages []*coldp.NameUsage
	basionyms := make(basionymLookup)
	blacklist := make(map[string]struct{})
	seenSynonyms := make(map[string]struct{})

	for _, synString := range synonymList {
		// Skip problematic synonyms
		if shouldSkipSynonym(synString) {
			slog.Debug("Skipping synonym", "reason", "obsolete rank or empty")
			continue
		}

		synName, synRef := parseSynonymString(synString)

		// Fix hybrid notation
		synName = fixHybridNotation(synName)

		// Parse synonym
		synParsed, err := wfwpParse(wp.parser, synName, synRef)
		if err != nil {
			slog.Warn("Failed to parse synonym", "name", synName, "error", err)
			continue
		}

		// Validate parsed synonym
		if !isValidSynonym(synParsed, synName) {
			continue
		}

		// Create synonym NameUsage
		synUsage, err := ctx.createSynonymUsage(
			synParsed,
			synRef,
			referenceLookup,
		)
		if err != nil {
			slog.Warn("Failed to create synonym", "name", synName, "error", err)
			continue
		}

		// Check for duplicates
		if _, exists := seenSynonyms[synUsage.ID]; exists {
			slog.Debug("Duplicate synonym excluded", "id", synUsage.ID)
			continue
		}
		seenSynonyms[synUsage.ID] = struct{}{}

		synonymUsages = append(synonymUsages, synUsage)

		// Update basionym lookup
		updateBasionymLookup(synUsage, synParsed, basionyms, blacklist)
	}

	return synonymUsages, basionyms, nil
}

// parseSynonymList splits the synonym string by = delimiter.
func parseSynonymList(synonyms string) []string {
	return strings.Split(synonyms, "=")
}

// shouldSkipSynonym checks if a synonym should be skipped.
func shouldSkipSynonym(synString string) bool {
	if synString == "" {
		return true
	}

	// Skip obsolete ranks that CLB doesn't accept
	obsoleteRanks := []string{
		"¿",
		" unranked ",
		" proles ",
		" nothossp. ",
		" convar. ",
		" agamosp. ",
		" race ",
		" nothovar. ",
		" nvar. ",
		" nothof. ",
		" microgen. ",
	}

	for _, rank := range obsoleteRanks {
		if strings.Contains(synString, rank) {
			return true
		}
	}

	return false
}

// parseSynonymString extracts the name and reference from a synonym string.
func parseSynonymString(synString string) (name, reference string) {
	name = strings.TrimSpace(synString)

	// Check if reference is embedded in brackets
	if strings.Contains(synString, "[") {
		parts := strings.Split(synString, "[")
		name = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			reference = strings.TrimSpace(strings.ReplaceAll(parts[1], "]", ""))
		}
	}

	return name, reference
}

// fixHybridNotation fixes hybrid names with no space (xGenus → × Genus).
func fixHybridNotation(name string) string {
	hybridRegex := regexp.MustCompile(`^x([A-Z][a-z]+.*)`)
	fixed := hybridRegex.ReplaceAllString(name, "× $1")

	if fixed != name {
		slog.Debug("Fixed hybrid notation", "original", name, "fixed", fixed)
	}

	return fixed
}

// isValidSynonym checks if a parsed synonym is valid.
func isValidSynonym(parsed gnparsed, originalName string) bool {
	// Exclude quadrinomials
	if parsed.cardinality > 3 {
		slog.Debug("Skipping quadrinomial synonym", "name", originalName)
		return false
	}

	// Exclude low quality parses
	if parsed.quality == 0 || parsed.quality > 2 {
		slog.Debug(
			"Skipping low quality synonym",
			"quality", parsed.quality,
			"name", originalName,
		)
		return false
	}

	return true
}

// createSynonymUsage creates a coldp.NameUsage for a synonym.
func (ctx *synonymContext) createSynonymUsage(
	parsed gnparsed,
	reference string,
	referenceLookup map[string]citation,
) (*coldp.NameUsage, error) {
	// Generate synonym ID
	idString := ctx.acceptedNameID + "_" +
		parsed.canonicalFull + "_" +
		strings.ReplaceAll(parsed.authorship, " ", "-")

	synonymID := uuid.NewSHA1(ctx.namespace, []byte(idString)).String()

	// Get reference information
	refID, page, year := addReference(
		referenceLookup,
		reference,
		parsed,
		ctx.namespace,
	)

	// Determine synonym rank
	synRank := synonymRank(parsed, ctx.acceptedNode.rank)

	usage := &coldp.NameUsage{
		ID:                   synonymID,
		ParentID:             ctx.acceptedNameID,
		ScientificName:       parsed.canonicalFull,
		Rank:                 synRank,
		Uninomial:            parsed.uninomial,
		GenericName:          parsed.genus,
		InfragenericEpithet:  parsed.subgenus,
		SpecificEpithet:      parsed.species,
		InfraspecificEpithet: parsed.infraspecies,
		Notho:                parsed.notho,
		TaxonomicStatus:      coldp.SynonymTS,
		Authorship:           parsed.authorship,
		ReferenceID:          refID,
		NameReferenceID:      refID,
		PublishedInPage:      page,
		PublishedInYear:      cleanYear(year),
		Code:                 nomcode.Botanical,
		Remarks:              ctx.acceptedNode.parsed.remarks,
		ScientificNameString: parsed.canonicalFull + " " + parsed.authorship,
	}

	return usage, nil
}

// updateBasionymLookup updates the basionym lookup map.
func updateBasionymLookup(
	usage *coldp.NameUsage,
	parsed gnparsed,
	lookup basionymLookup,
	blacklist map[string]struct{},
) {
	// Only original combinations can be basionyms
	if parsed.combinationAuthorship != "" {
		return
	}

	basionymID := getBasionymId(parsed)

	// Check if already blacklisted
	if _, blacklisted := blacklist[basionymID]; blacklisted {
		slog.Debug("Not adding blacklisted basionym", "id", basionymID)
		return
	}

	// Check for duplicates (ambiguous basionyms)
	if existing, exists := lookup[basionymID]; exists {
		slog.Warn(
			"Ambiguous basionym found",
			"id", basionymID,
			"existing", existing.ScientificName,
			"new", usage.ScientificName,
		)
		// Move to blacklist and remove from lookup
		blacklist[basionymID] = struct{}{}
		delete(lookup, basionymID)
		return
	}

	// Add to lookup
	lookup[basionymID] = *usage
}

// linkBasionyms links basionyms to their combinations.
func linkBasionyms(
	nameUsages []*coldp.NameUsage,
	lookup basionymLookup,
	parser *worldplants,
) error {
	for _, usage := range nameUsages {
		// Only link species-group ranks
		if usage.Rank != coldp.Species &&
			usage.Rank != coldp.Subspecies &&
			usage.Rank != coldp.Variety &&
			usage.Rank != coldp.Form {
			continue
		}

		// Parse the name to get basionym key
		fullName := usage.ScientificName + " " + usage.Authorship
		parsed, err := wfwpParse(parser.parser, fullName, "")
		if err != nil {
			slog.Debug("Failed to parse name for basionym linking", "name", fullName)
			continue
		}

		basionymID := getBasionymId(parsed)

		// Look up basionym
		if basionym, exists := lookup[basionymID]; exists {
			// Don't link to itself
			if usage.ID != basionym.ID {
				usage.BasionymID = basionym.ID
			}
		}
	}

	return nil
}

// addReference adds a reference to the lookup and returns reference info.
// Reference: original lines 541-583
func addReference(
	lookup map[string]citation,
	verbatimCitation string,
	parsed gnparsed,
	namespace uuid.UUID,
) (string, string, string) {
	var publishedInYear, publishedInPage string

	if strings.Contains(verbatimCitation, ":") {
		pageYearMatch := strings.Split(verbatimCitation, ":")[1]
		if strings.Contains(pageYearMatch, "(") {
			publishedInPage = strings.TrimSpace(
				strings.Split(pageYearMatch, "(")[0],
			)
			publishedInYear = strings.TrimSpace(
				strings.ReplaceAll(
					strings.Split(pageYearMatch, "(")[1],
					")",
					"",
				),
			)
		} else {
			publishedInPage = strings.TrimSpace(pageYearMatch)
		}
	}

	refAuthorship := parsed.authorship
	if parsed.combinationAuthorship != "" {
		refAuthorship = parsed.combinationAuthorship
	}

	var fullCitation string
	if refAuthorship != "" && publishedInYear != "" {
		fullCitation = refAuthorship + ". (" + publishedInYear + "). " +
			verbatimCitation
	} else if refAuthorship != "" && publishedInYear == "" {
		fullCitation = refAuthorship + ". " + verbatimCitation
	} else if refAuthorship == "" && publishedInYear != "" {
		fullCitation = "(" + publishedInYear + "). " + verbatimCitation
	}

	fullCitation = strings.ReplaceAll(fullCitation, "..", ".")

	referenceID := uuid.NewSHA1(namespace, []byte(fullCitation)).String()

	if _, ok := lookup[referenceID]; !ok {
		lookup[referenceID] = citation{
			id:       referenceID,
			author:   refAuthorship,
			year:     publishedInYear,
			title:    verbatimCitation,
			citation: fullCitation,
		}
	}

	return referenceID, publishedInPage, publishedInYear
}

// cleanYear validates and cleans a publication year.
// Reference: original lines 705-723
func cleanYear(year string) string {
	// Only accept 4 digit years from 1700-2099
	yearRegex := regexp.MustCompile(`^(17|18|19|20)\d{2}$`)
	year = yearRegex.FindString(year)

	// Validate year is not in the future
	// Note: In production, should check against current year
	// For now, just return the validated year
	return year
}
