package base_test

import (
	"testing"

	"github.com/sfborg/harvester/internal/io/base"
	"github.com/sfborg/harvester/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg, input, name, authorship, uninomial, species string
		rank                                             string
		quality                                          int
	}{
		{"uninomial", "Chalcodes Westwood", "Chalcodes", "Westwood",
			"Chalcodes", "", "", 1},
		{"infra", "Podocarpus pseudobracteatus var. sicaris de Laub.",
			"Podocarpus pseudobracteatus var. sicaris", "de Laub.", "",
			"pseudobracteatus", "var.", 1},
	}

	c := base.New(config.New(), nil)
	for _, v := range tests {
		parsed := c.Parse(v.input)
		assert.Equal(v.authorship, parsed.Authorship, v.msg)
		assert.Equal(v.uninomial, parsed.Uninomial, v.msg)
		assert.Equal(v.species, parsed.Species, v.msg)
		assert.Equal(v.rank, parsed.Rank, v.msg)
	}
}
