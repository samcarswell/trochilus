# Trochilus

Simple cron (or any script) monitoring

Named after the legendary bird described by Herodotus in [The Histories](https://en.wikipedia.org/wiki/Trochilus_(crocodile_bird)).
<p>
    <p>
        <img width="512" alt="TheCronsFriend" src="https://upload.wikimedia.org/wikipedia/commons/thumb/3/33/PloverCrocodileSymbiosis.jpg/512px-PloverCrocodileSymbiosis.jpg?20130116123116">
    </p>
    <p>
        <em>The Cron's Friend</em>
    </p>
</p>

## Features

- Automatically stores stdout/stderr logs of jobs.
- Watch and tail stdout/stderr of running or past jobs.
- Keeps a history of all job runs in a local sqlite database.
- Query job runs using the `troc` cli.
- Flock functionality; ensures that only one instance of a job is ran at a time; keeps a log of skipped runs.
- Posts run results to slack; tags `@channel` on failure.
- Single executable; no daemon.

## Build

Checkout the relevant release tag. eg. `0.2.0`; you can build on any commit but there might be issues.

```bash
git checkout 0.2.0
./build
```

Will create a `troc` binary in root directory of the repo. You can output elsewhere if you like: `./build ~/.local/bin`.

## Installation

Ensure the `troc` binary is available in your path. eg. `PATH=$PATH:~/.local/bin` assuming `troc` is in `~/.local/bin`.

Executing `troc` for the first time will setup your config and database with the default settings:

```log
fsh ❯ troc exec --name "test" "echo 'Testing...'"
time=2025-11-09T14:02:31.182+11:00 level=INFO msg="Creating config directory at /home/srcarswell/.config/troc"
time=2025-11-09T14:02:31.182+11:00 level=INFO msg="Creating initial config file at /home/srcarswell/.config/troc/config.yaml"
time=2025-11-09T14:02:31.184+11:00 level=INFO msg="Logging to /tmp/trocsys_wfnzz_20251109T030231.log"
time=2025-11-09T14:02:31.184+11:00 level=INFO msg="Creating: /home/srcarswell/.config/troc/troc.db"
time=2025-11-09T14:02:31.187+11:00 level=INFO msg="Applying: 20251104051400_initial.sql"
time=2025-11-09T14:02:31.189+11:00 level=INFO msg="Applied: 20251104051400_initial.sql in 1.886512ms"
time=2025-11-09T14:02:31.189+11:00 level=INFO msg="Applying: 20251108020521_message_cron.sql"
time=2025-11-09T14:02:31.191+11:00 level=INFO msg="Applied: 20251108020521_message_cron.sql in 2.176757ms"
time=2025-11-09T14:02:31.195+11:00 level=INFO msg="Job not registered. Creating new Job with name test"
time=2025-11-09T14:02:31.196+11:00 level=INFO msg="Created job lock at /tmp/test.lock"
time=2025-11-09T14:02:31.196+11:00 level=INFO msg="Run log created at: /tmp/test.4227158531.log"
time=2025-11-09T14:02:31.197+11:00 level=INFO msg="Run created with ID 1"
time=2025-11-09T14:02:31.199+11:00 level=INFO msg="Run 1 completed: Succeeded"
{
    "ID": 1,
    "JobName": "test",
    "StartTime": "2025-11-22 12:46:01 +1100 AEDT",
    "EndTime": "2025-11-22 12:46:02 +1100 AEDT",
    "LogFile": "/tmp/test.4227158531.log",
    "SystemLogFile": "/tmp/trocsys_wfnzz_20251109T030231.log",
    "Status": "Succeeded"
    "Duration": "1s"
}
```

Subsequent runs won't do this:

```log
time=2025-11-09T14:03:18.256+11:00 level=INFO msg="Logging to /tmp/trocsys_yncwi_20251109T030318.log"
time=2025-11-09T14:03:18.261+11:00 level=INFO msg="Created job lock at /tmp/test.lock"
time=2025-11-09T14:03:18.261+11:00 level=INFO msg="Run log created at: /tmp/test.2330262208.log"
time=2025-11-09T14:03:18.262+11:00 level=INFO msg="Run created with ID 2"
time=2025-11-09T14:03:18.264+11:00 level=INFO msg="Run 2 completed: Succeeded"
{
    "ID": 2,
...
```

Updating `troc` versions may also need to apply migrations to your database,
in which case these will be logged similarly to the first run. eg.

```log
...
time=2025-11-09T14:03:18.240+11:00 level=INFO msg="Applying: 20251108020521_a_new_migration.sql"
...
```

### Config

Running `troc` for the first time will create a config file at `~/.config/troc/config.yaml`
if it does not exist with default values.

Any of the config values can also be specified using env vars:
eg. `TROC_DATABASE` or `TROC_NOTIFY_SLACK_TOKEN`.


| Name | Description | Default |
| - | - | - |
| `database` | Path to the sqlite database. | `~/.config/troc/troc.db`
| `localtime` | Display dates in local time rather than UTC. | `true`
| `lockdir` | Directory of job lock files. | `$TMPDIR` if not empty, otherwise `/tmp`
| `logdir` | Directory of job log files. | `$TMPDIR` if not empty, otherwise `/tmp`
| `notify.hostname` | Name of server when pushing notifications. eg. `job-name@hostname` | Output of `hostname`
| `notify.slack.token` | Token for slack app. | 
| `notify.slack.channel` | Slack channel to post notifications. | 

Any invocation of `troc` will check for a database located at the `database` config value.
If it does not exist, it will create it.

## Usage

### Running a job

`troc exec` handles the execution of a job. Use `--name` to specify the name of
the job. If it does not exist, it will be created with the default settings;
for anything non-default use `troc job add` before `troc exec`.

Use the args to pass the command you intend to run.

eg. `troc exec --name 'daily-sync' "rsync --avh /tmp/source-dir /tmp/dest-dir"`

This will create (if it does not exist) a job named `daily-sync` and will
execute `rsync --avh /tmp/source-dir /tmp/dest-dir` as a run of that job.

The stdout log will display the id of the run:
```
...
time=2025-11-04T18:03:44.964+11:00 level=INFO msg="Run created with ID 1"
...
```

You can use this id to see the ongoing (for long-running jobs) or completed logs:
`troc run show -r 1`
Output:
```json
{
    "ID": 1,
    "JobName": "daily-sync",
    "StartTime": "2025-11-22 12:46:01 +1100 AEDT",
    "EndTime": "2025-11-22 12:46:02 +1100 AEDT",
    "LogFile": "/tmp/daily-sync.3159256558.log",
    "SystemLogFile": "/tmp/trocsys_pgqlq_20251104T070344.log",
    "Status": "Succeeded",
    "Duration": "1s"
}
```

If you have the `notify.slack.*` config values, you can append `--notify` to the `troc exec`
command to send a notification in slack:

```
daily-sync@example-server: run 84 - ✅: Succeeded
Log: /tmp/daily-sync.3159256558.log
```

### Watching a run

Use `troc run watch -r [RUN_ID]` to tail the logs of a running job until it completes. If the job has already ran, it will print the logs and immediately exit.

### Details of a run

`troc run show -r [RUN_ID]`

### Manually terminating a run

If a `troc` run process was killed using SIGKILL, it cannot be gracefully handled.
As a result the run will be left in a `Running` state.
These can be manually set to `Terminated` using `troc run term -r [RUN_ID]`.

Note that this will not check if the process is still running, or attempt to terminate it.
Only run this if you have determined that the run is not running and it's state is still `Running`.

If the run is still in progress, and `troc run term` has been ran on it,
the run will still correctly update it's state once it completes.

### Run history

Use `troc run list` to see a list of historical runs. Optionally filter on `--name`.

### Update job info

A job name and log settings can be updated using `troc job update`.

### Crontab example:

```
PATH=$PATH:/usr/local/bin:/usr/bin # Ensuring that troc and rsync is in the path

*/5 * * * * troc exec --name 'daily-sync' "rsync --avh /tmp/source-dir /tmp/dest-dir" --notify
```

### Run states

| Name | Description |
| - | - |
| `Running` | Run is still actively running. |
| `Skipped` | The run was skipped as there is already another run of the same job in progress. |
| `Succeeded` | The run completed with an exit code == 0. |
| `Failed` | The run completed with an exit code != 0. |
| `Terminated` | The run was interrupted. |


## Troubleshooting

If `exec` fails to create a run, it errored before it could create the run.
This is probably caused by configuration issues. You can debug this by looking through the
system logs: `[logdir]/trocsys*.log`

## Development

### SQL queries

All query code is generated from `./db/query.sql` using [sqlc](https://github.com/sqlc-dev/sqlc).

To update/add a query, make the changes needed to `./db/query.sql` and then run
`sqlc generate`. This will generate code in `./data`; no non-generated code
should be placed in this directory.

### Database migrations

`./migration [migration_name]`

This will create a new migration file in `./db/migrations/` which can be updated
to include the raw sql migration statements.

All `./db/migrations/*sql` files are embedded into the binary and ran
at startup; so just adding the migration file to that directory is enough.

### Running tests

`go test ./...`

### Test coverage

`./coverage` to open a browser with the test coverage info.

## Goals

- TUI interface.
- A local `troc` should be able to connect to a remote `troc` using SSH.
- More notification options.
