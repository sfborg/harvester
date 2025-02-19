package itis

import (
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sfborg/from-coldp/pkg/ent/sfgarc"
	"github.com/sfborg/from-coldp/pkg/io/sfgarcio"
	"github.com/sfborg/harvester/internal/io/sysio"
	"github.com/sfborg/sflib/io/dbio"
	"github.com/sfborg/sflib/io/schemaio"
	_ "modernc.org/sqlite"
)

func (i *itis) ToSFGA(sfga sfgarc.Archive) error {
	var err error
	i.sfga, err = i.initSFGA()
	if err != nil {
		return err
	}

	i.itisDb, err = i.itisConnect()
	if err != nil {
		return err
	}

	defer i.sfga.Close()
	if err != nil {
		return err
	}
	slog.Info("Importing Meta")
	i.importMeta()
	return nil
}

func (i *itis) initSFGA() (sfgarc.Archive, error) {
	sysio.EmptyDir(i.cfg.SfgaDir)
	coldpCfg := i.cfg.ToColdpConfig()

	sfgaSchema := schemaio.New(i.cfg.GitRepo, i.cfg.TempRepoDir)
	sfgaDB := dbio.New(i.cfg.SfgaDir)

	sfarc := sfgarcio.New(coldpCfg, sfgaSchema, sfgaDB)
	err := sfarc.Connect()
	if err != nil {
		return nil, err
	}
	return sfarc, nil
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
