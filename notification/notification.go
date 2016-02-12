package notification

import (
	"github.com/juju/errgo"
)

var (
	ErrNotConfigured        = errgo.Newf("No notification receiver configured")
	ErrInvalidConfiguration = errgo.Newf("Invalid configuration")
)

func IsNotConfigured(err error) bool {
	return errgo.Cause(err) == ErrNotConfigured
}

func IsInvalidConfiguration(err error) bool {
	return errgo.Cause(err) == ErrInvalidConfiguration
}
