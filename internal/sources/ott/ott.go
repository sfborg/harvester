package ott

import (
	"path/filepath"

	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type ott struct {
	data.Convertor
	cfg          config.Config
	sfga         sfga.Archive
	taxonomyPath string
	synonymsPath string
	synonyms     map[string][]synDatum
}

type synDatum struct {
	name    string
	synType string
}

type datum struct {
	taxonID  string
	parentID string
	name     string
	rank     string
	flags    string
	synonyms []synDatum
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "ott",
		Name:  "Open Tree Taxonomy",
		Notes: `OTT is the Open Tree Taxonomy, a comprehensive classification of
all life assembled from NCBI, GBIF, WoRMS, and other sources.

Data must be manually downloaded from:
https://tree.opentreeoflife.org/about/taxonomy-version

Download the tgz archive and load it via the --file flag.

Example:
  harvester get ott --file ~/Downloads/ott3.7.3.tgz`,
		ManualSteps: true,
	}
	res := ott{
		cfg:          cfg,
		Convertor:    base.New(cfg, &set),
		taxonomyPath: filepath.Join(cfg.ExtractDir, "ott3.7.3", "taxonomy.tsv"),
		synonymsPath: filepath.Join(cfg.ExtractDir, "ott3.7.3", "synonyms.tsv"),
		synonyms:     make(map[string][]synDatum),
	}
	return &res
}
