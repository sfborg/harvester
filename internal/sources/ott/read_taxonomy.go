package ott

import (
	"strings"
)

// parseTaxonLine converts split TSV fields into a datum.
// Returns the datum and false if the row should be skipped.
func parseTaxonLine(row []string, synonyms map[string][]synDatum) (datum, bool) {
	uid := strings.TrimSpace(row[0])
	if uid == "" {
		return datum{}, false
	}

	flags := ""
	if len(row) > 6 {
		flags = strings.TrimSpace(row[6])
	}
	if strings.Contains(flags, "merged") || strings.Contains(flags, "hidden") {
		return datum{}, false
	}

	parentID := strings.TrimSpace(row[1])
	name := strings.TrimSpace(row[2])
	rank := strings.TrimSpace(row[3])

	if rank == "no rank" || rank == "no rank - terminal" {
		rank = ""
	}

	return datum{
		taxonID:  uid,
		parentID: parentID,
		name:     name,
		rank:     rank,
		flags:    flags,
		synonyms: synonyms[uid],
	}, true
}
