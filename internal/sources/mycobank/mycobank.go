package mycobank

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gnames/gn"
	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
)

type mycobank struct {
	data.Convertor
	cfg      config.Config
	sfga     sfga.Archive
	xlsxPath string
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "mycobank",
		Name:  "MycoBank",
		Notes: `MycoBank is an online database documenting mycological
nomenclatural novelties. Data is downloaded automatically from
https://www.mycobank.org/images/MBList.zip.

Use the -f flag to provide a local copy of MBList.zip
or MBList.xlsx instead.`,
		ManualSteps: false,
		URL:         "https://www.mycobank.org/images/MBList.zip",
	}
	res := mycobank{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

// Extract handles both a downloaded zip (default) and a user-supplied xlsx
// file provided via the -f flag.
func (m *mycobank) Extract(path string) error {
	if path == "" {
		slog.Info("skip extraction (using cached files)")
		return nil
	}

	if strings.HasSuffix(strings.ToLower(path), ".xlsx") {
		slog.Info("copying MycoBank xlsx file")
		gn.Info("Copying MycoBank xlsx file")
		dest := filepath.Join(m.cfg.ExtractDir, filepath.Base(path))
		if _, err := gnsys.CopyFile(path, dest); err != nil {
			return err
		}
		m.xlsxPath = dest
		return nil
	}

	// Assume zip; delegate to base extractor.
	if err := m.Convertor.Extract(path); err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(m.cfg.ExtractDir, "*.xlsx"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no xlsx file found in %s after extraction", m.cfg.ExtractDir)
	}
	m.xlsxPath = matches[0]
	return nil
}
