package api

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"RegistryUI/internal/registry"
	"RegistryUI/internal/session"
)

// ---- auth ----

type loginRequest struct {
	RegistryURL string `json:"registryUrl"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type sessionResponse struct {
	RegistryURL string `json:"registryUrl"`
	Username    string `json:"username"`
}

// handleDefaults exposes the allow-listed registries the user may pick from.
// Credentials are never prefilled.
func (s *Server) handleDefaults(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"registries": s.cfg.Registries})
}

func (s *Server) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errBody("invalid request body"))
		return
	}
	req.RegistryURL = strings.TrimSpace(req.RegistryURL)
	// The registry must be one of the configured options; this prevents the
	// session client from being pointed at arbitrary (e.g. internal) URLs.
	if !s.cfg.AllowsRegistry(req.RegistryURL) {
		c.JSON(http.StatusBadRequest, errBody("unknown registry"))
		return
	}

	client := s.newRegistryClient(req.RegistryURL, req.Username, req.Password)
	if err := client.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusUnauthorized, errBody("cannot connect to registry: "+err.Error()))
		return
	}

	sess, err := s.sessions.Create(client, req.RegistryURL, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errBody("failed to create session"))
		return
	}
	s.setSessionCookie(c, sess.Token)
	c.JSON(http.StatusOK, sessionResponse{RegistryURL: sess.RegistryURL, Username: sess.Username})
}

func (s *Server) handleSession(c *gin.Context) {
	sess, ok := s.currentSession(c.Request)
	if !ok {
		c.JSON(http.StatusUnauthorized, errBody("not authenticated"))
		return
	}
	c.JSON(http.StatusOK, sessionResponse{RegistryURL: sess.RegistryURL, Username: sess.Username})
}

func (s *Server) handleLogout(c *gin.Context) {
	if cookie, err := c.Request.Cookie(session.CookieName); err == nil {
		s.sessions.Delete(cookie.Value)
	}
	s.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

func (s *Server) setSessionCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     session.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   c.Request.TLS != nil,
		Expires:  time.Now().Add(12 * time.Hour),
	})
}

func (s *Server) clearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     session.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   c.Request.TLS != nil,
		MaxAge:   -1,
	})
}

// ---- registry ----

func (s *Server) handleHealth(c *gin.Context) {
	if err := clientFrom(c).Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "unreachable", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleStats(c *gin.Context) {
	stats, err := clientFrom(c).Stats(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (s *Server) handleListRepositories(c *gin.Context) {
	repos, err := clientFrom(c).Catalog(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	if repos == nil {
		repos = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"repositories": repos})
}

func (s *Server) handleRepoSummary(c *gin.Context) {
	repo := c.Query("repo")
	if repo == "" {
		c.JSON(http.StatusBadRequest, errBody("missing required query parameter: repo"))
		return
	}
	sum, err := clientFrom(c).RepoSummary(c.Request.Context(), repo)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, sum)
}

func (s *Server) handleListTags(c *gin.Context) {
	repo := c.Query("repo")
	if repo == "" {
		c.JSON(http.StatusBadRequest, errBody("missing required query parameter: repo"))
		return
	}
	tags, err := clientFrom(c).Tags(c.Request.Context(), repo)
	if err != nil {
		writeError(c, err)
		return
	}
	if tags == nil {
		tags = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"repository": repo, "tags": tags})
}

func (s *Server) handleTagDetails(c *gin.Context) {
	repo := c.Query("repo")
	tag := c.Query("tag")
	if repo == "" || tag == "" {
		c.JSON(http.StatusBadRequest, errBody("missing required query parameters: repo and tag"))
		return
	}
	details, err := clientFrom(c).TagDetails(c.Request.Context(), repo, tag)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, details)
}

func (s *Server) handleDeleteTag(c *gin.Context) {
	repo := c.Query("repo")
	tag := c.Query("tag")
	if repo == "" || tag == "" {
		c.JSON(http.StatusBadRequest, errBody("missing required query parameters: repo and tag"))
		return
	}
	if err := clientFrom(c).DeleteTag(c.Request.Context(), repo, tag); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- helpers ----

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, registry.ErrNotFound):
		c.JSON(http.StatusNotFound, errBody("not found"))
	default:
		log.Printf("api error: %v", err)
		c.JSON(http.StatusBadGateway, errBody(err.Error()))
	}
}

func errBody(msg string) gin.H {
	return gin.H{"error": msg}
}
