package disk

import (
	"sync"

	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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

	wg sync.WaitGroup

	logger logr.Logger
}

func InitPayload(c PayloadConfig, logger logr.Logger) (*Payload, error) {
	switch c.Action {
	case Read:
		path, err := ReadPath(c.Path)
		if err != nil {
			return nil, err
		}
		byteSize, err := Count(c.Size, c.Percent, path)
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
		byteSize, err := Count(c.Size, c.Percent, path)
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

func (p *Payload) Inject(pid uint32) error {
	p.wg.Add(len(p.DdCmds))
	errs := make(chan error, len(p.DdCmds))

	for _, rawDD := range p.DdCmds {
		rawCmd, err := rawDD.ToCmd()
		if err != nil {
			return err
		}
		cmd := WrapCmd(rawCmd, pid)
		p.logger.Info(cmd.String())

		go func() {
			defer p.wg.Done()
			out, err := cmd.CombinedOutput()
			p.logger.Info(string(out))
			if err != nil {
				errs <- err
				return
			}
		}()

		if err != nil {
			return errors.WithStack(err)
		}
	}
	p.wg.Wait()
	close(errs)
	var result error
	for err := range errs {
		result = multierror.Append(result, err)
	}
	return errors.WithStack(result)
}
