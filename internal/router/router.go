// Package router wires the middleware and routes onto a gin engine.
package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/EmiraBlight/dev-crypt-blog-be/docs" // generated swagger spec
	"github.com/EmiraBlight/dev-crypt-blog-be/internal/auth"
	"github.com/EmiraBlight/dev-crypt-blog-be/internal/config"
	"github.com/EmiraBlight/dev-crypt-blog-be/internal/handlers"
)

// New builds the fully-wired gin engine.
func New(cfg *config.Config, pool *pgxpool.Pool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(corsConfig(cfg)))

	h := handlers.New(pool, cfg.BlogsJSONPath)
	verifier := auth.NewVerifier(cfg.FirebaseAPIKey)
	authed := auth.RequireAuth(verifier)

	r.GET("/", h.Root)
	r.GET("/blogs.json", h.BlogsJSON)

	r.POST("/comments", authed, h.CreateComment)
	r.DELETE("/comments/:commentId", authed,
		auth.RequireAnyRole(pool, auth.RoleAdmin, auth.RoleModerator), h.DeleteComment)

	r.GET("/posts", h.ListPosts)
	r.POST("/posts", authed,
		auth.RequireAnyRole(pool, auth.RoleAdmin, auth.RoleWriter), h.CreatePost)
	r.PUT("/posts/:postId", authed, h.UpdatePost)
	r.DELETE("/posts/:postId", authed, h.DeletePost)
	r.GET("/posts/:postId/comments", h.ListPostComments)

	r.POST("/users", h.CreateUser)
	r.GET("/users/me/roles", authed, h.GetMyRoles)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

// corsConfig mirrors the previous wai-cors policy: the simple methods plus
// PUT/DELETE, and the Content-Type and Authorization request headers.
func corsConfig(cfg *config.Config) cors.Config {
	c := cors.Config{
		AllowMethods:     []string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
		c.AllowAllOrigins = true
	} else {
		c.AllowOrigins = cfg.AllowedOrigins
	}

	return c
}
