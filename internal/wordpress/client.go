package wordpress

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"mcp-wp-go/internal/config"
)

const maxResponseBytes = 8 << 20

var leadingImageParagraph = regexp.MustCompile(`(?is)^(\s*<p[^>]*>\s*<img\b[^>]*>\s*</p>\s*)`)

// Client calls one configured WordPress site. Its base URL is never supplied by tools.
type Client struct {
	baseURL        *url.URL
	username       string
	appPassword    string
	httpClient     *http.Client
	publicClient   *http.Client
	maxUploadBytes int64
	uploadRoot     string
}

// New creates a WordPress REST client from validated configuration.
func New(cfg config.Config) *Client {
	redirectPolicy := func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		if !sameHost(req.URL, cfg.BaseURL) {
			return errors.New("redirect left configured WordPress host")
		}
		return nil
	}
	return &Client{
		baseURL:     cfg.BaseURL,
		username:    cfg.Username,
		appPassword: cfg.AppPassword,
		httpClient: &http.Client{
			Timeout:       cfg.Timeout,
			CheckRedirect: redirectPolicy,
		},
		publicClient: &http.Client{
			Timeout:       cfg.Timeout,
			CheckRedirect: redirectPolicy,
		},
		maxUploadBytes: cfg.MaxUploadBytes,
		uploadRoot:     cfg.UploadRoot,
	}
}

func sameHost(a, b *url.URL) bool {
	return strings.EqualFold(a.Hostname(), b.Hostname()) && effectivePort(a) == effectivePort(b)
}

func effectivePort(u *url.URL) string {
	if p := u.Port(); p != "" {
		return p
	}
	if u.Scheme == "https" {
		return "443"
	}
	return "80"
}

func (c *Client) endpoint(segment string, query url.Values) (*url.URL, error) {
	if strings.Contains(segment, "..") || strings.HasPrefix(segment, "/") {
		return nil, errors.New("invalid API endpoint")
	}
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, "wp-json", "wp", "v2", segment)
	u.RawQuery = query.Encode()
	return &u, nil
}

func (c *Client) doJSON(ctx context.Context, method, segment string, query url.Values, input, output any) error {
	var body io.Reader
	if input != nil {
		payload, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode WordPress request: %w", err)
		}
		body = bytes.NewReader(payload)
	}

	u, err := c.endpoint(segment, query)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return fmt.Errorf("build WordPress request: %w", err)
	}
	req.SetBasicAuth(c.username, c.appPassword)
	req.Header.Set("Accept", "application/json")
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.execute(c.httpClient, req, output)
}

func (c *Client) execute(client *http.Client, req *http.Request, output any) error {
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("WordPress request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read WordPress response: %w", err)
	}
	if len(body) > maxResponseBytes {
		return errors.New("WordPress response exceeds size limit")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var apiError struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(body, &apiError)
		if apiError.Message != "" {
			return fmt.Errorf("WordPress API returned HTTP %d (%s): %s", resp.StatusCode, apiError.Code, apiError.Message)
		}
		return fmt.Errorf("WordPress API returned HTTP %d", resp.StatusCode)
	}
	if output == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, output); err != nil {
		return fmt.Errorf("decode WordPress response: %w", err)
	}
	return nil
}

// ListPosts returns posts visible to the configured editor.
func (c *Client) ListPosts(ctx context.Context, status, search string, page, perPage int) ([]Post, error) {
	query := url.Values{"context": {"edit"}}
	if status != "" {
		query.Set("status", status)
	}
	if search != "" {
		query.Set("search", search)
	}
	if page > 0 {
		query.Set("page", fmt.Sprint(page))
	}
	if perPage > 0 {
		query.Set("per_page", fmt.Sprint(perPage))
	}
	var posts []Post
	if err := c.doJSON(ctx, http.MethodGet, "posts", query, nil, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPost gets raw editable content; it is not a public-only endpoint.
func (c *Client) GetPost(ctx context.Context, id int) (Post, error) {
	var post Post
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("posts/%d", id), url.Values{"context": {"edit"}}, nil, &post)
	return post, err
}

// CreatePost always closes comments. The caller controls only draft or publish status.
func (c *Client) CreatePost(ctx context.Context, title, content, status, excerpt string, categories, tags []int) (Post, error) {
	if title = strings.TrimSpace(title); title == "" {
		return Post{}, errors.New("post title is required")
	}
	if status == "" {
		status = "draft"
	}
	if status != "draft" && status != "publish" {
		return Post{}, errors.New("new post status must be draft or publish")
	}
	input := map[string]any{
		"title":          title,
		"content":        content,
		"status":         status,
		"comment_status": "closed",
	}
	if excerpt != "" {
		input["excerpt"] = excerpt
	}
	if categories != nil {
		input["categories"] = categories
	}
	if tags != nil {
		input["tags"] = tags
	}
	var post Post
	err := c.doJSON(ctx, http.MethodPost, "posts", nil, input, &post)
	return post, err
}

// UpdatePost deliberately excludes slug, date and status to protect indexed URLs.
func (c *Client) UpdatePost(ctx context.Context, id int, title, content, excerpt *string, categories, tags *[]int) (Post, error) {
	input := make(map[string]any)
	if title != nil {
		if strings.TrimSpace(*title) == "" {
			return Post{}, errors.New("post title cannot be empty")
		}
		input["title"] = *title
	}
	if content != nil {
		input["content"] = *content
	}
	if excerpt != nil {
		input["excerpt"] = *excerpt
	}
	if categories != nil {
		input["categories"] = *categories
	}
	if tags != nil {
		input["tags"] = *tags
	}
	if len(input) == 0 {
		return Post{}, errors.New("supply at least one editable field")
	}
	// Keep the site-wide no-comments policy true for every edited post.
	input["comment_status"] = "closed"
	var post Post
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("posts/%d", id), nil, input, &post)
	return post, err
}

