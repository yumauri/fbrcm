package parameters

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) loadParametersCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		cache, state, err := m.svc.InspectParametersCache(project.ProjectID)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}

		switch state {
		case core.ParametersCacheFresh:
			tree, hasDraft, err := m.svc.BuildDraftAwareParametersTree(project.ProjectID, cache)
			source := "cache"
			if hasDraft {
				source = "draft"
			}
			cacheVersion := remoteConfigVersion(cache.RemoteConfig)
			draftVersion := ""
			staleDraft := false
			if hasDraft && tree != nil {
				draftVersion = tree.Version
				staleDraft = versionLess(draftVersion, cacheVersion)
			}
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: "cache", CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, HasDraft: hasDraft, StaleDraft: staleDraft}
		case core.ParametersCacheStale:
			tree, hasDraft, err := m.svc.BuildDraftAwareParametersTree(project.ProjectID, cache)
			source := "cache-stale"
			if hasDraft {
				source = "draft"
			}
			cacheVersion := remoteConfigVersion(cache.RemoteConfig)
			draftVersion := ""
			staleDraft := false
			if hasDraft && tree != nil {
				draftVersion = tree.Version
				staleDraft = versionLess(draftVersion, cacheVersion)
			}
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: "cache-stale", CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, Revalidate: true, RevalidateCache: cache, HasDraft: hasDraft, StaleDraft: staleDraft}
		default:
			cache, source, err := m.svc.GetParameters(context.Background(), project.ProjectID, false)
			if err != nil {
				return messages.ParametersLoadedMsg{Project: project, Err: err}
			}
			tree, hasDraft, err := m.svc.BuildDraftAwareParametersTree(project.ProjectID, cache)
			cacheSource := source
			if hasDraft {
				source = "draft"
			}
			cacheVersion := remoteConfigVersion(cache.RemoteConfig)
			draftVersion := ""
			staleDraft := false
			if hasDraft && tree != nil {
				draftVersion = tree.Version
				staleDraft = versionLess(draftVersion, cacheVersion)
			}
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: cacheSource, CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, HasDraft: hasDraft, StaleDraft: staleDraft}
		}
	}
}

func (m Model) revalidateParametersCmd(project core.Project, previousCache *core.ParametersCache) tea.Cmd {
	return func() tea.Msg {
		cache, tree, source, hasDraft, staleDraft, err := m.svc.RefreshDraftAwareParameters(context.Background(), project.ProjectID, previousCache)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}
		_ = cache
		cacheSource := source
		if source == "draft" || source == "draft-stale" {
			cacheSource = "firebase"
		}
		cacheVersion := remoteConfigVersion(cache.RemoteConfig)
		draftVersion := ""
		if hasDraft && tree != nil {
			draftVersion = tree.Version
			staleDraft = staleDraft || versionLess(draftVersion, cacheVersion)
		}
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: cacheSource, CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, HasDraft: hasDraft, StaleDraft: staleDraft}
	}
}

func (m Model) forceParametersCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		cache, tree, source, hasDraft, staleDraft, err := m.svc.RefreshDraftAwareParameters(context.Background(), project.ProjectID, nil)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}
		_ = cache
		cacheSource := source
		if source == "draft" || source == "draft-stale" {
			cacheSource = "firebase"
		}
		cacheVersion := remoteConfigVersion(cache.RemoteConfig)
		draftVersion := ""
		if hasDraft && tree != nil {
			draftVersion = tree.Version
			staleDraft = staleDraft || versionLess(draftVersion, cacheVersion)
		}
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: cacheSource, CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, HasDraft: hasDraft, StaleDraft: staleDraft}
	}
}
