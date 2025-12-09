package base

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/fatih/color"
	"github.com/gnames/gn"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/sysio"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib"
	"github.com/sfborg/sflib/pkg/sfga"
)

// Convertor implements default methods of data.Convertor interface.
type Convertor struct {
	set *data.DataSet
	cfg config.Config
	gnp gnparser.GNparser
}

func New(cfg config.Config, s *data.DataSet) data.Convertor {
	res := Convertor{
		cfg: cfg,
		set: s,
	}
	gncfg := gnparser.NewConfig(
		gnparser.OptCode(cfg.Code),
		gnparser.OptWithDetails(true),
	)
	res.gnp = gnparser.New(gncfg)
	return &res
}

func (c *Convertor) Config() config.Config {
	return c.cfg
}

func (c *Convertor) Label() string {

	res := c.set.Label
	if c.set.ManualSteps {
		res += color.RedString("*")
	}
	return res
}

func (c *Convertor) Name() string {
	return c.set.Name
}

func (c *Convertor) Description() string {
	return c.set.Notes
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

	if c.cfg.LoadFile != "" {
		slog.Info(
			"using local file", "source", c.set.Label, "file", c.cfg.LoadFile,
		)
		gn.Info("Using local file for %s: %s", c.set.Label, c.cfg.LoadFile)
		return c.cfg.LoadFile, nil
	}

	if c.set.URL == "" {
		err = errors.New("no local file or URL given")
		return "", err
	}

	slog.Info("downloading", "source", c.set.Label)
	gn.Info("Downloading %s", c.set.Label)
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
	case gnsys.Bz2FT:
		f = gnsys.ExtractBz2
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

func (c *Convertor) InitSfga() (sfga.Archive, error) {
	sysio.EmptyDir(c.cfg.SfgaDir)

	sfga := sflib.NewSfga()
	err := sfga.Create(c.cfg.SfgaDir)
	if err != nil {
		return nil, err
	}
	_, err = sfga.Connect()
	if err != nil {
		return nil, err
	}
	return sfga, nil
}

func (c *Convertor) ToSfga(_ sfga.Archive) error {
	slog.Info("running a placeholder ToSfga method")
	gn.Warn("Runs a placeholder for ToSfga method")
	return nil
}
