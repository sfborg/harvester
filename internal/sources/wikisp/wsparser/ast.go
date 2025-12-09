package wsparser

// ast.go - AST walking for Wikispecies name parser
// This file contains the logic for traversing the parse tree and
// extracting semantic information (canonical name, authors, year, etc.)

import (
	"strings"
)

// parserState holds the AST root for walking.
type parserState struct {
	root *node32
}

var state = &parserState{}

// walkAST traverses the parse tree and extracts name components.
func (p *Parser) walkAST() ParsedName {
	result := ParsedName{
		Input: p.Buffer,
	}

	// Build the AST
	p.outputAST()

	// Start walking from root
	if state.root == nil {
		result.Quality = QualityUnparseable
		return result
	}

	// Track nodes for authorship construction and marker
	var authorNode, yearNode, parenAuthorNode, parenYearNode, markerNode *node32

	// Walk the tree to find author, year, and marker nodes
	n := state.root.up
	for n != nil {
		switch n.pegRule {
		case ruleMarker:
			if markerNode == nil {
				markerNode = n
			}
		case ruleParenthesizedAuthorPart:
			if parenAuthorNode == nil {
				parenAuthorNode = n
			}
		case ruleAuthorPart:
			if authorNode == nil {
				authorNode = n
			}
		case ruleYear:
			yearNode = n
		case ruleParenthesizedYear:
			parenYearNode = n
		}
		n = n.next
	}

	// Build authorship - handle various combinations
	if parenAuthorNode != nil {
		// Extract parenthesized authorship (original author)
		result.Authorship = p.extractParenthesizedAuthorship(parenAuthorNode)

		// Check if there's also a current author
		if authorNode != nil {
			authors := p.extractAuthorship(authorNode)
			if authors != "" {
				result.Authorship = result.Authorship + " " + authors
			}
		}

		// Add parenthesized year if present (approximate date)
		if parenYearNode != nil {
			parenYear := p.extractParenthesizedYear(parenYearNode)
			if parenYear != "" {
				result.Authorship = result.Authorship + " " + parenYear
			}
		}
	} else if authorNode != nil {
		// Regular authorship (no parenthesized author)
		authors := p.extractAuthorship(authorNode)
		if authors != "" {
			result.Authorship = authors

			// Add year (regular or parenthesized)
			if parenYearNode != nil {
				parenYear := p.extractParenthesizedYear(parenYearNode)
				if parenYear != "" {
					result.Authorship = authors + " " + parenYear
				}
			} else if yearNode != nil {
				runes := []rune(p.Buffer)
				// Get everything from end of author to end of year
				// This preserves commas, spaces, and the year exactly as in input
				betweenAndYear := string(runes[authorNode.end:yearNode.end])

				// Check if there's a comma in the between text
				hasComma := strings.Contains(betweenAndYear, ",")

				// Clean and extract just the year
				yearOnly := strings.TrimSpace(betweenAndYear)
				yearOnly = strings.TrimPrefix(yearOnly, ",")
				yearOnly = strings.TrimSpace(yearOnly)

				// Format: use comma if original had comma, otherwise just space
				if hasComma {
					result.Authorship = authors + ", " + yearOnly
				} else {
					result.Authorship = authors + " " + yearOnly
				}
			}
		}
	} else if parenYearNode != nil {
		// Parenthesized year only
		result.Authorship = p.extractParenthesizedYear(parenYearNode)
	} else if yearNode != nil {
		// Regular year only, no authors
		result.Authorship = p.extractYear(yearNode)
	}

	// Walk again for other components
	n = state.root.up
	for n != nil {
		p.processNode(n, &result)
		n = n.next
	}

	// Prepend marker to canonical name if present (preserve spacing from input)
	if markerNode != nil && result.Canonical != "" {
		marker := p.nodeValue(markerNode)

		// Find the NamePart node to check spacing
		var nameNode *node32
		n := state.root.up
		for n != nil {
			if n.pegRule == ruleNamePart {
				nameNode = n
				break
			}
			n = n.next
		}

		// Check if there's whitespace between marker and name in input
		if nameNode != nil {
			runes := []rune(p.Buffer)
			betweenText := string(runes[markerNode.end:nameNode.begin])
			// If there's any whitespace between them, add a space
			if strings.TrimSpace(betweenText) == "" && len(betweenText) > 0 {
				result.Canonical = marker + " " + result.Canonical
			} else {
				result.Canonical = marker + result.Canonical
			}
		} else {
			// Fallback: no space
			result.Canonical = marker + result.Canonical
		}
	}

	// Calculate quality
	result.Quality = calculateQuality(&result)

	return result
}

