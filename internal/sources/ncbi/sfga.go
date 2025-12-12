package ncbi

import (
	"fmt"

	"github.com/gnames/gnparser"
	"github.com/gnames/gnuuid"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/sfborg/sflib/pkg/sfga"
)

func (n *ncbi) ToSfga(sfga sfga.Archive) error {
	var err error
	n.sfga = sfga
	err = n.collectNames()
	if err != nil {
		return err
	}
	err = n.collectNodes()
	if err != nil {
		return err
	}

	err = n.populateSfga()
	if err != nil {
		return err
	}

	return nil
}

func (n *ncbi) populateSfga() error {
	var err error
	err = n.setMetadata()
	if err != nil {
		return err
	}

	err = n.setNameUsage()
	if err != nil {
		return err
	}

	return nil
}

func (n *ncbi) setMetadata() error {
	// Insert metadata
	meta := coldp.Meta{
		Title: "National Center for Biotechnology Information",
		Alias: "NCBI",
		Description: "The National Center for Biotechnology Information " +
			"advances science and health by providing access to biomedical " +
			"and genomic information.",
		URL:            "https://www.ncbi.nlm.nih.gov/",
		License:        "CC0",
		TaxonomicScope: "All life",
		Keywords:       []string{"taxonomy", "biodiversity", "species", "nomenclature"},
	}
	if err := n.sfga.InsertMeta(&meta); err != nil {
		return fmt.Errorf("failed to insert metadata: %w", err)
	}
	return nil
}

func (n *ncbi) setNameUsage() error {
	var rejectedNum, rejectedSyn int
	var nus []coldp.NameUsage
	cfg := gnparser.NewConfig(
		gnparser.OptWithDetails(true),
	)
	gnp := gnparser.New(cfg)
	for _, v := range n.data {
		status := coldp.AcceptedTS
		nu := coldp.NameUsage{
			ID:                   v.taxonID,
			TaxonomicStatus:      status,
			Rank:                 coldp.NewRank(v.rank),
			ScientificNameString: v.nameStr,
			ScientificName:       v.canonical,
			ParentID:             v.parentID,
		}
		data.AddParsedData(gnp, &nu)
		if !isParsedOK(&nu) {
			rejectedNum++
			continue
		}
		nus = append(nus, nu)
		var syn []coldp.NameUsage
		if len(v.synonyms) > 0 {
			rejectedSyn, syn = n.synonymNameUsage(gnp, v)
			nus = append(nus, syn...)
		}
	}
	fmt.Printf(`
Imported names:    %d
Rejected names:    %d
Rejected synonyms: %d
`,
		len(nus), rejectedNum, rejectedSyn,
	)
	err := n.sfga.InsertNameUsages(nus)
	if err != nil {
		return err
	}

	return nil
}

func (n *ncbi) synonymNameUsage(
	gnp gnparser.GNparser,
	d datum,
) (int, []coldp.NameUsage) {
	var rejectNum int
	var res []coldp.NameUsage
	for _, v := range d.synonyms {
		nu := coldp.NameUsage{
			ID:                   gnuuid.New(v.name).String(),
			ParentID:             d.taxonID,
			ScientificNameString: v.name,
			ScientificName:       v.name,
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
