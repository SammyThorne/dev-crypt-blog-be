package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/EmiraBlight/dev-crypt-blog-be/internal/config"
)

// newTestRouter builds the real router with a nil pool. Only routes that reject
// the request before touching the database may be exercised with it.
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return New(&config.Config{AllowedOrigins: []string{"*"}, FirebaseAPIKey: "test-key"}, nil)
}

func TestRootEchoesUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "curl/8.0")
	w := httptest.NewRecorder()

	newTestRouter(t).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := body["commentText"]; got != "Your User-Agent is: curl/8.0" {
		t.Fatalf("got commentText %q", got)
	}
	if got := body["userName"]; got != "System" {
		t.Fatalf("got userName %q, want System", got)
	}
	if got := body["commentId"]; got != float64(-1) {
		t.Fatalf("got commentId %v, want -1", got)
	}
}

func TestRootWithoutUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Del("User-Agent")
	w := httptest.NewRecorder()

	newTestRouter(t).ServeHTTP(w, req)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := body["commentText"]; got != "No User-Agent header was provided." {
		t.Fatalf("got commentText %q", got)
	}
}

// Every protected route must reject an anonymous request before it reaches the
// database, with a plain-text body the frontend can surface directly.
func TestProtectedRoutesRejectAnonymous(t *testing.T) {
	r := newTestRouter(t)

	routes := []struct{ method, path string }{
		{http.MethodPost, "/comments"},
		{http.MethodDelete, "/comments/1"},
		{http.MethodPost, "/posts"},
		{http.MethodPut, "/posts/1"},
		{http.MethodDelete, "/posts/1"},
		{http.MethodGet, "/users/me/roles"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(rt.method, rt.path, nil))

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("got status %d, want 401", w.Code)
			}
			if got := w.Body.String(); got != "Missing Authorization header" {
				t.Fatalf("got body %q", got)
			}
			if ct := w.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
				t.Fatalf("got content-type %q, want plain text", ct)
			}
		})
	}
}

// The frontend sends Content-Type and Authorization on cross-origin requests,
// and uses PUT/DELETE, so the preflight must allow all of them.
func TestCORSPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/posts/1", nil)
	req.Header.Set("Origin", "https://blog.example")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	w := httptest.NewRecorder()

	newTestRouter(t).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("got status %d, want a successful preflight", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("got allow-origin %q, want *", got)
	}
	for _, want := range []string{"PUT", "DELETE"} {
		if !headerContains(w.Header().Get("Access-Control-Allow-Methods"), want) {
			t.Fatalf("allow-methods %q missing %s", w.Header().Get("Access-Control-Allow-Methods"), want)
		}
	}
	for _, want := range []string{"Authorization", "Content-Type"} {
		if !headerContains(w.Header().Get("Access-Control-Allow-Headers"), want) {
			t.Fatalf("allow-headers %q missing %s", w.Header().Get("Access-Control-Allow-Headers"), want)
		}
	}
}

func TestSwaggerUIIsServed(t *testing.T) {
	w := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", w.Code)
	}

	var spec map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("swagger spec is not valid json: %v", err)
	}
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatal("swagger spec has no paths")
	}
	for _, want := range []string{"/", "/posts", "/posts/{postId}", "/posts/{postId}/comments", "/comments", "/comments/{commentId}", "/users", "/users/me/roles"} {
		if _, ok := paths[want]; !ok {
			t.Fatalf("swagger spec is missing path %s", want)
		}
	}
}

func headerContains(header, want string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.TrimSpace(part) == want {
			return true
		}
	}
	return false
}
