package config

import (
	neturl "net/url"
	"strings"
)

// url percent-encodes a connection-string component so passwords containing
// ":" or "@" do not corrupt the DSN.
func url(s string) string {
	return neturl.QueryEscape(s)
}

// splitAndTrim splits a comma-separated list, dropping empty entries.
func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
