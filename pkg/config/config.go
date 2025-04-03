package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnparser/ent/nomcode"
)

var (
	jobsNum = 5
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

	// LoadFile can be a local file or URl. In cases when there is no
	// stable link to a source's data the LoadFile is used.
	LoadFile string

	// Code provides nomenclatural code setting to use in GNparser.
	// This flag is only important for importing data from text, csv and
	// other ad-hoc files.
	Code nomcode.Code

	// ArchiveDate is used to set 'Issued' field in CoLDP/SFGA metadata.
	ArchiveDate string

	// ArchiveDate is used to set 'Version' firle for CoLDP/SFGA metadata.
	ArchiveVersion string

	// WithVerbose indicates that more information might be shown in the
	// output information. It is only important for listing a short list of
	// supported sources, or providing details about them.
	WithVerbose bool

	// SkipDownload disables source data download and extraction. Useful during
	// development to save time and bandwidth.
	SkipDownload bool

	// ColSep is used when importing CSV/TSV/PSV files. By default it is empty
	// and is determined automatically.
	ColSep string

	// WithoutQuotes can be used to parse correctly tab- or pipe-delimited
	// files where fields never escaped by quotes.
	WithoutQuotes bool

	// BadRow sets how to process rows with wrong number of fields in CSV
	// files. By default it is set to process such rows. Other options are
	// to return an error, or skip them.
	BadRow gnfmt.BadRow

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
		c.LoadFile = s
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

func OptArchiveDate(s string) Option {
	return func(c *Config) {
		c.ArchiveDate = s
	}
}

func OptArchiveVersion(s string) Option {
	return func(c *Config) {
		c.ArchiveVersion = s
	}
}

func OptWithoutQuotes(b bool) Option {
	return func(c *Config) {
		c.WithoutQuotes = b
	}
}

func OptColSep(s string) Option {
	return func(c *Config) {
		c.ColSep = s
	}
}

func OptBadRow(br gnfmt.BadRow) Option {
	return func(c *Config) {
		c.BadRow = br
	}
}

func New(opts ...Option) Config {
	tmpDir := os.TempDir()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = tmpDir
	}
	cacheDir = filepath.Join(cacheDir, "sfborg", "harvester")

	currentTime := time.Now()
	today := currentTime.Format("2006-01-02")

	res := Config{
		CacheDir:    cacheDir,
		JobsNum:     jobsNum,
		Code:        nomcode.Unknown,
		BadRow:      gnfmt.ProcessBadRow,
		BatchSize:   50_000,
		ArchiveDate: today,
	}
	for _, opt := range opts {
		opt(&res)
	}

	res.DownloadDir = filepath.Join(res.CacheDir, "download")
	res.ExtractDir = filepath.Join(res.CacheDir, "extract")
	res.SfgaDir = filepath.Join(res.CacheDir, "sfga")

	return res
}
