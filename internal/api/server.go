package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"RegistryUI/internal/config"
	"RegistryUI/internal/registry"
	"RegistryUI/internal/session"
)

// sessionKey is the gin context key under which the resolved session is stored.
const sessionKey = "session"

// Server wires session-scoped registry clients into a gin engine.
type Server struct {
	cfg      config.Config
	sessions *session.Store
}

// NewServer builds the API server.
func NewServer(cfg config.Config, sessions *session.Store) *Server {
	return &Server{cfg: cfg, sessions: sessions}
}

// Engine builds and returns the configured gin engine.
func (s *Server) Engine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), s.cors())

	// Auth + bootstrap (no session required).
	r.GET("/api/defaults", s.handleDefaults)
	r.POST("/api/session", s.handleLogin)
	r.GET("/api/session", s.handleSession)
	r.DELETE("/api/session", s.handleLogout)

	// Registry endpoints (session required). Docker repository names are
	// hierarchical, so repo and tag travel as query parameters.
	api := r.Group("/api", s.requireSession)
	api.GET("/health", s.handleHealth)
	api.GET("/stats", s.handleStats)
	api.GET("/repositories", s.handleListRepositories)
	api.GET("/repository", s.handleRepoSummary)
	api.GET("/tags", s.handleListTags)
	api.GET("/tag", s.handleTagDetails)
	api.DELETE("/tag", s.handleDeleteTag)

	// Serve the built frontend (single-binary / Docker deploy) with SPA fallback.
	r.NoRoute(s.serveStatic)

	return r
}

// requireSession resolves the session cookie to a registry client and stores it
// in the gin context, or aborts the request with 401.
func (s *Server) requireSession(c *gin.Context) {
	sess, ok := s.currentSession(c.Request)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errBody("not authenticated"))
		return
	}
	c.Set(sessionKey, sess)
	c.Next()
}

func (s *Server) currentSession(r *http.Request) (*session.Session, bool) {
	cookie, err := r.Cookie(session.CookieName)
	if err != nil {
		return nil, false
	}
	return s.sessions.Get(cookie.Value)
}

// clientFrom returns the session's registry client from the gin context.
func clientFrom(c *gin.Context) *registry.Client {
	return c.MustGet(sessionKey).(*session.Session).Client
}

// cors allows the Vite dev server to call the API cross-origin with cookies.
func (s *Server) cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", s.cfg.AllowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Vary", "Origin")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// serveStatic serves the built frontend from cfg.StaticDir with SPA fallback:
// unknown, non-API paths return index.html so client-side routing works on deep
// links and refreshes.
func (s *Server) serveStatic(c *gin.Context) {
	path := c.Request.URL.Path
	// Unmatched API paths are genuine 404s, never the SPA.
	if strings.HasPrefix(path, "/api/") {
		c.JSON(http.StatusNotFound, errBody("not found"))
		return
	}
	dir := s.cfg.StaticDir
	if dir == "" {
		c.Status(http.StatusNotFound)
		return
	}
	target := filepath.Join(dir, filepath.Clean(path))
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		c.File(target)
		return
	}
	index := filepath.Join(dir, "index.html")
	if _, err := os.Stat(index); err == nil {
		c.File(index)
		return
	}
	c.Status(http.StatusNotFound)
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
