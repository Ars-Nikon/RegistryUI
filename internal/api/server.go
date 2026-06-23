package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"RegistryUI/internal/auth"
	"RegistryUI/internal/config"
	"RegistryUI/internal/registry"
	"RegistryUI/internal/session"
)

const (
	cookieName = "Authorization"
	sessionKey = "session"
)

// userSession is the per-login state kept server-side: the registry
// credentials, including the password, which never reaches the browser. A
// fresh registry.Client is built from these on each request.
type userSession struct {
	RegistryURL string
	Username    string
	Password    string
}

type Server struct {
	cfg      config.Config
	auth     *auth.Service
	sessions *session.Store[*userSession]
}

func NewServer(cfg config.Config, authSvc *auth.Service) *Server {
	return &Server{
		cfg:      cfg,
		auth:     authSvc,
		sessions: session.NewStore[*userSession](cfg.JwtTTL),
	}
}

func (s *Server) Engine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), s.cors())

	// Auth + bootstrap (no auth required).
	r.GET("/api/defaults", s.handleDefaults)
	r.POST("/api/session", s.handleLogin)
	r.GET("/api/session", s.handleSession)
	r.DELETE("/api/session", s.handleLogout)

	// Registry endpoints (auth required). Docker repository names are
	// hierarchical, so repo and tag travel as query parameters.
	api := r.Group("/api", s.requireAuth)
	api.GET("/health", s.handleHealth)
	api.GET("/stats", s.handleStats)
	api.GET("/repositories", s.handleListRepositories)
	api.GET("/repository", s.handleRepoSummary)
	api.GET("/tags", s.handleListTags)
	api.GET("/tag", s.handleTagDetails)
	api.DELETE("/tag", s.handleDeleteTag)

	r.NoRoute(s.serveStatic)

	return r
}

func (s *Server) requireAuth(c *gin.Context) {
	sess, ok := s.currentSession(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errBody("not authenticated"))
		return
	}
	c.Set(sessionKey, sess)
	c.Next()
}
func (s *Server) currentSession(c *gin.Context) (*userSession, bool) {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return nil, false
	}
	identity, err := s.auth.Decode(cookie)
	if err != nil {
		return nil, false
	}
	return s.sessions.Get(identity.SessionID)
}

// clientFrom builds a registry client from the current session's stored
// credentials. A new client is created per request; the connection pool is not
// shared across requests.
func (s *Server) clientFrom(c *gin.Context) *registry.Client {
	us := c.MustGet(sessionKey).(*userSession)
	return registry.NewClient(us.RegistryURL, us.Username, us.Password, s.cfg.RequestTimeout)
}

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
