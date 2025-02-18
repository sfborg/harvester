package sysio

import (
	"fmt"

	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/pkg/config"
)

func ResetCache(cfg config.Config) error {
	err := emptyCacheDir(cfg.CacheDir)
	if err != nil {
		return err
	}
	err = gnsys.MakeDir(cfg.DownloadDir)
	if err != nil {
		return err
	}

	err = gnsys.MakeDir(cfg.ExtractDir)
	if err != nil {
		return err
	}

	err = gnsys.MakeDir(cfg.SfgaDir)
	if err != nil {
		return err
	}

	return nil
}

func emptyCacheDir(cacheDir string) error {
	switch gnsys.GetDirState(cacheDir) {
	case gnsys.DirAbsent:
		return gnsys.MakeDir(cacheDir)
	case gnsys.DirEmpty:
		return nil
	case gnsys.DirNotEmpty:
		return gnsys.CleanDir(cacheDir)
	default:
		return fmt.Errorf("cannot empty CacheDir '%s", cacheDir)
	}
}
