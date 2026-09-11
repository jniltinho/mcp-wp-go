// Package server exposes the WordPress REST client as a stdio MCP server.
package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-wp-go/internal/config"
	"mcp-wp-go/internal/wordpress"
)

// Version is set from the Git tag by the release build.
var Version = "dev"

// New creates an MCP server with WordPress post, media and verification tools.
func New(cfg config.Config) *mcp.Server {
	wp := wordpress.New(cfg)
	server := mcp.NewServer(&mcp.Implementation{
		Name:        "mcp-wp-go",
		Title:       "WordPress REST MCP",
		Description: "Manages WordPress posts and images through the REST API without SSH.",
		Version:     Version,
	}, nil)

	addPostTools(server, wp)
	addMediaTools(server, wp)
	addSiteTools(server, wp, cfg.BaseURL.String())
	return server
}

type listPostsInput struct {
	Status  string `json:"status,omitempty" jsonschema:"WordPress status, for example publish, draft, or pending."`
	Search  string `json:"search,omitempty" jsonschema:"Text search for posts."`
	Page    int    `json:"page,omitempty" jsonschema:"Page number, starting at 1."`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Items per page, maximum 100."`
}

type postsOutput struct {
	Posts []wordpress.Post `json:"posts"`
}

type postIDInput struct {
	ID int `json:"id" jsonschema:"Numeric WordPress post ID."`
}

type postOutput struct {
	Post wordpress.Post `json:"post"`
}

type createPostInput struct {
	Title      string `json:"title" jsonschema:"Title for the new post."`
	Content    string `json:"content" jsonschema:"Post HTML body. For covers, the image must be the first element."`
	Status     string `json:"status,omitempty" jsonschema:"draft (default) or publish."`
	Excerpt    string `json:"excerpt,omitempty" jsonschema:"Optional excerpt."`
	Categories []int  `json:"categories,omitempty" jsonschema:"Existing category IDs."`
	Tags       []int  `json:"tags,omitempty" jsonschema:"Existing tag IDs."`
}

type updatePostInput struct {
	ID         int     `json:"id" jsonschema:"Numeric post ID."`
	Title      *string `json:"title,omitempty" jsonschema:"New title. This tool cannot modify slug, date, or status."`
	Content    *string `json:"content,omitempty" jsonschema:"Complete replacement HTML body."`
	Excerpt    *string `json:"excerpt,omitempty" jsonschema:"New excerpt."`
	Categories *[]int  `json:"categories,omitempty" jsonschema:"Replaces category IDs."`
	Tags       *[]int  `json:"tags,omitempty" jsonschema:"Replaces tag IDs."`
}

type deleteInput struct {
	ID        int  `json:"id" jsonschema:"Numeric ID to delete."`
	Confirm   bool `json:"confirm" jsonschema:"Must be true to confirm deletion."`
	Permanent bool `json:"permanent,omitempty" jsonschema:"true deletes permanently; false moves to trash."`
}

type bulkPostStatusInput struct {
	IDs     []int `json:"ids" jsonschema:"One to 100 unique numeric WordPress post IDs."`
	Confirm bool  `json:"confirm" jsonschema:"Must be true to confirm this status change."`
}

type postStatusResult struct {
	ID    int             `json:"id"`
	Post  *wordpress.Post `json:"post,omitempty"`
	Error string          `json:"error,omitempty"`
}

type bulkPostStatusOutput struct {
	TargetStatus string             `json:"target_status"`
	Complete     bool               `json:"complete"`
	Results      []postStatusResult `json:"results"`
}

type deleteOutput struct {
	ID        int  `json:"id"`
	Deleted   bool `json:"deleted"`
	Permanent bool `json:"permanent"`
}

func addPostTools(server *mcp.Server, wp *wordpress.Client) {
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_list_posts", Description: "Lists posts, including drafts, through the authenticated API."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input listPostsInput) (*mcp.CallToolResult, postsOutput, error) {
			if err := validatePagination(input.Page, input.PerPage); err != nil {
				return nil, postsOutput{}, err
			}
			posts, err := wp.ListPosts(ctx, input.Status, input.Search, input.Page, input.PerPage)
			return nil, postsOutput{Posts: posts}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_get_post", Description: "Gets a post with editable HTML (content.raw)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input postIDInput) (*mcp.CallToolResult, postOutput, error) {
			if err := validID(input.ID, "post"); err != nil {
				return nil, postOutput{}, err
			}
			post, err := wp.GetPost(ctx, input.ID)
			return nil, postOutput{Post: post}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_create_post", Description: "Creates a post (draft by default) and always closes comments."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input createPostInput) (*mcp.CallToolResult, postOutput, error) {
			post, err := wp.CreatePost(ctx, input.Title, input.Content, input.Status, input.Excerpt, input.Categories, input.Tags)
			return nil, postOutput{Post: post}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_update_post", Description: "Edits title, HTML, excerpt, categories, or tags. It cannot change slug, date, or status."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input updatePostInput) (*mcp.CallToolResult, postOutput, error) {
			if err := validID(input.ID, "post"); err != nil {
				return nil, postOutput{}, err
			}
			post, err := wp.UpdatePost(ctx, input.ID, input.Title, input.Content, input.Excerpt, input.Categories, input.Tags)
			return nil, postOutput{Post: post}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_publish_posts", Description: "Publishes one or more posts. Requires confirm=true and reports each result; the operation is not transactional."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input bulkPostStatusInput) (*mcp.CallToolResult, bulkPostStatusOutput, error) {
			return changePostStatus(ctx, wp, input, "publish")
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_unpublish_posts", Description: "Moves one or more published posts back to draft. Requires confirm=true and reports each result; the operation is not transactional."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input bulkPostStatusInput) (*mcp.CallToolResult, bulkPostStatusOutput, error) {
			return changePostStatus(ctx, wp, input, "draft")
		})

	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_delete_post", Description: "Moves a post to trash or deletes it permanently. Requires confirm=true."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input deleteInput) (*mcp.CallToolResult, deleteOutput, error) {
			if err := validID(input.ID, "post"); err != nil {
				return nil, deleteOutput{}, err
			}
			if !input.Confirm {
				return nil, deleteOutput{}, errors.New("delete requires confirm=true")
			}
			err := wp.DeletePost(ctx, input.ID, input.Permanent)
			return nil, deleteOutput{ID: input.ID, Deleted: err == nil, Permanent: input.Permanent}, err
		})
}

