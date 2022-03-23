package disk

import (
	"testing"
)

func TestSpaceLock_Lock(t *testing.T) {
	var sLock SpaceLock
	sLock.Lock("20M")
	sLock.Unlock()
}
