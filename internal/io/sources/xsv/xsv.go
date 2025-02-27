package xsv

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/ent/sfga"
)

type xsv struct {
	data.Convertor
	cfg      config.Config
	sfga     sfga.Archive
	filePath string
}

func New(cfg config.Config) data.Convertor {
	set := data.Set{
		Label: "xsv",
		Name:  "CSV, TSV, PSV files",
		Description: `
It requires a CSV (comma-delimited), TSV (tab-delimited) or PSV (pipe-delimited)
file provided with --file option. The file must have headers that
use DarwinCore terms.
`,
		ManualSteps: true,
	}
	res := xsv{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (x *xsv) Import(src string) error {
	// isText, err := gnsys.IsTextFile(src)
	// if err != nil {
	// 	return err
	// }
	// if !isText {
	// 	err = fmt.Errorf("not a text file '%s'", filepath.Base(src))
	// 	return err
	// }
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dst := x.cfg.ExtractDir
	dstPath := filepath.Join(dst, filepath.Base(src))

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	x.filePath = dstPath

	return nil
}

// ToSFGA imports the ION archive into a sfga archive.
func (x *xsv) ToSFGA(sfga sfga.Archive) error {
	var err error
	x.sfga = sfga

	slog.Info("Importing Names")
	err = x.importNames()
	if err != nil {
		return err
	}

	return nil
}
