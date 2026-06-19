package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the runtime configuration, populated from environment variables.
type Config struct {
	// Addr is the address the HTTP API listens on, e.g. ":8080".
	Addr string
	// RegistryURL is the base URL of the Docker Registry v2, e.g. "http://localhost:5000".
	RegistryURL string
	// RegistryUser / RegistryPassword are optional basic-auth credentials.
	RegistryUser     string
	RegistryPassword string
	// RequestTimeout bounds a single upstream registry request.
	RequestTimeout time.Duration
	// AllowedOrigin is the CORS origin allowed during development (the Vite dev server).
	AllowedOrigin string
}

// Load reads configuration from the environment, applying sensible defaults.
// A local .env file (if present) is loaded first for convenience; it is
// gitignored and never overrides variables already set in the environment.
func Load() Config {
	loadDotEnv(".env")
	return Config{
		Addr:             env("PORT", ":8080", withColon),
		RegistryURL:      env("REGISTRY_URL", "http://localhost:5000", nil),
		RegistryUser:     env("REGISTRY_USERNAME", "registry", nil),
		RegistryPassword: env("REGISTRY_PASSWORD", "", nil),
		RequestTimeout:   envDuration("REGISTRY_TIMEOUT", 15*time.Second),
		AllowedOrigin:    env("CORS_ORIGIN", "http://localhost:5173", nil),
	}
}

// loadDotEnv reads simple KEY=VALUE lines from path into the process
// environment, skipping blanks, comments and keys that are already set.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
}

func env(key, def string, transform func(string) string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		v = def
	}
	if transform != nil {
		v = transform(v)
	}
	return v
}

func envDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	// Allow a bare number of seconds, e.g. "30".
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	return def
}

// withColon lets PORT be given as either "8080" or ":8080".
func withColon(v string) string {
	if v == "" || v[0] == ':' {
		return v
	}
	if _, err := strconv.Atoi(v); err == nil {
		return ":" + v
	}
	return v
}
