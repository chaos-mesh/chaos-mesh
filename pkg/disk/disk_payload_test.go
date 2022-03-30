package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/test/pkg/timer"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

func TestPayload_Inject(t *testing.T) {
	err := os.Chdir("../../")
	assert.NoError(t, err)
	timer, err := timer.StartTimer()
	assert.NoError(t, err)

	c := NewPayloadConfig(Write, CommonConfig{
		Path:    "",
		Size:    "10G",
		Percent: "",
		SLock:   NewSpaceLock("50M"),
	}, RuntimeConfig{
		ProcessNum:    2,
		LoopExecution: false,
	})

	logger, err := log.NewDefaultZapLogger()
	assert.NoError(t, err)

	p, err := InitPayload(c, logger)
	assert.NoError(t, err)
	pid := timer.Pid()
	var wg sync.WaitGroup
	wg.Add(1)

	err = p.Inject(uint32(pid))
	assert.NoError(t, err)

	p.Recover()
}
