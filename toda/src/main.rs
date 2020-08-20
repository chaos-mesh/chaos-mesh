#![feature(box_syntax)]
#![feature(async_closure)]

extern crate derive_more;

mod fd_replacer;
mod fuse_device;
mod hookfs;
mod injector;
mod mount;
mod mount_injector;
mod namespace;
mod ptrace;

use fd_replacer::FdReplacer;
use injector::InjectorConfig;
use mount_injector::MountInjector;

use anyhow::Result;
use nix::sys::signal::{signal, SigHandler, Signal};
use nix::sys::mman::{MlockAllFlags, mlockall};
use signal_hook::iterator::Signals;
use structopt::StructOpt;
use tracing::{info, Level};
use tracing_subscriber;

use std::path::PathBuf;
use std::str::FromStr;

#[derive(StructOpt, Debug)]
#[structopt(name = "basic")]
struct Options {
    #[structopt(short, long)]
    pid: i32,

    #[structopt(long)]
    path: PathBuf,

    #[structopt(short = "v", long = "verbose", default_value = "trace")]
    verbose: String,
}

fn main() -> Result<()> {
    mlockall(MlockAllFlags::MCL_CURRENT)?;
    // ignore dying children
    unsafe { signal(Signal::SIGCHLD, SigHandler::SigIgn)? };

    let option = Options::from_args();
    let verbose = Level::from_str(&option.verbose)?;
    let subscriber = tracing_subscriber::fmt().with_max_level(verbose).finish();
    tracing::subscriber::set_global_default(subscriber).expect("no global subscriber has been set");

    info!("parse injector configs");
    let injector_config: Vec<InjectorConfig> = serde_json::from_reader(std::io::stdin())?;
    info!("inject with config {:?}", injector_config);

    let path = option.path;
    let pid = option.pid;

    let mut fdreplacer = FdReplacer::enable(&path, pid)?;

    let mut injection = MountInjector::create_injection(&path, pid, injector_config)?;

    let fuse_dev = fuse_device::read_fuse_dev_t()?;

    let mut mount_injection = namespace::with_mnt_pid_namespace(
        box move || -> Result<_> {
            if let Err(err) = fuse_device::mkfuse_node(fuse_dev) {
                info!("fail to make /dev/fuse node: {}", err)
            }

            injection.mount()?;

            return Ok(injection);
        },
        option.pid,
    )?;

    fdreplacer.reopen()?;
    drop(fdreplacer);

    let signals = Signals::new(&[signal_hook::SIGTERM, signal_hook::SIGINT])?;

    info!("enable injection");
    mount_injection.enable_injection();

    info!("waiting for signal to exit");
    signals.forever().next();
    info!("start to recover and exit");

    info!("disable injection");
    mount_injection.disable_injection();

    fdreplacer = FdReplacer::disable(&path, pid)?;
    fdreplacer.reopen()?;

    info!("fdreplace reopened");

    namespace::with_mnt_pid_namespace(
        box move || -> Result<()> {
            info!("recovering mount");

            // TODO: retry umount multiple times
            mount_injection.recover_mount()?;
            return Ok(());
        },
        option.pid,
    )?;
    drop(fdreplacer);

    info!("fdreplace detached");
    info!("recover successfully");
    return Ok(());
}
