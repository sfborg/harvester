package clb

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gnames/gn"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib"
	sflibcfg "github.com/sfborg/sflib/config"
	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/sfborg/sflib/pkg/sfga"
	"golang.org/x/sync/errgroup"
)

// ColdpToSfga converts a COLDP zip archive to SFGA using sflib directly
// and returns the connected archive. Shared by all ChecklistBank-based sources.
func ColdpToSfga(coldpPath string, cfg config.Config) (sfga.Archive, error) {
	sfgaOutPath := filepath.Join(cfg.SfgaDir, "output")
	opts := sfgaOptions(cfg)

	arc := sflib.NewSfga(opts...)
	if err := arc.Create(cfg.SfgaDir); err != nil {
		return nil, fmt.Errorf("creating SFGA: %w", err)
	}
	if _, err := arc.Connect(); err != nil {
		return nil, fmt.Errorf("connecting to SFGA: %w", err)
	}

	slog.Info("fetching COLDP archive", "path", coldpPath)
	gn.Info("Converting COLDP to SFGA")

	cp := sflib.NewColdp(opts...)
	if err := cp.Fetch(coldpPath, cfg.ExtractDir); err != nil {
		return nil, fmt.Errorf("fetching COLDP: %w", err)
	}
	if err := cp.DirInfo(); err != nil {
		return nil, fmt.Errorf("reading COLDP layout: %w", err)
	}

	meta, err := cp.Meta()
	if err != nil {
		return nil, fmt.Errorf("reading COLDP meta: %w", err)
	}
	if err := arc.InsertMeta(meta); err != nil {
		return nil, fmt.Errorf("inserting meta: %w", err)
	}

	if err := importColdpData(cp, arc); err != nil {
		return nil, err
	}

	if err := arc.Close(); err != nil {
		return nil, fmt.Errorf("closing SFGA: %w", err)
	}

	outSqlite := sfgaOutPath + ".sqlite"
	schemaPath := filepath.Join(cfg.SfgaDir, "schema.sqlite")
	if err := os.Rename(schemaPath, outSqlite); err != nil {
		return nil, fmt.Errorf("finalizing SFGA: %w", err)
	}

	if cfg.WithZipOutput {
		if err := zipFile(outSqlite); err != nil {
			slog.Warn("failed to create SFGA zip", "error", err)
		}
	}

	arc.SetDb(outSqlite)
	if _, err := arc.Connect(); err != nil {
		return nil, fmt.Errorf("reconnecting to SFGA: %w", err)
	}

	slog.Info("COLDP to SFGA conversion complete")
	return arc, nil
}

func sfgaOptions(cfg config.Config) []sflibcfg.Option {
	var opts []sflibcfg.Option
	if cfg.LocalSchemaPath != "" {
		opts = append(opts, sflibcfg.OptLocalSchemaPath(cfg.LocalSchemaPath))
	}
	if cfg.Code != nomcode.Unknown {
		opts = append(opts, sflibcfg.OptNomCode(cfg.Code))
	}
	return opts
}

func importColdpData(cp coldp.Archive, arc sfga.Archive) error {
	cfg := cp.Config()
	paths := cp.DataPaths()

	if path, ok := paths[coldp.ReferenceDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertReferences); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.AuthorDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertAuthors); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.NameDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertNames); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.TaxonDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertTaxa); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.SynonymDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertSynonyms); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.NameUsageDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertNameUsages); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.VernacularNameDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertVernaculars); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.NameRelationDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertNameRelations); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.TypeMaterialDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertTypeMaterials); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.DistributionDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertDistributions); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.MediaDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertMedia); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.TreatmentDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertTreatments); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.SpeciesEstimateDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertSpeciesEstimates); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.TaxonPropertyDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertTaxonProperties); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.SpeciesInteractionDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertSpeciesInteractions); err != nil {
			return err
		}
	}
	if path, ok := paths[coldp.TaxonConceptRelationDT]; ok {
		if err := readAndInsert(cfg, path, arc.InsertTaxonConceptRelations); err != nil {
			return err
		}
	}

	return nil
}

// readAndInsert reads a COLDP CSV file and batch-inserts into SFGA.
// The insertion goroutine runs concurrently with the CSV read.
func readAndInsert[T coldp.DataLoader](
	cfg sflibcfg.Config,
	path string,
	insertFn func([]T) error,
) error {
	ch := make(chan T)
	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return batchInsert(ch, cfg.BatchSize, insertFn)
	})

	if err := coldp.Read(cfg, path, ch); err != nil {
		close(ch)
		g.Wait() //nolint:errcheck
		return fmt.Errorf("reading %s: %w", filepath.Base(path), err)
	}
	close(ch)

	return g.Wait()
}

func batchInsert[T any](
	ch <-chan T,
	batchSize int,
	insertFn func([]T) error,
) error {
	if batchSize <= 0 {
		batchSize = 50_000
	}
	batch := make([]T, 0, batchSize)
	for item := range ch {
		batch = append(batch, item)
		if len(batch) >= batchSize {
			if err := insertFn(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		return insertFn(batch)
	}
	return nil
}

func zipFile(path string) error {
	zipPath := path + ".zip"
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	hdr.Method = zip.Deflate

	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}
