package nzor

import "github.com/sfborg/sflib/pkg/coldp"

func (n *nzor) importMeta() error {
	meta := coldp.Meta{
		Title: "New Zealand Organism Register",
		Alias: "NZOR",
		Description: "NZOR is an actively maintained compilation of all organism names " +
			"relevant to New Zealand: indigenous, endemic or exotic species or species not " +
			"present in New Zealand but of national interest. NZOR is digitally and " +
			"automatically assembled from a number of taxonomic data providers.",
		URL:            "https://www.nzor.org.nz",
		License:        "CC BY 4.0",
		TaxonomicScope: "All organisms relevant to New Zealand",
		Keywords: []string{
			"taxonomy", "nomenclature", "New Zealand", "biodiversity",
		},
	}
	if n.cfg.ArchiveDate != "" {
		meta.Issued = n.cfg.ArchiveDate
	}
	if n.cfg.ArchiveVersion != "" {
		meta.Version = n.cfg.ArchiveVersion
	}
	return n.sfga.InsertMeta(&meta)
}
