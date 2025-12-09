package wsparser_test

import (
	"fmt"
	"github.com/sfborg/harvester/internal/sources/wikisp/wsparser"
)

func ExampleParse() {
	// Parse a Wikispecies name with template author (full and short forms)
	input := "''Felis catus'' {{a|Carl Linnaeus|L.}}, 1758: Systema Naturae"
	result, err := wsparser.Parse(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Canonical: %s\n", result.Canonical)
	fmt.Printf("Authorship: %s\n", result.Authorship)
	fmt.Printf("Reference: %s\n", result.Reference)

	// Output:
	// Canonical: Felis catus
	// Authorship: L., 1758
	// Reference: : Systema Naturae
}
