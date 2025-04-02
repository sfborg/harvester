package ion

import (
	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type ion struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "ion",
		Name:  "Index to Organism Names",
		Notes: `Download cached version of the file from box.com.
Ask Rod Page for an update.`,
		ManualSteps: true,
		URL:         "https://uofi.box.com/shared/static/tklh8i6q2kb33g6ki33k6s3is06lo9np.gz",
	}
	res := ion{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (i *ion) Import(path string) error {
	err := gnsys.ExtractTarGz(path, i.cfg.ExtractDir)
	if err != nil {
		return err
	}
	return nil
}
