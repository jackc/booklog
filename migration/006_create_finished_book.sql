create table finished_book (
  id bigint primary key,
  reader_id bigint not null references login_account on delete cascade,
  title text not null,
  author text not null,
  date_finished date not null,
  media text not null,
  insert_time timestamptz not null default now(),
  update_time timestamptz not null default now()
);
select set_default_to_next_duid_block('finished_book', 'id', 'finished_book_id_seq');

create trigger on_finished_book_update
before update on finished_book
for each row execute procedure timestamp_update();

grant select, insert, delete, update on table finished_book to {{.app_user}};
grant usage on sequence finished_book_id_seq to {{.app_user}}

---- create above / drop below ----

drop table finished_book;
drop sequence finished_book_id_seq
