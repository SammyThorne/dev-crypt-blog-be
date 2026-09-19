package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/EmiraBlight/dev-crypt-blog-be/internal/auth"
	"github.com/EmiraBlight/dev-crypt-blog-be/internal/models"
)

// ListPosts godoc
//
//	@Summary	List blog posts
//	@Tags		posts
//	@Produce	json
//	@Success	200	{array}	models.BlogPost
//	@Router		/posts [get]
func (h *Handler) ListPosts(c *gin.Context) {
	const q = `SELECT p.id, p.title, p.blurb, p.content, p.date_time, p.author_uuid, COALESCE(u.username, 'Unknown')
               FROM posts p LEFT JOIN usernames u ON p.author_uuid = u.uuid
               ORDER BY p.date_time DESC`

	rows, err := h.pool.Query(c.Request.Context(), q)
	if err != nil {
		log.Printf("list posts failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to fetch posts")
		return
	}
	defer rows.Close()

	posts := []models.BlogPost{}
	for rows.Next() {
		var p models.BlogPost
		if err := rows.Scan(&p.PID, &p.PTitle, &p.PBlurb, &p.PContent, &p.PDateTime, &p.PAuthorUUID, &p.PAuthorName); err != nil {
			log.Printf("list posts failed: %v", err)
			c.String(http.StatusInternalServerError, "Failed to fetch posts")
			return
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("list posts failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to fetch posts")
		return
	}

	c.JSON(http.StatusOK, posts)
}

// CreatePost godoc
//
//	@Summary		Publish a blog post
//	@Description	Creates a post authored by the authenticated user. Requires the admin or writer role.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			post	body		models.NewPost	true	"Post to create"
//	@Success		200		{object}	models.BlogPost
//	@Failure		400		{string}	string	"Malformed request body"
//	@Failure		401		{string}	string	"Missing or invalid token"
//	@Failure		403		{string}	string	"Insufficient permissions"
//	@Failure		500		{string}	string	"Failed to insert post"
//	@Security		BearerAuth
//	@Router			/posts [post]
func (h *Handler) CreatePost(c *gin.Context) {
	var body models.NewPost
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: %v", err)
		return
	}

	const q = `WITH inserted AS (
                   INSERT INTO posts (title, blurb, content, author_uuid)
                   VALUES ($1, $2, $3, $4)
                   RETURNING id, title, blurb, content, date_time, author_uuid
               )
               SELECT i.id, i.title, i.blurb, i.content, i.date_time, i.author_uuid, COALESCE(u.username, 'Unknown')
               FROM inserted i LEFT JOIN usernames u ON i.author_uuid = u.uuid`

	var p models.BlogPost
	err := h.pool.QueryRow(c.Request.Context(), q, body.NewPostTitle, body.NewPostBlurb, body.NewPostContent, auth.UID(c)).
		Scan(&p.PID, &p.PTitle, &p.PBlurb, &p.PContent, &p.PDateTime, &p.PAuthorUUID, &p.PAuthorName)
	if err != nil {
		log.Printf("insert post failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to insert post")
		return
	}

	c.JSON(http.StatusOK, p)
}

// UpdatePost godoc
//
//	@Summary		Edit a blog post
//	@Description	Updates a post. Admins may edit any post, writers only their own.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postId	path		int				true	"Post ID"
//	@Param			post	body		models.NewPost	true	"Updated post"
//	@Success		200		{object}	models.BlogPost
//	@Failure		400		{string}	string	"Malformed request body or post id"
//	@Failure		401		{string}	string	"Missing or invalid token"
//	@Failure		403		{string}	string	"Not permitted to modify this post"
//	@Failure		404		{string}	string	"Post not found"
//	@Failure		500		{string}	string	"Failed to update post"
//	@Security		BearerAuth
//	@Router			/posts/{postId} [put]
func (h *Handler) UpdatePost(c *gin.Context) {
	postID, ok := h.authorizePostMutation(c)
	if !ok {
		return
	}

	var body models.NewPost
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: %v", err)
		return
	}

	const q = `WITH updated AS (
                   UPDATE posts SET title = $1, blurb = $2, content = $3 WHERE id = $4
                   RETURNING id, title, blurb, content, date_time, author_uuid
               )
               SELECT p.id, p.title, p.blurb, p.content, p.date_time, p.author_uuid, COALESCE(u.username, 'Unknown')
               FROM updated p LEFT JOIN usernames u ON p.author_uuid = u.uuid`

	var p models.BlogPost
	err := h.pool.QueryRow(c.Request.Context(), q, body.NewPostTitle, body.NewPostBlurb, body.NewPostContent, postID).
		Scan(&p.PID, &p.PTitle, &p.PBlurb, &p.PContent, &p.PDateTime, &p.PAuthorUUID, &p.PAuthorName)
	if err != nil {
		log.Printf("update post failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to update post")
		return
	}

	c.JSON(http.StatusOK, p)
}

// DeletePost godoc
//
//	@Summary		Delete a blog post
//	@Description	Deletes a post and its comments. Admins may delete any post, writers only their own.
//	@Tags			posts
//	@Produce		json
//	@Param			postId	path		int	true	"Post ID"
//	@Success		200		{string}	string	"Post deleted successfully"
//	@Failure		400		{string}	string	"Invalid post id"
//	@Failure		401		{string}	string	"Missing or invalid token"
//	@Failure		403		{string}	string	"Not permitted to modify this post"
//	@Failure		404		{string}	string	"Post not found"
//	@Security		BearerAuth
//	@Router			/posts/{postId} [delete]
func (h *Handler) DeletePost(c *gin.Context) {
	postID, ok := h.authorizePostMutation(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		log.Printf("delete post failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to delete post")
		return
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op once the tx is committed

	if _, err := tx.Exec(ctx, `DELETE FROM comments WHERE post_id = $1`, postID); err != nil {
		log.Printf("delete post comments failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to delete post")
		return
	}

	tag, err := tx.Exec(ctx, `DELETE FROM posts WHERE id = $1`, postID)
	if err != nil {
		log.Printf("delete post failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to delete post")
		return
	}
	if tag.RowsAffected() == 0 {
		c.String(http.StatusNotFound, "Post not found")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("delete post commit failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to delete post")
		return
	}

	c.JSON(http.StatusOK, "Post deleted successfully")
}

// authorizePostMutation parses the :postId parameter and checks the caller may
// modify that post, writing the error response itself when they may not.
func (h *Handler) authorizePostMutation(c *gin.Context) (int, bool) {
	postID, err := strconv.Atoi(c.Param("postId"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid post id")
		return 0, false
	}

	switch err := auth.CanMutatePost(c.Request.Context(), h.pool, auth.UID(c), postID); {
	case err == nil:
		return postID, true
	case errors.Is(err, auth.ErrPostNotFound):
		c.String(http.StatusNotFound, err.Error())
	case errors.Is(err, auth.ErrNotPostOwner):
		c.String(http.StatusForbidden, err.Error())
	default:
		log.Printf("post authorization failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to authorize request")
	}
	return 0, false
}
