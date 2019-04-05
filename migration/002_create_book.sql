create table book (
  id uuid primary key default gen_random_uuid(),
  title text not null,
  author text not null,
  date_finished date not null,
  media text not null,
  insert_time timestamptz not null default now(),
  update_time timestamptz not null default now()
);

create trigger on_book_update
before update on book
for each row execute procedure timestamp_update();

grant select, insert, delete, update on table book to {{.app_user}};
