package errors

import (
	"errors"
)

var (
	ErrNoNeedSchedule = errors.New("no need to schedule")
	ErrNoSuchNode     = errors.New("no such node")
	ErrNoTemplates    = errors.New("no templates in workflow spec")
)
