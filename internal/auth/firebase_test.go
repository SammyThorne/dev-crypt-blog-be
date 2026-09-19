package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestVerifier points a Verifier at a stub Identity Toolkit.
func newTestVerifier(t *testing.T, handler http.HandlerFunc) *Verifier {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	v := NewVerifier("test-key")
	v.client = srv.Client()
	v.endpoint = srv.URL
	return v
}

func TestVerifyHeaderValidation(t *testing.T) {
	v := newTestVerifier(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("firebase should not be called for a malformed header")
	})

	tests := []struct {
		name   string
		header string
		want   error
	}{
		{"empty", "", ErrMissingHeader},
		{"no bearer prefix", "abc123", ErrMalformedHeader},
		{"wrong scheme", "Basic abc123", ErrMalformedHeader},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := v.Verify(context.Background(), tc.header); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestVerifySuccess(t *testing.T) {
	v := newTestVerifier(t, func(w http.ResponseWriter, r *http.Request) {
		var body firebaseRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.IDToken != "tok" {
			t.Fatalf("got idToken %q, want %q", body.IDToken, "tok")
		}
		_ = json.NewEncoder(w).Encode(firebaseResponse{Users: []firebaseUser{{LocalID: "uid-1"}}})
	})

	uid, err := v.Verify(context.Background(), "Bearer tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != "uid-1" {
		t.Fatalf("got uid %q, want %q", uid, "uid-1")
	}
}

func TestVerifyFirebaseFailures(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    error
	}{
		{
			name:    "rejected",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) },
			want:    ErrRejected,
		},
		{
			name:    "unparsable",
			handler: func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("not json")) },
			want:    ErrUnparsable,
		},
		{
			name:    "no user",
			handler: func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"users":[]}`)) },
			want:    ErrNoUser,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := newTestVerifier(t, tc.handler)
			if _, err := v.Verify(context.Background(), "Bearer tok"); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
