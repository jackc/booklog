# Booklog

Booklog is a simple tool to track read books.

## Development

Required Prerequisites:

https://github.com/jackc/tern - for database migrations
https://direnv.net/ - Manage environment variables

Highly Recommended:

https://github.com/asdf-vm/asdf - Version management for Ruby and Node
https://github.com/watchexec/watchexec - Restart server when files change (needed for `rake rerun`)

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

Run asset server:

```
npx vite
```

Site will be available at: http://localhost:5173/

## Iterm2 Script

`bin/start-booklog-dev.py.example` contains an example script to start all needed programs for development. It does the following:

* Start booklog server.
* Split the window and run the asset server
* Create a tab for a console.
* Open booklog in VS Code.

Make a copy of this file without the `.example`. Symlink the file into `~/Library/Application Support/iTerm2/Scripts`. e.g. `ln -s ~/dev/booklog/bin/start-isoamp-dev.py ~/Library/Application\ Support/iTerm2/Scripts`.  This script will then be available in the iTerm2 Scripts menu as well as the cmd+shift+o "Open Quickly" window. You can now edit the file if needed.

## Testing

The following environment variables must be set:

* `TEST_DATABASE`: the test database name
* `TEST_DATABASE_COUNT`: the number of test databases to use

Set these variables in `.envrc`.

Run tests with `rake`.
