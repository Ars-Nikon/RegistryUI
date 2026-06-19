package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

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

// handleDefaults exposes the env-configured registry URL/user to prefill the login form.
func (s *Server) handleDefaults(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"registryUrl": s.cfg.RegistryURL,
		"username":    s.cfg.RegistryUser,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid request body"))
		return
	}
	req.RegistryURL = strings.TrimSpace(req.RegistryURL)
	if req.RegistryURL == "" {
		req.RegistryURL = s.cfg.RegistryURL
	}
	if req.RegistryURL == "" {
		writeJSON(w, http.StatusBadRequest, errBody("registry URL is required"))
		return
	}

	client := s.newRegistryClient(req.RegistryURL, req.Username, req.Password)
	if err := client.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusUnauthorized, errBody("cannot connect to registry: "+err.Error()))
		return
	}

	sess, err := s.sessions.Create(client, req.RegistryURL, req.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody("failed to create session"))
		return
	}
	s.setSessionCookie(w, r, sess.Token)
	writeJSON(w, http.StatusOK, sessionResponse{RegistryURL: sess.RegistryURL, Username: sess.Username})
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.currentSession(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("not authenticated"))
		return
	}
	writeJSON(w, http.StatusOK, sessionResponse{RegistryURL: sess.RegistryURL, Username: sess.Username})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(session.CookieName); err == nil {
		s.sessions.Delete(cookie.Value)
	}
	s.clearSessionCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		Expires:  time.Now().Add(12 * time.Hour),
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   -1,
	})
}

// ---- registry ----

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := clientFrom(r).Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": "unreachable", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := clientFrom(r).Stats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleListRepositories(w http.ResponseWriter, r *http.Request) {
	repos, err := clientFrom(r).Catalog(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	if repos == nil {
		repos = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"repositories": repos})
}

func (s *Server) handleRepoSummary(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		writeJSON(w, http.StatusBadRequest, errBody("missing required query parameter: repo"))
		return
	}
	sum, err := clientFrom(r).RepoSummary(r.Context(), repo)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		writeJSON(w, http.StatusBadRequest, errBody("missing required query parameter: repo"))
		return
	}
	tags, err := clientFrom(r).Tags(r.Context(), repo)
	if err != nil {
		writeError(w, err)
		return
	}
	if tags == nil {
		tags = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"repository": repo, "tags": tags})
}

func (s *Server) handleTagDetails(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	tag := r.URL.Query().Get("tag")
	if repo == "" || tag == "" {
		writeJSON(w, http.StatusBadRequest, errBody("missing required query parameters: repo and tag"))
		return
	}
	details, err := clientFrom(r).TagDetails(r.Context(), repo, tag)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, details)
}

func (s *Server) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	tag := r.URL.Query().Get("tag")
	if repo == "" || tag == "" {
		writeJSON(w, http.StatusBadRequest, errBody("missing required query parameters: repo and tag"))
		return
	}
	if err := clientFrom(r).DeleteTag(r.Context(), repo, tag); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, registry.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errBody("not found"))
	default:
		log.Printf("api error: %v", err)
		writeJSON(w, http.StatusBadGateway, errBody(err.Error()))
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func errBody(msg string) map[string]any {
	return map[string]any{"error": msg}
}
