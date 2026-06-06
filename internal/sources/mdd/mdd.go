package mdd

import (
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type mdd struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label:       "mdd",
		Name:        "Mammal Diversity Database",
		ManualSteps: false,
		URL:         "https://github.com/mammaldiversity/mammaldiversity.github.io/raw/refs/heads/master/assets/data/MDD.zip",
		Notes: `MDD is the Mammal Diversity Database maintained by the
American Society of Mammalogists. Data is downloaded automatically
from the MDD website. Use --load-file to supply a local copy instead.

Example:
  harvester get mdd
  harvester get mdd --load-file ~/Downloads/MDD.zip`,
	}
	res := mdd{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (m *mdd) InitSfga() (sfga.Archive, error) {
	arc, err := m.Convertor.InitSfga()
	if err != nil {
		return nil, err
	}
	m.sfga = arc
	return arc, nil
}
