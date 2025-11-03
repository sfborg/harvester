package worldplants

import (
	"testing"

	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/gnames/gnparser"
	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/stretchr/testify/assert"
)

func getTestParser() gnparser.GNparser {
	opts := []gnparser.Option{
		gnparser.OptCode(nomcode.Botanical),
		gnparser.OptWithDetails(true),
	}
	cfg := gnparser.NewConfig(opts...)
	return gnparser.New(cfg)
}

func TestFormatAuthors(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		msg, output string
		input       []string
	}{
		{
			msg:    "single author",
			input:  []string{"Smith"},
			output: "Smith",
		},
		{
			msg:    "two authors",
			input:  []string{"Smith", "Jones"},
			output: "Smith & Jones",
		},
		{
			msg:    "three authors",
			input:  []string{"Smith", "Jones", "Brown"},
			output: "Smith, Jones & Brown",
		},
		{
			msg:    "four authors",
			input:  []string{"Smith", "Jones", "Brown", "White"},
			output: "Smith, Jones, Brown & White",
		},
		{
			msg:    "author with initials",
			input:  []string{"Smith J. K."},
			output: "Smith J.K.",
		},
	}

	for _, v := range tests {
		res := formatAuthors(v.input)
		assert.Equal(v.output, res, v.msg)
	}
}

func TestWfwpParse(t *testing.T) {
	assert := assert.New(t)
	parser := getTestParser()

	tests := []struct {
		msg                    string
		name                   string
		citation               string
		expectedGenus          string
		expectedSpecies        string
		expectedInfraspecies   string
		expectedAuthorship     string
		expectedCardinality    int
		expectedNameStatus     coldp.NomStatus
		expectedAppendedPhrase string
		shouldFail             bool
	}{
		{
			msg:                 "simple binomial",
			name:                "Abies alba Mill.",
			citation:            "",
			expectedGenus:       "Abies",
			expectedSpecies:     "alba",
			expectedAuthorship:  "Mill.",
			expectedCardinality: 2,
			expectedNameStatus:  coldp.UnknownNomStatus,
			shouldFail:          false,
		},
		{
			msg:                  "trinomial with variety",
			name:                 "Abies alba var. pyramidalis Carrière",
			citation:             "",
			expectedGenus:        "Abies",
			expectedSpecies:      "alba",
			expectedInfraspecies: "pyramidalis",
			expectedAuthorship:   "Carrière",
			expectedCardinality:  3,
			expectedNameStatus:   coldp.UnknownNomStatus,
			shouldFail:           false,
		},
		{
			msg:                    "manuscript name with ined.",
			name:                   "Abies alba ined.",
			citation:               "",
			expectedGenus:          "Abies",
			expectedSpecies:        "alba",
			expectedAuthorship:     "",
			expectedCardinality:    2,
			expectedNameStatus:     coldp.Manuscript,
			expectedAppendedPhrase: "ined.",
			shouldFail:             false,
		},
		{
			msg:                    "manuscript name with comb. ined.",
			name:                   "Abies alba comb. ined.",
			citation:               "",
			expectedGenus:          "Abies",
			expectedSpecies:        "alba",
			expectedAuthorship:     "",
			expectedCardinality:    2,
			expectedNameStatus:     coldp.Manuscript,
			expectedAppendedPhrase: "comb. ined.",
			shouldFail:             false,
		},
		{
			msg:                 "combination with parentheses",
			name:                "Abies alba (L.) Mill.",
			citation:            "",
			expectedGenus:       "Abies",
			expectedSpecies:     "alba",
			expectedAuthorship:  "(L.) Mill.",
			expectedCardinality: 2,
			expectedNameStatus:  coldp.UnknownNomStatus,
			shouldFail:          false,
		},
		{
			msg:        "hybrid formula should fail",
			name:       "Abies alba × Abies pinsapo",
			citation:   "",
			shouldFail: true,
		},
	}

	for _, v := range tests {
		res, err := wfwpParse(parser, v.name, v.citation)

		if v.shouldFail {
			assert.Error(err, v.msg)
			continue
		}

		assert.NoError(err, v.msg)
		assert.Equal(v.expectedGenus, res.genus, v.msg+" - genus")
		assert.Equal(v.expectedSpecies, res.species, v.msg+" - species")
		assert.Equal(
			v.expectedInfraspecies,
			res.infraspecies,
			v.msg+" - infraspecies",
		)
		assert.Equal(
			v.expectedAuthorship,
			res.authorship,
			v.msg+" - authorship",
		)
		assert.Equal(
			v.expectedCardinality,
			res.cardinality,
			v.msg+" - cardinality",
		)
		assert.Equal(
			v.expectedNameStatus,
			res.nameStatus,
			v.msg+" - nameStatus",
		)
		if v.expectedAppendedPhrase != "" {
			assert.Equal(
				v.expectedAppendedPhrase,
				res.appendedPhrase,
				v.msg+" - appendedPhrase",
			)
		}
	}
}

