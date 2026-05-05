package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"fbrcm/core/config"
	"fbrcm/core/firebase"
	corelog "fbrcm/core/log"
)

type ParametersCache = config.ParametersCache

const defaultParametersGroupKey = "__default__"

type ParametersCacheState int

const (
	ParametersCacheMissing ParametersCacheState = iota
	ParametersCacheFresh
	ParametersCacheStale
)

type ParametersTree struct {
	Version  string
	CachedAt time.Time
	ETag     string
	Groups   []ParametersGroup
}

type ParametersGroup struct {
	Key        string
	Label      string
	Parameters []ParametersEntry
}

type ParametersEntry struct {
	Key         string
	Description string
	Summary     string
	Values      []ParametersValue
}

type ParametersValue struct {
	Label     string
	Value     string
	Color     string
	Empty     bool
	EmptyType string
}

func (s *Core) InspectParametersCache(projectID string) (*ParametersCache, ParametersCacheState, error) {
	logger := corelog.For("core")
	logger.Debug("inspect parameters cache", "project_id", projectID)

	cache, err := config.LoadParametersCache(projectID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("parameters cache miss", "project_id", projectID)
			return nil, ParametersCacheMissing, nil
		}
		logger.Error("parameters cache load failed", "project_id", projectID, "err", err)
		return nil, ParametersCacheMissing, err
	}

	if _, err := firebase.ParseRemoteConfig(cache.RemoteConfig); err != nil {
		logger.Error("cached remote config decode failed", "project_id", projectID, "err", err)
		return nil, ParametersCacheMissing, fmt.Errorf("decode cached remote config: %w", err)
	}

	if cache.IsFresh(time.Now()) {
		logger.Info("parameters cache fresh", "project_id", projectID, "etag", cache.ETag)
		return cache, ParametersCacheFresh, nil
	}

	logger.Info("parameters cache stale", "project_id", projectID, "etag", cache.ETag)
	return cache, ParametersCacheStale, nil
}

func (s *Core) GetParameters(ctx context.Context, projectID string, force bool) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	logger.Debug("get parameters requested", "project_id", projectID, "force", force)

	if force {
		return s.fetchParameters(ctx, projectID)
	}

	cache, state, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, "", fmt.Errorf("inspect parameters cache: %w", err)
	}

	switch state {
	case ParametersCacheFresh:
		logger.Info("serving parameters from fresh cache", "project_id", projectID)
		return cache, "cache", nil
	case ParametersCacheStale:
		logger.Info("verifying stale parameters cache", "project_id", projectID)
		return s.verifyParameters(ctx, projectID, cache)
	default:
		logger.Warn("fetching parameters from firebase due to cache miss", "project_id", projectID)
		return s.fetchParameters(ctx, projectID)
	}
}

func (s *Core) RevalidateParameters(ctx context.Context, projectID string) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	logger.Debug("revalidate parameters requested", "project_id", projectID)

	cache, state, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, "", fmt.Errorf("inspect parameters cache: %w", err)
	}

	if state == ParametersCacheMissing || cache == nil {
		logger.Warn("fetching parameters from firebase due to cache miss during revalidation", "project_id", projectID)
		return s.fetchParameters(ctx, projectID)
	}

	return s.verifyParameters(ctx, projectID, cache)
}

func (s *Core) BuildParametersTree(cache *ParametersCache) (*ParametersTree, error) {
	if cache == nil {
		return nil, fmt.Errorf("parameters cache is nil")
	}

	remoteConfig, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}

	tree := &ParametersTree{
		Version:  remoteConfig.Version.VersionNumber,
		CachedAt: cache.CachedAt,
		ETag:     cache.ETag,
		Groups:   buildParametersGroups(remoteConfig),
	}

	corelog.For("core").Debug("built parameters tree", "version", tree.Version, "group_count", len(tree.Groups))
	return tree, nil
}

