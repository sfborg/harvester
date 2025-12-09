package errcode_test

import (
	"errors"
	"testing"

	"github.com/gnames/gn"
	"github.com/sfborg/harvester/pkg/errcode"
	"github.com/stretchr/testify/assert"
)

func TestIs(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg  string
		err  error
		code gn.ErrorCode
		res  bool
	}{
		{"not gn", errors.New("some error"), errcode.UnknownError, false},
		{"gn nil code", &gn.Error{}, errcode.UnknownError, true},
		{"gn nil code2", &gn.Error{}, errcode.WikispSkipPage, false},
		{"gn code", &gn.Error{Code: errcode.CreateDirError}, errcode.WikispSkipPage, false},
		{"gn code", &gn.Error{Code: errcode.WikispSkipPage}, errcode.WikispSkipPage, true},
	}

	for _, v := range tests {
		assert.Equal(v.res, errcode.Is(v.err, v.code))
	}
}
