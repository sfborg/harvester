package text

import (
	"bufio"
	"fmt"
	"iter"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/coldp/ent/coldp"
)

func (t *text) importNames() error {
	f, err := os.Open(t.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Skip the header row.
	if !scanner.Scan() {
		return scanner.Err()
	}

	iter := t.nameIterator(scanner)

	var count int
	names := make([]coldp.Name, 0, t.cfg.BatchSize)

	for n := range iter {
		count++
		names = append(names, n)
		if len(names) == t.cfg.BatchSize {
			if err := t.processBatch(names, count); err != nil {
				return err
			}
			names = names[:0]
		}
	}

	if len(names) > 0 {
		if err := t.processBatch(names, count); err != nil {
			return err
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}
	return nil
}

// nameIterator returns an iterator function that yields coldp.Name
// structs from a TSV file scanner.
func (t *text) nameIterator(scanner *bufio.Scanner) iter.Seq[coldp.Name] {
	return func(yield func(coldp.Name) bool) {
		ids := make(map[string]struct{})
		for scanner.Scan() {
			name := scanner.Text()
			parsed := t.Parse(name)
			// skip duplicates
			if _, ok := ids[parsed.NameID]; ok {
				continue
			}

			ids[parsed.NameID] = struct{}{}
			n := coldp.Name{
				ID:                    parsed.NameID,
				ScientificNameString:  name,
				ScientificName:        parsed.CanonicalFull,
				Authorship:            parsed.Authorship,
				CombinationAuthorship: parsed.CombinationAuthorship,
				Uninomial:             parsed.Uninomial,
				Genus:                 parsed.Genus,
				SpecificEpithet:       parsed.Species,
				Rank:                  coldp.NewRank(parsed.Rank),
				InfraspecificEpithet:  parsed.Infraspecies,
			}
			if !yield(n) {
				return
			}
		}
	}
}

func (t *text) processBatch(names []coldp.Name, count int) error {
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s lines", humanize.Comma(int64(count)))

	err := t.sfga.InsertNames(names)
	if err != nil {
		return err
	}
	return nil
}
