package mdd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnuuid"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

var parser = gnparser.New(gnparser.NewConfig(
	gnparser.OptCode(nomcode.Zoological),
	gnparser.OptWithDetails(true),
))

func (m *mdd) importSpecies() (map[string]string, error) {
	path, err := m.speciesFile()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("mdd: opening species file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	hdrRow, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("mdd: reading species header: %w", err)
	}
	col := headerIndex(hdrRow)

	sciNameToID := make(map[string]string)
	refs := make(map[string]coldp.Reference)
	var nameUsages []coldp.NameUsage
	var vernaculars []coldp.Vernacular
	var count int

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("mdd: reading species row: %w", err)
		}

		id := "mdd:" + row[col["id"]]
		sciNameToID[strings.ReplaceAll(row[col["sciName"]], "_", " ")] = id

		nu, ref := buildSpeciesUsage(row, col, id)
		data.AddParsedData(parser, nu)
		nameUsages = append(nameUsages, *nu)

		if ref != nil {
			if _, exists := refs[ref.ID]; !exists {
				refs[ref.ID] = *ref
			}
		}

		vernaculars = append(vernaculars, vernacularsForSpecies(row, col, id)...)
		count++
	}

	refSlice := make([]coldp.Reference, 0, len(refs))
	for _, v := range refs {
		refSlice = append(refSlice, v)
	}
	if err := m.sfga.InsertReferences(refSlice); err != nil {
		return nil, fmt.Errorf("mdd: inserting species references: %w", err)
	}
	if err := m.sfga.InsertNameUsages(nameUsages); err != nil {
		return nil, fmt.Errorf("mdd: inserting species name usages: %w", err)
	}
	if err := m.sfga.InsertVernaculars(vernaculars); err != nil {
		return nil, fmt.Errorf("mdd: inserting vernaculars: %w", err)
	}

	slog.Info("MDD species imported", "total", count)
	return sciNameToID, nil
}

func buildSpeciesUsage(
	row []string, col map[string]int, id string,
) (*coldp.NameUsage, *coldp.Reference) {
	genus := strings.TrimSpace(row[col["genus"]])
	epithet := strings.TrimSpace(row[col["specificEpithet"]])
	sciName := genus + " " + epithet

	author := strings.TrimSpace(row[col["authoritySpeciesAuthor"]])
	year := strings.TrimSpace(row[col["authoritySpeciesYear"]])
	parens := strings.TrimSpace(row[col["authorityParentheses"]])
	authorship := buildAuthorship(author, year, parens)

	citation := strings.TrimSpace(row[col["authoritySpeciesCitation"]])
	link := strings.TrimSpace(row[col["authoritySpeciesLink"]])
	if !strings.HasPrefix(link, "http") {
		link = ""
	}

	var refID string
	var ref *coldp.Reference
	if citation != "" {
		refID = "sf_" + gnuuid.New(citation).String()
		ref = &coldp.Reference{ID: refID, Citation: citation, Link: link}
	}

	mddID := row[col["id"]]
	nu := &coldp.NameUsage{
		ID:                   id,
		AlternativeID:        "gnoutlink:" + mddID,
		ScientificName:       sciName,
		ScientificNameString: strings.TrimSpace(sciName + " " + authorship),
		Authorship:           authorship,
		GenericName:          genus,
		SpecificEpithet:      epithet,
		TaxonomicStatus:      coldp.AcceptedTS,
		NameStatus:           coldp.Established,
		Code:                 nomcode.Zoological,
		Rank:                 coldp.NewRank("species"),
		Kingdom:              "Animalia",
		Phylum:               "Chordata",
		Class:                "Mammalia",
		Subclass:             normRank(row[col["subclass"]]),
		Order:                normRank(row[col["order"]]),
		Superfamily:          normRank(row[col["superfamily"]]),
		Family:               normRank(row[col["family"]]),
		Subfamily:            normRank(row[col["subfamily"]]),
		Tribe:                normRank(row[col["tribe"]]),
		Genus:                genus,
		Subgenus:             normRank(row[col["subgenus"]]),
		NameReferenceID:      refID,
		Link:                 fmt.Sprintf("https://www.mammaldiversity.org/taxon/%s/", mddID),
		Extinct:              coldp.ToBool(row[col["extinct"]]),
	}

	return nu, ref
}

func vernacularsForSpecies(
	row []string, col map[string]int, id string,
) []coldp.Vernacular {
	var res []coldp.Vernacular

	if main := strings.TrimSpace(row[col["mainCommonName"]]); main != "" {
		res = append(res, coldp.Vernacular{
			TaxonID:   id,
			Name:      main,
			Language:  "eng",
			Preferred: coldp.ToBool(true),
		})
	}

	others := strings.TrimSpace(row[col["otherCommonNames"]])
	for name := range strings.SplitSeq(others, "|") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		res = append(res, coldp.Vernacular{
			TaxonID:  id,
			Name:     name,
			Language: "eng",
		})
	}

	return res
}

func (m *mdd) speciesFile() (string, error) {
	pattern := filepath.Join(m.cfg.ExtractDir, "MDD", "MDD_v*species.csv")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("mdd: no species CSV found at %s", pattern)
	}
	return matches[0], nil
}

// buildAuthorship formats author + year, optionally wrapping in parentheses.
// parens is "1" when the name was originally described in a different genus.
func buildAuthorship(author, year, parens string) string {
	auth := author
	if year != "" {
		auth = author + ", " + year
	}
	if parens == "1" {
		auth = "(" + auth + ")"
	}
	return strings.TrimSpace(auth)
}

// normNA returns "" for blank or "NA" values, trimmed otherwise.
func normNA(s string) string {
	s = strings.TrimSpace(s)
	if strings.EqualFold(s, "NA") {
		return ""
	}
	return s
}

// normRank title-cases a classification rank value, treating blank/"NA" as "".
func normRank(s string) string {
	s = normNA(s)
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func headerIndex(hdr []string) map[string]int {
	res := make(map[string]int, len(hdr))
	for i, v := range hdr {
		res[v] = i
	}
	return res
}
