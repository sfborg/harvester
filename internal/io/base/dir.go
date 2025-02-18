package base

import (
	"fmt"

	"github.com/gnames/gnsys"
)

func (c *Convertor) ResetCache() error {
	err := emptyCacheDir(c.cfg.CacheDir)
	if err != nil {
		return err
	}
	err = gnsys.MakeDir(c.cfg.DownloadDir)
	if err != nil {
		return err
	}

	err = gnsys.MakeDir(c.cfg.ExtractDir)
	if err != nil {
		return err
	}

	err = gnsys.MakeDir(c.cfg.SfgaDir)
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
		err := fmt.Errorf("cannot emptry '%s' dir", cacheDir)
		return err
	}
}
