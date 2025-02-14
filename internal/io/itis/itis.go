package itis

import (
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/base"
)

type itis struct {
	data.Convertor
}

var New = func() data.Convertor {
	set := data.Set{
		Label:       "itis",
		Name:        "ITIS",
		ManualSteps: false,
		URL:         "https://www.itis.gov/downloads/itisSqlite.zip",
	}
	res := itis{
		Convertor: base.New(set),
	}
	return &res
}
