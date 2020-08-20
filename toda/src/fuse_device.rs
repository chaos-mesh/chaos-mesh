use nix::sys::stat::{mknod, stat, Mode, SFlag};
use nix::Error as NixError;

pub fn read_fuse_dev_t() -> anyhow::Result<u64> {
    let fuse_stat = stat("/dev/fuse")?;

    Ok(fuse_stat.st_rdev)
}

pub fn mkfuse_node(dev: u64) -> anyhow::Result<()> {
    let mode = unsafe { Mode::from_bits_unchecked(0o666) };
    match mknod("/dev/fuse", SFlag::S_IFCHR, mode, dev) {
        Ok(()) => Ok(()),
        Err(NixError::Sys(errno)) => {
            if errno == nix::errno::Errno::EEXIST {
                Ok(())
            } else {
                Err(NixError::from_errno(errno).into())
            }
        }
        Err(err) => Err(err.into()),
    }
}
