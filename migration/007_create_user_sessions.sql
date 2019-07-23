create extension pgcrypto;

create table user_sessionss (
  id uuid primary key default gen_random_uuid(),
  user_id bigint not null references users on delete cascade,
  login_time timestamptz not null default now()
);

create index on user_sessionss (user_id);

grant select, insert, delete, update on table user_sessionss to {{.app_user}};

---- create above / drop below ----

drop table user_sessionss;
