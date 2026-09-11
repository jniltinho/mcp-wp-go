# Markdown conversion examples

`post-completo.md` demonstrates the content elements normally used in a
WordPress post: a leading cover placeholder, an inline media-library image, a
responsive YouTube iframe, and fenced Bash, YAML, and JavaScript blocks.

Generate and compare its HTML with:

```bash
dist/mcp-wp-go post-html examples/post-completo.md -o /tmp/post-completo.html
diff -u examples/post-completo.html /tmp/post-completo.html
```

The `/wp-content/uploads/...` paths are examples. Upload the real images first
with `wordpress_upload_media` and replace the paths with the relative URLs
returned by WordPress. When creating a post through MCP, prefer
`wordpress_set_post_cover` for the first image: it uploads the real cover and
replaces the initial placeholder without setting a featured image.
