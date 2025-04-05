package paleodb

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
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

	go p.processRef(csv, ch, &wg)

	_, err = csv.ReadChunks(context.Background(), ch, p.cfg.BatchSize)
	if err != nil {
		return err
	}
	return nil
}

func (p *paleodb) processRef(
	csv gncsv.GnCSV,
	ch <-chan [][]string,
	wg *sync.WaitGroup,
) error {
	var err error
	defer wg.Done()
	for rows := range ch {
		res := make([]coldp.Reference, 0, len(rows[0]))
		for _, r := range rows {
			ref := coldp.Reference{
				ID:        csv.F(r, "id"),
				Type:      coldp.NewReferenceType(csv.F(r, "type")),
				Author:    csv.F(r, "author"),
				Citation:  citation(csv, r),
				Title:     csv.F(r, "title"),
				Volume:    csv.F(r, "volume"),
				Issue:     csv.F(r, "number"),
				Page:      csv.F(r, "pages"),
				ISBN:      csv.F(r, "isbn"),
				Publisher: csv.F(r, "publisher"),
			}
			res = append(res, ref)
		}
		err = p.sfga.InsertReferences(res)
		if err != nil {
			return err
		}
	}
	return nil
}

func citation(csv gncsv.GnCSV, r []string) string {
	journal := csv.F(r, "journal")
	journalNumber := csv.F(r, "number")
	year := csv.F(r, "year")
	volume := csv.F(r, "volume")
	pages := csv.F(r, "pages")
	booktitle := csv.F(r, "booktitle")
	title := csv.F(r, "title")
	authors := csv.F(r, "author")
	publisher := csv.F(r, "publisher")

	var citation string

	if authors != "" {
		citation += authors + ". "
	}

	if year != "" {
		citation += fmt.Sprintf("(%s). ", year)
	}

	if title != "" {
		citation += fmt.Sprintf("%s. ", title)
	}

	if booktitle != "" {
		citation += fmt.Sprintf("In %s. ", booktitle)
	} else if journal != "" {
		citation += fmt.Sprintf("%s", journal)
		if volume != "" {
			citation += fmt.Sprintf(", %s", volume)
		}
		if journalNumber != "" {
			citation += fmt.Sprintf("(%s)", journalNumber)
		}
		if pages != "" {
			citation += fmt.Sprintf(", %s", pages)
		}
		citation += ". "
	}
	if publisher != "" && booktitle != "" {
		citation += fmt.Sprintf("%s. ", publisher)
	}

	return strings.TrimSpace(citation)
}
