-- name: GetCron :one
select
    sqlc.embed(crons)
from crons
where crons.name = ?;

-- name: CreateCron :one
insert into crons
    (name)
values (?)
returning id;

-- name: StartRun :one
insert into runs
    (cron_id, start_time, stdout_log_file, stderr_log_file, exec_log_file)
values (?, current_timestamp, ?, ?, ?)
returning id;

-- name: EndRun :exec
update runs
set end_time = current_timestamp, succeeded = ?
where id = ?;

-- name: GetCrons :many
select
    sqlc.embed(crons)
from crons;

-- name: GetRuns :many
select
    sqlc.embed(runs),
    sqlc.embed(crons)
from runs, crons
where runs.cron_id = crons.id;

-- name: GetRun :one
select
    sqlc.embed(runs),
    sqlc.embed(crons)
from runs, crons
where runs.cron_id = crons.id
and runs.id = ?;

-- name: IsRunFinished :one
select runs.end_time is not null
from runs
where runs.id = ?;

