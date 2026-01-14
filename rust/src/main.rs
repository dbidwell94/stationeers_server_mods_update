mod modconfig;
mod update;
mod utils;

use clap::{Args, Parser};
use colored::Colorize;
use std::path::PathBuf;
use thiserror::Error;

#[derive(Error, Debug)]
enum ProgramError {
    #[error(transparent)]
    Update(#[from] update::UpdateError),
}

#[derive(Args, Debug)]
struct UpdateArgs {
    /// Which mods should be updated. If omitted, will attempt to update all mods
    #[arg(short = 'm', long = "mod-id")]
    mod_id: Option<Vec<String>>,

    /// Skip updating mods that are not enabled
    #[arg(short = 'd', long = "ignore-disabled", default_value_t = true)]
    ignore_disabled: bool,

    /// If a mod has an update, the previous version will be backed up
    #[arg(short = 'b', long = "backup-updated", default_value_t = false)]
    backup_updated: bool,

    /// Location for the modconfig.xml
    #[arg(short = 'c', long = "config-location", env)]
    config_location: PathBuf,

    /// The location to the SteamCMD binary. If omitted, assumes it's in `$PATH`
    #[arg(short = 's', long = "steam-cmd-path", env)]
    steam_cmd_path: Option<PathBuf>,
}

#[derive(Parser, Debug)]
#[command(version, about)]
enum Command {
    /// Attempt to update one or all mods
    Update(UpdateArgs),
}

#[tokio::main]
async fn main() {
    if let Err(e) = run().await {
        eprint!("{}: ", "Error".red().bold());

        eprintln!("{}", e);

        let mut src = e.source();

        while let Some(s) = src {
            eprintln!("  {}: {}", "Caused by".yellow().dimmed(), s);
            src = s.source();
        }

        std::process::exit(1);
    }
}

async fn run() -> anyhow::Result<()> {
    let args = Command::parse();

    let res: Result<(), ProgramError> = match args {
        Command::Update(args) => Ok(update::run(args).await?),
    };

    res.map_err(|e| anyhow::anyhow!(e).context("Program failed"))
}
