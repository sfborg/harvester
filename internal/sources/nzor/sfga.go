package nzor

import (
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/sflib/pkg/sfga"
)

func (n *nzor) ToSfga(sfga sfga.Archive) error {
	n.sfga = sfga

	slog.Info("importing Meta")
	gn.Info("Importing Meta")
	if err := n.importMeta(); err != nil {
		return err
	}

	slog.Info("importing Name Usages")
	gn.Info("Importing Name Usages")
	if err := n.importNameUsages(); err != nil {
		return err
	}

	return nil
}
