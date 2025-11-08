-- migrate:up
alter table crons
add column notify_log_content boolean not null default false;

-- migrate:down
alter table crons
drop column notify_log_content;
