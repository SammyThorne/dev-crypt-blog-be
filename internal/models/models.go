// Package models holds the API's request and response payloads.
//
// The JSON field names are inherited from the previous Haskell implementation,
// where Aeson derived them from the record selectors verbatim. They are kept
// exactly as-is so existing clients keep working.
package models

import "time"

// Comment is a comment on a blog post, joined with the commenter's display name.
type Comment struct {
	CommentID   int       `json:"commentId" example:"12"`
	UserID      string    `json:"userId" example:"G3QzWiV57VgYwAKbMTa07OaX3ra2"`
	PostID      int       `json:"postId" example:"3"`
	CommentText string    `json:"commentText" example:"Great write-up!"`
	PostTime    time.Time `json:"postTime" example:"2026-05-24T08:25:00Z"`
	UserName    string    `json:"userName" example:"Sammy"`
}

// NewComment is the request body for creating a comment.
type NewComment struct {
	NewUserID string `json:"newUserId" binding:"required" example:"G3QzWiV57VgYwAKbMTa07OaX3ra2"`
	// NewPostID is deliberately not `binding:"required"`: post ids are
	// client-generated hashes with no foreign key, so 0 is a legal value and
	// "required" would reject it.
	NewPostID      int    `json:"newPostId" example:"3"`
	NewCommentText string `json:"newCommentText" binding:"required" example:"Great write-up!"`
}

// BlogPost is a blog post, joined with the author's display name.
type BlogPost struct {
	PID         int       `json:"pId" example:"3"`
	PTitle      string    `json:"pTitle" example:"Writing a Gin backend"`
	PBlurb      string    `json:"pBlurb" example:"A short summary of the post."`
	PContent    string    `json:"pContent" example:"# Heading\n\nBody text."`
	PDateTime   time.Time `json:"pDateTime" example:"2026-05-24T08:25:00Z"`
	PAuthorUUID *string   `json:"pAuthorUuid" example:"G3QzWiV57VgYwAKbMTa07OaX3ra2"`
	PAuthorName string    `json:"pAuthorName" example:"Sammy"`
}

// NewPost is the request body for creating or updating a blog post.
type NewPost struct {
	NewPostTitle   string `json:"newPostTitle" binding:"required" example:"Writing a Gin backend"`
	NewPostBlurb   string `json:"newPostBlurb" binding:"required" example:"A short summary of the post."`
	NewPostContent string `json:"newPostContent" binding:"required" example:"# Heading\n\nBody text."`
}

// NewUser is the request body for registering a user's display name.
type NewUser struct {
	NewUserUUID string `json:"newUserUuid" binding:"required" example:"G3QzWiV57VgYwAKbMTa07OaX3ra2"`
	NewUserName string `json:"newUserName" binding:"required" example:"Sammy"`
}
