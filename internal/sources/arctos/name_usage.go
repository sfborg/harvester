package arctos

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnuuid"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

type synRec struct {
	relatedName  string
	relationship string
}

func (a *arctos) importNameUsages() error {
	syns, err := a.loadSynonyms()
	if err != nil {
		return err
	}

	names, err := a.pivotClassification()
	if err != nil {
		return err
	}

	gnp := gnparser.New(gnparser.NewConfig(
		gnparser.OptWithDetails(true),
	))

	var idCounter int
	idMap := make(map[string]string)
	makeID := func(name string) string {
		uuid := gnuuid.New(name).String()
		if id, ok := idMap[uuid]; ok {
			return id
		}
		idCounter++
		id := fmt.Sprintf("sf_%d", idCounter)
		idMap[uuid] = id
		return id
	}

	var batch []coldp.NameUsage
	total := 0

	for sciName, fields := range names {
		nu := buildNameUsage(sciName, fields)
		nu.ID = makeID(sciName)

		data.AddParsedData(gnp, nu)

		if nu.CanonicalFull != "" {
			nu.NameAlternativeID = "gnoutlink:" + url.QueryEscape(nu.CanonicalFull)
		} else {
			nu.NameAlternativeID = "gnoutlink:" + url.QueryEscape(nu.ScientificName)
		}

		batch = append(batch, *nu)
		total++

		if len(batch) >= a.cfg.BatchSize {
			if err := a.flushBatch(batch, total); err != nil {
				return err
			}
			batch = batch[:0]
		}

		if synList, ok := syns[sciName]; ok {
			for _, s := range synList {
				snu := buildSynonym(s, nu.ID)
				snu.ID = makeID(s.relatedName)
				data.AddParsedData(gnp, snu)
				if snu.CanonicalFull != "" {
					snu.NameAlternativeID = "gnoutlink:" + url.QueryEscape(snu.CanonicalFull)
				} else {
					snu.NameAlternativeID = "gnoutlink:" + url.QueryEscape(snu.ScientificName)
				}
				batch = append(batch, *snu)
				total++
				if len(batch) >= a.cfg.BatchSize {
					if err := a.flushBatch(batch, total); err != nil {
						return err
					}
					batch = batch[:0]
				}
			}
		}
	}

	if len(batch) > 0 {
		if err := a.flushBatch(batch, total); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
	return nil
}

func (a *arctos) flushBatch(batch []coldp.NameUsage, total int) error {
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s names", humanize.Comma(int64(total)))
	return a.sfga.InsertNameUsages(batch)
}

// loadSynonyms reads globalnames_relationships.csv and returns a map from
// accepted scientific_name to its synonyms.
func (a *arctos) loadSynonyms() (map[string][]synRec, error) {
	path := filepath.Join(a.cfg.ExtractDir, "globalnames_relationships.csv")

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := make(map[string]int)
	for i, h := range header {
		idx[h] = i
	}

	syns := make(map[string][]synRec)
	var count int

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		sciName := getField(row, idx, "scientific_name")
		related := getField(row, idx, "related_name")
		rel := getField(row, idx, "taxon_relationship")
		if sciName == "" || related == "" {
			continue
		}
		syns[sciName] = append(syns[sciName], synRec{
			relatedName:  related,
			relationship: rel,
		})

		count++
		if count%100_000 == 0 {
			fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 80))
			fmt.Fprintf(os.Stderr, "\rLoading synonyms: %s rows",
				humanize.Comma(int64(count)))
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
	return syns, nil
}

// pivotClassification reads globalnames_classification.csv and pivots the
// Entity-Attribute-Value rows into a map[scientificName]map[termType]term,
// keeping only name_type == "Linnean".
func (a *arctos) pivotClassification() (map[string]map[string]string, error) {
	path := filepath.Join(a.cfg.ExtractDir, "globalnames_classification.csv")

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	// read header
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := make(map[string]int)
	for i, h := range header {
		idx[h] = i
	}

	names := make(map[string]map[string]string)
	var count int

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if getField(row, idx, "name_type") != "Linnean" {
			continue
		}

		sciName := getField(row, idx, "scientific_name")
		termType := getField(row, idx, "term_type")
		term := getField(row, idx, "term")

		if sciName == "" || termType == "" {
			continue
		}

		if _, ok := names[sciName]; !ok {
			names[sciName] = make(map[string]string)
		}
		// keep the first value for each term_type
		if _, exists := names[sciName][termType]; !exists {
			names[sciName][termType] = term
		}

		count++
		if count%500_000 == 0 {
			fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 80))
			fmt.Fprintf(os.Stderr, "\rPivoting classification: %s rows",
				humanize.Comma(int64(count)))
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
	return names, nil
}

func buildNameUsage(sciName string, f map[string]string) *coldp.NameUsage {
	authors := strings.TrimSpace(f["author_text"])
	sciNameStr := strings.TrimSpace(sciName + " " + authors)

	nu := &coldp.NameUsage{
		ScientificName:       sciName,
		ScientificNameString: sciNameStr,
		Authorship:           authors,
		TaxonomicStatus:      parseTaxStatus(f["taxon_status"]),
		Code:                 parseNomCode(f["nomenclatural_code"]),
		Kingdom:              f["kingdom"],
		Phylum:               f["phylum"],
		Class:                f["class"],
		Subclass:             f["subclass"],
		Order:                f["order"],
		Suborder:             f["suborder"],
		Superfamily:          f["superfamily"],
		Family:               f["family"],
		Subfamily:            f["subfamily"],
		Tribe:                f["tribe"],
		Subtribe:             f["subtribe"],
		Genus:                f["genus"],
		Species:              f["species"],
	}

	if f["taxon_status"] == "extinct" {
		nu.Extinct = coldp.ToBool("true")
	}

	return nu
}

func buildSynonym(s synRec, parentID string) *coldp.NameUsage {
	return &coldp.NameUsage{
		ScientificName:       s.relatedName,
		ScientificNameString: s.relatedName,
		ParentID:             parentID,
		TaxonomicStatus:      coldp.SynonymTS,
		NameRemarks:          s.relationship,
	}
}

func parseTaxStatus(status string) coldp.TaxonomicStatus {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "valid", "accepted", "extant":
		return coldp.AcceptedTS
	case "extinct":
		return coldp.AcceptedTS
	case "nomen dubium", "nomen nudum", "nomen oblitum", "doubtful":
		return coldp.ProvisionallyAcceptedTS
	default:
		return coldp.UnknownTaxSt
	}
}

func parseNomCode(code string) nomcode.Code {
	switch strings.TrimSpace(code) {
	case "ICZN", "ICZ", "ICZA":
		return nomcode.Zoological
	case "ICBN", "ICN", "IFPNI":
		return nomcode.Botanical
	case "ICNP", "ICNB":
		return nomcode.Bacterial
	case "ICTV":
		return nomcode.Virus
	default:
		return nomcode.Unknown
	}
}

func getField(row []string, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

