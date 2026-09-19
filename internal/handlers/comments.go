package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/EmiraBlight/dev-crypt-blog-be/internal/auth"
	"github.com/EmiraBlight/dev-crypt-blog-be/internal/models"
)

// CreateComment godoc
//
//	@Summary		Post a comment
//	@Description	Creates a comment on a blog post. The Firebase UID in the token must match newUserId.
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			comment	body		models.NewComment	true	"Comment to create"
//	@Success		200		{object}	models.Comment
//	@Failure		400		{string}	string	"Malformed request body"
//	@Failure		401		{string}	string	"Missing, invalid, or mismatched token"
//	@Failure		500		{string}	string	"Failed to insert comment"
//	@Security		BearerAuth
//	@Router			/comments [post]
func (h *Handler) CreateComment(c *gin.Context) {
	var body models.NewComment
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: %v", err)
		return
	}

	if auth.UID(c) != body.NewUserID {
		c.String(http.StatusUnauthorized, "Unauthorized: Token UID does not match request UID.")
		return
	}

	const q = `WITH inserted AS (
                   INSERT INTO comments (user_id, post_id, comment)
                   VALUES ($1, $2, $3)
                   RETURNING id, user_id, post_id, comment, post_time
               )
               SELECT i.id, i.user_id, i.post_id, i.comment, i.post_time, COALESCE(u.username, 'Unknown')
               FROM inserted i LEFT JOIN usernames u ON i.user_id = u.uuid`

	var out models.Comment
	err := h.pool.QueryRow(c.Request.Context(), q, body.NewUserID, body.NewPostID, body.NewCommentText).
		Scan(&out.CommentID, &out.UserID, &out.PostID, &out.CommentText, &out.PostTime, &out.UserName)
	if err != nil {
		log.Printf("insert comment failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to insert comment")
		return
	}

	c.JSON(http.StatusOK, out)
}

// DeleteComment godoc
//
//	@Summary		Delete a comment
//	@Description	Deletes a comment. Requires the admin or moderator role.
//	@Tags			comments
//	@Produce		json
//	@Param			commentId	path		int	true	"Comment ID"
//	@Success		200			{string}	string	"Comment deleted successfully"
//	@Failure		400			{string}	string	"Invalid comment id"
//	@Failure		401			{string}	string	"Missing or invalid token"
//	@Failure		403			{string}	string	"Insufficient permissions"
//	@Failure		404			{string}	string	"Comment not found"
//	@Security		BearerAuth
//	@Router			/comments/{commentId} [delete]
func (h *Handler) DeleteComment(c *gin.Context) {
	commentID, err := strconv.Atoi(c.Param("commentId"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid comment id")
		return
	}

	tag, err := h.pool.Exec(c.Request.Context(), `DELETE FROM comments WHERE id = $1`, commentID)
	if err != nil {
		log.Printf("delete comment failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to delete comment")
		return
	}
	if tag.RowsAffected() == 0 {
		c.String(http.StatusNotFound, "Comment not found")
		return
	}

	c.JSON(http.StatusOK, "Comment deleted successfully")
}

// ListPostComments godoc
//
//	@Summary		List a post's comments
//	@Description	Returns comments for a post, oldest first, with optional pagination.
//	@Tags			comments
//	@Produce		json
//	@Param			postId	path		int	true	"Post ID"
//	@Param			offset	query		int	false	"Rows to skip"		default(0)
//	@Param			limit	query		int	false	"Maximum rows"		default(10)
//	@Success		200		{array}		models.Comment
//	@Failure		400		{string}	string	"Invalid post id"
//	@Router			/posts/{postId}/comments [get]
func (h *Handler) ListPostComments(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("postId"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid post id")
		return
	}

	offset := queryInt(c, "offset", 0)
	limit := queryInt(c, "limit", 10)

	const q = `SELECT c.id, c.user_id, c.post_id, c.comment, c.post_time, COALESCE(u.username, 'Unknown')
               FROM comments c LEFT JOIN usernames u ON c.user_id = u.uuid
               WHERE c.post_id = $1 ORDER BY c.id ASC OFFSET $2 LIMIT $3`

	rows, err := h.pool.Query(c.Request.Context(), q, postID, offset, limit)
	if err != nil {
		log.Printf("list comments failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to fetch comments")
		return
	}
	defer rows.Close()

	comments, err := scanComments(rows)
	if err != nil {
		log.Printf("list comments failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to fetch comments")
		return
	}

	c.JSON(http.StatusOK, comments)
}

// scanComments drains a comment row set into Comments. It
// returns an empty (non-nil) slice so the JSON body is [] rather than null.
func scanComments(rows pgx.Rows) ([]models.Comment, error) {
	comments := []models.Comment{}
	for rows.Next() {
		var cm models.Comment
		if err := rows.Scan(&cm.CommentID, &cm.UserID, &cm.PostID, &cm.CommentText, &cm.PostTime, &cm.UserName); err != nil {
			return nil, err
		}
		comments = append(comments, cm)
	}
	return comments, rows.Err()
}

// queryInt reads a non-negative integer query parameter, falling back to def
// when it is absent or unparsable.
func queryInt(c *gin.Context, name string, def int) int {
	raw := c.Query(name)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return def
	}
	return v
}
