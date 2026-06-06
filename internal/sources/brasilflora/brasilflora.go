package brasilflora

import (
	"github.com/sfborg/harvester/internal/sources/clbsrc"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
)

func New(cfg config.Config) data.Convertor {
	return clbsrc.New(cfg, clbsrc.Dataset{
		Label:     "brasil-flora",
		Name:      "Flora e Funga do Brasil",
		DatasetID: 2031,
		Notes: `Flora e Funga do Brasil - Lista Oficial from ChecklistBank (dataset 2031).
Data is downloaded automatically via the ChecklistBank API as a
ColDP export. Requires --clb-user and --clb-password flags for
authentication.`,
	})
}
