package paleodb

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/gnames/gnfmt/gncsv"
	"github.com/gnames/gnfmt/gncsv/config"
	"github.com/sfborg/sflib/pkg/coldp"
)

func (p *paleodb) importReferences() error {
	taxonPath := filepath.Join(p.cfg.ExtractDir, "ref.csv")
	cfg, err := config.New(config.OptPath(taxonPath))
	if err != nil {
		return err
	}
	csv := gncsv.New(cfg)

	ch := make(chan [][]string)
	var wg sync.WaitGroup
	wg.Add(1)

	go p.processRef(ch, &wg)

	csv := gncsv.New(cfg)
	ch := make(chan [][]string)
	_, err = csv.ReadChunks(context.Background(), ch, p.cfg.BatchSize)
	if err != nil {
		return err
	}
	return nil
}

func (p *paleodb) processRef(ch <-chan [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, r := range rows {
		refs := make([]coldp.Reference, 0, len(rows[0]))

	}
}
