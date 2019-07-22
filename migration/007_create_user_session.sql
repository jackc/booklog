create extension pgcrypto;

create table user_session (
  id uuid primary key default gen_random_uuid(),
  user_id bigint not null references login_account on delete cascade,
  login_time timestamptz not null default now()
);

create index on user_session (user_id);

grant select, insert, delete, update on table user_session to {{.app_user}};

---- create above / drop below ----

drop table user_session;
