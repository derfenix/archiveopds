package domerr

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidPath    = errors.New("invalid path")
	ErrArchiveOpen    = errors.New("archive open failed")
	ErrUnsupportedFmt = errors.New("unsupported format")
	ErrBookTooLarge   = errors.New("book exceeds configured size limit")
)
