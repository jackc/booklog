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

The following environment variables must be set:

* `TEST_DATABASE`: the test database name
* `TEST_DATABASE_COUNT`: the number of test databases to use

Set these variables in `.envrc`.

Run tests with `rake`.