type listMediaInput struct {
	Search    string `json:"search,omitempty" jsonschema:"Text search for media."`
	MediaType string `json:"media_type,omitempty" jsonschema:"Media type, normally image."`
	Page      int    `json:"page,omitempty" jsonschema:"Page number, starting at 1."`
	PerPage   int    `json:"per_page,omitempty" jsonschema:"Items per page, maximum 100."`
}

type mediaOutput struct {
	Media wordpress.Media `json:"media"`
}

type mediaListOutput struct {
	Media []wordpress.Media `json:"media"`
}

type uploadMediaInput struct {
	FilePath string `json:"file_path" jsonschema:"Local path to a GIF, JPEG, PNG, or WebP image."`
	PostID   int    `json:"post_id,omitempty" jsonschema:"Post ID to associate with this media item."`
	AltText  string `json:"alt_text,omitempty" jsonschema:"Accessible alternative text."`
	Caption  string `json:"caption,omitempty" jsonschema:"Optional caption."`
}

type updateMediaInput struct {
	ID      int     `json:"id" jsonschema:"Numeric media ID."`
	AltText *string `json:"alt_text,omitempty" jsonschema:"New alternative text."`
	Caption *string `json:"caption,omitempty" jsonschema:"New caption."`
	PostID  *int    `json:"post_id,omitempty" jsonschema:"New associated post ID; use 0 to detach."`
}

type setCoverInput struct {
	PostID          int    `json:"post_id" jsonschema:"Post ID that will receive the cover."`
	FilePath        string `json:"file_path" jsonschema:"Local path to the cover (GIF, JPEG, PNG, or WebP)."`
	AltText         string `json:"alt_text" jsonschema:"Cover alternative text."`
	ReplaceExisting *bool  `json:"replace_existing,omitempty" jsonschema:"Replaces the existing initial image when true; default is true."`
}

type coverOutput struct {
	Post  wordpress.Post  `json:"post"`
	Media wordpress.Media `json:"media"`
}

func addMediaTools(server *mcp.Server, wp *wordpress.Client) {
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_list_media", Description: "Lists images and other media through the authenticated API."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input listMediaInput) (*mcp.CallToolResult, mediaListOutput, error) {
			if err := validatePagination(input.Page, input.PerPage); err != nil {
				return nil, mediaListOutput{}, err
			}
			media, err := wp.ListMedia(ctx, input.Search, input.MediaType, input.Page, input.PerPage)
			return nil, mediaListOutput{Media: media}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_get_media", Description: "Gets editable metadata for a media item."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input postIDInput) (*mcp.CallToolResult, mediaOutput, error) {
			if err := validID(input.ID, "media"); err != nil {
				return nil, mediaOutput{}, err
			}
			media, err := wp.GetMedia(ctx, input.ID)
			return nil, mediaOutput{Media: media}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_upload_media", Description: "Securely uploads a local GIF, JPEG, PNG, or WebP and sets optional alt text, caption, and post."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input uploadMediaInput) (*mcp.CallToolResult, mediaOutput, error) {
			if input.PostID < 0 {
				return nil, mediaOutput{}, errors.New("post_id must be positive")
			}
			media, err := wp.UploadMedia(ctx, input.FilePath, input.PostID, input.AltText, input.Caption)
			return nil, mediaOutput{Media: media}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_update_media", Description: "Edits a media item's alt text, caption, or post association; it does not replace the binary file."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input updateMediaInput) (*mcp.CallToolResult, mediaOutput, error) {
			if err := validID(input.ID, "media"); err != nil {
				return nil, mediaOutput{}, err
			}
			if input.PostID != nil && *input.PostID < 0 {
				return nil, mediaOutput{}, errors.New("post_id must be zero or positive")
			}
			media, err := wp.UpdateMedia(ctx, input.ID, input.AltText, input.Caption, input.PostID)
			return nil, mediaOutput{Media: media}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_delete_media", Description: "Moves a media item to trash or deletes it permanently. Requires confirm=true."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input deleteInput) (*mcp.CallToolResult, deleteOutput, error) {
			if err := validID(input.ID, "media"); err != nil {
				return nil, deleteOutput{}, err
			}
			if !input.Confirm {
				return nil, deleteOutput{}, errors.New("delete requires confirm=true")
			}
			err := wp.DeleteMedia(ctx, input.ID, input.Permanent)
			return nil, deleteOutput{ID: input.ID, Deleted: err == nil, Permanent: input.Permanent}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_set_post_cover", Description: "Uploads an image and places it as the first body element using a relative URL. It does not set a featured image."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input setCoverInput) (*mcp.CallToolResult, coverOutput, error) {
			if err := validID(input.PostID, "post"); err != nil {
				return nil, coverOutput{}, err
			}
			replaceExisting := true
			if input.ReplaceExisting != nil {
				replaceExisting = *input.ReplaceExisting
			}
			post, media, err := wp.SetPostCover(ctx, input.PostID, input.FilePath, input.AltText, replaceExisting)
			return nil, coverOutput{Post: post, Media: media}, err
		})
}

