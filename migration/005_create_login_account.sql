create table login_account (
  id bigint primary key,
  username text not null unique,
  password_digest text not null,
  insert_time timestamptz not null default now(),
  update_time timestamptz not null default now()
);
select set_default_to_next_duid_block('login_account', 'id', 'login_account_id_seq');

create trigger on_login_account_update
before update on login_account
for each row execute procedure timestamp_update();

grant select, insert, update, delete on table login_account to {{.app_user}};
grant usage on sequence login_account_id_seq to {{.app_user}};

---- create above / drop below ----

drop table login_account;
drop sequence login_account_id_seq;
