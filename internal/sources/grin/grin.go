package grin

import (
	"bufio"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gnames/gnlib"
	"github.com/gnames/gnsys"
	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
	_ "modernc.org/sqlite"
)

type grin struct {
	data.Convertor
	cfg  config.Config
	sfga sfga.Archive
	db   *sql.DB
}

func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "grin",
		Name:  "GRIN Plant Taxonomy",
		Notes: `Create zip file from
https://npgsweb.ars-grin.gov/gringlobal/uploads/documents/taxonomy_data.cab
and save to the box.com.
`,
		ManualSteps: true,
		URL:         "https://uofi.box.com/shared/static/xob0fp0hw26hhz5lwdo421wspw9x8qbq.zip",
	}
	res := grin{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
	}
	return &res
}

func (g *grin) Import(path string) error {
	slog.Info("Importing GRIN data to a temporary SQLite database")
	err := gnsys.ExtractZip(path, g.cfg.ExtractDir)
	if err != nil {
		return err
	}

	files, err := grinFiles(g.cfg.ExtractDir)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite", filepath.Join(g.cfg.ExtractDir, "grin.sqlite"))
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA temp_store = MEMORY")
	if err != nil {
		return err
	}

	// Enable Write-Ahead Logging. Allow many reads and one write concurrently,
	// usually boosts write performance.
	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		return err
	}

	g.db = db

	for _, v := range files {
		file := filepath.Join(g.cfg.ExtractDir, v)
		err := createTable(db, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func grinFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".txt" {
			res = append(res, f.Name())
		}
	}
	return res, nil
}

func createTable(db *sql.DB, path string) error {
	name := filepath.Base(path)
	name = name[:len(name)-4]

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	line := scanner.Text()
	headers := strings.Split(line, "\t")

	fields := gnlib.Map(headers, func(s string) string {
		return fmt.Sprintf("    %s TEXT", s)
	})
	qFields := strings.Join(fields, ",\n")
	q := fmt.Sprintf(`
CREATE TABLE %s (
%s
)
`, name, qFields)

	_, err = db.Exec(q)
	if err != nil {
		return err
	}

	populateTable(db, scanner, name, headers)

	return nil
}

func populateTable(
	db *sql.DB,
	scanner *bufio.Scanner,
	table string,
	fields []string,
) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			slog.Error("Cannot finish transaction", "error", err)
			tx.Rollback()
		}
	}()

	questions := gnlib.Map(fields, func(s string) string {
		return "?"
	})
	q := fmt.Sprintf(`
INSERT INTO %s (%s)
		VALUES
		(%s)
`,
		table,
		strings.Join(fields, ","),
		strings.Join(questions, ","),
	)
	ch := make(chan []string)
	var wg sync.WaitGroup
	wg.Add(1)

	stmt, err := db.Prepare(q)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close() // Close the statement when done

	go func() {
		defer wg.Done()
		var count int
		for row := range ch {
			count++
			if count%10_000 == 0 {
				fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 60))
				fmt.Fprintf(os.Stderr, "\rTable '%s': processing %d's row", table, count)
			}
			// Convert []string to []any
			rowAny := make([]any, len(row))
			for i, v := range row {
				if v == "\\N" {
					v = ""
				}
				rowAny[i] = v
			}

			_, err := stmt.Exec(rowAny...)
			if err != nil {
				panic(fmt.Errorf("error executing statement: %w, row: %v", err, row))
			}
		}
	}()
	var num int
	for scanner.Scan() {
		num++
		line := scanner.Text()
		ch <- strings.Split(line, "\t")
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	close(ch)

	wg.Wait()

	err = tx.Commit()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 60))
	slog.Info("Imported data", "table", table, "rows", num)
	return nil
}
