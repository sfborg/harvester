package wcvp

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

const powoLinkBase = "https://powo.science.kew.org/taxon/urn:lsid:ipni.org:names:"

var yearRe = regexp.MustCompile(`\d{4}`)

func (w *wcvp) importNameUsages() error {
	gnp := gnparser.New(gnparser.NewConfig(
		gnparser.OptCode(nomcode.Botanical),
		gnparser.OptWithDetails(true),
	))

	f, err := os.Open(w.csvPath)
	if err != nil {
		return fmt.Errorf("opening WCVP csv for name usages: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '|'
	r.LazyQuotes = true

	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("reading WCVP csv headers: %w", err)
	}
	idx := buildIndex(headers)

	var total int
	var batch []coldp.NameUsage

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading WCVP csv: %w", err)
		}

		nu := buildNameUsage(row, idx, w.refMap)
		if nu == nil {
			continue
		}
		data.AddParsedData(gnp, nu)

		total++
		batch = append(batch, *nu)
		if len(batch) >= w.cfg.BatchSize {
			if err := w.flushBatch(batch, total); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := w.flushBatch(batch, total); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
	return nil
}

func (w *wcvp) flushBatch(batch []coldp.NameUsage, total int) error {
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s names", humanize.Comma(int64(total)))
	return w.sfga.InsertNameUsages(batch)
}

func buildNameUsage(row []string, idx map[string]int, refMap map[string]string) *coldp.NameUsage {
	get := getter(row, idx)

	id := get("plant_name_id")
	if id == "" {
		return nil
	}
	name := get("taxon_name")
	if name == "" {
		return nil
	}

	authors := get("taxon_authors")
	nameStr := strings.TrimSpace(name + " " + authors)

	ipniID := get("ipni_id")
	powoID := get("powo_id")

	taxStatus := mapTaxStatus(get("taxon_status"))
	acceptedID := get("accepted_plant_name_id")
	parentID := resolveParentID(taxStatus, id, acceptedID, get("parent_plant_name_id"))

	var refID string
	if pub := get("place_of_publication"); pub != "" && refMap != nil {
		key := pub + "|" + get("volume_and_page") + "|" + get("first_published")
		refID = refMap[key]
	}

	return &coldp.NameUsage{
		ID:                   id,
		NameAlternativeID:    buildAltID(ipniID, powoID),
		ScientificName:       name,
		Authorship:           authors,
		ScientificNameString: nameStr,
		Rank:                 coldp.NewRank(strings.ToUpper(get("taxon_rank"))),
		TaxonomicStatus:      taxStatus,
		ParentID:             parentID,
		BasionymID:           get("basionym_plant_name_id"),
		Family:               get("family"),
		Genus:                get("genus"),
		SpecificEpithet:      get("species"),
		InfraspecificEpithet: get("infraspecies"),
		NameRemarks:          get("nomenclatural_remarks"),
		PublishedInYear:      extractYear(get("first_published")),
		NameReferenceID:      refID,
		Link:                 powoLink(powoID),
		Code:                 nomcode.Botanical,
	}
}

func mapTaxStatus(s string) coldp.TaxonomicStatus {
	switch strings.ToLower(s) {
	case "accepted", "artificial hybrid":
		return coldp.AcceptedTS
	case "provisionally accepted":
		return coldp.ProvisionallyAcceptedTS
	case "synonym", "orthographic", "local biotype":
		return coldp.SynonymTS
	case "misapplied":
		return coldp.MisappliedTS
	case "illegitimate", "invalid":
		return coldp.BareNameTS
	default:
		return coldp.UnknownTaxSt
	}
}

func resolveParentID(ts coldp.TaxonomicStatus, id, acceptedID, parentID string) string {
	switch ts {
	case coldp.AcceptedTS, coldp.ProvisionallyAcceptedTS:
		return parentID
	case coldp.SynonymTS, coldp.MisappliedTS:
		if acceptedID != id {
			return acceptedID
		}
	}
	return ""
}

func buildAltID(ipniID, powoID string) string {
	var parts []string
	if ipniID != "" {
		parts = append(parts, "ipni:"+ipniID)
	}
	if powoID != "" {
		parts = append(parts, "powo:"+powoID)
		parts = append(parts, "gnoutlink:"+powoID)
	}
	return strings.Join(parts, ",")
}

func powoLink(powoID string) string {
	if powoID == "" {
		return ""
	}
	return powoLinkBase + powoID
}

func buildIndex(headers []string) map[string]int {
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[strings.TrimSpace(h)] = i
	}
	return idx
}

func getter(row []string, idx map[string]int) func(string) string {
	return func(col string) string {
		i, ok := idx[col]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
}

func extractYear(s string) string {
	return yearRe.FindString(s)
}
