# Content workflow

This is the recommended end-to-end flow for a post authored in Markdown with
body images, a cover, a YouTube video, and runnable scripts.

## 1. Prepare media

Store generated or downloaded images below `WP_UPLOAD_ROOT`, preferably as
optimized WebP files. Upload each body image with `wordpress_upload_media` and
record the returned `source_url`.

Use the path portion in Markdown:

```markdown
![Application running on Linux](/wp-content/uploads/2026/09/application.webp)
```

The `post-html` command creates the `<img>` element but does not upload files.
Using a Media Library URL avoids hotlinks and keeps WordPress metadata attached
to the asset.

## 2. Write the article

Use normal Markdown for headings, paragraphs, lists, links, blockquotes, tables,
images, and fenced code:

````markdown
## Install

Download the package for your operating system.

```bash
set -euo pipefail
curl -fsSL https://example.com/package.tar.gz -o package.tar.gz
tar -xzf package.tar.gz
```

```yaml
service:
  port: "{{ app_port }}"
```
````

Language names become classes such as `language-bash` and `language-yaml`.
Characters with HTML meaning are escaped inside code blocks, while template
expressions such as `{{ app_port }}` remain intact.

## 3. Embed YouTube responsively

Trusted inline HTML is preserved. A responsive embed can therefore be authored
directly in the Markdown file:

```html
<div class="video-container"><iframe src="https://www.youtube.com/embed/VIDEO_ID" title="Video title" loading="lazy" allowfullscreen></iframe></div>
```

Whether `.video-container` is responsive depends on the site's theme CSS. The
converter does not add styling.

## 4. Convert to HTML

```bash
mcp-wp-go post-html article.md --output article.html
```

Review the result before publication:

```bash
grep -E '<h2|<img|video-container|language-' article.html
```

The result intentionally has no `<html>` or `<body>` wrapper.

## 5. Create the draft

Send the complete generated HTML to `wordpress_create_post` with
`"status": "draft"`, existing category IDs, and existing tag IDs. The tool
closes comments automatically.

Read the new post back with `wordpress_get_post` and compare both:

- `content.raw`: editable HTML stored by WordPress;
- `content.rendered`: HTML after WordPress filters.

Confirm that the expected images, iframe, and code classes survived.

## 6. Set the cover

Call `wordpress_set_post_cover` after the draft exists:

```json
{
  "post_id": 123,
  "file_path": "/home/user/wordpress-media/article-cover.webp",
  "alt_text": "Descriptive cover alternative text"
}
```

The tool uploads the file, inserts it as the first image paragraph with a
site-relative URL, and does not set `featured_media`. If the converted Markdown
already starts with a cover placeholder, the default replacement behavior
swaps that paragraph for the real Media Library image.

## 7. Validate and publish

Before changing status:

1. retrieve the draft again;
2. confirm comments are closed;
3. confirm the cover is first and uses a relative URL;
4. confirm body images belong to the configured site;
5. confirm YouTube and code blocks appear in `content.rendered`;
6. verify title, excerpt, categories, and tags.

Publish only with explicit confirmation:

```json
{ "ids": [123], "confirm": true }
```

Then call `wordpress_check_post_live` and require both `live: true` and a public
HTTP 2xx status.

## What conversion does not do

- It does not upload local or remote images.
- It does not fetch YouTube metadata or validate a video ID.
- It does not generate theme CSS for responsive embeds.
- It does not parse front matter into title, excerpt, categories, or tags.
- It does not emit Gutenberg block comments.
- It does not sanitize embedded HTML.

Only convert trusted Markdown. Use the MCP tools for media, taxonomy, status,
and live-site validation.

See [`examples/post-completo.md`](../examples/post-completo.md) and its exact
[`post-completo.html`](../examples/post-completo.html) output.
