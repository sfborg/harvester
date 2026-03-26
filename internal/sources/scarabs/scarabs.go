package scarabs

import (
	"github.com/sfborg/harvester/internal/sources/clbsrc"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
)

func New(cfg config.Config) data.Convertor {
	return clbsrc.New(cfg, clbsrc.Dataset{
		Label:     "scarabs",
		Name:      "World Scarabaeidae Database",
		DatasetID: 1027,
		TaxonName: "Scarabaeoidea",
		TaxonRank: "superfamily",
		Notes: `World Scarabaeidae Database from ChecklistBank (dataset 1027).
Data is downloaded automatically via the ChecklistBank API as a
ColDP export. Requires --clb-user and --clb-password flags for
authentication.`,
	})
}
