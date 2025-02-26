package worldplants

import (
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/ent/sfga"
)

type worldplants struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
}

var New = func(cfg config.Config) data.Convertor {
	set := data.Set{
		Label: "world-plants",
		Name:  "World of Plants",
		Description: `
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
