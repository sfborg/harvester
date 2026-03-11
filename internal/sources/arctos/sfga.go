package arctos

import (
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/sflib/pkg/sfga"
)

func (a *arctos) ToSfga(sfga sfga.Archive) error {
	var err error
	a.sfga = sfga

	slog.Info("importing Meta")
	gn.Info("Importing Meta")
	if err = a.importMeta(); err != nil {
		return err
	}

	slog.Info("importing Name Usages")
	gn.Info("Importing Name Usages")
	if err = a.importNameUsages(); err != nil {
		return err
	}

	return nil
}
