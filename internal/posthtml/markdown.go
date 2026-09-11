// Package posthtml converts authoring formats into WordPress post HTML fragments.
package posthtml

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// FromMarkdown converts Markdown into an HTML fragment suitable for the
// WordPress post content field. It deliberately does not add html or body tags.
func FromMarkdown(source []byte) []byte {
	extensions := parser.CommonExtensions |
		parser.AutoHeadingIDs |
		parser.NoEmptyLineBeforeBlock
	markdownParser := parser.NewWithExtensions(extensions)
	document := markdownParser.Parse(source)

	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags,
	})

	return markdown.Render(document, renderer)
}
