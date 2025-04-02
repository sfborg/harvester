package ion

import "github.com/sfborg/sflib/pkg/coldp"

func (i *ion) importMeta() error {
	meta := coldp.Meta{
		Title: "Index to Organism Names",
		Description: "ION contains millions of animal names, both fossil and " +
			"recent, at all taxonomic ranks, reported from the scientific " +
			"literature. (Bacteria, plant and virus names will be added soon)." +
			"\n\n" +
			"These names are derived from premier Clarivate databases: " +
			"Zoological Record®, BIOSIS Previews®, and Biological Abstracts®. " +
			"All names are tied to at least one published article. Together, " +
			"these resources cover every aspect of the life sciences - " +
			"providing names from over 30 million scientific records, " +
			"including approximately ,000 international journals, patents, " +
			"books, and conference proceedings. They provide a powerful " +
			"foundation for the most complete collection of organism names " +
			"available today.",
	}
	if i.cfg.ArchiveDate != "" {
		meta.Issued = i.cfg.ArchiveDate
	}
	i.sfga.InsertMeta(&meta)
	return nil
}
