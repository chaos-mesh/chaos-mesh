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

// +build cgo

package ptrace

/*
#define _GNU_SOURCE
#include <sys/wait.h>
#include <sys/uio.h>
#include <errno.h>
*/
import "C"

func waitpid(pid int) int {
	return int(C.waitpid(C.int(pid), nil, C.__WALL))
}
