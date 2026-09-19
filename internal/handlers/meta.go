package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/EmiraBlight/dev-crypt-blog-be/internal/models"
)

// Root godoc
//
//	@Summary		Service greeting
//	@Description	Echoes the caller's User-Agent inside a placeholder Comment. Used as a liveness check.
//	@Tags			meta
//	@Produce		json
//	@Success		200	{object}	models.Comment
//	@Router			/ [get]
func (h *Handler) Root(c *gin.Context) {
	text := "No User-Agent header was provided."
	if ua := c.GetHeader("User-Agent"); ua != "" {
		text = "Your User-Agent is: " + ua
	}

	c.JSON(http.StatusOK, models.Comment{
		CommentID:   -1,
		UserID:      "0",
		PostID:      0,
		CommentText: text,
		PostTime:    time.Now().UTC(),
		UserName:    "System",
	})
}
