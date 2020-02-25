// Copyright 2020 PingCAP, Inc.
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
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("hook", func() {
	Context("PreOpen", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("open", &faultContext{
				errno: e,
				pct:   100,
			})
			_, _, err := h.PreOpen("", 0)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("open")
			_, _, err := h.PreOpen("", 0)
			Expect(err).To(BeNil())
		})
	})

	Context("PostOpen", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostOpen(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreRead", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("read", &faultContext{
				errno: e,
				pct:   100,
			})
			_, f, _, err := h.PreRead("", 0, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("read")
			_, f, _, err := h.PreRead("", 0, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostRead", func() {
		It("should work", func() {
			h := &InjuredHook{}
			_, f, e := h.PostRead(0, nil, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreWrite", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("write", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreWrite("", nil, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("write")
			f, _, err := h.PreWrite("", nil, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostWrite", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostWrite(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreMkdir", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("mkdir", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreMkdir("", 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("mkdir")
			f, _, err := h.PreMkdir("", 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostMkdir", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostMkdir(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreRmdir", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("rmdir", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreRmdir("")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("rmdir")
			f, _, err := h.PreRmdir("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostRmdir", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostRmdir(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreOpenDir", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("opendir", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreOpenDir("")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("opendir")
			f, _, err := h.PreOpenDir("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostOpenDir", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostOpenDir(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreFsync", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("fsync", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreFsync("", 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("fsync")
			f, _, err := h.PreFsync("", 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostFsync", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostFsync(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreFlush", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("flush", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreFlush("")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("flush")
			f, _, err := h.PreFlush("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostFlush", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostFlush(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreRelease", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("release", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _ := h.PreRelease("")
			Expect(f).To(Equal(false))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("release")
			f, _, err := h.PreFlush("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostRelease", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f := h.PostRelease(0)
			Expect(f).To(Equal(false))
		})
	})

	Context("PreTruncate", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("truncate", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreTruncate("", 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("truncate")
			f, _, err := h.PreTruncate("", 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostTruncate", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostTruncate(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreGetAttr", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("getattr", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreGetAttr("")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("getattr")
			f, _, err := h.PreGetAttr("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostGetAttr", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostGetAttr(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreChown", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("chown", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreChown("", 0, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("chown")
			f, _, err := h.PreChown("", 0, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostChown", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostChown(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreChmod", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("chmod", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreChmod("", 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("chmod")
			f, _, err := h.PreChmod("", 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostChmod", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostChmod(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreUtimens", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("utimens", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreUtimens("", nil, nil)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("utimens")
			f, _, err := h.PreUtimens("", nil, nil)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostUtimens", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostUtimens(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreAllocate", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("allocate", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreAllocate("", 0, 0, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("allocate")
			f, _, err := h.PreAllocate("", 0, 0, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostAllocate", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostAllocate(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreGetLk", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("getlk", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreGetLk("", 0, nil, 0, nil)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("getlk")
			f, _, err := h.PreGetLk("", 0, nil, 0, nil)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostGetLk", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostGetLk(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreSetLk", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("setlk", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreSetLk("", 0, nil, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("setlk")
			f, _, err := h.PreSetLk("", 0, nil, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostSetLk", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostSetLk(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreSetLkw", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("setlkw", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreSetLkw("", 0, nil, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("setlkw")
			f, _, err := h.PreSetLkw("", 0, nil, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostSetLkw", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostSetLkw(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreStatFs", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("statfs", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreStatFs("")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("statfs")
			f, _, err := h.PreStatFs("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostStatFs", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostStatFs(0)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreReadlink", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("readlink", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreReadlink("")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("readlink")
			f, _, err := h.PreReadlink("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostReadlink", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostReadlink(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreSymlink", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("symlink", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreSymlink("", "")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("symlink")
			f, _, err := h.PreSymlink("", "")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostSymlink", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostSymlink(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreCreate", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("create", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreCreate("", 0, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("create")
			f, _, err := h.PreCreate("", 0, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostCreate", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostCreate(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreAccess", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("access", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreAccess("", 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("access")
			f, _, err := h.PreAccess("", 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostAccess", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostAccess(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreLink", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("link", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreLink("", "")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("link")
			f, _, err := h.PreLink("", "")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostLink", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostLink(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreMknod", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("mknod", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreMknod("", 0, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("mknod")
			f, _, err := h.PreMknod("", 0, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostMknod", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostMknod(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreRename", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("rename", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreRename("", "")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("rename")
			f, _, err := h.PreRename("", "")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostRename", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostRename(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreUnlink", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("unlink", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreUnlink("")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("unlink")
			f, _, err := h.PreUnlink("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostUnlink", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostUnlink(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreGetXAttr", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("getxattr", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreGetXAttr("", "")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("getxattr")
			f, _, err := h.PreGetXAttr("", "")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostGetXAttr", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostGetXAttr(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreListXAttr", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("listxattr", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreListXAttr("")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("listxattr")
			f, _, err := h.PreListXAttr("")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostListXAttr", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostListXAttr(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreRemoveXAttr", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("removexattr", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreRemoveXAttr("", "")
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("removexattr")
			f, _, err := h.PreRemoveXAttr("", "")
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostRemoveXAttr", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostRemoveXAttr(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})

	Context("PreSetXAttr", func() {
		It("should work", func() {
			e := errors.New("mock error")
			h := &InjuredHook{}
			faultMap.Store("setxattr", &faultContext{
				errno: e,
				pct:   100,
			})
			f, _, err := h.PreSetXAttr("", "", nil, 0)
			Expect(f).To(Equal(true))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(err))
		})

		It("should skip", func() {
			h := &InjuredHook{}
			faultMap.Delete("setxattr")
			f, _, err := h.PreSetXAttr("", "", nil, 0)
			Expect(f).To(Equal(false))
			Expect(err).To(BeNil())
		})
	})

	Context("PostSetXAttr", func() {
		It("should work", func() {
			h := &InjuredHook{}
			f, e := h.PostSetXAttr(0, nil)
			Expect(f).To(Equal(false))
			Expect(e).To(BeNil())
		})
	})
})
