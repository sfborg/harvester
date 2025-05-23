package ioc

import (
	"log/slog"

	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga imports the IOC List into SFGA archive.
func (l *ioc) ToSfga(sfga sfga.Archive) error {
	var err error
	l.sfga = sfga

	slog.Info("Importing Meta")
	err = l.importMeta()
	if err != nil {
		return err
	}

	slog.Info("Importing Names")
	err = l.importNameUsages()
	if err != nil {
		return err
	}

	// slog.Info("Importing vernaculars")
	// err = l.importVern()
	// if err != nil {
	// 	return err
	// }

	return nil
}