func (s *Core) fetchParameters(ctx context.Context, projectID string) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	logger.Info("fetch parameters from firebase", "project_id", projectID)

	fb, err := s.firebaseService(ctx)
	if err != nil {
		return nil, "", err
	}

	raw, etag, err := fb.GetRemoteConfig(ctx, projectID)
	if err != nil {
		logger.Error("firebase remote config fetch failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		logger.Error("firebase remote config decode failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("decode firebase remote config: %w", err)
	}

	cache := &config.ParametersCache{
		ETag:         etag,
		CachedAt:     time.Now().UTC(),
		RemoteConfig: raw,
	}
	if firebase.IsDryRun(ctx) {
		logger.Warn("dry run, skip parameters cache save after fetch", "project_id", projectID, "etag", etag)
		return cache, "firebase", nil
	}
	if err := config.SaveParametersCache(projectID, cache); err != nil {
		logger.Error("save parameters cache failed", "project_id", projectID, "etag", etag, "err", err)
		return nil, "", fmt.Errorf("save parameters cache: %w", err)
	}

	logger.Info("parameters cached from firebase", "project_id", projectID, "etag", etag)
	return cache, "firebase", nil
}

func (s *Core) verifyParameters(ctx context.Context, projectID string, cache *ParametersCache) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	logger.Info("verify parameters cache against firebase", "project_id", projectID, "etag", cache.ETag)

	remoteConfig, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		logger.Error("decode cached remote config failed during verification", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("decode cached remote config: %w", err)
	}

	fb, err := s.firebaseService(ctx)
	if err != nil {
		return nil, "", err
	}

	latestVersion, err := fb.GetLatestRemoteConfigVersion(ctx, projectID)
	if err != nil {
		logger.Error("remote config version check failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	if latestVersion.VersionNumber != "" && latestVersion.VersionNumber == remoteConfig.Version.VersionNumber {
		refreshed := *cache
		refreshed.CachedAt = time.Now().UTC()
		if firebase.IsDryRun(ctx) {
			logger.Warn("dry run, skip parameters cache timestamp refresh", "project_id", projectID, "version", latestVersion.VersionNumber)
			return &refreshed, "cache-verified", nil
		}
		if err := config.SaveParametersCache(projectID, &refreshed); err != nil {
			logger.Error("refresh parameters cache timestamp failed", "project_id", projectID, "err", err)
			return nil, "", fmt.Errorf("refresh parameters cache timestamp: %w", err)
		}
		logger.Info("parameters cache verified as current", "project_id", projectID, "version", latestVersion.VersionNumber)
		return &refreshed, "cache-verified", nil
	}

	logger.Info("parameters cache outdated; refetching", "project_id", projectID, "cached_version", remoteConfig.Version.VersionNumber, "latest_version", latestVersion.VersionNumber)
	return s.fetchParameters(ctx, projectID)
}

func buildParametersGroups(remoteConfig *firebase.RemoteConfig) []ParametersGroup {
	if remoteConfig == nil {
		return nil
	}

	conditionColors := make(map[string]string, len(remoteConfig.Conditions))
	conditionOrder := make(map[string]int, len(remoteConfig.Conditions))
	for i, condition := range remoteConfig.Conditions {
		conditionColors[condition.Name] = condition.TagColor
		conditionOrder[condition.Name] = i
	}

	groupKeys := sortedKeys(remoteConfig.ParameterGroups)
	seen := make(map[string]struct{})
	groups := make([]ParametersGroup, 0, len(groupKeys)+1)

	for _, groupKey := range groupKeys {
		group := remoteConfig.ParameterGroups[groupKey]
		params := buildParametersEntries(group.Parameters, conditionColors, conditionOrder)
		for key := range group.Parameters {
			seen[key] = struct{}{}
		}
		groups = append(groups, ParametersGroup{
			Key:        groupKey,
			Label:      groupKey,
			Parameters: params,
		})
	}

	rootParams := make(map[string]firebase.RemoteConfigParam)
	for key, param := range remoteConfig.Parameters {
		if _, ok := seen[key]; ok {
			continue
		}
		rootParams[key] = param
	}
	if len(rootParams) > 0 {
		groups = append([]ParametersGroup{{
			Key:        defaultParametersGroupKey,
			Label:      "(default)",
			Parameters: buildParametersEntries(rootParams, conditionColors, conditionOrder),
		}}, groups...)
	}

	return groups
}

