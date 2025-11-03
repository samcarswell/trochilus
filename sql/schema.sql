create table if not exists crons (
    id integer primary key autoincrement,
    name varchar not null, -- TODO: This needs to contain no spaces, quotes etc
    unique(name)
);

create table if not exists runs (
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
