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
createdb booklog
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
