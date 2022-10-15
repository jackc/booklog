# Booklog

Booklog is a simple tool to track read books.

## Development

Required Prerequisites:

https://github.com/jackc/tern - for database migrations
https://direnv.net/ - Manage environment variables

Highly Recommended:

https://github.com/asdf-vm/asdf - Version management for Ruby and Node
https://github.com/jackc/react2fs - Restart server when files change

Make a copy of all files that end in `.example` but without the `.example` and edit the new files as needed to configure development environment.

Create database and user.

```
createdb --locale=en_US -T template0 booklog_dev
createuser booklog
```


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
PGDATABASE=booklog_test tern migrate
```

The `MT_CPU` environment variable must be set to determine how many parallel browser tests are run. Set that variable in `.envrc`.

Create all browser test databases.

```
ruby -e '(1..ENV["MT_CPU"].to_i).each { |n| `dropdb --if-exists booklog_browser_test_#{n}` }'
ruby -e '(1..ENV["MT_CPU"].to_i).each { |n| `createdb -T booklog_test booklog_browser_test_#{n}` }'
```
