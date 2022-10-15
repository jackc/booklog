create table users (
  id bigint primary key,
  username text not null unique,
  password_digest text not null,
  insert_time timestamptz not null default now(),
  update_time timestamptz not null default now()
);
select set_default_to_next_duid_block('users', 'id', 'user_id_seq');

create trigger on_user_update
before update on users
for each row execute procedure timestamp_update();

grant select, insert, update, delete on table users to {{.app_user}};
grant usage on sequence user_id_seq to {{.app_user}};

---- create above / drop below ----

drop table users;
drop sequence user_id_seq;
