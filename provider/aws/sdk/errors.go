package sdk

import "github.com/juju/errgo"

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)
