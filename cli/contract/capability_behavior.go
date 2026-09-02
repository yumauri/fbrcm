package contract

import "strings"

type capabilityBehavior struct {
	level              int
	effects            []effectBehavior
	network            string
	networkWhen        []BehaviorConditionClause
	destructive        bool
	destructiveWhen    []BehaviorConditionClause
	destructiveReasons []string
	idempotency        string
	idempotencyWhen    []IdempotencyCondition
	stdin              bool
	interaction        *InteractionCapability
	interactionWhen    []BehaviorConditionClause
}

type effectBehavior struct {
	name string
	when []BehaviorConditionClause
}

func effect(name string, when ...BehaviorConditionClause) effectBehavior {
	return effectBehavior{name: name, when: cloneConditions(when)}
}

func behavior(level int, network string, effects ...effectBehavior) capabilityBehavior {
	return capabilityBehavior{level: level, effects: effects, network: network, idempotency: "yes"}
}

func localWrite() capabilityBehavior {
	return behavior(1, "none", effect("local_state_write",
		conditionClause(
			predicate("runtime_state", "mutation_plan", "has_changes", nil),
			predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
		)))
}

func localMutationEffects(names ...string) capabilityBehavior {
	when := conditionClause(
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
		predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
	)
	effects := make([]effectBehavior, 0, len(names))
	for _, name := range names {
		effects = append(effects, effect(name, when))
	}
	return behavior(1, "none", effects...)
}

func authReplacement(b capabilityBehavior) capabilityBehavior {
	return b.withEffect("local_file_delete",
		conditionClause(
			predicate("runtime_state", "mutation_plan", "is_destructive", nil),
			predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
		))
}

func credentialFileWrite(b capabilityBehavior) capabilityBehavior {
	return b.withEffect("local_file_write",
		conditionClause(predicate("runtime_state", "credential_file", "write_succeeded", nil)))
}

func previewableLocalWrite() capabilityBehavior {
	b := behavior(1, "none", effect("local_state_write",
		conditionClause(
			predicate("option", "dry-run", "equals", false),
			predicate("runtime_state", "mutation_plan", "has_changes", nil),
			predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
		)))
	return b
}

func nonIdempotent(b capabilityBehavior) capabilityBehavior {
	b.idempotency = "no"
	b.idempotencyWhen = nil
	return b
}

func conditionalRemoteRead(when []BehaviorConditionClause, extra ...effectBehavior) capabilityBehavior {
	effects := []effectBehavior{
		effect("firebase_remote_read", when...),
		effect("local_cache_write", appendPredicates(when, predicate("runtime_state", "remote_read", "cache_write_succeeded", nil))...),
	}
	effects = append(effects, extra...)
	return capabilityBehavior{level: 2, effects: effects, network: "conditional", networkWhen: cloneConditions(when), idempotency: "yes"}
}

func statelessUpdatingRemoteRead(extra ...effectBehavior) capabilityBehavior {
	return conditionalRemoteRead([]BehaviorConditionClause{
		conditionClause(predicate("option", "stateless", "equals", true)),
		conditionClause(predicate("option", "stateless", "equals", false), predicate("option", "update", "equals", true)),
		conditionClause(predicate("option", "stateless", "equals", false), predicate("runtime_state", "required_cache", "not_usable", nil)),
	}, extra...)
}

func managedFeatureRead() capabilityBehavior {
	refreshWhen := []BehaviorConditionClause{
		conditionClause(predicate("option", "update", "equals", true)),
		conditionClause(predicate("runtime_state", "required_cache", "not_usable", nil)),
	}
	return behavior(2, "required",
		effect("firebase_remote_read"),
		effect("local_cache_write", appendPredicates(refreshWhen,
			predicate("runtime_state", "remote_read", "cache_write_succeeded", nil))...),
	)
}

func historicalVersionRead(extra ...effectBehavior) capabilityBehavior {
	versionWhen := []BehaviorConditionClause{conditionClause(
		predicate("option", "cached", "equals", false),
		predicate("runtime_state", "version_request", "requires_network", nil),
	)}
	networkWhen := append(cloneConditions(versionWhen), conditionClause(
		predicate("option", "cached", "equals", false),
		predicate("runtime_state", "project_registry", "sync_required", nil),
	))
	effects := []effectBehavior{
		effect("firebase_remote_read", networkWhen...),
		effect("local_cache_write", appendPredicates(versionWhen,
			predicate("runtime_state", "remote_read", "cache_write_succeeded", nil))...),
	}
	effects = append(effects, extra...)
	return capabilityBehavior{level: 2, effects: effects, network: "conditional", networkWhen: networkWhen, idempotency: "yes"}
}

func requiredRemoteRead(extra ...effectBehavior) capabilityBehavior {
	effects := []effectBehavior{effect("firebase_remote_read")}
	effects = append(effects, extra...)
	return behavior(2, "required", effects...)
}

