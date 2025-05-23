package ioc

// Fields in v15.1:
// Infraclass, Parvclass, Order, Family (Scientific),
// Family (English), Genus, Species (Scientific), Subspecies,
// Authority, Species (English), Breeding Range, Breeding Range-Subregion(s),
// Nonbreeding Range, Code, Comment

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnfmt/gncsv"
	csvCfg "github.com/gnames/gnfmt/gncsv/config"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/harvester/internal/util"
	"github.com/sfborg/sflib/pkg/coldp"
)

var ranks = []string{
	"Class", "Infraclass", "Parvclass", "Order", "Family (Scientific)",
	"Genus", "Species (Scientific)", "Subspecies",
}

var parser = gnparser.New(gnparser.NewConfig())

type name struct {
	id                                       string
	rank                                     string
	species                                  string
	path                                     map[string]string
	authority                                string
	speciesEng                               string
	breeding, breedingSubregion, nonbreeding string
	code, comment                            string
}

func newName() name {
	return name{rank: "class", path: map[string]string{"class": "Aves"}}
}

func (n name) update(row map[string]string) name {
	for _, v := range ranks {
		if row[v] != "" {
			n.rank = v
			n.path[v] = row[v]
			break
		}
	}
	n.authority = row["Authority"]
	n.speciesEng = row["Species (English)"]
	n.breeding = row["Breeding Range"]
	n.breedingSubregion = row["Breeding Range-Subregion(s)"]
	n.nonbreeding = row["Nonbreeding Range"]
	n.code = row["Code"]
	n.comment = row["Comment"]
	return n.clean()
}

func (n name) sciName() string {
	if n.path["Genus"] == "" {
		for i := 4; i >= 0; i-- {
			rnk := ranks[i]
			val := n.path[rnk]
			if val != "" {
				return util.ToTitleCaseWord(val)
			}
		}
	}
	ary := []string{n.path["Genus"], n.path["Species (Scientific)"],
		n.path["Subspecies"]}
	res := strings.Join(ary, " ")
	return strings.TrimSpace(res)
}

func (n name) usage() *coldp.NameUsage {
	res := &coldp.NameUsage{
		ID:                   n.id,
		TaxonomicStatus:      coldp.AcceptedTS,
		ScientificName:       n.sciName(),
		Authorship:           util.NormalizeAuthors(n.authority),
		GenericName:          n.path["Genus"],
		SpecificEpithet:      n.path["Species (Scientific)"],
		InfraspecificEpithet: n.path["Subspecies"],
		NameStatus:           coldp.Established,
		Code:                 nomcode.Zoological,
		Genus:                util.ToTitleCaseWord(n.path["Genus"]),
		Family:               util.ToTitleCaseWord(n.path["Family (Scientific)"]),
		Order:                util.ToTitleCaseWord(n.path["Order"]),
		Subclass:             util.ToTitleCaseWord(n.path["Infraclass"]),
		Class:                "Aves",
		Phylum:               "Chordata",
		Kingdom:              "Animalia",
	}
	if strings.Contains(res.ScientificName, "†") {
		res.Extinct = coldp.ToBool(true)
	}
	if n.path["Genus"] == "" {
		res.Uninomial = res.ScientificName
	}
	str := res.ScientificName + " " + res.Authorship
	str = strings.ReplaceAll(str, "†", "")
	res.ScientificNameString = strings.TrimSpace(str)

	rank := n.rank
	if rank == "Species (Scientific)" {
		rank = "species"
	}
	if rank == "Family (Scientific)" {
		rank = "family"
	}
	res.Rank = coldp.NewRank(rank)

	if res.ScientificName == "" {
		return nil
	}

	res.Amend(parser)

	return res
}

func (n name) vern() *coldp.Vernacular {
	if n.speciesEng == "" {
		return nil
	}
	res := coldp.Vernacular{
		TaxonID:  n.id,
		Name:     n.speciesEng,
		Language: "eng",
	}
	return &res
}

func (n name) clean() name {
	var found bool
	for _, v := range ranks {
		if v == n.rank {
			found = true
			continue
		}
		if found {
			n.path[v] = ""
		}
	}
	return n
}

func (l *ioc) importNameUsages() error {
	var res coldp.Data
	ch := make(chan []string)
	var wg sync.WaitGroup
	wg.Add(1)

	opts := []csvCfg.Option{
		csvCfg.OptPath(l.path),
		csvCfg.OptBadRowMode(gnfmt.ProcessBadRow),
		csvCfg.OptWithQuotes(true),
	}
	cfg, err := csvCfg.New(opts...)
	if err != nil {
		return err
	}
	csv := gncsv.New(cfg)

	n := newName()
	go func() {
		defer wg.Done()
		var count int
		for l := range ch {
			count++
			n = n.update(getRow(l, csv.Headers()))
			n.id = "gn_" + strconv.Itoa(count)
			nu := n.usage()
			if nu != nil {
				res.NameUsages = append(res.NameUsages, *nu)
			}
			vern := n.vern()
			if vern != nil && nu != nil {
				res.Vernaculars = append(res.Vernaculars, *vern)
			}

		}
	}()

	csv.Read(context.Background(), ch)
	close(ch)

	wg.Wait()

	l.sfga.InsertNameUsages(res.NameUsages)
	l.sfga.InsertVernaculars(res.Vernaculars)

	return nil
}

func getRow(l []string, headers []string) map[string]string {
	res := make(map[string]string)
	for i, v := range headers {
		res[v] = l[i]
	}
	return res
}
