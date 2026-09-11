# Getting started

This guide connects one WordPress site to an MCP client over a local stdio
process. The MCP server talks to WordPress through the REST API over HTTPS; it
does not need SSH or direct database access.

## Requirements

- Linux amd64, macOS arm64, or Windows amd64; or Go 1.27+ to build from source.
- A WordPress account with the capabilities needed for the intended operations.
- A WordPress application password created for that account.
- An MCP client that can launch a stdio server.

An Editor account is a practical default for content management. Use a more
restricted role when the site workflow allows it.

## Install a release

Download the archive for your operating system from
[GitHub Releases](https://github.com/jniltinho/mcp-wp-go/releases/latest).

### Linux amd64

```bash
tar -xzf mcp-wp-go_*_linux_amd64.tar.gz
sudo install -m 0755 mcp-wp-go /usr/local/bin/mcp-wp-go
```

### macOS arm64

```bash
tar -xzf mcp-wp-go_*_darwin_arm64.tar.gz
sudo install -m 0755 mcp-wp-go /usr/local/bin/mcp-wp-go
```

### Windows amd64

Extract `mcp-wp-go.exe` from the ZIP archive and place it in a directory listed
in `PATH`.

## Build from source

```bash
git clone https://github.com/jniltinho/mcp-wp-go.git
cd mcp-wp-go
make check
make build
```

The binary is written to `dist/mcp-wp-go`.

## Create an application password

In WordPress, open the profile page for the chosen user, create a new
application password, and copy it once. This credential is separate from the
normal interactive login password and can be revoked independently.

Create a private configuration file:

```bash
mkdir -p ~/.config
install -m 600 .env.example ~/.config/mcp-wp-go.env
${EDITOR:-vi} ~/.config/mcp-wp-go.env
```

Required values:

```dotenv
WP_BASE_URL=https://wordpress.example.com
WP_USERNAME=editor-user
WP_APP_PASSWORD=xxxx xxxx xxxx xxxx xxxx xxxx
```

Recommended upload restriction:

```dotenv
WP_UPLOAD_ROOT=/home/user/wordpress-media
```

Only files below that directory can then be read by media upload tools.

## Configuration reference

| Variable | Required | Default | Purpose |
| --- | ---: | --- | --- |
| `WP_BASE_URL` | yes | — | Canonical WordPress site URL |
| `WP_USERNAME` | yes | — | User that owns the application password |
| `WP_APP_PASSWORD` | yes | — | WordPress application password |
| `WP_TIMEOUT` | no | `30s` | Timeout for REST and public validation requests |
| `WP_MAX_UPLOAD_BYTES` | no | `26214400` | Maximum accepted upload size in bytes |
| `WP_UPLOAD_ROOT` | no | unrestricted | Filesystem root allowed for uploads |

Values already present in the process environment take precedence over values
from `--env-file`. The file parser accepts simple `KEY=VALUE` lines with
optional single or double quotes; it never executes shell expressions.

## Register the stdio server

Use absolute paths in the MCP client configuration:

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

Restart or reload the MCP client after changing its configuration.

## Validate the connection

Call `wordpress_site_health` with an empty input object. It validates the REST
endpoint and authentication and returns:

- the configured base URL;
- the application user's ID, name, username, slug, and roles.

It never returns the application password.

Next: [Usage](USAGE.md) or the complete
[content workflow](CONTENT-WORKFLOW.md).
