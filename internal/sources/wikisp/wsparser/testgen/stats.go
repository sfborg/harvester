package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// ShowStats displays quality distribution from a test file.
func ShowStats(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	qualityCounts := make(map[int]int)
	var total int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var tc TestCase
		if err := json.Unmarshal([]byte(line), &tc); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		qualityCounts[tc.Output.Quality]++
		total++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Printf("Quality Distribution (%d total):\n", total)
	fmt.Println("================================")
	fmt.Printf("  0 (Unparseable): %6d (%5.1f%%)\n",
		qualityCounts[0], 100.0*float64(qualityCounts[0])/float64(total))
	fmt.Printf("  1 (Partial):     %6d (%5.1f%%)\n",
		qualityCounts[1], 100.0*float64(qualityCounts[1])/float64(total))
	fmt.Printf("  2 (Good):        %6d (%5.1f%%)\n",
		qualityCounts[2], 100.0*float64(qualityCounts[2])/float64(total))
	fmt.Printf("  3 (Complete):    %6d (%5.1f%%)\n",
		qualityCounts[3], 100.0*float64(qualityCounts[3])/float64(total))

	return nil
}
