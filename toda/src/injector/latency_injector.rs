use async_trait::async_trait;

use std::path::Path;
use std::time::Duration;

use super::filter;
use super::injector_config::LatencyConfig;
use super::Injector;
use crate::hookfs::Result;

use tokio::time::delay_for;
use tracing::{trace ,info};

#[derive(Debug)]
pub struct LatencyInjector {
    latency: Duration,
    filter: filter::Filter,
}

#[async_trait]
impl Injector for LatencyInjector {
    #[tracing::instrument]
    async fn inject(&self, method: &filter::Method, path: &Path) -> Result<()> {
        trace!("test for filter");
        if self.filter.filter(method, path) {
            info!("inject io delay {:?}", self.latency);
            delay_for(self.latency).await;
            info!("latency finished");
        }

        return Ok(());
    }
}

impl LatencyInjector {
    #[tracing::instrument]
    pub fn build(conf: LatencyConfig) -> anyhow::Result<Self> {
        trace!("build latency injector");

        Ok(Self {
            latency: conf.latency,
            filter: filter::Filter::build(conf.filter)?,
        })
    }
}
