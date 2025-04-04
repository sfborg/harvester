package paleodb

import "github.com/sfborg/sflib/pkg/coldp"

func (g *paleodb) importMeta() error {
	meta := coldp.Meta{
		Title: "The Paleobiology Database",
		Alias: "Paleobiodb",
		Description: "The Paleobiology Database is an online, non-governmental, " +
			"non-profit public resource for paleontological data. It is organized " +
			"and operated by a multi-disciplinary, multi-institutional, " +
			"international group of paleobiological researchers.",
		URL: "https://paleobiodb.org",
	}
	if g.cfg.ArchiveDate != "" {
		meta.Issued = g.cfg.ArchiveDate
	}
	g.sfga.InsertMeta(&meta)

	return nil
}
