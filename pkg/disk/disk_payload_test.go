package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/test/pkg/timer"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPayload_Inject(t *testing.T) {
	err := os.Chdir("../../")
	assert.NoError(t, err)
	timer, err := timer.StartTimer()
	assert.NoError(t, err)

	c := NewPayloadConfig(Read, CommonConfig{
		Path:      "",
		Size:      "100M",
		Percent:   "",
		SpaceLock: "",
	}, RuntimeConfig{
		ProcessNum:    4,
		LoopExecution: false,
	})

	logger, err := log.NewDefaultZapLogger()
	assert.NoError(t, err)

	p, err := InitPayload(c, logger)
	assert.NoError(t, err)

	err = p.Inject(uint32(timer.Pid()))
	assert.NoError(t, err)
}
