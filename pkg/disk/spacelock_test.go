package disk

import (
	"testing"
)

func TestSpaceLock_Lock(t *testing.T) {
	sLock := NewSpaceLock("20M")
	sLock.Lock()
	sLock.Unlock()
}
