drop function set_default_to_next_duid_block(regclass, text, text);
drop function create_sequence_for_duid_block(int, text);

create function create_sequence_for_duid_block(_block_id int, _schema_name text, _sequence_name text) returns regclass
as $$
declare
  _minvalue bigint;
  _maxvalue bigint;
  _sequence regclass;
begin
  _minvalue := _block_id::bigint * (1::bigint << 32);
  _maxvalue := ((_block_id::bigint+1) * (1::bigint << 32))-1;

  execute format('create sequence %I.%I minvalue %s maxvalue %s',
    _schema_name,
    _sequence_name,
    _minvalue,
    _maxvalue
  );

  select pg_class.oid into _sequence
  from pg_class
    join pg_namespace on pg_class.relnamespace=pg_namespace.oid
  where pg_namespace.nspname=_schema_name
    and pg_class.relname=_sequence_name;

  return _sequence;
end;
$$ language plpgsql;

create function set_default_to_next_duid_block(_table regclass, _column text, _sequence_name text) returns void
as $$
declare
  _block_id int;
  _schema_name text;
  _table_name text;
  _sequence regclass;
begin
  select pg_namespace.nspname, pg_class.relname into _schema_name, _table_name
  from pg_class
    join pg_namespace on pg_class.relnamespace=pg_namespace.oid
  where pg_class.oid=_table;

  _block_id = allocate_duid_block(format('%I.%I.%I', _schema_name, _table_name, _column));
  _sequence = create_sequence_for_duid_block(_block_id, _schema_name, _sequence_name);

  update duid_blocks
  set table_name=_table,
    sequence_name=_sequence
  where id=_block_id;

  execute format('alter table %I.%I alter column %I set default nextval(%L)', _schema_name, _table_name, _column, _sequence);
end;
$$ language plpgsql;
