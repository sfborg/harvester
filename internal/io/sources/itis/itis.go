package itis

import (
	"database/sql"

	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/ent/sfga"
)

type itis struct {
	data.Convertor
	cfg    config.Config
	itisDb *sql.DB
	sfga   sfga.Archive
}

var New = func(cfg config.Config) data.Convertor {
	set := data.Set{
		Label:       "itis",
		Name:        "ITIS",
		ManualSteps: false,
		URL:         "https://www.itis.gov/downloads/itisSqlite.zip",
	}
	res := itis{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}
