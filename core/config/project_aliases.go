package config

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

var projectAliasNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,62}$`)

// ValidateProjectAliasName validates a repository project alias.
func ValidateProjectAliasName(alias string) error {
	if !projectAliasNamePattern.MatchString(alias) {
		return fmt.Errorf("invalid project alias %q: use 1-63 lowercase letters, digits, hyphens, or underscores, starting with a letter", alias)
	}
	return nil
}

// ValidateProjectAliasProjectID validates the literal physical project ID
// stored as an alias target. Accessibility is intentionally not checked.
func ValidateProjectAliasProjectID(projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project alias target must not be empty")
	}
	if projectID != strings.TrimSpace(projectID) {
		return fmt.Errorf("project alias target %q must not have surrounding whitespace", projectID)
	}
	if strings.ContainsAny(projectID, " \t\r\n") {
		return fmt.Errorf("project alias target %q must not contain whitespace", projectID)
	}
	lower := strings.ToLower(projectID)
	if strings.HasPrefix(lower, "client@") || strings.HasPrefix(lower, "server@") {
		return fmt.Errorf("project alias target %q must reference a physical project ID without a template prefix", projectID)
	}
	if strings.Contains(projectID, "@") {
		return fmt.Errorf("project alias target %q must reference a physical project ID", projectID)
	}
	if strings.ContainsAny(projectID[:1], "=^/~") {
		return fmt.Errorf("project alias target %q must not use a filter prefix", projectID)
	}
	return nil
}

// ValidateProjectAliases validates every repository alias mapping.
func ValidateProjectAliases(aliases map[string]string) error {
	names := make([]string, 0, len(aliases))
	for alias := range aliases {
		names = append(names, alias)
	}
	slices.Sort(names)
	for _, alias := range names {
		projectID := aliases[alias]
		if err := ValidateProjectAliasName(alias); err != nil {
			return err
		}
		if err := ValidateProjectAliasProjectID(projectID); err != nil {
			return fmt.Errorf("project alias %q: %w", alias, err)
		}
	}
	return nil
}

// RejectGlobalProjectAliases enforces repository-only ownership.
func RejectGlobalProjectAliases(cfg *AppConfig) error {
	if len(projectAliases(cfg)) > 0 {
		return fmt.Errorf("projects.aliases: project aliases are repository-scoped; move them to .fbrcm.toml")
	}
	return nil
}

func projectAliases(cfg *AppConfig) map[string]string {
	if cfg == nil || cfg.Projects == nil {
		return nil
	}
	return cfg.Projects.Aliases
}

// CloneProjectAliases returns a detached alias map.
func CloneProjectAliases(cfg *AppConfig) map[string]string {
	aliases := projectAliases(cfg)
	if len(aliases) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(aliases))
	maps.Copy(out, aliases)
	return out
}

// CloneAppConfig returns a detached application configuration.
func CloneAppConfig(cfg *AppConfig) *AppConfig {
	if cfg == nil {
		cfg = &AppConfig{}
	}
	out := &AppConfig{
		Profile: cfg.Profile,
		Keys:    make(map[string]map[string][]string, len(cfg.Keys)),
	}
	if cfg.PowerlineGlyphs != nil {
		value := *cfg.PowerlineGlyphs
		out.PowerlineGlyphs = &value
	}
	for block, actions := range cfg.Keys {
		out.Keys[block] = make(map[string][]string, len(actions))
		for action, keys := range actions {
			out.Keys[block][action] = append([]string(nil), keys...)
		}
	}
	if cfg.Hooks != nil {
		out.Hooks = &HooksConfig{
			Timeout:     cfg.Hooks.Timeout,
			PrePublish:  append([]string(nil), cfg.Hooks.PrePublish...),
			PostPublish: append([]string(nil), cfg.Hooks.PostPublish...),
		}
	}
	if cfg.Projects != nil {
		out.Projects = &ProjectsConfig{Aliases: CloneProjectAliases(cfg)}
	}
	return out
}

// SetProjectAlias creates or replaces one alias mapping.
func SetProjectAlias(cfg *AppConfig, alias, projectID string) (string, bool, error) {
	if cfg == nil {
		return "", false, fmt.Errorf("application config is nil")
	}
	if err := ValidateProjectAliasName(alias); err != nil {
		return "", false, err
	}
	if err := ValidateProjectAliasProjectID(projectID); err != nil {
		return "", false, fmt.Errorf("project alias %q: %w", alias, err)
	}
	if cfg.Projects == nil {
		cfg.Projects = &ProjectsConfig{}
	}
	if cfg.Projects.Aliases == nil {
		cfg.Projects.Aliases = map[string]string{}
	}
	previous, exists := cfg.Projects.Aliases[alias]
	if exists && previous == projectID {
		return previous, false, nil
	}
	cfg.Projects.Aliases[alias] = projectID
	return previous, true, nil
}

// RemoveProjectAlias removes one alias mapping.
func RemoveProjectAlias(cfg *AppConfig, alias string) (string, bool, error) {
	if err := ValidateProjectAliasName(alias); err != nil {
		return "", false, err
	}
	if cfg == nil || cfg.Projects == nil || cfg.Projects.Aliases == nil {
		return "", false, nil
	}
	previous, exists := cfg.Projects.Aliases[alias]
	if !exists {
		return "", false, nil
	}
	delete(cfg.Projects.Aliases, alias)
	if len(cfg.Projects.Aliases) == 0 {
		cfg.Projects = nil
	}
	return previous, true, nil
}

// LoadProjectAliases loads effective repository aliases. Missing configuration
// is equivalent to an empty mapping.
func LoadProjectAliases() (map[string]string, error) {
	registry, err := LoadProjectAliasRegistry()
	if err != nil {
		return nil, err
	}
	return maps.Clone(registry.Aliases), nil
}

// ResolveProjectAlias resolves an exact case-sensitive alias.
func ResolveProjectAlias(aliases map[string]string, query string) (string, string, bool) {
	projectID, ok := aliases[query]
	if ok {
		return query, projectID, true
	}
	return "", "", false
}

// ProjectAliasesByID reverses an alias mapping into sorted alias lists.
func ProjectAliasesByID(aliases map[string]string) map[string][]string {
	out := make(map[string][]string)
	for alias, projectID := range aliases {
		out[projectID] = append(out[projectID], alias)
	}
	for projectID := range out {
		slices.Sort(out[projectID])
	}
	return out
}
