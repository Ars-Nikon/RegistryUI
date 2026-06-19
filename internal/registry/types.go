package registry

import "time"

// Media types understood by this client (Docker schema2 and OCI).
const (
	MediaTypeManifestV2   = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	MediaTypeOCIManifest  = "application/vnd.oci.image.manifest.v1+json"
	MediaTypeOCIIndex     = "application/vnd.oci.image.index.v1+json"
)

// Descriptor is an OCI content descriptor referencing a blob or manifest.
type Descriptor struct {
	MediaType string    `json:"mediaType"`
	Digest    string    `json:"digest"`
	Size      int64     `json:"size"`
	Platform  *Platform `json:"platform,omitempty"`
}

// Platform identifies the OS/architecture of an image in a multi-arch index.
type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant,omitempty"`
}

// manifest is the wire representation of a schema2 / OCI image manifest or index.
type manifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Config        Descriptor   `json:"config"`
	Layers        []Descriptor `json:"layers"`
	// Manifests is populated only for manifest lists / image indexes.
	Manifests []Descriptor `json:"manifests"`
}

// imageConfig is the subset of the image config blob we surface to the UI.
type imageConfig struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Created      string `json:"created"`
	Config       struct {
		Entrypoint []string          `json:"Entrypoint"`
		Cmd        []string          `json:"Cmd"`
		WorkingDir string            `json:"WorkingDir"`
		Env        []string          `json:"Env"`
		Labels     map[string]string `json:"Labels"`
	} `json:"config"`
	History []struct {
		CreatedBy  string `json:"created_by"`
		EmptyLayer bool   `json:"empty_layer"`
	} `json:"history"`
}

// LayerInfo describes a single image layer for the UI.
type LayerInfo struct {
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	CreatedBy string `json:"createdBy"`
}

// PlatformInfo is a platform entry of a manifest list with its image size.
type PlatformInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant,omitempty"`
	Digest       string `json:"digest"`
	Size         int64  `json:"size"`
}

// TagDetails is the aggregated, UI-friendly view of a single tag.
type TagDetails struct {
	Name         string            `json:"name"`
	Digest       string            `json:"digest"`
	MediaType    string            `json:"mediaType"`
	Size         int64             `json:"size"`
	Created      *time.Time        `json:"created,omitempty"`
	Architecture string            `json:"architecture,omitempty"`
	OS           string            `json:"os,omitempty"`
	Entrypoint   []string          `json:"entrypoint,omitempty"`
	Cmd          []string          `json:"cmd,omitempty"`
	WorkingDir   string            `json:"workingDir,omitempty"`
	Env          []string          `json:"env,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Layers       []LayerInfo       `json:"layers,omitempty"`
	IsIndex      bool              `json:"isIndex"`
	Platforms    []PlatformInfo    `json:"platforms,omitempty"`
}

// RepoSummary is the lightweight catalog-card view of a repository.
type RepoSummary struct {
	Name        string     `json:"name"`
	TagCount    int        `json:"tagCount"`
	Size        int64      `json:"size"`
	Updated     *time.Time `json:"updated,omitempty"`
	Description string     `json:"description,omitempty"`
}

// Stats is the aggregate sidebar summary across the whole registry.
type Stats struct {
	Repositories int   `json:"repositories"`
	Tags         int   `json:"tags"`
	Storage      int64 `json:"storage"`
}
