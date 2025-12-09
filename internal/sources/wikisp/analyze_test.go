package wikisp

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestAnalyzeHierarchy analyzes the test data to show missing templates/parents
func TestAnalyzeHierarchy(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/wikisp_pages.xml")
	if err != nil {
		t.Fatal(err)
	}

	// Parse all pages
	content := string(data)
	pages := strings.Split(content, "<page>")

	templates := make(map[string]string)    // template name -> page ID
	taxonPages := make(map[string]string)   // taxon title -> page ID
	parentRefs := make(map[string][]string) // parent template -> list of children

	for _, pageStr := range pages {
		if !strings.Contains(pageStr, "</page>") {
			continue
		}

		pageStr = "<page>" + pageStr

		var page PageXML
		if err := xml.Unmarshal([]byte(pageStr), &page); err != nil {
			continue
		}

		// Collect templates
		if strings.HasPrefix(page.Title, "Template:") {
			templateName := strings.TrimPrefix(page.Title, "Template:")
			templates[templateName] = fmt.Sprintf("%d", page.ID)
			fmt.Printf("Found template: %s (ID: %d)\n", templateName, page.ID)
		}

		// Skip redirects
		if isRedirect(&page) {
			continue
		}

		// Parse taxon pages
		sections := parseSections(page.Revision.Text.Content)
		if isTaxonPage(&page, sections) {
			taxonPages[page.Title] = fmt.Sprintf("%d", page.ID)

			// Extract parent reference
			if taxonavSection, ok := sections["{{int:taxonavigation}}"]; ok {
				parentTemplate := extractParentTemplate(taxonavSection)
				if parentTemplate != "" {
					parentRefs[parentTemplate] = append(parentRefs[parentTemplate], page.Title)
					fmt.Printf("Taxon: %s (ID: %d) -> Parent template: %s\n",
						page.Title, page.ID, parentTemplate)
				}
			}
		}
	}

	fmt.Println("\n=== SUMMARY ===")
	fmt.Printf("Templates found: %d\n", len(templates))
	fmt.Printf("Taxon pages found: %d\n", len(taxonPages))
	fmt.Printf("Unique parent references: %d\n\n", len(parentRefs))

	// Check which parent templates are missing
	fmt.Println("=== MISSING PARENT TEMPLATES ===")
	for parentTemplate, children := range parentRefs {
		// Check if it exists as a template
		if _, ok := templates[parentTemplate]; ok {
			fmt.Printf("✓ %s (found as template)\n", parentTemplate)
			continue
		}

		// Check if it exists as a taxon page
		if _, ok := taxonPages[parentTemplate]; ok {
			fmt.Printf("✓ %s (found as taxon page)\n", parentTemplate)
			continue
		}

		// It's missing!
		fmt.Printf("✗ %s (MISSING) - needed by %d taxa:\n", parentTemplate, len(children))
		for _, child := range children {
			fmt.Printf("  - %s\n", child)
		}
	}

	fmt.Println("\n=== AVAILABLE TEMPLATES ===")
	for name, id := range templates {
		fmt.Printf("  %s (ID: %s)\n", name, id)
	}

	fmt.Println("\n=== TAXON PAGES (potential parent targets) ===")
	for title, id := range taxonPages {
		fmt.Printf("  %s (ID: %s)\n", title, id)
	}
}
