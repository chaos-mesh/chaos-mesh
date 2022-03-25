package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"os/exec"
)

type FAllocate struct {
	command.Exec `exec:"fallocate"`
	Length       string `para:"-l"`
	FileName     string `para:""`
}

func (c FAllocate) ToCmd() (*exec.Cmd, error) {
	path, fields, err := command.Marshal(c)
	if err != nil {
		return nil, err
	}

	params := make([]string, len(fields))
	for i := range params {
		params[i] = fields[i].Join(" ", "")
	}

	return exec.Command(path, params...), nil
}
