use super::UpdateArgs;
use crate::modconfig::{ConfigError, ModConfig};
use crate::utils::fs::copy_dir_content;
use anyhow::{Context, anyhow};
use colored::Colorize;
use regex::Regex;
use std::{collections::HashMap, path::PathBuf, sync::LazyLock};
use thiserror::Error;
use which::which;

const STATIONEERS_APP_ID: u64 = 544550;

static SUCCESS_CHECK_REGEX: LazyLock<Regex> = LazyLock::new(|| {
    Regex::new(r#"(?i)downloaded item (\d+) to \"([^\"]*)\""#).expect("Invalid regex")
});

#[derive(Error, Debug)]
pub enum UpdateError {
    #[error(transparent)]
    Config(#[from] ConfigError),

    #[error("SteamCMD is missing. Please download / install it before setting the steam path")]
    MissingSteamCMD,

    #[error("An error occurred with steamcmd: {0}")]
    SteamCmdRuntime(anyhow::Error),

    #[error("The specified ModID was not downloaded: {0}")]
    DownloadFailure(u64),

    #[error(transparent)]
    Unknown(anyhow::Error),

    #[error(transparent)]
    Copy(#[from] std::io::Error),
}

#[derive(Debug)]
struct ModToUpdate {
    pub workshop_id: u64,
    pub install_path: PathBuf,
}

pub async fn run(update_args: UpdateArgs) -> Result<(), UpdateError> {
    let config = ModConfig::try_from(update_args.config_location.as_path())?;

    let mods_dir = update_args
        .config_location
        .parent()
        .ok_or(UpdateError::Unknown(anyhow::anyhow!(
            "config file has no parent directory"
        )))?
        .join("mods");

    let steam_path = update_args
        .steam_cmd_path
        .or_else(|| which("steamcmd").ok())
        .ok_or(UpdateError::MissingSteamCMD)?;

    let ignore_disabled = update_args.ignore_disabled;

    let mods_to_update = config
        .locals
        .iter()
        .filter_map(|local| {
            if ignore_disabled && !local.enabled {
                return None;
            };

            let workshop_id = local.path.workshop_id()?;

            let install_path = local.path.value.as_ref().and_then(|path| {
                let folder_name = PathBuf::from(path);
                let folder_name = folder_name.file_name()?;

                Some(mods_dir.join(folder_name))
            })?;

            Some(ModToUpdate {
                workshop_id,
                install_path,
            })
        })
        .collect::<Vec<_>>();

    let downloads_args = mods_to_update.iter().map(|mod_item| {
        format!(
            "+workshop_download_item {STATIONEERS_APP_ID} {0} validate",
            mod_item.workshop_id
        )
    });

    let mut steam_cmd = tokio::process::Command::new(steam_path);

    let steam_cmd = steam_cmd
        .arg("+login anonymous")
        .args(downloads_args)
        .arg("+logoff")
        .arg("+quit")
        .output();

    let downloaded_mods = process_steam_cmd(steam_cmd, &mods_to_update).await?;

    relocate_mods(&mods_to_update, downloaded_mods).await?;
    Ok(())
}

async fn process_steam_cmd(
    cmd: impl Future<Output = Result<std::process::Output, std::io::Error>>,
    mods: &[ModToUpdate],
) -> Result<HashMap<u64, PathBuf>, UpdateError> {
    let stdout = match cmd.await.context("Failed to launch steamcmd") {
        Ok(output) => {
            let Some(status) = output.status.code() else {
                return Err(UpdateError::SteamCmdRuntime(anyhow!(
                    "SteamCMD exited unexpectedly"
                )));
            };

            if !output.status.success() {
                return Err(UpdateError::SteamCmdRuntime(anyhow!(
                    "Non-0 exit status: {status}"
                )));
            }

            String::from_utf8(output.stdout)
                .map_err(|e| UpdateError::SteamCmdRuntime(anyhow!("Invalid stdout: {e}")))?
        }
        Err(e) => return Err(UpdateError::SteamCmdRuntime(anyhow!(e))),
    };

    let downloaded_mods = SUCCESS_CHECK_REGEX
        .captures_iter(&stdout)
        .map(|cap| {
            let id = cap[1].parse::<u64>().unwrap_or(0);
            let path = PathBuf::from(cap[2].to_string());

            (id, path)
        })
        .collect::<HashMap<u64, PathBuf>>();

    for mod_item in mods {
        if !downloaded_mods.contains_key(&mod_item.workshop_id) {
            return Err(UpdateError::DownloadFailure(mod_item.workshop_id));
        }
        println!(
            "{0} downloaded update for ModID {1}",
            "Success:".green(),
            mod_item.workshop_id
        );
    }

    Ok(downloaded_mods)
}

async fn relocate_mods(
    mods_configs: &[ModToUpdate],
    downloaded_mods: HashMap<u64, PathBuf>,
) -> Result<(), UpdateError> {
    println!("\n--Relocating mods--\n");

    for mod_config in mods_configs {
        let downloaded_into = downloaded_mods
            .get(&mod_config.workshop_id)
            .ok_or(UpdateError::DownloadFailure(mod_config.workshop_id))?;

        println!("{}", mod_config.workshop_id.to_string().green().bold());
        println!("  From: {:#?}", downloaded_into);

        copy_dir_content(downloaded_into, &mod_config.install_path)?;
    }

    Ok(())
}
