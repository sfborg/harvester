package lpsn

import "github.com/sfborg/sflib/pkg/coldp"

func (l *lpsn) importMeta() error {
	meta := coldp.Meta{
		Title: "List of Prokaryotic names with Standing in Nomenclature",
		Alias: "LPSN",
		Description: "LPSN is the authoritative resource for the nomenclature " +
			"of prokaryotes (Bacteria and Archaea). It provides names that " +
			"are validly published under the rules of the International Code " +
			"of Nomenclature of Prokaryotes (ICNP), along with their " +
			"nomenclatural status and taxonomic recommendations.\n\n" +
			"LPSN was founded by Jean P. Euzéby and is currently maintained " +
			"by Aidan C. Parte at the Leibniz Institute DSMZ.",
		URL:            "https://lpsn.dsmz.de",
		License:        "CC BY-SA 4.0",
		TaxonomicScope: "Bacteria, Archaea",
		Keywords: []string{
			"taxonomy", "nomenclature", "bacteria", "archaea", "prokaryotes",
		},
	}
	if l.cfg.ArchiveDate != "" {
		meta.Issued = l.cfg.ArchiveDate
	}
	if l.cfg.ArchiveVersion != "" {
		meta.Version = l.cfg.ArchiveVersion
	}
	return l.sfga.InsertMeta(&meta)
}
