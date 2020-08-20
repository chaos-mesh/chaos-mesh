use crate::ptrace;

use std::fmt::Debug;
use std::fs::read_dir;
use std::fs::read_link;
use std::path::{Path, PathBuf};

use anyhow::{anyhow, Result};
use nix::fcntl::FcntlArg;
use nix::fcntl::OFlag;
use nix::sys::stat::Mode;

use tracing::info;

#[derive(PartialEq, Debug)]
pub enum MountDirection {
    EnableChaos,
    DisableChaos,
}

pub struct FdReplacer {
    pid: i32,
    original_path: PathBuf,
    new_path: PathBuf,
    direction: MountDirection,
    process: ptrace::TracedProcess,
}

impl FdReplacer {
    #[tracing::instrument()]
    pub fn new<P: AsRef<Path> + Debug>(
        path: P,
        pid: i32,
        direction: MountDirection,
    ) -> Result<FdReplacer> {
        let original_path: PathBuf = path.as_ref().to_owned();

        let mut base_path: PathBuf = path.as_ref().to_owned();
        if !base_path.pop() {
            return Err(anyhow!("path is the root"));
        }

        let mut new_path: PathBuf = base_path.clone();
        let original_filename = original_path
            .file_name()
            .ok_or(anyhow!("the path terminates in `..` or `/`"))?
            .to_str()
            .ok_or(anyhow!("path with non-UTF-8 character"))?;
        let new_filename = format!("__chaosfs__{}__", original_filename);
        new_path.push(new_filename.as_str());

        return Ok(FdReplacer {
            pid,
            original_path,
            new_path,
            direction,
            process: ptrace::TracedProcess::trace(pid)?,
        });
    }

    #[tracing::instrument()]
    pub fn enable<P: AsRef<Path> + Debug>(path: P, pid: i32) -> Result<FdReplacer> {
        Self::new(path, pid, MountDirection::EnableChaos)
    }

    #[tracing::instrument()]
    pub fn disable<P: AsRef<Path> + Debug>(path: P, pid: i32) -> Result<FdReplacer> {
        Self::new(path, pid, MountDirection::DisableChaos)
    }

    #[tracing::instrument(skip(self))]
    pub fn reopen(&self) -> Result<()> {
        info!("reopen fd for pid: {}", self.pid);

        let base_path = if self.direction == MountDirection::EnableChaos {
            self.new_path.as_path()
        } else {
            self.original_path.as_path()
        };

        for thread in self.process.threads() {
            let tid = thread.tid;
            let fd_dir_path = format!("/proc/{}/fd", tid);
            for fd in read_dir(fd_dir_path)?.into_iter() {
                let path = fd?.path();
                let fd = path
                    .file_name()
                    .ok_or(anyhow!("fd doesn't contain a filename"))?
                    .to_str()
                    .ok_or(anyhow!("fd contains non-UTF-8 character"))?
                    .parse()?;
                if let Ok(path) = read_link(&path) {
                    info!("handling path: {:?}", path);
                    if path.starts_with(base_path) {
                        info!("reopen file, fd: {:?}, path: {:?}", fd, path.as_path());
                        self.reopen_file(&thread, fd, path.as_path())?;
                    }
                }
            }
        }

        return Ok(());
    }

    #[tracing::instrument(skip(self, thread, path))]
    fn reopen_file<P: AsRef<Path>>(
        &self,
        thread: &ptrace::TracedThread,
        fd: u64,
        path: P,
    ) -> Result<()> {
        let base_path = if self.direction == MountDirection::EnableChaos {
            self.new_path.as_path()
        } else {
            self.original_path.as_path()
        };

        let striped_path = path.as_ref().strip_prefix(base_path)?;

        let original_path = if self.direction == MountDirection::EnableChaos {
            self.original_path.join(striped_path)
        } else {
            self.new_path.join(striped_path)
        };

        info!(
            "reopen fd: {} for pid {}, from {} to {}",
            fd,
            thread.tid,
            path.as_ref().display(),
            original_path.display()
        );

        let flags = thread.fcntl(fd, FcntlArg::F_GETFL)?;

        let flags = OFlag::from_bits_truncate(flags as i32);

        info!("fcntl get flags {:?}", flags);

        let new_open_fd = thread.open(original_path, flags, Mode::empty())?;
        thread.dup2(new_open_fd, fd)?;
        thread.close(new_open_fd)?;

        return Ok(());
    }
}

impl Drop for FdReplacer {
    #[tracing::instrument(skip(self))]
    fn drop(&mut self) {
        for thread in self.process.threads() {
            thread.detach().unwrap_or_else(|err| {
                panic!(
                    "fails to detach thread ({}/{}): {}",
                    self.pid, thread.tid, err
                )
            });
        }
    }
}