func requiredCacheableRemoteRead(extra ...effectBehavior) capabilityBehavior {
	effects := []effectBehavior{
		effect("firebase_remote_read"),
		effect("local_cache_write", conditionClause(predicate("runtime_state", "remote_read", "cache_write_succeeded", nil))),
	}
	effects = append(effects, extra...)
	return behavior(2, "required", effects...)
}

func requiredRegistrySync() capabilityBehavior {
	b := behavior(2, "required", effect("firebase_remote_read"), effect("local_state_write",
		conditionClause(predicate("runtime_state", "remote_read", "succeeded", nil))))
	b.idempotency = "no"
	return b
}

func conditionalReadOnly(when ...BehaviorConditionClause) capabilityBehavior {
	return capabilityBehavior{
		level:       2,
		effects:     []effectBehavior{effect("firebase_remote_read", when...)},
		network:     "conditional",
		networkWhen: cloneConditions(when),
		idempotency: "yes",
	}
}

func cachedRegistryRead() capabilityBehavior {
	return conditionalReadOnly(conditionClause(predicate("runtime_state", "required_cache", "not_usable", nil)))
}

func projectsListRead() capabilityBehavior {
	return conditionalReadOnly(
		conditionClause(predicate("option", "stateless", "equals", true)),
		conditionClause(predicate("option", "stateless", "equals", false), predicate("option", "update", "equals", true)),
		conditionClause(predicate("option", "stateless", "equals", false), predicate("runtime_state", "required_cache", "not_usable", nil)),
	)
}

func quotaProjectResolutionBehavior(base capabilityBehavior) capabilityBehavior {
	when := []BehaviorConditionClause{conditionClause(
		predicate("runtime_state", "quota_project_credentials", "requires_network", nil),
	)}
	base.level = 2
	base.network = "conditional"
	base.networkWhen = when
	base = base.withEffect("authentication_remote_access", when...)
	return base
}

func quotaProjectShowBehavior() capabilityBehavior {
	return quotaProjectResolutionBehavior(behavior(0, "none"))
}

func quotaProjectCredentialResolutionCommand(id string) bool {
	switch id {
	case "auth.quota-project.show", "auth.quota-project.unset", "project.quota-project.show", "project.quota-project.unset":
		return true
	default:
		return false
	}
}

func stdinRemoteMutation() capabilityBehavior {
	return remoteMutation("conditional", []BehaviorConditionClause{conditionClause(predicate("stdin", "document", "absent", nil))}, true)
}

func requiredRemoteMutation() capabilityBehavior {
	return remoteMutation("required", nil, true)
}

func projectImportMutation() capabilityBehavior {
	return remoteMutation("required", nil, false)
}

func remoteMutation(network string, remoteWhen []BehaviorConditionClause, validationAfterConfirmation bool) capabilityBehavior {
	validationPredicates := []BehaviorPredicate{
		predicate("option", "draft", "equals", false),
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
	}
	if validationAfterConfirmation {
		validationPredicates = append(validationPredicates,
			predicate("runtime_state", "confirmation", "authorized_or_not_required", nil))
	}
	validationWhen := appendPredicates(remoteWhen, validationPredicates...)
	publishWhen := appendPredicates(remoteWhen,
		predicate("option", "draft", "equals", false),
		predicate("option", "dry-run", "equals", false),
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
		predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
	)
	draftWhen := appendPredicates(remoteWhen,
		predicate("option", "draft", "equals", true),
		predicate("option", "dry-run", "equals", false),
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
		predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
	)
	preHookWhen := appendPredicates(remoteWhen,
		predicate("option", "draft", "equals", false),
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
		predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
		predicate("runtime_state", "trusted_hook", "configured_for_event", nil),
	)
	postHookWhen := appendPredicates(publishWhen,
		predicate("runtime_state", "publication", "accepted", nil),
		predicate("runtime_state", "trusted_hook", "configured_for_event", nil),
	)
	b := behavior(3, network,
		effect("firebase_remote_read", remoteWhen...),
		effect("local_cache_write", appendPredicates(remoteWhen,
			predicate("option", "dry-run", "equals", false),
			predicate("runtime_state", "remote_read", "cache_write_succeeded", nil))...),
		effect("firebase_remote_write", publishWhen...),
		effect("firebase_remote_validation", validationWhen...),
		effect("local_draft_write", draftWhen...),
		effect("trusted_hook_execution", append(preHookWhen, postHookWhen...)...),
	)
	b.networkWhen = cloneConditions(remoteWhen)
	b.idempotency = "conditional"
	b.idempotencyWhen = []IdempotencyCondition{
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("runtime_state", "confirmation", "required", nil))}},
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("option", "draft", "equals", true), predicate("runtime_state", "confirmation", "authorized_or_not_required", nil))}},
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("option", "dry-run", "equals", true), predicate("runtime_state", "trusted_hook", "not_executed", nil))}},
		{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(predicate("option", "dry-run", "equals", true), predicate("runtime_state", "trusted_hook", "executed", nil))}},
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("runtime_state", "mutation_plan", "has_no_changes", nil))}},
		{Idempotency: "no", When: appendPredicates(remoteWhen,
			predicate("option", "draft", "equals", false),
			predicate("option", "dry-run", "equals", false),
			predicate("runtime_state", "mutation_plan", "has_changes", nil),
			predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
		)},
	}
	if len(remoteWhen) > 0 {
		b.idempotencyWhen = append([]IdempotencyCondition{{
			Idempotency: "yes",
			When:        []BehaviorConditionClause{conditionClause(predicate("stdin", "document", "present", nil))},
		}}, b.idempotencyWhen...)
	}
	return b
}

