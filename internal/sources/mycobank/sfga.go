package mycobank

import (
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/sflib/pkg/sfga"
)

func (m *mycobank) ToSfga(sfga sfga.Archive) error {
	var err error
	m.sfga = sfga

	slog.Info("importing Meta")
	gn.Info("Importing Meta")
	if err = m.importMeta(); err != nil {
		return err
	}

	slog.Info("importing Name Usages")
	gn.Info("Importing Name Usages")
	if err = m.importNameUsages(); err != nil {
		return err
	}

	return nil
}
