package disk

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/command"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const DevZero = "/dev/zero"
const DevNull = "/dev/null"

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

func WritePath(path string) (string, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return "", errors.WithStack(err)
		}
	}

	fi, err := os.Stat(path)
	if err != nil {
		// check if Path of file is valid when Path is not empty
		if os.IsNotExist(err) {
			var b []byte
			if err := ioutil.WriteFile(path, b, 0600); err != nil {
				return "", errors.WithStack(err)
			}
			if err := os.Remove(path); err != nil {
				return "", errors.WithStack(err)
			}
			return path, nil
		}
		return "", errors.WithStack(err)
	}
	if fi.IsDir() {
		path, err = CreateTempFile(path)
		if err != nil {
			return "", err
		}
		if err := os.Remove(path); err != nil {
			return "", errors.WithStack(err)
		}
		return path, err
	}
	return "", errors.New("write path cannot be a existing file")
}

func ReadPath(path string) (string, error) {
	if path == "" {
		path, err := GetRootDevice()
		if err != nil {
			return "", err
		}
		if path == "" {
			err = errors.Errorf("can not get root device path")
			return "", err
		}
		return path, nil
	}
	var fi os.FileInfo
	var err error
	if fi, err = os.Stat(path); err != nil {
		return "", errors.WithStack(err)
	}
	if fi.IsDir() {
		return "", errors.Errorf("path is a dictory, path : %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = f.Close()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return path, nil
}

func Count(size string, percent string, path string) (uint64, error) {
	if size != "" {
		byteSize, err := ParseUnit(size)
		if err != nil {
			return 0, err
		}
		return byteSize, nil
	} else if percent != "" {
		percent = strings.Trim(percent, " %")
		percent, err := strconv.ParseUint(percent, 10, 0)
		if err != nil {
			return 0, errors.WithStack(err)
		}
		dir := filepath.Dir(path)
		totalSize, err := GetDiskTotalSize(dir)
		if err != nil {
			return 0, err
		}
		return totalSize * percent / 100, nil
	}

	return 0, errors.New("one of percent and size must not be empty")
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