func requiredPublication(cleanupDraft bool) capabilityBehavior {
	validationWhen := []BehaviorConditionClause{conditionClause(
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
		predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
	)}
	publishWhen := []BehaviorConditionClause{conditionClause(
		predicate("option", "dry-run", "equals", false),
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
		predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
	)}
	preHookWhen := appendPredicates(validationWhen,
		predicate("runtime_state", "trusted_hook", "configured_for_event", nil),
	)
	postHookWhen := appendPredicates(publishWhen,
		predicate("runtime_state", "publication", "accepted", nil),
		predicate("runtime_state", "trusted_hook", "configured_for_event", nil),
	)
	effects := []effectBehavior{
		effect("firebase_remote_read"),
		effect("local_cache_write", conditionClause(
			predicate("option", "dry-run", "equals", false),
			predicate("runtime_state", "remote_read", "cache_write_succeeded", nil))),
		effect("firebase_remote_write", publishWhen...),
		effect("firebase_remote_validation", validationWhen...),
		effect("trusted_hook_execution", append(preHookWhen, postHookWhen...)...),
	}
	if cleanupDraft {
		cleanupWhen := appendPredicates(publishWhen, predicate("runtime_state", "publication", "accepted", nil))
		cleanupWhen = append(cleanupWhen, conditionClause(
			predicate("option", "dry-run", "equals", false),
			predicate("runtime_state", "mutation_plan", "has_no_changes", nil),
		))
		effects = append(effects, effect("local_draft_delete", cleanupWhen...))
		effects = append(effects, effect("local_draft_write", conditionClause(
			predicate("option", "dry-run", "equals", false),
			predicate("runtime_state", "draft_change_note", "persisted", nil),
		)))
	}
	b := behavior(3, "required", effects...)
	b.idempotency = "conditional"
	b.idempotencyWhen = []IdempotencyCondition{
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("runtime_state", "confirmation", "required", nil))}},
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("option", "dry-run", "equals", true), predicate("runtime_state", "trusted_hook", "not_executed", nil))}},
		{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(predicate("option", "dry-run", "equals", true), predicate("runtime_state", "trusted_hook", "executed", nil))}},
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("runtime_state", "mutation_plan", "has_no_changes", nil))}},
		{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(predicate("option", "dry-run", "equals", false), predicate("runtime_state", "mutation_plan", "has_changes", nil), predicate("runtime_state", "confirmation", "authorized_or_not_required", nil))}},
	}
	return b
}

func appendPredicates(conditions []BehaviorConditionClause, predicates ...BehaviorPredicate) []BehaviorConditionClause {
	if len(conditions) == 0 {
		return []BehaviorConditionClause{conditionClause(predicates...)}
	}
	result := make([]BehaviorConditionClause, 0, len(conditions))
	for _, clause := range conditions {
		all := append(append([]BehaviorPredicate{}, clause.AllOf...), predicates...)
		result = append(result, conditionClause(all...))
	}
	return result
}

func (b capabilityBehavior) effectNames() []string {
	result := make([]string, 0, len(b.effects))
	for _, item := range b.effects {
		result = append(result, item.name)
	}
	return result
}

func (b capabilityBehavior) sideEffectConditions() []SideEffectCondition {
	result := make([]SideEffectCondition, 0, len(b.effects))
	for _, item := range b.effects {
		result = append(result, SideEffectCondition{Effect: item.name, When: cloneConditions(item.when)})
	}
	return result
}

func (b capabilityBehavior) withEffect(name string, when ...BehaviorConditionClause) capabilityBehavior {
	effects := make([]effectBehavior, len(b.effects))
	for index := range b.effects {
		effects[index] = effectBehavior{name: b.effects[index].name, when: cloneConditions(b.effects[index].when)}
	}
	b.effects = effects
	for index := range b.effects {
		if b.effects[index].name == name {
			if len(b.effects[index].when) == 0 || len(when) == 0 {
				b.effects[index].when = nil
				return b
			}
			b.effects[index].when = append(b.effects[index].when, cloneConditions(when)...)
			return b
		}
	}
	b.effects = append(b.effects, effect(name, when...))
	return b
}

