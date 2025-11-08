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

- Automatically stores stdout/stderr logs of cron jobs.
- Watch and tail stdout/stderr of running or past cron jobs.
- Keeps a history of all cron job runs in a local sqlite database.
- Query cron job runs using the `troc` cli.
- Flock functionality; ensures that only one instance of a cron job is ran at a time; keeps a log of skipped runs.
- Posts run results to slack; tags `@channel` on failure.

## Installation

Build the `troc` CLI:

`go build -o troc`

Move the `troc` CLI to somewhere in your `$PATH`:

eg. `mv troc /usr/local/bin`

Ensure this path is available in your crontab:

eg. `PATH=$PATH:/usr/local/bin`

### Config

Running `troc` for the first time will create a config file at `~/.config/troc/config.yaml`
if it does not exist with default values.

Any of the config values can also be specified using env vars:
eg. `TROC_DATABASE` or `TROC_SLACK_TOKEN`.


| Name | Description | Default |
| - | - | - |
| `database` | Path to the sqlite database. | `~/.config/troc/troc.db`
| `hostname` | Name of server when pushing notifications. eg. `cron-name@hostname` | Output of `hostname`
| `lockdir` | Directory of cron lock files. | `$TMPDIR` if not empty, otherwise `/tmp`
| `logdir` | Directory of cron log files. | `$TMPDIR` if not empty, otherwise `/tmp`
| `slack.token` | Token for slack app. | 
| `slack.channel` | Slack channel to post notifications. | 

Any invocation of `troc` will check for a database located at the `database` config value.
If it does not exist, it will create it.

## Running a cron

`troc exec` handles the execution of a job. Use `--name` to specify the name of
the cronjob. If it does not exist, it will be created.

Use the args to pass the command you intend to run.

eg. `troc exec --name 'daily-sync' "rsync --avh /tmp/source-dir /tmp/dest-dir"`

This will create (if it does not exist) a cron named `daily-sync` and will
execute `rsync --avh /tmp/source-dir /tmp/dest-dir` as a run of that cron.

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
    "CronName": "daily-sync",
    "StartTime": "2025-11-04T07:03:44Z",
    "EndTime": "2025-11-04T07:03:54Z",
    "LogFile": "/tmp/daily-sync.3159256558.log",
    "SystemLogFile": "/tmp/trocsys_pgqlq_20251104T070344.log",
    "Status": Succeeded
}
```

If you have the `slack.*` config values, you can append `--notify` to the `troc exec`
command to send a notification in slack:

```
daily-sync@example-server: run 84 - ✅: Succeeded
Log: /tmp/daily-sync.3159256558.log
```

## Watching a running cron

Use `troc run watch -r [RUN_ID]` to tail the logs of a running cron until it completes. If the cron has already ran, it will print the logs and immediately exit.

## Run history

Use `troc run list` to see a list of historical runs.

## Crontab example:

```
PATH=$PATH:/usr/local/bin:/usr/bin # Ensuring that troc and rsync is in the path

*/5 * * * * troc exec --name 'daily-sync' "rsync --avh /tmp/source-dir /tmp/dest-dir" --notify
```

## Troubleshooting

If `exec` fails to create a run, it errored before it could create the run.
This is probably caused by configuration issues. You can debug this by looking through the
system logs: `[logdir]/trocsys*.log`

## Goals

- TUI interface.
- A local `troc` should be able to connect to a remote `troc` using SSH.
- More notification options.
