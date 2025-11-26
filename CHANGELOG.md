# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Displaying of emojis using config value `display.emoji`. Defaults to `true`.
- Colour run status output using config value `display.color.status.[succeeded|failed|running|skipped|terminated]`. Defaults to `false`.
- Handling SIGTERM in `exec` by setting run state to `Terminated`.
- `run kill` to kill runs with status `Running`.

## [0.2.0] - 2025-11-23

### Added

- Handling SIGINT in `exec` by setting run state to `Terminated`.
- Filtering by job in `run list`.
- `job update` command.
- `run term` command to manually set run state to `Terminated` for orphaned `Running` runs. eg. runs that have been killed using SIGKILL, since we cannot gracefully handle it.

### Changed

- Moved `cron [add|list|update]` to `job [add|list|update]`.
- Display times in local timezone. This can be made UTC again using config value: `localtime=false`.

### Fixed

- Issues using `exec` with commands that chain or pipe multiple commands.

### Removed

- Jobs configured with `--notify-log` only notify the raw log content; the log path has been removed.

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

