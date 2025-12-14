package itis

import (
	"context"
	"log/slog"

	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga converts the ITIS data to SFGA format.
func (t *itis) ToSfga(arc sfga.Archive) error {
	var err error
	t.sfga = arc

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

	// Infer basionym relationships from authorship patterns.
	// ITIS doesn't have explicit basionym relationships, so we detect them
	// by matching stemmed epithets + authorship + year across names.
	slog.Info("Inferring basionym relationships")
	err = arc.InferBasionyms(context.Background(), sfga.BasionymInferenceConfig{
		SkipIfRelationsExist:       true, // skip if relations already exist
		CreateOriginalCombinations: true, // create OriginalGenus, OriginalSpecies, etc.
	})
	if err != nil {
		return err
	}

	return nil
}
