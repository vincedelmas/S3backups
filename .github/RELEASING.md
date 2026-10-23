# Releases

Google Release Please runs on pushes to `main`. Use Conventional Commit messages
for changes that should be released, for example `feat: add a backup option` or
`fix: handle a missing backup directory`. When squash merging, use that format in
the PR title / squash commit message.

Release Please opens or updates a release PR with the version and changelog.
Merge that PR into `main` to create the tag and GitHub Release. The first release
is `v0.1.0`; subsequent versions are calculated from commit messages.

The same workflow then tests and builds the tagged source for Linux amd64 with
CGO disabled. Download `s3-backups-<tag>-linux-amd64.tar.gz` and `SHA256SUMS` from
the GitHub Release or the workflow run's artifacts. The archive preserves the
binary's executable permission. Verify and extract it with:

```sh
sha256sum -c SHA256SUMS
tar -xzf s3-backups-v0.1.0-linux-amd64.tar.gz
```

The program still needs its `config.json` and `rclone` configured on the machine
where it runs.

Authentication uses the built-in `GITHUB_TOKEN`; no extra secret is required.
The repository must allow GitHub Actions to create pull requests under
Settings → Actions → General → Workflow permissions. This setting was enabled
when these workflows were added.

GitHub does not run PR workflows for PRs opened by `GITHUB_TOKEN`. Normal PRs and
pushes to `main` run CI; the release workflow also tests and vets the tagged
source before publishing binaries. Binary publishing stays in the Release
Please workflow because tags/releases created by `GITHUB_TOKEN` do not trigger
separate workflows.
