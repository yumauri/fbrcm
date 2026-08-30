package contract

import (
	"slices"
	"sort"
	"strings"
)

var knownProblemCodeValues = []string{
	"argument.invalid", "argument.unknown_command", "auth.configuration_invalid", "auth.credentials_invalid", "auth.id_invalid", "auth.not_found", "auth.quota_project_required", "auth.setup_required",
	"batch.failed", "batch.partial_success", "command.canceled", "command.not_executable", "command.not_found", "command.timeout",
	"condition.invalid", "condition.not_found", "configuration.invalid", "configuration.local_disabled", "configuration.local_not_found", "configuration.project_aliases_invalid",
	"diagnostic.failed", "draft.ambiguous", "draft.exists", "draft.not_found", "expression.invalid", "file.io_failed", "firebase.permission_denied", "firebase.rate_limited", "firebase.request_failed", "firebase.service_unavailable", "firebase.timeout", "filesystem.permission_denied",
	"group.not_found", "hook.failed", "hooks.changed", "hooks.not_configured", "interaction.required", "internal.contract_violation", "internal.unclassified",
	"network.offline", "network.timeout", "network.unavailable", "parameter.ambiguous", "parameter.exists", "parameter.not_found", "parameters_cache.not_found",
	"personalization.not_found", "profile.conflict", "profile.invalid", "profile.not_found", "project.ambiguous", "project.not_found", "project_alias.conflict", "project_alias.read_only", "publication.cache_failed", "publication.hook_failed",
	"theme.conflict", "theme.invalid", "theme.not_found",
	"remote_config.conflict", "remote_config.invalid", "remote_config.validation_failed", "resource.conflict", "resource.not_found", "result.unsuccessful", "schema.not_found", "stdin.remote_config.invalid", "validation.failed", "version.not_found",
}

var knownWarningCodeValues = []string{
	"cache.stale",
	"publication.cache_stale",
	"publication.draft_cleanup_failed",
	"publication.non_atomic",
	"publication.post_publish_hook_failed",
	"theme.already_exists",
}

// KnownProblemCodes returns the advisory catalog of stable codes currently
// emitted anywhere by the machine contract.
func KnownProblemCodes() []string {
	result := append([]string(nil), knownProblemCodeValues...)
	sort.Strings(result)
	return result
}

// KnownWarningCodes returns the advisory catalog of stable non-fatal warning
// codes currently emitted by the machine contract.
func KnownWarningCodes() []string {
	result := append([]string(nil), knownWarningCodeValues...)
	sort.Strings(result)
	return result
}

