package parameters

import (
	"strings"

	"charm.land/lipgloss/v2"

	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) layoutForProject(projectID string) projectLayout {
	layout := projectLayout{}

	project := m.projectByID(projectID)
	if project == nil {
		return layout
	}

	metadata := m.projectMeta(project, false)
	layout.metadataWidth = lipgloss.Width(metadata)
	return layout
}

func (m Model) projectMeta(project *projectState, selected bool) string {
	if project == nil {
		return ""
	}

	badge, rest := m.projectMetaSegments(project, selected)
	switch {
	case badge != "" && rest != "":
		return badge + " " + rest
	case badge != "":
		return badge
	default:
		return rest
	}
}

func (m Model) projectMetaSegments(project *projectState, selected bool) (badge string, rest string) {
	if project == nil {
		return "", ""
	}

	if project.hasDraft {
		label := "draft"
		if project.staleDraft {
			label = "staled draft"
			if project.draftVersion != "" {
				label += " v" + project.draftVersion
			}
		}
		badge = styles.RenderDraftBadge(label, selected)
	}

	parts := make([]string, 0, 3)
	version := project.displayVersion()
	if version != "" {
		parts = append(parts, "v"+version)
	}
	if project.loading || project.verifying {
		parts = append(parts, m.spin.View())
	} else if project.err != nil && project.tree != nil {
		parts = append(parts, "error")
	} else if state := project.cacheStateLabel(); state != "" {
		parts = append(parts, state)
	}
	if project.tree != nil && !project.tree.CachedAt.IsZero() {
		parts = append(parts, rcdisplay.FormatLocalDateTime(project.tree.CachedAt))
	}
	return badge, strings.Join(parts, " ")
}
