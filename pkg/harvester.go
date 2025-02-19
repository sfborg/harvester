package harvester

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sort"

	"github.com/sfborg/from-coldp/pkg/ent/sfgarc"
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/list"
	"github.com/sfborg/harvester/pkg/config"
)

type harvester struct {
	cfg    config.Config
	ds     map[string]data.Convertor
	itisDB *sql.DB
	sfga   sfgarc.Archive
}

func New(cfg config.Config) Harvester {
	res := harvester{
		cfg: cfg,
		ds:  list.DataSets(cfg),
	}

	return &res
}

func (h *harvester) List() []string {
	var res []string
	for k := range h.ds {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func (h *harvester) Convert(label, path string) error {
	var err error
	var ds data.Convertor
	var ok bool
	if ds, ok = h.ds[label]; !ok {
		err = fmt.Errorf("Label '%s' does not exist", label)
		return err
	}

	if h.cfg.SkipDownload {
		slog.Info("Skipping download step", "source", ds.Label())
	} else {
		slog.Info("Downloading", "source", ds.Label())
		path, err = ds.Download()
		if err != nil {
			return err
		}

		slog.Info("Extracting files", "source", ds.Label())
		err = ds.Extract(path)
		if err != nil {
			return err
		}
	}

	slog.Info("Creating SFG archive")
	err = ds.ToSFGA(h.sfga)
	if err != nil {
		return err
	}
	return nil
}
