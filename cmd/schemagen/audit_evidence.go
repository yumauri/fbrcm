package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
)

var generatedAuditClasses = []string{
	"artifact", "batch", "boundary", "determinism", "discovery", "effectiveness", "effects", "failure",
	"interaction", "invocation", "no_op", "selection", "stdin", "success", "warning",
}

type generatedAuditMatrix struct {
	AuditStandardVersion string                          `json:"audit_standard_version"`
	ContractVersion      string                          `json:"contract_version"`
	Catalog              map[string]string               `json:"catalog"`
	Commands             []generatedAuditCommandEvidence `json:"commands"`
}

type generatedAuditCommandEvidence struct {
	ID      string                                `json:"id"`
	Classes map[string]generatedAuditEvidenceCell `json:"classes"`
}

type generatedAuditEvidenceCell struct {
	Evidence      []string `json:"evidence,omitempty"`
	NotApplicable string   `json:"not_applicable,omitempty"`
}

func buildAuditEvidenceMatrix(root *cobra.Command, capabilities []contract.Capability) generatedAuditMatrix {
	scenarios := loadAuditScenarios()
	matrix := generatedAuditMatrix{
		AuditStandardVersion: "1.0.0",
		ContractVersion:      contract.Version,
		Catalog:              generatedAuditEvidenceCatalog(),
		Commands:             make([]generatedAuditCommandEvidence, 0, len(capabilities)),
	}
	for _, capability := range capabilities {
		command := root
		if capability.ID != "root" {
			var err error
			var remaining []string
			command, remaining, err = root.Find(capability.Path)
			if err != nil || len(remaining) != 0 {
				panic(fmt.Sprintf("find audit command %s: error=%v remaining=%v", capability.ID, err, remaining))
			}
		}
		dataSchema, err := contract.ResponseDataSchema(command)
		if err != nil {
			panic(err)
		}
		successSchema, err := contract.ResponseSuccessDataSchema(command)
		if err != nil {
			panic(err)
		}
		input := inputSchema(capability, command)
		response := responseSchema(capability, dataSchema, successSchema)
		entry := generatedAuditCommandEvidence{ID: capability.ID, Classes: make(map[string]generatedAuditEvidenceCell, len(generatedAuditClasses))}
		for _, class := range generatedAuditClasses {
			if !generatedAuditClassApplies(class, capability, input, response, successSchema) {
				entry.Classes[class] = generatedAuditEvidenceCell{NotApplicable: generatedAuditNAReason(class)}
				continue
			}
			evidence := append([]string(nil), generatedAuditBaseEvidence(class)...)
			if capability.ID == "auth.add.google" {
				switch class {
				case "boundary", "invocation":
					evidence = append(evidence, "app.auth_add_quota_schema", "auth.google_quota_failure")
				case "failure":
					evidence = append(evidence, "app.auth_google_configuration_failure", "auth.google_quota_failure")
				}
			}
			if capability.Supports.Plan {
				switch class {
				case "artifact", "success":
					evidence = append(evidence, "plan.artifact_runtime")
				case "boundary", "invocation", "stdin":
					evidence = append(evidence, "app.plan_invocation_schema")
				case "effectiveness", "effects", "interaction":
					evidence = append(evidence, "app.plan_capability_boundaries")
				}
			}
			if capability.ID == "apply" {
				switch class {
				case "boundary", "effects", "interaction":
					evidence = append(evidence, "app.plan_capability_boundaries")
				case "invocation", "stdin":
					evidence = append(evidence, "app.plan_invocation_schema")
				case "success", "no_op":
					evidence = append(evidence, "apply.status_runtime")
				case "warning":
					evidence = append(evidence, "apply.draft_cleanup_runtime", "apply.warning_runtime")
				}
			}
			for _, scenario := range scenarios[capability.ID] {
				evidence = append(evidence, "e2e:"+scenario)
			}
			if class == "success" && len(scenarios[capability.ID]) == 0 {
				evidence = append(evidence, generatedAuditSuccessFallback(capability.ID))
			}
			sort.Strings(evidence)
			entry.Classes[class] = generatedAuditEvidenceCell{Evidence: slices.Compact(evidence)}
		}
		matrix.Commands = append(matrix.Commands, entry)
	}
	return matrix
}

