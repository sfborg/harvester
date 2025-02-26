package ion

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/coldp/ent/coldp"
)

// nameIterator returns an iterator function that yields coldp.Name
// structs from a TSV file scanner.
func nameIterator(scanner *bufio.Scanner) func(func(coldp.Name) bool) {
	return func(yield func(coldp.Name) bool) {
		fields := struct {
			id, name, authorship int
		}{
			id:         0,
			name:       3,
			authorship: 4,
		}

		for scanner.Scan() {
			line := scanner.Text()
			row := strings.Split(line, "\t")
			n := coldp.Name{
				ID:                   row[fields.id],
				ScientificName:       row[fields.name],
				Authorship:           row[fields.authorship],
				ScientificNameString: row[fields.name] + " " + row[fields.authorship],
			}
			if !yield(n) {
				return
			}
		}
	}
}

// importNames reads names from a TSV file and processes them in batches.
// It uses a scanner to read the file line by line and an iterator function
// to yield coldp.Name structs. The names are processed in batches of size
// specified in the configuration.
func (i *ion) importNames() error {
	f, err := os.Open(filepath.Join(i.cfg.ExtractDir, "ion.tsv"))
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Skip the header row.
	if !scanner.Scan() {
		return scanner.Err()
	}

	iter := nameIterator(scanner)

	var count int
	names := make([]coldp.Name, 0, i.cfg.BatchSize)

	for n := range iter {
		count++
		names = append(names, n)
		if len(names) == i.cfg.BatchSize {
			if err := processBatch(i, names, count); err != nil {
				return err
			}
			names = names[:0]
		}
	}

	if len(names) > 0 {
		if err := processBatch(i, names, count); err != nil {
			return err
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

func processBatch(i *ion, names []coldp.Name, count int) error {
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s lines", humanize.Comma(int64(count)))
	err := i.sfga.InsertNames(names)

	if err != nil {
		return err
	}
	return nil
}
