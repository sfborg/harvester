package sysio

import (
	"fmt"
	"log/slog"

	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/pkg/config"
)

func ResetCache(cfg config.Config) error {
	slog.Info("Reseting cache")
	err := EmptyDir(cfg.CacheDir)
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

func EmptyDir(cacheDir string) error {
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
