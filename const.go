package gcm

import (
	"errors"
)

const (
	maxCapacity     = 512
	maxMultiplicity = 128
)

var (
	ErrFuncExists    = errors.New("speficied function already exists")
	ErrFuncNotExists = errors.New("speficied function not found")
)

type Status int

const (
	Running = iota + 1
	Stopping
	Stopped
)
