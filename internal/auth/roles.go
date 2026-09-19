package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Role names recognised by the API.
const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleWriter    = "writer"
)

// ErrPostNotFound is returned by CanMutatePost when the post does not exist.
var ErrPostNotFound = errors.New("Post not found")

// ErrForbidden is returned when a user lacks the required role.
var ErrForbidden = errors.New("Forbidden: Insufficient permissions")

// ErrNotPostOwner is returned when a writer targets someone else's post.
var ErrNotPostOwner = errors.New("Forbidden: You may only modify your own posts")

// GetUserRoles returns the role names granted to a Firebase UID.
func GetUserRoles(ctx context.Context, pool *pgxpool.Pool, uid string) ([]string, error) {
	const q = `SELECT r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_uuid = $1`
	rows, err := pool.Query(ctx, q, uid)
	if err != nil {
		return nil, fmt.Errorf("querying user roles: %w", err)
	}
	defer rows.Close()

	roles := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scanning user role: %w", err)
		}
		roles = append(roles, name)
	}
	return roles, rows.Err()
}

// HasAnyRole reports whether any of wanted appears in roles.
func HasAnyRole(roles []string, wanted ...string) bool {
	for _, r := range roles {
		for _, w := range wanted {
			if r == w {
				return true
			}
		}
	}
	return false
}

// CanMutatePost authorises a mutation on an existing post: admins may edit or
// delete any post, writers only the posts they authored. Returns ErrPostNotFound
// for a missing post so the caller does not have to look it up separately.
func CanMutatePost(ctx context.Context, pool *pgxpool.Pool, uid string, postID int) error {
	var authorUUID *string
	err := pool.QueryRow(ctx, `SELECT author_uuid FROM posts WHERE id = $1`, postID).Scan(&authorUUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPostNotFound
	}
	if err != nil {
		return fmt.Errorf("looking up post author: %w", err)
	}

	roles, err := GetUserRoles(ctx, pool, uid)
	if err != nil {
		return err
	}

	isAdmin := HasAnyRole(roles, RoleAdmin)
	isOwner := HasAnyRole(roles, RoleWriter) && authorUUID != nil && *authorUUID == uid
	if isAdmin || isOwner {
		return nil
	}
	return ErrNotPostOwner
}
