package paleodb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gnames/gnfmt/gncsv"
	"github.com/gnames/gnfmt/gncsv/config"
	"github.com/sfborg/sflib/pkg/coldp"
)

func (p *paleodb) importTypeMaterials(types map[string][]string) error {
	var err error
	ch := make(chan []string)
	var wg sync.WaitGroup
	wg.Add(1)

	specPath := filepath.Join(p.cfg.ExtractDir, "spec.csv")
	cfg, err := config.New(config.OptPath(specPath))
	if err != nil {
		return err
	}
	csv := gncsv.New(cfg)

	go p.processType(csv, types, ch, &wg)

	_, err = csv.Read(context.Background(), ch)
	if err != nil {
		return err
	}
	close(ch)
	wg.Wait()
	return nil
}

func (p *paleodb) processType(
	csv gncsv.GnCSV,
	types map[string][]string,
	ch <-chan []string,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	var res []coldp.TypeMaterial
	var count int
	for v := range ch {
		count++
		if count%1_000 == 0 {
			fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
			fmt.Fprintf(os.Stderr, "\rProcessed %s lines", humanize.Comma(int64(count)))
		}
		specID := csv.F(v, "specimen_no")
		taxonIDs, ok := types[specID]
		if specID == "" && !ok {
			continue
		}
		for _, taxonID := range taxonIDs {
			tm := coldp.TypeMaterial{
				ID:              specID,
				NameID:          taxonID,
				ReferenceID:     csv.F(v, "reference_no"),
				Longitude:       coldp.ToFloat(csv.F(v, "lng")),
				Latitude:        coldp.ToFloat(csv.F(v, "lat")),
				Collector:       csv.F(v, "collectors"),
				Date:            csv.F(v, "collection_dates"),
				InstitutionCode: csv.F(v, "museum"),
			}
			res = append(res, tm)
		}
	}
	p.sfga.InsertTypeMaterials(res)
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 80))
}
