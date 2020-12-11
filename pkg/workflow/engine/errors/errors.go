package errors

import (
	"errors"
)

var (
	ErrNoNeedSchedule = errors.New("no need to schedule")
	ErrNoSuchNode     = errors.New("no such node")
)
