package harvester

import "github.com/sfborg/harvester/pkg/data"

type Harvester interface {
	List() map[string]data.Convertor
	Get(datasetLabel, outPath string) error
}
