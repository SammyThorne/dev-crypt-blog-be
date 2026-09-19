package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// contextKeyUID is the gin context key holding the authenticated Firebase UID.
const contextKeyUID = "firebaseUID"

// UID returns the authenticated Firebase UID stored by RequireAuth.
func UID(c *gin.Context) string {
	uid, _ := c.Get(contextKeyUID)
	s, _ := uid.(string)
	return s
}

// RequireAuth validates the Authorization header and stores the resulting
// Firebase UID on the request context. Error bodies are plain text, matching
// the previous Servant implementation that clients read with res.text().
func RequireAuth(v *Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := v.Verify(c.Request.Context(), c.GetHeader("Authorization"))
		if err != nil {
			switch {
			case errors.Is(err, ErrMissingHeader),
				errors.Is(err, ErrMalformedHeader),
				errors.Is(err, ErrRejected),
				errors.Is(err, ErrUnparsable),
				errors.Is(err, ErrNoUser):
				c.String(http.StatusUnauthorized, err.Error())
			default:
				log.Printf("firebase token validation error: %v", err)
				c.String(http.StatusInternalServerError, "Failed to validate token")
			}
			c.Abort()
			return
		}

		c.Set(contextKeyUID, uid)
		c.Next()
	}
}

// RequireAnyRole aborts with 403 unless the authenticated user holds at least
// one of the given roles. It must run after RequireAuth.
func RequireAnyRole(pool *pgxpool.Pool, wanted ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := GetUserRoles(c.Request.Context(), pool, UID(c))
		if err != nil {
			log.Printf("role lookup failed: %v", err)
			c.String(http.StatusInternalServerError, "Failed to look up user roles")
			c.Abort()
			return
		}
		if !HasAnyRole(roles, wanted...) {
			c.String(http.StatusForbidden, ErrForbidden.Error())
			c.Abort()
			return
		}
		c.Next()
	}
}
