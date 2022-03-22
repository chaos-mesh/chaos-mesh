package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFAllocateOption_ToCmd(t *testing.T) {
	dd := FAllocateOption{
		Exec:     command.NewExec(),
		Length:   "20M",
		FileName: "/tmp/tmp_file",
	}
	_, err := dd.ToCmd()
	assert.NoError(t, err)
}
