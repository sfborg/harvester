package itis

import (
	"strconv"
	"time"

	"github.com/sfborg/sflib/pkg/coldp"
)

func (t *itis) importReferences() error {
	// Query all publications from ITIS.
	// COALESCE is used to handle NULL values.
	q := `
SELECT
	publication_id,
	COALESCE(reference_author, ''),
	COALESCE(title, ''),
	COALESCE(actual_pub_date, ''),
	COALESCE(publication_name, ''),
	COALESCE(pub_comment, '')
FROM publications
`

	rows, err := t.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var references []coldp.Reference

	for rows.Next() {
		var pubID int
		var author, title, pubDate, pubName, comment string

		err = rows.Scan(&pubID, &author, &title, &pubDate, &pubName, &comment)
		if err != nil {
			return err
		}

		ref := coldp.Reference{
			ID:             strconv.Itoa(pubID),
			Author:         author,
			Title:          title,
			ContainerTitle: pubName,
			Remarks:        comment,
		}

		// Extract year from actual_pub_date.
		if pubDate != "" {
			year := extractYear(pubDate)
			if year != "" {
				ref.Issued = year
			}
		}

		references = append(references, ref)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	err = t.sfga.InsertReferences(references)
	if err != nil {
		return err
	}

	return nil
}

// extractYear extracts the year from a date string.
func extractYear(dateStr string) string {
	// Try to parse as a date.
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return strconv.Itoa(t.Year())
		}
	}

	// If parsing fails, check if it's just a year.
	if len(dateStr) >= 4 {
		return dateStr[:4]
	}

	return ""
}
