# S3backups

Back up folders or SQLite files and send them to S3 with `rclone`.

1. Install `rclone` and `tar`. Set up your S3 remote in `rclone`.
2. Download and unpack the Linux binary from GitHub Releases.
3. Create `config.json`:

```json
{
  "db_backups_to_keep": 7,
  "sites": [
    {
      "name": "my-site",
      "data_folder_path": "/srv/my-site/data",
      "folder_backups_path": "/mnt/backups/my-site/folderBackups",
      "images_dir": "/srv/my-site/images",
      "s3_remote": "my-remote",
      "s3_bucket": "my-bucket",
      "s3_db_path": "my-site/backups",
      "s3_images_path": "my-site/images"
    }
  ]
}
```

Change the paths and S3 names to match your setup.

- `db_backups_to_keep`: how many local backups to keep.
- `data_folder_path`: local folder to back up as a `.tar.gz` file.
- `db_path`: local SQLite file to back up instead. Leave out `data_folder_path` to use it.
- `folder_backups_path`: where to save folder backups. Keep it outside the data folder.
  If left out, the app uses `folderBackups` next to the data folder.
- `images_dir`: local image folder to sync. Leave out to skip images.
- `s3_remote`: the remote name you set in `rclone`.
- `s3_bucket`: the S3 bucket name.
- `s3_db_path`: folder inside the bucket for SQLite or folder backups.
- `s3_images_path`: folder inside the bucket for images.

Sync removes files from these S3 folders when they are removed locally.

Run from the folder that holds `config.json`:

```sh
./s3-backups
```
