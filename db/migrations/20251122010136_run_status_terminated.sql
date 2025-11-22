-- migrate:up
create table if not exists runs1 (
    id integer primary key autoincrement,
    cron_id int not null,
    start_time timestamp not null,
    end_time timestamp,
    log_file varchar not null,
    exec_log_file varchar not null,
    status varchar not null,
    constraint fk_cron_id foreign key(cron_id) references crons(id),
    constraint ck_status check (status in ("Running", "Skipped", "Succeeded", "Failed", "Killed"))
);
insert into runs1
(id, cron_id, start_time, end_time, log_file, exec_log_file, status)
    select id, cron_id, start_time, end_time, log_file, exec_log_file, status
    from runs;
drop table runs;
alter table runs1 rename to runs;

-- migrate:down
create table if not exists runs1 (
    id integer primary key autoincrement,
    cron_id int not null,
    start_time timestamp not null,
    end_time timestamp,
    log_file varchar not null,
    exec_log_file varchar not null,
    status varchar not null,
    constraint fk_cron_id foreign key(cron_id) references crons(id),
    check (status in ("Running", "Skipped", "Succeeded", "Failed"))
);

insert into runs1
(id, cron_id, start_time, end_time, log_file, exec_log_file, status)
    select id, cron_id, start_time, end_time, log_file, exec_log_file, status
    from runs
    where status != "Killed";
insert into runs1
(id, cron_id, start_time, end_time, log_file, exec_log_file, "Failed")
    select id, cron_id, start_time, end_time, log_file, exec_log_file, status
    from runs
    where status = "Killed";
drop table runs;
alter table runs1 rename to runs;
