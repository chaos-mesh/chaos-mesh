// Copyright 2026 Chaos Mesh Authors.
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

package chaosdaemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDataIntoFileUsesPrivateTempDir(t *testing.T) {
	const rule = "RULE test\nIF true\nDO traceln(\"test\")\nENDRULE\n"

	filename, err := writeDataIntoFile(rule, "rule.btm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeDataFile(filename)

	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected rule file mode 0600, got %#o", info.Mode().Perm())
	}

	dir := filepath.Dir(filename)
	if dir == os.TempDir() {
		t.Fatalf("expected rule file under a daemon-private directory, got %s", dir)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0700 {
		t.Fatalf("expected rule directory mode 0700, got %#o", dirInfo.Mode().Perm())
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != rule {
		t.Fatalf("expected rule data %q, got %q", rule, string(data))
	}
}

func TestRemoveDataFileDoesNotRemovePrefixedDirOutsideTempDir(t *testing.T) {
	dir, err := os.MkdirTemp(".", jvmRuleTempDirPrefix)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	file, err := os.CreateTemp(dir, "rule.btm")
	if err != nil {
		t.Fatal(err)
	}
	filename := file.Name()
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	removeDataFile(filename)

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected cleanup to preserve non-temp parent directory: %v", err)
	}
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Fatalf("expected cleanup to remove only the rule file, got err %v", err)
	}
}
