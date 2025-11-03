package worldplants

import (
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/google/uuid"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type worldplants struct {
	data.Convertor
	cfg       config.Config
	set       data.DataSet
	sfga      sfga.Archive
	parser    gnparser.GNparser
	namespace uuid.UUID
}

var New = func(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "wfwp",
		Name:  "World Ferns and World Plants",
		Notes: `World Ferns and World Plants data must be provided as a zip
file or directory containing ferns.csv and numbered plant CSV files
(1.csv, 2.csv, etc.). Use --file option to specify the path.
Examples:
  harvester get wfwp --file ~/data/wfwp.zip
  harvester get wfwp --file ~/data/wfwp/`,
		ManualSteps: true,
	}

	opts := []gnparser.Option{
		gnparser.OptCode(nomcode.Botanical),
		gnparser.OptWithDetails(true),
	}
	parserCfg := gnparser.NewConfig(opts...)

	res := worldplants{
		cfg:       cfg,
		set:       set,
		Convertor: base.New(cfg, &set),
		parser:    gnparser.New(parserCfg),
		namespace: uuid.NewSHA1(uuid.NameSpaceOID, []byte("SFBORG::WFWP")),
	}
	return &res
}
