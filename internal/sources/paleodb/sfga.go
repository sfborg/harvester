package paleodb

import (
	"log/slog"

	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga imports the ION archive into a sfga archive.
func (p *paleodb) ToSfga(sfga sfga.Archive) error {
	var err error
	var citations map[string]string
	var types map[string][]string
	p.sfga = sfga

	slog.Info("Importing Meta")
	err = p.importMeta()
	if err != nil {
		return err
	}

	slog.Info("Importing Names Usages")
	citations, types, err = p.importNameUsages()
	if err != nil {
		return err
	}

	slog.Info("Importing Refernces")
	err = p.importReferences(citations)
	if err != nil {
		return err
	}

	slog.Info("Importing Type Materials")
	err = p.importTypeMaterials(types)
	if err != nil {
		return err
	}

	return nil
}
