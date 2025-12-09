package errcode

import (
	"errors"

	"github.com/gnames/gn"
)

const (
	UnknownError gn.ErrorCode = iota

	// File System errors
	CreateDirError
	OpenFileError

	// Wikisp
	WikispSkipPage
)

func Is(err error, code gn.ErrorCode) bool {
	var gnErr *gn.Error
	if errors.As(err, &gnErr) {
		return gnErr.Code == code
	}
	return false
}
