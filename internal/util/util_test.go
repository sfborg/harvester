package util_test

import (
	"testing"

	"github.com/sfborg/harvester/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNormAuthorship(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		msg, input, output string
	}{
		{"simple", "Neumann, 1898", "Neumann, 1898"},
		{"1 init", "Rothschild & Chubb, C, 1914", "Rothschild & C. Chubb, 1914"},
		{"2 init", "Saunders, GH, 1876", "G.H. Saunders, 1876"},
		{"suffix", "(Grandidier, G Jr & Berlioz, 1929)",
			"(G. Grandidier Jr & Berlioz, 1929)"},
		{"no yr", "Saunders, G", "G. Saunders"},
		// {"no yr, paren", "(Saunders, G)", "(G. Saunders)"},
	}

	for _, v := range tests {
		res := util.NormalizeAuthors(v.input)
		assert.Equal(v.output, res, v.msg)
	}
}
