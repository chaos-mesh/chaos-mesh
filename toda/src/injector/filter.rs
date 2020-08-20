use std::convert::TryFrom;
use std::path::Path;

use super::injector_config::FilterConfig;

use anyhow::{anyhow, Error, Result};
use bitflags::bitflags;
use glob::{MatchOptions, Pattern};
use rand::Rng;

use tracing::{trace, info};

bitflags! {
    pub struct Method: u32 {
        const LOOKUP = 1;
        const FORGET = 1<<1;
        const GETATTR = 1<<2;
        const SETATTR = 1<<3;
        const READLINK = 1<<4;
        const MKNOD = 1<<5;
        const MKDIR = 1<<6;
        const UNLINK = 1<<7;
        const RMDIR = 1<<8;
        const SYMLINK = 1<<9;
        const RENAME = 1<<10;
        const LINK = 1<<11;
        const OPEN = 1<<12;
        const READ = 1<<13;
        const WRITE = 1<<14;
        const FLUSH = 1<<15;
        const RELEASE = 1<<16;
        const FSYNC = 1<<17;
        const OPENDIR = 1<<18;
        const READDIR = 1<<19;
        const RELEASEDIR = 1<<20;
        const FSYNCDIR = 1<<21;
        const STATFS = 1<<22;
        const SETXATTR = 1<<23;
        const GETXATTR = 1<<24;
        const LISTXATTR = 1<<25;
        const REMOVEXATTR = 1<<26;
        const ACCESS = 1<<27;
        const CREATE = 1<<28;
        const GETLK = 1<<29;
        const SETLK = 1<<30;
        const BMAP = 1<<31;
    }
}

impl TryFrom<&str> for Method {
    fn try_from(s: &str) -> Result<Method> {
        match s.to_lowercase().as_str() {
            "lookup" => Ok(Method::LOOKUP),
            "forget" => Ok(Method::FORGET),
            "getattr" => Ok(Method::GETATTR),
            "setattr" => Ok(Method::SETATTR),
            "readlink" => Ok(Method::READLINK),
            "mknod" => Ok(Method::MKNOD),
            "mkdir" => Ok(Method::MKDIR),
            "unlink" => Ok(Method::UNLINK),
            "rmdir" => Ok(Method::RMDIR),
            "symlink" => Ok(Method::SYMLINK),
            "rename" => Ok(Method::RENAME),
            "link" => Ok(Method::LINK),
            "open" => Ok(Method::OPEN),
            "read" => Ok(Method::READ),
            "write" => Ok(Method::WRITE),
            "flush" => Ok(Method::FLUSH),
            "release" => Ok(Method::RELEASE),
            "fsync" => Ok(Method::FSYNC),
            "opendir" => Ok(Method::OPENDIR),
            "readdir" => Ok(Method::READDIR),
            "releasedir" => Ok(Method::RELEASEDIR),
            "fsyncdir" => Ok(Method::FSYNCDIR),
            "statfs" => Ok(Method::STATFS),
            "setxattr" => Ok(Method::SETXATTR),
            "getxattr" => Ok(Method::GETXATTR),
            "listxattr" => Ok(Method::LISTXATTR),
            "removexattr" => Ok(Method::REMOVEXATTR),
            "access" => Ok(Method::ACCESS),
            "create" => Ok(Method::CREATE),
            "getlk" => Ok(Method::GETLK),
            "setlk" => Ok(Method::SETLK),
            "bmap" => Ok(Method::BMAP),
            _ => Err(anyhow!("")),
        }
    }
    type Error = Error;
}

#[derive(Debug)]
pub struct Filter {
    path_filter: Pattern,
    methods: Method,
    probability: f64,
}

impl Filter {
    #[tracing::instrument]
    pub fn build(conf: FilterConfig) -> Result<Self> {
        info!("build filter");
        let mut methods = Method::empty();
        if let Some(conf_methods) = conf.methods {
            for method in conf_methods.iter() {
                methods |= Method::try_from(method.as_str())?;
            }
        } else {
            methods = Method::all()
        }

        Ok(Self {
            path_filter: Pattern::new(&conf.path)?,
            methods,
            probability: conf.percent as f64 / 100f64,
        })
    }
    #[tracing::instrument]
    pub fn filter(&self, method: &Method, path: &Path) -> bool {
        let mut rng = rand::thread_rng();
        let p: f64 = rng.gen();

        let match_path = self.path_filter.matches_path_with(
            path,
            MatchOptions {
                case_sensitive: true,
                require_literal_separator: true,
                require_literal_leading_dot: false,
            },
        );
        let match_method = !(self.methods & *method).is_empty();
        let match_probability = p < self.probability;
        trace!("path filter: {}", match_path);
        trace!("method filter: {}", match_method);
        trace!("probability: {}", match_probability);

        return match_path && match_method
            && match_probability;
    }
}
