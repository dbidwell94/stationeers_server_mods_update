#![allow(dead_code)]

use regex::Regex;
use serde::Deserialize;
use std::{path::Path, sync::LazyLock};
use thiserror::Error;

static WORKSHOP_REGEX: LazyLock<Regex> =
    LazyLock::new(|| Regex::new(r#"(?i)workshop_(\d+)"#).expect("Invalid Workshop regex"));

#[derive(Error, Debug)]
pub enum ConfigError {
    #[error("Could not open config file at {path}")]
    IoError {
        path: std::path::PathBuf,
        #[source]
        source: std::io::Error,
    },
    #[error(transparent)]
    ParseError(#[from] quick_xml::de::DeError),
}

// 1. Root Struct
#[derive(Deserialize, Debug)]
#[serde(rename = "ModConfig")]
pub struct ModConfig {
    #[serde(rename = "Core")]
    pub core: ConfigItem,

    // Use default to return an empty list if no <Local> tags exist
    #[serde(rename = "Local", default)]
    pub locals: Vec<ConfigItem>,
}

// 2. The Item (Core/Local)
#[derive(Deserialize, Debug)]
pub struct ConfigItem {
    #[serde(rename = "@Enabled")]
    pub enabled: bool,

    #[serde(rename = "Path")]
    pub path: PathItem,
}

// 3. The Path Struct
#[derive(Deserialize, Debug)]
pub struct PathItem {
    #[serde(rename = "@Value")]
    pub value: Option<String>,
}

impl PathItem {
    pub fn workshop_id(&self) -> Option<u64> {
        let text = self.value.as_deref()?;

        let cap = WORKSHOP_REGEX.captures(text)?;

        cap[1].parse::<u64>().ok()
    }
}

impl TryFrom<&Path> for ModConfig {
    type Error = ConfigError;

    fn try_from(path: &Path) -> Result<Self, Self::Error> {
        let file_data = std::fs::read_to_string(path).map_err(|e| ConfigError::IoError {
            path: path.to_path_buf(),
            source: e,
        })?;

        let config: ModConfig = quick_xml::de::from_str(&file_data)?;
        Ok(config)
    }
}