// Every JSON invocation builds an envelope whose profile context bootstraps
// the default profile when no effective profile state can be loaded.
func withJSONEnvelopeProfileBootstrap(b capabilityBehavior) capabilityBehavior {
	if b.level < 1 {
		b.level = 1
	}
	return b.withEffect("local_state_write", conditionClause(
		predicate("runtime_state", "profile_bootstrap", "required", nil),
	))
}

func withProjectRegistrySync(b capabilityBehavior) capabilityBehavior {
	return b.withEffect("local_state_write", conditionClause(
		predicate("runtime_state", "project_registry", "sync_write_succeeded", nil),
	))
}

// withMachineAuthenticationEffects records authentication work performed while
// constructing a Firebase client. Doctor uses a diagnostic client that never
// persists credentials; auth login declares its more specific behavior in the
// command manifest below.
func withMachineAuthenticationEffects(id string, b capabilityBehavior) capabilityBehavior {
	if b.network == "none" || id == "auth.login" || id == "theme.import" || quotaProjectCredentialResolutionCommand(id) {
		return b
	}
	b = b.withEffect("authentication_remote_access", conditionClause(
		predicate("runtime_state", "authentication", "requires_network", nil),
	))
	if id != "doctor" {
		b = b.withEffect("local_file_write", conditionClause(
			predicate("runtime_state", "authentication", "token_persisted", nil),
		))
	}
	return b
}

func withStatelessCommandEffects(b capabilityBehavior) capabilityBehavior {
	effects := make([]effectBehavior, len(b.effects))
	for index := range b.effects {
		effects[index] = effectBehavior{name: b.effects[index].name, when: cloneConditions(b.effects[index].when)}
	}
	b.effects = effects
	b.networkWhen = cloneConditions(b.networkWhen)
	for effectIndex := range b.effects {
		for clauseIndex := range b.effects[effectIndex].when {
			clause := &b.effects[effectIndex].when[clauseIndex]
			if statelessPolicyDisablesEffect(b.effects[effectIndex].name) || clauseHasStatefulCommandPredicate(*clause) {
				clause.AllOf = append(clause.AllOf, predicate("option", "stateless", "equals", false))
			}
		}
	}
	for clauseIndex := range b.networkWhen {
		clause := &b.networkWhen[clauseIndex]
		if clauseHasStatefulCommandPredicate(*clause) {
			clause.AllOf = append(clause.AllOf, predicate("option", "stateless", "equals", false))
		}
	}
	return b
}

func statelessPolicyDisablesEffect(name string) bool {
	return strings.HasPrefix(name, "local_cache_") || strings.HasPrefix(name, "local_draft_") ||
		name == "local_state_write" || name == "trusted_hook_execution"
}

func withStatelessCommandInteractions(conditions []BehaviorConditionClause) []BehaviorConditionClause {
	result := cloneConditions(conditions)
	for index := range result {
		if containsPredicate([]BehaviorConditionClause{result[index]}, "runtime_state", "authentication", "requires_human_authorization") {
			result[index].AllOf = append(result[index].AllOf, predicate("option", "stateless", "equals", false))
		}
	}
	return result
}

func clauseHasStatefulCommandPredicate(clause BehaviorConditionClause) bool {
	for _, item := range clause.AllOf {
		if item.Source != "runtime_state" {
			continue
		}
		if item.Name == "project_registry" || item.Name == "profile_bootstrap" {
			return true
		}
		if item.Name == "authentication" && (item.Operator == "requires_network" || item.Operator == "token_persisted") {
			return true
		}
	}
	return false
}

func managedFeatureDelete() capabilityBehavior {
	b := behavior(3, "required",
		effect("firebase_remote_read"),
		effect("firebase_managed_feature_delete", conditionClause(predicate("option", "yes", "equals", true))),
	)
	b.idempotency = "conditional"
	b.idempotencyWhen = []IdempotencyCondition{
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("option", "yes", "equals", false))}},
		{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(predicate("option", "yes", "equals", true))}},
	}
	return withInteraction(b, "optional", "deletion_preview_requires_confirmation",
		conditionClause(predicate("option", "yes", "equals", false)))
}

func destructive(b capabilityBehavior, reason string) capabilityBehavior {
	b.destructive = true
	b.destructiveWhen = []BehaviorConditionClause{conditionClause(predicate("runtime_state", "mutation_plan", "is_destructive", nil))}
	b.destructiveReasons = []string{reason}
	return b
}

func withStdin(b capabilityBehavior) capabilityBehavior {
	b.stdin = true
	return b
}

