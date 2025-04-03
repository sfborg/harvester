package paleodb

import (
	"database/sql"

	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type paleodb struct {
	data.Convertor
	cfg  config.Config
	set  data.DataSet
	sfga sfga.Archive
	db   *sql.DB
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label:       "paleodb",
		Name:        "Paleobiology Database",
		Notes:       ``,
		ManualSteps: false,
		URL:         "https://paleobiodb.org/data1.2",
	}
	res := paleodb{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
		set:       set,
	}
	return &res
}
