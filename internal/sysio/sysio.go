package sysio

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gnames/gn"
	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/errcode"
)

func EnsureLogDir(homeDir string) error {
	dir := config.LogDir(homeDir)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return &gn.Error{
			Code: errcode.CreateDirError,
			Msg:  "Cannot create directory %s",
			Vars: []any{dir},
			Err:  fmt.Errorf("cannot make %s: %w", dir, err),
		}
	}
	return nil
}

// LogFile return LogFile or error.
func LogFile(homeDir string) (*os.File, error) {
	err := EnsureLogDir(homeDir)
	if err != nil {
		return nil, err
	}
	path := config.LogPath(homeDir)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		err = &gn.Error{
			Code: errcode.OpenFileError,
			Msg:  "Cannot open %s",
			Vars: []any{path},
			Err:  fmt.Errorf("cannot open %s: %w", path, err),
		}
		return nil, err
	}
	return file, nil
}

func ResetCache(cfg config.Config) error {
	slog.Info("reset cache")
	gn.Message("Resetting cache")
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
