package nzor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/gnfmt/gnlang"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

// JSON structures for the NZOR API response.

type nzorResponse struct {
	Names []nzorName `json:"names"`
}

type nzorName struct {
	NameID         string          `json:"nameId"`
	FullName       string          `json:"fullName"`
	Rank           string          `json:"rank"`
	Status         string          `json:"status"`
	Class          string          `json:"class"`
	GovCode        string          `json:"governingCode"`
	AcceptedName   *nzorRef        `json:"acceptedName"`
	Concepts       []nzorConcept   `json:"concepts"`
	ClassHierarchy []nzorHierEntry `json:"classificationHierarchy"`
	Language       string          `json:"language"`
}

type nzorRef struct {
	NameID string `json:"nameId"`
}

type nzorConcept struct {
	Applications []nzorApplication `json:"applications"`
}

type nzorApplication struct {
	Type    string          `json:"type"`
	Concept *nzorConceptRef `json:"concept"`
}

type nzorConceptRef struct {
	Name *nzorRef `json:"name"`
}

type nzorHierEntry struct {
	Rank        string `json:"rank"`
	PartialName string `json:"partialName"`
}

const nzorLinkBase = "https://www.nzor.org.nz/names/"

func (n *nzor) importNameUsages() error {
	gnp := gnparser.New(gnparser.NewConfig(
		gnparser.OptWithDetails(true),
	))

	f, err := os.Open(n.jsonlPath)
	if err != nil {
		return fmt.Errorf("opening nzor.jsonl: %w", err)
	}
	defer f.Close()

	var (
		totalNU   int
		totalVern int
		nuBatch   []coldp.NameUsage
		vernBatch []coldp.Vernacular
	)

	scanner := bufio.NewScanner(f)
	// 10 MB buffer — NZOR pages can be large.
	scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var resp nzorResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue // skip malformed lines
		}

		for i := range resp.Names {
			nm := &resp.Names[i]
			switch nm.Class {
			case "Scientific Name":
				nu := buildNameUsage(nm)
				data.AddParsedData(gnp, nu)
				nuBatch = append(nuBatch, *nu)
				totalNU++
			case "Vernacular Name":
				if v := buildVernacular(nm); v != nil {
					vernBatch = append(vernBatch, *v)
					totalVern++
				}
			}
		}

		if len(nuBatch) >= n.cfg.BatchSize {
			if err := n.flushBatch(nuBatch, vernBatch, totalNU, totalVern); err != nil {
				return err
			}
			nuBatch = nuBatch[:0]
			vernBatch = vernBatch[:0]
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading nzor.jsonl: %w", err)
	}

	if len(nuBatch) > 0 || len(vernBatch) > 0 {
		if err := n.flushBatch(nuBatch, vernBatch, totalNU, totalVern); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
	return nil
}

func (n *nzor) flushBatch(
	nuBatch []coldp.NameUsage,
	vernBatch []coldp.Vernacular,
	totalNU, totalVern int,
) error {
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s names, %s vernaculars",
		humanize.Comma(int64(totalNU)),
		humanize.Comma(int64(totalVern)),
	)
	if len(nuBatch) > 0 {
		if err := n.sfga.InsertNameUsages(nuBatch); err != nil {
			return err
		}
	}
	if len(vernBatch) > 0 {
		if err := n.sfga.InsertVernaculars(vernBatch); err != nil {
			return err
		}
	}
	return nil
}

func buildNameUsage(nm *nzorName) *coldp.NameUsage {
	taxStatus := coldp.AcceptedTS
	parentID := ""
	if nm.AcceptedName != nil && nm.AcceptedName.NameID != nm.NameID {
		taxStatus = coldp.SynonymTS
		parentID = nm.AcceptedName.NameID
	}

	cl := parseClassification(nm.ClassHierarchy)

	return &coldp.NameUsage{
		ID:                   nm.NameID,
		ScientificName:       nm.FullName,
		ScientificNameString: nm.FullName,
		Rank:                 coldp.NewRank(nm.Rank),
		TaxonomicStatus:      taxStatus,
		ParentID:             parentID,
		Code:                 parseCode(nm.GovCode),
		Link:                 nzorLinkBase + nm.NameID,
		NameRemarks:          nm.Status,
		Kingdom:              cl["kingdom"],
		Phylum:               cl["phylum"],
		Class:                cl["class"],
		Order:                cl["order"],
		Family:               cl["family"],
		Genus:                cl["genus"],
	}
}

func buildVernacular(nm *nzorName) *coldp.Vernacular {
	taxonID := vernTaxonID(nm)
	if taxonID == "" {
		return nil
	}
	return &coldp.Vernacular{
		TaxonID:  taxonID,
		Name:     nm.FullName,
		Language: gnlang.LangCode(nm.Language),
	}
}

// vernTaxonID finds the scientific name ID that this vernacular name refers to
// by looking for an application of type "is vernacular for".
func vernTaxonID(nm *nzorName) string {
	for _, c := range nm.Concepts {
		for _, a := range c.Applications {
			if a.Type == "is vernacular for" && a.Concept != nil && a.Concept.Name != nil {
				return a.Concept.Name.NameID
			}
		}
	}
	return ""
}

func parseClassification(entries []nzorHierEntry) map[string]string {
	cl := make(map[string]string)
	for _, e := range entries {
		switch strings.ToLower(e.Rank) {
		case "kingdom":
			cl["kingdom"] = e.PartialName
		case "phylum":
			cl["phylum"] = e.PartialName
		case "class":
			cl["class"] = e.PartialName
		case "order":
			cl["order"] = e.PartialName
		case "family":
			cl["family"] = e.PartialName
		case "genus":
			cl["genus"] = e.PartialName
		}
	}
	return cl
}

func parseCode(govCode string) nomcode.Code {
	switch strings.ToUpper(govCode) {
	case "ICN", "ICBN":
		return nomcode.Botanical
	case "ICZN":
		return nomcode.Zoological
	case "ICNB", "ICNP":
		return nomcode.Bacterial
	case "ICTV":
		return nomcode.Virus
	default:
		return nomcode.Unknown
	}
}
