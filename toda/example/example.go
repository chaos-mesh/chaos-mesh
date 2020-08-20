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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	err := ioutil.WriteFile("/mnt/test/test", []byte("HELLO WORLD"), 0644)
	if err != nil {
		fmt.Printf("Error: %v+", err)
		return
	}

	f, err := os.Open("/mnt/test/test")
	if err != nil {
		fmt.Printf("Error: %v+", err)
		return
	}

	for {
		time.Sleep(time.Second)

		buf := make([]byte, 5)
		n, err := f.Read(buf)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("%v %d bytes: %s\n", time.Now(), n, string(buf[:n]))

		_, err = f.Seek(0, 0)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
	}
}
