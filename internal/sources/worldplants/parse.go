package worldplants

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/sfborg/sflib/pkg/coldp"
)

// gnparsed holds parsed scientific name components.
type gnparsed struct {
	quality               int
	cardinality           int
	verbatim              string
	canonicalFull         string
	canonicalSimple       string
	canonicalStemmed      string
	authorship            string
	originalAuthorship    string
	combinationAuthorship string
	uninomial             string
	genus                 string
	subgenus              string
	species               string
	infraspecies          string
	notho                 coldp.NamePart
	rank                  string
	remarks               string
	nameStatus            coldp.NomStatus
	appendedPhrase        string
	unparsedTail          string
}

func formatAuthors(authorship []string) string {
	result := ""
	for i, author := range authorship {
		authorship[i] = strings.ReplaceAll(author, ". ", ".")
	}

	switch len(authorship) {
	case 1:
		result = authorship[0]
	case 2:
		result = strings.Join(authorship, " & ")
	default:
		result = strings.Join(
			authorship[0:len(authorship)-1],
			", ",
		)
		result = result + " & " + authorship[len(authorship)-1]
	}
	return result
}

func wfwpParse(
	parser gnparser.GNparser,
	name string,
	citation string,
) (gnparsed, error) {
	parsedName := parser.ParseName(name)
	if !parsedName.Parsed {
		return gnparsed{}, fmt.Errorf(
			"Failed to parse name: %s",
			name,
		)
	}

	result := gnparsed{
		verbatim:         parsedName.Verbatim,
		quality:          parsedName.ParseQuality,
		cardinality:      parsedName.Cardinality,
		canonicalFull:    parsedName.Canonical.Full,
		canonicalSimple:  parsedName.Canonical.Simple,
		canonicalStemmed: parsedName.Canonical.Stemmed,
	}

	// handle manuscript names
	if strings.Contains(name, "comb. ined.") ||
		strings.Contains(citation, "comb. ined.") {
		result.nameStatus = coldp.Manuscript
		result.appendedPhrase = "comb. ined."
		name = strings.ReplaceAll(name, "comb. ined.", "")
	} else if strings.Contains(name, "ined.") ||
		strings.Contains(citation, "ined.") {
		result.nameStatus = coldp.Manuscript
		result.appendedPhrase = "ined."
		name = strings.ReplaceAll(name, "ined.", "")
	}

	if parsedName.Authorship != nil &&
		parsedName.Authorship.Original != nil {
		result.originalAuthorship = formatAuthors(
			parsedName.Authorship.Original.Authors,
		)

		if parsedName.Authorship.Original.ExAuthors != nil {
			result.originalAuthorship = result.originalAuthorship +
				" ex " + formatAuthors(
				parsedName.Authorship.Original.ExAuthors.Authors,
			)
		} else if parsedName.Authorship.Original.InAuthors != nil {
			result.originalAuthorship = result.originalAuthorship +
				" in " + formatAuthors(
				parsedName.Authorship.Original.InAuthors.Authors,
			)
		}
		result.authorship = result.originalAuthorship
	}

	if parsedName.Authorship != nil &&
		parsedName.Authorship.Combination != nil {
		result.combinationAuthorship = formatAuthors(
			parsedName.Authorship.Combination.Authors,
		)

		if parsedName.Authorship.Combination.ExAuthors != nil {
			result.combinationAuthorship = result.combinationAuthorship +
				" ex " + formatAuthors(
				parsedName.Authorship.Combination.ExAuthors.Authors,
			)
		} else if parsedName.Authorship.Combination.InAuthors != nil {
			result.combinationAuthorship = result.combinationAuthorship +
				" in " + formatAuthors(
				parsedName.Authorship.Combination.InAuthors.Authors,
			)
		}
		result.authorship = "(" + result.originalAuthorship + ") " +
			result.combinationAuthorship
	}

	if parsedName.Hybrid != nil &&
		parsedName.Hybrid.String() == "HYBRID_FORMULA" {
		return gnparsed{}, fmt.Errorf(
			"Hybrid formulas are not allowed: %s",
			name,
		)
	}

	if parsedName.Rank != "" {
		result.rank = parsedName.Rank
	}

	switch detail := parsedName.Details.(type) {
	case parsed.DetailsUninomial:
		if result.rank == "sect." {
			result.genus = detail.Uninomial.Parent
			result.subgenus = detail.Uninomial.Value
		} else {
			result.uninomial = detail.Uninomial.Value
		}
	case parsed.DetailsSpecies:
		result.genus = detail.Species.Genus
		result.species = detail.Species.Species
	case parsed.DetailsInfraspecies:
		if len(detail.Infraspecies.Infraspecies) == 1 {
			result.genus = detail.Infraspecies.Genus
			result.species = detail.Infraspecies.Species.Species
			result.infraspecies =
				detail.Infraspecies.Infraspecies[0].Value
		}
	}

	if parsedName.Hybrid != nil &&
		parsedName.Hybrid.String() == "NAMED_HYBRID" {
		result.remarks = "Hybrid taxon."

		nothoRegexp := regexp.MustCompile(`x([A-Z][a-z]+)`)

		addHybridMarker := false
		for _, word := range parsedName.Words {
			if word.Type.String() == "HYBRID_CHAR" {
				addHybridMarker = true
			} else {
				if addHybridMarker {
					switch word.Type.String() {
					case "GENUS":
						result.notho = coldp.GenericNP
						result.genus = "× " +
							nothoRegexp.ReplaceAllString(
								result.genus,
								"$1",
							)
					case "SUBGENUS":
						result.notho = coldp.InfragenericNP
						result.subgenus = "× " +
							nothoRegexp.ReplaceAllString(
								result.subgenus,
								"$1",
							)
					case "SPECIES":
						result.notho = coldp.SpecificNP
						result.species = "× " +
							nothoRegexp.ReplaceAllString(
								result.species,
								"$1",
							)
					case "INFRASPECIES":
						result.notho = coldp.InfraspecificNP
						result.infraspecies = "× " +
							nothoRegexp.ReplaceAllString(
								result.infraspecies,
								"$1",
							)
					}
				}
				addHybridMarker = false
			}
		}
	}

	return result, nil
}

