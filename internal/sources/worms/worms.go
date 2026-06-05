package worms

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/gn"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib"
	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/sfborg/sflib/pkg/dwca"
	"github.com/sfborg/sflib/pkg/sfga"
	"golang.org/x/sync/errgroup"
)

type worms struct {
	data.Convertor
	cfg     config.Config
	sfga    sfga.Archive
	dwca    dwca.Archive
	zipPath string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "worms",
		Name:  "World Register of Marine Species",
		Notes: `WoRMS is the World Register of Marine Species, an authoritative
classification and catalogue of marine organisms.

Data must be manually downloaded. To request a full copy go to
https://www.marinespecies.org/usersrequest.php.
Load DwCA zip file via the --load-file flag.

Example:
  harvester worms --load-file ~/Downloads/WoRMS_DwC-A.zip`,
		ManualSteps: true,
	}
	res := worms{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

// Extract stores the zip path; dwca.Fetch in ToSfga handles the actual extraction.
func (w *worms) Extract(path string) error {
	w.zipPath = path
	return nil
}

func (w *worms) InitSfga() (sfga.Archive, error) {
	arc, err := w.Convertor.InitSfga()
	if err != nil {
		return nil, err
	}
	w.sfga = arc
	return arc, nil
}

func (w *worms) ToSfga(arc sfga.Archive) error {
	w.sfga = arc
	w.dwca = sflib.NewDwca()

	gn.Info("Fetching WoRMS DwCA data")
	if err := w.dwca.Fetch(w.zipPath, w.cfg.ExtractDir); err != nil {
		return fmt.Errorf("worms: fetching DwCA: %w", err)
	}

	if err := w.importMeta(); err != nil {
		return err
	}

	if err := w.importCore(); err != nil {
		return err
	}

	return w.importExtensions()
}

func (w *worms) importCore() error {
	slog.Info("Importing WoRMS core data")
	ch := make(chan coldp.Data)
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		defer close(ch)
		return w.dwca.LoadCore(ctx, ch)
	})

	g.Go(func() error {
		return w.writeCoreData(ctx, ch)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (w *worms) writeCoreData(ctx context.Context, ch <-chan coldp.Data) error {
	var rows int
	ids := make(map[string]struct{})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-ch:
			if !ok {
				fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 50))
				fmt.Fprintf(os.Stderr, "\rSaved %s core records\n", humanize.Comma(int64(rows)))
				return nil
			}
			if len(d.NameUsages) > 0 {
				dedup := make([]coldp.NameUsage, 0, len(d.NameUsages))
				rows += len(d.NameUsages)
				fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 50))
				fmt.Fprintf(os.Stderr, "\rSaved %s records", humanize.Comma(int64(rows)))
				for _, nu := range d.NameUsages {
					if _, ok := ids[nu.ID]; ok {
						slog.Error("Duplicated TaxonID, skipping", "id", nu.ID)
						continue
					}
					ids[nu.ID] = struct{}{}
					nu.AlternativeID = aphiaGnoutlink(nu.ID)
					dedup = append(dedup, nu)
				}
				if err := w.sfga.InsertNameUsages(dedup); err != nil {
					return err
				}
			}
			if len(d.References) > 0 {
				if err := w.sfga.InsertReferences(d.References); err != nil {
					return err
				}
			}
		}
	}
}

func (w *worms) importExtensions() error {
	for i, ext := range w.dwca.Meta().Extensions {
		if ext == nil {
			continue
		}
		rowType := strings.ToLower(filepath.Base(ext.RowType))
		switch {
		case strings.Contains(rowType, "vernacular"):
			if err := w.importVernacular(i); err != nil {
				return err
			}
		case strings.Contains(rowType, "distribution"):
			if err := w.importDistribution(i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *worms) importVernacular(idx int) error {
	slog.Info("Importing WoRMS vernacular names")
	ch := make(chan []coldp.Vernacular)
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		defer close(ch)
		return w.dwca.LoadVernacular(ctx, idx, ch)
	})

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case v, ok := <-ch:
				if !ok {
					return nil
				}
				if err := w.sfga.InsertVernaculars(v); err != nil {
					return err
				}
			}
		}
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (w *worms) importDistribution(idx int) error {
	slog.Info("Importing WoRMS distribution data")
	ch := make(chan coldp.Data)
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		defer close(ch)
		return w.dwca.LoadDistribution(ctx, idx, ch)
	})

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case d, ok := <-ch:
				if !ok {
					return nil
				}
				if err := w.sfga.InsertDistributions(d.Distributions); err != nil {
					return err
				}
				if len(d.References) > 0 {
					if err := w.sfga.InsertReferences(d.References); err != nil {
						return err
					}
				}
			}
		}
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

// aphiaGnoutlink extracts the numeric AphiaID from a WoRMS LSID and formats
// it as a gnoutlink alternative ID. For example:
// "urn:lsid:marinespecies.org:taxname:12345" → "gnoutlink:12345".
func aphiaGnoutlink(taxonID string) string {
	idx := strings.LastIndex(taxonID, ":")
	if idx < 0 || idx == len(taxonID)-1 {
		return ""
	}
	return "gnoutlink:" + taxonID[idx+1:]
}
