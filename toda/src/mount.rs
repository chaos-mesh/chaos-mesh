use std::fs::create_dir_all;
use std::fs::File;
use std::io::Read;
use std::path::Path;

use anyhow::{anyhow, Context, Result};

use nix::mount::{mount, MsFlags};

#[derive(Debug)]
pub struct MountsInfo {
    mounts: Vec<String>,
}

impl MountsInfo {
    pub fn parse_mounts(pid: i32) -> Result<Self> {
        let mut mounts = File::open(format!("/proc/{}/mounts", pid))?;
        let mut contents = String::new();
        mounts.read_to_string(&mut contents)?;

        let mounts = contents
            .split("\n")
            .map(|item| item.split(" ").nth(1).unwrap_or("").to_owned())
            .collect();

        return Ok(MountsInfo { mounts });
    }

    pub fn non_root<P: AsRef<Path>>(&self, path: P) -> Result<bool> {
        let path = path
            .as_ref()
            .to_str()
            .ok_or(anyhow!("path with non-UTF-8 character"))?;

        for mount_point in self.mounts.iter() {
            if mount_point.contains(path) {
                // The relationship is "contain" because if we want to inject /a/b, and /a is a mount point, we can still
                // use this method.
                return Ok(true);
            }
        }
        return Ok(false);
    }

    pub fn move_mount<P1: AsRef<Path>, P2: AsRef<Path>>(
        &self,
        original_path: P1,
        target_path: P2,
    ) -> Result<()> {
        create_dir_all(target_path.as_ref())?;

        mount::<_, _, str, str>(
            Some(original_path.as_ref()),
            target_path.as_ref(),
            None,
            MsFlags::MS_MOVE,
            None,
        )
        .context(format!(
            "source: {}, target: {}",
            original_path.as_ref().display(),
            target_path.as_ref().display()
        ))?;

        return Ok(());
    }
}