// CommandProblemCodes returns the closed set of top-level problem codes that
// current runtime control flow can emit for one executable command. Nested
// target failures in batch details remain an explicitly open extension point
// because their command is the target operation, not the aggregating command.
func CommandProblemCodes(capability Capability) []string {
	set := make(map[string]bool)
	add := func(values ...string) {
		for _, value := range values {
			set[value] = true
		}
	}
	add(
		"argument.invalid",
		"command.canceled",
		"command.timeout",
		"internal.contract_violation",
	)
	if commandCanReturnUnclassified(capability.ID) {
		add("internal.unclassified")
	}
	if commandLoadsConfiguration(capability.ID) {
		add("configuration.invalid")
	}
	if commandUsesProfile(capability.ID) {
		add("profile.invalid")
	}
	if commandCanReturnPathError(capability.ID) {
		add("file.io_failed", "filesystem.permission_denied")
	}
	if capability.Interaction.Mode != "none" {
		add("interaction.required")
	}
	if capability.NetworkAccess != "none" {
		add("network.offline", "network.timeout", "network.unavailable")
	}
	if hasCapabilityEffect(capability, "authentication_remote_access") {
		add("auth.configuration_invalid", "auth.credentials_invalid", "auth.not_found", "auth.quota_project_required", "auth.setup_required")
	}
	if hasAnyCapabilityEffect(capability,
		"firebase_remote_read", "firebase_remote_validation", "firebase_remote_write", "firebase_managed_feature_delete",
	) {
		add("firebase.permission_denied", "firebase.rate_limited", "firebase.request_failed", "firebase.service_unavailable", "firebase.timeout", "resource.not_found")
	}
	if hasCapabilityEffect(capability, "firebase_remote_write") {
		add(
			"batch.failed", "batch.partial_success", "hook.failed",
			"publication.cache_failed", "publication.hook_failed", "remote_config.conflict",
			"remote_config.invalid", "remote_config.validation_failed",
		)
		if capability.ID != "draft.publish" {
			add("draft.exists")
		}
	}
	if capability.Supports.Stdin {
		add("stdin.remote_config.invalid")
	}
	if capabilityHasFlag(capability, "--expr") {
		add("expression.invalid")
	}
	if capabilityHasFlag(capability, "--project") || capabilityHasArgument(capability, "project") || capabilityHasArgument(capability, "source_project") || capabilityHasArgument(capability, "target_project") {
		add("project.ambiguous", "project.not_found")
	}

	id := capability.ID
	switch {
	case id == "root":
		add("argument.unknown_command", "command.not_executable", "command.not_found")
	case id == "capabilities":
		add("command.not_executable", "command.not_found")
	case id == "schema.show":
		add("schema.not_found")
	case strings.HasPrefix(id, "config."):
		add("configuration.local_disabled", "configuration.local_not_found")
		if id == "config.validate" {
			add("validation.failed")
		}
	case id == "doctor":
		add("diagnostic.failed")
	case strings.HasPrefix(id, "auth."):
		add("auth.configuration_invalid")
		switch {
		case id == "auth.add.google":
			// The built-in client is build-scoped, so this command can fail
			// configuration validation but does not parse caller credentials or
			// perform a conflict-producing create operation.
		case strings.HasPrefix(id, "auth.add."):
			add("auth.credentials_invalid", "auth.id_invalid", "resource.conflict")
		case id == "auth.bind":
			add("auth.credentials_invalid", "auth.not_found", "auth.quota_project_required", "auth.setup_required")
		case id == "auth.login":
			add("auth.credentials_invalid", "auth.not_found", "auth.setup_required")
		case id == "auth.delete" || id == "auth.path" || strings.HasPrefix(id, "auth.quota-project."):
			add("auth.not_found")
		}
	case strings.HasPrefix(id, "draft."):
		if slices.Contains([]string{"draft.change-note", "draft.diff", "draft.discard", "draft.publish", "draft.show"}, id) {
			add("draft.ambiguous", "draft.not_found")
		}
		if id == "draft.publish" {
			add("parameters_cache.not_found")
		}
	case strings.HasPrefix(id, "profile"):
		switch id {
		case "profile.switch":
			add("profile.invalid")
		case "profile.path":
			add("profile.invalid", "profile.not_found")
		case "profile.delete", "profile.rename":
			add("profile.conflict", "profile.invalid", "profile.not_found")
		}
	case strings.HasPrefix(id, "theme"):
		switch id {
		case "theme.delete", "theme.rename":
			add("theme.conflict", "theme.invalid", "theme.not_found")
		case "theme.import":
			add("theme.conflict", "theme.invalid", "validation.failed")
		case "theme.switch":
			add("theme.invalid", "theme.not_found")
		}
	case strings.HasPrefix(id, "hooks."):
		if id == "hooks.fingerprint" || id == "hooks.trust" {
			add("hooks.not_configured")
		}
		if id == "hooks.trust" {
			add("hooks.changed")
		}
	case strings.HasPrefix(id, "projects.aliases."):
		add("configuration.local_disabled", "configuration.project_aliases_invalid")
		if id != "projects.aliases.list" {
			add("project_alias.conflict", "project_alias.read_only")
		}
	case strings.HasPrefix(id, "conditions."):
		if id == "conditions.show" || slices.Contains([]string{"conditions.delete", "conditions.edit", "conditions.move", "conditions.rename"}, id) {
			add("condition.not_found")
		}
		if slices.Contains([]string{"conditions.add", "conditions.delete", "conditions.edit", "conditions.move", "conditions.rename", "conditions.validate"}, id) {
			add("condition.invalid")
		}
	case strings.HasPrefix(id, "groups."):
		if id != "groups.list" {
			add("group.not_found")
		}
	case strings.HasPrefix(id, "versions."):
		add("version.not_found")
	case strings.HasPrefix(id, "personalizations."):
		if id == "personalizations.show" {
			add("personalization.not_found")
		}
	}

	switch id {
	case "add":
		add("parameter.exists")
	case "duplicate":
		add("parameter.ambiguous", "parameter.exists", "parameter.not_found")
	case "get", "delete", "update":
		add("parameter.ambiguous", "parameter.not_found")
	case "project.import":
		add("group.not_found")
	case "projects.diff", "projects.promote":
		add("group.not_found")
	case "experiments.delete", "rollouts.delete":
		add("interaction.required")
	}

	result := make([]string, 0, len(set))
	for value := range set {
		if !slices.Contains(knownProblemCodeValues, value) {
			panic("unknown command problem code " + value)
		}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func commandIsContractMetadata(commandID string) bool {
	return commandID == "help" || commandID == "capabilities" || strings.HasPrefix(commandID, "completion.") || strings.HasPrefix(commandID, "schema.")
}

func commandCanReturnUnclassified(commandID string) bool {
	return !commandIsContractMetadata(commandID)
}

func commandLoadsConfiguration(commandID string) bool {
	if commandIsContractMetadata(commandID) || commandID == "config.edit" {
		return false
	}
	return true
}

func commandUsesProfile(commandID string) bool {
	if commandIsContractMetadata(commandID) || strings.HasPrefix(commandID, "config.") || strings.HasPrefix(commandID, "theme.") || strings.HasPrefix(commandID, "hooks.") || strings.HasPrefix(commandID, "projects.aliases.") {
		return false
	}
	return true
}

func commandCanReturnPathError(commandID string) bool {
	if commandIsContractMetadata(commandID) || commandID == "root" || commandID == "config.edit" {
		return false
	}
	return true
}

func capabilityHasFlag(capability Capability, name string) bool {
	return slices.ContainsFunc(capability.Flags, func(flag FlagCapability) bool { return flag.Name == name })
}

func capabilityHasArgument(capability Capability, name string) bool {
	return slices.ContainsFunc(capability.Arguments, func(argument ArgumentCapability) bool { return argument.Name == name })
}

func hasCapabilityEffect(capability Capability, effect string) bool {
	return slices.Contains(capability.SideEffects, effect)
}

func hasAnyCapabilityEffect(capability Capability, effects ...string) bool {
	return slices.ContainsFunc(effects, func(effect string) bool { return hasCapabilityEffect(capability, effect) })
}
