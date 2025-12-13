package itis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

func (t *itis) importSynonyms() error {
	// Query synonyms and invalid names, excluding database artifacts.
	// COALESCE is used to handle NULL values from LEFT JOINs.
	q := `
SELECT
	tu.tsn,
	sl.tsn_accepted,
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
INNER JOIN synonym_links sl ON tu.tsn = sl.tsn
LEFT JOIN taxon_authors_lkp tal ON tu.taxon_author_id = tal.taxon_author_id
LEFT JOIN taxon_unit_types tut ON tu.rank_id = tut.rank_id
	AND tu.kingdom_id = tut.kingdom_id
WHERE COALESCE(tu.unaccept_reason, '') NOT IN (
	'unavailable, database artifact', 'database artifact'
)
`

	rows, err := t.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var nameUsages []coldp.NameUsage

	for rows.Next() {
		var tsn, tsnAccepted int
		var completeName, author, rankName string
		var rankID, kingdomID int
		var unitName1, unitName2, unitName3, unitName4 string
		var nameUsage, unacceptReason, updateDate string

		err = rows.Scan(
			&tsn, &tsnAccepted, &completeName, &author, &rankName,
			&rankID, &unitName1, &unitName2, &unitName3, &unitName4,
			&nameUsage, &unacceptReason, &updateDate, &kingdomID,
		)
		if err != nil {
			return err
		}

		nu := t.buildSynonymUsage(
			tsn, tsnAccepted, completeName, author, rankName, rankID,
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

func (t *itis) buildSynonymUsage(
	tsn, tsnAccepted int,
	completeName, author, rankName string,
	rankID int,
	unitName1, unitName2, unitName3, unitName4,
	nameUsage, unacceptReason, updateDate string,
	kingdomID int,
) coldp.NameUsage {
	id := strconv.Itoa(tsn)

	nu := coldp.NameUsage{
		ID:              id,
		ParentID:        strconv.Itoa(tsnAccepted),
		ScientificName:  completeName,
		Authorship:      author,
		TaxonomicStatus: coldp.SynonymTS,
		Code:            kingdomToCode(kingdomID),
		Link:            fmt.Sprintf("https://www.itis.gov/servlet/SingleRpt/SingleRpt?search_topic=TSN&search_value=%d", tsn),
		Modified:        updateDate,
	}

	// Set name status based on unaccept_reason.
	nu.NameStatus = mapNameStatus(kingdomID, unacceptReason)

	// Set rank.
	nu.Rank = coldp.NewRank(rankName)

	// Build scientific name string with authorship.
	nu.ScientificNameString = strings.TrimSpace(completeName + " " + author)

	// Set epithet fields based on rank.
	if rankID < 220 {
		nu.Uninomial = completeName
	} else {
		nu.GenericName = unitName1

		if strings.HasPrefix(unitName2, "(") && strings.HasSuffix(unitName2, ")") {
			nu.InfragenericEpithet = strings.Trim(unitName2, "()")
			nu.SpecificEpithet = unitName3
			nu.InfraspecificEpithet = unitName4
		} else {
			nu.SpecificEpithet = unitName2
			nu.InfraspecificEpithet = unitName3
		}
	}

	// Parse the name to get additional data.
	data.AddParsedData(parser, &nu)

	return nu
}

// mapNameStatus maps ITIS kingdom and unaccept_reason to COLDP name status.
func mapNameStatus(kingdomID int, unacceptReason string) coldp.NomStatus {
	if unacceptReason == "" {
		return coldp.Established
	}

	// Zoological codes (kingdoms 2, 5).
	if kingdomID == 2 || kingdomID == 5 {
		switch unacceptReason {
		case "junior synonym", "homonym & junior synonym":
			return coldp.Unacceptable
		case "junior homonym":
			return coldp.Unacceptable
		case "misapplied":
			return coldp.Unacceptable
		case "nomen dubium":
			return coldp.Doubtful
		case "nomen oblitum":
			return coldp.Unacceptable
		case "original name/combination":
			return coldp.Established
		case "subsequent name/combination":
			return coldp.Established
		case "unavailable, nomen nudum":
			return coldp.NotEstablished
		case "unavailable, incorrect orig. spelling":
			return coldp.NotEstablished
		case "unavailable, literature misspelling":
			return coldp.NotEstablished
		case "unavailable, suppressed by ruling":
			return coldp.Rejected
		case "unavailable, other":
			return coldp.NotEstablished
		case "unjustified emendation":
			return coldp.Unacceptable
		case "unnecessary replacement":
			return coldp.Unacceptable
		}
	}

	// Botanical codes (kingdoms 3, 4, 6).
	if kingdomID == 3 || kingdomID == 4 || kingdomID == 6 {
		switch unacceptReason {
		case "synonym":
			return coldp.Unacceptable
		case "homonym (illegitimate)":
			return coldp.Unacceptable
		case "invalidly published, nomen nudum":
			return coldp.NotEstablished
		case "invalidly published, other":
			return coldp.NotEstablished
		case "orthographic variant (misspelling)":
			return coldp.Unacceptable
		case "rejected name":
			return coldp.Rejected
		case "superfluous renaming (illegitimate)":
			return coldp.Unacceptable
		case "misapplied":
			return coldp.Unacceptable
		case "horticultural":
			return coldp.Unacceptable
		}
	}

	// Bacterial codes (kingdoms 1, 7).
	if kingdomID == 1 || kingdomID == 7 {
		switch unacceptReason {
		case "junior synonym":
			return coldp.Unacceptable
		case "original name/combination":
			return coldp.Established
		case "subsequent name/combination":
			return coldp.Established
		case "unavailable, other":
			return coldp.NotEstablished
		case "unavailable, suppressed by ruling":
			return coldp.Rejected
		}
	}

	return coldp.UnknownNomStatus
}
