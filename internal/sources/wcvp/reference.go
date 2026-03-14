package wcvp

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sfborg/sflib/pkg/coldp"
)

func (w *wcvp) importReferences() error {
	f, err := os.Open(w.csvPath)
	if err != nil {
		return fmt.Errorf("opening WCVP csv for references: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '|'
	r.LazyQuotes = true

	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("reading WCVP csv headers: %w", err)
	}
	idx := buildIndex(headers)

	w.refMap = make(map[string]string)
	var refs []coldp.Reference
	counter := 0

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading WCVP csv: %w", err)
		}

		get := getter(row, idx)
		pub := get("place_of_publication")
		if pub == "" {
			continue
		}

		vol := get("volume_and_page")
		year := get("first_published")
		key := pub + "|" + vol + "|" + year

		if _, exists := w.refMap[key]; exists {
			continue
		}

		counter++
		id := fmt.Sprintf("sf_%d", counter)
		w.refMap[key] = id

		ref := coldp.Reference{
			ID:             id,
			ContainerTitle: pub,
			Page:           strings.TrimSpace(vol),
			Author:         get("publication_author"),
			Issued:         extractYear(year),
		}
		refs = append(refs, ref)

		if len(refs) >= w.cfg.BatchSize {
			if err := w.sfga.InsertReferences(refs); err != nil {
				return err
			}
			refs = refs[:0]
		}
	}

	if len(refs) > 0 {
		if err := w.sfga.InsertReferences(refs); err != nil {
			return err
		}
	}

	return nil
}
