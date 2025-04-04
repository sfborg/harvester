package paleodb

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gnames/gnfmt/gncsv"
	"github.com/gnames/gnfmt/gncsv/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

func (p *paleodb) importNameUsages() error {
	taxonPath := filepath.Join(p.cfg.ExtractDir, "taxon.csv")
	cfg, err := config.New(config.OptPath(taxonPath))
	if err != nil {
		return err
	}

	csv := gncsv.New(cfg)
	ch := make(chan [][]string)

	go func() {
		for rows := range ch {

			nus := make([]coldp.NameUsage, 0, len(rows[0]))
			verns := make([]coldp.Vernacular, 0, len(rows[0]))

			for _, r := range rows {
				taxStatus := coldp.AcceptedTS
				id := csv.F(r, "orig_no")
				parentID := csv.F(r, "parent_no")
				acceptedID := csv.F(r, "accepted_no")
				if acceptedID != id {
					parentID = acceptedID
					taxStatus = coldp.SynonymTS
				}
				if parentID == "0" {
					parentID = ""
				}

				name := csv.F(r, "taxon_name")
				au := csv.F(r, "taxon_attr")
				nameString := strings.TrimSpace(name + " " + au)

				rank := coldp.NewRank(csv.F(r, "accepted_rank"))
				remark := csv.F(r, "difference")

				vern := csv.F(r, "common_name")
				if vern != "" && taxStatus != coldp.SynonymTS {
					verns = append(verns, coldp.Vernacular{
						TaxonID:  id,
						Name:     vern,
						Language: "eng",
					})
				}
				start := coldp.NewGeoTime(csv.F(r, "early_interval"))
				end := coldp.NewGeoTime(csv.F(r, "late_interval"))

				nu := coldp.NameUsage{
					ID:                   id,
					AlternativeID:        csv.F(r, "taxon_no"),
					ParentID:             parentID,
					ScientificName:       name,
					Authorship:           au,
					ScientificNameString: nameString,
					Rank:                 rank,
					TaxonomicStatus:      taxStatus,
					NamePhrase:           remark,
					ReferenceID:          csv.F(r, "reference_no"),
					TemporalRangeStart:   start,
					TemporalRangeEnd:     end,
					Genus:                csv.F(r, "genus"),
					Family:               csv.F(r, "family"),
					Order:                csv.F(r, "order"),
					Class:                csv.F(r, "class"),
					Phylum:               csv.F(r, "phylum"),
				}

				switch csv.F(r, "is_extant") {
				case "extant":
					nu.Extinct = coldp.ToBool(false)
				case "extinct":
					nu.Extinct = coldp.ToBool(true)
				}

				data.AddParsedData(p.p, &nu)
				nus = append(nus, nu)
			}
			p.sfga.InsertNameUsages(nus)
			p.sfga.InsertVernaculars(verns)
		}
	}()

	_, err = csv.ReadChunks(context.Background(), ch, p.cfg.BatchSize)
	if err != nil {
		return err
	}
	return nil
}
