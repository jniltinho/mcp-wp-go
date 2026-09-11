# Development

## Local setup

```bash
git clone https://github.com/jniltinho/mcp-wp-go.git
cd mcp-wp-go
go mod download
make check
make build
```

The module targets the Go version declared in `go.mod`.

## Commands

| Command | Purpose |
| --- | --- |
| `make build` | Build the static local binary in `dist/` |
| `make fmt` | Format all Go files with `gofmt` |
| `make vet` | Run standard Go static analysis |
| `make lint` | Run `go vet` and fail on unformatted Go source |
| `make test` | Run every package test with the race detector |
| `make check` | Format, lint, race-test, and verify downloaded modules |
| `make tidy` | Normalize `go.mod` and `go.sum` |
| `make release-cross` | Build the three supported release archives |
| `make clean` | Remove `dist/` |

Additional local QA used before releases:

```bash
golangci-lint run ./...
staticcheck ./...
```

## Test strategy

Run the complete suite:

```bash
go test -race ./...
```

Run only Markdown conversion tests:

```bash
go test ./internal/posthtml
```

Generate coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

The Markdown package uses a golden pair under `internal/posthtml/testdata/`.
When conversion behavior intentionally changes, inspect the diff between input
and generated HTML before updating the expected fixture.

The public `examples/post-completo.md` and `examples/post-completo.html` pair can
also be checked directly:

```bash
make build
dist/mcp-wp-go post-html examples/post-completo.md > /tmp/post-completo.html
diff -u examples/post-completo.html /tmp/post-completo.html
```

## Adding an MCP tool

1. Add the REST behavior and response type to `internal/wordpress`.
2. Cover it with an `httptest` server in `internal/wordpress/*_test.go`.
3. Add the typed MCP input/output and registration in `internal/server`.
4. Enforce confirmation for destructive or status-changing behavior.
5. Update [`TOOLS.md`](TOOLS.md), usage examples, and the changelog.
6. Run `make check` and the additional static-analysis commands.

Keep stdout protocol-only. Do not log ordinary diagnostics there.

## Adding a CLI command

1. Add one command file under `cmd/`.
2. Use `RunE`, Cobra argument validators, and `cmd.OutOrStdout()`.
3. Put reusable domain logic in an `internal/` package.
4. Build a fresh command tree in every test.
5. Document the command in the repository README.

## CI

`.github/workflows/ci.yml` runs on pushes and pull requests to `main`:

1. checkout;
2. install the Go version from `go.mod`;
3. `make lint`;
4. `make test`;
5. `make build`.

## Releases

The release version comes from an annotated `v*` Git tag. The release workflow:

1. runs race-enabled tests;
2. builds Linux amd64, macOS arm64, and Windows amd64 archives;
3. extracts the matching release notes from `CHANGELOG.md`;
4. creates the GitHub Release and uploads all archives.

Before tagging, move the relevant changelog items from `Unreleased` into a
dated semantic-version section and ensure all local QA is green. Never commit
credentials or files from `dist/`.
