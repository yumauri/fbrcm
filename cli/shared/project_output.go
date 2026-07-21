package shared

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

// ProjectJSON is the shared machine-readable representation of a project.
type ProjectJSON struct {
	Project      string   `json:"project"`
	ProjectID    string   `json:"project_id"`
	Number       string   `json:"number,omitempty"`
	State        string   `json:"state,omitempty"`
	ETag         string   `json:"etag,omitempty"`
	AuthID       string   `json:"auth_id"`
	Disabled     bool     `json:"disabled"`
	DiscoveredBy []string `json:"discovered_by,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
	SyncedAt     string   `json:"synced_at,omitempty"`
	URL          string   `json:"url,omitempty"`
}

// NewProjectJSON copies a project into its CLI JSON representation.
func NewProjectJSON(project core.Project, withURL bool) ProjectJSON {
	out := ProjectJSON{
		Project:      project.Name,
		ProjectID:    project.ProjectID,
		Number:       project.ProjectNumber,
		State:        project.State,
		ETag:         project.ETag,
		AuthID:       project.AuthID,
		Disabled:     project.Disabled,
		DiscoveredBy: append([]string(nil), project.DiscoveredBy...),
		UpdatedAt:    project.UpdatedAt,
		SyncedAt:     project.SyncedAt,
	}
	if withURL {
		out.URL = firebase.RemoteConfigConsoleURL(project.ProjectID)
	}
	return out
}
