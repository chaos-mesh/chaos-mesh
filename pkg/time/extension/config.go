package extension

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/chaos-mesh/chaos-mesh/pkg/time"
	"github.com/pkg/errors"
)

type Config struct {
	inner time.Config
}

func (c *Config) DeepCopy() tasks.Object {
	return &Config{*c.inner.DeepCopy().(*time.Config)}
}

func (c *Config) Add(a tasks.Addable) error {
	A, OK := a.(*Config)
	if OK {
		err := c.inner.Add(&A.inner)
		if err != nil {
			return err
		}
		return nil
	}

	return errors.Wrapf(tasks.ErrCanNotAdd, "expect type : *extension.Config, got : %T", a)
}

func (c *Config) New(values interface{}) (tasks.Injectable, error) {
	skew, err := time.NewSkew()
	if err != nil {
		return nil, err
	}
	skew.SkewConfig = *c.inner.DeepCopy().(*time.Config)
	groupProcessHandler, ok := values.(*tasks.ProcessGroupHandler)
	if !ok {
		return nil, errors.New(fmt.Sprintf("type %t is not *tasks.ProcessGroupHandler", values))
	}
	_, ok = groupProcessHandler.Main.(*time.Skew)
	if !ok {
		return nil, errors.New(fmt.Sprintf("type %t is not *Skew", groupProcessHandler.Main))
	}
	newGroupProcessHandler :=
		tasks.NewProcessGroupHandler(groupProcessHandler.Logger, &skew)
	return &newGroupProcessHandler, nil
}

func (c *Config) Assign(injectable tasks.Injectable) error {
	groupProcessHandler, ok := injectable.(*tasks.ProcessGroupHandler)
	if !ok {
		return errors.New(fmt.Sprintf("type %t is not *tasks.ProcessGroupHandler", injectable))
	}
	I, ok := groupProcessHandler.Main.(*time.Skew)
	if !ok {
		return errors.New(fmt.Sprintf("type %t is not *Skew", groupProcessHandler.Main))
	}

	I.SkewConfig = (*c).inner
	return nil
}
