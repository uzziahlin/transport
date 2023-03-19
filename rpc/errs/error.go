package errs

import "errors"

var (
	ErrOneway = errors.New("micro: oneway error")
)
