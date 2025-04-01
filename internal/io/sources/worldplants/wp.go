package worldplants

import (
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/pkg/sfga"
)

type worldplants struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
}

var New = func(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "world-plants",
		Name:  "World of Plants",
		Notes: `
    World of Plants data file has to provided either local path or remote URL.
    Use --file option.
    `,
		ManualSteps: true,
	}
	res := worldplants{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}
