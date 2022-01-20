package tasks

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type FakeConfig struct {
	i int
}

func (f *FakeConfig) Add(a Addable) error {
	A, OK := a.(*FakeConfig)
	if OK {
		f.i += A.i
		return nil
	}
	return errors.Wrapf(ErrCanNotAdd, "expect type : *FakeConfig, got : %T", a)
}

func (f *FakeConfig) Assign(c ChaosOnProcess) error {
	C, OK := c.(*FakeChaos)
	if OK {
		C.c = *f
		return nil
	}
	return errors.Wrapf(ErrCanNotAssign, "expect type : *FakeChaos, got : %T", c)
}

func (f *FakeConfig) New(logger logr.Logger) (ChaosOnProcess, error) {
	return &FakeChaos{
		c:      *f,
		logger: logger,
		j:      1,
	}, nil
}

type FakeChaos struct {
	c      FakeConfig
	logger logr.Logger
	j      int
}

func (f *FakeChaos) Inject(pid PID) error {
	f.logger.Info("inject", "pid", pid, "FakeChaos", f.c.i)
	return nil
}

func (f *FakeChaos) Recover(pid PID) error {
	f.logger.Info("recover", "pid", pid, "FakeChaos", f.c.i)
	return nil
}

func TestMAA(t *testing.T) {
	var log logr.Logger

	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)

	m := NewChaosOnProcessManager(log)
	fmt.Println(m.Apply("1", 1, &FakeConfig{i: 1}))
	fmt.Println(m.Apply("2", 1, &FakeConfig{i: 1}))
	fmt.Println("up", m.Update("1", 1, &FakeConfig{i: 2}))
	fmt.Println("rec", m.Recover("1", 1))
	fmt.Println("rec", m.Recover("2", 1))
	fmt.Println(m.Apply("1", 1, &FakeConfig{i: 1}))
	fmt.Println(m.Recover("1", 1))
}
