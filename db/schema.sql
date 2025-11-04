CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE crons (
    id integer primary key autoincrement,
    name varchar not null, -- TODO: This needs to contain no spaces, quotes etc
    unique(name)
);
CREATE TABLE runs (
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
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20251104051400');