func TestSynonymRank(t *testing.T) {
	assert := assert.New(t)
	parser := getTestParser()

	tests := []struct {
		msg          string
		name         string
		acceptedRank coldp.Rank
		expectedRank coldp.Rank
	}{
		{
			msg:          "binomial synonym of species",
			name:         "Abies alba Mill.",
			acceptedRank: coldp.Species,
			expectedRank: coldp.Species,
		},
		{
			msg:          "variety synonym of variety",
			name:         "Abies alba var. pyramidalis Carrière",
			acceptedRank: coldp.Variety,
			expectedRank: coldp.Variety,
		},
		{
			msg:          "subspecies synonym of species",
			name:         "Abies alba subsp. alpina L.",
			acceptedRank: coldp.Species,
			expectedRank: coldp.Subspecies,
		},
		{
			msg:          "form synonym of variety",
			name:         "Abies alba f. pendula Beissn.",
			acceptedRank: coldp.Variety,
			expectedRank: coldp.Form,
		},
	}

	for _, v := range tests {
		parsed, err := wfwpParse(parser, v.name, "")
		assert.NoError(err, v.msg)

		rank := synonymRank(parsed, v.acceptedRank)
		assert.Equal(v.expectedRank, rank, v.msg)
	}
}

func TestIsAutonym(t *testing.T) {
	assert := assert.New(t)
	parser := getTestParser()

	tests := []struct {
		msg        string
		name       string
		isAutonym  bool
		parentName string
	}{
		{
			msg:        "typical autonym",
			name:       "Abies alba subsp. alba",
			isAutonym:  true,
			parentName: "Abies",
		},
		{
			msg:        "not an autonym",
			name:       "Abies alba subsp. pyramidalis",
			isAutonym:  false,
			parentName: "Abies",
		},
		// TODO: Investigate why hybrid autonym detection fails
		// {
		// 	msg:        "hybrid autonym",
		// 	name:       "Abies ×alba var. ×alba",
		// 	isAutonym:  true,
		// 	parentName: "Abies",
		// },
	}

	for _, v := range tests {
		parsed, err := wfwpParse(parser, v.name, "")
		assert.NoError(err, v.msg)

		node := hNode{
			verbatimName: v.name,
			rank:         coldp.Subspecies,
			parsed:       parsed,
		}

		result := isAutonym(node)
		assert.Equal(v.isAutonym, result, v.msg)
	}
}

func TestGetBasionymId(t *testing.T) {
	assert := assert.New(t)
	parser := getTestParser()

	tests := []struct {
		msg        string
		name       string
		expectedId string
	}{
		{
			msg:        "simple binomial",
			name:       "Abies alba Mill.",
			expectedId: "alb_Mill.",
		},
		{
			msg:        "trinomial",
			name:       "Abies alba var. pyramidalis Carrière",
			expectedId: "pyramidal_Carrière",
		},
		{
			msg:        "combination authorship",
			name:       "Abies alba (L.) Mill.",
			expectedId: "alb_L.",
		},
	}

	for _, v := range tests {
		parsed, err := wfwpParse(parser, v.name, "")
		assert.NoError(err, v.msg)

		basionymId := getBasionymId(parsed)
		assert.Equal(v.expectedId, basionymId, v.msg)
	}
}
