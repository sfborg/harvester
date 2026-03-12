package mycobank

import "github.com/sfborg/sflib/pkg/coldp"

func (m *mycobank) importMeta() error {
	meta := coldp.Meta{
		Title: "MycoBank",
		Alias: "MycoBank",
		Description: "MycoBank is an online database aimed as a service to the " +
			"mycological and scientific community by documenting mycological " +
			"nomenclatural novelties (new names and combinations) and associated " +
			"data. Westerdijk Fungal Biodiversity Institute.",
		URL:            "https://www.mycobank.org",
		License:        "CC BY 4.0",
		TaxonomicScope: "Fungi",
		Keywords: []string{
			"taxonomy", "nomenclature", "fungi", "mycology",
		},
	}
	if m.cfg.ArchiveDate != "" {
		meta.Issued = m.cfg.ArchiveDate
	}
	if m.cfg.ArchiveVersion != "" {
		meta.Version = m.cfg.ArchiveVersion
	}
	return m.sfga.InsertMeta(&meta)
}
