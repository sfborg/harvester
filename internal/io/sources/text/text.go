package text

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/harvester/internal/io/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/ent/sfga"
)

type text struct {
	data.Convertor
	cfg      config.Config
	filePath string
	sfga     sfga.Archive
}

func New(cfg config.Config) data.Convertor {
	set := data.Set{
		Label: "text",
		Name:  "Name list",
		Description: `
Imports UTF8-encoded file with one name per line.
The file with names has to be provided with --file option.`,
		ManualSteps: true,
	}
	res := text{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (t *text) Import(src string) error {
	isText, err := gnsys.IsTextFile(src)
	if err != nil {
		return err
	}
	if !isText {
		err = fmt.Errorf("not a text file '%s'", filepath.Base(src))
		return err
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dst := t.cfg.ExtractDir
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

	t.filePath = dstPath

	return nil
}

// ToSFGA imports the ION archive into a sfga archive.
func (t *text) ToSFGA(sfga sfga.Archive) error {
	var err error
	t.sfga = sfga

	slog.Info("Importing Names")
	err = t.importNames()
	if err != nil {
		return err
	}

	return nil
}
