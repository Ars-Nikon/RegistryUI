package api

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"RegistryUI/internal/auth"
	"RegistryUI/internal/registry"
)

type loginRequest struct {
	RegistryURL string `json:"registryUrl"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type sessionResponse struct {
	RegistryURL string `json:"registryUrl"`
	Username    string `json:"username"`
}

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

	if !s.cfg.AllowsRegistry(req.RegistryURL) {
		c.JSON(http.StatusBadRequest, errBody("unknown registry"))
		return
	}

	timeout := s.cfg.RequestTimeout
	client := registry.NewClient(req.RegistryURL, req.Username, req.Password, timeout)

	if err := client.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusUnauthorized, errBody("cannot connect to registry: "+err.Error()))
		return
	}

	id, err := s.sessions.Create(&userSession{
		RegistryURL: req.RegistryURL,
		Username:    req.Username,
		Password:    req.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errBody("failed to create session"))
		return
	}

	token, err := s.auth.Generate(auth.Identity{
		UserName:    req.Username,
		RegistryURL: req.RegistryURL,
		SessionID:   id,
	})
	if err != nil {
		s.sessions.Delete(id)
		c.JSON(http.StatusInternalServerError, errBody("failed to issue token"))
		return
	}

	s.setAuthCookie(c, token)
	c.JSON(http.StatusOK, sessionResponse{RegistryURL: req.RegistryURL, Username: req.Username})
}

func (s *Server) handleSession(c *gin.Context) {
	sess, ok := s.currentSession(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errBody("not authenticated"))
		return
	}
	c.JSON(http.StatusOK, sessionResponse{RegistryURL: sess.RegistryURL, Username: sess.Username})
}

func (s *Server) handleLogout(c *gin.Context) {
	if cookie, err := c.Cookie(cookieName); err == nil {
		if identity, err := s.auth.Decode(cookie); err == nil {
			s.sessions.Delete(identity.SessionID)
		}
	}
	s.clearAuthCookie(c)
	c.Status(http.StatusNoContent)
}

func (s *Server) setAuthCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   c.Request.TLS != nil,
		Expires:  time.Now().Add(s.cfg.JwtTTL),
	})
}

func (s *Server) clearAuthCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
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
	if err := s.clientFrom(c).Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "unreachable", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleStats(c *gin.Context) {
	stats, err := s.clientFrom(c).Stats(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (s *Server) handleListRepositories(c *gin.Context) {
	repos, err := s.clientFrom(c).Catalog(c.Request.Context())
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
	sum, err := s.clientFrom(c).RepoSummary(c.Request.Context(), repo)
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
	tags, err := s.clientFrom(c).Tags(c.Request.Context(), repo)
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
	details, err := s.clientFrom(c).TagDetails(c.Request.Context(), repo, tag)
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
	if err := s.clientFrom(c).DeleteTag(c.Request.Context(), repo, tag); err != nil {
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
