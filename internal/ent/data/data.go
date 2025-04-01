package data

// DataSet contains metadata required for conversion of an external
// biodiversity data source to SFGA format.
type DataSet struct {
	// Label is a short, unique identifier for the external data source. This
	// label is typically used for internal identification and file naming. For
	// example: "itis".
	Label string

	// Name returns the official, human-readable name of the external data
	// source. For example: "Integrated Taxonomic Information System".
	Name string
	// Notes provides a detailed description of the data source,
	// including information about its data structure, update frequency, and
	// any known limitations.  If the conversion process involves manual steps,
	// those steps MUST be documented clearly in this description.
	Notes string

	// ManualSteps is true if the conversion process requires manual
	// intervention or steps that cannot be fully automated.  If true, the
	// Description MUST provide detailed instructions for these manual steps.
	ManualSteps bool
	// URL provides the URL from which the source's data can be downloaded.
	// The URL can be provided by the maintainers of the source, or be
	// manually created if not available otherwise.
	URL string

	// New function creates the instance of Convertor interface from the
	// data provided in the set.
	New func(DataSet) Convertor
}
