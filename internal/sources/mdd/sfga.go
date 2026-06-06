package mdd

import (
	"log/slog"

	"github.com/gnames/gn"
	"github.com/sfborg/sflib/pkg/sfga"
)

func (m *mdd) ToSfga(arc sfga.Archive) error {
	m.sfga = arc

	slog.Info("importing MDD meta")
	gn.Info("Importing MDD meta")
	if err := m.importMeta(); err != nil {
		return err
	}

	slog.Info("importing MDD species")
	gn.Info("Importing MDD species")
	sciNameToID, err := m.importSpecies()
	if err != nil {
		return err
	}

	slog.Info("importing MDD synonyms and subspecies")
	gn.Info("Importing MDD synonyms and subspecies")
	return m.importSynonyms(sciNameToID)
}
