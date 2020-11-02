// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package fusedev

// #include <sys/sysmacros.h>
// #include <sys/types.h>
// // makedev is a macro, so a wrapper is needed
// dev_t Makedev(unsigned int maj, unsigned int min) {
//   return makedev(maj, min);
// }
import "C"
import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/pingcap/errors"
)

// EnsureFuseDev ensures /dev/fuse exists. If not, it will create one
func EnsureFuseDev() {
	if _, err := os.Open("/dev/fuse"); os.IsNotExist(err) {
		// 10, 229 according to https://www.kernel.org/doc/Documentation/admin-guide/devices.txt
		fuse := C.Makedev(10, 229)
		syscall.Mknod("/dev/fuse", 0o666|syscall.S_IFCHR, int(fuse))
	}
}

// GrantAccess appends 'c 10:229 rwm' to devices.allow
func GrantAccess() error {
	pid := os.Getpid()
	cgroupPath := fmt.Sprintf("/proc/%d/cgroup", pid)

	cgroupFile, err := os.Open(cgroupPath)
	if err != nil {
		return err
	}

	// TODO: encapsulate these logic with chaos-daemon StressChaos part
	cgroupScanner := bufio.NewScanner(cgroupFile)
	var deviceCgroupPath string
	for cgroupScanner.Scan() {
		if err := cgroupScanner.Err(); err != nil {
			return err
		}
		var (
			text  = cgroupScanner.Text()
			parts = strings.SplitN(text, ":", 3)
		)
		if len(parts) < 3 {
			return errors.Errorf("invalid cgroup entry: %q", text)
		}

		if parts[1] == "devices" {
			deviceCgroupPath = parts[2]
		}
	}

	if len(deviceCgroupPath) == 0 {
		return errors.Errorf("fail to find device cgroup")
	}

	deviceCgroupPath = "/sys/fs/cgroup/devices" + deviceCgroupPath + "/devices.allow"
	f, err := os.OpenFile(deviceCgroupPath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	// 10, 229 according to https://www.kernel.org/doc/Documentation/admin-guide/devices.txt
	content := "c 10:229 rwm"
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
