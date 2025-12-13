package itis_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfborg/harvester/internal/sources/itis"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

const testDataDir = "../../../testdata/itis"

func TestITISIntegration(t *testing.T) {
	assert := assert.New(t)

	// Skip if testdata not available.
	testDB := filepath.Join(testDataDir, "ITIS.sqlite")
	if _, err := os.Stat(testDB); os.IsNotExist(err) {
		t.Skip("testdata/itis/ITIS.sqlite not found, skipping integration test")
	}

	// Create temp directory for test output.
	tmpDir, err := os.MkdirTemp("", "itis-test-*")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)

	// Create directory structure expected by harvester.
	extractDir := filepath.Join(tmpDir, "extract", "itisSqlite")
	err = os.MkdirAll(extractDir, 0755)
	assert.NoError(err)

	// Copy test database to extract directory.
	testData, err := os.ReadFile(testDB)
	assert.NoError(err)
	err = os.WriteFile(filepath.Join(extractDir, "ITIS.sqlite"), testData, 0644)
	assert.NoError(err)

	// Create config with temp cache.
	cfg := config.New(
		config.OptCacheDir(tmpDir),
		config.OptSkipDownload(true),
	)

	// Create ITIS convertor.
	convertor := itis.New(cfg)
	assert.NotNil(convertor)

	// Test label.
	assert.Equal("itis", convertor.Label())

	// Test Extract (opens database, loads extinct TSNs).
	err = convertor.Extract("")
	assert.NoError(err, "Extract should succeed")

	// Test InitSfga.
	sfgaArchive, err := convertor.InitSfga()
	assert.NoError(err, "InitSfga should succeed")
	assert.NotNil(sfgaArchive)

	// Test ToSfga.
	err = convertor.ToSfga(sfgaArchive)
	assert.NoError(err, "ToSfga should succeed")

	// Verify database was created.
	sfgaDir := filepath.Join(tmpDir, "sfga")
	dbPath := filepath.Join(sfgaDir, "schema.sqlite")
	assert.FileExists(dbPath, "SFGA database should be created")

	// Validate database content.
	validateDatabase(t, dbPath)
}

func validateDatabase(t *testing.T, dbPath string) {
	assert := assert.New(t)

	db, err := sql.Open("sqlite", dbPath)
	assert.NoError(err)
	defer db.Close()

	// Test metadata table.
	var title string
	err = db.QueryRow("SELECT col__title FROM metadata LIMIT 1").Scan(&title)
	assert.NoError(err)
	assert.Contains(title, "ITIS")

	// Test name table.
	var nameCount int
	err = db.QueryRow("SELECT COUNT(*) FROM name").Scan(&nameCount)
	assert.NoError(err)
	assert.Greater(nameCount, 100, "Should have names from all kingdoms")

	// Test taxon table.
	var taxonCount int
	err = db.QueryRow("SELECT COUNT(*) FROM taxon").Scan(&taxonCount)
	assert.NoError(err)
	assert.Greater(taxonCount, 100, "Should have taxa")

	// Test synonym table.
	var synonymCount int
	err = db.QueryRow("SELECT COUNT(*) FROM synonym").Scan(&synonymCount)
	assert.NoError(err)
	assert.Greater(synonymCount, 0, "Should have synonyms")

	// Test vernacular table.
	var vernCount int
	err = db.QueryRow("SELECT COUNT(*) FROM vernacular").Scan(&vernCount)
	assert.NoError(err)
	assert.Greater(vernCount, 0, "Should have vernacular names")

	// Test distribution table.
	var distCount int
	err = db.QueryRow("SELECT COUNT(*) FROM distribution").Scan(&distCount)
	assert.NoError(err)
	assert.Greater(distCount, 0, "Should have distributions")

	// Test reference table.
	var refCount int
	err = db.QueryRow("SELECT COUNT(*) FROM reference").Scan(&refCount)
	assert.NoError(err)
	assert.Greater(refCount, 0, "Should have references")

	// Verify extinct taxa are marked.
	var extinctCount int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM taxon WHERE col__extinct = 1",
	).Scan(&extinctCount)
	assert.NoError(err)
	assert.Greater(extinctCount, 0, "Should have extinct taxa marked")

	// Verify hierarchy integrity - no orphaned taxa.
	var orphanCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM taxon t
		WHERE col__parent_id != ''
		AND col__parent_id NOT IN (SELECT col__id FROM taxon)
		AND col__parent_id != '0'
	`).Scan(&orphanCount)
	assert.NoError(err)
	assert.Equal(0, orphanCount, "Should have no orphaned taxa")

	// Verify all synonyms point to valid taxa.
	var invalidSynonymCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM synonym s
		WHERE s.col__taxon_id NOT IN (SELECT col__id FROM taxon)
	`).Scan(&invalidSynonymCount)
	assert.NoError(err)
	assert.Equal(0, invalidSynonymCount, "All synonyms should reference valid taxa")

	// Verify all synonyms have corresponding names.
	var missingSynonymNames int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM synonym s
		WHERE s.col__name_id NOT IN (SELECT col__id FROM name)
	`).Scan(&missingSynonymNames)
	assert.NoError(err)
	assert.Equal(0, missingSynonymNames, "All synonyms should have names")

	// Verify GNparser data is populated.
	var parsedCount int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM name WHERE gn__canonical_simple != ''",
	).Scan(&parsedCount)
	assert.NoError(err)
	assert.Greater(parsedCount, 50, "Most names should be parsed")
}

func TestITISLabel(t *testing.T) {
	cfg := config.New()
	convertor := itis.New(cfg)
	assert.Equal(t, "itis", convertor.Label())
	assert.Equal(t, "Integrated Taxonomic Information System", convertor.Name())
}