type termsInput struct {
	Search  string `json:"search,omitempty" jsonschema:"Search by name."`
	Page    int    `json:"page,omitempty" jsonschema:"Page number, starting at 1."`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Items per page, maximum 100."`
}

type termsOutput struct {
	Terms []wordpress.Term `json:"terms"`
}

type healthOutput struct {
	BaseURL string         `json:"base_url"`
	User    wordpress.User `json:"user"`
}

type liveOutput struct {
	Post       wordpress.Post `json:"post"`
	HTTPStatus int            `json:"http_status"`
	Live       bool           `json:"live"`
}

func addSiteTools(server *mcp.Server, wp *wordpress.Client, baseURL string) {
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_list_categories", Description: "Lists existing categories for reuse."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input termsInput) (*mcp.CallToolResult, termsOutput, error) {
			if err := validatePagination(input.Page, input.PerPage); err != nil {
				return nil, termsOutput{}, err
			}
			terms, err := wp.ListTerms(ctx, "categories", input.Search, input.Page, input.PerPage)
			return nil, termsOutput{Terms: terms}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_list_tags", Description: "Lists existing tags for reuse."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input termsInput) (*mcp.CallToolResult, termsOutput, error) {
			if err := validatePagination(input.Page, input.PerPage); err != nil {
				return nil, termsOutput{}, err
			}
			terms, err := wp.ListTerms(ctx, "tags", input.Search, input.Page, input.PerPage)
			return nil, termsOutput{Terms: terms}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_site_health", Description: "Validates REST authentication and shows the application user without revealing credentials."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, healthOutput, error) {
			user, err := wp.CurrentUser(ctx)
			return nil, healthOutput{BaseURL: baseURL, User: user}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "wordpress_check_post_live", Description: "Confirms that a post is published and that its public permalink returns 2xx."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input postIDInput) (*mcp.CallToolResult, liveOutput, error) {
			if err := validID(input.ID, "post"); err != nil {
				return nil, liveOutput{}, err
			}
			post, httpStatus, live, err := wp.CheckPostLive(ctx, input.ID)
			return nil, liveOutput{Post: post, HTTPStatus: httpStatus, Live: live}, err
		})
}

func changePostStatus(ctx context.Context, wp *wordpress.Client, input bulkPostStatusInput, status string) (*mcp.CallToolResult, bulkPostStatusOutput, error) {
	if !input.Confirm {
		return nil, bulkPostStatusOutput{}, errors.New("status change requires confirm=true")
	}
	if err := validateBulkPostIDs(input.IDs); err != nil {
		return nil, bulkPostStatusOutput{}, err
	}

	output := bulkPostStatusOutput{
		TargetStatus: status,
		Complete:     true,
		Results:      make([]postStatusResult, 0, len(input.IDs)),
	}
	for _, id := range input.IDs {
		post, err := wp.SetPostStatus(ctx, id, status)
		if err != nil {
			output.Complete = false
			output.Results = append(output.Results, postStatusResult{ID: id, Error: err.Error()})
			continue
		}
		output.Results = append(output.Results, postStatusResult{ID: id, Post: &post})
	}
	return nil, output, nil
}

func validateBulkPostIDs(ids []int) error {
	if len(ids) == 0 || len(ids) > 100 {
		return errors.New("ids must contain between one and 100 post IDs")
	}
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if err := validID(id, "post"); err != nil {
			return err
		}
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("post ID %d is duplicated", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validID(id int, resource string) error {
	if id <= 0 {
		return fmt.Errorf("%s ID must be positive", resource)
	}
	return nil
}

func validatePagination(page, perPage int) error {
	if page < 0 {
		return errors.New("page must be positive")
	}
	if perPage < 0 || perPage > 100 {
		return errors.New("per_page must be between 1 and 100")
	}
	return nil
}
