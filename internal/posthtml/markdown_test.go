package posthtml

import (
	"os"
	"strings"
	"testing"
)

func TestFromMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		source   string
		expected []string
	}{
		{
			name:   "post structure",
			source: "## Install\n\nUse **Linux**.\n",
			expected: []string{
				`<h2 id="install">Install</h2>`,
				`<p>Use <strong>Linux</strong>.</p>`,
			},
		},
		{
			name:   "fenced code language",
			source: "```bash\necho '<ok>'\n```\n",
			expected: []string{
				`<pre><code class="language-bash">`,
				`echo '&lt;ok&gt;'`,
			},
		},
		{
			name:   "table",
			source: "Name | Value\n--- | ---\nLinux | yes\n",
			expected: []string{
				"<table>",
				"<th>Name</th>",
				"<td>Linux</td>",
			},
		},
		{
			name:   "embedded wordpress html",
			source: `<div class="notice">Keep this block.</div>`,
			expected: []string{
				`<div class="notice">Keep this block.</div>`,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			converted := string(FromMarkdown([]byte(test.source)))
			for _, expected := range test.expected {
				if !strings.Contains(converted, expected) {
					t.Errorf("FromMarkdown() = %q, expected it to contain %q", converted, expected)
				}
			}
			if strings.Contains(converted, "<html") || strings.Contains(converted, "<body") {
				t.Errorf("FromMarkdown() returned a complete document: %q", converted)
			}
		})
	}
}

func TestFromMarkdownMatchesWordPressGoldenFile(t *testing.T) {
	t.Parallel()

	markdownSource, err := os.ReadFile("testdata/post.md")
	if err != nil {
		t.Fatalf("ReadFile(markdown): %v", err)
	}
	expectedHTML, err := os.ReadFile("testdata/post.html")
	if err != nil {
		t.Fatalf("ReadFile(html): %v", err)
	}

	actualHTML := FromMarkdown(markdownSource)
	if string(actualHTML) != string(expectedHTML) {
		t.Fatalf("FromMarkdown() mismatch\nactual:\n%s\nexpected:\n%s", actualHTML, expectedHTML)
	}
}

func TestFromMarkdownSupportsWordPressMediaAndScripts(t *testing.T) {
	t.Parallel()

	markdownSource, err := os.ReadFile("testdata/post.md")
	if err != nil {
		t.Fatalf("ReadFile(markdown): %v", err)
	}

	converted := string(FromMarkdown(markdownSource))
	if !strings.HasPrefix(converted, `<p><img src="/wp-content/uploads/2026/09/capa-qa.webp"`) {
		t.Fatalf("cover is not the first HTML element: %q", converted)
	}

	expected := []string{
		`<img src="/wp-content/uploads/2026/09/tela-qa.webp" alt="Captura do aplicativo"`,
		`<div class="video-container"><iframe src="https://www.youtube.com/embed/jNQXAC9IVRw"`,
		`<code class="language-bash">`,
		`<code class="language-yaml">porta: &quot;{{ app_port }}&quot;`,
		`<code class="language-javascript">`,
		`&lt;script seguro&gt;`,
	}
	for _, fragment := range expected {
		if !strings.Contains(converted, fragment) {
			t.Errorf("FromMarkdown() = %q, expected it to contain %q", converted, fragment)
		}
	}
}
