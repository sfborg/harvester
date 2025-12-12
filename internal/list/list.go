package list

import (
	"github.com/sfborg/harvester/internal/sources/grin"
	"github.com/sfborg/harvester/internal/sources/ioc"
	"github.com/sfborg/harvester/internal/sources/ion"
	"github.com/sfborg/harvester/internal/sources/ncbi"
	"github.com/sfborg/harvester/internal/sources/paleodb"
	"github.com/sfborg/harvester/internal/sources/wikisp"
	"github.com/sfborg/harvester/internal/sources/worldplants"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
)

func GetDataSets(cfg config.Config) map[string]data.Convertor {
	// The keys of the map are the names of the data sources, and the
	//  values are the corresponding data converters.
	ds := []data.Convertor{
		grin.New(cfg),
		ion.New(cfg),
		ioc.New(cfg),
		ncbi.New(cfg),
		paleodb.New(cfg),
		worldplants.New(cfg),
		wikisp.New(cfg),
	}

	res := make(map[string]data.Convertor)

	for i := range ds {
		res[ds[i].Label()] = ds[i]
	}
	return res
}
