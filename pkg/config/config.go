package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	CacheDir    string
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
	return res
}
