package paleodb

import (
	"database/sql"
	"net/http"
	"time"

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
	http *http.Client
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
		http:      httpClient(),
	}
	return &res
}

func httpClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:      10,
		IdleConnTimeout:   600 * time.Second,
		ForceAttemptHTTP2: false,
	}
	return &http.Client{Timeout: 5 * time.Minute, Transport: tr}
}
