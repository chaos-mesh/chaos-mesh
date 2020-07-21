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

package watcher

import (
	"html/template"
	"testing"
)

func TestRenderTemplateWithArgs(t *testing.T) {
	tmpl := template.Must(template.New("common-template").Parse(`initContainers:
- name: inject-scripts
  image: pingcap/chaos-scripts:latest
  imagePullpolicy: Always
  command: ["sh", "-c", "/scripts/init.sh -d {{.DataPath}} -f {{.MountPath}}/fuse-data"]
containers:
- name: chaosfs
  image: pingcap/chaos-fs:latest
  imagePullpolicy: Always
  ports:
  - containerPort: 65534
  securityContext:
    privileged: true
  command:
    - /usr/local/bin/chaosfs
    - -addr=:65534
    - -pidfile=/tmp/fuse/pid
    - -original={{.MountPath}}/fuse-data
    - -mountpoint={{.DataPath}}
  volumeMounts:
  - name: {{.VolumeName}}
    mountPath: {{.MountPath}}
    mountPropagation: Bidirectional
volumeMounts:
- name: {{.VolumeName}}
  mountPath: {{.MountPath}}
  mountPropagation: HostToContainer
- name: scripts
  mountPath: /tmp/scripts
- name: fuse
  mountPath: /tmp/fuse
volumes:
- name: scripts
  emptyDir: {}
- name: fuse
  emptyDir: {}
postStart:
  {{.ContainerName}}:
    command:
      - /tmp/scripts/wait-fuse.sh`))

	args := map[string]string{
		"DataPath":      "/var/lib/pd/data",
		"VolumeName":    "pd",
		"MountPath":     "/var/lib/pd",
		"ContainerName": "pd",
	}
	out, err := renderTemplateWithArgs(tmpl, args)
	if err != nil {
		t.Error("failed to render template", err)
	}
	expected := `initContainers:
- name: inject-scripts
  image: pingcap/chaos-scripts:latest
  imagePullpolicy: Always
  command: ["sh", "-c", "/scripts/init.sh -d /var/lib/pd/data -f /var/lib/pd/fuse-data"]
containers:
- name: chaosfs
  image: pingcap/chaos-fs:latest
  imagePullpolicy: Always
  ports:
  - containerPort: 65534
  securityContext:
    privileged: true
  command:
    - /usr/local/bin/chaosfs
    - -addr=:65534
    - -pidfile=/tmp/fuse/pid
    - -original=/var/lib/pd/fuse-data
    - -mountpoint=/var/lib/pd/data
  volumeMounts:
  - name: pd
    mountPath: /var/lib/pd
    mountPropagation: Bidirectional
volumeMounts:
- name: pd
  mountPath: /var/lib/pd
  mountPropagation: HostToContainer
- name: scripts
  mountPath: /tmp/scripts
- name: fuse
  mountPath: /tmp/fuse
volumes:
- name: scripts
  emptyDir: {}
- name: fuse
  emptyDir: {}
postStart:
  pd:
    command:
      - /tmp/scripts/wait-fuse.sh`
	if string(out) != expected {
		t.Error("expected to get", expected, "but got", string(out))
	}
}
