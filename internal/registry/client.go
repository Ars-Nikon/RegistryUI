package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrNotFound is returned when the registry responds with 404 for a resource.
var ErrNotFound = errors.New("registry: resource not found")

// manifestAccept advertises every manifest type we can parse so the registry
// returns the concrete one rather than negotiating down to schema1.
var manifestAccept = strings.Join([]string{
	MediaTypeManifestV2,
	MediaTypeManifestList,
	MediaTypeOCIManifest,
	MediaTypeOCIIndex,
}, ", ")

// Client talks to a Docker Registry v2 HTTP API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

// NewClient builds a registry client for baseURL. username/password are optional.
func NewClient(baseURL, username, password string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
		username:   username,
		password:   password,
	}
}

// BaseURL returns the registry base URL this client targets.
func (c *Client) BaseURL() string { return c.baseURL }

// Username returns the configured user (may be empty).
func (c *Client) Username() string { return c.username }

func (c *Client) newRequest(ctx context.Context, method, path, accept string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	return req, nil
}

// do executes the request and maps non-2xx statuses to errors.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry request failed: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		resp.Body.Close()
		return nil, fmt.Errorf("registry returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

// Ping verifies the registry implements the v2 API and is reachable.
func (c *Client) Ping(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodGet, "/v2/", "")
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Catalog returns every repository name, following pagination.
func (c *Client) Catalog(ctx context.Context) ([]string, error) {
	const pageSize = 100
	var repos []string
	last := ""
	for {
		path := fmt.Sprintf("/v2/_catalog?n=%d", pageSize)
		if last != "" {
			path += "&last=" + url.QueryEscape(last)
		}
		req, err := c.newRequest(ctx, http.MethodGet, path, "")
		if err != nil {
			return nil, err
		}
		resp, err := c.do(req)
		if err != nil {
			return nil, err
		}
		var page struct {
			Repositories []string `json:"repositories"`
		}
		err = json.NewDecoder(resp.Body).Decode(&page)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode catalog: %w", err)
		}
		repos = append(repos, page.Repositories...)
		if len(page.Repositories) < pageSize {
			break
		}
		last = page.Repositories[len(page.Repositories)-1]
	}
	sort.Strings(repos)
	return repos, nil
}

// Tags returns the tag names for a repository. A repository with no tags yields
// an empty slice rather than an error.
func (c *Client) Tags(ctx context.Context, repo string) ([]string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/v2/"+repo+"/tags/list", "")
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var res struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode tags: %w", err)
	}
	sort.Strings(res.Tags)
	return res.Tags, nil
}

// rawManifest is a manifest fetched together with its resolved digest and type.
type rawManifest struct {
	body        []byte
	digest      string
	contentType string
	parsed      manifest
}

func (c *Client) fetchManifest(ctx context.Context, repo, ref string) (*rawManifest, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/v2/"+repo+"/manifests/"+ref, manifestAccept)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	rm := &rawManifest{
		digest:      resp.Header.Get("Docker-Content-Digest"),
		contentType: resp.Header.Get("Content-Type"),
	}
	rm.body, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	if err := json.Unmarshal(rm.body, &rm.parsed); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return rm, nil
}

// imageSize returns the total blob size of a non-index image manifest.
func imageSize(m manifest) int64 {
	total := m.Config.Size
	for _, l := range m.Layers {
		total += l.Size
	}
	return total
}

// TagDetails fetches and aggregates everything the UI shows for a single tag.
func (c *Client) TagDetails(ctx context.Context, repo, tag string) (*TagDetails, error) {
	rm, err := c.fetchManifest(ctx, repo, tag)
	if err != nil {
		return nil, err
	}
	m := rm.parsed

	details := &TagDetails{
		Name:      tag,
		Digest:    rm.digest,
		MediaType: rm.contentType,
	}

	if isIndex(rm.contentType, m.MediaType) {
		details.IsIndex = true
		for _, d := range m.Manifests {
			p := PlatformInfo{Digest: d.Digest, Size: d.Size}
			if d.Platform != nil {
				p.OS = d.Platform.OS
				p.Architecture = d.Platform.Architecture
				p.Variant = d.Platform.Variant
			}
			// Resolve the child image's true total size; fall back to the
			// descriptor size if the child cannot be read.
			if child, err := c.fetchManifest(ctx, repo, d.Digest); err == nil && !isIndex(child.contentType, child.parsed.MediaType) {
				p.Size = imageSize(child.parsed)
			}
			details.Size += p.Size
			details.Platforms = append(details.Platforms, p)
		}
		return details, nil
	}

	details.Size = imageSize(m)

	cfg, _ := c.configBlob(ctx, repo, m.Config.Digest)
	if cfg != nil {
		details.Architecture = cfg.Architecture
		details.OS = cfg.OS
		details.Entrypoint = cfg.Config.Entrypoint
		details.Cmd = cfg.Config.Cmd
		details.WorkingDir = cfg.Config.WorkingDir
		details.Env = cfg.Config.Env
		if len(cfg.Config.Labels) > 0 {
			details.Labels = cfg.Config.Labels
		}
		if t, err := time.Parse(time.RFC3339Nano, cfg.Created); err == nil {
			details.Created = &t
		}
	}

	details.Layers = buildLayers(m.Layers, cfg)
	return details, nil
}