func isAutonym(node hNode) bool {
	result := false

	// must eliminate any hybrid symbols for comparison
	canonicalSpecies := strings.ReplaceAll(node.parsed.species, "× ", "")
	canonicalInfraspecies := strings.ReplaceAll(
		node.parsed.infraspecies,
		"× ",
		"",
	)

	if canonicalSpecies == canonicalInfraspecies {
		result = true
	}
	return result
}

func spFromAutonym(
	child hNode,
	parser gnparser.GNparser,
	rankStack []hNode,
) (hNode, hNode) {
	autonym := child.parsed.genus + " " + child.parsed.species
	parsed, err := wfwpParse(parser, autonym, "")
	if err != nil {
		panic(fmt.Errorf("Error parsing autonym: %w", err))
	}

	parent := hNode{
		id:           "",
		parentId:     "",
		verbatimName: autonym,
		rank:         coldp.Species,
		parsed:       parsed,
	}

	parent.parentId = getParentId(rankStack, parent)

	// fix authorships so that trinomial has no authorship and
	// binomial receives the original authorship
	authorship := child.parsed.authorship
	child.parsed.authorship = ""
	parent.parsed.authorship = authorship

	return parent, child
}

func synonymRank(parsed gnparsed, acceptedRank coldp.Rank) coldp.Rank {
	result := coldp.UnknownRank
	acceptedGroup := rankGroup(acceptedRank)

	switch acceptedGroup {
	case "species":
		if parsed.infraspecies == "" {
			result = coldp.Species
		} else {
			switch parsed.rank {
			case "sp.":
				result = coldp.Species
			case "subsp.":
				result = coldp.Subspecies
			case "var.":
				result = coldp.Variety
			case "f.":
				result = coldp.Form
			case "subvar.":
				result = coldp.Subvariety
			case "subf.":
				result = coldp.Subform
			default:
				result = coldp.Subspecies
			}
		}
	case "genus":
		if parsed.rank == "sect." {
			result = coldp.Section
		} else {
			if parsed.subgenus != "" {
				result = coldp.Subgenus
			} else {
				result = coldp.Genus
			}
		}
	case "higher":
		result = rankBySuffix(parsed.canonicalFull)
	}

	// as a last resort set to accepted rank
	if result == coldp.UnknownRank {
		result = acceptedRank
	}

	return result
}

func getBasionymId(parsed gnparsed) string {
	stemmedParts := strings.Split(parsed.canonicalStemmed, " ")
	lowestPart := stemmedParts[len(stemmedParts)-1]
	return lowestPart + "_" + parsed.originalAuthorship
}
