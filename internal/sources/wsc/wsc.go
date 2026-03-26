package wsc

import (
	"github.com/sfborg/harvester/internal/sources/clbsrc"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
)

func New(cfg config.Config) data.Convertor {
	return clbsrc.New(cfg, clbsrc.Dataset{
		Label:     "wsc",
		Name:      "World Spider Catalog",
		DatasetID: 56185,
		TaxonName: "Arachnida",
		TaxonRank: "class",
		Notes: `World Spider Catalog from ChecklistBank (dataset 56185).
Data is downloaded automatically via the ChecklistBank API as a
ColDP export. Requires --clb-user and --clb-password flags for
authentication.

Root taxon 'Arachnida' is resolved dynamically by name search.
Use --clb-taxon-id to override with a specific taxon ID.`,
	})
}
