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
    (cron_id, start_time, log_file, exec_log_file, status)
values (?, current_timestamp, ?, ?, "Running")
returning id;

-- name: EndRun :exec
update runs
set end_time = current_timestamp, status = ?
where id = ?;

-- name: SkipRun :one
insert into runs
    (cron_id, start_time, end_time, log_file, exec_log_file, status)
values (?, current_timestamp, current_timestamp, "", ?, "Skipped")
returning id;

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

