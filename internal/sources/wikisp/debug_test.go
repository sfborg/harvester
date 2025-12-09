package wikisp

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/harvester/internal/sources/wikisp/wsparser"
)

func TestParseAnthozoa(t *testing.T) {
	// Read the test file
	data, err := os.ReadFile("../../../testdata/wikisp_pages.xml")
	if err != nil {
		t.Fatal(err)
	}

	// Find Anthozoa page
	pages := splitPages(string(data))
	var anthozoaPage string
	for _, page := range pages {
		if contains(page, "<title>Anthozoa</title>") {
			anthozoaPage = page
			break
		}
	}

	if anthozoaPage == "" {
		t.Fatal("Anthozoa page not found")
	}

	// Parse XML
	var page PageXML
	if err := xml.Unmarshal([]byte(anthozoaPage), &page); err != nil {
		t.Fatalf("XML unmarshal failed: %v", err)
	}

	t.Logf("Title: %s", page.Title)
	t.Logf("ID: %d", page.ID)
	t.Logf("Text length: %d", len(page.Revision.Text.Content))
	t.Logf(
		"First 500 chars:\n%s",
		page.Revision.Text.Content[:min(500, len(page.Revision.Text.Content))],
	)

	// Parse sections
	sections := parseSections(page.Revision.Text.Content)
	t.Logf("Found %d sections", len(sections))
	for header := range sections {
		t.Logf("  Section: %s", header)
	}

	// Check if it's a taxon page
	if !isTaxonPage(&page, sections) {
		t.Fatal("Not recognized as taxon page")
	}

	// Initialize gnparser
	gnp := gnparser.New(gnparser.NewConfig(
		gnparser.OptWithDetails(true),
		gnparser.OptCode(nomcode.Unknown),
	))

	// Initialize wsparser with gnparser fallback
	wsp := wsparser.New(&gnparserAdapter{gnp: gnp})

	// Try to extract data
	pd, err := extractPageData(&page, wsp)
	if err != nil {
		t.Fatalf("extractPageData failed: %v", err)
	}

	t.Logf("Scientific name: %s", pd.ScientificName)
	t.Logf("Parent template: %s", pd.ParentTemplate)
	t.Logf("Vernacular names: %d", len(pd.VernacularNames))
}

func TestBASEPAGENAME(t *testing.T) {
	// Read the test file
	data, err := os.ReadFile("../../../testdata/wikisp_pages.xml")
	if err != nil {
		t.Fatal(err)
	}

	// Find Template:Clowesia page (has BASEPAGENAME)
	pages := splitPages(string(data))
	var targetPage string
	for _, page := range pages {
		if contains(page, "<title>Template:Clowesia</title>") {
			targetPage = page
			break
		}
	}

	if targetPage == "" {
		t.Fatal("Template:Clowesia page not found")
	}

	// Parse XML
	var page PageXML
	if err := xml.Unmarshal([]byte(targetPage), &page); err != nil {
		t.Fatalf("XML unmarshal failed: %v", err)
	}

	t.Logf("Title: %s", page.Title)
	t.Logf("Content before substitution:\n%s", page.Revision.Text.Content)

	// Verify BASEPAGENAME is in content
	if !contains(page.Revision.Text.Content, "{{BASEPAGENAME}}") {
		t.Fatal("BASEPAGENAME not found in content")
	}

	// Extract BASEPAGENAME from title
	baseName := page.Title
	if idx := findInString(page.Title, ":"); idx != -1 {
		baseName = page.Title[idx+1:]
	}

	t.Logf("Extracted base name: %s", baseName)

	if baseName != "Clowesia" {
		t.Errorf("Expected base name 'Clowesia', got '%s'", baseName)
	}

	// Substitute
	contentAfter := ""
	for i := 0; i < len(page.Revision.Text.Content); i++ {
		if i+len("{{BASEPAGENAME}}") <= len(page.Revision.Text.Content) &&
			page.Revision.Text.Content[i:i+len("{{BASEPAGENAME}}")] == "{{BASEPAGENAME}}" {
			contentAfter += baseName
			i += len("{{BASEPAGENAME}}") - 1
		} else {
			contentAfter += string(page.Revision.Text.Content[i])
		}
	}

	t.Logf("Content after substitution:\n%s", contentAfter)

	// Verify substitution worked
	if contains(contentAfter, "{{BASEPAGENAME}}") {
		t.Error("BASEPAGENAME still present after substitution")
	}

	if !contains(contentAfter, "Clowesia") {
		t.Error("Clowesia not found after substitution")
	}

	t.Log("BASEPAGENAME substitution test passed!")
}

func splitPages(content string) []string {
	// Simple split by <page> tags
	var pages []string
	var current []rune
	inPage := false

	for _, line := range split(content, '\n') {
		if contains(line, "<page>") {
			inPage = true
			current = []rune{}
		}
		if inPage {
			current = append(current, []rune(line)...)
			current = append(current, '\n')
		}
		if contains(line, "</page>") {
			pages = append(pages, string(current))
			inPage = false
		}
	}
	return pages
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findInString(s, substr) >= 0
}

func findInString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func split(s string, sep rune) []string {
	var result []string
	var current []rune
	for _, r := range s {
		if r == sep {
			result = append(result, string(current))
			current = []rune{}
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
