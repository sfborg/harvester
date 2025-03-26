package data

import (
	"github.com/gnames/gnparser"
	"github.com/sfborg/sflib/pkg/coldp"
)

func AddParsedData(p gnparser.GNparser, nu *coldp.NameUsage) {
	prsd := p.ParseName(nu.ScientificNameString).Flatten()

	if prsd.Parsed {
		nu.ParseQuality = coldp.ToInt(prsd.ParseQuality)
		if prsd.ParseQuality > 2 {
			return
		}
		nu.CanonicalSimple = prsd.CanonicalSimple
		nu.CanonicalFull = prsd.CanonicalFull
		nu.CanonicalStemmed = prsd.CanonicalStemmed
		nu.Cardinality = coldp.ToInt(prsd.Cardinality)
		nu.Virus = coldp.ToBool(prsd.Virus)
		nu.Hybrid = prsd.Hybrid
		nu.Surrogate = prsd.Surrogate
		nu.Authors = prsd.Authors
		nu.GnID = prsd.VerbatimID

		nu.Authorship = pick(nu.Authorship, prsd.Authorship)
		nu.Rank = coldp.NewRank(pick(nu.Rank.String(), prsd.Rank))
		nu.Uninomial = pick(nu.Uninomial, prsd.Uninomial)
		nu.GenericName = pick(nu.GenericName, prsd.Genus)
		nu.InfragenericEpithet = pick(nu.InfragenericEpithet, prsd.Subgenus)
		nu.SpecificEpithet = pick(nu.SpecificEpithet, prsd.Species)
		nu.InfraspecificEpithet = pick(
			nu.InfraspecificEpithet,
			prsd.Infraspecies,
		)
		nu.CultivarEpithet = pick(nu.CultivarEpithet, prsd.CultivarEpithet)

		nu.CombinationAuthorship = pick(
			nu.CombinationAuthorship,
			prsd.CombinationAuthorship,
		)
		nu.CombinationExAuthorship = pick(
			nu.CombinationExAuthorship,
			prsd.CombinationExAuthorship,
		)
		nu.CombinationAuthorshipYear = pick(
			nu.CombinationAuthorshipYear,
			prsd.CombinationAuthorshipYear,
		)

		nu.BasionymAuthorship = pick(
			nu.BasionymAuthorship,
			prsd.BasionymAuthorship,
		)
		nu.BasionymExAuthorship = pick(
			nu.BasionymExAuthorship,
			prsd.BasionymExAuthorship,
		)
		nu.BasionymAuthorshipYear = pick(
			nu.BasionymAuthorshipYear,
			prsd.BasionymAuthorshipYear,
		)
	}
}

func pick(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
