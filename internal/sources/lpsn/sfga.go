package lpsn

import (
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga imports the LPSN CSV into a SFGA archive.
func (l *lpsn) ToSfga(sfga sfga.Archive) error {
	var err error
	l.sfga = sfga

	slog.Info("importing Meta")
	gn.Info("Importing Meta")
	if err = l.importMeta(); err != nil {
		return err
	}

	slog.Info("importing Name Usages")
	gn.Info("Importing Name Usages")
	if err = l.importNameUsages(); err != nil {
		return err
	}

	return nil
}
