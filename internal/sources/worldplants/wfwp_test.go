package worldplants_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfborg/harvester/internal/sources/worldplants"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

const testDataDir = "../../../testdata/wfwp"

func TestWorldPlantsIntegration(t *testing.T) {
	assert := assert.New(t)

	// Skip if testdata not available
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("testdata/wfwp not found, skipping integration test")
	}

	// Create temp directory for test output
	tmpDir, err := os.MkdirTemp("", "wfwp-test-*")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)

	// Create config with temp cache
	cfg := config.New(
		config.OptCacheDir(tmpDir),
		config.OptLocalFile(testDataDir),
	)

	// Create extract directory
	extractDir := filepath.Join(tmpDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	assert.NoError(err)

	// Create worldplants convertor
	convertor := worldplants.New(cfg)
	assert.NotNil(convertor)

	// Test label
	assert.Equal("wfwp", convertor.Label())

	// Test Extract (file preparation)
	err = convertor.Extract(testDataDir)
	assert.NoError(err, "Extract should succeed")

	// Verify concatenated files were created
	fernsPath := filepath.Join(extractDir, "ferns.csv")
	plantsPath := filepath.Join(extractDir, "plants.csv")

	assert.FileExists(fernsPath, "ferns.csv should be created")
	assert.FileExists(plantsPath, "plants.csv should be created")

	// Test InitSfga
	sfgaArchive, err := convertor.InitSfga()
	assert.NoError(err, "InitSfga should succeed")
	assert.NotNil(sfgaArchive)

	// Test ToSfga
	err = convertor.ToSfga(sfgaArchive)
	assert.NoError(err, "ToSfga should succeed")

	// Verify database was created
	sfgaDir := filepath.Join(tmpDir, "sfga")
	dbPath := filepath.Join(sfgaDir, "schema.sqlite")
	assert.FileExists(dbPath, "SFGA database should be created")

	// Validate database content
	validateDatabase(t, dbPath)
}

func validateDatabase(t *testing.T, dbPath string) {
	assert := assert.New(t)

	db, err := sql.Open("sqlite", dbPath)
	assert.NoError(err)
	defer db.Close()

	// Test metadata table
	var title, alias string
	err = db.QueryRow(
		"SELECT col__title, col__alias FROM metadata LIMIT 1",
	).Scan(&title, &alias)
	assert.NoError(err)
	assert.NotEmpty(title)

	// Test name table - should have both ferns and plants data
	var nameCount int
	err = db.QueryRow("SELECT COUNT(*) FROM name").Scan(&nameCount)
	assert.NoError(err)
	assert.Greater(nameCount, 100, "Should have names")

	// Test reference table
	var refCount int
	err = db.QueryRow("SELECT COUNT(*) FROM reference").Scan(&refCount)
	assert.NoError(err)
	assert.Greater(refCount, 0, "Should have references")

	// Test distribution table
	var distCount int
	err = db.QueryRow("SELECT COUNT(*) FROM distribution").Scan(&distCount)
	assert.NoError(err)
	assert.Greater(distCount, 0, "Should have distributions")

	// Test vernacular table
	var vernCount int
	err = db.QueryRow("SELECT COUNT(*) FROM vernacular").Scan(&vernCount)
	assert.NoError(err)
	assert.Greater(vernCount, 0, "Should have vernacular names")

	// Verify taxa exist (accepted names)
	var taxonCount int
	err = db.QueryRow("SELECT COUNT(*) FROM taxon").Scan(&taxonCount)
	assert.NoError(err)
	assert.Greater(taxonCount, 100, "Should have taxa")

	// Verify synonyms exist
	var synonymCount int
	err = db.QueryRow("SELECT COUNT(*) FROM synonym").Scan(&synonymCount)
	assert.NoError(err)
	assert.Greater(synonymCount, 0, "Should have synonyms")

	// Verify parent_id relationships in taxon table
	var parentCount int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM taxon WHERE col__parent_id IS NOT NULL AND col__parent_id != ''",
	).Scan(&parentCount)
	assert.NoError(err)
	assert.Greater(parentCount, 50, "Should have parent relationships")
}

