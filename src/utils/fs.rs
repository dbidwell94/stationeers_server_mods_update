use std::fs;
use std::path::Path;

use colored::Colorize;

pub fn copy_dir_content(
    src: impl AsRef<Path>,
    dst: impl AsRef<Path>,
) -> Result<(), std::io::Error> {
    let src = src.as_ref();
    let dst = dst.as_ref();

    // Ensure destination exists (including parents)
    fs::create_dir_all(dst)?;

    // Iterate through the source directory
    for entry in fs::read_dir(src)? {
        let entry = entry?;
        let file_type = entry.file_type()?;

        // Append the file name to the destination path
        // e.g. "backup/mods" + "my_mod.dll"
        let dest_path = dst.join(entry.file_name());

        if file_type.is_dir() {
            // Recurse!
            copy_dir_content(entry.path(), dest_path)?;
        } else {
            // remove the existing file first so a hard link will work
            if dest_path.exists() {
                let _ = fs::remove_file(&dest_path);
            }
            // Attempt to hard-link first.
            if let Err(e) = fs::hard_link(entry.path(), &dest_path) {
                println!("  {e}");
                println!(
                    "  {}: Hard link failed. Falling back to copy",
                    "Warning:".yellow().bold()
                );
                fs::copy(entry.path(), &dest_path)?;
                println!("  Copied to: {:#?}", dest_path);
            } else {
                println!("  Linked to: {:#?}", dest_path);
            }
        }
    }

    Ok(())
}
