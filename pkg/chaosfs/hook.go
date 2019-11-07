package chaosfs

import (
	"time"

	"github.com/ethercflow/hookfs/hookfs"
	"github.com/hanwen/go-fuse/fuse"
)

type InjuredHookContext struct {
}

type InjuredHook struct {
	Addr string
}

func (h *InjuredHook) Init() error {
	StartServer(h.Addr)
	return nil
}

func (h *InjuredHook) PreOpen(path string, flags uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "open")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostOpen(int32, hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreRead(path string, length int64, offset int64) ([]byte, bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "read")
	if err != nil {
		return nil, true, ctx, err
	}
	return nil, false, ctx, nil
}

func (h *InjuredHook) PostRead(realRetCode int32, realBuf []byte, prehookCtx hookfs.HookContext) ([]byte, bool, error) {
	return nil, false, nil
}

func (h *InjuredHook) PreWrite(path string, buf []byte, offset int64) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "write")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostWrite(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreMkdir(path string, mode uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "mkdir")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostMkdir(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreRmdir(path string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "rmdir")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostRmdir(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreOpenDir(path string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "opendir")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostOpenDir(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreFsync(path string, flags uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "fsync")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostFsync(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreFlush(path string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "flush")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostFlush(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreRelease(path string) (bool, hookfs.HookContext) {
	ctx := &InjuredHookContext{}
	_ = faultInject(path, "release")
	return false, ctx
}

func (h *InjuredHook) PostRelease(prehookCtx hookfs.HookContext) (hooked bool) {
	return false
}

func (h *InjuredHook) PreTruncate(path string, size uint64) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "truncate")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostTruncate(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreGetAttr(path string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "getattr")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostGetAttr(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreChown(path string, uid uint32, gid uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "chown")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostChown(realRetCode int32, prehookCtx hookfs.HookContext) (hooked bool, err error) {
	return false, nil
}

func (h *InjuredHook) PreChmod(path string, perms uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "chmod")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostChmod(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreUtimens(path string, atime *time.Time, mtime *time.Time) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "utimens")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostUtimens(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreAllocate(path string, off uint64, size uint64, mode uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "allocate")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostAllocate(realRetCode int32, prehookCtx hookfs.HookContext) (hooked bool, err error) {
	return false, nil
}

func (h *InjuredHook) PreGetLk(path string, owner uint64, lk *fuse.FileLock, flags uint32, out *fuse.FileLock) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "getlk")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostGetLk(realRetCode int32, prehookCtx hookfs.HookContext) (hooked bool, err error) {
	return false, nil
}

func (h *InjuredHook) PreSetLk(path string, owner uint64, lk *fuse.FileLock, flags uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "setlk")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostSetLk(realRetCode int32, prehookCtx hookfs.HookContext) (hooked bool, err error) {
	return false, nil
}

func (h *InjuredHook) PreSetLkw(path string, owner uint64, lk *fuse.FileLock, flags uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "setlkw")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostSetLkw(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreStatFs(path string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(path, "statfs")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostStatFs(prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreReadlink(name string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "readlink")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostReadlink(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreSymlink(value string, linkName string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(value, "symlink")
	if err != nil {
		return true, ctx, err
	}
	err = faultInject(linkName, "symlink")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostSymlink(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreCreate(name string, flags uint32, mode uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "create")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostCreate(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreAccess(name string, mode uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "access")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostAccess(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreLink(oldName string, newName string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(oldName, "link")
	if err != nil {
		return true, ctx, err
	}
	err = faultInject(newName, "link")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostLink(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreMknod(name string, mode uint32, dev uint32) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "mknod")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostMknod(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreRename(oldName string, newName string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(oldName, "rename")
	if err != nil {
		return true, ctx, err
	}
	err = faultInject(newName, "rename")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostRename(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreUnlink(name string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "unlink")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil

}
func (h *InjuredHook) PostUnlink(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreGetXAttr(name string, attribute string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "getxattr")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostGetXAttr(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreListXAttr(name string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "listxattr")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostListXAttr(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreRemoveXAttr(name string, attr string) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "removexattr")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostRemoveXAttr(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}

func (h *InjuredHook) PreSetXAttr(name string, attr string, data []byte, flags int) (bool, hookfs.HookContext, error) {
	ctx := &InjuredHookContext{}
	err := faultInject(name, "setxattr")
	if err != nil {
		return true, ctx, err
	}
	return false, ctx, nil
}

func (h *InjuredHook) PostSetXAttr(realRetCode int32, prehookCtx hookfs.HookContext) (bool, error) {
	return false, nil
}
