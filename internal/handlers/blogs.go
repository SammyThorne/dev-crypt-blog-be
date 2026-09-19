package handlers

import (
	"crypto/md5" //nolint:gosec // used for an ETag validator, not for security
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// BlogsJSON godoc
//
//	@Summary		Static blog index
//	@Description	Serves the blogs.json file with an ETag; send If-None-Match to receive 304 when unchanged.
//	@Tags			meta
//	@Produce		json
//	@Param			If-None-Match	header		string	false	"ETag from a previous response"
//	@Success		200				{array}		object
//	@Success		304				{string}	string	"Not modified"
//	@Failure		404				{object}	map[string]string	"blogs.json not found"
//	@Failure		500				{object}	map[string]string	"Failed to read blogs.json"
//	@Router			/blogs.json [get]
func (h *Handler) BlogsJSON(c *gin.Context) {
	content, err := os.ReadFile(h.blogsJSONPath)
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, gin.H{"error": "blogs.json not found"})
		return
	}
	if err != nil {
		log.Printf("reading %s failed: %v", h.blogsJSONPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	sum := md5.Sum(content) //nolint:gosec // ETag only
	etag := `"` + hex.EncodeToString(sum[:]) + `"`
	c.Header("ETag", etag)

	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "application/json", content)
}
