package mdd

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/sfborg/sflib/pkg/coldp"
)

func (m *mdd) importMeta() error {
	meta := coldp.Meta{
		Title: "Mammal Diversity Database",
		Description: "The Mammal Diversity Database (MDD) is a continuously " +
			"updated, species-level consensus mammal taxonomy, with all species " +
			"of currently recognized extant and recently extinct mammals.",
		URL:             "https://www.mammaldiversity.org",
		DOI:             "https://doi.org/10.5281/zenodo.4139722",
		TaxonomicScope:  "Mammalia",
		GeographicScope: "global",
	}

	tomlPath := filepath.Join(m.cfg.ExtractDir, "MDD", "release.toml")
	if f, err := os.Open(tomlPath); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			k, v, ok := parseTomlString(scanner.Text())
			if !ok {
				continue
			}
			switch k {
			case "version":
				meta.Version = v
			case "release_date":
				meta.Issued = v
			case "zenodo_citation":
				meta.Citation = v
			}
		}
	}

	m.sfga.InsertMeta(&meta)
	return nil
}

func parseTomlString(line string) (key, val string, ok bool) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
		return "", "", false
	}
	var found bool
	key, val, found = strings.Cut(line, "=")
	if !found {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	val = strings.Trim(strings.TrimSpace(val), `"`)
	return key, val, true
}
