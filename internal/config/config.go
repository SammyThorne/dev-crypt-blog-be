// Package config loads the server's runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds every setting the server needs to start.
type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string

	Port   int
	UseTLS bool

	TLSCertPath string
	TLSKeyPath  string

	FirebaseAPIKey string

	// BlogsJSONPath is the file served at /blogs.json.
	BlogsJSONPath string

	// AllowedOrigins is the CORS origin allowlist; a single "*" allows any origin.
	AllowedOrigins []string
}

// getEnvDefault reads an environment variable, falling back to a default if unset
// (keeps production behaviour unchanged unless the corresponding env var is set).
func getEnvDefault(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}

// Load reads the configuration from the environment. DB host/user/name and TLS
// settings default to the production values so deploying with no environment
// variables set behaves as before.
func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnvDefault("PORT", "8081"))
	if err != nil {
		port = 8081
	}

	origins := []string{"*"}
	if raw := getEnvDefault("CORS_ALLOWED_ORIGINS", "*"); raw != "*" {
		origins = splitAndTrim(raw)
	}

	return &Config{
		DBHost:     getEnvDefault("DB_HOST", "localhost"),
		DBUser:     getEnvDefault("DB_USER", "postgres"),
		DBPassword: getEnvDefault("DB_PASSWORD", "postgres"),
		DBName:     getEnvDefault("DB_NAME", "test_db"),

		Port:   port,
		UseTLS: getEnvDefault("USE_TLS", "true") != "false",

		TLSCertPath: getEnvDefault("TLS_CERT_PATH", "/etc/letsencrypt/live/srv915664.hstgr.cloud/cert.pem"),
		TLSKeyPath:  getEnvDefault("TLS_KEY_PATH", "/etc/letsencrypt/live/srv915664.hstgr.cloud/privkey.pem"),

		FirebaseAPIKey: getEnvDefault("FIREBASE_API_KEY", "AIzaSyCw50RiNb7HeK9_-fDpzGqVGDcPFC4U0JI"),

		BlogsJSONPath: getEnvDefault("BLOGS_JSON_PATH", "public/blogs.json"),

		AllowedOrigins: origins,
	}, nil
}

// DatabaseURL renders the connection settings as a libpq-style URL for pgx.
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		url(c.DBUser), url(c.DBPassword), c.DBHost, c.DBName)
}
