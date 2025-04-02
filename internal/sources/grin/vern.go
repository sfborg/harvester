package grin

import (
	"strings"

	"github.com/gnames/gnfmt/gnlang"
	"github.com/sfborg/sflib/pkg/coldp"
)

func (g *grin) importVern() error {
	q := `
SELECT name, language_description, taxonomy_species_id
	FROM taxonomy_common_name
	WHERE taxonomy_species_id != ''
`
	rows, err := g.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var res []coldp.Vernacular
	for rows.Next() {
		var name, lang, id string
		err := rows.Scan(&name, &lang, &id)
		if err != nil {
			return err
		}
		langCode, countryCode := processLang(lang)
		vern := coldp.Vernacular{
			TaxonID:  id,
			Name:     name,
			Language: langCode,
			Country:  countryCode,
			Remarks:  lang,
		}
		res = append(res, vern)
	}
	g.sfga.InsertVernaculars(res)
	return nil
}

func processLang(lang string) (string, string) {
	ls := strings.Split(lang, "(")
	var country string
	lang = strings.TrimSpace(ls[0])
	if len(ls) > 1 {
		country = strings.Trim(ls[1], " )")
	}
	langCode := gnlang.LangCode(lang)
	cntCode := gnlang.CountryCode(country)
	return langCode, cntCode
}
