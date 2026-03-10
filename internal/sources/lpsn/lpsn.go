package lpsn

import (
	"log/slog"
	"path/filepath"

	"github.com/gnames/gn"
	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type lpsn struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
	path string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "lpsn",
		Name:  "List of Prokaryotic names with Standing in Nomenclature",
		Notes: `Download the CSV file manually from https://lpsn.dsmz.de/downloads
(requires free registration). Save the file locally and provide it
with the -f flag:

  harvester get lpsn -f path/to/lpsn_gss_<date>.csv`,
		ManualSteps: true,
		URL:         "",
	}
	res := lpsn{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (l *lpsn) Extract(path string) error {
	slog.Info("copying LPSN CSV file")
	gn.Info("Copying LPSN CSV file")
	file := filepath.Base(path)
	l.path = filepath.Join(l.cfg.ExtractDir, file)
	_, err := gnsys.CopyFile(path, l.path)
	if err != nil {
		return err
	}
	return nil
}
