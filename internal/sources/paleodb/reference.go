package paleodb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/sfborg/sflib/pkg/coldp"
)

func (p *paleodb) importReferences(citations map[string]string) error {
	taxonPath := filepath.Join(p.cfg.ExtractDir, "ref.json")
	jsonRef, err := os.ReadFile(taxonPath)
	var refs References
	err = json.Unmarshal(jsonRef, &refs)
	if err != nil {
		return err
	}
	res := make([]coldp.Reference, len(refs.Records))
	for i, v := range refs.Records {
		cit := citations[v.ID]
		ref := coldp.Reference{
			ID:        v.ID[4:],
			Type:      coldp.NewReferenceType(v.Type),
			Author:    authors(v.Author),
			Citation:  cit,
			Title:     v.Title,
			Volume:    v.Volume,
			Issue:     v.Number,
			Page:      v.Pages,
			ISBN:      v.ISBN,
			Publisher: v.Publisher,
			DOI:       doi(v.Identifier),
		}
		res[i] = ref

	}

	p.sfga.InsertReferences(res)
	return nil
}

func doi(id Identifier) string {
	if id.Type == "doi" {
		return id.ID
	}
	return ""
}

func authors(au []Author) string {
	res := make([]string, len(au))
	for i, v := range au {
		res[i] = strings.TrimSpace(v.Firstname + " " + v.Lastname)
	}
	return strings.Join(res, ", ")
}
