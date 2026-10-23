# S3backups

Back up SQLite databases or whole folders and sync them, along with optional image
directories, to S3 using rclone. Old local backups are removed according to the
configured retention count; `rclone sync` mirrors that retention to S3.

## Run on Linux

Download `s3-backups-linux-amd64` (x86-64) or `s3-backups-linux-arm64` (ARM64) from
[Releases](https://github.com/vincedelmas/S3backups/releases). Download `SHA256SUMS`
alongside your binary and verify it:

```sh
sha256sum --ignore-missing -c SHA256SUMS
chmod +x s3-backups-linux-amd64
```

Install and configure [rclone](https://rclone.org/) with access to your S3 storage.
The program also uses `cp` for database backups and `tar` for folder backups.
Create a `config.json` in the directory where you run the program:

```json
{
  "db_backups_to_keep": 3,
  "sites": [
    {
      "name": "example",
      "s3_remote": "my-s3",
      "s3_bucket": "my-backup-bucket",
      "s3_db_path": "example/database",
      "s3_images_path": "example/images",
      "db_path": "/srv/example/site.db",
      "images_dir": "/srv/example/images",
      "data_folder_path": ""
    }
  ]
}
```

Use `data_folder_path` instead of `db_path` to archive an entire folder. Leave
`images_dir` empty to skip image syncing. Keep the retention count positive.
`config.json` is ignored by Git; keep credentials in your external rclone config.

Run from that directory, or schedule the command with cron:

```sh
./s3-backups-linux-amd64
```

## Development and releases

Build with the Go version specified in `go.mod`:

```sh
go test ./...
go vet ./...
go build .
```

CI checks pull requests and builds Linux amd64 and arm64 binaries. Use
[Conventional Commits](https://www.conventionalcommits.org/) in commits or squash
merge titles: `fix:` for patches, `feat:` for features, and `feat!:` for breaking
changes.

[Release Please](https://github.com/googleapis/release-please-action) opens or
updates a release PR after changes land on `master`. Merge that PR to publish the
version and changelog. The same workflow builds both Linux binaries and attaches
them with `SHA256SUMS` to the GitHub release. The first release starts at `v0.1.0`.

The workflow uses the built-in `GITHUB_TOKEN`; no extra secret is required.
Repository settings must allow GitHub Actions to create pull requests. PRs
created with that token do not trigger PR workflows; the release workflow checks
and builds the merged code before publishing.
