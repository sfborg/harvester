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
	"github.com/gnames/gnuuid"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

func (m *mdd) importSynonyms(sciNameToID map[string]string) error {
	path, err := m.synonymFile()
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("mdd: opening synonym file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	hdrRow, err := r.Read()
	if err != nil {
		return fmt.Errorf("mdd: reading synonym header: %w", err)
	}
	col := headerIndex(hdrRow)

	refs := make(map[string]coldp.Reference)
	seen := make(map[string]struct{})
	var nameUsages []coldp.NameUsage
	var count int

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("mdd: reading synonym row: %w", err)
		}

		// Skip rows that mirror accepted species — already imported.
		if strings.TrimSpace(row[col["MDD_validity"]]) == "species" {
			continue
		}

		nu, ref := buildSynonymUsage(row, col, sciNameToID)
		if nu == nil {
			continue
		}
		if _, dup := seen[nu.ID]; dup {
			continue
		}
		seen[nu.ID] = struct{}{}

		data.AddParsedData(parser, nu)
		nameUsages = append(nameUsages, *nu)

		if ref != nil {
			if _, exists := refs[ref.ID]; !exists {
				refs[ref.ID] = *ref
			}
		}

		count++
		if count%5000 == 0 {
			slog.Info("importing MDD synonyms", "count", count)
		}
	}

	refSlice := make([]coldp.Reference, 0, len(refs))
	for _, v := range refs {
		refSlice = append(refSlice, v)
	}
	if err := m.sfga.InsertReferences(refSlice); err != nil {
		return fmt.Errorf("mdd: inserting synonym references: %w", err)
	}
	if err := m.sfga.InsertNameUsages(nameUsages); err != nil {
		return fmt.Errorf("mdd: inserting synonyms: %w", err)
	}

	slog.Info("MDD synonyms and subspecies imported", "total", count)
	return nil
}

func buildSynonymUsage(
	row []string, col map[string]int, sciNameToID map[string]string,
) (*coldp.NameUsage, *coldp.Reference) {
	genus := normNA(row[col["MDD_genus"]])
	epithet := normNA(row[col["MDD_specificEpithet"]])
	origRank := strings.TrimSpace(row[col["MDD_original_rank"]])

	// MDD_subspecificEpithet is empty/NA for all subspecies rows;
	// the infraspecific epithet is stored in MDD_root_name instead.
	subEpithet := normNA(row[col["MDD_subspecificEpithet"]])
	if subEpithet == "" && origRank == "subspecies" {
		subEpithet = normNA(row[col["MDD_root_name"]])
	}

	parts := []string{genus, epithet}
	if subEpithet != "" {
		parts = append(parts, subEpithet)
	}
	sciName := strings.TrimSpace(strings.Join(parts, " "))
	if sciName == "" {
		return nil, nil
	}

	// A subspecies must be trinomial; skip if we could not form three parts.
	if origRank == "subspecies" && len(parts) < 3 {
		return nil, nil
	}

	id := "mdd-syn:" + strings.TrimSpace(row[col["MDD_syn_ID"]])
	acceptedID := sciNameToID[strings.TrimSpace(row[col["MDD_species"]])]

	author := strings.TrimSpace(row[col["MDD_author"]])
	year := strings.TrimSpace(row[col["MDD_year"]])
	parens := strings.TrimSpace(row[col["MDD_authority_parentheses"]])
	authorship := buildAuthorship(author, year, parens)

	citation := strings.TrimSpace(row[col["MDD_authority_citation"]])
	link := strings.TrimSpace(row[col["MDD_authority_link"]])
	if !strings.HasPrefix(link, "http") {
		link = ""
	}

	var refID string
	var ref *coldp.Reference
	if citation != "" {
		refID = "sf_" + gnuuid.New(citation).String()
		ref = &coldp.Reference{ID: refID, Citation: citation, Link: link}
	}

	validity := strings.TrimSpace(row[col["MDD_validity"]])
	rank := coldp.NewRank(origRank)

	taxStatus := taxStatusFromValidity(validity)

	nu := &coldp.NameUsage{
		ID:       id,
		ParentID: acceptedID,
		ScientificName:       sciName,
		ScientificNameString: strings.TrimSpace(sciName + " " + authorship),
		Authorship:           authorship,
		GenericName:          genus,
		SpecificEpithet:      epithet,
		InfraspecificEpithet: subEpithet,
		TaxonomicStatus:      taxStatus,
		Code:                 nomcode.Zoological,
		Rank:                 rank,
		NameReferenceID:      refID,
	}

	return nu, ref
}

func taxStatusFromValidity(validity string) coldp.TaxonomicStatus {
	switch validity {
	case "synonym":
		return coldp.SynonymTS
	case "nomen_dubium", "species_inquirenda", "composite", "hybrid":
		return coldp.AmbiguousSynonymTS
	default:
		return coldp.SynonymTS
	}
}

func (m *mdd) synonymFile() (string, error) {
	pattern := filepath.Join(m.cfg.ExtractDir, "MDD", "Species_Syn_v*.csv")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("mdd: no synonym CSV found at %s", pattern)
	}
	return matches[0], nil
}
