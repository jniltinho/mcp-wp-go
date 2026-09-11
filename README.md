# mcp-wp-go

A **Go MCP server** for administering one WordPress site through its REST API,
without SSH. It uses only the official MCP Go SDK:
[`github.com/modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk).

It authenticates with a **WordPress application password**, never with a server
password. Communication runs over `stdio`: `stdout` is reserved for the MCP
protocol and diagnostic messages go to `stderr`.

## Features

| Area | Tools |
|---|---|
| Posts | `wordpress_list_posts`, `wordpress_get_post`, `wordpress_create_post`, `wordpress_update_post`, `wordpress_delete_post` |
| Media | `wordpress_list_media`, `wordpress_get_media`, `wordpress_upload_media`, `wordpress_update_media`, `wordpress_delete_media` |
| Covers | `wordpress_set_post_cover` |
| Taxonomy and validation | `wordpress_list_categories`, `wordpress_list_tags`, `wordpress_site_health`, `wordpress_check_post_live` |

### Safety controls

- The WordPress URL comes only from configuration; no tool accepts an arbitrary
  URL. HTTPS is mandatory except for `localhost` development.
- Deletions require `confirm: true` and move resources to trash by default.
  Permanent deletion also requires `permanent: true`.
- Uploads allow only GIF, JPEG, PNG, and WebP. The server checks the extension,
  detected file signature, and size. Set `WP_UPLOAD_ROOT` to restrict which
  local paths the MCP process can read.
- `wordpress_create_post` always closes comments.
- `wordpress_update_post` cannot change an existing post's slug, date, or
  status, and keeps comments closed.
- `wordpress_set_post_cover` uploads the image, uses the relative URL returned
  by the site, and places it as the first body element. It **does not** set a
  featured image, avoiding duplicated covers in themes that render featured images separately.
- `wordpress_check_post_live` reads the authenticated post and requests its
  public permalink. It reports `live: true` only when the post is `publish` and
  the page returns HTTP 2xx.

## Requirements

- Go 1.27 or newer to build.
- A WordPress account with the necessary capabilities (Editor is appropriate for
  content; upload and delete capabilities depend on the account).
- A [WordPress application password](https://wordpress.org/documentation/article/application-passwords/).

> Do not use `root`, your normal account password, or database credentials.
> Revoke an application password immediately if it is exposed.

## Build

```bash
cd mcp-wp-go
cp .env.example .env
chmod 600 .env
# Edit .env with the site URL, WordPress user, and application password.
make check
make build
```

The binary is created at `dist/mcp-wp-go`. The local `.env` file is ignored by
Git. Its loader accepts only simple `KEY=VALUE` lines (optional single or double
quotes); it never executes shell expressions. Values already present in the
process environment take precedence over values in the file.

## Configuration

| Variable | Required | Description |
|---|---:|---|
| `WP_BASE_URL` | yes | Canonical URL, for example `https://wp-domain.com` |
| `WP_USERNAME` | yes | WordPress account associated with the application password |
| `WP_APP_PASSWORD` | yes | WordPress application password |
| `WP_TIMEOUT` | no | HTTP timeout; default `30s` |
| `WP_MAX_UPLOAD_BYTES` | no | Maximum upload size; default `26214400` (25 MiB) |
| `WP_UPLOAD_ROOT` | no, recommended | Root directory from which images may be uploaded |

After registering the server, call `wordpress_site_health` to validate the
credentials. It returns the authenticated user, never the application password.

## Register with an MCP client

Use absolute paths for both the binary and the private configuration file. This
is an example of a local MCP configuration entry:

```json
{
  "mcpServers": {
    "wordpress_go": {
      "type": "stdio",
      "command": "/absolute/path/to/mcp-wp-go/dist/mcp-wp-go",
      "args": [
        "--env-file",
        "/home/user/.config/mcp-wp-go.env"
      ]
    }
  }
}
```

Create `~/.config/mcp-wp-go.env` with mode `600`, using the values from
`.env.example`. Never put an application password in `.mcp.json`, a README,
Git, or logs. Reload the MCP client after changing its configuration.

## MCP call examples

Create a draft:

```json
{
  "title": "New post",
  "content": "<p>Introduction.</p>",
  "status": "draft",
  "categories": [115],
  "tags": [42]
}
```

Upload an accessible image:

```json
{
  "file_path": "/authorized/path/cover.webp",
  "alt_text": "Observability dashboards for a production service",
  "post_id": 123
}
```

Replace the first body cover image (replacement is the default):

```json
{
  "post_id": 123,
  "file_path": "/authorized/path/cover.webp",
  "alt_text": "Cover for the OpenObserve on Ubuntu guide"
}
```

Check the public post:

```json
{ "id": 1533 }
```

## Development

```bash
make fmt            # format Go source
make vet            # run standard static analysis
make test           # run tests with the race detector
make lint           # fail if formatting or vet checks fail
make check          # format, vet, test, and verify modules
make release-cross  # build Linux amd64 release archive
make clean          # remove local artifacts
```

The project does not execute remote shells and does not use SSH. Publishing is
performed through the WordPress REST API so the installation's own permissions,
hooks, and cache integration remain in effect.

## Releases

Pushing a `v*` tag runs `.github/workflows/release.yml`. The workflow tests the
module, builds a static Linux amd64 archive, and publishes it as a GitHub Release.
