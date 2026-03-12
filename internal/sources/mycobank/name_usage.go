package mycobank

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/xuri/excelize/v2"
)

// xlsx column indices (0-based).
const (
	colID         = 0
	colTaxonName  = 1
	colAuthors    = 3 // abbreviated authors (col D), standard for botanical names
	colRank       = 4
	colYear       = 5
	colNameStatus = 6
	colMBNum      = 7 // MycoBank # – used for external links; not unique
	colLink       = 8
	colCurrentMB  = 10 // formula: MB# of the current/accepted name
)

// mbRow holds the raw fields needed from one xlsx row.
type mbRow struct {
	id        string // col A – unique row ID, same number used in the URL
	taxonName string // col B
	authors   string // col C
	rank      string // col E
	year      string // col F
	status    string // col G
	mbNum     string // col H – MycoBank # (used as outlink, not primary key)
	link      string // col I
	currentMB string // col K – MB# of the accepted name (formula result)
}

func (m *mycobank) importNameUsages() error {
	gnp := gnparser.New(gnparser.NewConfig(
		gnparser.OptCode(nomcode.Botanical),
		gnparser.OptWithDetails(true),
	))

	f, err := excelize.OpenFile(m.xlsxPath)
	if err != nil {
		return fmt.Errorf("opening xlsx: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("no sheets found in %s", m.xlsxPath)
	}

	rows, err := f.Rows(sheets[0])
	if err != nil {
		return fmt.Errorf("reading rows from %s: %w", sheets[0], err)
	}
	defer rows.Close()

	// Pass 1: collect all rows and build MB# → row-ID index.
	fmt.Fprint(os.Stderr, "Reading xlsx...")
	var allRows []mbRow
	// mbNumToID maps MycoBank # (col H) → row ID (col A).
	mbNumToID := make(map[string]string)

	rows.Next() // skip header
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			continue
		}
		id := strings.TrimSpace(getCol(cols, colID))
		mbNum := strings.TrimSpace(getCol(cols, colMBNum))
		taxonName := strings.TrimSpace(getCol(cols, colTaxonName))
		if id == "" || taxonName == "" {
			continue
		}

		r := mbRow{
			id:        id,
			taxonName: taxonName,
			authors:   strings.TrimSpace(getCol(cols, colAuthors)),
			rank:      strings.TrimSpace(getCol(cols, colRank)),
			year:      strings.TrimSpace(getCol(cols, colYear)),
			status:    strings.TrimSpace(getCol(cols, colNameStatus)),
			mbNum:     mbNum,
			link:      strings.TrimSpace(getCol(cols, colLink)),
			currentMB: strings.TrimSpace(getCol(cols, colCurrentMB)),
		}
		allRows = append(allRows, r)

		// Only record the first occurrence of each MB# to avoid overwriting
		// the accepted name with a duplicate entry.
		if mbNum != "" {
			if _, exists := mbNumToID[mbNum]; !exists {
				mbNumToID[mbNum] = id
			}
		}
	}
	fmt.Fprintf(os.Stderr, " done (%s rows)\n", humanize.Comma(int64(len(allRows))))

	// Pass 2: build NameUsages with resolved parent IDs and flush in batches.
	var total int
	var batch []coldp.NameUsage

	for i := range allRows {
		nu := buildNameUsage(&allRows[i], mbNumToID)
		data.AddParsedData(gnp, nu)

		total++
		batch = append(batch, *nu)
		if len(batch) >= m.cfg.BatchSize {
			if err := m.flushBatch(batch, total); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := m.flushBatch(batch, total); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "\n")
	return nil
}

func (m *mycobank) flushBatch(batch []coldp.NameUsage, total int) error {
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s names", humanize.Comma(int64(total)))
	return m.sfga.InsertNameUsages(batch)
}

func buildNameUsage(r *mbRow, mbNumToID map[string]string) *coldp.NameUsage {
	sciNameStr := strings.TrimSpace(r.taxonName + " " + r.authors)
	taxStatus := parseTaxStatus(r.mbNum, r.currentMB)

	nu := &coldp.NameUsage{
		ID:                   r.id,
		ScientificName:       r.taxonName,
		ScientificNameString: sciNameStr,
		Authorship:           r.authors,
		Rank:                 coldp.NewRank(r.rank),
		TaxonomicStatus:      taxStatus,
		Code:                 nomcode.Botanical,
		Link:                 r.link,
		NameRemarks:          r.status,
		NameAlternativeID:    "mycobank:" + r.mbNum,
		PublishedInYear:      r.year,
	}

	// Resolve synonym → accepted name using the MB# index.
	if taxStatus == coldp.SynonymTS && r.currentMB != "" {
		if parentID, ok := mbNumToID[r.currentMB]; ok {
			nu.ParentID = parentID
		}
	}

	return nu
}

// parseTaxStatus checks whether the entry's own MB# matches the current/
// accepted name's MB# (pre-computed in the xlsx via formula).
func parseTaxStatus(mbNum, currentMB string) coldp.TaxonomicStatus {
	if currentMB == "" || currentMB == mbNum {
		return coldp.AcceptedTS
	}
	return coldp.SynonymTS
}

func getCol(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}
