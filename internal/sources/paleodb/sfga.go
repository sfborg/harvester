package paleodb

import (
	"log/slog"

	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga imports the ION archive into a sfga archive.
func (p *paleodb) ToSfga(sfga sfga.Archive) error {
	var err error
	p.sfga = sfga

	slog.Info("Importing Meta")
	err = p.importMeta()
	if err != nil {
		return err
	}

	slog.Info("Importing Names Usages")
	err = p.importNameUsages()
	if err != nil {
		return err
	}

	slog.Info("Importing Refernces")
	err = p.importReferences()
	if err != nil {
		return err
	}

	return nil
}
