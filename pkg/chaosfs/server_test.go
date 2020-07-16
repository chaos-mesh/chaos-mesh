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

package chaosfs

import (
	"context"
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosfs/pb"
)

var _ = Describe("server", func() {
	const faultInjectMethod = "open"
	const faultInjectPath = "fault-inject-path"

	Context("randomErrno", func() {
		It("pseudo random", func() {
			errSets := make(map[string]bool)
			for i := 0; i < 5; i++ {
				e := randomErrno()
				errSets[e.Error()] = true
			}
			Expect(len(errSets) > 1).Should(Equal(true))
		})
	})

	Context("probab", func() {
		It("should be true", func() {
			Expect(probab(100)).To(Equal(true))
			Expect(probab(101)).To(Equal(true))
			Expect(probab(100000000)).To(Equal(true))
		})

		It("should be false", func() {
			Expect(probab(0)).To(Equal(false))
		})
	})

	Context("faultInject", func() {
		It("should work", func() {
			faultMap.Store(faultInjectMethod, &faultContext{
				pct:    100,
				random: true,
			})
			err := faultInject(faultInjectPath, faultInjectMethod)
			Expect(err).ToNot(BeNil())
		})

		It("should skip unknow method", func() {
			err := faultInject(faultInjectPath, "unknow method")
			Expect(err).To(BeNil())
		})

		It("should skip on wrong regex", func() {
			faultMap.Store(faultInjectMethod, &faultContext{
				pct:    100,
				random: true,
				path:   `^\/(?!\/)(.*?)`,
			})
			err := faultInject(faultInjectPath, faultInjectMethod)
			Expect(err).To(BeNil())
		})

		It("should skip on mismatch path", func() {
			faultMap.Store(faultInjectMethod, &faultContext{
				pct:    100,
				random: true,
				path:   `mismatch-path`,
			})
			err := faultInject(faultInjectPath, faultInjectMethod)
			Expect(err).To(BeNil())
		})

		It("should return specified errno", func() {
			e := errors.New("mock err")
			faultMap.Store(faultInjectMethod, &faultContext{
				pct:   100,
				errno: e,
			})
			err := faultInject(faultInjectPath, faultInjectMethod)
			Expect(err).To(Equal(e))
		})
	})

	Context("RecoverAll", func() {
		It("should work", func() {
			s := &server{}
			count := 0
			s.RecoverAll(context.TODO(), nil)
			faultMap.Range(func(k, v interface{}) bool {
				count++
				return true
			})
			Expect(count).To(Equal(0))
			faultMap.Store(faultInjectMethod, &faultContext{})
			faultMap.Range(func(k, v interface{}) bool {
				count++
				return true
			})
			Expect(count).To(Equal(1))
			count = 0
			s.RecoverAll(context.TODO(), nil)
			faultMap.Range(func(k, v interface{}) bool {
				count++
				return true
			})
			Expect(count).To(Equal(0))
		})
	})

	Context("setFault and RecoverMethod", func() {
		It("should work", func() {
			s := &server{}
			s.setFault([]string{faultInjectMethod}, &faultContext{})
			_, ok := faultMap.Load(faultInjectMethod)
			Expect(ok).To(Equal(true))
			s.RecoverMethod(context.TODO(), &pb.Request{
				Methods: []string{faultInjectMethod},
			})
			_, ok = faultMap.Load(faultInjectMethod)
			Expect(ok).To(Equal(false))
		})
	})

	Context("SetFault", func() {
		It("should work", func() {
			s := &server{}
			faultMap.Delete(faultInjectMethod)
			_, ok := faultMap.Load(faultInjectMethod)
			Expect(ok).To(Equal(false))
			s.SetFault(context.TODO(), &pb.Request{
				Methods: []string{faultInjectMethod},
				Random:  true,
				Pct:     100,
			})
			_, ok = faultMap.Load(faultInjectMethod)
			Expect(ok).To(Equal(true))
		})
	})

	Context("SetFaultAll", func() {
		It("should work", func() {
			s := &server{}
			faultMap.Delete(faultInjectMethod)
			_, ok := faultMap.Load(faultInjectMethod)
			Expect(ok).To(Equal(false))
			s.SetFaultAll(context.TODO(), &pb.Request{
				Random: true,
				Pct:    100,
			})
			s.Injected(context.TODO(), &empty.Empty{})
			_, ok = faultMap.Load(faultInjectMethod)
			Expect(ok).To(Equal(true))
		})
	})

	Context("Injected", func() {
		It("should return false", func() {
			s := &server{}
			faultMap.Range(func(k, v interface{}) bool {
				faultMap.Delete(k)
				return true
			})

			resp, _ := s.Injected(context.TODO(), &empty.Empty{})
			Expect(resp.Injected).To(Equal(false))
		})

		It("should return true", func() {
			s := &server{}
			faultMap.Range(func(k, v interface{}) bool {
				faultMap.Delete(k)
				return true
			})

			faultMap.Store(faultInjectMethod, &pb.Request{Delay: 1000})
			resp, _ := s.Injected(context.TODO(), &empty.Empty{})
			Expect(resp.Injected).To(Equal(true))
		})
	})
})
