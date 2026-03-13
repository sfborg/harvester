package ipni

import (
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/sflib/pkg/sfga"
)

func (i *ipni) ToSfga(sfga sfga.Archive) error {
	i.sfga = sfga

	slog.Info("importing Meta")
	gn.Info("Importing Meta")
	if err := i.importMeta(); err != nil {
		return err
	}

	slog.Info("importing Name Usages")
	gn.Info("Importing Name Usages")
	if err := i.importNameUsages(); err != nil {
		return err
	}

	return nil
}
