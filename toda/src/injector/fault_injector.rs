use super::filter;
use super::Injector;

use super::injector_config::FaultsConfig;
use crate::hookfs::{Error, Result};

use async_trait::async_trait;
use nix::errno::Errno;
use rand::Rng;
use tracing::{trace, info};

use std::path::Path;

#[derive(Debug)]
pub struct FaultInjector {
    filter: filter::Filter,

    errnos: Vec<(Errno, i32)>,

    sum: i32,
}

#[async_trait]
impl Injector for FaultInjector {
    #[tracing::instrument]
    async fn inject(&self, method: &filter::Method, path: &Path) -> Result<()> {
        info!("test filter");
        if self.filter.filter(method, path) {
            info!("inject io fault");
            let mut rng = rand::thread_rng();
            let attempt: f64 = rng.gen();
            let mut attempt = (attempt * (self.sum as f64)) as i32;

            for (err, p) in self.errnos.iter() {
                attempt -= p;

                if attempt < 0 {
                    info!("return with error {}", err);
                    return Err(Error::Sys(*err));
                }
            }
        }

        return Ok(());
    }
}

impl FaultInjector {
    #[tracing::instrument]
    pub fn build(conf: FaultsConfig) -> anyhow::Result<Self> {
        trace!("build fault injector");

        let errnos: Vec<_> = conf
            .faults
            .iter()
            .map(|item| (Errno::from_i32(item.errno), item.weight))
            .collect();

        let sum = errnos.iter().fold(0, |acc, w| acc + w.1);
        Ok(Self {
            filter: filter::Filter::build(conf.filter)?,
            errnos,
            sum,
        })
    }
}
