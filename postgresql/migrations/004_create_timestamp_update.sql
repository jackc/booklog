create function timestamp_update() returns trigger
language plpgsql
as $$
  begin
    new.update_time = now();
    return new;
  end;
$$;
