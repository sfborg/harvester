package grin

import (
	"log/slog"

	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga imports the ION archive into a sfga archive.
func (g *grin) ToSfga(sfga sfga.Archive) error {
	var err error
	g.sfga = sfga

	slog.Info("Importing Meta")
	err = g.importMeta()
	if err != nil {
		return err
	}

	slog.Info("Importing Names")
	err = g.importNameUsages()
	if err != nil {
		return err
	}

	slog.Info("Importing vernaculars")
	err = g.importVern()
	if err != nil {
		return err
	}

	return nil
}