// buildLayers zips the manifest's physical layers with the config history to
// attach the "created by" command to each layer.
func buildLayers(layers []Descriptor, cfg *imageConfig) []LayerInfo {
	var createdBy []string
	if cfg != nil {
		for _, h := range cfg.History {
			if !h.EmptyLayer {
				createdBy = append(createdBy, h.CreatedBy)
			}
		}
	}
	out := make([]LayerInfo, len(layers))
	for i, l := range layers {
		li := LayerInfo{Digest: l.Digest, Size: l.Size}
		if i < len(createdBy) {
			li.CreatedBy = createdBy[i]
		}
		out[i] = li
	}
	return out
}

// configBlob fetches the image config blob and parses the fields we expose.
func (c *Client) configBlob(ctx context.Context, repo, digest string) (*imageConfig, error) {
	if digest == "" {
		return nil, errors.New("registry: empty config digest")
	}
	req, err := c.newRequest(ctx, http.MethodGet, "/v2/"+repo+"/blobs/"+digest, "")
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cfg imageConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode image config: %w", err)
	}
	return &cfg, nil
}

// RepoSummary returns the catalog-card view of a repository: tag count, size of
// the most recent tag, last-updated time and an optional description label.
func (c *Client) RepoSummary(ctx context.Context, repo string) (*RepoSummary, error) {
	tags, err := c.Tags(ctx, repo)
	if err != nil {
		return nil, err
	}
	sum := &RepoSummary{Name: repo, TagCount: len(tags)}
	if len(tags) == 0 {
		return sum, nil
	}
	// Inspect the most recently created tag for size, update time and labels.
	var newest *TagDetails
	for _, tag := range tags {
		d, err := c.TagDetails(ctx, repo, tag)
		if err != nil {
			continue
		}
		if newest == nil || (d.Created != nil && (newest.Created == nil || d.Created.After(*newest.Created))) {
			newest = d
		}
	}
	if newest != nil {
		sum.Size = newest.Size
		sum.Updated = newest.Created
		if desc := newest.Labels["org.opencontainers.image.description"]; desc != "" {
			sum.Description = desc
		}
	}
	return sum, nil
}

// Stats aggregates repository, tag and (deduplicated) storage totals.
func (c *Client) Stats(ctx context.Context) (*Stats, error) {
	repos, err := c.Catalog(ctx)
	if err != nil {
		return nil, err
	}
	stats := &Stats{Repositories: len(repos)}

	var (
		mu        sync.Mutex
		seenBlobs = map[string]int64{}
		sem       = make(chan struct{}, 8)
		wg        sync.WaitGroup
	)
	for _, repo := range repos {
		tags, err := c.Tags(ctx, repo)
		if err != nil {
			continue
		}
		stats.Tags += len(tags)
		for _, tag := range tags {
			wg.Add(1)
			sem <- struct{}{}
			go func(repo, tag string) {
				defer wg.Done()
				defer func() { <-sem }()
				c.collectBlobs(ctx, repo, tag, &mu, seenBlobs)
			}(repo, tag)
		}
	}
	wg.Wait()

	for _, size := range seenBlobs {
		stats.Storage += size
	}
	return stats, nil
}

// collectBlobs records the unique config/layer blob sizes referenced by a tag.
func (c *Client) collectBlobs(ctx context.Context, repo, tag string, mu *sync.Mutex, seen map[string]int64) {
	rm, err := c.fetchManifest(ctx, repo, tag)
	if err != nil {
		return
	}
	record := func(m manifest) {
		mu.Lock()
		defer mu.Unlock()
		if m.Config.Digest != "" {
			seen[m.Config.Digest] = m.Config.Size
		}
		for _, l := range m.Layers {
			seen[l.Digest] = l.Size
		}
	}
	if isIndex(rm.contentType, rm.parsed.MediaType) {
		for _, d := range rm.parsed.Manifests {
			if child, err := c.fetchManifest(ctx, repo, d.Digest); err == nil {
				record(child.parsed)
			}
		}
		return
	}
	record(rm.parsed)
}

// Digest resolves the content digest for a tag without downloading the manifest body.
func (c *Client) Digest(ctx context.Context, repo, tag string) (string, error) {
	req, err := c.newRequest(ctx, http.MethodHead, "/v2/"+repo+"/manifests/"+tag, manifestAccept)
	if err != nil {
		return "", err
	}
	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return "", errors.New("registry: missing Docker-Content-Digest header")
	}
	return digest, nil
}

// DeleteTag deletes the manifest a tag points to. The registry must be started
// with REGISTRY_STORAGE_DELETE_ENABLED=true, otherwise it responds 405.
func (c *Client) DeleteTag(ctx context.Context, repo, tag string) error {
	digest, err := c.Digest(ctx, repo, tag)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodDelete, "/v2/"+repo+"/manifests/"+digest, manifestAccept)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func isIndex(contentType, manifestMediaType string) bool {
	for _, t := range []string{contentType, manifestMediaType} {
		if strings.HasPrefix(t, MediaTypeManifestList) || strings.HasPrefix(t, MediaTypeOCIIndex) {
			return true
		}
	}
	return false
}
