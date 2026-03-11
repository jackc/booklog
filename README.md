# Booklog

Booklog is a simple tool to track read books.

## Development

The preferred development environment is the provided devcontainer. There are VS Code tasks defined to automatically
start the Go HTTP server and the Vite server.

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

## Testing

The following environment variables must be set:

* `TEST_DATABASE`: the test database name
* `TEST_DATABASE_COUNT`: the number of test databases to use

They are preset in `.mise.toml`. If you want to use a different number of parallel tests change `TEST_DATABASE_COUNT` in
`.mise.local.toml`.

Run tests with `rake`.
