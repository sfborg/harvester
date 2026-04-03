package iucn

import (
	"github.com/sfborg/harvester/internal/sources/clbsrc"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
)

func New(cfg config.Config) data.Convertor {
	return clbsrc.New(cfg, clbsrc.Dataset{
		Label:     "iucn",
		Name:      "The IUCN Red List of Threatened Species",
		DatasetID: 53131,
		Notes: `IUCN Red List from ChecklistBank (dataset 53131).
Data is downloaded automatically via the ChecklistBank API as a
ColDP export. Requires --clb-user and --clb-password flags for
authentication.`,
	})
}
