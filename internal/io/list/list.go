package list

import (
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/itis"
	"github.com/sfborg/harvester/pkg/config"
)

func DataSets(cfg config.Config) map[string]data.Convertor {
	res := map[string]data.Convertor{
		"itis": itis.New(cfg),
	}
	return res
}