// withPlanOutputEffects specializes a mutation capability for --plan-out.
// Planning still reads and validates Remote Config and may execute a trusted
// pre-publish hook, but it writes only the exclusive local plan artifact.
func withPlanOutputEffects(b capabilityBehavior) capabilityBehavior {
	b.destructiveWhen = cloneConditions(b.destructiveWhen)
	b.idempotencyWhen = cloneIdempotencyConditions(b.idempotencyWhen)
	effects := make([]effectBehavior, len(b.effects))
	for index, item := range b.effects {
		effects[index] = effectBehavior{name: item.name, when: cloneConditions(item.when)}
		if item.name == "firebase_remote_write" || strings.HasPrefix(item.name, "local_draft_") {
			effects[index].when = appendPredicates(effects[index].when,
				predicate("option", "plan-out", "equals", ""))
		}
	}
	b.effects = effects
	b = b.withEffect("local_file_write", conditionClause(
		predicate("option", "plan-out", "present", nil),
		predicate("runtime_state", "output_destination", "write_authorized", nil),
	))
	for index := range b.destructiveWhen {
		b.destructiveWhen[index].AllOf = append(b.destructiveWhen[index].AllOf,
			predicate("option", "plan-out", "equals", ""))
	}
	for index := range b.idempotencyWhen {
		if b.idempotencyWhen[index].Idempotency != "no" {
			continue
		}
		for clauseIndex := range b.idempotencyWhen[index].When {
			b.idempotencyWhen[index].When[clauseIndex].AllOf = append(
				b.idempotencyWhen[index].When[clauseIndex].AllOf,
				predicate("option", "plan-out", "equals", ""),
			)
		}
	}
	b.idempotencyWhen = append([]IdempotencyCondition{
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(
			predicate("option", "plan-out", "present", nil),
			predicate("runtime_state", "trusted_hook", "not_executed", nil),
		)}},
		{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(
			predicate("option", "plan-out", "present", nil),
			predicate("runtime_state", "trusted_hook", "executed", nil),
		)}},
	}, b.idempotencyWhen...)
	return b
}

func publicationPlanApplyBehavior() capabilityBehavior {
	publishTargets := conditionClause(predicate("runtime_state", "publication_plan", "has_publish_targets", nil))
	noPublishTargets := conditionClause(predicate("runtime_state", "publication_plan", "has_no_publish_targets", nil))
	cleanup := conditionClause(
		predicate("option", "dry-run", "equals", false),
		predicate("runtime_state", "source_draft_cleanup", "required", nil),
	)
	b := destructive(requiredPublication(false), "publishes the exact Remote Config templates recorded in a plan")
	b.network = "conditional"
	b.networkWhen = []BehaviorConditionClause{publishTargets}
	for index := range b.effects {
		if b.effects[index].name == "firebase_remote_read" {
			b.effects[index].when = []BehaviorConditionClause{publishTargets}
		}
	}
	b = b.withEffect("local_draft_delete", cleanup)
	cleanupDestructive := conditionClause(append(append([]BehaviorPredicate{}, cleanup.AllOf...), predicate("option", "stateless", "equals", false))...)
	b.destructiveWhen = append(b.destructiveWhen, cleanupDestructive)
	b.idempotency = "conditional"
	b.idempotencyWhen = []IdempotencyCondition{
		{Idempotency: "yes", When: []BehaviorConditionClause{noPublishTargets}},
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("runtime_state", "confirmation", "required", nil))}},
		{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(
			predicate("runtime_state", "publication_plan", "has_publish_targets", nil),
			predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
			predicate("runtime_state", "trusted_hook", "not_executed", nil),
		)}},
		{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(
			predicate("runtime_state", "publication_plan", "has_publish_targets", nil),
			predicate("runtime_state", "confirmation", "authorized_or_not_required", nil),
			predicate("runtime_state", "trusted_hook", "executed", nil),
		)}},
	}
	return withStdin(b)
}

func withInteraction(b capabilityBehavior, mode, jsonBehavior string, when ...BehaviorConditionClause) capabilityBehavior {
	b.interaction = &InteractionCapability{Mode: mode, JSONBehavior: jsonBehavior}
	b.interactionWhen = cloneConditions(when)
	return b
}

func themeSwitchBehavior() capabilityBehavior {
	b := localWrite()
	b.destructive = true
	b.destructiveWhen = []BehaviorConditionClause{conditionClause(
		predicate("argument", "name", "equals", "built-in"),
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
	)}
	b.destructiveReasons = []string{"the built-in selector removes a persisted theme selection"}
	return b
}

