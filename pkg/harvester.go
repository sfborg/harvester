package harvester

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sort"

	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/list"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/pkg/sfga"
)

type harvester struct {
	cfg    config.Config
	ds     map[string]data.Convertor
	itisDB *sql.DB
	sfga   sfga.Archive
}

func New(cfg config.Config) Harvester {
	res := harvester{
		cfg: cfg,
		ds:  list.GetDataSets(cfg),
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

func (h *harvester) Get(label, outPath string) error {
	var err error
	var sfga sfga.Archive
	var ds data.Convertor
	var ok bool
	var dlPath string
	if ds, ok = h.ds[label]; !ok {
		err = fmt.Errorf("Label '%s' does not exist", label)
		return err
	}

	if h.cfg.SkipDownload {
		slog.Info("Skipping download step", "source", ds.Label())
	} else {
		dlPath, err = ds.Download()
		if err != nil {
			return err
		}

		slog.Info("Extracting files", "source", ds.Label())
		err = ds.Import(dlPath)
		if err != nil {
			return err
		}
	}

	slog.Info("Creating SFG archive")
	sfga, err = ds.InitSfga()
	if err != nil {
		return err
	}

	err = ds.ToSfga(sfga)
	if err != nil {
		return err
	}

	sfga.Export(outPath, ds.Config().WithZipOutput)
	return nil
}
