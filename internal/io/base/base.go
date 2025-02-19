package base

import (
	"fmt"
	"log/slog"

	"github.com/gnames/gnsys"
	"github.com/sfborg/from-coldp/pkg/ent/sfgarc"
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/sysio"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/ent/sfga"
)

// Convertor implements default methods of data.Convertor interface.
type Convertor struct {
	set data.Set
	cfg config.Config
	sch sfga.Schema
}

func New(cfg config.Config, s data.Set) data.Convertor {
	res := Convertor{cfg: cfg, set: s}
	return &res
}

func (c *Convertor) Label() string {
	return c.set.Label
}

func (c *Convertor) Name() string {
	return c.set.Name
}

func (c *Convertor) Description() string {
	return c.set.Description
}
func (c *Convertor) ManualSteps() bool {
	return c.set.ManualSteps
}

func (c *Convertor) Download() (string, error) {
	var err error
	var path string
	err = sysio.ResetCache(c.cfg)
	if err != nil {
		return "", err
	}

	path, err = gnsys.Download(c.set.URL, c.cfg.DownloadDir, true)
	if err != nil {
		return path, err
	}
	return path, nil
}

func (c *Convertor) Extract(path string) error {
	var f gnsys.Extractor
	switch gnsys.GetFileType(path) {
	case gnsys.ZipFT:
		f = gnsys.ExtractZip
	case gnsys.TarFT:
		f = gnsys.ExtractTar
	case gnsys.TarGzFT:
		f = gnsys.ExtractTarGz
	case gnsys.TarBzFT:
		f = gnsys.ExtractTarBz2
	case gnsys.TarXzFt:
		f = gnsys.ExtractTarXz
	default:
		return fmt.Errorf("cannot determine file format of '%s'", path)
	}
	err := f(path, c.cfg.ExtractDir)
	if err != nil {
		return err
	}
	return nil
}

func (c *Convertor) ToSFGA(_ sfgarc.Archive) error {
	slog.Info("Running a placeholder ToSFGA method")
	return nil
}
