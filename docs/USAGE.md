# Usage

Once the stdio server is registered, the MCP client discovers the tools and
their JSON schemas automatically. The snippets below show each tool's
`arguments` object rather than a complete JSON-RPC envelope.

## Inspect the site first

Validate credentials with `wordpress_site_health`:

```json
{}
```

Inspect content totals with `wordpress_content_stats`:

```json
{}
```

The result separates published posts from visible non-published posts and
reports the Media Library total.

## Find existing content

Search all visible posts with `wordpress_list_posts`:

```json
{
  "search": "OpenChatCut",
  "page": 1,
  "per_page": 20
}
```

Add `"status": "draft"` or `"status": "publish"` to narrow the result.

Read editable HTML with `wordpress_get_post`:

```json
{ "id": 123 }
```

Use `content.raw` when preparing an update; `content.rendered` shows the output
after WordPress filters.

## Reuse taxonomy

List or search categories with `wordpress_list_categories`:

```json
{ "search": "Linux", "per_page": 50 }
```

List or search tags with `wordpress_list_tags`:

```json
{ "search": "opensource", "per_page": 50 }
```

The create and update tools accept existing numeric IDs; they do not create new
taxonomy terms.

## Create a draft

Call `wordpress_create_post`:

```json
{
  "title": "Running an open source editor on Linux",
  "content": "<h2>Install</h2>\n<p>Download the package.</p>",
  "status": "draft",
  "excerpt": "A practical installation guide.",
  "categories": [2],
  "tags": [42]
}
```

`status` defaults to `draft`, and comments are always closed.

## Update the draft

Call `wordpress_update_post`:

```json
{
  "id": 123,
  "title": "Updated title",
  "content": "<h2>Install</h2>\n<p>Updated complete body.</p>",
  "categories": [2, 25],
  "tags": [42, 87]
}
```

Supplied fields replace their existing values; `content` is a complete body
replacement. The tool cannot change slug, date, or status.

## Upload an image

Call `wordpress_upload_media`:

```json
{
  "file_path": "/home/user/wordpress-media/screenshot.webp",
  "post_id": 123,
  "alt_text": "Application settings screen",
  "caption": "Settings used in this guide"
}
```

Supported formats are GIF, JPEG, PNG, and WebP. The extension, detected media
type, allowed root, and maximum size are checked.

The response contains `source_url`. When writing post HTML, keep only its
site-relative path, such as `/wp-content/uploads/2026/09/screenshot.webp`.

## Set the cover

Call `wordpress_set_post_cover`:

```json
{
  "post_id": 123,
  "file_path": "/home/user/wordpress-media/cover.webp",
  "alt_text": "Open source video editor running on Linux"
}
```

It uploads the cover and inserts a relative image URL as the first body
element. If the content already starts with an image paragraph, it is replaced
by default. Set `"replace_existing": false` to prepend instead.

The tool deliberately leaves `featured_media` unset to prevent themes from
rendering the cover twice.

## Publish and verify

Call `wordpress_publish_posts` as a separate confirmed operation:

```json
{
  "ids": [123],
  "confirm": true
}
```

Up to 100 unique IDs are accepted. Results are reported individually because a
bulk status change is not transactional.

Verify the public permalink with `wordpress_check_post_live`:

```json
{ "id": 123 }
```

`live` is true only when the authenticated post status is `publish` and the
public permalink returns HTTP 2xx.

## Unpublish or delete

Move published posts back to drafts with `wordpress_unpublish_posts`:

```json
{ "ids": [123], "confirm": true }
```

Move a post to trash with `wordpress_delete_post`:

```json
{ "id": 123, "confirm": true }
```

Permanent deletion additionally requires `"permanent": true`. Media deletion
follows the same confirmation model.

For the full draft-to-live sequence, see
[Content workflow](CONTENT-WORKFLOW.md).
