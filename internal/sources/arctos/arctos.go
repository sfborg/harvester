package arctos

import (
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type arctos struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "arctos",
		Name:  "Arctos",
		Notes: `Arctos is an ongoing effort to integrate access to specimen
data, collection-management tools, and external resources.

The source file is large (~1GB). It is often preferable to download
it separately and provide it with the -f flag:

  wget https://arctos.database.museum/cache/gn_merge.tgz
  harvester get arctos -f path/to/gn_merge.tgz`,
		URL: "https://arctos.database.museum/cache/gn_merge.tgz",
	}
	res := arctos{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}
