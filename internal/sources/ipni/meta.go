package ipni

import "github.com/sfborg/sflib/pkg/coldp"

func (i *ipni) importMeta() error {
	meta := coldp.Meta{
		Title: "The International Plant Names Index",
		Alias: "IPNI",
		Description: "The International Plant Names Index (IPNI) is a database of " +
			"the names and associated basic bibliographical details of seed plants, " +
			"ferns and lycophytes. Its goal is to eliminate the need for repeated " +
			"reference to primary sources for basic bibliographic information about " +
			"plant names. The data are freely available and are gradually being " +
			"standardized and checked. IPNI will be a dynamic resource, depending " +
			"on direct contributions by all members of the botanical community.",
		URL:            "https://www.ipni.org",
		License:        "CC BY 4.0",
		TaxonomicScope: "Seed plants, ferns and lycophytes",
		Keywords: []string{
			"taxonomy", "nomenclature", "botany", "plants", "biodiversity",
		},
	}
	if i.cfg.ArchiveDate != "" {
		meta.Issued = i.cfg.ArchiveDate
	}
	if i.cfg.ArchiveVersion != "" {
		meta.Version = i.cfg.ArchiveVersion
	}
	return i.sfga.InsertMeta(&meta)
}
