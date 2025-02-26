package list

import (
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/ion"
	"github.com/sfborg/harvester/internal/io/itis"
	worldplants "github.com/sfborg/harvester/internal/io/world-plants"
	"github.com/sfborg/harvester/pkg/config"
)

func DataSets(cfg config.Config) map[string]data.Convertor {
	res := map[string]data.Convertor{
		"ion":          ion.New(cfg),
		"itis":         itis.New(cfg),
		"world-plants": worldplants.New(cfg),
	}
	return res
}
