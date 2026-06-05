package worms

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/sfborg/sflib/pkg/dwca"
)

var orcidRe = regexp.MustCompile(`\d{4}-\d{4}-\d{4}-\d{4}`)

func (w *worms) importMeta() error {
	eml := w.dwca.EML()
	if eml == nil {
		slog.Warn("no EML data in WoRMS DwCA, skipping metadata")
		return nil
	}

	var doi string
	if strings.Contains(eml.Dataset.ID, "doi.") {
		doi = eml.Dataset.ID
	}

	meta := coldp.Meta{
		Title:       eml.Dataset.Title,
		DOI:         doi,
		Description: eml.Dataset.Abstract.Para,
		Creators:    emlCreators(eml),
		Contact:     emlContact(eml),
	}

	return w.sfga.InsertMeta(&meta)
}

func emlCreators(eml *dwca.EML) []coldp.Actor {
	var res []coldp.Actor
	for _, c := range eml.Dataset.Creators {
		a := coldp.Actor{Email: c.ElectronicMailAddress}
		if c.IndividualName != nil {
			a.Given = c.IndividualName.GivenName
			a.Family = c.IndividualName.SurName
		}
		if c.OrganizationName != nil {
			a.Organization = c.OrganizationName.Value
		}
		if orcidRe.MatchString(c.ID) {
			a.Orcid = c.ID
		}
		res = append(res, a)
	}
	return res
}

func emlContact(eml *dwca.EML) *coldp.Actor {
	if len(eml.Dataset.Contacts) == 0 {
		return nil
	}
	c := eml.Dataset.Contacts[0]
	a := coldp.Actor{Email: c.ElectronicMailAddress}
	if c.IndividualName != nil {
		a.Given = c.IndividualName.GivenName
		a.Family = c.IndividualName.SurName
	}
	if c.OrganizationName != nil {
		a.Organization = c.OrganizationName.Value
	}
	return &a
}
