use serde::{Deserialize, Serialize};

use std::time::Duration;

#[derive(Serialize, Deserialize, Clone, Debug)]
#[serde(tag = "type")]
#[serde(rename_all = "camelCase")]
pub enum InjectorConfig {
    Latency(LatencyConfig),
    Fault(FaultsConfig),
    AttrOverride(AttrOverrideConfig),
}

#[derive(Serialize, Deserialize, Clone, Debug)]
#[serde(rename_all = "camelCase")]
pub struct LatencyConfig {
    #[serde(flatten)]
    pub filter: FilterConfig,
    #[serde(with = "humantime_serde")]
    pub latency: Duration,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
#[serde(rename_all = "camelCase")]
pub struct FaultsConfig {
    #[serde(flatten)]
    pub filter: FilterConfig,

    pub faults: Vec<FaultConfig>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
#[serde(rename_all = "camelCase")]
pub struct FilterConfig {
    pub path: String,
    pub methods: Option<Vec<String>>,
    pub percent: i32,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
#[serde(rename_all = "camelCase")]
pub struct FaultConfig {
    pub errno: i32,
    pub weight: i32,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
#[serde(rename_all = "camelCase")]
pub struct AttrOverrideConfig {
    pub path: String,
    pub percent: i32,

    pub ino: Option<u64>,
    pub size: Option<u64>,
    pub blocks: Option<u64>,
    pub atime: Option<Timespec>,
    pub mtime: Option<Timespec>,
    pub ctime: Option<Timespec>,
    pub kind: Option<FileType>,
    pub perm: Option<u16>,
    pub nlink: Option<u32>,
    pub uid: Option<u32>,
    pub gid: Option<u32>,
    pub rdev: Option<u32>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
#[serde(rename_all = "camelCase")]
pub enum FileType {
    NamedPipe,
    CharDevice,
    BlockDevice,
    Directory,
    RegularFile,
    Symlink,
    Socket,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
#[serde(rename_all = "camelCase")]
pub struct Timespec {
    pub sec: i64,
    pub nsec: i32,
}
