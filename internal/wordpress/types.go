// Package wordpress implements the small, safety-oriented subset of the
// WordPress REST API exposed by this MCP server.
package wordpress

// RenderedField is how WordPress returns rendered title and content values.
type RenderedField struct {
	Rendered string `json:"rendered"`
	Raw      string `json:"raw,omitempty"`
}

// Post is the response shape used by the post tools.
type Post struct {
	ID            int           `json:"id"`
	FeaturedMedia int           `json:"featured_media"`
	Date          string        `json:"date"`
	Modified      string        `json:"modified"`
	Slug          string        `json:"slug"`
	Status        string        `json:"status"`
	Link          string        `json:"link"`
	Title         RenderedField `json:"title"`
	Content       RenderedField `json:"content"`
	Excerpt       RenderedField `json:"excerpt"`
	Categories    []int         `json:"categories"`
	Tags          []int         `json:"tags"`
	CommentStatus string        `json:"comment_status"`
}

// MediaDetails has the dimensions needed to write a body cover image.
type MediaDetails struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Media is the response shape used by media tools.
type Media struct {
	ID           int           `json:"id"`
	Date         string        `json:"date"`
	Slug         string        `json:"slug"`
	Link         string        `json:"link"`
	Title        RenderedField `json:"title"`
	SourceURL    string        `json:"source_url"`
	AltText      string        `json:"alt_text"`
	Caption      RenderedField `json:"caption"`
	Description  RenderedField `json:"description"`
	MediaType    string        `json:"media_type"`
	MimeType     string        `json:"mime_type"`
	MediaDetails MediaDetails  `json:"media_details"`
	Post         int           `json:"post"`
}

// Term is shared by the category and tag endpoints.
type Term struct {
	ID          int    `json:"id"`
	Count       int    `json:"count"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
}

// User identifies the account associated with the configured application password.
type User struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Slug     string   `json:"slug"`
	Roles    []string `json:"roles"`
	Username string   `json:"username"`
}