// outputAST assembles PEG nodes' AST structure (adapted from gnparser).
func (p *Parser) outputAST() {
	type element struct {
		node *node32
		down *element
	}
	var node *node32
	var skip bool
	var stack *element
	for _, token := range p.Tokens() {
		if node, skip = p.newNode(token); skip {
			continue
		}
		for stack != nil && stackNodeIsWithin(stack.node, token) {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		state.root = stack.node
	}
}

func stackNodeIsWithin(n *node32, t token32) bool {
	return n.begin >= t.begin && n.end <= t.end
}

// newNode creates a new AST node from a token.
func (p *Parser) newNode(t token32) (*node32, bool) {
	// We want to keep all structural nodes
	return &node32{token32: t}, false
}

// processNode handles a single node in the parse tree.
func (p *Parser) processNode(n *node32, result *ParsedName) {
	switch n.pegRule {
	case ruleNamePart:
		result.Canonical = p.extractCanonical(n)
	case ruleReference:
		result.Reference = p.nodeValue(n)
	case ruleTail:
		result.Tail = p.nodeValue(n)
	}
}

// nodeValue extracts the text value of a node.
func (p *Parser) nodeValue(n *node32) string {
	if n == nil {
		return ""
	}
	t := n.token32
	return string([]rune(p.Buffer)[t.begin:t.end])
}

// extractCanonical extracts and cleans the canonical scientific name.
func (p *Parser) extractCanonical(n *node32) string {
	if n == nil || n.up == nil {
		return ""
	}

	child := n.up
	switch child.pegRule {
	case ruleItalicInfraspeciesWithAuthors:
		return p.extractInfraspeciesWithAuthors(child)
	case ruleItalicInfraspecies:
		return p.extractItalicInfraspecies(child)
	case ruleItalicName:
		return p.extractItalicName(child)
	case ruleBareName:
		return p.extractBareName(child)
	}

	return ""
}

// extractInfraspeciesWithAuthors handles: ”Species” Authors rank ”infraspecies”
// Returns: Species rank infraspecies (authors ignored for autonyms)
func (p *Parser) extractInfraspeciesWithAuthors(n *node32) string {
	var species, rank, epithet string

	child := n.up
	for child != nil {
		text := p.nodeValue(child)
		switch child.pegRule {
		case ruleItalicContent:
			if species == "" {
				species = text
			} else {
				epithet = text
			}
		case ruleInfraRank:
			rank = text
		}
		child = child.next
	}

	if species != "" && rank != "" && epithet != "" {
		return species + " " + rank + " " + epithet
	}
	return ""
}

// extractItalicInfraspecies handles: ”Species” rank ”infraspecies”
func (p *Parser) extractItalicInfraspecies(n *node32) string {
	text := p.nodeValue(n)
	// Remove italic markup
	text = strings.ReplaceAll(text, "''", "")
	return text
}

// extractItalicName handles: ”Name” or ”'Name”'
func (p *Parser) extractItalicName(n *node32) string {
	child := n.up
	for child != nil {
		switch child.pegRule {
		case ruleSimpleItalicName, ruleBoldItalicName:
			return p.extractItalicContent(child)
		}
		child = child.next
	}
	return ""
}

// extractItalicContent extracts text from within italic markup.
func (p *Parser) extractItalicContent(n *node32) string {
	// For bold italic, there's no ItalicContent child, just raw text
	if n.pegRule == ruleBoldItalicName {
		text := p.nodeValue(n)
		// Remove the triple quotes
		text = strings.TrimPrefix(text, "'''")
		text = strings.TrimSuffix(text, "'''")
		return text
	}

	// For simple italic, look for ItalicContent child
	child := n.up
	for child != nil {
		if child.pegRule == ruleItalicContent {
			return p.nodeValue(child)
		}
		child = child.next
	}
	return ""
}

// extractBareName handles: Genus species (no italic markup)
func (p *Parser) extractBareName(n *node32) string {
	child := n.up
	for child != nil {
		if child.pegRule == ruleBareNameContent {
			return p.nodeValue(child)
		}
		child = child.next
	}
	return ""
}

// extractAuthorship extracts the complete authorship string.
// For templates like {{a|Full Name|Short}}, uses the short form if available.
// For brackets like [[Full Name|Short]], uses the short form.
func (p *Parser) extractAuthorship(n *node32) string {
	if n == nil || n.up == nil {
		return ""
	}

	var parts []string

	child := n.up
	for child != nil {
		if child.pegRule == ruleAuthors {
			parts = append(parts, p.extractAuthorshipFromAuthors(child))
		}
		child = child.next
	}

	return strings.Join(parts, "")
}

// extractAuthorshipFromAuthors extracts authorship from Authors node.
func (p *Parser) extractAuthorshipFromAuthors(n *node32) string {
	var authorParts []string
	var italicAuthorPart string

	child := n.up
	for child != nil {
		if child.pegRule == ruleAuthor {
			if auth := p.extractAuthorString(child); auth != "" {
				authorParts = append(authorParts, auth)
			}
		} else if child.pegRule == ruleItalicAuthorText {
			// Handle italic author text (like "et al.") separately
			// It should be appended with space, not joined with "&"
			if auth := p.extractAuthorFromItalic(child); auth != "" {
				italicAuthorPart = auth
			}
		}
		child = child.next
	}

	result := strings.Join(authorParts, " & ")
	if italicAuthorPart != "" {
		result = result + " " + italicAuthorPart
	}
	return result
}

// extractAuthorString extracts a single author string (using short form if available).
func (p *Parser) extractAuthorString(n *node32) string {
	child := n.up
	if child == nil {
		return ""
	}

	switch child.pegRule {
	case ruleAuthorTemplate:
		return p.extractAuthorFromTemplate(child)
	case ruleBracketAuthor:
		return p.extractAuthorFromBracket(child)
	case ruleItalicAuthorText:
		return p.extractAuthorFromItalic(child)
	}

	return ""
}

// extractAuthorFromTemplate handles: {{a|Full Name|Short|param=val}}
// Returns the short form if available, otherwise the full name.
func (p *Parser) extractAuthorFromTemplate(n *node32) string {
	// Get the raw template text
	text := p.nodeValue(n)

	// Remove template wrapper: {{type|...}}
	text = strings.TrimPrefix(text, "{{")
	text = strings.TrimSuffix(text, "}}")

	// Split by |
	parts := strings.Split(text, "|")
	if len(parts) < 2 {
		return ""
	}

	// parts[0] is template type (a, au, aut)
	// parts[1] is full author name
	// parts[2] is short author name (if present)

	fullName := strings.TrimSpace(parts[1])
	if len(parts) > 2 {
		shortName := strings.TrimSpace(parts[2])
		// Skip if it looks like a parameter (key=value)
		if shortName != "" && !strings.Contains(shortName, "=") {
			return shortName
		}
	}

	return fullName
}

// extractAuthorFromBracket handles: [[Full Name|Short]]
// Returns the short form if available.
func (p *Parser) extractAuthorFromBracket(n *node32) string {
	text := p.nodeValue(n)
	// Remove brackets
	text = strings.TrimPrefix(text, "[[")
	text = strings.TrimSuffix(text, "]]")

	// Split on |
	parts := strings.Split(text, "|")
	if len(parts) >= 2 {
		// Return short form (second part)
		return strings.TrimSpace(parts[1])
	}

	// No short form, return full name
	return strings.TrimSpace(text)
}

// extractAuthorFromItalic handles: ''et al''. (with optional trailing period)
// Returns the text inside the italic markup plus period if present.
func (p *Parser) extractAuthorFromItalic(n *node32) string {
	text := p.nodeValue(n)
	// Check if there's a trailing period (after the closing '')
	hasPeriod := strings.HasSuffix(text, ".")
	if hasPeriod {
		text = strings.TrimSuffix(text, ".")
	}
	// Remove italic markup
	text = strings.TrimPrefix(text, "''")
	text = strings.TrimSuffix(text, "''")
	text = strings.TrimSpace(text)
	// Add period back if it was present
	if hasPeriod {
		text = text + "."
	}
	return text
}

// extractParenthesizedAuthorship extracts authorship from parentheses.
// Returns the complete authorship wrapped in parentheses, e.g., "(Hübner, 1827)"
func (p *Parser) extractParenthesizedAuthorship(n *node32) string {
	if n == nil || n.up == nil {
		return ""
	}

	// Find Authors and Year nodes within the parenthesized node
	var authorsNode, yearNode *node32
	child := n.up
	for child != nil {
		switch child.pegRule {
		case ruleAuthors:
			if authorsNode == nil {
				authorsNode = child
			}
		case ruleYear:
			yearNode = child
		}
		child = child.next
	}

	// Build the authorship string
	var parts []string

	if authorsNode != nil {
		authors := p.extractAuthorshipFromAuthors(authorsNode)
		if authors != "" {
			parts = append(parts, authors)
		}
	}

	if yearNode != nil {
		year := p.extractYear(yearNode)
		if year != "" {
			parts = append(parts, year)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	// Join with comma and space, wrap in parentheses
	return "(" + strings.Join(parts, ", ") + ")"
}

// extractYear extracts the publication year.
func (p *Parser) extractYear(n *node32) string {
	text := p.nodeValue(n)
	// Remove comma and whitespace
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, ",")
	text = strings.TrimSpace(text)
	return text
}

// extractParenthesizedYear extracts a parenthesized year (approximate date).
// Returns the year with parentheses, e.g., "(1899)"
func (p *Parser) extractParenthesizedYear(n *node32) string {
	text := p.nodeValue(n)
	// The text already includes parentheses, just trim whitespace
	text = strings.TrimSpace(text)
	return text
}

// calculateQuality determines parse quality based on extracted fields.
func calculateQuality(result *ParsedName) Quality {
	// If no canonical name extracted, it's unparseable
	if result.Canonical == "" {
		return QualityUnparseable
	}

	// If there's a tail (unparsed content without ':'), quality is partial
	if result.Tail != "" {
		return QualityPartial
	}

	// If we have canonical + reference (or just canonical), it's good
	return QualityGood
}
