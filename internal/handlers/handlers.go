// Package handlers implements the HTTP endpoints.
package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler carries the dependencies shared by every endpoint.
type Handler struct {
	pool          *pgxpool.Pool
	blogsJSONPath string
}

// New builds a Handler backed by the given connection pool. blogsJSONPath is
// the file served at /blogs.json.
func New(pool *pgxpool.Pool, blogsJSONPath string) *Handler {
	return &Handler{pool: pool, blogsJSONPath: blogsJSONPath}
}
