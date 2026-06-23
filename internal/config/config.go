package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type RegistryOption struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// RegistryOptions implements envconfig.Decoder so REGISTRIES can be given as
// comma-separated "Name=URL" pairs.
type RegistryOptions []RegistryOption

func (o *RegistryOptions) Decode(value string) error {
	*o = parseRegistries(value)
	return nil
}

type Config struct {
	Addr           string          `envconfig:"PORT" default:":8080"`
	RegistryURL    string          `envconfig:"REGISTRY_URL" default:"http://localhost:5000"`
	Registries     RegistryOptions `envconfig:"REGISTRIES"`
	RequestTimeout time.Duration   `ignored:"true" default:"15s"`
	AllowedOrigin  string          `envconfig:"CORS_ORIGIN" default:"http://localhost:5173"`
	StaticDir      string          `envconfig:"STATIC_DIR" default:"web/dist"`
	TLSCertFile    string          `envconfig:"TLS_CERT_FILE"`
	TLSKeyFile     string          `envconfig:"TLS_KEY_FILE"`
	JwtSecret      string          `envconfig:"JWT_SECRET"`
	JwtTTL         time.Duration   `envconfig:"JWT_TTL" default:"24h"`
	JwtIssuer      string          `envconfig:"JWT_ISSUER" default:"registryui"`
}

func (c Config) TLSEnabled() bool {
	return c.TLSCertFile != "" && c.TLSKeyFile != ""
}

func Load() Config {
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		panic(err)
	}

	cfg.Addr = withColon(cfg.Addr)
	cfg.RequestTimeout = envDuration("REGISTRY_TIMEOUT", 15*time.Second)
	if len(cfg.Registries) == 0 && cfg.RegistryURL != "" {
		cfg.Registries = RegistryOptions{{Name: registryName(cfg.RegistryURL), URL: cfg.RegistryURL}}
	}

	return cfg
}

func (c Config) AllowsRegistry(registryURL string) bool {
	for _, r := range c.Registries {
		if r.URL == registryURL {
			return true
		}
	}
	return false
}

func parseRegistries(raw string) []RegistryOption {
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
	return opts
}

func registryName(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		return u.Host
	}
	return rawURL
}

func envDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	return def
}

func withColon(v string) string {
	if v == "" || v[0] == ':' {
		return v
	}
	if _, err := strconv.Atoi(v); err == nil {
		return ":" + v
	}
	return v
}
