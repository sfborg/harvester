package list

import (
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/sources/ion"
	"github.com/sfborg/harvester/internal/io/sources/itis"
	"github.com/sfborg/harvester/internal/io/sources/text"
	"github.com/sfborg/harvester/internal/io/sources/worldplants"
	"github.com/sfborg/harvester/pkg/config"
)

func GetDataSets(cfg config.Config) map[string]data.Convertor {
	// The keys of the map are the names of the data sources, and the
	//  values are the corresponding data converters.
	res := map[string]data.Convertor{
		"ion":          ion.New(cfg),
		"itis":         itis.New(cfg),
		"text":         text.New(cfg),
		"world-plants": worldplants.New(cfg),
	}
	return res
}
