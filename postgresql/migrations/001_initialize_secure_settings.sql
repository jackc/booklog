-- Note that "public" has two meanings. It is the name of a schema but it is all a keywork meaning all roles.

do $$
declare
  dbname text;
begin
  select current_database() into dbname;
  execute 'revoke connect on database ' || quote_ident(dbname) || ' from public';
  execute 'grant connect on database ' || quote_ident(dbname) || ' to {{.app_user}}';
end
$$;

revoke all privileges on schema public from public;
grant usage on schema public to {{.app_user}};
