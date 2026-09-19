package models

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
	"time"
)

// The frontend reads these exact keys (dev-crypt-blog-fe/src/components/Comments.tsx,
// src/pages/Home.tsx). Renaming any of them is a breaking API change.
func TestJSONFieldNames(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  []string
	}{
		{
			name:  "Comment",
			value: Comment{},
			want:  []string{"commentId", "commentText", "postId", "postTime", "userId", "userName"},
		},
		{
			name:  "BlogPost",
			value: BlogPost{},
			want:  []string{"pAuthorName", "pAuthorUuid", "pBlurb", "pContent", "pDateTime", "pId", "pTitle"},
		},
		{
			name:  "NewComment",
			value: NewComment{},
			want:  []string{"newCommentText", "newPostId", "newUserId"},
		},
		{
			name:  "NewPost",
			value: NewPost{},
			want:  []string{"newPostBlurb", "newPostContent", "newPostTitle"},
		},
		{
			name:  "NewUser",
			value: NewUser{},
			want:  []string{"newUserName", "newUserUuid"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var m map[string]json.RawMessage
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got := make([]string, 0, len(m))
			for k := range m {
				got = append(got, k)
			}
			sort.Strings(got)

			if len(got) != len(tc.want) {
				t.Fatalf("got keys %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got keys %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// A null author_uuid must serialise as JSON null, which the frontend checks with
// `item.pAuthorUuid ?? null`.
func TestBlogPostNullAuthor(t *testing.T) {
	raw, err := json.Marshal(BlogPost{PDateTime: time.Unix(0, 0).UTC()})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if want := `"pAuthorUuid":null`; !strings.Contains(string(raw), want) {
		t.Fatalf("got %s, want it to contain %s", raw, want)
	}
	if want := `"pDateTime":"1970-01-01T00:00:00Z"`; !strings.Contains(string(raw), want) {
		t.Fatalf("got %s, want it to contain %s", raw, want)
	}
}
