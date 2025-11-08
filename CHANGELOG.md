# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Moved `hostname` config to `notify.hostname`.
- Moved `slack.channel` config to `notify.slack.channel`.
- Moved `slack.token` config to `notify.slack.token`.
- Allow crons to be configured to include log contents in notify messages.

## [0.0.1] - 2025-11-04

### Added

- `exec` command to run crons.
- `run list` command to list cron runs.
- `run watch` command to tail logs of runs.
- `cron add` command to add crons.
- `cron list` command to list crons.

