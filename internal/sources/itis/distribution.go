package itis

import (
	"strconv"

	"github.com/sfborg/sflib/pkg/coldp"
)

func (t *itis) importDistributions() error {
	// Query geographic distributions for valid taxa only.
	// COALESCE is used to handle NULL values.
	q := `
SELECT
	gd.tsn,
	COALESCE(gd.geographic_value, '')
FROM geographic_div gd
INNER JOIN taxonomic_units tu ON gd.tsn = tu.tsn
WHERE tu.name_usage IN ('valid', 'accepted')
`

	rows, err := t.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var distributions []coldp.Distribution

	for rows.Next() {
		var tsn int
		var area string

		err = rows.Scan(&tsn, &area)
		if err != nil {
			return err
		}

		dist := coldp.Distribution{
			TaxonID:   strconv.Itoa(tsn),
			Area:      area,
			Gazetteer: coldp.TextGz,
		}

		distributions = append(distributions, dist)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	err = t.sfga.InsertDistributions(distributions)
	if err != nil {
		return err
	}

	return nil
}
