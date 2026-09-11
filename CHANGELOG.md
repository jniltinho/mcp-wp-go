# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Documentation
- Add a structured documentation hub with getting-started, usage, content workflow, MCP tool reference, architecture, and development guides.
- Redesign the repository README with release badges, a quick-start path, capability overview, safety summary, and links to deeper documentation.

## [0.4.0] — 2026-09-11

**Markdown authoring is now a first-class CLI workflow for WordPress posts.**

### Added
- Cobra command structure while preserving the no-subcommand stdio MCP server behavior.
- `post-html` command (aliases: `markdown`, `md`) to convert trusted Markdown files into WordPress-ready HTML fragments using `gomarkdown/markdown`.
- `--output`/`-o` support for writing converted HTML to a file; stdout remains the default.
- Golden fixtures and a complete `examples/` pair covering a leading cover, media-library images, a responsive YouTube embed, and Bash, YAML, and JavaScript code blocks.

### Changed
- Document the media workflow: upload body images separately and use `wordpress_set_post_cover` to insert the cover without setting a duplicate featured image.
- Resolve existing static-analysis findings around URL validation and read-only resource cleanup.

## [0.3.0] — 2026-09-11

### Added
- `wordpress_content_stats` for total posts, active published posts, inactive non-published posts, and total media.
- Explicit names and WordPress status values for `publish`, `draft`, `pending`, `future`, and `private` in the statistics response.
- Media titles in `wordpress_list_media` results, alongside the existing URL, MIME type, alt text, and dimensions.

### Notes
- Statistics are read from `X-WP-Total` pagination headers using one item per request; WordPress trash is excluded.

## [0.2.0] — 2026-09-11

### Added
- Publish and unpublish one or up to 100 WordPress posts in a confirmed bulk operation.
- Per-post status results that make partial failures visible to MCP clients.

### Safety
- Require explicit confirmation and preserve post content, slug, date, taxonomy, and comments during status-only changes.

## [0.1.4] — 2026-09-11

**Release process aligned with the repository release standard.**

### Changed
- Build release archives for Linux amd64, macOS arm64, and Windows amd64.
- Derive build version metadata from the Git tag instead of a release version constant.
- Add a maintained changelog and prepare curated GitHub Release notes.

## [0.1.3] — 2026-09-11

### Added
- Include `featured_media` in post responses for automated cover validation.

### Fixed
- Preserve and validate media library alternative text for post covers.

## [0.1.2] — 2026-09-11

### Changed
- Move the executable entry point to the repository root to match the project layout standard.

## [0.1.1] — 2026-09-11

### Changed
- Generalize project branding, configuration examples, repository metadata, and documentation.

## [0.1.0] — 2026-09-11

### Added
- Initial WordPress REST MCP server for posts, media, covers, taxonomy, and publication checks.
- GitHub Actions CI and release automation.
