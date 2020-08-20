use async_trait::async_trait;
use fuse::*;
use time::Timespec;

use super::errors::Result;
use super::reply::*;
use super::runtime::spawn;

use std::ffi::OsString;
use std::fmt::Debug;
use std::sync::Arc;
use std::{
    future::Future,
    path::{Path, PathBuf},
};

pub fn spawn_reply<F, R, V>(reply: R, f: F)
where
    F: Future<Output = Result<V>> + Send + 'static,
    R: FsReply<V> + Send + 'static,
    V: Debug,
{
    spawn(async move {
        let result = f.await;
        reply.reply(result);
    });
}

#[async_trait]
pub trait AsyncFileSystemImpl: Send + Sync {
    fn init(&self) -> Result<()>;

    fn destroy(&self);

    async fn lookup(&self, parent: u64, name: OsString) -> Result<Entry>;

    async fn forget(&self, ino: u64, nlookup: u64);

    async fn getattr(&self, ino: u64) -> Result<Attr>;

    async fn setattr(
        &self,
        ino: u64,
        mode: Option<u32>,
        uid: Option<u32>,
        gid: Option<u32>,
        size: Option<u64>,
        atime: Option<Timespec>,
        mtime: Option<Timespec>,
        fh: Option<u64>,
        crtime: Option<Timespec>,
        chgtime: Option<Timespec>,
        bkuptime: Option<Timespec>,
        flags: Option<u32>,
    ) -> Result<Attr>;

    async fn readlink(&self, ino: u64) -> Result<Data>;

    async fn mknod(&self, parent: u64, name: OsString, mode: u32, rdev: u32) -> Result<Entry>;

    async fn mkdir(&self, parent: u64, name: OsString, mode: u32) -> Result<Entry>;

    async fn unlink(&self, parent: u64, name: OsString) -> Result<()>;

    async fn rmdir(&self, parent: u64, name: OsString) -> Result<()>;

    async fn symlink(&self, parent: u64, name: OsString, link: PathBuf) -> Result<Entry>;

    async fn rename(
        &self,
        parent: u64,
        name: OsString,
        newparent: u64,
        newname: OsString,
    ) -> Result<()>;

    async fn link(&self, ino: u64, newparent: u64, newname: OsString) -> Result<Entry>;

    async fn open(&self, ino: u64, flags: u32) -> Result<Open>;

    async fn read(&self, ino: u64, fh: u64, offset: i64, size: u32) -> Result<Data>;

    async fn write(
        &self,
        ino: u64,
        fh: u64,
        offset: i64,
        data: Vec<u8>,
        flags: u32,
    ) -> Result<Write>;

    async fn flush(&self, ino: u64, fh: u64, lock_owner: u64) -> Result<()>;

    async fn release(
        &self,
        ino: u64,
        fh: u64,
        flags: u32,
        lock_owner: u64,
        flush: bool,
    ) -> Result<()>;

    async fn fsync(&self, ino: u64, fh: u64, datasync: bool) -> Result<()>;

    async fn opendir(&self, ino: u64, flags: u32) -> Result<Open>;

    async fn readdir(&self, ino: u64, fh: u64, offset: i64, reply: ReplyDirectory);

    async fn releasedir(&self, ino: u64, fh: u64, flags: u32) -> Result<()>;

    async fn fsyncdir(&self, ino: u64, fh: u64, datasync: bool) -> Result<()>;

    async fn statfs(&self, ino: u64) -> Result<StatFs>;

    async fn setxattr(
        &self,
        ino: u64,
        name: OsString,
        value: Vec<u8>,
        flags: u32,
        position: u32,
    ) -> Result<()>;

    async fn getxattr(&self, ino: u64, name: OsString, size: u32) -> Result<Xattr>;

    async fn listxattr(&self, ino: u64, size: u32) -> Result<Xattr>;

    async fn removexattr(&self, ino: u64, name: OsString) -> Result<()>;

    async fn access(&self, ino: u64, mask: u32) -> Result<()>;

    async fn create(
        &self,
        parent: u64,
        name: OsString,
        mode: u32,
        flags: u32,
        uid: u32,
        gid: u32,
    ) -> Result<Create>;

    async fn getlk(
        &self,
        ino: u64,
        fh: u64,
        lock_owner: u64,
        start: u64,
        end: u64,
        typ: u32,
        pid: u32,
    ) -> Result<Lock>;

    async fn setlk(
        &self,
        ino: u64,
        fh: u64,
        lock_owner: u64,
        start: u64,
        end: u64,
        typ: u32,
        pid: u32,
        sleep: bool,
    ) -> Result<()>;

    async fn bmap(&self, ino: u64, blocksize: u32, idx: u64, reply: ReplyBmap);
}

pub struct AsyncFileSystem<T>(Arc<T>);

impl<T> AsyncFileSystem<T> {
    pub fn clone_inner(&self) -> Arc<T> {
        self.0.clone()
    }
}

