package config

import (
	"os"
	"path/filepath"

	"github.com/gnames/gnparser/ent/nomcode"
)

var (
	jobsNum = 5
)

type Config struct {
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

	// LocalFile if set to a local file path, this file will be used as a
	// source data instead of a download from internet.
	LocalFile string

	// Code provides nomenclatural code setting to use in GNparser.
	Code nomcode.Code

	// WithVerbose indicates that more information might be shown in the
	// output information.
	WithVerbose bool

	// SkipDownload disables source data download and extraction. Useful during
	// development to save time and bandwidth.
	SkipDownload bool

	// JobsNum sets the number of concurrent jobs to set, if it is
	// needed.
	JobsNum int

	// BatchSize determines the size of slices to import into SFGA.
	BatchSize int

	// WithZipOutput indicates that zipped archives have to be created.
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

func OptLocalFile(s string) Option {
	return func(c *Config) {
		c.LocalFile = s
	}
}

func OptWithZipOutput(b bool) Option {
	return func(c *Config) {
		c.WithZipOutput = b
	}
}

func OptSkipDownload(b bool) Option {
	return func(c *Config) {
		c.SkipDownload = b
	}
}

func OptCode(code nomcode.Code) Option {
	return func(c *Config) {
		c.Code = code
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
		TempRepoDir: schemaRepo,
		CacheDir:    cacheDir,
		JobsNum:     jobsNum,
		Code:        nomcode.Unknown,
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
