package itis

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sfborg/sflib/ent/sfga"
	_ "modernc.org/sqlite"
)

func (i *itis) ToSFGA(sfga sfga.Archive) error {
	var err error
	var itisDb *sql.DB
	i.sfga = sfga

	itisDb, err = i.itisConnect()
	if err != nil {
		return err
	}
	i.itisDb = itisDb
	defer i.itisDb.Close()
	defer i.sfga.Close()

	slog.Info("Importing Meta")
	err = i.importMeta()
	if err != nil {
		return err
	}

	slog.Info("Importing NameUsage")
	err = i.importNameUsage()
	if err != nil {
		return err
	}

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
		i.cfg.ExtractDir,
		func(path string, info os.FileInfo, err error,
		) error {
			if err != nil {
				return err // Handle errors, but keep walking
			}
			if info.Name() == "ITIS.sqlite" {
				itisPath = path
				return filepath.SkipAll // Found it, no need to continue in this dir
			}
			return nil // Continue walking
		})
	if err != nil && err != filepath.SkipAll {
		return "", err // Return the error if it wasn't a SkipDir
	}
	if itisPath == "" {
		return "", errors.New("cannot find ITIS.sqlite")
	}
	return itisPath, nil
}
