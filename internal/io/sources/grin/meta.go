package grin

import "github.com/sfborg/sflib/pkg/coldp"

func (g *grin) importMeta() error {
	meta := coldp.Meta{
		Title: "USDA National Plant Germplasm System",
		Alias: "GRIN Plant Taxonomy",
		Description: "The USDA National Plant Germplasm System (NPGS), " +
			"often referred to through its associated database, " +
			"the Germplasm Resources Information Network (GRIN), " +
			"is a vital resource for preserving and providing access " +
			"to plant genetic diversity.\n\n" +

			"The Germplasm Resources Information Network " +
			"(GRIN) provides information about the United States Department " +
			"of Agriculture (USDA national collections of animal, microbial, " +
			"and plant genetic resources (germplasm) important for food " +
			"and agricultural production. GRIN documents these collections " +
			"through informational pages, searchable databases, and links " +
			"to USDA-ARS projects that curate the collections. ",
		URL: "https://npgsweb.ars-grin.gov/gringlobal/taxon/abouttaxonomy",
	}
	if g.cfg.ArchiveDate != "" {
		meta.Issued = g.cfg.ArchiveDate
	}
	g.sfga.InsertMeta(&meta)
	return nil
}