func generatedAuditEvidenceCatalog() map[string]string {
	return map[string]string{
		"app.arity":                             "cli/app/contract_test.go#TestEveryExecutableCommandCobraArityMatchesCapability",
		"app.auth_add_quota_schema":             "cli/app/contract_test.go#TestAuthAddInvocationSchemasPublishQuotaProjectNormalizationAndGrammar",
		"app.auth_google_configuration_failure": "cli/app/contract_test.go#TestAuthAddGoogleMissingBuiltInClientReturnsConformingConfigurationFailure",
		"app.artifact_schema":                   "cli/app/contract_test.go#TestArtifactSchemasConstrainCommandReachableEncodings",
		"app.boundaries":                        "cli/app/contract_test.go#TestInvocationSchemasRejectMachineOnlyAndTypedValueContradictions",
		"app.capability_behavior":               "cli/app/contract_test.go#TestDetailedCapabilityGoldenCoversAuthoritativeBehavior",
		"app.completion_success":                "cli/app/contract_test.go#TestJSONCompletionCommandsUseConformingResponseSchemas",
		"app.discovery":                         "cli/app/contract_test.go#TestEveryExecutableCommandHasCapabilityAndPublishedSchemas",
		"app.effectiveness":                     "cli/app/contract_test.go#TestMachineIgnoredCommandOptionsAreExplicitInInvocationSchemas",
		"app.failure_codes":                     "cli/app/contract_test.go#TestCommandResponseSchemasConstrainReachableProblemCodes",
		"app.failure_runtime":                   "cli/app/contract_test.go#TestTypedFailureScenariosConformToRuntimeSchema",
		"app.interaction":                       "cli/app/contract_test.go#TestCapabilitiesDescribeMachineModeSafetyAndInteraction",
		"app.invocation_semantics":              "cli/app/contract_test.go#TestSemanticInvocationSchemasRejectInvalidCombinations",
		"app.no_op":                             "cli/app/contract_test.go#TestEmptyCollectionAndNoOpRuntimeEnvelopesConform",
		"app.outcomes":                          "cli/app/contract_test.go#TestCommandResponseSchemasConstrainReachableOutcomesAndWarnings",
		"app.plan_capability_boundaries":        "cli/app/contract_test.go#TestPublicationPlanCapabilitiesDescribeRuntimeBoundaries",
		"app.plan_invocation_schema":            "cli/app/contract_test.go#TestPublicationPlanInvocationSchemasPublishIntegrityAndInputSelection",
		"app.response_invariants":               "cli/app/contract_test.go#TestResponseSchemasRejectImpossibleDTOStates",
		"app.selection":                         "cli/app/contract_test.go#TestInvocationSchemasPublishCommandLocalSelectionSemantics",
		"app.stdin":                             "cli/app/contract_test.go#TestPublishedStdinSchemaDescribesRemoteConfig",
		"app.stdin_restrictions":                "cli/app/contract_test.go#TestStdinMutationSchemasRejectIgnoredRemoteOptions",
		"app.unknown_option":                    "cli/app/contract_test.go#TestEveryExecutableCommandFailureEnvelopeConformsToItsSchema",
		"app.warning_runtime":                   "cli/app/contract_test.go#TestPostPublicationFailureEnvelopesAndWarningsConform",
		"apply.no_change_success":               "cli/commands/apply/commands_test.go#TestApplyNoChangePlanSucceedsWithoutFirebase",
		"apply.status_runtime":                  "cli/commands/apply/commands_test.go#TestClassifyPublishResultCoversEveryStatusAndWarning",
		"apply.draft_cleanup_runtime":           "cli/commands/apply/commands_test.go#TestCleanupMatchingDraftDeletesOnlyExactSourceAndWarnsOnDriftOrFailure",
		"apply.warning_runtime":                 "cli/commands/apply/commands_test.go#TestNonAtomicWarningHasTypedDetailsAndSkipsDryRun",
		"auth.oauth_success":                    "core/firebase/auth_oauth_reauthorize_test.go#TestRecoverRejectedOAuthTokenReauthorizesWhenRefreshTokenIsInvalid",
		"auth.google_quota_failure":             "cli/commands/auth/commands_test.go#TestAuthAddGoogleRejectsInvalidQuotaProjectAsArgumentFailure",
		"contract.artifact_runtime":             "cli/contract/contract_test.go#TestArtifactEncodesBinaryContent",
		"contract.batch_runtime":                "cli/contract/contract_test.go#TestAllFailedBatchPreservesTypedTargetProblems",
		"draft.publish_success":                 "core/draft/pipeline_test.go#TestPublishExistingDraftSuccessRemovesDraft",
		"profile.root_success":                  "cli/commands/profile/commands_test.go#TestProfileRootJSON",
		"plan.metadata_success":                 "cli/commands/plan/commands_test.go#TestPlanShowAndValidateJSON",
		"plan.artifact_runtime":                 "cli/shared/rc/plan_test.go#TestWritePublicationPlanReportsExactPrivateArtifact",
		"schemagen.determinism":                 "cmd/schemagen/determinism_test.go#TestStageGeneratedContractIsByteDeterministic",
		"theme.mutation_success":                "cli/commands/theme/reset_test.go#TestSwitchBuiltInAndResetClearSelections",
		"versions.restore_success":              "cli/commands/versions/contracts_test.go#TestVersionPublishJSONRepresentsNoOp",
	}
}

