package itis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

var parser = gnparser.New(gnparser.NewConfig(gnparser.OptWithDetails(true)))

func (t *itis) importNameUsages() error {
	// Query to get all accepted taxa with their hierarchy and name information.
	// This query handles the different name constructions for different ranks.
	// COALESCE is used to handle NULL values from LEFT JOINs.
	q := `
SELECT
	tu.tsn,
	COALESCE(h.parent_tsn, 0),
	tu.complete_name,
	COALESCE(tal.taxon_author, ''),
	COALESCE(LOWER(tut.rank_name), ''),
	tu.rank_id,
	COALESCE(tu.unit_name1, ''),
	COALESCE(tu.unit_name2, ''),
	COALESCE(tu.unit_name3, ''),
	COALESCE(tu.unit_name4, ''),
	COALESCE(tu.name_usage, ''),
	COALESCE(tu.unaccept_reason, ''),
	COALESCE(tu.update_date, ''),
	tu.kingdom_id
FROM taxonomic_units tu
LEFT JOIN hierarchy h ON tu.tsn = h.tsn
LEFT JOIN taxon_authors_lkp tal ON tu.taxon_author_id = tal.taxon_author_id
LEFT JOIN taxon_unit_types tut ON tu.rank_id = tut.rank_id
	AND tu.kingdom_id = tut.kingdom_id
WHERE tu.name_usage IN ('valid', 'accepted')
	AND (tu.unaccept_reason IS NULL OR tu.unaccept_reason = '')
`

	rows, err := t.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var nameUsages []coldp.NameUsage

	for rows.Next() {
		var tsn, parentTSN int
		var completeName, author, rankName string
		var rankID, kingdomID int
		var unitName1, unitName2, unitName3, unitName4 string
		var nameUsage, unacceptReason, updateDate string

		err = rows.Scan(
			&tsn, &parentTSN, &completeName, &author, &rankName,
			&rankID, &unitName1, &unitName2, &unitName3, &unitName4,
			&nameUsage, &unacceptReason, &updateDate, &kingdomID,
		)
		if err != nil {
			return err
		}

		nu := t.buildNameUsage(
			tsn, parentTSN, completeName, author, rankName, rankID,
			unitName1, unitName2, unitName3, unitName4,
			nameUsage, unacceptReason, updateDate, kingdomID,
		)

		nameUsages = append(nameUsages, nu)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	err = t.sfga.InsertNameUsages(nameUsages)
	if err != nil {
		return err
	}

	return nil
}

func (t *itis) buildNameUsage(
	tsn, parentTSN int,
	completeName, author, rankName string,
	rankID int,
	unitName1, unitName2, unitName3, unitName4,
	nameUsage, unacceptReason, updateDate string,
	kingdomID int,
) coldp.NameUsage {
	id := strconv.Itoa(tsn)

	nu := coldp.NameUsage{
		ID:              id,
		ScientificName:  completeName,
		Authorship:      author,
		TaxonomicStatus: coldp.AcceptedTS,
		NameStatus:      coldp.Established,
		Code:            kingdomToCode(kingdomID),
		Link:            fmt.Sprintf("https://www.itis.gov/servlet/SingleRpt/SingleRpt?search_topic=TSN&search_value=%d", tsn),
		Modified:        updateDate,
	}

	// Set parent ID for accepted taxa.
	if parentTSN > 0 && parentTSN != tsn {
		nu.ParentID = strconv.Itoa(parentTSN)
	}

	// Set rank.
	nu.Rank = coldp.NewRank(rankName)

	// Build scientific name string with authorship.
	nu.ScientificNameString = strings.TrimSpace(completeName + " " + author)

	// Set epithet fields based on rank.
	if rankID < 220 {
		// Ranks above genus level use uninomial.
		nu.Uninomial = completeName
	} else {
		// Genus and below.
		nu.GenericName = unitName1

		// Check if unit_name2 is a subgenus (in parentheses).
		if strings.HasPrefix(unitName2, "(") && strings.HasSuffix(unitName2, ")") {
			// Has subgenus.
			nu.InfragenericEpithet = strings.Trim(unitName2, "()")
			nu.SpecificEpithet = unitName3
			nu.InfraspecificEpithet = unitName4
		} else {
			nu.SpecificEpithet = unitName2
			nu.InfraspecificEpithet = unitName3
		}
	}

	// Check extinct status.
	if t.extinct[tsn] {
		nu.Extinct = coldp.ToBool(true)
	}

	// Parse the name to get additional data.
	data.AddParsedData(parser, &nu)

	return nu
}

// kingdomToCode maps ITIS kingdom_id to nomenclatural codes.
// Kingdom IDs:
//
//	1 = Bacteria (bacterial)
//	2 = Protozoa (zoological)
//	3 = Plantae (botanical)
//	4 = Fungi (botanical)
//	5 = Animalia (zoological)
//	6 = Chromista (botanical)
//	7 = Archaea (bacterial)
func kingdomToCode(kingdomID int) nomcode.Code {
	switch kingdomID {
	case 1, 7:
		return nomcode.Bacterial
	case 2, 5:
		return nomcode.Zoological
	case 3, 4, 6:
		return nomcode.Botanical
	default:
		return nomcode.Unknown
	}
}
