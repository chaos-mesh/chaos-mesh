package disk

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/disk"
	"io/ioutil"
	"strconv"
	"syscall"
)

// DdArgBlock is command arg for dd. BlockSize is bs.Count is count.
type DdArgBlock struct {
	BlockSize string
	Count     string
}

// SplitBytesByProcessNum split bytes in to processNum + 1 dd arg blocks.
// Every ddArgBlock can generate one dd command.
// If bytes is bigger than processNum M ,
// bytes will be split into processNum dd commands with bs = 1M ,count = bytes/ processNum M.
// If bytes is not bigger than processNum M ,
// bytes will be split into processNum dd commands with bs = bytes / uint64(processNum) ,count = 1.
// And one ddArgBlock stand by the rest bytes will also add to the end of slice,
// even if rest bytes = 0.
func SplitBytesByProcessNum(bytes uint64, processNum uint8) ([]DdArgBlock, error) {
	if bytes == 0 {
		return []DdArgBlock{{
			BlockSize: "1M",
			Count:     "0",
		}}, nil
	}
	if processNum == 0 {
		return nil, errors.Errorf("num must not be zero")
	}
	ddArgBlocks := make([]DdArgBlock, processNum)
	if bytes > uint64(processNum)*(1<<20) {
		count := (bytes >> 20) / uint64(processNum)
		for i := range ddArgBlocks {
			ddArgBlocks[i].Count = strconv.FormatUint(count, 10)
			ddArgBlocks[i].BlockSize = "1M"
			bytes -= count << 20
		}
	} else {
		blockSize := bytes / uint64(processNum)
		for i := range ddArgBlocks {
			ddArgBlocks[i].Count = "1"
			ddArgBlocks[i].BlockSize = strconv.FormatUint(blockSize, 10) + "c"
			bytes -= blockSize
		}
	}

	if bytes == 0 {
		ddArgBlocks = append(ddArgBlocks, DdArgBlock{
			Count:     "0",
			BlockSize: "1M",
		})
	} else {
		ddArgBlocks = append(ddArgBlocks, DdArgBlock{
			Count:     "1",
			BlockSize: strconv.FormatUint(bytes, 10) + "c",
		})
	}
	return ddArgBlocks, nil
}

// CreateTempFile will create a temp file in current directory.
func CreateTempFile(path string) (string, error) {
	tempFile, err := ioutil.TempFile(path, "example")
	if err != nil {
		return "", errors.WithStack(err)
	}

	if tempFile != nil {
		err = tempFile.Close()
		if err != nil {
			return "", errors.WithStack(err)
		}
	} else {
		err := errors.Errorf("unexpected err : file get from ioutil.TempFile is nil")
		return "", err
	}
	return tempFile.Name(), nil
}

// GetDiskTotalSize returns the total bytes in disk
func GetDiskTotalSize(path string) (total uint64, err error) {
	s := syscall.Statfs_t{}
	err = syscall.Statfs(path, &s)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	reservedBlocks := s.Bfree - s.Bavail
	total = uint64(s.Frsize) * (s.Blocks - reservedBlocks)
	return total, nil
}

// GetRootDevice returns the device which "/" mount on.
func GetRootDevice() (string, error) {
	mapStat, err := disk.Partitions(false)
	if err != nil {
		return "", errors.WithStack(err)
	}
	for _, stat := range mapStat {
		if stat.Mountpoint == "/" {
			return stat.Device, nil
		}
	}
	return "", nil
}
