package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"github.com/go-logr/logr"
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

type Payload struct {
	PayloadConfig
	DdCmds []DD

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
