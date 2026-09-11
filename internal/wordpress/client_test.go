package wordpress

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mcp-wp-go/internal/config"
)

func testClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	site := httptest.NewServer(handler)
	u, err := url.Parse(site.URL)
	if err != nil {
		t.Fatal(err)
	}
	return New(config.Config{
		BaseURL:        u,
		Username:       "editor",
		AppPassword:    "not-a-real-secret",
		Timeout:        time.Second,
		MaxUploadBytes: 1 << 20,
	}), site
}

func TestCreatePostClosesComments(t *testing.T) {
	client, site := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wp-json/wp/v2/posts" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if got := body["comment_status"]; got != "closed" {
			t.Fatalf("comment_status = %#v", got)
		}
		if got := body["status"]; got != "draft" {
			t.Fatalf("status = %#v", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"status":"draft","title":{"rendered":"Test"}}`))
	}))
	defer site.Close()

	post, err := client.CreatePost(context.Background(), "Test", "<p>Conteúdo</p>", "", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if post.ID != 42 {
		t.Fatalf("post ID = %d", post.ID)
	}
}

func TestSetPostCoverUsesRelativeURLAndReplacesLeadingImage(t *testing.T) {
	var site *httptest.Server
	var updateContent string
	client, localSite := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/wp-json/wp/v2/posts/7":
			if r.URL.Query().Get("context") != "edit" {
				t.Fatalf("post context = %q", r.URL.Query().Get("context"))
			}
			_, _ = w.Write([]byte(`{"id":7,"status":"draft","content":{"raw":"<p><img src=\"/wp-content/uploads/old.webp\" alt=\"old\" /></p>\n<h2>Texto</h2>"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/media":
			if got := r.Header.Get("Content-Type"); got != "image/webp" {
				t.Fatalf("content type = %q", got)
			}
			if !strings.Contains(r.Header.Get("Content-Disposition"), "cover.webp") {
				t.Fatalf("content disposition = %q", r.Header.Get("Content-Disposition"))
			}
			_, _ = io.Copy(io.Discard, r.Body)
			_, _ = w.Write([]byte(`{"id":88,"source_url":"` + site.URL + `/wp-content/uploads/2026/09/cover.webp","media_details":{"width":1440,"height":810}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/media/88":
			_, _ = w.Write([]byte(`{"id":88,"source_url":"` + site.URL + `/wp-content/uploads/2026/09/cover.webp","media_details":{"width":1440,"height":810}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/wp-json/wp/v2/posts/7":
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			updateContent = payload["content"]
			_, _ = w.Write([]byte(`{"id":7,"status":"draft"}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL)
		}
	}))
	site = localSite
	defer site.Close()

	image := filepath.Join(t.TempDir(), "cover.webp")
	if err := os.WriteFile(image, []byte("RIFF\x24\x00\x00\x00WEBPVP8 \x18\x00\x00\x00"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, media, err := client.SetPostCover(context.Background(), 7, image, "Test cover", true)
	if err != nil {
		t.Fatal(err)
	}
	if media.ID != 88 {
		t.Fatalf("media ID = %d", media.ID)
	}
	if !strings.Contains(updateContent, `src="/wp-content/uploads/2026/09/cover.webp"`) {
		t.Fatalf("cover URL was not relative: %s", updateContent)
	}
	if strings.Contains(updateContent, "old.webp") {
		t.Fatalf("old cover remains: %s", updateContent)
	}
	if !strings.Contains(updateContent, "<h2>Texto</h2>") {
		t.Fatalf("body content was lost: %s", updateContent)
	}
}

func TestCheckPostLive(t *testing.T) {
	var site *httptest.Server
	site = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/wp-json/wp/v2/posts/9":
			_, _ = w.Write([]byte(`{"id":9,"status":"publish","link":"` + site.URL + `/2026/09/post/"}`))
		case "/2026/09/post/":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer site.Close()
	u, err := url.Parse(site.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := New(config.Config{BaseURL: u, Username: "editor", AppPassword: "password", Timeout: time.Second, MaxUploadBytes: 1 << 20})
	post, status, live, err := client.CheckPostLive(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if post.ID != 9 || status != http.StatusOK || !live {
		t.Fatalf("CheckPostLive() = post=%d status=%d live=%t", post.ID, status, live)
	}
}
