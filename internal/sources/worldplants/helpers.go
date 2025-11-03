package worldplants

import (
	"strings"

	"github.com/google/uuid"
	"github.com/sfborg/sflib/pkg/coldp"
)

// hNode represents a hierarchical node in the taxonomy.
type hNode struct {
	id                     string
	parentId               string
	verbatimName           string
	verbatimCitation       string
	verbatimDistribution   string
	verbatimRemarks        string
	varbatimStatus         string
	verbatimSynonyms       string
	verbatimVernacularName string
	rank                   coldp.Rank
	parsed                 gnparsed
}

// citation holds reference information.
type citation struct {
	id       string
	author   string
	year     string
	title    string
	citation string
}

func rank(rank string) coldp.Rank {
	rank = strings.ToUpper(rank)
	rankMap := map[string]coldp.Rank{
		"K":          coldp.Kingdom,
		"C":          coldp.Class,
		"O":          coldp.Order,
		"_SO_":       coldp.Suborder,
		"F":          coldp.Family,
		"SF":         coldp.Subfamily,
		"T":          coldp.Tribe,
		"ST":         coldp.Subtribe,
		"G":          coldp.Genus,
		"_SG_":       coldp.Subgenus,
		"_SG2_":      coldp.Subgenus,
		"S":          coldp.Species,
		"SS":         coldp.Subspecies,
		"_SV_":       coldp.Subvariety,
		"_SF_2":      coldp.Subform,
		"_SSP_":      coldp.Subspecies,
		"V":          coldp.Variety,
		"FM":         coldp.Form,
		"_FM2_":      coldp.Form,
		"_NSSP_":     coldp.UnknownRank,
		"_RC_":       coldp.UnknownRank,
		"KINGDOM":    coldp.Kingdom,
		"CLASS":      coldp.Class,
		"ORDER":      coldp.Order,
		"SUBORDER":   coldp.Suborder,
		"FAMILY":     coldp.Family,
		"SUBFAMILY":  coldp.Subfamily,
		"TRIBE":      coldp.Tribe,
		"SUBTRIBE":   coldp.Subtribe,
		"GENUS":      coldp.Genus,
		"SUBGENUS":   coldp.Subgenus,
		"SPECIES":    coldp.Species,
		"SUBSPECIES": coldp.Subspecies,
		"VARIETY":    coldp.Variety,
		"FORM":       coldp.Form,
		"SUBVARIETY": coldp.Subvariety,
		"SUBFORM":    coldp.Subform,
		"UNKNOWN":    coldp.UnknownRank,
	}

	if result, ok := rankMap[rank]; ok {
		return result
	}
	return coldp.UnknownRank
}

func rankBySuffix(name string) coldp.Rank {
	var result coldp.Rank

	suffixes := []struct {
		suffix string
		rank   coldp.Rank
	}{
		{"mycetidae", coldp.Subclass},
		{"mycotina", coldp.Subphylum},
		{"phycidae", coldp.Subclass},
		{"mycetes", coldp.Class},
		{"phyceae", coldp.Class},
		{"phytina", coldp.Subphylum},
		{"mycota", coldp.Phylum},
		{"opsida", coldp.Class},
		{"oideae", coldp.Subfamily},
		{"phyta", coldp.Phylum},
		{"aceae", coldp.Family},
		{"oidea", coldp.Superfamily},
		{"ineae", coldp.Suborder},
		{"idae", coldp.Subclass},
		{"inae", coldp.Subtribe},
		{"anae", coldp.Superorder},
		{"ales", coldp.Order},
		{"ana", coldp.Superorder},
		{"eae", coldp.Tribe},
	}

	for _, r := range suffixes {
		if strings.HasSuffix(name, r.suffix) {
			result = r.rank
			break
		}
	}

	return result
}

func rankLevel(rank coldp.Rank) int {
	rankOrder := map[coldp.Rank]int{
		coldp.Order:       1,
		coldp.Suborder:    2,
		coldp.Family:      3,
		coldp.Subfamily:   4,
		coldp.Tribe:       5,
		coldp.Subtribe:    6,
		coldp.Genus:       7,
		coldp.Subgenus:    8,
		coldp.Section:     9,
		coldp.Species:     10,
		coldp.Subspecies:  11,
		coldp.Variety:     12,
		coldp.Form:        13,
		coldp.Subvariety:  14,
		coldp.Subform:     15,
		coldp.UnknownRank: 16,
	}
	return rankOrder[rank]
}

func rankGroup(rank coldp.Rank) string {
	var result string
	switch rank {
	case coldp.Species, coldp.Subspecies, coldp.Variety, coldp.Form,
		coldp.Subvariety, coldp.Subform:
		result = "species"
	case coldp.Genus, coldp.Subgenus, coldp.Section:
		result = "genus"
	default:
		result = "higher"
	}
	return result
}

func getParentId(rankStack []hNode, child hNode) string {
	// iterate up rankLevel from the current hNode rank to
	// find next highest parent
	for i := len(rankStack) - 1; i >= 0; i-- {
		if rankLevel(rankStack[i].rank) < rankLevel(child.rank) {
			return rankStack[i].id
		}
	}
	return ""
}

func getPersistentId(
	node hNode,
	nodesMap map[string]hNode,
	namespace uuid.UUID,
	verboseIds bool,
) string {
	parents := []string{node.verbatimName}
	for {
		parentNode := nodesMap[node.parentId]
		if parentNode.verbatimName == "" {
			break
		}
		parents = append([]string{parentNode.verbatimName}, parents...)
		if parentNode.parentId == "" {
			break
		}
		node = parentNode
	}
	result := strings.Join(parents, "_")

	if !verboseIds {
		if result == "" {
			result = uuid.Nil.String()
		} else {
			result = uuid.NewSHA1(namespace, []byte(result)).String()
		}
	}
	return result
}
