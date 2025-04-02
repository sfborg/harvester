package ion

import (
	"log/slog"

	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga imports the ION archive into a sfga archive.
func (i *ion) ToSfga(sfga sfga.Archive) error {
	var err error
	i.sfga = sfga

	slog.Info("Importing Meta")
	err = i.importMeta()
	if err != nil {
		return err
	}

	slog.Info("Importing Names")
	err = i.importNames()
	if err != nil {
		return err
	}

	return nil
}
