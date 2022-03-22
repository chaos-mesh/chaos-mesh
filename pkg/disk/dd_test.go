package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"github.com/stretchr/testify/assert"
	"testing"
)

//TODO : check
func TestDD_ToCmd(t *testing.T) {
	dd := DD{
		Exec:      command.NewExec(),
		ReadPath:  "/dev/zero",
		WritePath: "/dev/null",
		BlockSize: "20M",
		Count:     "400",
		Iflag:     "dsync,fullblock,nocache",
	}
	cmd, err := dd.ToCmd()
	assert.NoError(t, err)
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, string(output))
}
