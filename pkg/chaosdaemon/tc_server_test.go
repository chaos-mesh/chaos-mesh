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

package chaosdaemon

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

func commonTcTest(t *testing.T, fpname, errString string, tcFunc func(s *daemonServer) error) {
	g := NewWithT(t)

	defer mock.With("MockContainerdClient", &MockClient{})()
	c, _ := CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
	s := &daemonServer{c}

	if errString == "" {
		defer mock.With(fpname, true)()
	} else {
		defer mock.With(fpname, errors.New(errString))()
	}

	err := tcFunc(s)

	if errString == "" {
		g.Expect(err).To(BeNil())
	} else {
		g.Expect(err).ToNot(BeNil())
		g.Expect(err.Error()).To(ContainSubstring(errString))
	}
}

func Test_daemonServer_AddQdisc(t *testing.T) {
	req := &pb.QdiscRequest{
		Qdisc: &pb.Qdisc{
			Type: "netem",
		},
		ContainerId: "containerd://container-id",
	}

	tests := []struct {
		tname     string
		fpname    string
		errString string
	}{
		{"should work", "TcApplyError", ""},
		{"should fail on get pid", "TaskError", "mock error on Task()"},
		{"should fail on apply tc", "TcApplyError", "mock error on applyTc()"},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.tname, func(t *testing.T) {
			commonTcTest(t, tt.fpname, tt.errString, func(s *daemonServer) error {
				_, err := s.AddQdisc(context.TODO(), req)
				return err
			})
		})
	}
}

func Test_daemonServer_DelQdisc(t *testing.T) {
	req := &pb.QdiscRequest{
		Qdisc: &pb.Qdisc{
			Type: "netem",
		},
		ContainerId: "containerd://container-id",
	}

	tests := []struct {
		tname     string
		fpname    string
		errString string
	}{
		{"should work", "TcApplyError", ""},
		{"should fail on get pid", "TaskError", "mock error on Task()"},
		{"should fail on apply tc", "TcApplyError", "mock error on applyTc()"},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.tname, func(t *testing.T) {
			commonTcTest(t, tt.fpname, tt.errString, func(s *daemonServer) error {
				_, err := s.DelQdisc(context.TODO(), req)
				return err
			})
		})
	}
}

func Test_daemonServer_AddEmatchFilter(t *testing.T) {
	req := &pb.EmatchFilterRequest{
		Filter: &pb.EmatchFilter{
			Match:   "test",
			Parent:  &pb.TcHandle{Major: 1, Minor: 0},
			Classid: &pb.TcHandle{Major: 10, Minor: 0},
		},
		ContainerId: "containerd://container-id",
	}

	tests := []struct {
		tname     string
		fpname    string
		errString string
	}{
		{"should work", "TcApplyError", ""},
		{"should fail on get pid", "TaskError", "mock error on Task()"},
		{"should fail on apply tc", "TcApplyError", "mock error on applyTc()"},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.tname, func(t *testing.T) {
			commonTcTest(t, tt.fpname, tt.errString, func(s *daemonServer) error {
				_, err := s.AddEmatchFilter(context.TODO(), req)
				return err
			})
		})
	}
}

func Test_daemonServer_DelTcFilter(t *testing.T) {
	req := &pb.TcFilterRequest{
		Filter: &pb.TcFilter{
			Parent: &pb.TcHandle{Major: 1, Minor: 0},
		},
		ContainerId: "containerd://container-id",
	}

	tests := []struct {
		tname     string
		fpname    string
		errString string
	}{
		{"should work", "TcApplyError", ""},
		{"should fail on get pid", "TaskError", "mock error on Task()"},
		{"should fail on apply tc", "TcApplyError", "mock error on applyTc()"},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.tname, func(t *testing.T) {
			commonTcTest(t, tt.fpname, tt.errString, func(s *daemonServer) error {
				_, err := s.DelTcFilter(context.TODO(), req)
				return err
			})
		})
	}
}

func Test_generateQdiscArgs(t *testing.T) {
	g := NewWithT(t)

	typ := "netem"

	t.Run("without parent and handle", func(t *testing.T) {

		args, err := generateQdiscArgs("add", &pb.Qdisc{Type: typ})

		g.Expect(err).To(BeNil())
		g.Expect(args).To(Equal([]string{"qdisc", "add", "dev", "eth0", "root", "handle", "1:0", typ}))
	})

	t.Run("with parent and handle", func(t *testing.T) {
		args, err := generateQdiscArgs("add", &pb.Qdisc{
			Type: typ,
			Parent: &pb.TcHandle{
				Major: 1,
				Minor: 1,
			},
			Handle: &pb.TcHandle{
				Major: 10,
				Minor: 0,
			},
		})

		g.Expect(err).To(BeNil())
		g.Expect(args).To(Equal([]string{"qdisc", "add", "dev", "eth0", "parent", "1:1", "handle", "10:0", typ}))
	})
}
