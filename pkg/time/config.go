package time

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

type SkewConfig struct {
	deltaSeconds     int64
	deltaNanoSeconds int64
	clockIDsMask     uint64
}

func (c *SkewConfig) Add(a tasks.Addable) error {
	if a, ok := a.(*SkewConfig); ok {
		c.deltaSeconds += a.deltaSeconds
		c.deltaNanoSeconds += a.deltaNanoSeconds
		return nil
	}
	return errors.Wrapf(tasks.ErrCanNotAdd, "type1 : %T, type2 : %T", c, a)
}

func (c *SkewConfig) New(logger logr.Logger) (tasks.ChaosOnProcess, error) {
	skew, err := NewTimeSkew(c.deltaSeconds, c.deltaNanoSeconds, c.clockIDsMask)
	if err != nil {
		return nil, err
	}
	gp := tasks.NewGroupProcessHandler(logger, skew)
	return &gp, nil
}

func (c *SkewConfig) Assign(task tasks.ChaosOnProcess) error {
	if t, ok := task.(*tasks.GroupProcessHandler); ok {
		if p, ok := t.Main.(*Skew); ok {
			p.deltaSeconds = c.deltaSeconds
			p.deltaNanoSeconds = c.deltaNanoSeconds
			return nil
		}
		return errors.Wrapf(tasks.ErrCanNotAssign, "expect type : *Skew, got : %T.", task)
	}
	return errors.Wrapf(tasks.ErrCanNotAssign, "expect type : *tasks.GroupProcessHandler, got : %T.", task)
}
