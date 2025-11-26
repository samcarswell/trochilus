-- migrate:up
alter table runs
add column pid int default null;

-- migrate:down
alter table runs
drop column pid;
