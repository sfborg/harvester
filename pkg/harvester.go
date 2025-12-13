package harvester

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/harvester/internal/list"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
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

func (h *harvester) List() map[string]data.Convertor {
	return h.ds
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
		slog.Info("skip download step", "source", ds.Label())
		gn.Message("Skipping download for <em>%s</em>", ds.Label())
	} else {
		dlPath, err = ds.Download()
		if err != nil {
			return err
		}
	}

	slog.Info("extracting files", "source", ds.Label())
	gn.Message("Extracting files of <em>%s</em>", ds.Label())
	err = ds.Extract(dlPath)
	if err != nil {
		return err
	}

	slog.Info("creating SFG archive")
	gn.Message("Creating empty SFGA file")
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
