package wikisp

import (
	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/sfborg/harvester/internal/sources/wikisp/wsparser"
)

// gnparserAdapter adapts gnparser.GNparser to wsparser.GNparser interface.
type gnparserAdapter struct {
	gnp gnparser.GNparser
}

// ParseName implements wsparser.GNparser interface.
func (g *gnparserAdapter) ParseName(name string) wsparser.ParseResult {
	return &parseResultAdapter{result: g.gnp.ParseName(name)}
}

// parseResultAdapter adapts gnparser result to wsparser.ParseResult interface.
type parseResultAdapter struct {
	result parsed.Parsed
}

// Flatten implements wsparser.ParseResult interface.
func (p *parseResultAdapter) Flatten() wsparser.FlattenedResult {
	flat := p.result.Flatten()
	return wsparser.FlattenedResult{
		Parsed:        flat.Parsed,
		ParseQuality:  flat.ParseQuality,
		CanonicalFull: flat.CanonicalFull,
		Authorship:    flat.Authorship,
	}
}