impl<T: AsyncFileSystemImpl> From<T> for AsyncFileSystem<T> {
    fn from(inner: T) -> Self {
        Self(Arc::new(inner))
    }
}

impl<T: Debug> Debug for AsyncFileSystem<T> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        self.0.fmt(f)
    }
}

impl<T: AsyncFileSystemImpl + 'static> Filesystem for AsyncFileSystem<T> {
    fn init(&mut self, _req: &fuse::Request) -> std::result::Result<(), nix::libc::c_int> {
        self.0.init().map_err(|err| err.into())
    }

    fn destroy(&mut self, _req: &fuse::Request) {
        self.0.destroy()
    }

    fn lookup(&mut self, _req: &Request, parent: u64, name: &std::ffi::OsStr, reply: ReplyEntry) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        spawn_reply(reply, async move { async_impl.lookup(parent, name).await });
    }

    fn forget(&mut self, _req: &Request, ino: u64, nlookup: u64) {
        let async_impl = self.0.clone();

        // TODO: union the spawn function for request without reply
        spawn(async move {
            async_impl.forget(ino, nlookup).await;
        });
    }

    fn getattr(&mut self, _req: &Request, ino: u64, reply: ReplyAttr) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move { async_impl.getattr(ino).await });
    }

    fn setattr(
        &mut self,
        _req: &Request,
        ino: u64,
        mode: Option<u32>,
        uid: Option<u32>,
        gid: Option<u32>,
        size: Option<u64>,
        atime: Option<Timespec>,
        mtime: Option<Timespec>,
        fh: Option<u64>,
        crtime: Option<Timespec>,
        chgtime: Option<Timespec>,
        bkuptime: Option<Timespec>,
        flags: Option<u32>,
        reply: ReplyAttr,
    ) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move {
            async_impl
                .setattr(
                    ino, mode, uid, gid, size, atime, mtime, fh, crtime, chgtime, bkuptime, flags,
                )
                .await
        });
    }

    fn readlink(&mut self, _req: &Request, ino: u64, reply: ReplyData) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move { async_impl.readlink(ino).await });
    }
    fn mknod(
        &mut self,
        _req: &Request,
        parent: u64,
        name: &std::ffi::OsStr,
        mode: u32,
        rdev: u32,
        reply: ReplyEntry,
    ) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        spawn_reply(reply, async move {
            async_impl.mknod(parent, name, mode, rdev).await
        });
    }
    fn mkdir(
        &mut self,
        _req: &Request,
        parent: u64,
        name: &std::ffi::OsStr,
        mode: u32,
        reply: ReplyEntry,
    ) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        spawn_reply(
            reply,
            async move { async_impl.mkdir(parent, name, mode).await },
        );
    }
    fn unlink(&mut self, _req: &Request, parent: u64, name: &std::ffi::OsStr, reply: ReplyEmpty) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        spawn_reply(reply, async move { async_impl.unlink(parent, name).await });
    }
    fn rmdir(&mut self, _req: &Request, parent: u64, name: &std::ffi::OsStr, reply: ReplyEmpty) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        spawn_reply(reply, async move { async_impl.rmdir(parent, name).await });
    }
    fn symlink(
        &mut self,
        _req: &Request,
        parent: u64,
        name: &std::ffi::OsStr,
        link: &Path,
        reply: ReplyEntry,
    ) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        let link = link.to_owned();
        spawn_reply(reply, async move {
            async_impl.symlink(parent, name, link).await
        });
    }
    fn rename(
        &mut self,
        _req: &Request,
        parent: u64,
        name: &std::ffi::OsStr,
        newparent: u64,
        newname: &std::ffi::OsStr,
        reply: ReplyEmpty,
    ) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        let newname = newname.to_owned();
        spawn_reply(reply, async move {
            async_impl.rename(parent, name, newparent, newname).await
        });
    }
    fn link(
        &mut self,
        _req: &Request,
        ino: u64,
        newparent: u64,
        newname: &std::ffi::OsStr,
        reply: ReplyEntry,
    ) {
        let async_impl = self.0.clone();
        let newname = newname.to_owned();
        spawn_reply(reply, async move {
            async_impl.link(ino, newparent, newname).await
        });
    }
    fn open(&mut self, _req: &Request, ino: u64, flags: u32, reply: ReplyOpen) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move { async_impl.open(ino, flags).await });
    }
    fn read(
        &mut self,
        _req: &Request,
        ino: u64,
        fh: u64,
        offset: i64,
        size: u32,
        reply: ReplyData,
    ) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move {
            async_impl.read(ino, fh, offset, size).await
        });
    }
    fn write(
        &mut self,
        _req: &Request,
        ino: u64,
        fh: u64,
        offset: i64,
        data: &[u8],
        flags: u32,
        reply: ReplyWrite,
    ) {
        let async_impl = self.0.clone();
        let data = data.to_owned();
        spawn_reply(reply, async move {
            async_impl.write(ino, fh, offset, data, flags).await
        });
    }
    fn flush(&mut self, _req: &Request, ino: u64, fh: u64, lock_owner: u64, reply: ReplyEmpty) {
        let async_impl = self.0.clone();
        spawn_reply(
            reply,
            async move { async_impl.flush(ino, fh, lock_owner).await },
        );
    }
    fn release(
        &mut self,
        _req: &Request,
        ino: u64,
        fh: u64,
        flags: u32,
        lock_owner: u64,
        flush: bool,
        reply: ReplyEmpty,
    ) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move {
            async_impl.release(ino, fh, flags, lock_owner, flush).await
        });
    }
    fn fsync(&mut self, _req: &Request, ino: u64, fh: u64, datasync: bool, reply: ReplyEmpty) {
        let async_impl = self.0.clone();
        spawn_reply(
            reply,
            async move { async_impl.fsync(ino, fh, datasync).await },
        );
    }
    fn opendir(&mut self, _req: &Request, ino: u64, flags: u32, reply: ReplyOpen) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move { async_impl.opendir(ino, flags).await });
    }
    fn readdir(&mut self, _req: &Request, ino: u64, fh: u64, offset: i64, reply: ReplyDirectory) {
        let async_impl = self.0.clone();
        spawn(async move {
            async_impl.readdir(ino, fh, offset, reply).await;
        });
    }
    fn releasedir(&mut self, _req: &Request, ino: u64, fh: u64, flags: u32, reply: ReplyEmpty) {
        let async_impl = self.0.clone();
        spawn_reply(
            reply,
            async move { async_impl.releasedir(ino, fh, flags).await },
        );
    }
    fn fsyncdir(&mut self, _req: &Request, ino: u64, fh: u64, datasync: bool, reply: ReplyEmpty) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move {
            async_impl.fsyncdir(ino, fh, datasync).await
        });
    }
    fn statfs(&mut self, _req: &Request, ino: u64, reply: ReplyStatfs) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move { async_impl.statfs(ino).await });
    }
    fn setxattr(
        &mut self,
        _req: &Request,
        ino: u64,
        name: &std::ffi::OsStr,
        value: &[u8],
        flags: u32,
        position: u32,
        reply: ReplyEmpty,
    ) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        let value = value.to_owned();
        spawn_reply(reply, async move {
            async_impl.setxattr(ino, name, value, flags, position).await
        });
    }
    fn getxattr(
        &mut self,
        _req: &Request,
        ino: u64,
        name: &std::ffi::OsStr,
        size: u32,
        reply: ReplyXattr,
    ) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        spawn_reply(
            reply,
            async move { async_impl.getxattr(ino, name, size).await },
        );
    }
    fn listxattr(&mut self, _req: &Request, ino: u64, size: u32, reply: ReplyXattr) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move { async_impl.listxattr(ino, size).await });
    }
    fn removexattr(&mut self, _req: &Request, ino: u64, name: &std::ffi::OsStr, reply: ReplyEmpty) {
        let async_impl = self.0.clone();
        let name = name.to_owned();
        spawn_reply(
            reply,
            async move { async_impl.removexattr(ino, name).await },
        );
    }
    fn access(&mut self, _req: &Request, ino: u64, mask: u32, reply: ReplyEmpty) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move { async_impl.access(ino, mask).await });
    }
    fn create(
        &mut self,
        req: &Request,
        parent: u64,
        name: &std::ffi::OsStr,
        mode: u32,
        flags: u32,
        reply: ReplyCreate,
    ) {
        let uid = req.uid();
        let gid = req.gid();

        let async_impl = self.0.clone();
        let name = name.to_owned();
        spawn_reply(reply, async move {
            async_impl.create(parent, name, mode, flags, uid, gid).await
        });
    }
    fn getlk(
        &mut self,
        _req: &Request,
        ino: u64,
        fh: u64,
        lock_owner: u64,
        start: u64,
        end: u64,
        typ: u32,
        pid: u32,
        reply: ReplyLock,
    ) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move {
            async_impl
                .getlk(ino, fh, lock_owner, start, end, typ, pid)
                .await
        });
    }
    fn setlk(
        &mut self,
        _req: &Request,
        ino: u64,
        fh: u64,
        lock_owner: u64,
        start: u64,
        end: u64,
        typ: u32,
        pid: u32,
        sleep: bool,
        reply: ReplyEmpty,
    ) {
        let async_impl = self.0.clone();
        spawn_reply(reply, async move {
            async_impl
                .setlk(ino, fh, lock_owner, start, end, typ, pid, sleep)
                .await
        });
    }
    fn bmap(&mut self, _req: &Request, ino: u64, blocksize: u32, idx: u64, reply: ReplyBmap) {
        let async_impl = self.0.clone();
        spawn(async move {
            async_impl.bmap(ino, blocksize, idx, reply).await;
        });
    }
}
