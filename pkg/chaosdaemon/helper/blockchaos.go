// Copyright 2022 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package helper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/sys/mountinfo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var NormalizeVolumeNameCmd = &cobra.Command{
	Use:   "normalize-volume-name [path]",
	Short: "get the device name from the path",
	Long: `Get the device name from the path.
The path could be a directory, a partition, or a block device.
The block device name will be printed out.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(1)
		}

		volumePath := args[0]
		deviceName, err := normalizeVolumeName(volumePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println(deviceName)
	},
}

func normalizeVolumeName(volumePath string) (string, error) {
	// before resolving the soft link the volumePath inside the request have three possible situations:
	// 1. the volumePath is a partition of a block device, e.g. /dev/sda1, or /mnt/disks/ata-CT2000MX500SSD1_2117E599E804-part1
	// 2. the volumePath is a block file path, e.g. /dev/sda, or /mnt/disks/ata-CT2000MX500SSD1_2117E599E804
	// 3. the volumePath is a directory path, e.g. /var/lib/docker/volumes/my-volume
	//
	// if it's a partition of a block device, we need to convert it to the block file path
	// if it's a block device, the client library of chaos-driver can handle it
	// if it's a directory, chaos-daemon should automatically convert it to the corresponding block device name
	// For example, the return value of this function could be: `sda`, `sdb`, `nvme0n1`,

	volumePath, err := filepath.EvalSymlinks(volumePath)
	if err != nil {
		return "", errors.Wrapf(err, "resolving symlink %s", volumePath)
	}

	stat, err := os.Stat(volumePath)
	if err != nil {
		return "", errors.Wrapf(err, "getting stat of %s", volumePath)
	}

	if stat.IsDir() {
		parentMounts, err := mountinfo.GetMounts(mountinfo.ParentsFilter(volumePath))
		if err != nil {
			return "", errors.Wrap(err, "read mountinfo")
		}

		if len(parentMounts) == 0 {
			return "", errors.Errorf("cannot find the mount point which contains the volume path %s", volumePath)
		}

		bestMatch := &mountinfo.Info{}
		for _, mount := range parentMounts {
			mount := mount
			if len(mount.Mountpoint) > len(bestMatch.Mountpoint) {
				bestMatch = mount
			}
		}

		if bestMatch.Source == "none" || len(bestMatch.Source) == 0 {
			return "", errors.Errorf("unknown source of the mount point %v", bestMatch)
		}
		volumePath = bestMatch.Source
	}

	// now, the `volumePath` is either a partition, or a block device
	volumeName := filepath.Base(volumePath)

	// volumeName is either a partition (`sda1`, `nvme0n1p1`), or a block device (`sda`, `nvme0n1`)
	if _, err := os.Stat("/sys/block/" + volumeName); errors.Is(err, os.ErrNotExist) {
		// the volumeName is a partition, convert it to the block device name
		partitionSysPath, err := filepath.EvalSymlinks("/sys/class/block/" + volumeName)
		if err != nil {
			return "", errors.Wrapf(err, "resolving symlink %s", "/sys/class/block/"+volumeName)
		}

		volumeName = filepath.Base(filepath.Dir(partitionSysPath))
	}
	return volumeName, nil
}

// Manually test has been done for the following situations:
// 1. cdh normalize-volume-name /home, where /home is a simple directory
// 2. cdh normalize-volume-name /dev/vda1
// 3. cdh normalize-volume-name /dev/vda