func generatedAuditBaseEvidence(class string) []string {
	switch class {
	case "artifact":
		return []string{"app.artifact_schema", "contract.artifact_runtime"}
	case "batch":
		return []string{"app.outcomes", "contract.batch_runtime"}
	case "boundary":
		return []string{"app.boundaries", "app.invocation_semantics"}
	case "determinism":
		return []string{"schemagen.determinism"}
	case "discovery":
		return []string{"app.discovery"}
	case "effectiveness":
		return []string{"app.capability_behavior", "app.effectiveness"}
	case "effects":
		return []string{"app.capability_behavior"}
	case "failure":
		return []string{"app.failure_codes", "app.failure_runtime", "app.unknown_option"}
	case "interaction":
		return []string{"app.interaction"}
	case "invocation":
		return []string{"app.arity", "app.invocation_semantics", "app.unknown_option"}
	case "no_op":
		return []string{"app.no_op", "app.response_invariants"}
	case "selection":
		return []string{"app.selection"}
	case "stdin":
		return []string{"app.stdin", "app.stdin_restrictions"}
	case "success":
		return []string{"app.outcomes", "app.response_invariants"}
	case "warning":
		return []string{"app.outcomes", "app.warning_runtime"}
	default:
		panic("unknown generated audit class " + class)
	}
}

func generatedAuditSuccessFallback(commandID string) string {
	switch commandID {
	case "apply":
		return "apply.no_change_success"
	case "auth.login":
		return "auth.oauth_success"
	case "completion.bash", "completion.fish", "completion.powershell", "completion.zsh":
		return "app.completion_success"
	case "draft.publish":
		return "draft.publish_success"
	case "profile":
		return "profile.root_success"
	case "plan.show", "plan.validate":
		return "plan.metadata_success"
	case "theme.delete", "theme.reset", "theme.switch":
		return "theme.mutation_success"
	case "versions.restore":
		return "versions.restore_success"
	default:
		panic("successful command has neither an E2E scenario nor an audited unit fallback: " + commandID)
	}
}

