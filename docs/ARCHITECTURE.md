# Architecture

`mcp-wp-go` is one static Go binary with two execution paths: a stdio MCP server
when invoked without a subcommand, and a Markdown converter through the Cobra
`post-html` command.

```text
                              mcp-wp-go
                                  │
                         ┌────────▼────────┐
                         │   cmd/ (Cobra)  │
                         └───────┬─────────┘
                                 │
             no subcommand       │        post-html FILE.md
                  ┌──────────────┴──────────────┐
                  │                             │
         ┌────────▼────────┐          ┌─────────▼─────────┐
         │ internal/config│          │ internal/posthtml │
         │ env + validation│          │ gomarkdown parser │
         └────────┬────────┘          └─────────┬─────────┘
                  │                             │
         ┌────────▼────────┐                    ▼
         │ internal/server│             HTML fragment
         │ MCP tool schemas│
         └────────┬────────┘
                  │
       ┌──────────▼───────────┐
       │ internal/wordpress  │
       │ bounded REST client │
       └──────────┬───────────┘
                  │ HTTPS
          ┌───────▼────────┐
          │ WordPress REST │
          │ /wp-json/wp/v2 │
          └────────────────┘
```

## Package layout

| Path | Responsibility |
| --- | --- |
| `main.go` | Process boundary: execute the Cobra tree, print terminal errors, set exit status |
| `cmd/root.go` | Root command, persistent `--env-file`, default stdio server path |
| `cmd/post_html.go` | Markdown CLI validation, file I/O, stdout/output selection |
| `internal/config` | Environment loading, URL policy, timeouts, upload limits and roots |
| `internal/server` | MCP implementation metadata, tool schemas, confirmation and input checks |
| `internal/wordpress` | WordPress REST requests, bounded responses, media validation, cover editing |
| `internal/posthtml` | Markdown parser extensions and HTML-fragment renderer |

The `cmd` package contains orchestration. WordPress behavior lives in
`internal/wordpress`; Markdown rendering lives in `internal/posthtml`. This
keeps command parsing independently testable and prevents CLI concerns from
leaking into the REST client.

## MCP request flow

```text
MCP client
   │ JSON-RPC over stdin/stdout
   ▼
internal/server
   │ typed input + local validation
   ▼
internal/wordpress
   │ authenticated HTTPS + bounded body
   ▼
WordPress REST API
   │ typed response or contextual error
   └──────────────────────────────────▶ MCP client
```

The process reserves stdout for protocol messages. Startup and terminal errors
go to stderr so diagnostics cannot corrupt the JSON-RPC stream.

## Markdown conversion flow

```text
.md/.markdown file
   │ extension and path validation
   ▼
gomarkdown parser
   │ common extensions + tables + fenced code + heading IDs
   ▼
HTML fragment renderer
   │ stdout or --output file
   ▼
WordPress content.raw
```

The renderer intentionally preserves embedded HTML and does not create a full
HTML document. That allows responsive video wrappers and other trusted blocks,
but it makes trust at the input boundary essential.

## Media and cover flow

Media is deliberately separate from Markdown conversion:

1. `wordpress_upload_media` validates and uploads a local image.
2. WordPress returns its canonical Media Library URL and metadata.
3. Post HTML uses a site-relative `/wp-content/uploads/...` path.
4. `wordpress_set_post_cover` uploads the cover and updates the first body
   image without setting `featured_media`.

The upload root is resolved to a canonical filesystem path, limiting which
local files an MCP caller can request when `WP_UPLOAD_ROOT` is configured.

## Design decisions

- **REST instead of SSH or SQL.** WordPress permissions, hooks, and cache
  invalidation remain authoritative.
- **One configured site.** Tools cannot turn the server into an arbitrary HTTP
  client by supplying another base URL.
- **Draft-first publishing.** Create and edit are distinct from confirmed
  status changes.
- **Immutable permalink fields on update.** Slug and date changes are outside
  the exposed edit contract.
- **Relative body media URLs.** Content remains portable across canonical-host
  redirects while the configured host is validated at upload time.
- **No duplicated cover.** The cover lives in post content; `featured_media`
  remains unset.
- **Bounded network input.** REST response bodies and upload sizes have limits.
- **Static release binary.** Builds use `CGO_ENABLED=0`, `-trimpath`, and stripped
  linker flags; UPX is optional for local builds.
- **Tag-derived version.** Release tags inject the MCP implementation version
  through linker flags.

## Error and trust boundaries

- Configuration errors stop startup before the MCP server begins.
- REST errors include operation context without exposing credentials.
- Redirects are limited and cannot leave the configured WordPress host.
- Media signatures must agree with their extensions.
- Embedded Markdown HTML is trusted input and is not sanitized by the CLI.
- WordPress remains the final authorization and rendering boundary.

## Tests

- `cmd/` tests command behavior without starting a real MCP transport.
- `internal/posthtml` uses exact Markdown-to-HTML golden fixtures.
- `internal/config` tests URL and environment policies.
- `internal/wordpress` uses HTTP test servers for REST, cover, and media behavior.
- CI runs formatting checks, `go vet`, race-enabled tests, and a static build.
