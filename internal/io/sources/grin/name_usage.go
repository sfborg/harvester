package grin

import (
	"log/slog"
	"strings"

	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/nomcode"
	"github.com/gnames/gnuuid"
	"github.com/sfborg/harvester/internal/ent/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

func (g *grin) importNameUsages() error {
	q := `
SELECT 
  s.taxonomy_species_id AS id, s.current_taxonomy_species_id AS accepted_id,
	s.name, s.name_authority, s.synonym_code, s.protologue,
	s.protologue_virtual_path, s.modified_date, 
	f.suprafamily_rank_code, f.suprafamily_rank_name, f.family_name, f.family_authority,
	f.subfamily_name, f.tribe_name, f.subtribe_name,
	g.genus_name, g.genus_authority, g.subgenus_name, g.section_name
FROM taxonomy_species s 
	JOIN taxonomy_genus g 
	  ON g.taxonomy_genus_id = s.taxonomy_genus_id
	JOIN taxonomy_family f 
	  ON f.taxonomy_family_id = g.taxonomy_family_id
`

	basionyms, err := g.getBasionyms()
	if err != nil {
		return err
	}
	p := gnparser.New(gnparser.NewConfig(
		[]gnparser.Option{
			gnparser.OptWithDetails(true),
			gnparser.OptCode(nomcode.Botanical),
		}...))
	rows, err := g.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()
	slog.Info("Collecting name usages")
	var res []coldp.NameUsage
	refs := make(map[string]string)

	for rows.Next() {
		var id, acceptedID, name, authority, synonymCode string
		var protologue, protologueUrl, modified string
		var suprafamilyType, suprafamily, family, familyAuthority string
		var subfamily, tribe, subtribe string
		var genus, genusAuthority, subgenus, section string
		err = rows.Scan(
			&id, &acceptedID, &name, &authority, &synonymCode,
			&protologue, &protologueUrl, &modified,
			&suprafamilyType, &suprafamily, &family, &familyAuthority,
			&subfamily, &tribe, &subtribe, &genus, &genusAuthority,
			&subgenus, &section,
		)

		basionymID := basionyms[id]
		tStatus, nStatus := getStatus(id, acceptedID, synonymCode)
		nameString := name + " " + authority
		// family = family + " " + familyAuthority
		// genus = genus + " " + genusAuthority
		var ord string
		if suprafamilyType == "ORDER" {
			ord = suprafamily
			suprafamily = ""
		}

		var refID string
		if protologue != "" {
			var ok bool
			if refID, ok = refs[protologue]; !ok {
				refID = gnuuid.New(protologue).String()
				refs[protologue] = refID
			}
		}
		if !strings.HasPrefix(protologueUrl, "http") {
			protologueUrl = ""
		}

		nu := coldp.NameUsage{
			ID:                   id,
			ScientificNameString: nameString,
			ScientificName:       name,
			Authorship:           authority,
			ParentID:             getParentID(id, acceptedID),
			BasionymID:           basionymID,
			TaxonomicStatus:      tStatus,
			NameStatus:           nStatus,
			Code:                 coldp.Botanical,
			ReferenceID:          refID,
			Order:                ord,
			Superfamily:          suprafamily,
			Family:               family,
			Subfamily:            subfamily,
			Tribe:                tribe,
			Subtribe:             subtribe,
			Genus:                genus,
			Subgenus:             subgenus,
			Section:              section,
			Link:                 protologueUrl,
			Modified:             modified,
		}

		data.AddParsedData(p, &nu)

		res = append(res, nu)
	}
	err = g.sfga.InsertNameUsages(res)
	if err != nil {
		return err
	}

	var references []coldp.Reference
	for prot, refID := range refs {
		ref := coldp.Reference{
			ID:       refID,
			Citation: prot,
		}
		references = append(references, ref)
	}
	err = g.sfga.InsertReferences(references)
	if err != nil {
		return err
	}

	return nil
}

func getStatus(
	id, acceptedID, synonymCode string,
) (coldp.TaxonomicStatus, coldp.NomStatus) {
	if id == acceptedID {
		return coldp.AcceptedTS, coldp.Established
	}
	switch synonymCode {
	case "":
		slog.Warn("no annotation for synonym", "id", id)
		return coldp.SynonymTS, coldp.UnknownNomStatus
	case "Invalid":
		return coldp.SynonymTS, coldp.Unacceptable
	default:
		return coldp.SynonymTS, coldp.UnknownNomStatus
	}
}

func getParentID(id, acceptedID string) string {
	if id == acceptedID {
		return ""
	}
	return acceptedID
}

func (g *grin) getBasionyms() (map[string]string, error) {
	slog.Info("Getting basionyms")
	q := `
SELECT taxonomy_species_id, current_taxonomy_species_id
	FROM taxonomy_species
  WHERE synonym_code = 'B'
`
	rows, err := g.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]string)
	for rows.Next() {
		var basionym, current string
		err = rows.Scan(&basionym, &current)
		if err != nil {
			return nil, err
		}
		res[current] = basionym
	}
	return res, nil
}
