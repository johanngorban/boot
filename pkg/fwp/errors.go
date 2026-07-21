package fwp

import "errors"

var (
	ErrPayloadOverflow = errors.New("payload is too long")
)
