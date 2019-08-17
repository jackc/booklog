# Booklog

Booklog is a simple tool to track read books.

## Development

Required Prerequisites:

https://github.com/jackc/tern - for database migrations

Highly Recommended:

https://direnv.net/ - Manage environment variables
https://github.com/asdf-vm/asdf - Version management for Ruby and Node
https://github.com/jackc/react2fs - Restart server when files change

Create database and user.

```
createdb --locale=en_US -T template0 booklog_dev
createuser booklog
```

Make a copy of all files that end in `.example` but without the `.example` and edit the new files as needed to configure development environment.

```
npm install
bundle install
tern migrate -m migration -c migration/development.conf
```

Run server with rake:

```
rake rerun
```

## Testing

Create the database for the Go tests

```
createdb --locale=en_US -T template0 booklog_test
PGDATABASE=booklog_test tern migrate -m migration -c migration/test.conf
```

The `N` environment variable must be set to determine how many parallel browser tests are run. Set that variable in `.envrc`.

Create all browser test databases.

```
ruby -e '(1..ENV["N"].to_i).each { |n| `createdb --locale=en_US -T template0 booklog_browser_test_#{n}` }'
```

Migrate all browser test databases.

```
ruby -e '(1..ENV["N"].to_i).each { |n| `PGDATABASE=booklog_browser_test_#{n} tern migrate -c migration/test.conf -m migration` }'
```
