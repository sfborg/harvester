package itis

import (
	"bufio"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sfborg/harvester/internal/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/sfborg/harvester/pkg/data"
	"github.com/sfborg/sflib/pkg/sfga"
	_ "modernc.org/sqlite"
)

const extinctURL = "https://raw.githubusercontent.com/CatalogueOfLife/" +
	"data-itis/master/raw/extinct.tsv"

type itis struct {
	data.Convertor
	cfg     config.Config
	sfga    sfga.Archive
	db      *sql.DB
	dbPath  string
	extinct map[int]bool
}

// New creates a new ITIS data convertor.
func New(cfg config.Config) data.Convertor {
	set := data.DataSet{
		Label: "itis",
		Name:  "Integrated Taxonomic Information System",
		Notes: `ITIS provides taxonomic information on plants, animals,
fungi, and microbes of North America and the world. The SQLite database
is downloaded directly from ITIS.gov. Note that ITIS does not provide
extinction status, so this data comes from a manually maintained list.`,
		ManualSteps: false,
		URL:         "https://itis.gov/downloads/itisSqlite.zip",
	}
	res := itis{
		cfg:       cfg,
		Convertor: base.New(cfg, &set),
		extinct:   make(map[int]bool),
	}
	return &res
}

// Extract extracts the ITIS SQLite database from the downloaded ZIP file.
// If path is empty (skip download mode), it tries to find an existing
// database in the extract directory.
func (t *itis) Extract(path string) error {
	slog.Info("Extracting ITIS SQLite database")

	// If path is provided, extract the archive first.
	if path != "" {
		err := t.Convertor.Extract(path)
		if err != nil {
			return err
		}
	}

	// Find the SQLite database file in the extracted directory.
	dbPath, err := findSQLiteDB(t.cfg.ExtractDir)
	if err != nil {
		return err
	}
	t.dbPath = dbPath

	slog.Info("Found ITIS database", "path", dbPath)

	// Open the database connection.
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// Set SQLite pragmas for better performance.
	_, err = db.Exec("PRAGMA temp_store = MEMORY")
	if err != nil {
		return err
	}

	t.db = db

	// Load extinct TSNs from GitHub.
	slog.Info("Loading extinct TSNs from GitHub")
	err = t.loadExtinctTSNs()
	if err != nil {
		slog.Warn("Could not load extinct TSNs", "error", err)
		// Not fatal - continue without extinction data.
	} else {
		slog.Info("Loaded extinct TSNs", "count", len(t.extinct))
	}

	return nil
}

// loadExtinctTSNs downloads and parses the extinct.tsv file from GitHub.
func (t *itis) loadExtinctTSNs() error {
	resp, err := http.Get(extinctURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	scanner := bufio.NewScanner(resp.Body)

	// Skip header line.
	if scanner.Scan() {
		// First line is "tsn" header.
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		tsn, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		t.extinct[tsn] = true
	}

	return scanner.Err()
}

// findSQLiteDB searches for the SQLite database file in the given directory.
func findSQLiteDB(dir string) (string, error) {
	var dbPath string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sqlite") {
			dbPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return "", err
	}
	if dbPath == "" {
		return "", os.ErrNotExist
	}
	return dbPath, nil
}
