package wsparser

// types.go - Type definitions for Wikispecies parser

// Quality represents the parse quality level.
type Quality int

const (
	// QualityUnparseable means the input failed to parse.
	QualityUnparseable Quality = 0

	// QualityPartial means parsed but has significant unparsed tail.
	QualityPartial Quality = 1

	// QualityGood means parsed without tail.
	QualityGood Quality = 2
)

// ParsedName contains the extracted components of a Wikispecies name.
type ParsedName struct {
	Input      string  // Original input string
	Canonical  string  // Scientific name (e.g., "Homo sapiens")
	Authorship string  // Complete authorship (authors + year, e.g., "L., 1758")
	Reference  string  // Everything after ':' (bibliographic references)
	Tail       string  // Unparsed text after last component (no ':')
	Quality    Quality // Parse quality (0-2)
}
