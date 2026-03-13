package ipni

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

const ipniLinkBase = "https://www.ipni.org/n/"

func (i *ipni) importNameUsages() error {
	gnp := gnparser.New(gnparser.NewConfig(
		gnparser.OptCode(nomcode.Botanical),
		gnparser.OptWithDetails(true),
	))

	f, err := os.Open(i.csvPath)
	if err != nil {
		return fmt.Errorf("opening IPNI csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '|'
	r.LazyQuotes = true

	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("reading IPNI csv headers: %w", err)
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
			return fmt.Errorf("reading IPNI csv: %w", err)
		}

		nu := buildNameUsage(row, idx)
		if nu == nil {
			continue
		}
		data.AddParsedData(gnp, nu)

		total++
		batch = append(batch, *nu)
		if len(batch) >= i.cfg.BatchSize {
			if err := i.flushBatch(batch, total); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := i.flushBatch(batch, total); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "\n")
	return nil
}

func (i *ipni) flushBatch(batch []coldp.NameUsage, total int) error {
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s names", humanize.Comma(int64(total)))
	return i.sfga.InsertNameUsages(batch)
}

func buildNameUsage(row []string, idx map[string]int) *coldp.NameUsage {
	get := func(col string) string {
		if i, ok := idx[col]; ok && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}

	lsid := get("id")
	if lsid == "" {
		return nil
	}

	id := lsid[strings.LastIndex(lsid, ":")+1:]
	name := get("taxon_scientific_name_s_lower")
	if name == "" {
		return nil
	}

	authors := get("authors_t")
	nameStr := strings.TrimSpace(name + " " + authors)

	return &coldp.NameUsage{
		ID:                   id,
		NameAlternativeID:    "lsid:" + lsid,
		ScientificName:       name,
		ScientificNameString: nameStr,
		Authorship:           authors,
		Rank:                 coldp.NewRank(get("rank_s_alphanum")),
		Code:                 nomcode.Botanical,
		Link:                 ipniLinkBase + id,
		Family:               get("family_s_lower"),
		Genus:                get("genus_s_lower"),
		InfragenericEpithet:  get("infra_genus_s_lower"),
		SpecificEpithet:      get("species_s_lower"),
		InfraspecificEpithet: get("infraspecies_s_lower"),
		NameRemarks:          get("name_status_s_lower"),
		NameStatus:           nomStatus(get("name_status_s_lower"), get("name_status_bot_code_type_s_lower")),
		BasionymID:           get("lookup_basionym_id"),
		BasionymAuthorship:   get("basionym_author_s_lower"),
	}
}

// nomStatus derives NomStatus from IPNI's two status fields.
// bot_code_type is the clean ICBN abbreviation when set — use it directly.
// Fall back to name_status_s_lower, extracting the leading recognised prefix
// since that field may contain compound values like "nom. illeg. later homonym".
func nomStatus(statusStr, botCodeType string) coldp.NomStatus {
	if botCodeType != "" {
		return coldp.NewNomStatus(botCodeType)
	}
	if statusStr == "" {
		return coldp.NewNomStatus("")
	}
	lower := strings.ToLower(statusStr)
	// Longest-first to avoid partial matches.
	prefixes := []string{
		"nom. et orth. cons.",
		"nom. inval.",
		"nom. illeg.",
		"nom. superfl.",
		"nom. subnud.",
		"nom. cons.",
		"nom. rej.",
		"nom. nud.",
		"orth. var.",
		"isonym",
		"pro syn.",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return coldp.NewNomStatus(p)
		}
	}
	return coldp.NewNomStatus(statusStr)
}

func buildIndex(headers []string) map[string]int {
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[strings.TrimSpace(h)] = i
	}
	return idx
}
