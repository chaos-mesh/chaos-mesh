package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"os/exec"
)

type DD struct {
	command.Exec `exec:"dd"`
	ReadPath     string `para:"if"`
	WritePath    string `para:"of"`
	BlockSize    string `para:"bs"`
	Count        string `para:"count"`
	Iflag        string `para:"iflag"`
	Oflag        string `para:"oflag"`
	Conv         string `para:"conv"`
}

func (c DD) ToCmd() (*exec.Cmd, error) {
	path, fields, err := command.Marshal(c)
	if err != nil {
		return nil, err
	}

	params := make([]string, len(fields))
	for i := range params {
		params[i] = fields[i].Join("=", "")
	}

	return exec.Command(path, params...), nil
}
