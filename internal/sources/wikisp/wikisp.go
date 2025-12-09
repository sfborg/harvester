package wikisp

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnames/gn"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/internal/sources/wikisp/wsparser"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type wikisp struct {
	data.Convertor
	cfg        config.Config
	sfga       sfga.Archive
	gnp        gnparser.GNparser
	wsp        *wsparser.WSParser
	stats      *parseStats
	storage    *tempStorage
	synonymMap map[string]*synonym
	taxonPages []*PageData
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label:       "wikispecies",
		Name:        "Wikispecies",
		ManualSteps: false,
		URL: "https://dumps.wikimedia.org/specieswiki/latest/" +
			"specieswiki-latest-pages-articles.xml.bz2",
	}

	gnp := gnparser.New(gnparser.NewConfig(
		gnparser.OptWithDetails(true),
		gnparser.OptCode(nomcode.Unknown),
	))
	res := wikisp{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
		storage: &tempStorage{
			redirects:   make(map[string]string),
			templateIDs: make(map[string]string),
			taxonIDs:    make(map[string]string),
		},
		stats: &parseStats{
			MissingParents:         make(map[string][]string),
			MissingRedirectTargets: make(map[string][]string),
		},
		gnp:        gnp,
		wsp:        wsparser.New(&gnparserAdapter{gnp: gnp}),
		synonymMap: make(map[string]*synonym),
	}
	return &res
}

// Extract handles both compressed archives and plain XML files.
func (w *wikisp) Extract(path string) error {
	// If it's already XML, just copy it to extract directory
	if strings.HasSuffix(strings.ToLower(path), ".xml") {
		slog.Info("plain XML file, copying to extract directory", "path", path)
		gn.Info("Copy XML file to extract directory")

		// Create extract directory
		if err := os.MkdirAll(w.cfg.ExtractDir, 0755); err != nil {
			return err
		}

		// Copy file
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dstPath := filepath.Join(w.cfg.ExtractDir, filepath.Base(path))
		dst, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		slog.Info("copied XML file", "destination", dstPath)
		gn.Info("XML file copied to <em>%s</em>", dstPath)
		return nil
	}

	// Otherwise use base implementation for compressed files
	return w.Convertor.Extract(path)
}
