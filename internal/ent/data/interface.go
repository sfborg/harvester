package data

import (
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/sflib/ent/sfga"
)

// Convertor provides methods for converting data from an external source to
// the SFGA file format. Implementations of this interface handle the specific
// details of each external data source (e.g., ITIS, GBIF, etc.).
type Convertor interface {
	Accessor
	Processor
	Parser
}

// Accessor provides methods to get data from internal fields of the
// Accessor implenentation.
type Accessor interface {
	// Label returns a short, unique identifier for the external data source.
	// This label is typically used for internal identification and file naming.
	// For example: "itis".
	Label() string

	// Name returns the official, human-readable name of the external data
	// source. For example: "Integrated Taxonomic Information System".
	Name() string

	// Config returns the configuration data.
	Config() config.Config

	// Description provides a detailed description of the data source, including
	// information about its data structure, update frequency, and any known
	// limitations.  If the conversion process involves manual steps, those steps
	// MUST be documented clearly in this description.
	Description() string

	// ManualSteps returns true if the conversion process requires manual
	// intervention or steps that cannot be fully automated.  If true, the
	// Description() method MUST provide detailed instructions for these manual
	// steps.
	ManualSteps() bool
}

// Processor contains actions for converting source into SFGArchive.
type Processor interface {
	// Download downloads the data from the external source and stores it in a
	// local cache. It returns the path to the downloaded file in the cache.  If
	// the download fails, an error is returned, and the returned path may be an
	// empty string.  Implementations should handle caching appropriately (e.g.,
	// checking if the file already exists before downloading).
	Download() (string, error)

	// Import extracts the relevant data from the downloaded file.  The 'path'
	// argument is the path returned by the Download() method.  The extracted
	// data should be placed in a separate cache directory.  An error is returned
	// if extraction fails.
	Import(path string) error

	// InitSFGA creates empty sfga.Archive and returns its instance.
	// In case of a failure it returns an error.
	InitSFGA() (sfga.Archive, error)

	// ToSFGA converts the extracted data to the SFGA file format.  This method
	// should use the data previously extracted by the Extract() method.  An
	// error is returned if the conversion fails.
	ToSFGA(sfga.Archive) error
}

// Parser contains method to parse scientific names and
// collect relevant data to Parsed object.
type Parser interface {
	// Parse converts string into Parsed result.
	Parse(name string) Parsed
}
