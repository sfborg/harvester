package ncbi

import (
	"path/filepath"

	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type ncbi struct {
	data.Convertor
	cfg                config.Config
	sfga               sfga.Archive
	namePath, nodePath string
	names              map[string]map[string]string
	data               []datum
}

type synonym struct {
	name            string
	taxonomicStatus string
}

type datum struct {
	taxonID   string
	parentID  string
	canonical string
	nameStr   string
	rank      string
	vernNames []string
	synonyms  []synonym
}

var vernType = []string{"common name", "genbank common name"}
var nameType = []string{"valid", "authority"}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label:       "ncbi",
		Name:        "National Center for Biotechnology Information",
		ManualSteps: false,
		URL:         "https://ftp.ncbi.nlm.nih.gov/pub/taxonomy/taxdump.tar.gz",
	}
	res := ncbi{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
		namePath:  filepath.Join(cfg.ExtractDir, "names.dmp"),
		nodePath:  filepath.Join(cfg.ExtractDir, "nodes.dmp"),
		names:     make(map[string]map[string]string),
	}
	return &res
}
