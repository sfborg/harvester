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

func (i *ion) importNames() error {
	f, err := os.Open(filepath.Join(i.cfg.ExtractDir, "ion.tsv"))
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// skip headers
	scanner.Scan()

	fld := struct {
		id, name, authorship int
	}{
		id:         0,
		name:       3,
		authorship: 4,
	}
	var count int
	names := make([]coldp.Name, 0, i.cfg.BatchSize)
	for scanner.Scan() {
		count++
		line := scanner.Text()
		row := strings.Split(line, "\t")
		n := coldp.Name{
			ID:                   row[fld.id],
			ScientificName:       row[fld.name],
			Authorship:           row[fld.authorship],
			ScientificNameString: row[fld.name] + " " + row[fld.authorship],
		}
		names = append(names, n)
		if count%i.cfg.BatchSize == 0 {
			fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
			fmt.Fprintf(os.Stderr, "\rProcessed %s lines", humanize.Comma(int64(count)))
			err = i.sfga.InsertNames(names)
			if err != nil {
				return err
			}
			names = names[:0]
		}
	}
	fmt.Fprint(os.Stderr, "\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "\rProcessed %s lines", humanize.Comma(int64(count)))
	if len(names) == 0 {
		return nil
	}

	if err = i.sfga.InsertNames(names); err != nil {
		return err
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
