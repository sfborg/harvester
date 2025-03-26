package grin

import (
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/pkg/sfga"
)

type grin struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
}

func New(cfg config.Config) data.Convertor {
	set := data.Set{
		Label: "grin",
		Name:  "USDA-ARS Germplasm Resources Information Network (GRIN)",
		Description: `
The Germplasm Resources Information Network (GRIN) provides information
about USDA national collections of plant genetic resources (germplasm).
`,
		ManualSteps: true,
		URL:         "https://uofi.box.com/v/grin-taxonomy",
	}
	res := grin{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (g *grin) Import(path string) error {
	return nil
}
