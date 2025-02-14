package list

import (
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/itis"
)

var DataSets = map[string]data.Convertor{
	"itis": itis.New(),
}
