package itis

import "github.com/gnames/coldp/ent/coldp"

func (i *itis) importMeta() error {
	meta := coldp.Meta{
		Title: "Integrated Taxonomic Information System",
		Description: "The White House Subcommittee on Biodiversity and Ecosystem " +
			"Dynamics has identified systematics as a research priority" +
			"that is fundamental to ecosystem management and biodiversity " +
			"conservation. This primary need identified by the Subcommittee " +
			"requires improvements in the organization of, and access to, " +
			"standardized nomenclature. ITIS (originally referred to as the " +
			"Interagency Taxonomic Information System) was designed to fulfill " +
			"these requirements. In the future, the ITIS will provide taxonomic " +
			"data and a directory of taxonomic expertise that will support the " +
			"system",
	}
	i.sfga.InsertMeta(&meta)
	return nil
}
