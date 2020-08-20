use super::filter;
use super::Injector;

use super::injector_config::{AttrOverrideConfig, FileType as ConfigFileType, FilterConfig};
use crate::hookfs::{Reply, Result};

use async_trait::async_trait;
use fuse::{FileAttr, FileType};
use time::Timespec;
use tracing::{info, trace};

use std::path::Path;

#[derive(Debug)]
pub struct AttrOverrideInjector {
    filter: filter::Filter,

    ino: Option<u64>,
    size: Option<u64>,
    blocks: Option<u64>,
    atime: Option<Timespec>,
    mtime: Option<Timespec>,
    ctime: Option<Timespec>,
    kind: Option<FileType>,
    perm: Option<u16>,
    nlink: Option<u32>,
    uid: Option<u32>,
    gid: Option<u32>,
    rdev: Option<u32>,
}

impl AttrOverrideInjector {
    fn inject_attr(&self, attr: &mut FileAttr) {
        if let Some(ino) = self.ino {
            trace!("overriding ino");
            attr.ino = ino
        }
        if let Some(size) = self.size {
            trace!("overriding size");
            attr.size = size
        }
        if let Some(blocks) = self.blocks {
            trace!("overriding block");
            attr.blocks = blocks
        }
        if let Some(atime) = self.atime {
            trace!("overriding atime");
            attr.atime = atime
        }
        if let Some(mtime) = self.mtime {
            trace!("overriding mtime");
            attr.mtime = mtime
        }
        if let Some(ctime) = self.ctime {
            trace!("overriding ctime");
            attr.ctime = ctime
        }
        if let Some(kind) = self.kind {
            trace!("overriding kind");
            attr.kind = kind
        }
        if let Some(perm) = self.perm {
            trace!("overriding perm");
            attr.perm = perm
        }
        if let Some(nlink) = self.nlink {
            trace!("overriding nlink");
            attr.nlink = nlink
        }
        if let Some(uid) = self.uid {
            trace!("overriding uid");
            attr.uid = uid
        }
        if let Some(gid) = self.gid {
            trace!("overriding gid");
            attr.gid = gid
        }
        if let Some(rdev) = self.rdev {
            trace!("overriding rdev");
            attr.rdev = rdev
        }
    }
}

#[async_trait]
impl Injector for AttrOverrideInjector {
    #[tracing::instrument]
    async fn inject(&self, _: &filter::Method, _: &Path) -> Result<()> {
        Ok(())
    }
    fn inject_reply(&self, method: &filter::Method, path: &Path, reply: &mut Reply) -> Result<()> {
        if !self.filter.filter(method, path) {
            return Ok(());
        }

        info!("overriding attributes");
        match reply {
            Reply::Entry(entry) => {
                self.inject_attr(&mut entry.stat);
            }
            Reply::Attr(attr) => {
                self.inject_attr(&mut attr.attr);
            }
            _ => {
                info!("reply without attributes");
            },
        }
        Ok(())
    }
}

impl AttrOverrideInjector {
    #[tracing::instrument]
    pub fn build(conf: AttrOverrideConfig) -> anyhow::Result<Self> {
        info!("build attr override injector");

        let methods = vec![
            String::from("getattr"),
            String::from("lookup"),
            String::from("mknod"),
            String::from("mkdir"),
            String::from("symlink"),
            String::from("link"),
        ];
        let filter = filter::Filter::build(FilterConfig {
            path: conf.path,
            methods: Some(methods),
            percent: conf.percent,
        })?;

        let atime = conf.atime.map(|item| Timespec {
            sec: item.sec,
            nsec: item.nsec,
        });
        let mtime = conf.mtime.map(|item| Timespec {
            sec: item.sec,
            nsec: item.nsec,
        });
        let ctime = conf.ctime.map(|item| Timespec {
            sec: item.sec,
            nsec: item.nsec,
        });

        let kind = conf.kind.map(|item| match item {
            ConfigFileType::Directory => FileType::Directory,
            ConfigFileType::NamedPipe => FileType::NamedPipe,
            ConfigFileType::RegularFile => FileType::RegularFile,
            ConfigFileType::Socket => FileType::Socket,
            ConfigFileType::Symlink => FileType::Symlink,
            ConfigFileType::CharDevice => FileType::CharDevice,
            ConfigFileType::BlockDevice => FileType::BlockDevice,
        });

        Ok(Self {
            filter,

            ino: conf.ino,
            size: conf.size,
            blocks: conf.blocks,
            atime,
            mtime,
            ctime,
            kind,
            perm: conf.perm,
            nlink: conf.nlink,
            uid: conf.uid,
            gid: conf.gid,
            rdev: conf.rdev,
        })
    }
}
