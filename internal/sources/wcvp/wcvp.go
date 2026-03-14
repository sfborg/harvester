package wcvp

import (
	"path/filepath"

	"github.com/gnames/gn"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type wcvp struct {
	data.Convertor
	cfg     config.Config
	sfga    sfga.Archive
	csvPath string
	refMap  map[string]string // citation key → reference ID
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "wcvp",
		Name:  "The World Checklist of Vascular Plants",
		Notes: `WCVP is a global consensus view of all known vascular plant
species. Data is downloaded automatically from
https://sftp.kew.org/pub/data-repositories/WCVP/wcvp.zip`,
		ManualSteps: false,
		URL:         "https://sftp.kew.org/pub/data-repositories/WCVP/wcvp.zip",
	}
	res := wcvp{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (w *wcvp) Extract(path string) error {
	gn.Info("Extracting WCVP data")
	if err := w.Convertor.Extract(path); err != nil {
		return err
	}
	w.csvPath = filepath.Join(w.cfg.ExtractDir, "wcvp_names.csv")
	return nil
}
