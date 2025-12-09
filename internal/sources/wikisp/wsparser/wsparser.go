package wsparser

import (
	"fmt"
	"regexp"
	"strings"
)

// GNparser interface defines the minimal gnparser functionality needed.
type GNparser interface {
	ParseName(name string) ParseResult
}

// ParseResult represents the result from gnparser.
type ParseResult interface {
	Flatten() FlattenedResult
}

// FlattenedResult represents flattened gnparser output.
type FlattenedResult struct {
	Parsed        bool
	ParseQuality  int
	CanonicalFull string
	Authorship    string
}

// WSParser wraps the PEG parser with optional gnparser fallback.
type WSParser struct {
	gnparser GNparser
}

// New creates a new NameParser with gnparser fallback support.
// If gnparser is nil, only wsparser will be used.
func New(gnparser GNparser) *WSParser {
	return &WSParser{
		gnparser: gnparser,
	}
}

// Parse parses a Wikispecies name string using wsparser first,
// with gnparser fallback if wsparser fails.
func (np *WSParser) Parse(input string) (ParsedName, error) {
	// Try wsparser first
	result, err := parseWithPEG(input)

	// If wsparser succeeded, return its result
	if err == nil && result.Quality != QualityUnparseable {
		return result, nil
	}

	// If no gnparser fallback available, return wsparser result
	if np.gnparser == nil {
		return result, err
	}

	// Try gnparser fallback
	cleaned := cleanWikiMarkup(input)
	gnResult := np.gnparser.ParseName(cleaned).Flatten()

	// Only use gnparser if quality is 1 (highest)
	if gnResult.Parsed && gnResult.ParseQuality < 3 {
		return ParsedName{
			Input:      input,
			Canonical:  gnResult.CanonicalFull,
			Authorship: gnResult.Authorship,
			Quality:    QualityGood, // Highest quality from wsparser
		}, nil
	}

	// Both parsers failed, return original wsparser result
	return result, err
}

// Parse is the legacy function that parses without gnparser fallback.
// Kept for backward compatibility.
func Parse(input string) (ParsedName, error) {
	return parseWithPEG(input)
}

// parseWithPEG performs the actual PEG parsing.
func parseWithPEG(input string) (ParsedName, error) {
	p := &Parser{
		Buffer: input,
	}

	p.Init()
	if err := p.Parse(); err != nil {
		return ParsedName{
			Input:   input,
			Quality: QualityUnparseable,
		}, fmt.Errorf("parse failed: %w", err)
	}

	// Walk the AST to extract semantic information
	result := p.walkAST()

	// Post-process: clean wiki links from canonical name and authorship
	// [[Link|Display]] → Display, [[Link]] → Link
	result.Canonical = wikiLink.ReplaceAllString(result.Canonical, "$2")
	result.Authorship = wikiLink.ReplaceAllString(result.Authorship, "$2")

	return result, nil
}

var (
	templateLink = regexp.MustCompile(`\{\{([^}]*)\}\}`)
	wikiLink     = regexp.MustCompile(`\[\[([^\]|]+\|)?([^\]]*)\]\]`)
	boldItalic   = regexp.MustCompile(`'{2,}`)
	htmlTag      = regexp.MustCompile(`<[^>]+>`)
	whitespace   = regexp.MustCompile(`\s+`)
)

// cleanWikiMarkup removes wiki markup for gnparser fallback.
func cleanWikiMarkup(text string) string {
	// Remove wiki links: [[Link|Display]] → Display, [[Link]] → Link
	text = wikiLink.ReplaceAllString(text, "$2")

	// Remove bold/italic markup
	text = boldItalic.ReplaceAllString(text, "")

	// Replace templates: {{template|param1|param2}} → param2 (keep last part)
	for {
		before := text
		text = templateLink.ReplaceAllStringFunc(text, func(match string) string {
			content := match[2 : len(match)-2]
			parts := strings.Split(content, "|")
			if len(parts) > 1 {
				for i := len(parts) - 1; i >= 0; i-- {
					if strings.TrimSpace(parts[i]) != "" {
						return strings.TrimSpace(parts[i])
					}
				}
			}
			return ""
		})
		if text == before {
			break
		}
	}

	// Remove HTML tags
	text = htmlTag.ReplaceAllString(text, "")

	// Remove extinct markers
	text = strings.ReplaceAll(text, "†", "")
	text = strings.ReplaceAll(text, "&dagger;", "")

	// Clean HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")

	// Normalize whitespace
	text = whitespace.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}
