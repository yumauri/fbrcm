package contract

import (
	"sort"
)

var knownProblemCodeValues = []string{
	"argument.invalid", "argument.unknown_command", "auth.configuration_invalid", "auth.credentials_invalid", "auth.id_invalid", "auth.not_found", "auth.setup_required",
	"batch.failed", "batch.partial_success", "command.canceled", "command.not_executable", "command.not_found", "command.timeout",
	"condition.ambiguous", "condition.invalid", "condition.not_found", "configuration.invalid", "configuration.local_disabled", "configuration.local_not_found", "configuration.project_aliases_invalid",
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
