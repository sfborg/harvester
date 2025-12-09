package wikisp

import (
	"regexp"

	"github.com/sfborg/harvester/internal/sources/wikisp/wsparser"
)

var (
	pageStart     = regexp.MustCompile(`^\s*<page>\s*$`)
	pageEnd       = regexp.MustCompile(`^\s*</page>\s*$`)
	sectionHeader = regexp.MustCompile(`^(=+)([^=]+)=+$`)
	templateLink  = regexp.MustCompile(`\{\{([^}]*)\}\}`)
	wikiLink      = regexp.MustCompile(`\[\[([^\]|]+\|)?([^\]]*)\]\]`)
	boldItalic    = regexp.MustCompile(`'{2,}`)
	htmlTag       = regexp.MustCompile(`<[^>]+>`)
	whitespace    = regexp.MustCompile(`\s+`)
	redirectLink  = regexp.MustCompile(`#(?i)redirect\s*\[\[([^\]]+)\]\]`)
	redirectTitle = regexp.MustCompile(`<redirect title="([^"]+)"`)
)

type PageXML struct {
	Title    string   `xml:"title"`
	ID       int      `xml:"id"`
	Revision Revision `xml:"revision"`
}

type Revision struct {
	Timestamp string      `xml:"timestamp"`
	Format    string      `xml:"format"`
	Text      TextElement `xml:"text"`
}

type TextElement struct {
	Content string `xml:",chardata"`
}

type PageData struct {
	ID              string
	Title           string
	ScientificName  string
	Authorship      string
	ParseQuality    wsparser.Quality // wsparser parse quality
	ParentTemplate  string
	Synonyms        []string
	VernacularNames map[string]string
}

type Section struct {
	Header string
	Level  int
	Lines  []string
}

type tempStorage struct {
	redirects   map[string]string
	templateIDs map[string]string
	taxonIDs    map[string]string
}

type synonym struct {
	CanonicalName     string
	Authorship        string
	Quality           wsparser.Quality
	AcceptedID        string
	HasRedirect       bool
	HasSynonymSection bool
}

type parseStats struct {
	TotalPages             int
	SkippedNamespace       int
	SkippedRedirects       int
	SkippedTemplates       int
	SkippedInvalidXML      int
	TaxonPages             int
	TaxonPagesFailed       int
	ParentResolved         int
	ParentNotFound         int
	SynonymsTotal          int
	SynonymsParseFailed    int
	SynonymDuplicates      int
	RedirectTargetNotFound int
	NamesAccepted          int
	NamesRejected          int
	MissingParents         map[string][]string // template -> list of taxa that need it
	MissingRedirectTargets map[string][]string // target -> list of redirects
}
