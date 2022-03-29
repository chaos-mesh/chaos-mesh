package disk

import (
	"bytes"
	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type RuntimeConfig struct {
	ProcessNum    uint8
	LoopExecution bool
}

type PayloadAction int

const (
	Read PayloadAction = iota
	Write
)

type PayloadConfig struct {
	Action PayloadAction
	CommonConfig
	RuntimeConfig
}

func NewPayloadConfig(action PayloadAction, c CommonConfig, rc RuntimeConfig) PayloadConfig {
	return PayloadConfig{
		Action:        action,
		CommonConfig:  c,
		RuntimeConfig: rc,
	}
}

type Payload struct {
	PayloadConfig
	DdCmds []DD

	wg        sync.WaitGroup
	processes []*exec.Cmd

	locker sync.Mutex

	logger logr.Logger
}

func InitPayload(c PayloadConfig, logger logr.Logger) (*Payload, error) {
	switch c.Action {
	case Read:
		path, err := ReadPath(c.Path)
		if err != nil {
			return nil, err
		}
		c.Path = path

		byteSize, err := Count(c.Size, c.Percent, c.Path)
		if err != nil {
			return nil, err
		}
		ddBlocks, err := SplitBytesByProcessNum(byteSize, c.ProcessNum)
		if err != nil {
			return nil, err
		}

		var cmds []DD

		for _, block := range ddBlocks {
			cmds = append(cmds, DD{
				Exec:      command.NewExec(),
				ReadPath:  path,
				WritePath: DevNull,
				BlockSize: block.BlockSize,
				Count:     block.Count,
				Iflag:     "dsync,fullblock,nocache", // nocache : Request to drop cache.
			})
		}
		return &Payload{
			PayloadConfig: c,
			DdCmds:        cmds,
			logger:        logger,
		}, nil
	case Write:
		path, err := WritePath(c.Path)
		if err != nil {
			return nil, err
		}
		c.Path = path

		byteSize, err := Count(c.Size, c.Percent, c.Path)
		if err != nil {
			return nil, err
		}
		ddBlocks, err := SplitBytesByProcessNum(byteSize, c.ProcessNum)
		if err != nil {
			return nil, err
		}

		var cmds []DD

		for _, block := range ddBlocks {
			cmds = append(cmds, DD{
				Exec:      command.NewExec(),
				ReadPath:  DevZero,
				WritePath: path,
				BlockSize: block.BlockSize,
				Count:     block.Count,
				Oflag:     "dsync", // dsync : use synchronized I/O for data.
			})
		}
		return &Payload{
			PayloadConfig: c,
			DdCmds:        cmds,
			logger:        logger,
		}, nil
	default:
		return nil, errors.New("action must be Read or Write")
	}
}

func StartCmd(c *exec.Cmd) (*bytes.Buffer, error) {
	if c.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if c.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	err := c.Start()
	return &b, errors.WithStack(err)
}

func (p *Payload) Inject(pid uint32) error {
	p.locker.Lock()
	cmds := make([]*exec.Cmd, len(p.DdCmds))
	for i, rawDD := range p.DdCmds {
		rawCmd, err := rawDD.ToCmd()
		if err != nil {
			return err
		}
		cmds[i] = WrapCmd(rawCmd, pid)
	}

	testCmd := cmds[len(cmds)-1]
	p.logger.Info(testCmd.String())
	out, err := testCmd.CombinedOutput()
	p.logger.Info(string(out))
	if err != nil {
		p.locker.Unlock()
		err = multierror.Append(err, p.Recover())
		return errors.Wrap(err, string(out))
	}

	cmds = cmds[:len(cmds)-1]
	p.wg.Add(len(cmds))
	errs := make(chan error, len(cmds))
	p.processes = make([]*exec.Cmd, len(cmds))
	for i, cmd := range cmds {
		p.logger.Info(cmd.String())
		out, err := StartCmd(cmd)
		if err != nil {
			p.locker.Unlock()
			err = multierror.Append(err, p.Recover())
			return err
		}
		p.processes[i] = cmd

		cmd := cmd
		go func() {
			defer p.wg.Done()
			err = cmd.Wait()
			p.logger.Info(out.String())
			if err != nil {
				errs <- err
				return
			}
		}()
	}
	p.locker.Unlock()
	p.logger.Info("UNLOCK")
	p.wg.Wait()
	close(errs)
	var result error
	for err := range errs {
		result = multierror.Append(result, err)
	}
	return errors.WithStack(result)
}

const WaitTime = time.Second * 10

func (p *Payload) Recover() error {
	p.locker.Lock()
	p.logger.Info("RECOVER")
	defer p.locker.Unlock()

	var result error

	for _, process := range p.processes {
		if process != nil {
			if process.Process != nil {
				if err := process.Process.Signal(syscall.SIGTERM); err != nil {
					result = multierror.Append(result, errors.WithStack(err))
				} else {
					_, err := process.Process.Wait()
					if err != nil {
						result = multierror.Append(result, errors.WithStack(err))
					}
				}
			} else {
				p.logger.Info("I don't know why , but process.Process is nil")
			}
		} else {
			p.logger.Info("I don't know why , but process is nil")
		}

	}
	p.processes = []*exec.Cmd{}

	if p.Action == Write {
		if _, err := os.Stat(p.Path); err == nil {
			return os.Remove(p.Path)
		} else if errors.Is(err, os.ErrNotExist) {
			return nil
		} else {
			return errors.WithStack(err)
		}
	}

	return errors.WithStack(result)
}
