package wikisp

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfborg/harvester/pkg/config"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// TestFullParsing tests the complete parsing pipeline with the test data
func TestFullParsing(t *testing.T) {
	// Create temp directory for output
	tmpDir := t.TempDir()
	extractDir := filepath.Join(tmpDir, "extract")

	// Copy test file to extract dir
	err := os.MkdirAll(extractDir, 0755)
	assert.Nil(t, err)

	testData, err := os.ReadFile("../../../testdata/wikisp_pages.xml")
	assert.Nil(t, err)

	err = os.WriteFile(
		filepath.Join(extractDir, "wikisp_pages.xml"),
		testData, 0644,
	)
	assert.Nil(t, err)

	// Create config
	cfg := config.New(
		config.OptCacheDir(tmpDir),
	)

	// Create wikisp convertor
	convertor := New(cfg)

	// Extract (pass the XML file path, not the directory)
	xmlPath := filepath.Join(extractDir, "wikisp_pages.xml")
	if err := convertor.Extract(xmlPath); err != nil {
		t.Fatal(err)
	}

	// Initialize SFGA
	sfgaArchive, err := convertor.InitSfga()
	if err != nil {
		t.Fatal(err)
	}

	// Convert to SFGA
	if err := convertor.ToSfga(sfgaArchive); err != nil {
		t.Fatalf("ToSfga failed: %v", err)
	}

	// Open database for validation
	dbPath := filepath.Join(tmpDir, "sfga", "schema.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// First, let's see what tables exist
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n=== AVAILABLE TABLES ===")
	var tables []string
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
		t.Logf("  %s", table)
	}
	rows.Close()

	// Show name table structure
	t.Logf("\n=== NAME TABLE STRUCTURE ===")
	rows, _ = db.Query("PRAGMA table_info(name)")
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt sql.NullString
		rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk)
		if cid < 15 { // Show first 15 columns
			t.Logf("  %s (%s)", name, typ)
		}
	}
	rows.Close()

	// Check metadata table
	t.Logf("\n=== METADATA CHECK ===")
	var metaTitle, metaDescription, metaURL, metaLicense sql.NullString
	err = db.QueryRow("SELECT col__title, col__description, col__url, col__license FROM metadata LIMIT 1").
		Scan(&metaTitle, &metaDescription, &metaURL, &metaLicense)
	if err == nil {
		t.Logf("Title: %s", metaTitle.String)
		t.Logf("Description: %s", metaDescription.String)
		t.Logf("URL: %s", metaURL.String)
		t.Logf("License: %s", metaLicense.String)

		if metaTitle.String != "Wikispecies" {
			t.Errorf("Expected title 'Wikispecies', got '%s'", metaTitle.String)
		}
	} else {
		t.Errorf("Failed to read metadata: %v", err)
	}

	// Check col__scientific_name is populated
	t.Logf("\n=== COL__SCIENTIFIC_NAME CHECK ===")
	var countWithScientificName, countTotal int
	db.QueryRow("SELECT COUNT(*) FROM name WHERE col__scientific_name IS NOT NULL AND col__scientific_name != ''").
		Scan(&countWithScientificName)
	db.QueryRow("SELECT COUNT(*) FROM name").Scan(&countTotal)
	t.Logf("Records with col__scientific_name: %d / %d", countWithScientificName, countTotal)

	if countWithScientificName != countTotal {
		t.Errorf(
			"Expected all %d records to have col__scientific_name populated, but only %d do",
			countTotal,
			countWithScientificName,
		)
	}

	// Show sample records to verify content
	rows, _ = db.Query(
		"SELECT col__scientific_name, gn__canonical_full, gn__scientific_name_string FROM name LIMIT 5",
	)
	t.Logf("\nSample records:")
	for rows.Next() {
		var sciName, canonicalFull, fullString sql.NullString
		rows.Scan(&sciName, &canonicalFull, &fullString)
		t.Logf("  col__scientific_name: '%s'", sciName.String)
		t.Logf("    gn__canonical_full: '%s'", canonicalFull.String)
		t.Logf("    gn__scientific_name_string: '%s'", fullString.String)
	}
	rows.Close()

	// Count total records in main tables
	var totalTaxon, totalSynonym, totalVernacular int
	if err := db.QueryRow("SELECT COUNT(*) FROM taxon").Scan(&totalTaxon); err == nil {
		t.Logf("\nTotal taxon records: %d", totalTaxon)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM synonym").Scan(&totalSynonym); err == nil {
		t.Logf("Total synonym records: %d", totalSynonym)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM vernacular").Scan(&totalVernacular); err == nil {
		t.Logf("Total vernacular records: %d", totalVernacular)
	}

	t.Logf("\n=== PARSING SUMMARY ===")
	t.Logf("Taxon records: %d", totalTaxon)
	t.Logf("Synonym records: %d", totalSynonym)
	t.Logf("Vernacular records: %d", totalVernacular)
	t.Logf("Total: %d", totalTaxon+totalSynonym)

	// Show classification hierarchy with ranks
	t.Logf("\n=== CLASSIFICATION BY RANK ===")

	// Get rank information from the rank table
	rankMap := make(map[string]string)
	rows, err = db.Query("SELECT id, id FROM rank") // rank table has rank names
	if err == nil {
		for rows.Next() {
			var id, name string
			rows.Scan(&id, &name)
			rankMap[id] = name
		}
		rows.Close()
	}

	// Query taxa grouped by rank
	rows, err = db.Query(`
		SELECT DISTINCT
			t.col__genus,
			t.col__family,
			t.col__order,
			t.col__class,
			t.col__phylum,
			t.col__kingdom,
			n.gn__scientific_name_string,
			n.gn__canonical_simple
		FROM taxon t
		JOIN name n ON t.col__name_id = n.col__id
		WHERE n.gn__canonical_simple IS NOT NULL
		ORDER BY t.col__phylum, t.col__class, t.col__order, t.col__family, t.col__genus
		LIMIT 20
	`)
	if err == nil {
		t.Logf("\nHigher classification:")
		for rows.Next() {
			var genus, family, order, class, phylum, kingdom sql.NullString
			var fullName, canonical string
			rows.Scan(&genus, &family, &order, &class, &phylum, &kingdom, &fullName, &canonical)

			t.Logf("  %s", canonical)
			if kingdom.Valid && kingdom.String != "" {
				t.Logf("    Kingdom: %s", kingdom.String)
			}
			if phylum.Valid && phylum.String != "" {
				t.Logf("    Phylum: %s", phylum.String)
			}
			if class.Valid && class.String != "" {
				t.Logf("    Class: %s", class.String)
			}
			if order.Valid && order.String != "" {
				t.Logf("    Order: %s", order.String)
			}
			if family.Valid && family.String != "" {
				t.Logf("    Family: %s", family.String)
			}
			if genus.Valid && genus.String != "" {
				t.Logf("    Genus: %s", genus.String)
			}
		}
		rows.Close()
	} else {
		t.Logf("Error querying classification: %v", err)
	}

	// Show parent-child relationships
	t.Logf("\n=== PARENT-CHILD HIERARCHY ===")
	rows, err = db.Query(`
		SELECT
			t.col__id as taxon_id,
			n1.gn__scientific_name_string as taxon_name,
			t.col__parent_id as parent_id
		FROM taxon t
		JOIN name n1 ON t.col__name_id = n1.col__id
		WHERE t.col__parent_id IS NOT NULL
			AND t.col__parent_id != ''
		ORDER BY taxon_name
		LIMIT 10
	`)
	if err == nil {
		for rows.Next() {
			var taxonID, taxonName, parentID string
			rows.Scan(&taxonID, &taxonName, &parentID)

			// Check if parent exists as a taxon
			var parentName sql.NullString
			db.QueryRow(`
				SELECT n.gn__scientific_name_string
				FROM taxon t
				JOIN name n ON t.col__name_id = n.col__id
				WHERE t.col__id = ?
			`, parentID).Scan(&parentName)

			if parentName.Valid && parentName.String != "" {
				t.Logf("  ✓ %s → %s", taxonName, parentName.String)
			} else {
				t.Logf("  ✗ %s (parent_id=%s not found)", taxonName, parentID)
			}
		}
		rows.Close()
	} else {
		t.Logf("Error: %v", err)
	}
}
