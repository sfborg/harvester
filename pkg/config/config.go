package config

import (
	"os"
	"path/filepath"

	"github.com/sfborg/sflib/ent/sfga"
)

var (
	// repoURL is the URL to the SFGA schema repository.
	repoURL = "https://github.com/sfborg/sfga"

	// tag of the sfga repo to get correct schema version.
	repoTag = "v0.3.24"

	// schemaHash is the sha256 sum of the correponding schema version.
	schemaHash = "b1db9df2e759f"

	jobsNum = 5
)

type Config struct {
	// GitRepo contains data for sfga schema Git repository.
	sfga.GitRepo

	// TempRepoDir is a temporary location to schema files downloaded from GitHub.
	TempRepoDir string

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

	// SkipDownload disables source data download and extraction. Useful during
	// development to save time and bandwidth.
	SkipDownload bool

	JobsNum int

	BatchSize int

	WithZipOutput bool
}

// Option is the type for all option functions available to modify
// configuration.
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

func OptSkipDownload(b bool) Option {
	return func(c *Config) {
		c.SkipDownload = b
	}
}

func New(opts ...Option) Config {
	tmpDir := os.TempDir()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = tmpDir
	}

	cacheDir = filepath.Join(cacheDir, "sfborg", "harvester")
	schemaRepo := filepath.Join(tmpDir, "sfborg", "sfga")

	res := Config{
		GitRepo: sfga.GitRepo{
			URL:          repoURL,
			Tag:          repoTag,
			ShaSchemaSQL: schemaHash,
		},
		TempRepoDir: schemaRepo,
		CacheDir:    cacheDir,
		JobsNum:     jobsNum,
		BatchSize:   50_000,
	}
	for _, opt := range opts {
		opt(&res)
	}

	res.DownloadDir = filepath.Join(res.CacheDir, "download")
	res.ExtractDir = filepath.Join(res.CacheDir, "extract")
	res.SfgaDir = filepath.Join(res.CacheDir, "sfga")
	return res
}
