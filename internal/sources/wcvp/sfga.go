package wcvp

import (
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/sflib/pkg/sfga"
)

func (w *wcvp) ToSfga(sfga sfga.Archive) error {
	w.sfga = sfga

	slog.Info("importing Meta")
	gn.Info("Importing Meta")
	if err := w.importMeta(); err != nil {
		return err
	}

	slog.Info("importing References")
	gn.Info("Importing References")
	if err := w.importReferences(); err != nil {
		return err
	}

	slog.Info("importing Name Usages")
	gn.Info("Importing Name Usages")
	if err := w.importNameUsages(); err != nil {
		return err
	}

	return nil
}