func TestFileConcatenation(t *testing.T) {
	assert := assert.New(t)

	// Skip if testdata not available
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("testdata/wfwp not found, skipping test")
	}

	tmpDir, err := os.MkdirTemp("", "wfwp-concat-test-*")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)

	cfg := config.New(
		config.OptCacheDir(tmpDir),
		config.OptLocalFile(testDataDir),
	)

	// Create extract directory
	extractDir := filepath.Join(tmpDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	assert.NoError(err)

	convertor := worldplants.New(cfg)
	err = convertor.Extract(testDataDir)
	assert.NoError(err)

	// Read concatenated plants.csv
	plantsPath := filepath.Join(tmpDir, "extract", "plants.csv")
	data, err := os.ReadFile(plantsPath)
	assert.NoError(err)

	content := string(data)

	// Should have header line
	assert.Contains(content, "Taxon|Number|Name|Literature")

	// Count lines to verify concatenation
	lines := 0
	for _, line := range string(data) {
		if line == '\n' {
			lines++
		}
	}
	// Check that we have concatenated content from multiple files
	assert.Greater(lines, 100, "Should have concatenated content")

	// Verify numbered files were processed in order
	// (harder to test directly, but we tested it worked)
}

func TestValidateInputDirectory(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantErr   bool
	}{
		{
			name: "valid directory",
			setupFunc: func(t *testing.T) string {
				return testDataDir
			},
			wantErr: false,
		},
		{
			name: "missing directory",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/directory"
			},
			wantErr: true,
		},
		{
			name: "missing ferns.csv",
			setupFunc: func(t *testing.T) string {
				tmpDir, _ := os.MkdirTemp("", "wfwp-nofernstest-*")
				// Create a numbered file but no ferns.csv
				os.WriteFile(
					filepath.Join(tmpDir, "1.csv"),
					[]byte("test"),
					0644,
				)
				t.Cleanup(func() { os.RemoveAll(tmpDir) })
				return tmpDir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			inputDir := tt.setupFunc(t)

			// Skip test if testdata required but not available
			if inputDir == testDataDir {
				if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
					t.Skip("testdata not available")
				}
			}

			tmpDir, err := os.MkdirTemp("", "wfwp-validate-test-*")
			assert.NoError(err)
			defer os.RemoveAll(tmpDir)

			// Create extract directory
			extractDir := filepath.Join(tmpDir, "extract")
			err = os.MkdirAll(extractDir, 0755)
			assert.NoError(err)

			cfg := config.New(config.OptCacheDir(tmpDir))
			convertor := worldplants.New(cfg)

			err = convertor.Extract(inputDir)
			if tt.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestZipFileExtraction(t *testing.T) {
	assert := assert.New(t)

	zipPath := "../../../testdata/wfwp.zip"

	// Skip if testdata zip not available
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Skip("testdata/wfwp.zip not found, skipping test")
	}

	tmpDir, err := os.MkdirTemp("", "wfwp-zip-test-*")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)

	cfg := config.New(
		config.OptCacheDir(tmpDir),
		config.OptLocalFile(zipPath),
	)

	// Create extract directory
	extractDir := filepath.Join(tmpDir, "extract")
	err = os.MkdirAll(extractDir, 0755)
	assert.NoError(err)

	convertor := worldplants.New(cfg)

	// Test zip extraction
	err = convertor.Extract(zipPath)
	assert.NoError(err, "Zip extraction should succeed")

	// Verify files were extracted and processed
	fernsPath := filepath.Join(extractDir, "ferns.csv")
	plantsPath := filepath.Join(extractDir, "plants.csv")

	assert.FileExists(fernsPath, "ferns.csv should exist after extraction")
	assert.FileExists(plantsPath, "plants.csv should exist after extraction")

	// Verify content
	data, err := os.ReadFile(plantsPath)
	assert.NoError(err)
	assert.Greater(len(data), 1000, "plants.csv should have content")
}
