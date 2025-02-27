package data

// Parsed represents the parsed scientific name data.
type Parsed struct {
	// Quality indicates the quality score of the parsed name.
	Quality int
	// NameID is a UUID v5 generated from the name verbatim string.
	NameID string
	// CanonicalFull is the full canonical form of the name.
	CanonicalFull string
	// CanonicalSimple is the simple canonical form of the name.
	// It does not have ranks, hybrid signs etc.
	CanonicalSimple string
	// CanonicalStemmed are stemmed CanonicalSimple names where
	// suffixes of specific and infraspecific epithets are removed.
	CanonicalStemmed string
	// Authorship is the authorship of the name.
	Authorship string
	// CombinationAuthorship is the authorship of the combination.
	CombinationAuthorship string
	// Uninomial is the uninomial part of the name.
	Uninomial string
	// Genus is the genus part of the name.
	Genus string
	// Subgenus is the subgenus part of the name.
	Subgenus string
	// Species is the species part of the name.
	Species string
	// Rank of the name.
	Rank string
	// Infraspecies is the infraspecies part of the name.
	Infraspecies string
	// UnparsedTail is the unparsed tail of the name.
	UnparsedTail string
}
