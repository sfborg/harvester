package ott

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gnames/gnparser"
	"github.com/gnames/gnuuid"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/sfborg/sflib/pkg/sfga"
)

func (o *ott) ToSfga(sfga sfga.Archive) error {
	o.sfga = sfga

	if err := o.collectSynonyms(); err != nil {
		return err
	}
	if err := o.setMetadata(); err != nil {
		return err
	}
	return o.setNameUsage()
}

func (o *ott) setMetadata() error {
	meta := coldp.Meta{
		Title: "Open Tree Taxonomy",
		Alias: "OTT",
		Description: "The Open Tree Taxonomy (OTT) is a comprehensive taxonomy " +
			"of all life assembled from multiple source taxonomies including " +
			"NCBI, GBIF, WoRMS, and others.",
		URL:            "https://tree.opentreeoflife.org/about/taxonomy-version",
		License:        "CC0",
		TaxonomicScope: "All life",
		Keywords:       []string{"taxonomy", "biodiversity", "species", "nomenclature"},
	}
	if err := o.sfga.InsertMeta(&meta); err != nil {
		return fmt.Errorf("insert metadata: %w", err)
	}
	return nil
}

func (o *ott) setNameUsage() error {
	f, err := os.Open(o.taxonomyPath)
	if err != nil {
		return fmt.Errorf("open taxonomy: %w", err)
	}
	defer f.Close()

	gnp := gnparser.New(gnparser.NewConfig(gnparser.OptWithDetails(true)))

	var nus []coldp.NameUsage
	var totalImported, rejectedNum, rejectedSyn int
	batchSize := o.cfg.BatchSize

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		line = strings.TrimRight(line, "\t| ")
		fields := strings.Split(line, "\t|\t")
		if len(fields) < 4 {
			continue
		}

		d, ok := parseTaxonLine(fields, o.synonyms)
		if !ok {
			continue
		}

		nu := coldp.NameUsage{
			ID:                   d.taxonID,
			ParentID:             d.parentID,
			ScientificNameString: d.name,
			ScientificName:       d.name,
			TaxonomicStatus:      coldp.AcceptedTS,
			Rank:                 coldp.NewRank(d.rank),
			AlternativeID:        ottAlternativeID(d.taxonID),
		}
		data.AddParsedData(gnp, &nu)
		if !isParsedOK(&nu) {
			rejectedNum++
			continue
		}
		nus = append(nus, nu)

		rejSyn, syns := o.synonymNameUsages(gnp, d)
		rejectedSyn += rejSyn
		nus = append(nus, syns...)

		if len(nus) >= batchSize {
			if err := o.sfga.InsertNameUsages(nus); err != nil {
				return err
			}
			totalImported += len(nus)
			nus = nus[:0]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(nus) > 0 {
		if err := o.sfga.InsertNameUsages(nus); err != nil {
			return err
		}
		totalImported += len(nus)
	}

	fmt.Printf(`
Imported names:    %d
Rejected names:    %d
Rejected synonyms: %d
`, totalImported, rejectedNum, rejectedSyn)

	return nil
}

func (o *ott) synonymNameUsages(
	gnp gnparser.GNparser,
	d datum,
) (int, []coldp.NameUsage) {
	var rejectNum int
	var res []coldp.NameUsage
	seen := make(map[string]bool)
	for _, s := range d.synonyms {
		id := "sf-" + gnuuid.New(s.name).String()
		if seen[id] {
			continue
		}
		seen[id] = true
		nu := coldp.NameUsage{
			ID:                   id,
			ParentID:             d.taxonID,
			ScientificNameString: s.name,
			ScientificName:       s.name,
			TaxonomicStatus:      coldp.SynonymTS,
		}
		data.AddParsedData(gnp, &nu)
		if !isParsedOK(&nu) {
			rejectNum++
			continue
		}
		res = append(res, nu)
	}
	return rejectNum, res
}

func isParsedOK(nu *coldp.NameUsage) bool {
	if nu.Virus.Bool {
		return true
	}
	switch nu.ParseQuality.Int64 {
	case 1, 2, 3:
		return true
	default:
		return false
	}
}

func ottAlternativeID(uid string) string {
	return fmt.Sprintf("ott:%s,gnoutlink:%s", uid, uid)
}