func generatedAuditClassApplies(class string, capability contract.Capability, input, response, successSchema map[string]any) bool {
	switch class {
	case "discovery", "invocation", "boundary", "effectiveness", "failure", "determinism":
		return true
	case "selection":
		return generatedSchemaContainsKey(input, "x-fbrcm-matching")
	case "stdin":
		return capability.Supports.Stdin
	case "success":
		return slices.Contains(commandReachableOutcomes(capability, successSchema), "success")
	case "no_op":
		return generatedAuditMutation(capability.ID)
	case "interaction":
		return capability.Interaction.Mode != "none"
	case "warning":
		return len(commandWarningCodes(capability.ID)) > 0
	case "batch":
		return slices.Contains(commandReachableOutcomes(capability, successSchema), "partial_success")
	case "effects":
		return capability.SideEffectLevel > 0 || capability.NetworkAccess != "none"
	case "artifact":
		return generatedSchemaContainsKey(response, "sha256") && generatedSchemaContainsKey(response, "size_bytes")
	default:
		panic("unknown generated audit class " + class)
	}
}

func generatedSchemaContainsKey(value any, key string) bool {
	switch typed := value.(type) {
	case map[string]any:
		if _, ok := typed[key]; ok {
			return true
		}
		for _, child := range typed {
			if generatedSchemaContainsKey(child, key) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if generatedSchemaContainsKey(child, key) {
				return true
			}
		}
	}
	return false
}

func generatedAuditMutation(commandID string) bool {
	return slices.Contains([]string{
		"add", "apply", "auth.add.gcloud", "auth.add.google", "auth.add.oauth", "auth.add.service-account", "auth.bind", "auth.delete", "auth.login", "auth.quota-project.set", "auth.quota-project.unset",
		"cache.clear", "conditions.add", "conditions.delete", "conditions.edit", "conditions.move", "conditions.rename", "config.edit", "config.reset", "config.set", "delete", "draft.change-note", "draft.discard", "draft.publish", "duplicate",
		"experiments.delete", "groups.add", "groups.delete", "groups.edit", "groups.rename", "hooks.trust", "hooks.untrust", "profile", "profile.delete", "profile.rename", "profile.switch", "project.import", "project.quota-project.set", "project.quota-project.unset",
		"projects.aliases.import", "projects.aliases.remove", "projects.aliases.set", "projects.forget", "projects.promote", "projects.reset", "projects.update", "rollouts.delete", "theme.delete", "theme.import", "theme.rename", "theme.reset", "theme.switch", "update", "versions.restore", "versions.rollback",
	}, commandID)
}

func generatedAuditNAReason(class string) string {
	switch class {
	case "artifact":
		return "The command does not return a contract artifact DTO."
	case "batch":
		return "The command cannot produce a partial-success batch outcome."
	case "effects":
		return "The command declares no effects or network access."
	case "interaction":
		return "The command has no interactive branch in JSON mode."
	case "no_op":
		return "The command is not a mutation operation."
	case "selection":
		return "The command has no selector with matching or lookup semantics."
	case "stdin":
		return "The command does not accept a normalized stdin document."
	case "success":
		return "The JSON operation is interaction-only and has no success outcome."
	case "warning":
		return "The command has no reachable non-fatal warning branch."
	default:
		panic("audit class cannot be N/A: " + class)
	}
}

func loadAuditScenarios() map[string][]string {
	root := filepath.Join("e2e", "testdata", "scenarios")
	entries, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}
	result := make(map[string][]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, entry.Name(), "scenario.json"))
		if err != nil {
			panic(err)
		}
		var scenario struct {
			Name      string `json:"name"`
			CommandID string `json:"command_id"`
		}
		if err := json.Unmarshal(raw, &scenario); err != nil {
			panic(err)
		}
		if scenario.Name == "" {
			scenario.Name = entry.Name()
		}
		if scenario.Name != entry.Name() || strings.TrimSpace(scenario.CommandID) == "" {
			panic(fmt.Sprintf("invalid audit scenario %s", entry.Name()))
		}
		result[scenario.CommandID] = append(result[scenario.CommandID], scenario.Name)
	}
	for id := range result {
		sort.Strings(result[id])
	}
	return result
}
