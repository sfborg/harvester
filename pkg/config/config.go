package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/nomcode"
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

	// LocalSchemaPath is the path to a local schema.sql file to use
	// instead of fetching from GitHub. Useful for development.
	LocalSchemaPath string

	// CLBApi is the ChecklistBank API base URL.
	CLBApi string

	// CLBUser is the username for ChecklistBank authentication.
	CLBUser string

	// CLBPassword is the password for ChecklistBank authentication.
	CLBPassword string

	// CLBDatasetID is the ChecklistBank dataset identifier.
	CLBDatasetID int

	// CLBDatasetAlias is used as the output file base name. If empty,
	// the alias is fetched from the CLB dataset metadata or derived
	// from the dataset title.
	CLBDatasetAlias string

	// CLBTaxonID is the root taxon ID for filtering CLB exports.
	CLBTaxonID string

	// CLBFormat is the CLB export format (e.g. ColDP, DWCA).
	CLBFormat string

	// CLBSynonyms includes synonyms in the CLB export.
	CLBSynonyms bool

	// CLBBareNames includes bare names in the CLB export.
	CLBBareNames bool

	// CLBExtended requests extended data in the CLB export.
	CLBExtended bool

	// CLBExtinct filters by extinction status. nil means unset.
	CLBExtinct *bool

	// CLBClassification includes classification in the CLB export.
	CLBClassification bool

	// CLBTaxGroups includes taxonomic groups in the CLB export.
	CLBTaxGroups bool

	// CLBMinRank sets the minimum taxonomic rank for the CLB export.
	CLBMinRank string

	// CLBTabFormat sets the tabular format (CSV or TSV) for CLB export.
	CLBTabFormat string
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

func OptLocalSchemaPath(s string) Option {
	return func(c *Config) {
		c.LocalSchemaPath = s
	}
}

func OptCLBApi(s string) Option {
	return func(c *Config) {
		c.CLBApi = s
	}
}

func OptCLBUser(s string) Option {
	return func(c *Config) {
		c.CLBUser = s
	}
}

func OptCLBPassword(s string) Option {
	return func(c *Config) {
		c.CLBPassword = s
	}
}

func OptCLBDatasetID(i int) Option {
	return func(c *Config) {
		c.CLBDatasetID = i
	}
}

func OptCLBDatasetAlias(s string) Option {
	return func(c *Config) {
		c.CLBDatasetAlias = s
	}
}

func OptCLBTaxonID(s string) Option {
	return func(c *Config) {
		c.CLBTaxonID = s
	}
}

func OptCLBFormat(s string) Option {
	return func(c *Config) {
		c.CLBFormat = s
	}
}

func OptCLBSynonyms(b bool) Option {
	return func(c *Config) {
		c.CLBSynonyms = b
	}
}

func OptCLBBareNames(b bool) Option {
	return func(c *Config) {
		c.CLBBareNames = b
	}
}

func OptCLBExtended(b bool) Option {
	return func(c *Config) {
		c.CLBExtended = b
	}
}

func OptCLBExtinct(b *bool) Option {
	return func(c *Config) {
		c.CLBExtinct = b
	}
}

func OptCLBClassification(b bool) Option {
	return func(c *Config) {
		c.CLBClassification = b
	}
}

func OptCLBTaxGroups(b bool) Option {
	return func(c *Config) {
		c.CLBTaxGroups = b
	}
}

func OptCLBMinRank(s string) Option {
	return func(c *Config) {
		c.CLBMinRank = s
	}
}

func OptCLBTabFormat(s string) Option {
	return func(c *Config) {
		c.CLBTabFormat = s
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
		CacheDir:          cacheDir,
		JobsNum:           jobsNum,
		Code:              nomcode.Unknown,
		BadRow:            gnfmt.ProcessBadRow,
		BatchSize:         50_000,
		ArchiveDate:       today,
		CLBApi:            "https://api.checklistbank.org",
		CLBFormat:         "ColDP",
		CLBSynonyms:       true,
		CLBExtended:       true,
		CLBClassification: true,
		CLBTaxGroups:      true,
	}
	for _, opt := range opts {
		opt(&res)
	}

	res.DownloadDir = filepath.Join(res.CacheDir, "download")
	res.ExtractDir = filepath.Join(res.CacheDir, "extract")
	res.SfgaDir = filepath.Join(res.CacheDir, "sfga")

	return res
}
