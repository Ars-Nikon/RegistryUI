package config

import (
	"bufio"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// RegistryOption is one selectable registry presented on the login screen.
type RegistryOption struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Config holds the runtime configuration, populated from environment variables.
type Config struct {
	// Addr is the address the HTTP API listens on, e.g. ":8080".
	Addr string
	// RegistryURL is the default registry base URL, used as the sole fallback
	// entry when REGISTRIES is not set.
	RegistryURL string
	// Registries is the allow-list of registries the user may pick from at login.
	Registries []RegistryOption
	// RequestTimeout bounds a single upstream registry request.
	RequestTimeout time.Duration
	// AllowedOrigin is the CORS origin allowed during development (the Vite dev server).
	AllowedOrigin string
	// StaticDir is the directory of the built frontend to serve. When the
	// directory is absent (e.g. local dev with the Vite server) static serving
	// is disabled and only the API is exposed.
	StaticDir string
	// TLSCertFile / TLSKeyFile enable HTTPS when both are set; otherwise the
	// server listens on plain HTTP.
	TLSCertFile string
	TLSKeyFile  string
}

// TLSEnabled reports whether HTTPS is configured (both cert and key are set).
func (c Config) TLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

// Load reads configuration from the environment, applying sensible defaults.
// A local .env file (if present) is loaded first for convenience; it is
// gitignored and never overrides variables already set in the environment.
func Load() Config {
	loadDotEnv(".env")
	registryURL := env("REGISTRY_URL", "http://localhost:5000", nil)
	return Config{
		Addr:           env("PORT", ":8080", withColon),
		RegistryURL:    registryURL,
		Registries:     parseRegistries(os.Getenv("REGISTRIES"), registryURL),
		RequestTimeout: envDuration("REGISTRY_TIMEOUT", 15*time.Second),
		AllowedOrigin:  env("CORS_ORIGIN", "http://localhost:5173", nil),
		StaticDir:      env("STATIC_DIR", "web/dist", nil),
		TLSCertFile:    env("TLS_CERT_FILE", "", nil),
		TLSKeyFile:     env("TLS_KEY_FILE", "", nil),
	}
}

// AllowsRegistry reports whether the given URL is one of the configured
// registries. Login is restricted to this allow-list.
func (c Config) AllowsRegistry(registryURL string) bool {
	for _, r := range c.Registries {
		if r.URL == registryURL {
			return true
		}
	}
	return false
}

// parseRegistries builds the selectable registry list from REGISTRIES, given as
// comma-separated "Name=URL" pairs (the name is optional and defaults to the URL
// host). When REGISTRIES is empty it falls back to a single entry from fallbackURL.
func parseRegistries(raw, fallbackURL string) []RegistryOption {
	var opts []RegistryOption
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, urlStr, ok := strings.Cut(part, "=")
		if !ok {
			urlStr, name = name, ""
		}
		name = strings.TrimSpace(name)
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			continue
		}
		if name == "" {
			name = registryName(urlStr)
		}
		opts = append(opts, RegistryOption{Name: name, URL: urlStr})
	}
	if len(opts) == 0 && fallbackURL != "" {
		opts = append(opts, RegistryOption{Name: registryName(fallbackURL), URL: fallbackURL})
	}
	return opts
}

// registryName derives a display name from a registry URL (its host).
func registryName(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		return u.Host
	}
	return rawURL
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
