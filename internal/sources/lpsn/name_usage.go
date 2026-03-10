package lpsn

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gnames/gnfmt/gncsv"
	csvCfg "github.com/gnames/gnfmt/gncsv/config"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnuuid"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
)

func (l *lpsn) importNameUsages() error {
	var mu sync.Mutex
	var count, total int
	var batch []coldp.NameUsage
	refs := make(map[string]coldp.Reference)

	gnp := gnparser.New(gnparser.NewConfig(
		gnparser.OptCode(nomcode.Bacterial),
		gnparser.OptWithDetails(true),
	))

	ch := make(chan []string)
	var wg sync.WaitGroup
	wg.Add(1)

	opts := []csvCfg.Option{
		csvCfg.OptPath(l.path),
	}
	cfg, err := csvCfg.New(opts...)
	if err != nil {
		return err
	}
	csv := gncsv.New(cfg)

	go func() {
		defer wg.Done()
		for row := range ch {
			fields := getRow(row, csv.Headers())
			nu := buildNameUsage(fields)
			if nu == nil {
				continue
			}

			if citation := strings.TrimSpace(fields["reference"]); citation != "" {
				refID := "sf_" + gnuuid.New(citation).String()
				if _, exists := refs[citation]; !exists {
					refs[citation] = coldp.Reference{
						ID:       refID,
						Citation: citation,
					}
				}
				nu.NameReferenceID = refID
			}

			data.AddParsedData(gnp, nu)

			mu.Lock()
			count++
			total++
			batch = append(batch, *nu)
			if len(batch) >= l.cfg.BatchSize {
				if err := l.flushBatch(batch, total); err != nil {
					fmt.Fprintf(os.Stderr, "\nError flushing batch: %v\n", err)
				}
				batch = batch[:0]
				count = 0
			}
			mu.Unlock()
		}
	}()

	csv.Read(context.Background(), ch)
	close(ch)
	wg.Wait()

	if len(batch) > 0 {
		if err := l.flushBatch(batch, total); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "\n")

	refSlice := make([]coldp.Reference, 0, len(refs))
	for _, r := range refs {
		refSlice = append(refSlice, r)
	}
	return l.sfga.InsertReferences(refSlice)
}

func (l *lpsn) flushBatch(batch []coldp.NameUsage, total int) error {
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s names", humanize.Comma(int64(total)))
	return l.sfga.InsertNameUsages(batch)
}

// buildNameUsage converts a CSV row into a coldp.NameUsage.
// Returns nil if the row has no genus name.
func buildNameUsage(f map[string]string) *coldp.NameUsage {
	genus := strings.TrimSpace(f["genus_name"])
	if genus == "" {
		return nil
	}
	sp := strings.TrimSpace(f["sp_epithet"])
	subsp := strings.TrimSpace(f["subsp_epithet"])
	authors := strings.TrimSpace(f["authors"])
	status := strings.TrimSpace(f["status"])
	recordNo := strings.TrimSpace(f["record_no"])
	recordLnk := strings.TrimSpace(f["record_lnk"])
	address := strings.TrimSpace(f["address"])

	sciName := buildSciName(genus, sp, subsp)
	rank := determineRank(sp, subsp)
	taxStatus := parseTaxStatus(status)

	sciNameStr := strings.TrimSpace(sciName + " " + authors)

	nu := &coldp.NameUsage{
		ID:                   recordNo,
		ScientificName:       sciName,
		ScientificNameString: sciNameStr,
		Authorship:           authors,
		Rank:                 coldp.NewRank(rank),
		TaxonomicStatus:      taxStatus,
		Code:                 nomcode.Bacterial,
		Link:                 address,
		NameRemarks:          status,
	}

	nu.GenericName = genus
	if sp != "" {
		nu.SpecificEpithet = sp
		if subsp != "" {
			nu.InfraspecificEpithet = subsp
		}
	}

	// For synonyms, record_lnk points to the accepted name's record_no.
	if recordLnk != "" {
		nu.ParentID = recordLnk
	}

	return nu
}

func buildSciName(genus, sp, subsp string) string {
	if sp == "" {
		return genus
	}
	if subsp == "" {
		return genus + " " + sp
	}
	return genus + " " + sp + " subsp. " + subsp
}

func determineRank(sp, subsp string) string {
	if subsp != "" {
		return "subspecies"
	}
	if sp != "" {
		return "species"
	}
	return "genus"
}

// parseTaxStatus derives taxonomic status from the last segment of the
// semicolon-delimited LPSN status string.
func parseTaxStatus(status string) coldp.TaxonomicStatus {
	parts := strings.Split(status, ";")
	if len(parts) == 0 {
		return coldp.SynonymTS
	}
	last := strings.ToLower(strings.TrimSpace(parts[len(parts)-1]))

	switch {
	case strings.Contains(last, "correct name"):
		return coldp.AcceptedTS
	case strings.Contains(last, "orphaned"):
		return coldp.ProvisionallyAcceptedTS
	default:
		return coldp.SynonymTS
	}
}

func getRow(row []string, headers []string) map[string]string {
	res := make(map[string]string)
	for i, h := range headers {
		if i < len(row) {
			res[h] = row[i]
		}
	}
	return res
}
