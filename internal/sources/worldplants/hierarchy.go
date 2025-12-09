package worldplants

import (
	"bufio"
	"fmt"
	"html"
	"log/slog"
	"os"
	"strings"

	"github.com/gnames/gn"
	"github.com/google/uuid"
	"github.com/sfborg/sflib/pkg/coldp"
)

// getNode parses a CSV line into an hNode structure.
// Reference: original lines 370-420
func (wp *worldplants) getNode(line string) (hNode, error) {
	row := strings.Split(line, "|")
	if len(row) < 9 {
		return hNode{}, fmt.Errorf("invalid CSV row: not enough fields")
	}

	rankValue := rank(row[0])
	accepted := row[2]
	citation := row[3]
	vernacularName := row[4]
	distribution := row[5]
	synonyms := row[6]
	status := row[7]
	remarks := row[8]

	// Validate name before processing
	if err := validateName(accepted); err != nil {
		return hNode{}, err
	}

	// Parse the accepted name
	acceptedParsed, err := wfwpParse(wp.parser, accepted, citation)
	if err != nil {
		return hNode{}, err
	}

	// Reject quadrinomials
	if acceptedParsed.cardinality > 3 {
		return hNode{}, fmt.Errorf(
			"illegal name (no quadrinomials) \"%s\"",
			accepted,
		)
	}

	// Generate temporary ID (will be replaced with persistent ID later)
	id := uuid.NewString()

	result := hNode{
		id:                     id,
		verbatimName:           accepted,
		rank:                   rankValue,
		parsed:                 acceptedParsed,
		verbatimCitation:       citation,
		verbatimDistribution:   distribution,
		verbatimRemarks:        remarks,
		varbatimStatus:         status,
		verbatimSynonyms:       synonyms,
		verbatimVernacularName: vernacularName,
	}

	return result, nil
}

// validateName checks if a name should be skipped.
func validateName(name string) error {
	nameLower := strings.ToLower(name)

	// Skip problematic names
	skipNames := []string{
		"incertae sedis",
		"undescribed",
		"nov. ined.",
		"taxonomic position unknown",
		"undescribed clade",
		"probably to excluded from celastraceae",
	}

	for _, skip := range skipNames {
		if nameLower == skip {
			return fmt.Errorf("illegal name \"%s\"", name)
		}
	}

	// Skip clade names
	if strings.HasSuffix(nameLower, " clade") {
		return fmt.Errorf("illegal name \"%s\"", name)
	}

	// Skip subgroup names
	if strings.HasSuffix(nameLower, " subgroup") {
		return fmt.Errorf("illegal name \"%s\"", name)
	}

	return nil
}

// buildHierarchy reads a CSV file and builds a hierarchical node structure.
// Reference: original lines 838-911
func (wp *worldplants) buildHierarchy(
	csvPath string,
) ([]hNode, map[string]hNode, error) {
	slog.Info("building hierarchy", "file", csvPath)
	gn.Info("Building hierarchy <em>%s</em>", csvPath)

	file, err := os.Open(csvPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open CSV file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Skip header
	_, err = reader.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read header: %w", err)
	}

	var nodes []hNode
	nodeMap := make(map[string]hNode)
	rankStack := make([]hNode, 0, 30)

	lineNum := 1
	for {
		lineNum++

		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				slog.Error("error reading line", "line", lineNum, "error", err)
			}
			break
		}

		// Unescape HTML entities
		line = html.UnescapeString(line)

		node, err := wp.getNode(line)
		if err != nil {
			slog.Debug("skipping node", "line", lineNum, "error", err)
			continue
		}

		// Process node based on rank
		// Returns the processed node, potentially a parent node
		// (for autonyms), and updated rankStack
		var parentNode *hNode
		node, parentNode, rankStack = wp.processNodeByRank(
			node,
			nodes,
			rankStack,
		)

		// If autonym created parent, add parent first
		if parentNode != nil {
			nodes = append(nodes, *parentNode)
			nodeMap[parentNode.id] = *parentNode
		}

		nodes = append(nodes, node)
		nodeMap[node.id] = node

		if lineNum%1000 == 0 {
			slog.Info("processed rows", "count", lineNum)
		}
	}

	slog.Info("hierarchy built", "nodes", len(nodes))
	gn.Info("Finished building hierarchy: <em>%d nodes</em>", len(nodes))

	return nodes, nodeMap, nil
}

