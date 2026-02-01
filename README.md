# Booklog

Booklog is a simple tool to track read books.

## Development

The preferred development environment is the provided devcontainer. There are VS Code tasks defined to automatically
start the Go HTTP server and the Vite server. Unfortunately, the first time the devcontainer is created the tasks will
run before the devcontainer is fully setup. So those tasks will need to be manually restarted the first time the project
is used.

Tests are run with `rake`.

```
rake
```

There is a rake task that will automatically recompile and restart the backend server whenever any Go code changes.

```
rake rerun
```

In another terminal start the vite development server.

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