var capabilityBehaviors = map[string]capabilityBehavior{
	"root":                  behavior(0, "none"),
	"help":                  behavior(0, "none"),
	"capabilities":          behavior(0, "none"),
	"schema.list":           behavior(0, "none"),
	"schema.show":           behavior(0, "none"),
	"plan.show":             withStdin(behavior(0, "none")),
	"plan.validate":         withStdin(behavior(0, "none")),
	"completion.bash":       behavior(0, "none"),
	"completion.fish":       behavior(0, "none"),
	"completion.powershell": behavior(0, "none"),
	"completion.zsh":        behavior(0, "none"),

	"auth.list":                  behavior(0, "none"),
	"auth.path":                  behavior(0, "none"),
	"auth.quota-project.show":    quotaProjectShowBehavior(),
	"cache.list":                 behavior(0, "none"),
	"cache.path":                 behavior(0, "none"),
	"config.path":                behavior(0, "none"),
	"config.show":                behavior(0, "none"),
	"config.validate":            behavior(0, "none"),
	"draft.list":                 behavior(0, "none"),
	"draft.path":                 behavior(0, "none"),
	"hooks.fingerprint":          behavior(0, "none"),
	"hooks.status":               behavior(0, "none"),
	"profile":                    behavior(0, "none"),
	"profile.list":               behavior(0, "none"),
	"profile.path":               behavior(0, "none"),
	"theme":                      behavior(0, "none"),
	"theme.list":                 behavior(0, "none"),
	"theme.path":                 behavior(0, "none"),
	"project.templates.show":     behavior(0, "none"),
	"project.quota-project.show": quotaProjectShowBehavior(),
	"projects.aliases.list":      behavior(0, "none"),
	"projects.path":              behavior(0, "none"),

	"auth.add.gcloud": destructive(authReplacement(nonIdempotent(localWrite())), "replaces an existing auth identity and may remove its credential files"),
	"auth.add.google": destructive(authReplacement(nonIdempotent(localWrite())), "replaces an existing auth identity and may remove its cached token or credential files"),
	"auth.add.oauth": withInteraction(withStdin(destructive(credentialFileWrite(authReplacement(nonIdempotent(localWrite()))), "replaces an existing auth identity and may remove its cached token or credential files")), "optional", "missing_input_returns_interaction",
		conditionClause(predicate("stdin", "document", "absent", nil), predicate("option", "from", "equals", ""))),
	"auth.add.service-account": withInteraction(withStdin(destructive(credentialFileWrite(authReplacement(nonIdempotent(localWrite()))), "replaces an existing auth identity and may remove its credential files")), "optional", "missing_input_returns_interaction",
		conditionClause(predicate("stdin", "document", "absent", nil), predicate("option", "from", "equals", ""))),
	"auth.bind":                localWrite(),
	"auth.quota-project.set":   localWrite(),
	"auth.quota-project.unset": quotaProjectResolutionBehavior(localWrite()),
	"auth.delete":              destructive(localMutationEffects("local_state_write", "local_file_delete"), "removes stored authentication material"),
	"cache.clear":              destructive(localMutationEffects("local_cache_delete"), "removes cached Remote Config data"),
	"config.reset":             destructive(localWrite(), "removes persisted configuration values"),
	"config.set":               localWrite(),
	"draft.change-note": {
		level: 1,
		effects: []effectBehavior{effect("local_draft_write",
			conditionClause(predicate("argument", "text", "present", nil)),
			conditionClause(predicate("option", "clear", "equals", true)),
		)},
		network:     "none",
		idempotency: "conditional",
		idempotencyWhen: []IdempotencyCondition{
			{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("argument", "text", "absent", nil), predicate("option", "clear", "equals", false))}},
			{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(predicate("argument", "text", "present", nil))}},
			{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(predicate("option", "clear", "equals", true))}},
		},
	},
	"draft.discard":  destructive(localMutationEffects("local_draft_delete"), "removes local Remote Config drafts"),
	"hooks.trust":    localWrite(),
	"hooks.untrust":  destructive(localWrite(), "removes persisted hook trust"),
	"profile.delete": destructive(localMutationEffects("local_file_delete", "local_cache_delete", "local_draft_delete"), "removes a persisted profile"),
	"profile.rename": localWrite().withEffect("local_cache_move", conditionClause(
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
		predicate("runtime_state", "profile_cache", "available", nil),
	)),
	"profile.switch": behavior(1, "none", effect("local_state_write")),
	"theme.delete":   destructive(localMutationEffects("local_file_delete"), "removes an installed theme file"),
	"theme.rename": localWrite().withEffect("local_file_move", conditionClause(
		predicate("runtime_state", "mutation_plan", "has_changes", nil),
	)),
	"theme.switch": themeSwitchBehavior(),
	"theme.reset":  destructive(localWrite(), "removes a persisted theme selection"),
	"theme.import": withInteraction(withStdin(capabilityBehavior{
		level: 2,
		effects: []effectBehavior{effect("local_file_write", conditionClause(
			predicate("runtime_state", "mutation_plan", "has_changes", nil),
		))},
		network: "conditional",
		networkWhen: []BehaviorConditionClause{conditionClause(
			predicate("runtime_state", "theme_source", "requires_network", nil),
		)},
		idempotency: "conditional",
		idempotencyWhen: []IdempotencyCondition{
			{Idempotency: "yes", When: []BehaviorConditionClause{conditionClause(predicate("runtime_state", "theme_source", "is_directory", nil))}},
			{Idempotency: "no", When: []BehaviorConditionClause{conditionClause(predicate("runtime_state", "theme_source", "is_single", nil))}},
		},
	}), "optional", "missing_input_returns_interaction",
		conditionClause(predicate("argument", "source", "absent", nil), predicate("stdin", "document", "absent", nil))),
	"project.templates.set":       localWrite(),
	"project.quota-project.set":   localWrite(),
	"project.quota-project.unset": quotaProjectResolutionBehavior(localWrite()),
	"projects.aliases.import":     destructive(previewableLocalWrite(), "--conflict overwrite may replace persisted project aliases"),
	"projects.aliases.remove":     destructive(localWrite(), "removes persisted project aliases"),
	"projects.aliases.set":        destructive(localWrite(), "replaces an existing project alias when the mapping changes"),
	"projects.forget":             destructive(localMutationEffects("local_state_write", "local_cache_delete", "local_draft_delete"), "removes projects from the local registry and deletes their cached templates, version snapshots, and drafts"),
	"projects.reset":              destructive(localMutationEffects("local_state_write", "local_file_delete"), "replaces the local project registry"),

	"draft.show": withInteraction(behavior(1, "none", effect("local_file_write",
		conditionClause(predicate("runtime_state", "output_destination", "write_authorized", nil)))), "optional", "destination_conflict_returns_interaction",
		conditionClause(predicate("runtime_state", "output_destination", "conflicts", nil))),
	"config.edit": withInteraction(behavior(0, "none"), "required", "external_input_returns_interaction", conditionClause(predicate("runtime_state", "external_editor", "required", nil))),
	// JSON invocation rejects the streaming mode before starting a server.
	"mcp":          behavior(0, "none"),
	"project.open": withInteraction(cachedRegistryRead(), "none", "browser_launch_suppressed"),
	"auth.login": withInteraction(capabilityBehavior{
		level: 2,
		effects: []effectBehavior{
			effect("authentication_remote_access", conditionClause(predicate("runtime_state", "authentication", "requires_network", nil))),
			effect("local_file_write", conditionClause(predicate("runtime_state", "authentication", "token_persisted", nil))),
		},
		network: "conditional", networkWhen: []BehaviorConditionClause{conditionClause(predicate("runtime_state", "authentication", "requires_network", nil))}, idempotency: "yes",
	},
		"optional", "oauth_authorization_returns_interaction",
		conditionClause(predicate("runtime_state", "authentication", "requires_human_authorization", nil))),

	"conditions.list":     statelessUpdatingRemoteRead(),
	"conditions.show":     statelessUpdatingRemoteRead(),
	"conditions.validate": requiredCacheableRemoteRead(effect("firebase_remote_validation")),
	"doctor": {
		level: 2,
		effects: []effectBehavior{
			effect("local_file_write", conditionClause(predicate("runtime_state", "diagnostic_cache_probe", "write_succeeded", nil))),
			effect("local_file_delete", conditionClause(predicate("runtime_state", "diagnostic_cache_probe", "delete_succeeded", nil))),
			effect("firebase_remote_read", conditionClause(
				predicate("context", "offline", "equals", false),
				predicate("runtime_state", "diagnostic_identity", "available", nil),
			)),
		},
		network: "conditional",
		networkWhen: []BehaviorConditionClause{conditionClause(
			predicate("context", "offline", "equals", false),
			predicate("runtime_state", "diagnostic_identity", "available", nil),
		)},
		idempotency: "yes",
	},
	"draft.diff": conditionalRemoteRead([]BehaviorConditionClause{conditionClause(
		predicate("option", "against", "equals", "current"),
		predicate("option", "cached", "equals", false),
	)}),
	"experiments.list": managedFeatureRead(),
	"experiments.show": managedFeatureRead(),
	"get": withStdin(conditionalRemoteRead([]BehaviorConditionClause{
		conditionClause(predicate("stdin", "document", "absent", nil), predicate("option", "stateless", "equals", true)),
		conditionClause(predicate("stdin", "document", "absent", nil), predicate("option", "stateless", "equals", false), predicate("option", "update", "equals", true)),
		conditionClause(predicate("stdin", "document", "absent", nil), predicate("option", "stateless", "equals", false), predicate("runtime_state", "required_cache", "not_usable", nil)),
	})),
	"groups.list":           statelessUpdatingRemoteRead(),
	"personalizations.list": statelessUpdatingRemoteRead(),
	"personalizations.show": statelessUpdatingRemoteRead(),
	"project.show":          projectsListRead(),
	"projects.diff":         conditionalReadOnly(conditionClause(predicate("option", "cached", "equals", false))),
	"projects.list":         projectsListRead(),
	"rollouts.list":         managedFeatureRead(),
	"rollouts.show":         managedFeatureRead(),
	"versions.diff":         historicalVersionRead(),
	"versions.list":         conditionalReadOnly(conditionClause(predicate("option", "cached", "equals", false))),
	"versions.show":         historicalVersionRead(),
	"projects.update":       requiredRegistrySync(),
	"project.defaults": destructive(requiredRemoteRead(effect("local_file_write",
		conditionClause(predicate("runtime_state", "output_destination", "write_authorized", nil)))), "an existing destination file may be overwritten"),
	"project.export": destructive(requiredRemoteRead(effect("local_file_write",
		conditionClause(predicate("runtime_state", "output_destination", "write_authorized", nil)))), "an existing destination file may be overwritten"),
	"versions.export": destructive(historicalVersionRead(
		effect("local_file_write", conditionClause(predicate("runtime_state", "output_destination", "write_authorized", nil)))), "an existing destination file may be overwritten"),

	"add":               withStdin(stdinRemoteMutation()),
	"conditions.add":    requiredRemoteMutation(),
	"conditions.delete": destructive(requiredRemoteMutation(), "removes Remote Config conditions"),
	"conditions.edit":   requiredRemoteMutation(),
	"conditions.move":   requiredRemoteMutation(),
	"conditions.rename": requiredRemoteMutation(),
	"delete":            withStdin(destructive(stdinRemoteMutation(), "removes Remote Config parameters")),
	"duplicate":         requiredRemoteMutation(),
	"groups.add":        requiredRemoteMutation(),
	"groups.delete":     destructive(requiredRemoteMutation(), "removes Remote Config parameter groups"),
	"groups.edit":       requiredRemoteMutation(),
	"groups.rename":     requiredRemoteMutation(),
	"update":            withStdin(destructive(stdinRemoteMutation(), "specific flags or selected changes may replace or remove existing values")),
	"project.import": withInteraction(withStdin(destructive(projectImportMutation(), "selected changes may replace or remove existing values")),
		"optional", "import_requires_explicit_strategy_and_confirmation",
		conditionClause(predicate("option", "yes", "equals", false), predicate("runtime_state", "confirmation", "required", nil)),
		conditionClause(predicate("option", "merge", "equals", false), predicate("option", "override", "equals", false), predicate("runtime_state", "import_strategy", "required", nil)),
		conditionClause(predicate("runtime_state", "import_merge_resolution", "required", nil)),
		conditionClause(predicate("stdin", "document", "absent", nil), predicate("option", "from", "equals", ""))),

	"draft.publish":      destructive(requiredPublication(true), "a draft may replace or remove current Remote Config content"),
	"experiments.delete": destructive(managedFeatureDelete(), "removes a Firebase Remote Config experiment"),
	"rollouts.delete":    destructive(managedFeatureDelete(), "removes a Firebase Remote Config rollout"),
	"projects.promote": withInteraction(destructive(requiredPublication(false), "selected changes may replace or remove existing values"),
		"optional", "promotion_requires_explicit_selection_and_confirmation",
		conditionClause(predicate("option", "yes", "equals", false), predicate("runtime_state", "confirmation", "required", nil)),
		conditionClause(predicate("runtime_state", "promotion_selection", "required", nil))),
	"versions.restore":  destructive(requiredPublication(false), "replaces the current Remote Config template"),
	"versions.rollback": destructive(requiredPublication(false), "replaces the current Remote Config template"),
	"apply":             publicationPlanApplyBehavior(),
}

func init() {
	for _, id := range []string{
		"add", "delete", "duplicate", "update",
		"conditions.add", "conditions.delete", "conditions.edit", "conditions.list", "conditions.move", "conditions.rename", "conditions.show", "conditions.validate",
		"experiments.delete", "experiments.list", "experiments.show",
		"get",
		"groups.add", "groups.delete", "groups.edit", "groups.list", "groups.rename",
		"personalizations.list", "personalizations.show",
		"project.defaults", "project.export", "project.import", "project.open", "project.show",
		"projects.diff", "projects.list", "projects.promote",
		"rollouts.delete", "rollouts.list", "rollouts.show",
		"versions.diff", "versions.export", "versions.list", "versions.restore", "versions.rollback", "versions.show",
	} {
		behavior, ok := capabilityBehaviors[id]
		if !ok {
			panic("project-registry capability references unknown command " + id)
		}
		capabilityBehaviors[id] = withProjectRegistrySync(behavior)
	}
}
