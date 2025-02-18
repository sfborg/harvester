package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	// CacheDir is the directory where all temporary files are located.
	CacheDir string

	// DownloadDir contains temporary files for download.
	DownloadDir string

	// ExtractDir contains temporary files where original archive is extracted to.
	ExtractDir string

	// SfgaDir contains files of the built SFGA file.
	SfgaDir string

	// WithVerbose indicates that more information might be shown in the
	// output information.
	WithVerbose bool
}

type Option func(*Config)

func OptCacheDir(s string) Option {
	return func(c *Config) {
		c.CacheDir = s
	}
}

func OptWithVerbose(b bool) Option {
	return func(c *Config) {
		c.WithVerbose = b
	}
}

func New(opts ...Option) Config {
	tmpDir := os.TempDir()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = tmpDir
	}

	cacheDir = filepath.Join(cacheDir, "sfborg", "harvester")

	res := Config{
		CacheDir: cacheDir,
	}
	for _, opt := range opts {
		opt(&res)
	}

	res.DownloadDir = filepath.Join(res.CacheDir, "download")
	res.ExtractDir = filepath.Join(res.CacheDir, "extract")
	res.SfgaDir = filepath.Join(res.CacheDir, "sfga")
	return res
}
