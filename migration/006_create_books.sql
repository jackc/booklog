create table books (
  id bigint primary key,
  user_id bigint not null references users on delete cascade,
  title text not null,
  author text not null,
  finish_date date not null,
  media text not null,
  insert_time timestamptz not null default now(),
  update_time timestamptz not null default now()
);
select set_default_to_next_duid_block('books', 'id', 'book_id_seq');

create trigger on_book_update
before update on books
for each row execute procedure timestamp_update();

grant select, insert, delete, update on table books to {{.app_user}};
grant usage on sequence book_id_seq to {{.app_user}}

---- create above / drop below ----

drop table books;
drop sequence book_id_seq
