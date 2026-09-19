package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EmiraBlight/dev-crypt-blog-be/internal/auth"
	"github.com/EmiraBlight/dev-crypt-blog-be/internal/models"
)

// CreateUser godoc
//
//	@Summary		Register a display name
//	@Description	Stores the Firebase UID to display-name mapping. Existing users are left untouched.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			user	body		models.NewUser	true	"User to register"
//	@Success		200		{string}	string			"User created successfully"
//	@Failure		400		{string}	string			"Malformed request body"
//	@Router			/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var body models.NewUser
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Invalid request body: %v", err)
		return
	}

	const q = `INSERT INTO usernames (uuid, username) VALUES ($1, $2) ON CONFLICT (uuid) DO NOTHING`
	if _, err := h.pool.Exec(c.Request.Context(), q, body.NewUserUUID, body.NewUserName); err != nil {
		log.Printf("insert user failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to create user")
		return
	}

	c.JSON(http.StatusOK, "User created successfully")
}

// GetMyRoles godoc
//
//	@Summary	List the authenticated user's roles
//	@Tags		users
//	@Produce	json
//	@Success	200	{array}		string
//	@Failure	401	{string}	string	"Missing or invalid token"
//	@Security	BearerAuth
//	@Router		/users/me/roles [get]
func (h *Handler) GetMyRoles(c *gin.Context) {
	roles, err := auth.GetUserRoles(c.Request.Context(), h.pool, auth.UID(c))
	if err != nil {
		log.Printf("role lookup failed: %v", err)
		c.String(http.StatusInternalServerError, "Failed to look up user roles")
		return
	}

	c.JSON(http.StatusOK, roles)
}
