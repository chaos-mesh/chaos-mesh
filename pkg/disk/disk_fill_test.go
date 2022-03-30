package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/test/pkg/timer"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFill_Inject(t *testing.T) {
	err := os.Chdir("../../")
	assert.NoError(t, err)
	timer, err := timer.StartTimer()
	assert.NoError(t, err)

	c := NewFillConfig(false, CommonConfig{
		Path:    "",
		Size:    "20M",
		Percent: "",
		SLock:   NewSpaceLock("50M"),
	})
	logger, err := log.NewDefaultZapLogger()
	assert.NoError(t, err)

	f, err := InitFill(c, logger)
	assert.NoError(t, err)

	err = f.Inject(uint32(timer.Pid()))
	assert.NoError(t, err)
	err = f.Recover()
	assert.NoError(t, err)
}
