# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Handling SIGINT in `exec` by setting run state to `Killed`.
- Filtering by cron in `run list`.
- `cron update` command.
- Optionally display times in local timezone. Config value: `localtime=true`.
- `run kill` command to manually set run state to `Killed` for orphaned `Running` runs. eg. runs that have been killed using SIGKILL, since we cannot gracefully handle it.

### Fixed

- Issues using `exec` with commands that chain or pipe multiple commands.

### Removed

- Crons configured with `--notify-log` only notify the raw log content; the log path has been removed.

## [0.1.1] - 2025-11-09

### Fixed

- Issue where database migration logs were sent stdout, breaking JSON output.
- Uniform formatting for logs. Stderr is formatted as text, files are JSON.
- `run watch` now fails for non-existent runs.

## [0.1.0] - 2025-11-09

### Added

- Allow crons to be configured to include log contents in notify messages.
- `lockdir` can now be used to configure lock directory.
- `--version` option to display `troc` version.

### Changed

- Moved `hostname` config to `notify.hostname`.
- Moved `slack.channel` config to `notify.slack.channel`.
- Moved `slack.token` config to `notify.slack.token`.

### Fixed

- Some concurrency issues with the database. 

## [0.0.1] - 2025-11-04

### Added

- `exec` command to run crons.
- `run list` command to list cron runs.
- `run watch` command to tail logs of runs.
- `cron add` command to add crons.
- `cron list` command to list crons.
- `run show` command to display run.

