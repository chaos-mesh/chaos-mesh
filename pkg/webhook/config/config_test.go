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

package config

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("webhook config", func() {
	Context("Test webhook config", func() {
		It("unmarshal TemplateArgs", func() {
			template := `
name: chaosfs-etcd
selector:
  labelSelectors:
    app: etcd
template: chaosfs-sidecar
arguments:
  ContainerName: "etcd"
  DataPath: "/var/run/etcd/default.etcd"
  MountPath: "/var/run/etcd"
  VolumeName: "datadir"`

			var cfg TemplateArgs
			err := yaml.Unmarshal([]byte(template), &cfg)
			Expect(err).To(BeNil())
		})

		It("unmarshal Injection Config", func() {
			template := `
initContainers:
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
      - /tmp/scripts/wait-fuse.sh
`
			var cfg InjectionConfig
			err := yaml.Unmarshal([]byte(template), &cfg)
			Expect(err).To(BeNil())
		})

		It("should return request on RequestAnnotationKey", func() {
			var cfg Config
			res := cfg.RequestAnnotationKey()
			Expect(res).To(Equal("/request"))
		})

		It("should return status on StatusAnnotationKey", func() {
			var cfg Config
			res := cfg.StatusAnnotationKey()
			Expect(res).To(Equal("/status"))
		})

		It("should return init-request on RequestInitAnnotationKey", func() {
			var cfg Config
			res := cfg.RequestInitAnnotationKey()
			Expect(res).To(Equal("/init-request"))
		})

	})
})
