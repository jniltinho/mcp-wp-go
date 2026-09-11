# mcp-wp-go

[![CI](https://github.com/jniltinho/mcp-wp-go/actions/workflows/ci.yml/badge.svg)](https://github.com/jniltinho/mcp-wp-go/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/jniltinho/mcp-wp-go?color=blue)](https://github.com/jniltinho/mcp-wp-go/releases)
[![Go](https://img.shields.io/github/go-mod/go-version/jniltinho/mcp-wp-go)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Manage **WordPress posts and media through MCP**, without SSH, and turn Markdown
articles into publication-ready HTML — one small, static Go binary.

```text
AI assistant ──stdio/MCP──▶ mcp-wp-go ──HTTPS/REST──▶ WordPress
Markdown file ──post-html─▶ post HTML fragment ─────▶ post content
Local images ──media tools──────────────────────────▶ Media Library
```

| Capability | What it does |
| --- | --- |
| Posts | List, read, create drafts, edit, publish, unpublish, and delete |
| Media | Upload, inspect, update, associate, and delete images |
| Covers | Upload a cover and place it first without duplicating a featured image |
| Markdown | Convert `.md`/`.markdown` into WordPress-ready HTML |
| Taxonomy | Reuse existing categories and tags |
| Validation | Check credentials, content totals, and whether a post is publicly live |

## Install

Download the package for your platform from the
[latest release](https://github.com/jniltinho/mcp-wp-go/releases/latest):

- Linux: `mcp-wp-go_<version>_linux_amd64.tar.gz`
- macOS: `mcp-wp-go_<version>_darwin_arm64.tar.gz`
- Windows: `mcp-wp-go_<version>_windows_amd64.zip`

Linux example:

```bash
tar -xzf mcp-wp-go_*_linux_amd64.tar.gz
sudo install -m 0755 mcp-wp-go /usr/local/bin/mcp-wp-go
mcp-wp-go --help
```

Or build from source with Go 1.27 or newer:

```bash
git clone https://github.com/jniltinho/mcp-wp-go.git
cd mcp-wp-go
make check && make build
sudo make install
```

## Quick start

### 1. Configure WordPress

Create a dedicated
[WordPress application password](https://developer.wordpress.org/rest-api/using-the-rest-api/authentication/#basic-authentication-with-application-passwords),
then store it in a private file:

```bash
mkdir -p ~/.config
install -m 600 .env.example ~/.config/mcp-wp-go.env
${EDITOR:-vi} ~/.config/mcp-wp-go.env
```

```dotenv
WP_BASE_URL=https://wordpress.example.com
WP_USERNAME=editor-user
WP_APP_PASSWORD=xxxx xxxx xxxx xxxx xxxx xxxx
WP_UPLOAD_ROOT=/home/user/wordpress-media
```

Use an Editor or another account with only the capabilities the workflow needs.
Never use a server password or commit the application password to Git.

### 2. Register the MCP server

Add the binary to the MCP client's stdio configuration using absolute paths:

```json
{
  "mcpServers": {
    "wordpress": {
      "type": "stdio",
      "command": "/usr/local/bin/mcp-wp-go",
      "args": [
        "--env-file",
        "/home/user/.config/mcp-wp-go.env"
      ]
    }
  }
}
```

Reload the MCP client and call `wordpress_site_health`. A successful response
shows the configured site and authenticated WordPress user without exposing the
application password.

### 3. Create safely, then publish deliberately

Ask the MCP client to:

1. list categories and tags;
2. create the post as a draft;
3. upload body images;
4. set the post cover;
5. read the draft back for validation;
6. publish with explicit confirmation;
7. run `wordpress_check_post_live`.

Publishing, unpublishing, and deletion require `confirm: true`. Draft-first is
the recommended workflow.

## Markdown to WordPress HTML

Convert an article to stdout:

```bash
mcp-wp-go post-html article.md
```

Or write it directly to a file:

```bash
mcp-wp-go post-html article.md --output article.html
# aliases: markdown, md    short flag: -o
```

The output is an HTML fragment rather than a complete document, so it can be
sent directly as WordPress post content. Tables, fenced code languages, inline
images, links, blockquotes, and trusted embedded HTML are preserved.

````markdown
## Install

![Application screen](/wp-content/uploads/2026/09/application.webp)

```bash
curl -fsSL https://example.com/install.sh | sh
```

<div class="video-container"><iframe src="https://www.youtube.com/embed/VIDEO_ID" title="Demo" loading="lazy" allowfullscreen></iframe></div>
````

The converter does **not** upload images. Upload them through
`wordpress_upload_media`, use the returned site-relative path, and use
`wordpress_set_post_cover` for the cover. Embedded HTML is not sanitized; only
convert trusted Markdown files.

See the executable pair in [`examples/`](examples/) and the complete
[content workflow](docs/CONTENT-WORKFLOW.md).

## Safety by default

- WordPress requests are restricted to the configured site and use HTTPS;
  plain HTTP is accepted only for loopback development.
- `stdout` is reserved for MCP JSON-RPC; diagnostics go to `stderr`.
- New and updated posts always have comments closed.
- Existing post slugs and dates cannot be changed by the update tool.
- Status changes and deletions require explicit confirmation.
- Uploads are size-limited and checked by extension and detected file type.
- Covers use relative Media Library URLs and never set `featured_media`.
- REST API calls preserve WordPress permissions, hooks, and cache behavior.

## Documentation

| Guide | Contents |
| --- | --- |
| [Getting started](docs/GETTING-STARTED.md) | Installation, credentials, MCP registration, first health check |
| [Usage](docs/USAGE.md) | Post, media, taxonomy, publication, and validation examples |
| [Content workflow](docs/CONTENT-WORKFLOW.md) | Markdown, images, covers, YouTube, and script blocks |
| [Tool reference](docs/TOOLS.md) | Every MCP tool, input, and safety behavior |
| [Architecture](docs/ARCHITECTURE.md) | Package boundaries, data flows, and design decisions |
| [Development](docs/DEVELOPMENT.md) | Tests, linting, cross-builds, CI, and releases |

## Architecture

One binary, two entry paths, and clear package boundaries:

```text
main.go → cmd/ ┬→ MCP server → internal/server → internal/wordpress → WP REST API
               └→ post-html  → internal/posthtml → gomarkdown
```

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full design.

## Development

```bash
make build          # static binary → dist/mcp-wp-go
make test           # tests with the race detector
make lint           # gofmt verification + go vet
make check          # format, lint, test, and module verification
make release-cross  # Linux amd64, macOS arm64, Windows amd64
```

## License

[MIT](LICENSE) © Nilton Oliveira