func buildParametersEntries(params map[string]firebase.RemoteConfigParam, conditionColors map[string]string, conditionOrder map[string]int) []ParametersEntry {
	keys := sortedKeys(params)
	out := make([]ParametersEntry, 0, len(keys))
	for _, key := range keys {
		param := params[key]
		values := make([]ParametersValue, 0, len(param.ConditionalValues)+1)
		conditionKeys := sortedConditionalKeys(param.ConditionalValues, conditionOrder)
		for _, condition := range conditionKeys {
			values = append(values, ParametersValue{
				Label:     condition,
				Value:     formatRemoteConfigValue(param.ConditionalValues[condition], param.ValueType),
				Empty:     isEmptyRemoteConfigValue(param.ConditionalValues[condition]),
				EmptyType: emptyValueType(param.ValueType),
				Color:     conditionColors[condition],
			})
		}
		if param.DefaultValue != nil {
			values = append(values, ParametersValue{
				Label:     "default",
				Value:     formatRemoteConfigValue(*param.DefaultValue, param.ValueType),
				Empty:     isEmptyRemoteConfigValue(*param.DefaultValue),
				EmptyType: emptyValueType(param.ValueType),
			})
		}

		out = append(out, ParametersEntry{
			Key:         key,
			Description: strings.TrimSpace(param.Description),
			Summary:     summarizeParameterValues(values),
			Values:      values,
		})
	}
	return out
}

func summarizeParameterValues(values []ParametersValue) string {
	if len(values) == 0 {
		return "no values"
	}
	if len(values) == 1 {
		return values[0].Value
	}
	return fmt.Sprintf("%d values", len(values))
}

func formatRemoteConfigValue(value firebase.RemoteConfigValue, valueType string) string {
	switch {
	case value.UseInAppDefault:
		return "<in-app default>"
	case len(value.PersonalizationValue) > 0:
		return "<personalization>"
	case len(value.RolloutValue) > 0:
		return "<rollout>"
	case value.Value == "":
		return "(empty " + emptyValueType(valueType) + ")"
	default:
		return strings.ReplaceAll(value.Value, "\n", "\\n")
	}
}

func isEmptyRemoteConfigValue(value firebase.RemoteConfigValue) bool {
	return !value.UseInAppDefault && len(value.PersonalizationValue) == 0 && len(value.RolloutValue) == 0 && value.Value == ""
}

func emptyValueType(valueType string) string {
	valueType = strings.TrimSpace(strings.ToLower(valueType))
	if valueType == "" {
		return "string"
	}
	return valueType
}

func sortedKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := strings.ToLower(keys[i])
		right := strings.ToLower(keys[j])
		if left == right {
			return keys[i] < keys[j]
		}
		return left < right
	})
	return keys
}

func sortedConditionalKeys(items map[string]firebase.RemoteConfigValue, order map[string]int) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		left, leftOK := order[keys[i]]
		right, rightOK := order[keys[j]]
		switch {
		case leftOK && rightOK && left != right:
			return left < right
		case leftOK != rightOK:
			return leftOK
		default:
			leftKey := strings.ToLower(keys[i])
			rightKey := strings.ToLower(keys[j])
			if leftKey == rightKey {
				return keys[i] < keys[j]
			}
			return leftKey < rightKey
		}
	})

	return keys
}

func (s *Core) MarshalParametersTree(tree *ParametersTree) ([]byte, error) {
	return json.Marshal(tree)
}

func ParametersStatusLabel(source string, cachedAt time.Time, hasTree bool, err error) string {
	if err != nil && hasTree {
		return "error"
	}
	switch source {
	case "firebase":
		return "fetch"
	case "cache", "cache-verified", "cache-stale":
		if time.Since(cachedAt) > 10*time.Minute {
			return "staled"
		}
		if time.Since(cachedAt) < time.Minute {
			return "fetch"
		}
		return "cached"
	default:
		return ""
	}
}
