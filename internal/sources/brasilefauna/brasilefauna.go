package brasilefauna

import (
	"github.com/sfborg/harvester/internal/sources/clbsrc"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
)

func New(cfg config.Config) data.Convertor {
	return clbsrc.New(cfg, clbsrc.Dataset{
		Label:     "brasil-fauna",
		Name:      "Catálogo Taxonômico da Fauna do Brasil",
		DatasetID: 281817,
		Notes: `Catálogo Taxonômico da Fauna do Brasil from ChecklistBank (dataset 281817).
Data is downloaded automatically via the ChecklistBank API as a
ColDP export. Requires --clb-user and --clb-password flags for
authentication.`,
	})
}