// SetPostStatus changes only a post's publication status. It does not change its
// slug, date, content, taxonomy, or comment configuration.
func (c *Client) SetPostStatus(ctx context.Context, id int, status string) (Post, error) {
	if status != "publish" && status != "draft" {
		return Post{}, errors.New("post status must be publish or draft")
	}
	var post Post
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("posts/%d", id), nil, map[string]string{"status": status}, &post)
	return post, err
}

// DeletePost moves a post to trash unless permanent is true.
func (c *Client) DeletePost(ctx context.Context, id int, permanent bool) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("posts/%d", id), url.Values{"force": {fmt.Sprint(permanent)}}, nil, nil)
}

// ListMedia returns media visible to the configured editor.
func (c *Client) ListMedia(ctx context.Context, search, mediaType string, page, perPage int) ([]Media, error) {
	query := url.Values{"context": {"edit"}}
	if search != "" {
		query.Set("search", search)
	}
	if mediaType != "" {
		query.Set("media_type", mediaType)
	}
	if page > 0 {
		query.Set("page", fmt.Sprint(page))
	}
	if perPage > 0 {
		query.Set("per_page", fmt.Sprint(perPage))
	}
	var media []Media
	if err := c.doJSON(ctx, http.MethodGet, "media", query, nil, &media); err != nil {
		return nil, err
	}
	return media, nil
}

// GetMedia returns editable metadata for a media item.
func (c *Client) GetMedia(ctx context.Context, id int) (Media, error) {
	var media Media
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("media/%d", id), url.Values{"context": {"edit"}}, nil, &media)
	return media, err
}

var allowedImages = map[string]string{
	".gif":  "image/gif",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

func (c *Client) uploadFile(pathname string) (string, string, *os.File, error) {
	resolved, err := filepath.EvalSymlinks(pathname)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve upload file: %w", err)
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", "", nil, fmt.Errorf("make upload path absolute: %w", err)
	}
	if c.uploadRoot != "" {
		relative, err := filepath.Rel(c.uploadRoot, resolved)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", "", nil, errors.New("upload file is outside WP_UPLOAD_ROOT")
		}
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", "", nil, fmt.Errorf("stat upload file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", "", nil, errors.New("upload file must be a regular file")
	}
	if info.Size() > c.maxUploadBytes {
		return "", "", nil, fmt.Errorf("upload exceeds configured limit of %d bytes", c.maxUploadBytes)
	}
	ext := strings.ToLower(filepath.Ext(resolved))
	expected, allowed := allowedImages[ext]
	if !allowed {
		return "", "", nil, errors.New("only GIF, JPEG, PNG and WebP image uploads are allowed")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return "", "", nil, fmt.Errorf("open upload file: %w", err)
	}
	var header [512]byte
	n, readErr := file.Read(header[:])
	if readErr != nil && readErr != io.EOF {
		file.Close()
		return "", "", nil, fmt.Errorf("read upload file: %w", readErr)
	}
	detected := http.DetectContentType(header[:n])
	if detected != expected {
		file.Close()
		return "", "", nil, fmt.Errorf("file type %q does not match %s extension", detected, ext)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		return "", "", nil, fmt.Errorf("rewind upload file: %w", err)
	}
	return filepath.Base(resolved), expected, file, nil
}

// UploadMedia uploads an image and then sets its optional metadata with the REST API.
func (c *Client) UploadMedia(ctx context.Context, pathname string, postID int, altText, caption string) (Media, error) {
	filename, contentType, file, err := c.uploadFile(pathname)
	if err != nil {
		return Media{}, err
	}
	defer file.Close()

	u, err := c.endpoint("media", nil)
	if err != nil {
		return Media{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), file)
	if err != nil {
		return Media{}, fmt.Errorf("build media upload request: %w", err)
	}
	req.SetBasicAuth(c.username, c.appPassword)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	var media Media
	if err := c.execute(c.httpClient, req, &media); err != nil {
		return Media{}, err
	}
	if postID != 0 || altText != "" || caption != "" {
		input := make(map[string]any)
		if postID != 0 {
			input["post"] = postID
		}
		if altText != "" {
			input["alt_text"] = altText
		}
		if caption != "" {
			input["caption"] = caption
		}
		if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("media/%d", media.ID), nil, input, &media); err != nil {
			return Media{}, fmt.Errorf("media uploaded as ID %d but metadata update failed: %w", media.ID, err)
		}
	}
	return media, nil
}

