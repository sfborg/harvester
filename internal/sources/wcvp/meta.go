package wcvp

import "github.com/sfborg/sflib/pkg/coldp"

func (w *wcvp) importMeta() error {
	meta := coldp.Meta{
		Title: "The World Checklist of Vascular Plants (WCVP)",
		Alias: "WCVP",
		Description: "The World Checklist of Vascular Plants (WCVP) is a global " +
			"consensus view of all known vascular plant species " +
			"(flowering plants, conifers, ferns, clubmosses and " +
			"firmosses). WCVP aims to represent a global consensus view " +
			"of current plant taxonomy by reflecting the latest " +
			"published taxonomies while incorporating the opinions of " +
			"taxonomists based around the world. WCVP is built on the " +
			"nomenclatural data provided by the International Plant " +
			"Names Index (IPNI), which is the product of a " +
			"collaboration between The Royal Botanic Gardens, Kew, " +
			"The Harvard University Herbaria, and the Australian " +
			"National Herbarium, combined with the taxonomic data " +
			"provided by an international collaborative programme with " +
			"a large number of contributors from around the world.",
		URL:      "https://wcvp.science.kew.org",
		License:  "CC BY 3.0",
		DOI:      "https://doi.org/10.34885/rvc3-4d77",
		Keywords: []string{"Kew", "WCVP", "Plants", "World", "Taxonomy"},
		Creators: []coldp.Actor{
			{
				Given:        "Rafaël",
				Family:       "Govaerts",
				Orcid:        "0000-0003-2991-5282",
				Organization: "The Royal Botanic Gardens, Kew",
				City:         "London",
				Country:      "UK",
				URL:          "https://www.kew.org/",
			},
		},
	}
	if w.cfg.ArchiveDate != "" {
		meta.Issued = w.cfg.ArchiveDate
	}
	if w.cfg.ArchiveVersion != "" {
		meta.Version = w.cfg.ArchiveVersion
	}
	return w.sfga.InsertMeta(&meta)
}
