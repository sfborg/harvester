package itis

import (
	"strconv"

	"github.com/sfborg/sflib/pkg/coldp"
)

func (t *itis) importVernaculars() error {
	// Query vernacular names for valid taxa only.
	// COALESCE is used to handle NULL values.
	q := `
SELECT DISTINCT
	v.tsn,
	COALESCE(v.vernacular_name, ''),
	COALESCE(v.language, '')
FROM vernaculars v
INNER JOIN taxonomic_units tu ON tu.tsn = v.tsn
WHERE tu.name_usage IN ('valid', 'accepted')
`

	rows, err := t.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var vernaculars []coldp.Vernacular

	for rows.Next() {
		var tsn int
		var name, language string

		err = rows.Scan(&tsn, &name, &language)
		if err != nil {
			return err
		}

		vern := coldp.Vernacular{
			TaxonID:  strconv.Itoa(tsn),
			Name:     name,
			Language: normalizeLanguage(language),
		}

		vernaculars = append(vernaculars, vern)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	err = t.sfga.InsertVernaculars(vernaculars)
	if err != nil {
		return err
	}

	return nil
}

// normalizeLanguage converts ITIS language names to ISO 639-3 codes.
func normalizeLanguage(lang string) string {
	switch lang {
	case "English":
		return "eng"
	case "Spanish":
		return "spa"
	case "French":
		return "fra"
	case "Portuguese":
		return "por"
	case "Italian":
		return "ita"
	case "German":
		return "deu"
	case "Hawaiian":
		return "haw"
	case "unspecified":
		return ""
	default:
		return lang
	}
}
