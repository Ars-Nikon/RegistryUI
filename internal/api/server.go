package api

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"RegistryUI/internal/config"
	"RegistryUI/internal/registry"
	"RegistryUI/internal/session"
)

type ctxKey int

const sessionKey ctxKey = 0

// Server wires session-scoped registry clients into an HTTP handler.
type Server struct {
	cfg      config.Config
	sessions *session.Store
}

// NewServer builds the API server.
func NewServer(cfg config.Config, sessions *session.Store) *Server {
	return &Server{cfg: cfg, sessions: sessions}
}

// Handler returns the root http.Handler with routes and middleware applied.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Auth + bootstrap (no session required).
	mux.HandleFunc("GET /api/defaults", s.handleDefaults)
	mux.HandleFunc("POST /api/session", s.handleLogin)
	mux.HandleFunc("GET /api/session", s.handleSession)
	mux.HandleFunc("DELETE /api/session", s.handleLogout)

	// Registry endpoints (session required). Docker repository names are
	// hierarchical, so repo and tag travel as query parameters.
	mux.Handle("GET /api/health", s.requireSession(http.HandlerFunc(s.handleHealth)))
	mux.Handle("GET /api/stats", s.requireSession(http.HandlerFunc(s.handleStats)))
	mux.Handle("GET /api/repositories", s.requireSession(http.HandlerFunc(s.handleListRepositories)))
	mux.Handle("GET /api/repository", s.requireSession(http.HandlerFunc(s.handleRepoSummary)))
	mux.Handle("GET /api/tags", s.requireSession(http.HandlerFunc(s.handleListTags)))
	mux.Handle("GET /api/tag", s.requireSession(http.HandlerFunc(s.handleTagDetails)))
	mux.Handle("DELETE /api/tag", s.requireSession(http.HandlerFunc(s.handleDeleteTag)))

	// Serve the built frontend (single-binary / Docker deploy). When the
	// directory is absent (local dev with the Vite server) the catch-all route
	// is not registered and only the API is exposed.
	if h := s.staticHandler(); h != nil {
		mux.Handle("/", h)
	}

	return s.withCORS(mux)
}

// staticHandler serves the built frontend from cfg.StaticDir with SPA
// fallback: unknown, non-API paths return index.html so client-side routing
// works on deep links and refreshes. Returns nil when the directory is missing.
func (s *Server) staticHandler() http.Handler {
	dir := s.cfg.StaticDir
	if dir == "" {
		return nil
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil
	}
	fileServer := http.FileServer(http.Dir(dir))
	indexPath := filepath.Join(dir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API paths never fall through to the SPA.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		target := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}

// requireSession resolves the session cookie to a registry client and stores it
// in the request context, or rejects the request with 401.
func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := s.currentSession(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errBody("not authenticated"))
			return
		}
		ctx := context.WithValue(r.Context(), sessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) currentSession(r *http.Request) (*session.Session, bool) {
	cookie, err := r.Cookie(session.CookieName)
	if err != nil {
		return nil, false
	}
	return s.sessions.Get(cookie.Value)
}

// clientFrom returns the session's registry client from the request context.
func clientFrom(r *http.Request) *registry.Client {
	return r.Context().Value(sessionKey).(*session.Session).Client
}

// withCORS allows the Vite dev server to call the API cross-origin with cookies.
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.cfg.AllowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// newRegistryClient builds a registry client for the given connection details,
// falling back to the configured timeout.
func (s *Server) newRegistryClient(registryURL, username, password string) *registry.Client {
	timeout := s.cfg.RequestTimeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	return registry.NewClient(registryURL, username, password, timeout)
}
