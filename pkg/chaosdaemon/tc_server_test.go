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

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
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

func Test_convertNetemToArgs(t *testing.T) {
	g := NewWithT(t)

	t.Run("convert network delay", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Time: 1000,
		})
		g.Expect(args).To(Equal("delay 1000"))

		args = convertNetemToArgs(&pb.Netem{
			Time:      1000,
			DelayCorr: 25,
		})
		g.Expect(args).To(Equal("delay 1000"))

		args = convertNetemToArgs(&pb.Netem{
			Time:      1000,
			Jitter:    10000,
			DelayCorr: 25,
		})
		g.Expect(args).To(Equal("delay 1000 10000 25"))
	})

	t.Run("convert packet limit", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Limit: 1000,
		})
		g.Expect(args).To(Equal("limit 1000"))
	})

	t.Run("convert packet loss", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Loss: 100,
		})
		g.Expect(args).To(Equal("loss 100"))

		args = convertNetemToArgs(&pb.Netem{
			Loss:     50,
			LossCorr: 12,
		})
		g.Expect(args).To(Equal("loss 50 12"))
	})

	t.Run("convert packet reorder", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Reorder:     5,
			ReorderCorr: 10,
		})
		g.Expect(args).To(Equal(""))

		args = convertNetemToArgs(&pb.Netem{
			Time:        1000,
			Jitter:      10000,
			DelayCorr:   25,
			Reorder:     5,
			ReorderCorr: 10,
			Gap:         10,
		})
		g.Expect(args).To(Equal("delay 1000 10000 25 reorder 5 10 gap 10"))

		args = convertNetemToArgs(&pb.Netem{
			Time:        1000,
			Jitter:      10000,
			DelayCorr:   25,
			Reorder:     5,
			ReorderCorr: 10,
			Gap:         10,
		})
		g.Expect(args).To(Equal("delay 1000 10000 25 reorder 5 10 gap 10"))

		args = convertNetemToArgs(&pb.Netem{
			Time:      1000,
			Jitter:    10000,
			DelayCorr: 25,
			Reorder:   5,
			Gap:       10,
		})
		g.Expect(args).To(Equal("delay 1000 10000 25 reorder 5 gap 10"))
	})

	t.Run("convert packet duplication", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Duplicate: 10,
		})
		g.Expect(args).To(Equal("duplicate 10"))

		args = convertNetemToArgs(&pb.Netem{
			Duplicate:     10,
			DuplicateCorr: 50,
		})
		g.Expect(args).To(Equal("duplicate 10 50"))
	})

	t.Run("convert packet corrupt", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Corrupt: 10,
		})
		g.Expect(args).To(Equal("corrupt 10"))

		args = convertNetemToArgs(&pb.Netem{
			Corrupt:     10,
			CorruptCorr: 50,
		})
		g.Expect(args).To(Equal("corrupt 10 50"))
	})

	t.Run("complicate cases", func(t *testing.T) {
		args := convertNetemToArgs(&pb.Netem{
			Time:        1000,
			Jitter:      10000,
			Reorder:     5,
			Gap:         10,
			Corrupt:     10,
			CorruptCorr: 50,
		})
		g.Expect(args).To(Equal("delay 1000 10000 reorder 5 gap 10 corrupt 10 50"))
	})
}
