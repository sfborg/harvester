package itis

import (
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sfborg/sflib/ent/sfga"
	_ "modernc.org/sqlite"
)

func (i *itis) ToSFGA(sfga sfga.Archive) error {
	var err error
	i.sfga = sfga

	i.itisDb, err = i.itisConnect()
	if err != nil {
		return err
	}

	defer i.sfga.Close()

	slog.Info("Importing Meta")
	i.importMeta()
	return nil
}

func (i *itis) itisConnect() (*sql.DB, error) {
	var err error
	var db *sql.DB
	itisPath, err := i.getItisFile()
	if err != nil {
		return nil, err
	}

	db, err = sql.Open("sqlite", itisPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (i *itis) getItisFile() (string, error) {
	var itisPath string
	err := filepath.Walk(
		i.cfg.SfgaDir,
		func(path string, info os.FileInfo, err error,
		) error {
			if err != nil {
				return err // Handle errors, but keep walking
			}
			if info.Name() == "ITIS.sqlite" {
				itisPath = path
				return filepath.SkipDir // Found it, no need to continue in this dir
			}
			return nil // Continue walking
		})
	if err != nil && err != filepath.SkipDir {
		return "", err // Return the error if it wasn't a SkipDir
	}
	return itisPath, nil
}
