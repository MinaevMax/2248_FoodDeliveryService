package service_errors

import (
	"errors"
)

var (
	ErrEnvNotSet         = errors.New("environment variable is not set")
	ErrEnvUnexpectedBool = errors.New("forbidden environment bool variable set")
)
