package ipni

import (
	"path/filepath"
	"strings"

	"github.com/gnames/gn"
	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type ipni struct {
	data.Convertor
	cfg      config.Config
	sfga     sfga.Archive
	csvPath  string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "ipni",
		Name:  "The International Plant Names Index",
		Notes: `Download the dataset manually from
https://storage.cloud.google.com/ipni-data/ipniWebName.csv.xz
(requires a Google account). Both compressed and uncompressed files
are accepted via the -f flag:

  harvester get ipni -f path/to/ipniWebName.csv.xz
  harvester get ipni -f path/to/ipniWebName.csv`,
		ManualSteps: true,
		URL:         "",
	}
	res := ipni{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (i *ipni) Extract(path string) error {
	gn.Info("Extracting IPNI data")

	if strings.HasSuffix(path, ".csv") {
		i.csvPath = filepath.Join(i.cfg.ExtractDir, filepath.Base(path))
		_, err := gnsys.CopyFile(path, i.csvPath)
		return err
	}

	// .xz — delegate to base (which calls gnsys.ExtractXz)
	if err := i.Convertor.Extract(path); err != nil {
		return err
	}
	// After extraction the file is named ipniWebName.csv in the extract dir.
	base := strings.TrimSuffix(filepath.Base(path), ".xz")
	i.csvPath = filepath.Join(i.cfg.ExtractDir, base)
	return nil
}
