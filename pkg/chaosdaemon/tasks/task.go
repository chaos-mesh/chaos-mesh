package tasks

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

var ErrCanNotAdd = errors.New("can not add")
var ErrCanNotAssign = errors.New("can not assign")

type Addable interface {
	Add(a Addable) error
}

type NewChaosOnProcess interface {
	New(logger logr.Logger) (ChaosOnProcess, error)
}

type AssignChaosOnProcess interface {
	Assign(ChaosOnProcess) error
}

type Tasker interface {
	Addable
	NewChaosOnProcess
	AssignChaosOnProcess
}
