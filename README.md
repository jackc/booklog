# Booklog

Booklog is a simple tool to track read books.

## Development

The preferred development environment is the provided devcontainer. There are VS Code tasks defined to automatically
start the Go HTTP server and the Vite server. The backend server will be recompiled and restarted whenever any Go code
changes.

If you are not using the VS Code tasks then you can manually run `rake rerun` and `npx vite` to start the Go HTTP server
and Vite server respectively.

Site will be available at: http://localhost:5173/

## Testing

The following environment variables must be set:

* `TEST_DATABASE`: the test database name
* `TEST_DATABASE_COUNT`: the number of test databases to use

They are preset in `.mise.toml`. If you want to use a different number of parallel tests change `TEST_DATABASE_COUNT` in
`.mise.local.toml`.

Run tests with `rake`.
