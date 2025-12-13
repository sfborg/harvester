package itis

import (
	"log/slog"

	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga converts the ITIS data to SFGA format.
func (t *itis) ToSfga(sfga sfga.Archive) error {
	var err error
	t.sfga = sfga

	slog.Info("Importing Meta")
	err = t.importMeta()
	if err != nil {
		return err
	}

	slog.Info("Importing References")
	err = t.importReferences()
	if err != nil {
		return err
	}

	slog.Info("Importing Name Usages")
	err = t.importNameUsages()
	if err != nil {
		return err
	}

	slog.Info("Importing Synonyms")
	err = t.importSynonyms()
	if err != nil {
		return err
	}

	slog.Info("Importing Vernacular Names")
	err = t.importVernaculars()
	if err != nil {
		return err
	}

	slog.Info("Importing Distributions")
	err = t.importDistributions()
	if err != nil {
		return err
	}

	return nil
}
