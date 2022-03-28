package disk

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"os/exec"
	"strconv"
)

type CommonConfig struct {
	Path string

	Size    string
	Percent string

	SpaceLock string
}

type FillConfig struct {
	CommonConfig
	FillByFAllocate bool
}

func NewFillConfig(fillByFallocate bool, c CommonConfig) FillConfig {
	return FillConfig{
		CommonConfig:    c,
		FillByFAllocate: fillByFallocate,
	}
}

type Fill struct {
	FillConfig
	DdCmds       []DD
	FallocateCmd *FAllocate

	logger logr.Logger
}

func InitFill(c FillConfig, logger logr.Logger) (*Fill, error) {
	path, err := WritePath(c.Path)
	if err != nil {
		return nil, err
	}
	byteSize, err := Count(c.Size, c.Percent, path)
	if c.FillByFAllocate {
		fallocateCmd := FAllocate{
			Exec:     command.NewExec(),
			Length:   strconv.FormatUint(byteSize, 10),
			FileName: path,
		}
		return &Fill{
			FillConfig:   c,
			DdCmds:       nil,
			FallocateCmd: &fallocateCmd,
			logger:       logger,
		}, nil
	} else {
		ddBlocks, err := SplitBytesByProcessNum(byteSize, 1)
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
				Iflag:     "fullblock", // fullblock : accumulate full blocks of input.
				Oflag:     "append",
				Conv:      "notrunc", // notrunc : do not truncate the output file.
			})
		}
		return &Fill{
			FillConfig:   c,
			DdCmds:       cmds,
			FallocateCmd: nil,
			logger:       logger,
		}, nil
	}
}

func WrapCmd(rawCmd *exec.Cmd, pid uint32) *bpm.ManagedCommand {
	// cmd.Args == path + args
	processBuilder := bpm.DefaultProcessBuilder(rawCmd.Path, rawCmd.Args[1:]...).
		EnableLocalMnt().
		SetNS(pid, bpm.MountNS).
		SetNS(pid, bpm.PidNS)

	return processBuilder.Build(context.Background())
}

func (f *Fill) Inject(pid uint32) error {
	if f.FillByFAllocate && f.FallocateCmd != nil {
		rawCmd, err := f.FallocateCmd.ToCmd()
		if err != nil {
			return err
		}
		cmd := WrapCmd(rawCmd, pid)
		f.logger.Info(cmd.String())
		out, err := cmd.CombinedOutput()
		f.logger.Info(string(out))
		if err != nil {
			return errors.WithStack(err)
		}
	} else if !f.FillByFAllocate && f.DdCmds != nil {
		for _, rawDD := range f.DdCmds {
			rawCmd, err := rawDD.ToCmd()
			if err != nil {
				return err
			}
			cmd := WrapCmd(rawCmd, pid)
			f.logger.Info(cmd.String())
			out, err := cmd.CombinedOutput()
			f.logger.Info(string(out))
			if err != nil {
				return errors.WithStack(err)
			}
		}
	} else {
		return errors.New("unexpected situation")
	}
	return nil
}
