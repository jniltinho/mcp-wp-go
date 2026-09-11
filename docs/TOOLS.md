# MCP tool reference

All tools operate on the single WordPress site defined by `WP_BASE_URL`. Tool
inputs never accept an arbitrary base URL.

## Posts

| Tool | Main input | Behavior |
| --- | --- | --- |
| `wordpress_list_posts` | `status?`, `search?`, `page?`, `per_page?` | Lists posts visible to the authenticated user, including drafts |
| `wordpress_get_post` | `id` | Returns editable `content.raw` and rendered fields |
| `wordpress_create_post` | `title`, `content`, `status?`, `excerpt?`, `categories?`, `tags?` | Creates a draft by default and closes comments |
| `wordpress_update_post` | `id` plus fields to replace | Updates content metadata but not slug, date, or status |
| `wordpress_publish_posts` | `ids`, `confirm` | Publishes 1–100 unique posts and reports each result |
| `wordpress_unpublish_posts` | `ids`, `confirm` | Moves 1–100 posts to draft and reports each result |
| `wordpress_delete_post` | `id`, `confirm`, `permanent?` | Trashes by default; permanent deletion is explicit |

Pagination starts at page 1, and `per_page` cannot exceed 100.

## Media and covers

| Tool | Main input | Behavior |
| --- | --- | --- |
| `wordpress_list_media` | `search?`, `media_type?`, `page?`, `per_page?` | Lists Media Library items |
| `wordpress_get_media` | `id` | Returns source URL, type, dimensions, alt text, and association |
| `wordpress_upload_media` | `file_path`, `post_id?`, `alt_text?`, `caption?` | Validates and uploads GIF, JPEG, PNG, or WebP |
| `wordpress_update_media` | `id`, `alt_text?`, `caption?`, `post_id?` | Updates metadata or post association |
| `wordpress_delete_media` | `id`, `confirm`, `permanent?` | Trashes by default; permanent deletion is explicit |
| `wordpress_set_post_cover` | `post_id`, `file_path`, `alt_text`, `replace_existing?` | Uploads and places a relative cover URL first |

`wordpress_set_post_cover` defaults `replace_existing` to true. It never sets
WordPress's featured-image field.

Upload validation checks:

- the file is regular and readable;
- its canonical path is below `WP_UPLOAD_ROOT`, when configured;
- its size is within `WP_MAX_UPLOAD_BYTES`;
- its extension is one of the allowed image formats;
- its detected content type matches the extension.

## Taxonomy and site checks

| Tool | Main input | Behavior |
| --- | --- | --- |
| `wordpress_list_categories` | `search?`, `page?`, `per_page?` | Finds existing categories for reuse |
| `wordpress_list_tags` | `search?`, `page?`, `per_page?` | Finds existing tags for reuse |
| `wordpress_content_stats` | `{}` | Returns published, non-published, and media totals |
| `wordpress_site_health` | `{}` | Validates authentication and returns user identity |
| `wordpress_check_post_live` | `id` | Requires published status plus a public HTTP 2xx response |

`wordpress_content_stats` reads `X-WP-Total` using small one-item collection
requests. Trash is excluded. `inactive` is the visible total minus posts with
status `publish`, so registered custom statuses can be included.

## Confirmation model

The server rejects status changes and deletion unless `confirm` is true.

```json
{ "ids": [123], "confirm": true }
```

```json
{ "id": 123, "confirm": true, "permanent": false }
```

Bulk publication is intentionally not transactional. Always inspect the
per-item results instead of assuming that every ID succeeded.