// processNodeByRank handles node processing based on its rank.
// Returns: processed node, optional parent node (for autonyms), rankStack.
// Reference: original lines 860-906
func (wp *worldplants) processNodeByRank(
	node hNode,
	nodes []hNode,
	rankStack []hNode,
) (hNode, *hNode, []hNode) {
	switch {
	case node.rank == coldp.Order:
		return wp.handleOrderRank(node, rankStack)

	case node.rank == coldp.Species:
		return wp.handleSpeciesRank(node, rankStack)

	case node.rank == coldp.Subspecies ||
		node.rank == coldp.Variety ||
		node.rank == coldp.Form:
		return wp.handleInfraspeciesRank(node, nodes, rankStack)

	case node.rank == coldp.UnknownRank:
		// Skip unknown ranks (excluded from COL)
		return node, nil, rankStack

	case len(nodes) > 0 &&
		rankLevel(node.rank) < rankLevel(nodes[len(nodes)-1].rank):
		return wp.handleHigherRank(node, rankStack)

	default:
		return wp.handleDefaultRank(node, rankStack)
	}
}

// handleOrderRank processes Order rank nodes.
func (wp *worldplants) handleOrderRank(
	node hNode,
	rankStack []hNode,
) (hNode, *hNode, []hNode) {
	rankStack = append(rankStack, node)
	return node, nil, rankStack
}

// handleSpeciesRank processes Species rank nodes.
func (wp *worldplants) handleSpeciesRank(
	node hNode,
	rankStack []hNode,
) (hNode, *hNode, []hNode) {
	// Pop species from stack if last item is species
	if len(rankStack) > 0 && rankStack[len(rankStack)-1].rank == coldp.Species {
		rankStack = rankStack[:len(rankStack)-1]
	}

	node.parentId = getParentId(rankStack, node)
	return node, nil, rankStack
}

// handleInfraspeciesRank processes infraspecific rank nodes.
func (wp *worldplants) handleInfraspeciesRank(
	node hNode,
	nodes []hNode,
	rankStack []hNode,
) (hNode, *hNode, []hNode) {
	if len(nodes) == 0 {
		node.parentId = getParentId(rankStack, node)
		return node, nil, rankStack
	}

	lastNode := nodes[len(nodes)-1]

	// Check if parent is the previous species
	if lastNode.rank == coldp.Species &&
		strings.HasPrefix(
			node.parsed.canonicalSimple,
			lastNode.parsed.canonicalSimple,
		) {
		rankStack = append(rankStack, lastNode)
		node.parentId = getParentId(rankStack, node)
		return node, nil, rankStack
	}

	// Check for autonym
	if isAutonym(node) {
		parentNode, childNode := spFromAutonym(node, wp.parser, rankStack)
		rankStack = append(rankStack, parentNode)
		childNode.parentId = getParentId(rankStack, childNode)

		// Return child node and parent node (to be added to nodes list)
		return childNode, &parentNode, rankStack
	}

	node.parentId = getParentId(rankStack, node)
	return node, nil, rankStack
}

// handleHigherRank processes nodes with higher rank than previous.
func (wp *worldplants) handleHigherRank(
	node hNode,
	rankStack []hNode,
) (hNode, *hNode, []hNode) {
	// Pop stack until we find appropriate parent
	for len(rankStack) > 0 &&
		rankLevel(node.rank) <= rankLevel(rankStack[len(rankStack)-1].rank) {
		rankStack = rankStack[:len(rankStack)-1]
	}

	node.parentId = getParentId(rankStack, node)
	rankStack = append(rankStack, node)
	return node, nil, rankStack
}

// handleDefaultRank processes nodes with default behavior.
func (wp *worldplants) handleDefaultRank(
	node hNode,
	rankStack []hNode,
) (hNode, *hNode, []hNode) {
	node.parentId = getParentId(rankStack, node)
	rankStack = append(rankStack, node)
	return node, nil, rankStack
}
