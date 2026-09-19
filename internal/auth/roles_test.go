package auth

import "testing"

func TestHasAnyRole(t *testing.T) {
	tests := []struct {
		name   string
		roles  []string
		wanted []string
		want   bool
	}{
		{"match", []string{"writer"}, []string{"admin", "writer"}, true},
		{"no match", []string{"writer"}, []string{"admin", "moderator"}, false},
		{"empty roles", nil, []string{"admin"}, false},
		{"empty wanted", []string{"admin"}, nil, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasAnyRole(tc.roles, tc.wanted...); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
