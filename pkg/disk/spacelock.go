package disk

import (
	"github.com/pingcap/errors"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
)

type SpaceLock struct {
	size string
	file *os.File
}

func NewSpaceLock(size string) SpaceLock {
	if size == "" {
		size = "0"
	}
	return SpaceLock{
		size: size,
	}
}

const RandomFileNameSize = 24

func (l *SpaceLock) Lock() error {
	bSize, err := ParseUnit(l.size)
	if err != nil {
		return err
	}

	f, err := os.Create(rand.String(RandomFileNameSize))
	if err != nil {
		return err
	}

	if uint64(int64(bSize)) != bSize {
		os.Remove(f.Name())
		return errors.New("cannot Truncate, because size is too big")
	}
	if err := f.Truncate(int64(bSize)); err != nil {
		os.Remove(f.Name())
		return err
	}
	l.file = f
	return nil
}

func (l *SpaceLock) Unlock() error {
	return os.Remove(l.file.Name())
}
