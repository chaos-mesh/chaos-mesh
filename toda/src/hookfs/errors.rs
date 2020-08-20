use nix::errno::Errno;
use nix::Error;
use thiserror::Error;

use tracing::error;

#[derive(Error, Debug)]
pub enum HookFsError {
    #[error("errno {0}")]
    Sys(Errno),

    #[error("cannot find inode({inode})")]
    InodeNotFound { inode: u64 },

    #[error("cannot find fh({fh})")]
    FhNotFound { fh: u64 },

    #[error("invalid string")]
    InvalidStr,

    #[error("unknown file type")]
    UnknownFileType,

    #[error("strip prefix error")]
    StripPrefixError(#[from] std::path::StripPrefixError),

    #[error("unknown error")]
    UnknownError,
}

pub type Result<T> = std::result::Result<T, HookFsError>;

impl HookFsError {
    pub fn last() -> HookFsError {
        HookFsError::from(nix::Error::last())
    }
}

impl From<nix::Error> for HookFsError {
    #[tracing::instrument]
    fn from(err: Error) -> HookFsError {
        // TODO: match more error types
        match err {
            Error::Sys(errno) => HookFsError::Sys(errno),
            _ => {
                error!("unknown error {:?}", err);
                HookFsError::UnknownError
            }
        }
    }
}

impl From<std::ffi::NulError> for HookFsError {
    fn from(_: std::ffi::NulError) -> HookFsError {
        HookFsError::InvalidStr
    }
}

impl From<std::io::Error> for HookFsError {
    #[tracing::instrument]
    fn from(err: std::io::Error) -> HookFsError {
        error!("unknown error {:?}", err);
        HookFsError::UnknownError
    }
}

impl From<tokio::task::JoinError> for HookFsError {
    #[tracing::instrument]
    fn from(err: tokio::task::JoinError) -> HookFsError {
        error!("unknown error {:?}", err);
        HookFsError::UnknownError
    }
}

impl Into<libc::c_int> for HookFsError {
    fn into(self) -> libc::c_int {
        use HookFsError::*;

        match self {
            Sys(errno) => errno as i32,
            InodeNotFound { inode: _ } => libc::EFAULT,
            FhNotFound { fh: _ } => libc::EFAULT,
            UnknownFileType => libc::EINVAL,
            InvalidStr => libc::EINVAL,
            _ => libc::EFAULT,
        }
    }
}
