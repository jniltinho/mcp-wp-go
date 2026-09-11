# mcp-wp-go

A **Go MCP server and content utility CLI** for administering one WordPress
site through its REST API, without SSH. The MCP server uses the official MCP Go
SDK, and the command tree uses Cobra.

It authenticates with a **WordPress application password**, never with a server
password. Communication runs over `stdio`: `stdout` is reserved for the MCP
protocol and diagnostic messages go to `stderr`.

## Features

| Area | Tools |
|---|---|
| Posts | `wordpress_list_posts`, `wordpress_get_post`, `wordpress_create_post`, `wordpress_update_post`, `wordpress_publish_posts`, `wordpress_unpublish_posts`, `wordpress_delete_post` |
| Media | `wordpress_list_media`, `wordpress_get_media`, `wordpress_upload_media`, `wordpress_update_media`, `wordpress_delete_media` |
| Covers | `wordpress_set_post_cover` |
| Taxonomy and validation | `wordpress_list_categories`, `wordpress_list_tags`, `wordpress_content_stats`, `wordpress_site_health`, `wordpress_check_post_live` |

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
  status, and keeps comments closed. Dedicated publish and unpublish tools
  require `confirm: true`, accept one to 100 unique IDs, and report every
  result because bulk status changes are not transactional.
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

## Convert Markdown to WordPress HTML

The `post-html` subcommand converts a `.md` or `.markdown` file into an HTML
fragment suitable for the WordPress `content` field. It supports headings,
lists, links, blockquotes, tables, fenced code blocks with `language-*` classes,
and embedded HTML. It does not add `<html>` or `<body>` wrappers.

Write the HTML to stdout:

```bash
dist/mcp-wp-go post-html article.md
```

Write it to a file:

```bash
dist/mcp-wp-go post-html article.md --output article.html
```

`markdown` and `md` are aliases for `post-html`, and `-o` is the short form of
`--output`. The command refuses to overwrite its Markdown input. Embedded HTML
is preserved, so only convert trusted files before sending the result to
WordPress.

Images referenced by Markdown are written as `<img>` elements, but the command
does not upload their files. Upload images with `wordpress_upload_media` and use
the returned site-relative `/wp-content/uploads/...` URL. For a cover, create
the draft and call `wordpress_set_post_cover`; it uploads the image and places
it first without setting a duplicated featured image. Responsive YouTube
iframes can be included as trusted HTML, and fenced script blocks retain their
language class, escaping code characters such as `<` and `>`.

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

Publish one or more posts (explicit confirmation is required):

```json
{ "ids": [123, 124], "confirm": true }
```

Unpublish one or more posts by moving them to drafts:

```json
{ "ids": [123, 124], "confirm": true }
```

Check the public post:

```json
{ "id": 123 }
```

### Content totals and list values

Call `wordpress_content_stats` with an empty object to obtain a lightweight
summary. It makes three `per_page=1` REST requests and reads WordPress
`X-WP-Total` headers, so it does not download the full post or media library.

```json
{}
```

The response uses explicit metric names and WordPress status values:

```json
{
  "posts": {
    "total": 120,
    "active": 100,
    "inactive": 20,
    "active_option": {
      "name": "published",
      "value": "publish"
    },
    "inactive_options": [
      { "name": "draft", "value": "draft" },
      { "name": "pending", "value": "pending" },
      { "name": "scheduled", "value": "future" },
      { "name": "private", "value": "private" }
    ]
  },
  "media": { "total": 340 }
}
```

`active` means `status=publish`. `inactive` is every visible non-published post
(`total - active`), so it can also include custom statuses registered by a
WordPress installation. WordPress trash is excluded from these counts. Use
`wordpress_list_posts` to enumerate posts by status; use
`wordpress_list_media` to enumerate media. Media results include
`title.rendered`, `source_url`, `mime_type`, `alt_text`, and dimensions.

## Development

```bash
make fmt            # format Go source
make vet            # run standard static analysis
make test           # run tests with the race detector
make lint           # fail if formatting or vet checks fail
make check          # format, vet, test, and verify modules
make release-cross  # build Linux amd64, macOS arm64, and Windows amd64 archives
make clean          # remove local artifacts
```

The project does not execute remote shells and does not use SSH. Publishing is
performed through the WordPress REST API so the installation's own permissions,
hooks, and cache integration remain in effect.

## Releases

Pushing a `v*` tag runs `.github/workflows/release.yml`. The workflow tests the
module, builds static archives for Linux amd64, macOS arm64, and Windows amd64,
and publishes them as a GitHub Release. The release version is derived from the
Git tag; release notes are curated after the workflow completes.