// UpdateMedia changes safe metadata only; it does not replace the binary.
func (c *Client) UpdateMedia(ctx context.Context, id int, altText, caption *string, postID *int) (Media, error) {
	input := make(map[string]any)
	if altText != nil {
		input["alt_text"] = *altText
	}
	if caption != nil {
		input["caption"] = *caption
	}
	if postID != nil {
		input["post"] = *postID
	}
	if len(input) == 0 {
		return Media{}, errors.New("supply alt_text, caption or post_id")
	}
	var media Media
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("media/%d", id), nil, input, &media)
	return media, err
}

// DeleteMedia moves a media item to trash unless permanent is true.
func (c *Client) DeleteMedia(ctx context.Context, id int, permanent bool) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("media/%d", id), url.Values{"force": {fmt.Sprint(permanent)}}, nil, nil)
}

// SetPostCover uploads an image and puts it first in the post body. It never sets
// WordPress's featured-image field because some themes render it twice.
func (c *Client) SetPostCover(ctx context.Context, postID int, pathname, altText string, replaceExisting bool) (Post, Media, error) {
	post, err := c.GetPost(ctx, postID)
	if err != nil {
		return Post{}, Media{}, err
	}
	media, err := c.UploadMedia(ctx, pathname, postID, altText, "")
	if err != nil {
		return Post{}, Media{}, err
	}
	src, err := c.relativeMediaURL(media.SourceURL)
	if err != nil {
		return Post{}, media, err
	}
	image := fmt.Sprintf(`<p><img src="%s" alt="%s"`, html.EscapeString(src), html.EscapeString(altText))
	if media.MediaDetails.Width > 0 {
		image += fmt.Sprintf(` width="%d"`, media.MediaDetails.Width)
	}
	if media.MediaDetails.Height > 0 {
		image += fmt.Sprintf(` height="%d"`, media.MediaDetails.Height)
	}
	image += ` /></p>`
	content := post.Content.Raw
	if replaceExisting && leadingImageParagraph.MatchString(content) {
		content = leadingImageParagraph.ReplaceAllString(content, image+"\n")
	} else {
		content = image + "\n" + content
	}
	updated, err := c.UpdatePost(ctx, postID, nil, &content, nil, nil, nil)
	if err != nil {
		return Post{}, media, err
	}
	return updated, media, nil
}

func (c *Client) relativeMediaURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("WordPress returned an invalid media URL")
	}
	if !sameHost(u, c.baseURL) {
		return "", errors.New("WordPress returned media hosted outside configured site")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", errors.New("WordPress returned media URL with query or fragment")
	}
	return u.EscapedPath(), nil
}

// ListTerms lists either categories or tags.
func (c *Client) ListTerms(ctx context.Context, kind, search string, page, perPage int) ([]Term, error) {
	if kind != "categories" && kind != "tags" {
		return nil, errors.New("term kind must be categories or tags")
	}
	query := url.Values{"context": {"edit"}}
	if search != "" {
		query.Set("search", search)
	}
	if page > 0 {
		query.Set("page", fmt.Sprint(page))
	}
	if perPage > 0 {
		query.Set("per_page", fmt.Sprint(perPage))
	}
	var terms []Term
	if err := c.doJSON(ctx, http.MethodGet, kind, query, nil, &terms); err != nil {
		return nil, err
	}
	return terms, nil
}

// CurrentUser validates the configured application password and exposes no secret.
func (c *Client) CurrentUser(ctx context.Context) (User, error) {
	var user User
	err := c.doJSON(ctx, http.MethodGet, "users/me", url.Values{"context": {"edit"}}, nil, &user)
	return user, err
}

// CheckPostLive confirms that a post is published and that its public permalink
// returns a successful HTTP response from the same configured host.
func (c *Client) CheckPostLive(ctx context.Context, id int) (Post, int, bool, error) {
	post, err := c.GetPost(ctx, id)
	if err != nil {
		return Post{}, 0, false, err
	}
	link, err := url.Parse(post.Link)
	if err != nil || !sameHost(link, c.baseURL) {
		return post, 0, false, errors.New("WordPress returned a permalink outside configured site")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link.String(), nil)
	if err != nil {
		return post, 0, false, fmt.Errorf("build permalink check: %w", err)
	}
	resp, err := c.publicClient.Do(req)
	if err != nil {
		return post, 0, false, fmt.Errorf("public permalink check failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	live := post.Status == "publish" && resp.StatusCode >= 200 && resp.StatusCode < 300
	return post, resp.StatusCode, live, nil
}
