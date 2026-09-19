package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func newBlogsRouter(path string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/blogs.json", New(nil, path).BlogsJSON)
	return r
}

func TestBlogsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blogs.json")
	if err := os.WriteFile(path, []byte(`[{"title":"x"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	r := newBlogsRouter(path)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/blogs.json", nil))
	if w.Code != http.StatusOK || w.Body.String() != `[{"title":"x"}]` {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-cache" || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected headers: %v", w.Header())
	}
	etag := w.Header().Get("ETag")
	if len(etag) < 3 || etag[0] != '"' {
		t.Fatalf("bad etag %q", etag)
	}

	req := httptest.NewRequest(http.MethodGet, "/blogs.json", nil)
	req.Header.Set("If-None-Match", etag)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotModified || w.Body.Len() != 0 {
		t.Fatalf("got %d with body %q, want empty 304", w.Code, w.Body.String())
	}
}

func TestBlogsJSONMissing(t *testing.T) {
	w := httptest.NewRecorder()
	newBlogsRouter(filepath.Join(t.TempDir(), "nope.json")).
		ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/blogs.json", nil))
	if w.Code != http.StatusNotFound || w.Body.String() != `{"error":"blogs.json not found"}` {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}
