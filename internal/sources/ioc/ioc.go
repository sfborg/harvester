package ioc

import (
	"log/slog"
	"path/filepath"

	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type ioc struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
	path string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "ioc",
		Name:  "IOC World ",
		Notes: `Create tsv file from current master file at
https://www.worldbirdnames.org/new/ioc-lists/master-list-2/
and save to the box.com, generate new URL and update it here.`,
		ManualSteps: true,
		URL:         "https://uofi.box.com/shared/static/x9f7o161l81my22by0k8ov2kgfmuuunu.tsv",
	}
	res := ioc{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (l *ioc) Extract(path string) error {
	slog.Info("Copying IOC World Birds List")
	file := filepath.Base(path)
	l.path = filepath.Join(l.cfg.ExtractDir, file)
	_, err := gnsys.CopyFile(path, l.path)
	if err != nil {
		return err
	}
	return nil
}
