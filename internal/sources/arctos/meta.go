package arctos

import "github.com/sfborg/sflib/pkg/coldp"

func (a *arctos) importMeta() error {
	meta := coldp.Meta{
		Title: "Arctos",
		Alias: "Arctos",
		Description: "Arctos is an ongoing effort to integrate access to " +
			"specimen data, collection-management tools, and external " +
			"resources on the internet. It serves as a collection management " +
			"system for natural history collections and provides access to " +
			"taxonomic and nomenclatural data aggregated from multiple sources.",
		URL:     "https://arctos.database.museum",
		License: "CC0",
		Keywords: []string{
			"taxonomy", "nomenclature", "natural history", "specimens",
			"collections",
		},
	}
	if a.cfg.ArchiveDate != "" {
		meta.Issued = a.cfg.ArchiveDate
	}
	if a.cfg.ArchiveVersion != "" {
		meta.Version = a.cfg.ArchiveVersion
	}
	return a.sfga.InsertMeta(&meta)
}
