package disk

import (
	"github.com/pingcap/errors"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
)

type SpaceLock struct {
	file *os.File
}

const RandomFileSize = 24

func (l *SpaceLock) Lock(size string) error {
	bSize, err := ParseUnit(size)
	if err != nil {
		return err
	}

	f, err := os.Create(rand.String(RandomFileSize))
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
