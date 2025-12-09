package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sfborg/harvester/internal/sources/wikisp/wsparser"
)

// TestCase represents a single test case with input and expected output.
type TestCase struct {
	Input  string     `json:"in"`
	Output TestOutput `json:"out"`
}

// TestOutput represents the expected output from parsing.
type TestOutput struct {
	Canonical  string `json:"canonical,omitempty"`
	Authorship string `json:"authorship,omitempty"`
	Reference  string `json:"reference,omitempty"`
	Tail       string `json:"tail,omitempty"`
	Quality    int    `json:"quality"`
	Error      string `json:"error,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s <input.txt> <output.jsonl>  - Generate test data\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s stats <test.jsonl>          - Show quality stats\n", os.Args[0])
		os.Exit(1)
	}

	// Check for stats command
	if os.Args[1] == "stats" {
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s stats <test.jsonl>\n", os.Args[0])
			os.Exit(1)
		}
		if err := ShowStats(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Generate mode
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.txt> <output.jsonl>\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Open input file
	inFile, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input: %v\n", err)
		os.Exit(1)
	}
	defer inFile.Close()

	// Create output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(inFile)
	encoder := json.NewEncoder(outFile)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		testCase := TestCase{
			Input: line,
		}

		// Parse the input
		result, err := wsparser.Parse(line)
		if err != nil {
			testCase.Output.Error = "parse failed"
		}

		// Always capture quality and parsed fields
		testCase.Output.Canonical = result.Canonical
		testCase.Output.Authorship = result.Authorship
		testCase.Output.Reference = result.Reference
		testCase.Output.Tail = result.Tail
		testCase.Output.Quality = int(result.Quality)

		// Write as JSON line
		if err := encoder.Encode(testCase); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding line %d: %v\n",
				lineNum, err)
			os.Exit(1)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d test cases from %s -> %s\n",
		lineNum, inputFile, outputFile)
}
