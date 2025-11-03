package list

import (
	"github.com/sfborg/harvester/internal/sources/grin"
	"github.com/sfborg/harvester/internal/sources/ioc"
	"github.com/sfborg/harvester/internal/sources/ion"
	"github.com/sfborg/harvester/internal/sources/paleodb"
	"github.com/sfborg/harvester/internal/sources/worldplants"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
)

func GetDataSets(cfg config.Config) map[string]data.Convertor {
	// The keys of the map are the names of the data sources, and the
	//  values are the corresponding data converters.
	res := map[string]data.Convertor{
		"grin":      grin.New(cfg),
		"ion":       ion.New(cfg),
		"ioc birds": ioc.New(cfg),
		"paleodb":   paleodb.New(cfg),
		"wfwp":      worldplants.New(cfg),
	}
	return res
}
