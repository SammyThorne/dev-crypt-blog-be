package config

import (
	"os"
	"strings"
	"testing"
)

func TestDBPasswordDefault(t *testing.T) {
	os.Unsetenv("DB_PASSWORD") //nolint:errcheck
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBPassword != "postgres" {
		t.Fatalf("got %q, want the postgres default", cfg.DBPassword)
	}

	t.Setenv("DB_PASSWORD", "secret")
	if cfg, _ = Load(); cfg.DBPassword != "secret" {
		t.Fatalf("got %q, want the explicit value", cfg.DBPassword)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DB_PASSWORD", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DBHost != "localhost" || cfg.DBUser != "postgres" || cfg.DBName != "test_db" {
		t.Fatalf("unexpected db defaults: %+v", cfg)
	}
	if cfg.Port != 8081 {
		t.Fatalf("got port %d, want 8081", cfg.Port)
	}
	if !cfg.UseTLS {
		t.Fatal("TLS should default to enabled")
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Fatalf("got origins %v, want [*]", cfg.AllowedOrigins)
	}
}

// Only the literal string "false" disables TLS, matching the previous behaviour.
func TestUseTLSParsing(t *testing.T) {
	tests := map[string]bool{"false": false, "true": true, "0": true, "no": true, "": true}

	for value, want := range tests {
		t.Run("USE_TLS="+value, func(t *testing.T) {
			t.Setenv("DB_PASSWORD", "secret")
			t.Setenv("USE_TLS", value)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.UseTLS != want {
				t.Fatalf("got UseTLS=%v, want %v", cfg.UseTLS, want)
			}
		})
	}
}

func TestPortFallsBackWhenUnparsable(t *testing.T) {
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("PORT", "not-a-number")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8081 {
		t.Fatalf("got port %d, want the 8081 fallback", cfg.Port)
	}
}

// A password containing DSN-significant characters must not corrupt the URL.
func TestDatabaseURLEscapesCredentials(t *testing.T) {
	cfg := &Config{DBHost: "db", DBUser: "postgres", DBPassword: "p@ss:w/rd", DBName: "test_db"}

	got := cfg.DatabaseURL()
	if strings.Contains(got, "p@ss:w/rd") {
		t.Fatalf("password was not escaped: %s", got)
	}
	if !strings.HasSuffix(got, "@db/test_db?sslmode=disable") {
		t.Fatalf("unexpected dsn shape: %s", got)
	}
}

func TestAllowedOriginsList(t *testing.T) {
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example, https://b.example ,")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"https://a.example", "https://b.example"}
	if len(cfg.AllowedOrigins) != len(want) {
		t.Fatalf("got %v, want %v", cfg.AllowedOrigins, want)
	}
	for i := range want {
		if cfg.AllowedOrigins[i] != want[i] {
			t.Fatalf("got %v, want %v", cfg.AllowedOrigins, want)
		}
	}
}
