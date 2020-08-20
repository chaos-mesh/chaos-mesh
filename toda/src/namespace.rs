use anyhow::Result;

use nix::fcntl::{open, OFlag};
use nix::sched::setns;
use nix::sched::CloneFlags;
use nix::sys::stat;

use std::sync::atomic::{AtomicPtr, Ordering};

pub fn enter_mnt_namespace(pid: i32) -> Result<()> {
    let mnt_ns_path = format!("/proc/{}/ns/mnt", pid);
    let mnt_ns = open(mnt_ns_path.as_str(), OFlag::O_RDONLY, stat::Mode::all())?;

    setns(mnt_ns, CloneFlags::CLONE_NEWNS)?;

    Ok(())
}

pub fn with_mnt_pid_namespace<F: FnOnce() -> Result<R>, R>(f: Box<F>, pid: i32) -> Result<R> {
    // FIXME: memory leak here
    let mut ret = None;
    let ret_ptr: AtomicPtr<Option<Result<R>>> = AtomicPtr::new(&mut ret);

    let args = Box::new((f, &ret_ptr, Box::new(pid)));

    extern "C" fn callback<F: FnOnce() -> Result<R>, R>(args: *mut libc::c_void) -> libc::c_int {
        let args =
            unsafe { Box::from_raw(args as *mut (Box<F>, &AtomicPtr<Option<Result<R>>>, Box<i32>)) };
        let (f, ret_ptr, pid) = *args;

        if let Err(err) = enter_mnt_namespace(*pid) {
            let ret = Box::new(Some(Err(err)));
            ret_ptr.store(Box::leak(ret) , Ordering::SeqCst)
        }

        let ret = Box::new(Some(f()));
        ret_ptr.store(Box::leak(ret), Ordering::SeqCst);

        return 0;
    };

    let mut stack = vec![0u8; 1024 * 1024];

    let pid_ns_path = format!("/proc/{}/ns/pid", pid);
    let pid_ns = open(pid_ns_path.as_str(), OFlag::O_RDONLY, stat::Mode::all())?;

    setns(pid_ns, CloneFlags::CLONE_NEWPID)?;

    let pid = unsafe {
        libc::clone(
            callback::<F, R>,
            (stack.as_mut_ptr() as *mut libc::c_void).add(1024 * 1024),
            libc::CLONE_VM | libc::CLONE_FILES | libc::CLONE_SIGHAND | libc::SIGCHLD,
            Box::into_raw(args) as *mut libc::c_void,
        )
    };
    println!("clone returned {}", pid);

    loop {
        let ret = ret_ptr.load(Ordering::SeqCst);
        unsafe {
            match (&mut *ret).take() {
                Some(ret) => {
                    return ret
                }
                None => {}
            }
        }
    }
}
