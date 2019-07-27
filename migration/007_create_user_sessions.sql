create extension pgcrypto;

create table user_sessions (
  id uuid primary key default gen_random_uuid(),
  user_id bigint not null references users on delete cascade,
  login_time timestamptz not null default now()
);

create index on user_sessions (user_id);

grant select, insert, delete, update on table user_sessions to {{.app_user}};

---- create above / drop below ----

drop table user_sessions;
