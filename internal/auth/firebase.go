// Package auth verifies Firebase ID tokens and enforces the role rules.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Errors returned by Verify. Callers map these onto HTTP statuses.
var (
	// ErrMissingHeader is returned when no Authorization header was sent.
	ErrMissingHeader = errors.New("Missing Authorization header")
	// ErrMalformedHeader is returned when the header is not a Bearer token.
	ErrMalformedHeader = errors.New("Invalid Authorization header format. Expected 'Bearer <token>'")
	// ErrRejected is returned when Firebase refuses the token.
	ErrRejected = errors.New("Unauthorized: Invalid token (rejected by Firebase).")
	// ErrUnparsable is returned when Firebase's response cannot be decoded.
	ErrUnparsable = errors.New("Unauthorized: Could not parse token info.")
	// ErrNoUser is returned when the token is valid but maps to no user.
	ErrNoUser = errors.New("Unauthorized: Token is valid but represents no user.")
)

type firebaseRequest struct {
	IDToken string `json:"idToken"`
}

type firebaseUser struct {
	LocalID string `json:"localId"`
}

type firebaseResponse struct {
	Users []firebaseUser `json:"users"`
}

// defaultEndpoint is Google's Identity Toolkit account lookup endpoint.
const defaultEndpoint = "https://identitytoolkit.googleapis.com/v1/accounts:lookup"

// Verifier exchanges a Firebase ID token for the user's UID via the Identity
// Toolkit lookup endpoint.
type Verifier struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// NewVerifier builds a Verifier using the given Firebase Web API key.
func NewVerifier(apiKey string) *Verifier {
	return &Verifier{
		apiKey:   apiKey,
		endpoint: defaultEndpoint,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Verify validates an "Bearer <token>" Authorization header and returns the
// Firebase UID it belongs to.
func (v *Verifier) Verify(ctx context.Context, authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingHeader
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrMalformedHeader
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	body, err := json.Marshal(firebaseRequest{IDToken: token})
	if err != nil {
		return "", fmt.Errorf("encoding firebase request: %w", err)
	}

	url := v.endpoint + "?key=" + v.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("building firebase request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := v.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling firebase: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading firebase response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", ErrRejected
	}

	var parsed firebaseResponse
	if err := json.Unmarshal(resBody, &parsed); err != nil {
		return "", ErrUnparsable
	}
	if len(parsed.Users) == 0 {
		return "", ErrNoUser
	}
	return parsed.Users[0].LocalID, nil
}
