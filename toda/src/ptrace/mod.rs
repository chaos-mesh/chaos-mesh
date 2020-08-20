use anyhow::{anyhow, Result};
use nix::fcntl::FcntlArg;
use nix::fcntl::FcntlArg::*;
use nix::fcntl::OFlag;
use nix::sys::mman::{MapFlags, ProtFlags};
use nix::sys::ptrace;
use nix::sys::stat::Mode;
use nix::sys::uio::{process_vm_writev, IoVec, RemoteIoVec};
use nix::sys::wait;
use nix::unistd::Pid;

use tracing::trace;

use std::fs::read_dir;
use std::os::unix::ffi::OsStrExt;
use std::path::Path;

pub struct TracedProcess {
    tids: Vec<i32>,
}

impl TracedProcess {
    pub fn trace(pid: i32) -> Result<TracedProcess> {
        let mut tids = Vec::new();

        let tid_strs = read_dir(format!("/proc/{}/task", pid))?.map(|entry| -> Result<String> {
            Ok(entry?
                .path()
                .file_name()
                .ok_or(anyhow!("unexpected filename in task"))?
                .to_str()
                .ok_or(anyhow!("unexpected non-UTF-8 path in task"))?
                .to_owned())
        });

        for tid_str in tid_strs {
            let tid: i32 = tid_str?.parse()?;
            let pid = Pid::from_raw(tid);

            // TODO: retry here
            ptrace::attach(pid)?;

            let _ = wait::waitpid(pid, None)?;
            // TODO: check wait result

            tids.push(tid);
        }

        return Ok(TracedProcess { tids });
    }

    pub fn threads(&self) -> Vec<TracedThread> {
        self.tids
            .iter()
            .map(|tid| TracedThread { tid: *tid })
            .collect()
    }
}

#[derive(Debug)]
pub struct TracedThread {
    pub tid: i32,
}

impl TracedThread {
    #[tracing::instrument]
    fn protect(&self) -> Result<ThreadGuard> {
        let regs = ptrace::getregs(Pid::from_raw(self.tid))?;

        let rip = regs.rip;
        let rip_ins = ptrace::read(Pid::from_raw(self.tid), rip as *mut libc::c_void)?;

        let guard = ThreadGuard {
            tid: self.tid,
            regs,
            rip_ins,
        };
        return Ok(guard);
    }

    #[tracing::instrument(skip(f))]
    fn with_protect<R, F: Fn(&Self) -> Result<R>>(&self, f: F) -> Result<R> {
        let guard = self.protect()?;

        let ret = f(self)?;

        drop(guard);

        return Ok(ret);
    }

    #[tracing::instrument]
    fn syscall(&self, id: u64, args: &[u64]) -> Result<u64> {
        trace!("run syscall {} {:?}", id, args);

        return self.with_protect(|thread| -> Result<u64> {
            let pid = Pid::from_raw(thread.tid);

            let mut regs = ptrace::getregs(pid)?;
            let cur_ins_ptr = regs.rip;

            regs.rax = id;
            for (index, arg) in args.iter().enumerate() {
                // All these registers are hard coded for x86 platform
                if index == 0 {
                    regs.rdi = arg.clone()
                } else if index == 1 {
                    regs.rsi = arg.clone()
                } else if index == 2 {
                    regs.rdx = arg.clone()
                } else if index == 3 {
                    regs.r10 = arg.clone()
                } else if index == 4 {
                    regs.r8 = arg.clone()
                } else if index == 5 {
                    regs.r9 = arg.clone()
                } else {
                    return Err(anyhow!("too many arguments for a syscall"));
                }
            }
            ptrace::setregs(pid, regs)?;

            // We only support x86-64 platform now, so using hard coded `LittleEndian` here is ok.
            unsafe {
                ptrace::write(
                    pid,
                    cur_ins_ptr as *mut libc::c_void,
                    0x050f as *mut libc::c_void,
                )?
            };
            ptrace::step(pid, None)?;

            let status = wait::waitpid(pid, None)?;
            trace!("wait status: {:?}", status);
            // TODO: check wait result

            let regs = ptrace::getregs(pid)?;

            return Ok(regs.rax);
        });
    }

    #[tracing::instrument]
    pub fn detach(&self) -> Result<()> {
        ptrace::detach(Pid::from_raw(self.tid), None)?;

        return Ok(());
    }

    #[tracing::instrument]
    pub fn dup2(&self, old_fd: u64, new_fd: u64) -> Result<u64> {
        return self.syscall(33, &[old_fd, new_fd]);
    }

    #[tracing::instrument]
    pub fn close(&self, fd: u64) -> Result<u64> {
        return self.syscall(3, &[fd]);
    }

    #[tracing::instrument]
    pub fn fcntl(&self, fd: u64, arg: FcntlArg) -> Result<u64> {
        let (cmd, args) = match arg {
            F_GETFD => (libc::F_GETFD, 0),
            F_GETFL => (libc::F_GETFL, 0),
            _ => unimplemented!(),
        };
        return self.syscall(72, &[fd, cmd as u64, args as u64]);
    }

    #[tracing::instrument]
    pub fn mmap(&self, length: u64, fd: u64) -> Result<u64> {
        let prot = ProtFlags::PROT_READ | ProtFlags::PROT_WRITE | ProtFlags::PROT_EXEC;
        let flags = MapFlags::MAP_PRIVATE | MapFlags::MAP_ANON;

        return self.syscall(
            9,
            &[0, length, prot.bits() as u64, flags.bits() as u64, fd, 0],
        );
    }

    #[tracing::instrument]
    pub fn munmap(&self, addr: u64, len: u64) -> Result<u64> {
        return self.syscall(11, &[addr, len]);
    }

    #[tracing::instrument(skip(f))]
    fn with_mmap<R, F: Fn(&Self, u64) -> Result<R>>(&self, len: u64, f: F) -> Result<R> {
        let addr = self.mmap(len, 0)?;

        let ret = f(self, addr)?;

        self.munmap(addr, len)?;

        return Ok(ret);
    }

    #[tracing::instrument(skip(path))]
    pub fn open<P: AsRef<Path>>(&self, path: P, flags: OFlag, mode: Mode) -> Result<u64> {
        // TODO: 4096 is hard coded size. Replace it.
        return self.with_mmap(4096, |thread, addr| -> Result<u64> {
            let pid = Pid::from_raw(thread.tid);
            let path: &[u8] = path.as_ref().as_os_str().as_bytes();

            process_vm_writev(
                pid,
                &[IoVec::from_slice(path.as_ref())],
                &[RemoteIoVec {
                    base: addr as usize,
                    len: path.len(),
                }],
            )?;

            let ret = self.syscall(2, &[addr, flags.bits() as u64, mode.bits() as u64])?;

            return Ok(ret);
        });
    }
}

#[derive(Debug)]
struct ThreadGuard {
    tid: i32,
    regs: libc::user_regs_struct,
    rip_ins: i64,
}

impl Drop for ThreadGuard {
    fn drop(&mut self) {
        let pid = Pid::from_raw(self.tid);
        unsafe {
            ptrace::write(
                pid,
                self.regs.rip as *mut libc::c_void,
                self.rip_ins as *mut libc::c_void,
            )
            .unwrap();
        }
        ptrace::setregs(pid, self.regs).unwrap();
    }
}
