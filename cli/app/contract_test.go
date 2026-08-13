package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	sharedrc "github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	corehooks "github.com/yumauri/fbrcm/core/hooks"
	"github.com/yumauri/fbrcm/schemas"
)

func TestEveryExecutableCommandHasCapabilityAndPublishedSchemas(t *testing.T) {
	root := NewRootForContract("schema")
	index := contract.Capabilities(root)
	detailed := contract.DetailedCapabilities(root)
	if index.Count == 0 {
		t.Fatal("capability index is empty")
	}
	ids, err := schemas.List()
	if err != nil {
		t.Fatal(err)
	}
	currentPrefix := "urn:fbrcm:schema:cli:" + contract.Version + ":"
	currentCount := 0
	for _, id := range ids {
		if strings.HasPrefix(id, currentPrefix) {
			currentCount++
		}
	}
	if currentCount != index.Count*2+9 {
		t.Fatalf("published current-schema count = %d, want %d", currentCount, index.Count*2+9)
	}
	seen := make(map[string]bool, index.Count)
	for _, capability := range detailed {
		if capability.ID == "" || seen[capability.ID] {
			t.Fatalf("invalid or duplicate capability id %q", capability.ID)
		}
		seen[capability.ID] = true
		if capability.InvocationSchema == "" || capability.ResponseSchema == "" {
			t.Fatalf("%s has incomplete schema identifiers", capability.ID)
		}
		if _, err := schemas.ReadByID(capability.InvocationSchema); err != nil {
			t.Fatalf("%s input schema: %v", capability.ID, err)
		}
		responseRaw, err := schemas.ReadByID(capability.ResponseSchema)
		if err != nil {
			t.Fatalf("%s response schema: %v", capability.ID, err)
		}
		command := commandForCapability(t, root, capability)
		wantData, err := contract.ResponseDataSchema(command)
		if err != nil {
			t.Fatalf("%s response DTO: %v", capability.ID, err)
		}
		gotData := responseDataSchema(t, responseRaw)
		if capability.ID != "schema.show" && containsArbitraryJSONDocument(gotData) {
			t.Fatalf("%s uses an unconstrained top-level response document", capability.ID)
		}
		gotDataRaw, _ := json.Marshal(gotData)
		wantDataRaw, _ := json.Marshal(wantData)
		if !bytes.Equal(gotDataRaw, wantDataRaw) {
			t.Fatalf("%s generated data schema is stale\ngot:  %s\nwant: %s", capability.ID, gotDataRaw, wantDataRaw)
		}
		assertResolvablePublishedSchemaReferences(t, capability.ID, responseRaw)
		if capability.Supports.DryRun != hasContractFlag(root, capability.Path, "dry-run") {
			t.Fatalf("%s dry-run metadata does not match its flags", capability.ID)
		}
		if capability.Supports.Draft != hasContractFlag(root, capability.Path, "draft") {
			t.Fatalf("%s draft metadata does not match its flags", capability.ID)
		}
	}
}

func TestCollectionResponseSchemasDescribeItemDTOs(t *testing.T) {
	tests := []struct {
		id     string
		fields []string
	}{
		{"urn:fbrcm:schema:cli:1.0.0:command:get:response", []string{"project", "project_id", "group", "key", "description", "default_value", "conditional", "conditions", "type", "version", "cached_at", "status"}},
		{"urn:fbrcm:schema:cli:1.0.0:command:projects.list:response", []string{"project", "project_id", "aliases", "number", "auth_id", "updated_at", "synced_at", "url", "state", "disabled", "discovered_by", "etag", "templates", "primary_template"}},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			raw, err := schemas.ReadByID(test.id)
			if err != nil {
				t.Fatal(err)
			}
			data := responseDataSchema(t, raw)
			itemProperties := collectionItemProperties(t, data)
			for _, field := range test.fields {
				if _, ok := itemProperties[field]; !ok {
					t.Errorf("item schema is missing %q", field)
				}
			}
		})
	}
}

func TestResponseStatusAndValidationSourceFieldsAreSemantic(t *testing.T) {
	capabilities := contract.DetailedCapabilities(NewRootForContract("schema"))
	for _, capability := range capabilities {
		raw, err := schemas.ReadByID(capability.ResponseSchema)
		if err != nil {
			t.Fatal(err)
		}
		var document any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		walkSchemaProperties(document, func(name string, schema map[string]any) {
			switch name {
			case "status":
				if schema["enum"] == nil && schema["$ref"] == nil && schema["const"] == nil {
					t.Errorf("%s has unconstrained status schema %#v", capability.ID, schema)
				}
			case "validation_source":
				if schema["enum"] == nil && schema["const"] == nil && schema["$ref"] != contract.SemanticRef("validation_source") {
					t.Errorf("%s has unconstrained validation_source schema %#v", capability.ID, schema)
				}
			}
		})
	}
}

func TestResponseStringFormatsAreSemantic(t *testing.T) {
	tests := []struct {
		command string
		fields  map[string]string
	}{
		{"project.open", map[string]string{"url": "uri"}},
		{"projects.list", map[string]string{"url": "uri"}},
		{"versions.list", map[string]string{"updateTime": "date-time", "email": "email", "imageUrl": "uri"}},
		{"experiments.list", map[string]string{"startTime": "date-time", "endTime": "date-time", "lastUpdateTime": "date-time"}},
	}
	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			raw, err := schemas.ReadByID(contract.SchemaID(test.command))
			if err != nil {
				t.Fatal(err)
			}
			var document any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			found := make(map[string]bool, len(test.fields))
			walkSchemaProperties(document, func(name string, schema map[string]any) {
				format, expected := test.fields[name]
				if expected && schema["format"] == format {
					found[name] = true
				}
			})
			for field := range test.fields {
				if !found[field] {
					t.Errorf("%s does not publish the expected format for %s", test.command, field)
				}
			}
		})
	}
}

func TestCapabilityGoldenCoversEveryExecutableCommand(t *testing.T) {
	root := NewRootForContract("schema")
	got, err := json.MarshalIndent(contract.Capabilities(root), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	want, err := os.ReadFile("testdata/contract_v1_capabilities.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("capability contract changed; run go run ./cmd/schemagen and review the golden diff")
	}
}

func TestCapabilitiesDescribeMachineModeSafetyAndInteraction(t *testing.T) {
	root := NewRootForContract("test")
	draftPublish, err := contract.FindCapability(root, []string{"draft", "publish"})
	if err != nil {
		t.Fatal(err)
	}
	if !draftPublish.Destructive || len(draftPublish.DestructiveWhen) == 0 {
		t.Fatalf("draft.publish capability = %#v", draftPublish)
	}
	projectOpen, err := contract.FindCapability(root, []string{"project", "open"})
	if err != nil {
		t.Fatal(err)
	}
	if projectOpen.SideEffectLevel == 3 || projectOpen.Interaction.Mode != "optional" || projectOpen.Interaction.JSONBehavior != "browser_launch_suppressed_and_oauth_authorization_returns_interaction" || projectOpen.NetworkAccess != "conditional" || slices.Contains(projectOpen.SideEffects, "local_cache_write") || !capabilityHasPredicate(projectOpen, "local_state_write", "project_registry", "sync_write_succeeded") || !capabilityConditionsHavePredicate(projectOpen.InteractionWhen, "runtime_state", "authentication", "requires_human_authorization") {
		t.Fatalf("project.open capability = %#v", projectOpen)
	}
	authLogin, err := contract.FindCapability(root, []string{"auth", "login"})
	if err != nil {
		t.Fatal(err)
	}
	if authLogin.Interaction.Mode != "optional" || authLogin.NetworkAccess != "conditional" || len(authLogin.NetworkWhen) == 0 || !capabilityHasPredicate(authLogin, "authentication_remote_access", "authentication", "requires_network") || !capabilityHasPredicate(authLogin, "local_file_write", "authentication", "token_persisted") || capabilityHasPredicate(authLogin, "local_state_write", "authentication", "token_persisted") || authLogin.Idempotency != "yes" || len(authLogin.IdempotencyWhen) != 0 {
		t.Fatalf("auth.login capability = %#v", authLogin)
	}
	if flag := slices.IndexFunc(authLogin.Flags, func(flag contract.FlagCapability) bool { return flag.Name == "--noopen" }); flag < 0 || authLogin.Flags[flag].Effective {
		t.Fatalf("auth.login noopen flag should be accepted but ineffective in JSON mode: %#v", authLogin.Flags)
	}
	configEdit, err := contract.FindCapability(root, []string{"config", "edit"})
	if err != nil {
		t.Fatal(err)
	}
	if configEdit.SideEffectLevel != 1 || configEdit.NetworkAccess != "none" || !slices.Contains(configEdit.SideEffects, "local_state_write") || configEdit.Interaction.Mode != "required" {
		t.Fatalf("config.edit capability = %#v", configEdit)
	}
	if !capabilityHasPredicate(configEdit, "local_state_write", "profile_bootstrap", "required") {
		t.Fatalf("config.edit profile bootstrap side effect = %#v", configEdit.SideEffectWhen)
	}
	if flag := slices.IndexFunc(configEdit.Flags, func(flag contract.FlagCapability) bool { return flag.Name == "--profile" }); flag < 0 || configEdit.Flags[flag].Effective {
		t.Fatalf("config.edit profile flag should be accepted but ineffective: %#v", configEdit.Flags)
	}
	for _, name := range []string{"--editor", "--full", "--scope"} {
		if flag := slices.IndexFunc(configEdit.Flags, func(flag contract.FlagCapability) bool { return flag.Name == name }); flag < 0 || configEdit.Flags[flag].Effective {
			t.Fatalf("config.edit %s flag should be accepted but ineffective in JSON mode: %#v", name, configEdit.Flags)
		}
	}
	rootCapability, err := contract.FindCapability(root, []string{"root"})
	if err != nil {
		t.Fatal(err)
	}
	if flag := slices.IndexFunc(rootCapability.Flags, func(flag contract.FlagCapability) bool { return flag.Name == "--profile" }); flag < 0 || !rootCapability.Flags[flag].Effective || len(rootCapability.Flags[flag].EffectiveWhen) != 1 || rootCapability.Flags[flag].EffectiveWhen[0].AllOf[0].Name != "version" {
		t.Fatalf("root profile flag should be effective: %#v", rootCapability.Flags)
	}
	if flag := slices.IndexFunc(rootCapability.Flags, func(flag contract.FlagCapability) bool { return flag.Name == "--version" }); flag < 0 || !slices.Contains(rootCapability.Flags[flag].Aliases, "-v") {
		t.Fatalf("root version flag should publish -v: %#v", rootCapability.Flags)
	}
	draftShow, err := contract.FindCapability(root, []string{"draft", "show"})
	if err != nil {
		t.Fatal(err)
	}
	if draftShow.Destructive || draftShow.SideEffectLevel != 1 || draftShow.Interaction.Mode != "optional" {
		t.Fatalf("draft.show machine behavior = %#v", draftShow)
	}
	getCapability, err := contract.FindCapability(root, []string{"get"})
	if err != nil {
		t.Fatal(err)
	}
	if getCapability.NetworkAccess != "conditional" || len(getCapability.NetworkWhen) != 2 || getCapability.NetworkWhen[0].AllOf[0].Source != "stdin" || getCapability.NetworkWhen[0].AllOf[0].Operator != "absent" {
		t.Fatalf("get network condition = %#v", getCapability.NetworkWhen)
	}
	if !slices.Equal(getCapability.StdinModes, []string{"json_document"}) {
		t.Fatalf("get stdin modes = %#v", getCapability.StdinModes)
	}
	for _, path := range [][]string{{"auth", "add", "gcloud"}, {"auth", "add", "oauth"}, {"auth", "add", "service-account"}} {
		capability, findErr := contract.FindCapability(root, path)
		if findErr != nil {
			t.Fatal(findErr)
		}
		if !capability.Destructive || len(capability.DestructiveWhen) == 0 {
			t.Fatalf("%v capability = %#v", path, capability)
		}
	}
	for _, path := range [][]string{{"auth", "add", "oauth"}, {"auth", "add", "service-account"}} {
		capability, findErr := contract.FindCapability(root, path)
		if findErr != nil || !slices.Contains(capability.SideEffects, "local_file_write") || !capabilityHasPredicate(capability, "local_file_write", "credential_file", "write_succeeded") {
			t.Fatalf("%v capability credential write = %#v, %v", path, capability.SideEffectWhen, findErr)
		}
	}
	projectImport, err := contract.FindCapability(root, []string{"project", "import"})
	if err != nil {
		t.Fatal(err)
	}
	if projectImport.NetworkAccess != "required" || !slices.Contains(projectImport.SideEffects, "firebase_remote_read") || !slices.Contains(projectImport.SideEffects, "local_draft_write") {
		t.Fatalf("project.import capability = %#v", projectImport)
	}
	if projectImport.StdinSchema == nil || *projectImport.StdinSchema != "urn:fbrcm:schema:cli:1.0.0:stdin:remote_config_import" {
		t.Fatalf("project.import stdin schema = %#v", projectImport.StdinSchema)
	}
	if len(projectImport.InteractionWhen) != 5 || projectImport.InteractionWhen[0].AllOf[1].Name != "confirmation" || projectImport.InteractionWhen[1].AllOf[2].Name != "import_strategy" || projectImport.InteractionWhen[2].AllOf[0].Name != "import_merge_resolution" || !capabilityConditionsHavePredicate(projectImport.InteractionWhen, "runtime_state", "authentication", "requires_human_authorization") {
		t.Fatalf("project.import interaction conditions = %#v", projectImport.InteractionWhen)
	}
	promote, err := contract.FindCapability(root, []string{"projects", "promote"})
	if err != nil {
		t.Fatal(err)
	}
	if len(promote.InteractionWhen) != 3 || promote.InteractionWhen[1].AllOf[0].Name != "promotion_selection" || !capabilityConditionsHavePredicate(promote.InteractionWhen, "runtime_state", "authentication", "requires_human_authorization") {
		t.Fatalf("projects.promote interaction conditions = %#v", promote.InteractionWhen)
	}
	configReset, err := contract.FindCapability(root, []string{"config", "reset"})
	if err != nil {
		t.Fatal(err)
	}
	if configReset.Idempotency != "yes" || !capabilityHasPredicate(configReset, "local_state_write", "mutation_plan", "has_changes") || !capabilityHasPredicate(configReset, "local_state_write", "profile_bootstrap", "required") || len(configReset.InteractionWhen) != 1 || configReset.InteractionWhen[0].AllOf[1].Name != "confirmation" {
		t.Fatalf("config.reset capability = %#v", configReset)
	}
	stdinAdd, err := contract.FindCapability(root, []string{"add"})
	if err != nil {
		t.Fatal(err)
	}
	if stdinAdd.Idempotency != "conditional" || !idempotencyHasPredicate(stdinAdd.IdempotencyWhen, "yes", "stdin", "document", "present") || !idempotencyHasPredicate(stdinAdd.IdempotencyWhen, "yes", "runtime_state", "authentication", "requires_human_authorization") {
		t.Fatalf("add idempotency conditions = %#v", stdinAdd.IdempotencyWhen)
	}
	conditionsValidate, err := contract.FindCapability(root, []string{"conditions", "validate"})
	if err != nil {
		t.Fatal(err)
	}
	if conditionsValidate.NetworkAccess != "required" || !slices.Contains(conditionsValidate.SideEffects, "firebase_remote_validation") {
		t.Fatalf("conditions.validate capability = %#v", conditionsValidate)
	}
	doctor, err := contract.FindCapability(root, []string{"doctor"})
	if err != nil {
		t.Fatal(err)
	}
	if doctor.NetworkAccess != "conditional" || len(doctor.NetworkWhen) != 1 || len(doctor.NetworkWhen[0].AllOf) != 2 || doctor.NetworkWhen[0].AllOf[0].Source != "context" || doctor.NetworkWhen[0].AllOf[0].Name != "offline" || doctor.NetworkWhen[0].AllOf[0].Value != false || doctor.NetworkWhen[0].AllOf[1].Name != "diagnostic_identity" || !slices.Contains(doctor.SideEffects, "local_file_write") || !slices.Contains(doctor.SideEffects, "local_file_delete") || !capabilityHasPredicate(doctor, "authentication_remote_access", "authentication", "requires_network") || capabilityHasPredicate(doctor, "local_file_write", "authentication", "token_persisted") || doctor.Interaction.Mode != "none" {
		t.Fatalf("doctor capability = %#v", doctor)
	}
	draftChangeNote, err := contract.FindCapability(root, []string{"draft", "change-note"})
	if err != nil {
		t.Fatal(err)
	}
	if draftChangeNote.Idempotency != "conditional" || !slices.Contains(draftChangeNote.SideEffects, "local_draft_write") || capabilityHasPredicate(draftChangeNote, "local_state_write", "mutation_plan", "has_changes") || !capabilityEffectHasPredicate(draftChangeNote, "local_draft_write", "argument", "text", "present") || !capabilityEffectHasPredicate(draftChangeNote, "local_draft_write", "option", "clear", "equals") || !idempotencyHasPredicate(draftChangeNote.IdempotencyWhen, "yes", "argument", "text", "absent") || !idempotencyHasPredicate(draftChangeNote.IdempotencyWhen, "no", "argument", "text", "present") {
		t.Fatalf("draft.change-note capability = %#v", draftChangeNote)
	}
	for _, capability := range contract.DetailedCapabilities(root) {
		if capability.NetworkAccess == "none" || capability.ID == "auth.login" || capability.ID == "doctor" {
			continue
		}
		if !capabilityHasPredicate(capability, "authentication_remote_access", "authentication", "requires_network") || !capabilityHasPredicate(capability, "local_file_write", "authentication", "token_persisted") || capability.Interaction.Mode != "optional" || !capabilityConditionsHavePredicate(capability.InteractionWhen, "runtime_state", "authentication", "requires_human_authorization") {
			t.Errorf("%s omits shared machine authentication behavior: %#v", capability.ID, capability)
		}
	}
	for _, test := range []struct {
		path   []string
		effect string
	}{
		{[]string{"cache", "clear"}, "local_cache_delete"},
		{[]string{"draft", "discard"}, "local_draft_delete"},
		{[]string{"profile", "delete"}, "local_file_delete"},
		{[]string{"profile", "delete"}, "local_cache_delete"},
		{[]string{"profile", "delete"}, "local_draft_delete"},
		{[]string{"profile", "rename"}, "local_cache_move"},
		{[]string{"auth", "delete"}, "local_file_delete"},
		{[]string{"auth", "add", "gcloud"}, "local_file_delete"},
		{[]string{"auth", "add", "oauth"}, "local_file_delete"},
		{[]string{"auth", "add", "service-account"}, "local_file_delete"},
		{[]string{"projects", "forget"}, "local_cache_delete"},
		{[]string{"projects", "forget"}, "local_draft_delete"},
		{[]string{"projects", "reset"}, "local_file_delete"},
	} {
		capability, findErr := contract.FindCapability(root, test.path)
		if findErr != nil || !slices.Contains(capability.SideEffects, test.effect) {
			t.Fatalf("%v capability missing %s: %#v, %v", test.path, test.effect, capability.SideEffects, findErr)
		}
	}
	draftCleanupIndex := slices.Index(draftPublish.SideEffects, "local_draft_delete")
	if draftCleanupIndex < 0 || len(draftPublish.SideEffectWhen[draftCleanupIndex].When) != 2 || !slices.ContainsFunc(draftPublish.SideEffectWhen[draftCleanupIndex].When, func(clause contract.BehaviorConditionClause) bool {
		return slices.ContainsFunc(clause.AllOf, func(predicate contract.BehaviorPredicate) bool {
			return predicate.Name == "mutation_plan" && predicate.Operator == "has_no_changes"
		})
	}) {
		t.Fatalf("draft.publish cleanup conditions = %#v", draftPublish.SideEffectWhen)
	}
	profileSwitch, err := contract.FindCapability(root, []string{"profile", "switch"})
	if err != nil || !slices.ContainsFunc(profileSwitch.SideEffectWhen, func(item contract.SideEffectCondition) bool {
		return item.Effect == "local_state_write" && len(item.When) == 0
	}) {
		t.Fatalf("profile.switch should publish an unconditional local write: %#v, %v", profileSwitch, err)
	}
	for _, path := range [][]string{{"projects", "diff"}, {"versions", "diff"}, {"versions", "list"}, {"versions", "show"}} {
		capability, findErr := contract.FindCapability(root, path)
		if findErr != nil {
			t.Fatal(findErr)
		}
		wantClauses := 1
		wantPredicates := 1
		if path[0] == "versions" && path[1] != "list" {
			wantClauses, wantPredicates = 2, 2
		}
		if len(capability.NetworkWhen) != wantClauses || len(capability.NetworkWhen[0].AllOf) != wantPredicates {
			t.Fatalf("%v network conditions = %#v", path, capability.NetworkWhen)
		}
		cached := capability.NetworkWhen[0].AllOf[0]
		if cached.Source != "option" || cached.Name != "cached" || cached.Operator != "equals" || cached.Value != false {
			t.Fatalf("%v cached predicate = %#v", path, cached)
		}
		if wantPredicates == 2 && capability.NetworkWhen[0].AllOf[1].Name != "version_request" {
			t.Fatalf("%v version request predicate = %#v", path, capability.NetworkWhen[0].AllOf[1])
		}
		if wantClauses == 2 && capability.NetworkWhen[1].AllOf[1].Name != "project_registry" {
			t.Fatalf("%v project registry predicate = %#v", path, capability.NetworkWhen[1].AllOf[1])
		}
		if !capabilityHasPredicate(capability, "local_state_write", "project_registry", "sync_write_succeeded") {
			t.Fatalf("%v omits project-registry persistence: %#v", path, capability.SideEffectWhen)
		}
		for _, flag := range capability.Flags {
			if flag.Name == "--cached" {
				if flag.Default != false {
					t.Fatalf("%v cached default is not typed: %#v", path, flag.Default)
				}
				break
			}
		}
	}
	draftDiff, err := contract.FindCapability(root, []string{"draft", "diff"})
	if err != nil {
		t.Fatal(err)
	}
	if len(draftDiff.NetworkWhen) != 1 || len(draftDiff.NetworkWhen[0].AllOf) != 2 || draftDiff.NetworkWhen[0].AllOf[0].Name != "against" || draftDiff.NetworkWhen[0].AllOf[1].Name != "cached" {
		t.Fatalf("draft.diff network conditions = %#v", draftDiff.NetworkWhen)
	}
	conditionAdd, err := contract.FindCapability(root, []string{"conditions", "add"})
	if err != nil {
		t.Fatal(err)
	}
	for _, effect := range []string{"firebase_remote_read", "firebase_remote_validation", "firebase_remote_write", "local_draft_write", "trusted_hook_execution"} {
		if !slices.Contains(conditionAdd.SideEffects, effect) {
			t.Errorf("conditions.add is missing %s: %#v", effect, conditionAdd.SideEffects)
		}
	}
	for _, item := range conditionAdd.SideEffectWhen {
		if item.Effect != "trusted_hook_execution" {
			continue
		}
		if len(item.When) != 2 || slices.ContainsFunc(item.When[0].AllOf, func(predicate contract.BehaviorPredicate) bool {
			return predicate.Name == "publication"
		}) || !slices.ContainsFunc(item.When[1].AllOf, func(predicate contract.BehaviorPredicate) bool {
			return predicate.Name == "publication" && predicate.Operator == "accepted"
		}) {
			t.Fatalf("conditions.add hook conditions = %#v", item.When)
		}
	}
	for _, path := range [][]string{{"experiments", "list"}, {"experiments", "show"}, {"rollouts", "list"}, {"rollouts", "show"}} {
		capability, findErr := contract.FindCapability(root, path)
		if findErr != nil || capability.NetworkAccess != "required" || !slices.Contains(capability.SideEffects, "firebase_remote_read") {
			t.Fatalf("%v managed-feature capability = %#v, %v", path, capability, findErr)
		}
	}
}

func capabilityHasPredicate(capability contract.Capability, effect, name, operator string) bool {
	return capabilityEffectHasPredicate(capability, effect, "runtime_state", name, operator)
}

func capabilityEffectHasPredicate(capability contract.Capability, effect, source, name, operator string) bool {
	for _, item := range capability.SideEffectWhen {
		if item.Effect != effect {
			continue
		}
		if capabilityConditionsHavePredicate(item.When, source, name, operator) {
			return true
		}
	}
	return false
}

func capabilityConditionsHavePredicate(conditions []contract.BehaviorConditionClause, source, name, operator string) bool {
	for _, clause := range conditions {
		if slices.ContainsFunc(clause.AllOf, func(predicate contract.BehaviorPredicate) bool {
			return predicate.Source == source && predicate.Name == name && predicate.Operator == operator
		}) {
			return true
		}
	}
	return false
}

func idempotencyHasPredicate(conditions []contract.IdempotencyCondition, idempotency, source, name, operator string) bool {
	return slices.ContainsFunc(conditions, func(condition contract.IdempotencyCondition) bool {
		return condition.Idempotency == idempotency && capabilityConditionsHavePredicate(condition.When, source, name, operator)
	})
}

func TestDetailedCapabilitiesConformToStandaloneSchema(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":capability"
	for _, capability := range contract.DetailedCapabilities(NewRootForContract("schema")) {
		validateContractValue(t, id, structToContractValue(t, capability), true)
		unconditionalLocalWrite := slices.ContainsFunc(capability.SideEffectWhen, func(item contract.SideEffectCondition) bool {
			return item.Effect == "local_state_write" && len(item.When) == 0
		})
		if capability.SideEffectLevel < 1 || !unconditionalLocalWrite && !capabilityHasPredicate(capability, "local_state_write", "profile_bootstrap", "required") {
			t.Fatalf("%s omits JSON envelope profile bootstrap: %#v", capability.ID, capability.SideEffectWhen)
		}
		if capability.SideEffects == nil || capability.SideEffectWhen == nil || capability.NetworkWhen == nil || capability.DestructiveWhen == nil || capability.IdempotencyWhen == nil || capability.StdinModes == nil || capability.InteractionWhen == nil {
			t.Fatalf("%s has nullable machine arrays: %#v", capability.ID, capability)
		}
		if capability.NetworkAccess == "conditional" && len(capability.NetworkWhen) == 0 {
			t.Fatalf("%s has conditional network access without conditions", capability.ID)
		}
		if len(capability.SideEffects) != len(capability.SideEffectWhen) {
			t.Fatalf("%s side effects lack per-effect conditions: %#v", capability.ID, capability)
		}
		for index, effect := range capability.SideEffects {
			if capability.SideEffectWhen[index].Effect != effect {
				t.Fatalf("%s side-effect condition %d describes %q, want %q", capability.ID, index, capability.SideEffectWhen[index].Effect, effect)
			}
		}
	}
}

func TestDetailedCapabilityGoldenCoversAuthoritativeBehavior(t *testing.T) {
	root := NewRootForContract("schema")
	got, err := json.MarshalIndent(contract.DetailedCapabilities(root), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	want, err := os.ReadFile("testdata/contract_v1_capabilities_detailed.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("detailed capability behavior changed; run go run ./cmd/schemagen and review the golden diff")
	}
}

func TestJSONEnvelopeIsSingleMachineReadableDocument(t *testing.T) {
	envelope, raw := executeJSONContract(t, "capabilities", "--json")
	if envelope.ContractVersion != contract.Version || envelope.Command != "capabilities" || envelope.Outcome != "success" || envelope.ExitCode != 0 {
		t.Fatalf("envelope = %#v", envelope)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var decoded contract.Envelope
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("decode envelope: %v\n%s", err, raw)
	}
	if decoder.More() {
		t.Fatalf("more than one JSON document: %s", raw)
	}
	if !bytes.HasSuffix(raw, []byte("\n")) || strings.Contains(string(raw), "Usage:") {
		t.Fatalf("invalid machine stdout: %q", raw)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestEnvelopeSchemaRejectsContradictoryOutcomeState(t *testing.T) {
	envelope, raw := executeJSONContract(t, "capabilities", "--json")
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document["outcome"] = "failure"
	validateContractValue(t, envelope.Schema, document, false)

	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document["outcome"] = "partial_success"
	document["exit_code"] = float64(12)
	document["errors"] = []any{}
	validateContractValue(t, envelope.Schema, document, false)

	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document["outcome"] = "failure"
	document["exit_code"] = float64(15)
	document["data"] = nil
	document["errors"] = []any{map[string]any{
		"code": "argument.invalid", "category": "argument", "message": "bad argument", "retryable": false,
		"target": nil, "stage": nil, "details": nil, "remediation": []any{},
	}}
	validateContractValue(t, envelope.Schema, document, false)
}

func TestEnvelopeSchemaConstrainsKnownWarningDetails(t *testing.T) {
	document := map[string]any{
		"schema": contract.SchemaID("get"), "contract_version": contract.Version,
		"command": "get", "requested_command": "get", "outcome": "success", "exit_code": 0,
		"producer": map[string]any{"name": "fbrcm", "version": "test"},
		"context":  map[string]any{"profile": "default", "offline": false, "dry_run": false, "draft": false},
		"data":     map[string]any{"count": 0, "items": []any{}},
		"errors":   []any{}, "warnings": []any{},
	}
	document["warnings"] = []any{map[string]any{
		"code": "cache.stale", "message": "cache update failed", "target": "demo", "details": nil, "remediation": []any{},
	}}
	validateContractValue(t, contract.SchemaID("get"), document, false)
	document["warnings"].([]any)[0].(map[string]any)["details"] = map[string]any{"source": "expired-cache"}
	validateContractValue(t, contract.SchemaID("get"), document, true)
}

func TestPublishedProblemSchemaCatalogsCodesAndDiscriminatesDetails(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":error"
	problem := contract.Classify(&core.OAuthInteractionError{AuthID: "personal", Err: errors.New("authorization required")})
	value := structToContractValue(t, problem)
	validateContractValue(t, id, value, true)
	value.(map[string]any)["details"] = map[string]any{"kind": "oauth_authorization"}
	validateContractValue(t, id, value, false)

	raw, err := schemas.ReadByID(id)
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	properties := schema["properties"].(map[string]any)
	code := properties["code"].(map[string]any)
	known := code["enum"].([]any)
	if !slices.Contains(known, any("interaction.required")) {
		t.Fatalf("known problem codes = %#v", known)
	}
}

func TestCapabilityDiscoveryIsCompactAndExact(t *testing.T) {
	root := NewRootForContract("test")
	index := contract.Capabilities(root)
	raw, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(`"flags"`)) || bytes.Contains(raw, []byte(`"arguments"`)) || len(raw) > 100_000 {
		t.Fatalf("capability index is not compact: %d bytes", len(raw))
	}
	capability, err := contract.FindCapability(root, []string{"projects", "aliases", "set"})
	if err != nil || capability.ID != "projects.aliases.set" || len(capability.Flags) == 0 {
		t.Fatalf("exact capability = %#v, %v", capability, err)
	}
	rootCapability, err := contract.FindCapability(root, []string{"root"})
	if err != nil || rootCapability.ID != "root" || len(rootCapability.Path) != 0 {
		t.Fatalf("root capability = %#v, %v", rootCapability, err)
	}
	for _, test := range []struct {
		path     []string
		category string
		code     int
	}{
		{path: []string{"does-not-exist"}, category: "not_found", code: 6},
		{path: []string{"projects"}, category: "argument", code: 2},
	} {
		_, err := contract.FindCapability(root, test.path)
		if err == nil || contract.Classify(err).Category != test.category || contract.ExitCode(nil, err) != test.code {
			t.Fatalf("FindCapability(%v) = %v, category %q, exit %d", test.path, err, contract.Classify(err).Category, contract.ExitCode(nil, err))
		}
	}
}

func TestDiscoveryLookupFailuresAreTyped(t *testing.T) {
	for _, args := range [][]string{
		{"capabilities", "does-not-exist", "--json"},
		{"schema", "show", "urn:fbrcm:schema:missing", "--json"},
	} {
		envelope, raw := executeJSONContract(t, args...)
		if envelope.Outcome != "failure" || envelope.ExitCode != 6 || len(envelope.Errors) != 1 || envelope.Errors[0].Category != "not_found" {
			t.Fatalf("%v envelope = %#v", args, envelope)
		}
		validateContractDocument(t, envelope.Schema, raw)
	}
}

func TestJSONHelpAndVersionUseConformingResponseSchemas(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCommand string
	}{
		{name: "implicit command help", args: []string{"update", "--help", "--json"}, wantCommand: "help"},
		{name: "command group help", args: []string{"projects", "--json"}, wantCommand: "help"},
		{name: "root help", args: []string{"--help", "--json"}, wantCommand: "help"},
		{name: "root version", args: []string{"--version", "--json"}, wantCommand: "root"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope, raw := executeJSONContract(t, test.args...)
			if envelope.Command != test.wantCommand || envelope.Outcome != "success" || envelope.ExitCode != 0 {
				t.Fatalf("envelope = %#v", envelope)
			}
			validateContractDocument(t, envelope.Schema, raw)
		})
	}
}

func TestInvalidChangeNoteIsTypedArgumentFailure(t *testing.T) {
	envelope, raw := executeJSONContract(t, "add", "flag", "--type", "string", "--value", "on", "--change-note", "line one\nline two", "--yes", "--json")
	if envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" || envelope.Errors[0].Category != "argument" {
		t.Fatalf("invalid change-note envelope = %#v", envelope)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestJSONUnknownCommandBelowGroupUsesPublishedRootErrorSchema(t *testing.T) {
	envelope, raw := executeJSONContract(t, "projects", "does-not-exist", "--json")
	if envelope.Command != "root" || envelope.RequestedCommand != "projects.does-not-exist" || envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.unknown_command" || envelope.Errors[0].Category != "argument" {
		t.Fatalf("envelope = %#v", envelope)
	}
	detailsRaw, err := json.Marshal(envelope.Errors[0].Details)
	if err != nil {
		t.Fatal(err)
	}
	var details map[string]any
	if err := json.Unmarshal(detailsRaw, &details); err != nil {
		t.Fatal(err)
	}
	if details["kind"] != "invocation" || details["requested_command"] != "projects.does-not-exist" || details["resolved_command"] != "projects" {
		t.Fatalf("unknown-command details = %#v", envelope.Errors[0].Details)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestRuntimeRejectsMismatchedCountCorrelation(t *testing.T) {
	root := NewRootForContract("test")
	cmd, remaining, err := root.Find([]string{"capabilities"})
	if err != nil || len(remaining) != 0 {
		t.Fatalf("find capabilities: error=%v remaining=%v", err, remaining)
	}
	envelope := contract.BuildEnvelope(cmd, "test", []byte(`{"contract_version":"1.0.0","count":1,"commands":[]}`), nil)
	if envelope.Outcome != "failure" || envelope.ExitCode != 15 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "internal.contract_violation" || !strings.Contains(envelope.Errors[0].Message, "count correlation") {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestCapabilitiesResponseUsesStrictStandaloneCapabilitySchema(t *testing.T) {
	_, raw := executeJSONContract(t, "capabilities", "project", "import", "--json")
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	data := document["data"].(map[string]any)
	data["network_when"] = nil
	validateContractValue(t, contract.SchemaID("capabilities"), document, false)
}

func TestProjectAliasSourceValuesConformToResponseSchema(t *testing.T) {
	root := NewRootForContract("test")
	cmd, remaining, err := root.Find([]string{"projects", "aliases", "list"})
	if err != nil || len(remaining) != 0 {
		t.Fatalf("find aliases list: error=%v remaining=%v", err, remaining)
	}
	captured := []byte(`[
		{"alias":"fbrcm","project_id":"one","source":"fbrcm"},
		{"alias":"firebase","project_id":"two","source":"firebase"},
		{"alias":"both","project_id":"three","source":"both"}
	]`)
	envelope := contract.BuildEnvelope(cmd, "test", captured, nil)
	validateContractDocument(t, envelope.Schema, marshalEnvelope(t, envelope))
}

func TestEveryExecutableCommandFailureEnvelopeConformsToItsSchema(t *testing.T) {
	capabilities := contract.DetailedCapabilities(NewRootForContract("schema"))
	for _, capability := range capabilities {
		t.Run(capability.ID, func(t *testing.T) {
			args := append(append([]string(nil), capability.Path...), "--contract-conformance-unknown-flag", "--json")
			envelope, raw := executeJSONContract(t, args...)
			if envelope.Command != capability.ID || envelope.Outcome != "failure" || envelope.ExitCode != 2 {
				t.Fatalf("envelope = %#v", envelope)
			}
			validateContractDocument(t, capability.ResponseSchema, raw)
		})
	}
}

func TestEmptyCollectionAndNoOpRuntimeEnvelopesConform(t *testing.T) {
	root := NewRootForContract("test")
	projectsList, _, err := root.Find([]string{"projects", "list"})
	if err != nil {
		t.Fatal(err)
	}
	empty := contract.BuildEnvelope(projectsList, "test", []byte(`[]`), nil)
	emptyRaw := marshalEnvelope(t, empty)
	validateContractDocument(t, empty.Schema, emptyRaw)
	data, ok := empty.Data.(struct {
		Count int               `json:"count"`
		Items []json.RawMessage `json:"items"`
	})
	if !ok || data.Count != 0 || len(data.Items) != 0 {
		t.Fatalf("empty data = %#v", empty.Data)
	}

	update, _, err := root.Find([]string{"update"})
	if err != nil {
		t.Fatal(err)
	}
	reason := sharedrc.NoOpAlreadyApplied
	result := sharedrc.RemoteMutationJSONResult{
		Target:           "demo",
		Status:           sharedrc.RemoteMutationUnchanged,
		ChangedItemCount: 0,
		Validated:        true,
		ValidationSource: core.ValidationSourceLocal,
		Selection:        sharedrc.SelectionMetadata{DefaultScope: true, ResolvedTargetCount: 1, MatchedItemCount: 17},
		NoOpReason:       &reason,
	}
	captured, err := json.Marshal([]sharedrc.RemoteMutationJSONResult{result})
	if err != nil {
		t.Fatal(err)
	}
	noOp := contract.BuildEnvelope(update, "test", captured, nil)
	validateContractDocument(t, noOp.Schema, marshalEnvelope(t, noOp))
}

func TestManagedFeatureDeletePreviewFailureRetainsConformingData(t *testing.T) {
	root := NewRootForContract("test")
	for _, path := range [][]string{{"experiments", "delete"}, {"rollouts", "delete"}} {
		cmd, remaining, err := root.Find(path)
		if err != nil || len(remaining) != 0 {
			t.Fatalf("find %v: error=%v remaining=%v", path, err, remaining)
		}
		kind := path[0][:len(path[0])-1]
		captured := []byte(`{"kind":"` + kind + `","id":"7","display_name":"Signup","project_id":"demo","status":"would-delete"}`)
		envelope := contract.BuildEnvelope(cmd, "test", captured, shared.InteractionRequired("confirmation is required", true, "--yes"))
		if envelope.Outcome != "failure" || envelope.ExitCode != 10 || envelope.Data == nil {
			t.Fatalf("%v preview envelope = %#v", path, envelope)
		}
		validateContractDocument(t, envelope.Schema, marshalEnvelope(t, envelope))
		document := structToContractValue(t, envelope).(map[string]any)
		document["data"].(map[string]any)["status"] = "deleted"
		validateContractValue(t, envelope.Schema, document, false)

		deleted := []byte(`{"kind":"` + kind + `","id":"7","display_name":"Signup","project_id":"demo","status":"deleted"}`)
		success := contract.BuildEnvelope(cmd, "test", deleted, nil)
		validateContractDocument(t, success.Schema, marshalEnvelope(t, success))
		document = structToContractValue(t, success).(map[string]any)
		document["data"].(map[string]any)["status"] = "would-delete"
		validateContractValue(t, success.Schema, document, false)
	}
}

func TestTypedFailureScenariosConformToRuntimeSchema(t *testing.T) {
	root := NewRootForContract("test")
	update, _, err := root.Find([]string{"update"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		err  error
	}{
		{"invalid expression", &shared.ExpressionError{Expression: "key ==", Context: "parameter", Err: errors.New("unexpected end")}},
		{"interaction", shared.InteractionRequired("confirmation required", true, "--yes")},
		{"local validation", &core.RemoteConfigValidationError{Source: core.ValidationSourceLocal, Err: errors.New("invalid local candidate")}},
		{"firebase validation", &core.RemoteConfigValidationError{Source: core.ValidationSourceFirebase, Err: errors.New("Firebase rejected candidate")}},
		{"auth missing", &core.AuthError{Kind: "not_found", AuthID: "missing", Err: errors.New("auth is not configured")}},
		{"auth setup", &core.AuthError{Kind: "setup_required", Err: errors.New("no auth identities configured")}},
		{"auth configuration", &core.AuthError{Kind: "configuration", Err: errors.New("invalid auth registry")}},
		{"oauth authorization", &core.OAuthInteractionError{AuthID: "personal", Err: errors.New("authorization required")}},
		{"timeout", context.DeadlineExceeded},
		{"cancellation", context.Canceled},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := contract.BuildEnvelope(update, "test", nil, test.err)
			validateContractDocument(t, envelope.Schema, marshalEnvelope(t, envelope))
		})
	}
}

func TestPostPublicationFailureEnvelopesAndWarningsConform(t *testing.T) {
	root := NewRootForContract("test")
	update, _, err := root.Find([]string{"update"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		status  sharedrc.RemoteMutationStatus
		stage   string
		err     error
		warning shared.MachineWarning
	}{
		{
			name:   "cache",
			status: sharedrc.RemoteMutationPublishedCacheFailed,
			stage:  "cache",
			err:    &core.RemoteConfigPublishedCacheError{ProjectID: "demo", Err: errors.New("disk full")},
			warning: shared.MachineWarning{Code: "publication.cache_stale", Message: "published but cache refresh failed", Target: "demo", Details: map[string]any{"stage": "cache"}, Remediation: []shared.Remediation{{
				Description: "refresh the cache", Strategy: shared.RemediationRunCommand, Argv: []string{"get", "--update", "--project", "=demo"},
			}}},
		},
		{
			name:   "hook",
			status: sharedrc.RemoteMutationPublishedHookFailed,
			stage:  "post_publish_hook",
			err:    &core.RemoteConfigPublishedHookError{ProjectID: "demo", HookErr: &corehooks.Error{Event: corehooks.PostPublish, Err: errors.New("exit 1")}},
			warning: shared.MachineWarning{Code: "publication.post_publish_hook_failed", Message: "published but hook failed", Target: "demo", Details: map[string]any{"stage": "post_publish_hook"}, Remediation: []shared.Remediation{{
				Description: "inspect hook status", Strategy: shared.RemediationRunCommand, Argv: []string{"hooks", "status"},
			}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			update.SetContext(shared.WithMachineState(context.Background()))
			shared.AddMachineWarning(update, test.warning)
			publishedVersion := "42"
			payload := []sharedrc.RemoteMutationJSONResult{{
				Target: "demo", Status: test.status, ChangedItemCount: 1, Validated: true, ValidationSource: core.ValidationSourceFirebase,
				PublishedVersion: &publishedVersion,
				Selection:        sharedrc.SelectionMetadata{ResolvedTargetCount: 1, MatchedItemCount: 1},
				Error:            &sharedrc.RemoteMutationJSONError{Stage: test.stage, Message: test.err.Error()},
			}}
			captured, marshalErr := json.Marshal(payload)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			envelope := contract.BuildEnvelope(update, "test", captured, test.err)
			if envelope.Outcome != "partial_success" || envelope.ExitCode != 12 || len(envelope.Warnings) != 1 {
				t.Fatalf("envelope = %#v", envelope)
			}
			validateContractDocument(t, envelope.Schema, marshalEnvelope(t, envelope))
		})
	}
}

func TestSemanticInvocationSchemasRejectInvalidCombinations(t *testing.T) {
	updateID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:update:input"
	validateContractValue(t, updateID, map[string]any{
		"arguments": map[string]any{},
		"options":   map[string]any{"condition": "beta", "value": "on", "type": "string"},
		"stdin":     nil,
	}, true)
	validateContractValue(t, updateID, map[string]any{
		"arguments": map[string]any{},
		"options":   map[string]any{"condition": "beta"},
		"stdin":     nil,
	}, false)
	validateContractValue(t, updateID, map[string]any{
		"arguments": map[string]any{},
		"options":   map[string]any{"condition": "beta", "use-in-app-default": false},
		"stdin":     nil,
	}, false)
	validateContractValue(t, updateID, map[string]any{
		"arguments": map[string]any{},
		"options":   map[string]any{"condition": "beta", "value": "on", "type": "binary"},
		"stdin":     nil,
	}, false)
	validateContractValue(t, updateID, map[string]any{
		"arguments": map[string]any{},
		"options":   map[string]any{"condition": "beta", "remove-all-conditional-values": true, "value": "on", "type": "string"},
		"stdin":     nil,
	}, false)
}

func TestStdinMutationSchemasRejectIgnoredRemoteOptions(t *testing.T) {
	stdin := map[string]any{"parameters": map[string]any{}}
	for _, commandID := range []string{"add", "update", "delete"} {
		arguments := map[string]any{}
		options := map[string]any{}
		if commandID == "add" {
			arguments["parameter"] = "flag"
			options = map[string]any{"value": "on", "type": "string"}
		}
		id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:" + commandID + ":input"
		for _, unavailable := range []map[string]any{
			{"project": []any{"demo"}},
			{"project": []any{}},
			{"dry-run": true},
			{"dry-run": false},
			{"draft": true},
			{"draft": false},
			{"change-note": "release"},
			{"change-note": ""},
		} {
			candidate := maps.Clone(options)
			maps.Copy(candidate, unavailable)
			validateContractValue(t, id, map[string]any{"arguments": arguments, "options": candidate, "stdin": stdin}, false)
		}
		validateContractValue(t, id, map[string]any{"arguments": arguments, "options": options, "stdin": stdin}, true)
	}
}

func TestInvocationSchemasDistinguishSelectorsFromLiteralIdentifiers(t *testing.T) {
	aliasID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:projects.aliases.set:input"
	base := func(alias, projectID string) map[string]any {
		return map[string]any{"arguments": map[string]any{"alias": alias, "project_id": projectID}, "options": map[string]any{}, "stdin": nil}
	}
	validateContractValue(t, aliasID, base("prod", "physical-project"), true)
	for _, invalid := range []map[string]any{base("Prod", "physical-project"), base("prod", "=physical-project"), base("prod", "server@physical-project")} {
		validateContractValue(t, aliasID, invalid, false)
	}
	validateContractValue(t, aliasID, base("prod", "$physical-project"), true)

	versionID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:versions.diff:input"
	versionInput := func(from string) map[string]any {
		return map[string]any{"arguments": map[string]any{"project": "demo", "from": from}, "options": map[string]any{}, "stdin": nil}
	}
	validateContractValue(t, versionID, versionInput("current~299"), true)
	validateContractValue(t, versionID, versionInput("current~+00299"), true)
	validateContractValue(t, versionID, versionInput("current~300"), false)
	validateContractValue(t, versionID, versionInput("+00042"), true)

	projectID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:project.show:input"
	projectInput := func(project string) map[string]any {
		return map[string]any{"arguments": map[string]any{"project": project}, "options": map[string]any{}, "stdin": nil}
	}
	validateContractValue(t, projectID, projectInput("SERVER@demo"), true)
	validateContractValue(t, projectID, projectInput("server@=demo"), true)
	validateContractValue(t, projectID, projectInput("server@"), true)

	getID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:get:input"
	validateContractValue(t, getID, map[string]any{"arguments": map[string]any{}, "options": map[string]any{"filter": []any{"/"}}, "stdin": nil}, false)

	authBindID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:auth.bind:input"
	authBindInput := func(project string) map[string]any {
		return map[string]any{"arguments": map[string]any{}, "options": map[string]any{"auth": "main", "project": []any{project}}, "stdin": nil}
	}
	validateContractValue(t, authBindID, authBindInput("=demo"), true)
	validateContractValue(t, authBindID, authBindInput("server@=demo"), false)
}

func TestPhysicalProjectSchemasRejectTemplatePrefixes(t *testing.T) {
	tests := []struct {
		id      string
		args    map[string]any
		options map[string]any
	}{
		{"project.templates.show", map[string]any{"project": "demo"}, map[string]any{}},
		{"project.templates.set", map[string]any{"project": "demo"}, map[string]any{"templates": []any{"client"}}},
		{"experiments.list", map[string]any{"project": "demo"}, map[string]any{}},
		{"experiments.show", map[string]any{"project": "demo", "experiment_id": "exp"}, map[string]any{}},
		{"experiments.delete", map[string]any{"project": "demo", "experiment_id": "exp"}, map[string]any{}},
		{"rollouts.list", map[string]any{"project": "demo"}, map[string]any{}},
		{"rollouts.show", map[string]any{"project": "demo", "rollout_id": "rollout"}, map[string]any{}},
		{"rollouts.delete", map[string]any{"project": "demo", "rollout_id": "rollout"}, map[string]any{}},
		{"personalizations.list", map[string]any{"project": "demo"}, map[string]any{}},
		{"personalizations.show", map[string]any{"project": "demo", "personalization_id": "personalization"}, map[string]any{}},
	}
	for _, test := range tests {
		id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:" + test.id + ":input"
		input := map[string]any{"arguments": test.args, "options": test.options, "stdin": nil}
		validateContractValue(t, id, input, true)
		for _, prefix := range []string{"client@", "SERVER@"} {
			prefixedArgs := maps.Clone(test.args)
			prefixedArgs["project"] = prefix + "demo"
			validateContractValue(t, id, map[string]any{"arguments": prefixedArgs, "options": test.options, "stdin": nil}, false)
		}
		paddedArgs := maps.Clone(test.args)
		paddedArgs["project"] = " client@demo"
		validateContractValue(t, id, map[string]any{"arguments": paddedArgs, "options": test.options, "stdin": nil}, true)
		for key := range test.args {
			if strings.HasSuffix(key, "_id") {
				blankArgs := maps.Clone(test.args)
				blankArgs[key] = "   "
				validateContractValue(t, id, map[string]any{"arguments": blankArgs, "options": test.options, "stdin": nil}, false)
			}
		}
	}
}

func TestProjectArgumentSchemasMatchRuntimeResolverKinds(t *testing.T) {
	tests := map[string]string{
		"project.open":           "project_positional_selector",
		"project.show":           "project_positional_selector",
		"project.templates.show": "physical_project_selector",
		"project.templates.set":  "physical_project_selector",
		"experiments.list":       "physical_project_selector",
		"experiments.show":       "physical_project_selector",
		"experiments.delete":     "physical_project_selector",
		"rollouts.list":          "physical_project_selector",
		"rollouts.show":          "physical_project_selector",
		"rollouts.delete":        "physical_project_selector",
		"personalizations.list":  "physical_project_selector",
		"personalizations.show":  "physical_project_selector",
	}
	for commandID, definition := range tests {
		raw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:" + commandID + ":input")
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		properties := document["properties"].(map[string]any)
		arguments := properties["arguments"].(map[string]any)
		argumentProperties := arguments["properties"].(map[string]any)
		project := argumentProperties["project"].(map[string]any)
		if got, want := project["$ref"], "#/$defs/"+definition; got != want {
			t.Errorf("%s project schema = %v, want %q", commandID, got, want)
		}
	}
}

func TestLiteralExistingResourceArgumentsPublishNoTrimNormalization(t *testing.T) {
	for commandID, argumentName := range map[string]string{
		"conditions.show":       "condition",
		"duplicate":             "source",
		"experiments.show":      "experiment_id",
		"groups.edit":           "group",
		"personalizations.show": "personalization_id",
	} {
		raw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:" + commandID + ":input")
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		argument := document["properties"].(map[string]any)["arguments"].(map[string]any)["properties"].(map[string]any)[argumentName].(map[string]any)
		if argument["x-fbrcm-normalization"] != nil {
			t.Errorf("%s argument %s falsely publishes normalization: %#v", commandID, argumentName, argument["x-fbrcm-normalization"])
		}
	}

	raw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:versions.show:input")
	if err != nil {
		t.Fatal(err)
	}
	var versions map[string]any
	if err := json.Unmarshal(raw, &versions); err != nil {
		t.Fatal(err)
	}
	versionSelector := versions["$defs"].(map[string]any)["version_selector"].(map[string]any)
	if versionSelector["x-fbrcm-normalization"] != nil {
		t.Fatalf("version selector falsely publishes normalization: %#v", versionSelector["x-fbrcm-normalization"])
	}
}

func TestCommandResponseSchemasRejectSuccessWithNullData(t *testing.T) {
	root := NewRootForContract("test")
	cmd, _, err := root.Find([]string{"projects", "list"})
	if err != nil {
		t.Fatal(err)
	}
	envelope := contract.BuildEnvelope(cmd, "test", []byte(`[]`), nil)
	document := structToContractValue(t, envelope).(map[string]any)
	document["data"] = nil
	validateContractValue(t, envelope.Schema, document, false)

	statusOne := contract.BuildEnvelope(cmd, "test", nil, shared.WithExitCode(nil, 1))
	if statusOne.Outcome != "failure" || statusOne.ExitCode != 15 || len(statusOne.Errors) != 1 || statusOne.Errors[0].Code != "internal.contract_violation" {
		t.Fatalf("invalid command result was not converted to a contract violation: %#v", statusOne)
	}
	validateContractValue(t, statusOne.Schema, structToContractValue(t, statusOne), true)
	doctor, _, err := root.Find([]string{"doctor"})
	if err != nil {
		t.Fatal(err)
	}
	diagnosticFailure := contract.BuildEnvelope(doctor, "test", nil, shared.WithExitCode(nil, 1))
	validateContractValue(t, diagnosticFailure.Schema, structToContractValue(t, diagnosticFailure), true)
}

func TestBuildEnvelopeConvertsMismatchedRegisteredPayloadToContractViolation(t *testing.T) {
	root := NewRootForContract("test")
	cmd, remaining, err := root.Find([]string{"projects", "list"})
	if err != nil || len(remaining) != 0 {
		t.Fatalf("find projects list: error=%v remaining=%v", err, remaining)
	}
	envelope := contract.BuildEnvelope(cmd, "test", []byte(`{"unexpected":true}`), nil)
	if envelope.Outcome != "failure" || envelope.ExitCode != 15 || envelope.Data != nil || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "internal.contract_violation" {
		t.Fatalf("envelope = %#v", envelope)
	}
	validateContractValue(t, envelope.Schema, structToContractValue(t, envelope), true)
}

func TestDiffResponseSchemasCorrelateChangedAndExitCode(t *testing.T) {
	for _, command := range []string{"draft.diff", "projects.diff", "versions.diff"} {
		t.Run(command, func(t *testing.T) {
			schemaID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:" + command + ":response"
			raw, err := schemas.ReadByID(schemaID)
			if err != nil {
				t.Fatal(err)
			}
			var document map[string]any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			constraints := document["allOf"].([]any)[1].(map[string]any)["allOf"].([]any)
			got := make(map[bool]int)
			for _, rawConstraint := range constraints {
				constraint := rawConstraint.(map[string]any)
				ifSchema, ok := constraint["if"].(map[string]any)
				if !ok {
					continue
				}
				properties, ok := ifSchema["properties"].(map[string]any)
				if !ok {
					continue
				}
				data, ok := properties["data"].(map[string]any)
				if !ok {
					continue
				}
				dataProperties, ok := data["properties"].(map[string]any)
				if !ok {
					continue
				}
				changedSchema, ok := dataProperties["changed"].(map[string]any)
				if !ok {
					continue
				}
				changed, ok := changedSchema["const"].(bool)
				if !ok {
					continue
				}
				thenSchema := constraint["then"].(map[string]any)
				exitCode := thenSchema["properties"].(map[string]any)["exit_code"].(map[string]any)["const"].(float64)
				got[changed] = int(exitCode)
			}
			if got[false] != 0 || got[true] != 1 || len(got) != 2 {
				t.Fatalf("changed/exit-code constraints = %#v, want false:0 and true:1", got)
			}
		})
	}
}

func TestAliasesImportInvocationSchemaMatchesNoPositionalRuntime(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:projects.aliases.import:input"
	validateContractValue(t, id, map[string]any{
		"arguments": map[string]any{},
		"options":   map[string]any{"from": ".firebaserc"},
		"stdin":     nil,
	}, true)
	validateContractValue(t, id, map[string]any{
		"arguments": map[string]any{"path": ".firebaserc"},
		"options":   map[string]any{"from": ".firebaserc"},
		"stdin":     nil,
	}, false)
}

func TestInvocationSchemasEncodeRuntimeSemanticRequirements(t *testing.T) {
	tests := []struct {
		id        string
		arguments map[string]any
		options   map[string]any
		valid     bool
	}{
		{"conditions.edit", map[string]any{"project": "demo", "condition": "beta"}, map[string]any{}, false},
		{"conditions.edit", map[string]any{"project": "demo", "condition": "beta"}, map[string]any{"expression": "true"}, true},
		{"conditions.edit", map[string]any{"project": "demo", "condition": "beta"}, map[string]any{"no-color": false}, false},
		{"conditions.edit", map[string]any{"project": "demo", "condition": "beta"}, map[string]any{"color": "BLUE", "no-color": false}, true},
		{"conditions.edit", map[string]any{"project": "demo", "condition": "beta"}, map[string]any{"color": "BLUE", "no-color": true}, false},
		{"groups.edit", map[string]any{"group": "checkout"}, map[string]any{}, false},
		{"groups.edit", map[string]any{"group": "checkout"}, map[string]any{"no-description": true}, true},
		{"groups.edit", map[string]any{"group": "checkout"}, map[string]any{"no-description": false}, false},
		{"groups.edit", map[string]any{"group": "checkout"}, map[string]any{"description": "kept", "no-description": false}, true},
		{"groups.edit", map[string]any{"group": "checkout"}, map[string]any{"description": "removed", "no-description": true}, false},
		{"project.templates.set", map[string]any{"project": "demo"}, map[string]any{}, false},
		{"project.templates.set", map[string]any{"project": "demo"}, map[string]any{"templates": []any{"CLIENT"}, "primary": "client"}, true},
		{"project.templates.set", map[string]any{"project": "demo"}, map[string]any{"templates": []any{"server", "client", "server"}, "primary": "client"}, true},
		{"project.templates.set", map[string]any{"project": "demo"}, map[string]any{"templates": []any{"client"}, "primary": "server"}, false},
		{"project.templates.show", map[string]any{"project": "client@demo"}, map[string]any{}, false},
		{"experiments.list", map[string]any{"project": "server@demo"}, map[string]any{}, false},
		{"personalizations.show", map[string]any{"project": "demo", "personalization_id": "id"}, map[string]any{}, true},
		{"draft.publish", map[string]any{}, map[string]any{}, false},
		{"draft.publish", map[string]any{"project": []any{"demo"}}, map[string]any{}, true},
		{"draft.publish", map[string]any{}, map[string]any{"all": true}, true},
		{"draft.publish", map[string]any{"project": []any{"demo"}}, map[string]any{"all": true}, false},
		{"draft.change-note", map[string]any{"project": "demo", "text": "release"}, map[string]any{"clear": true}, false},
		{"draft.change-note", map[string]any{"project": "demo", "text": ""}, map[string]any{}, false},
		{"draft.change-note", map[string]any{"project": "demo", "text": "line one\nline two"}, map[string]any{}, false},
		{"project.import", map[string]any{"project": "demo"}, map[string]any{"from": "config.json", "merge": false, "merge-resolve": "current", "yes": true}, false},
		{"project.import", map[string]any{"project": "demo"}, map[string]any{"from": "config.json", "merge": true, "merge-resolve": "current", "yes": true}, true},
		{"project.import", map[string]any{"project": "demo"}, map[string]any{"from": "config.json", "group": []any{"   "}, "override": true, "yes": true}, false},
		{"versions.list", map[string]any{"project": "demo"}, map[string]any{"before": ""}, false},
		{"config.show", map[string]any{"key": "hooks.post_publish"}, map[string]any{}, true},
		{"config.show", map[string]any{"key": "hooks.unknown"}, map[string]any{}, false},
		{"config.reset", map[string]any{"key": "keys.global.quit"}, map[string]any{}, true},
		{"config.reset", map[string]any{"key": "hooks"}, map[string]any{}, false},
		{"config.edit", map[string]any{}, map[string]any{"editor": "ignored", "full": true, "scope": "invalid"}, true},
		{"config.edit", map[string]any{}, map[string]any{"editor": "   "}, false},
		{"update", map[string]any{"parameter": "flag"}, map[string]any{"filter": []any{"=other"}}, false},
		{"delete", map[string]any{"parameter": "flag"}, map[string]any{"filter": []any{"=other"}}, false},
		{"update", map[string]any{"parameter": "   "}, map[string]any{}, false},
		{"duplicate", map[string]any{"source": "source", "target": "   "}, map[string]any{}, false},
		{"duplicate", map[string]any{"source": strings.Repeat("s", 257), "target": "target"}, map[string]any{}, false},
	}
	for _, test := range tests {
		validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:"+test.id+":input", map[string]any{
			"arguments": test.arguments,
			"options":   test.options,
			"stdin":     nil,
		}, test.valid)
	}
}

func TestInvocationSchemasRejectMachineOnlyAndTypedValueContradictions(t *testing.T) {
	versionsDiff := "urn:fbrcm:schema:cli:" + contract.Version + ":command:versions.diff:input"
	validateContractValue(t, versionsDiff, map[string]any{
		"arguments": map[string]any{"project": "demo", "from": "1", "to": "2"},
		"options":   map[string]any{"side-by-side": true},
		"stdin":     nil,
	}, false)

	updateID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:update:input"
	update := func(options map[string]any) map[string]any {
		return map[string]any{"arguments": map[string]any{}, "options": options, "stdin": nil}
	}
	validateContractValue(t, updateID, update(map[string]any{"type": "boolean"}), false)
	validateContractValue(t, updateID, update(map[string]any{"type": "boolean", "use-in-app-default": false}), false)
	validateContractValue(t, updateID, update(map[string]any{"type": "boolean", "value": "maybe"}), false)
	validateContractValue(t, updateID, update(map[string]any{"type": "boolean", "value": "false"}), true)
	validateContractValue(t, updateID, update(map[string]any{"type": "number", "value": "not-a-number"}), false)
	validateContractValue(t, updateID, update(map[string]any{"type": "number", "value": "-1.5e+3"}), true)
	for _, value := range []string{"1_0", "nan", "NAN", "0x1_ffp2", "0x_1p2", "infinity"} {
		validateContractValue(t, updateID, update(map[string]any{"type": "number", "value": value}), true)
	}
	for _, value := range []string{"+NaN", "1__0", "1_", "0x1"} {
		validateContractValue(t, updateID, update(map[string]any{"type": "number", "value": value}), false)
	}
	validateContractValue(t, updateID, update(map[string]any{"condition": "   ", "type": "string", "value": "on"}), false)
	validateContractValue(t, updateID, update(map[string]any{"remove-conditional-value": []any{"   "}}), false)
	validateContractValue(t, updateID, update(map[string]any{"name": strings.Repeat("x", 257)}), false)
	raw, err := schemas.ReadByID(updateID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte(`"operator": "parse_json"`)) || !bytes.Contains(raw, []byte(`"specification": "RFC 8259"`)) {
		t.Fatal("update value schema does not publish its JSON parser invariant")
	}
	var updateSchema map[string]any
	if err := json.Unmarshal(raw, &updateSchema); err != nil {
		t.Fatal(err)
	}
	parameterSchema := updateSchema["properties"].(map[string]any)["arguments"].(map[string]any)["properties"].(map[string]any)["parameter"].(map[string]any)
	if parameterSchema["x-fbrcm-normalization"] != nil {
		t.Fatal("update parameter schema falsely publishes argv whitespace trimming")
	}
	duplicateRaw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:duplicate:input")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(duplicateRaw, []byte(`"operator": "fields_differ"`)) || !bytes.Contains(duplicateRaw, []byte(`"comparison": "exact_codepoint"`)) {
		t.Fatal("duplicate invocation schema does not publish its exact source/target inequality")
	}
	versionsListRaw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:versions.list:input")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(versionsListRaw, []byte(`"operator": "parse_time"`)) || !bytes.Contains(versionsListRaw, []byte(`"specification": "Go time.RFC3339"`)) {
		t.Fatal("version range schema does not publish the runtime timestamp parser")
	}
	conditionsAddID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:conditions.add:input"
	validateContractValue(t, conditionsAddID, map[string]any{
		"arguments": map[string]any{"project": "demo", "name": strings.Repeat("x", 101)},
		"options":   map[string]any{"expression": "true"}, "stdin": nil,
	}, false)
	validateContractValue(t, conditionsAddID, map[string]any{
		"arguments": map[string]any{"project": "demo", "name": "beta"},
		"options":   map[string]any{"expression": "   "}, "stdin": nil,
	}, false)
}

func TestConfigSetInvocationSchemaPublishesClosedKeyGrammar(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:config.set:input"
	input := func(key string, values []any, options map[string]any) map[string]any {
		return map[string]any{"arguments": map[string]any{"key": key, "value": values}, "options": options, "stdin": nil}
	}
	validateContractValue(t, id, input("powerline_glyphs", []any{"true"}, map[string]any{}), true)
	validateContractValue(t, id, input("powerline_glyphs", []any{"yes"}, map[string]any{}), false)
	validateContractValue(t, id, input("unknown", []any{"value"}, map[string]any{}), false)
	validateContractValue(t, id, input("keys.global.quit", []any{"q"}, map[string]any{}), true)
	validateContractValue(t, id, input("keys.global.quit", []any{" "}, map[string]any{}), false)
	validateContractValue(t, id, input("keys.global.quit", []any{""}, map[string]any{}), false)
	validateContractValue(t, id, input("keys.global.quit", []any{"ctrl+"}, map[string]any{}), false)
	validateContractValue(t, id, input("keys.global.quit", []any{"ctrl+r"}, map[string]any{}), true)
	validateContractValue(t, id, input("keys.global.not-an-action", []any{"q"}, map[string]any{}), false)
	validateContractValue(t, id, input("projects.aliases.demo", []any{"physical-project"}, map[string]any{}), false)
	validateContractValue(t, id, input("projects.aliases.demo", []any{"physical-project"}, map[string]any{"scope": "local"}), true)
	raw, err := schemas.ReadByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if got := bytes.Count(raw, []byte(`"operator": "trim_unicode_whitespace"`)); got != 3 {
		t.Fatalf("config.set normalization rule count = %d, want nested key, alias, and case-insensitive scope normalization", got)
	}
}

func TestIgnoredProfileOptionsAreExplicitInInvocationSchemas(t *testing.T) {
	for _, id := range []string{"capabilities", "config.show", "help", "hooks.status", "projects.aliases.list", "schema.list"} {
		raw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:" + id + ":input")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(raw, []byte(`"x-fbrcm-effective": false`)) {
			t.Errorf("%s does not mark --profile ineffective", id)
		}
	}
	raw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:root:input")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(`"x-fbrcm-effective": false`)) {
		t.Fatal("root schema marks an effective option as ignored")
	}
	if got := bytes.Count(raw, []byte(`"x-fbrcm-effective-when"`)); got != 2 {
		t.Fatalf("root schema conditional option effectiveness count = %d, want profile and no-local-config", got)
	}
}

func TestMachineIgnoredCommandOptionsAreExplicitInInvocationSchemas(t *testing.T) {
	for id, optionNames := range map[string][]string{
		"auth.login":  {"noopen"},
		"config.edit": {"editor", "full", "scope"},
	} {
		raw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:" + id + ":input")
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		properties := document["properties"].(map[string]any)["options"].(map[string]any)["properties"].(map[string]any)
		for _, name := range optionNames {
			option := properties[name].(map[string]any)
			if option["x-fbrcm-effective"] != false {
				t.Errorf("%s --%s does not publish x-fbrcm-effective false: %#v", id, name, option)
			}
		}
	}
}

func TestConfigKeySchemasPublishConditionalNestedKeyNormalization(t *testing.T) {
	for id, prefixes := range map[string][]string{
		"config.show":  {"keys.", "projects.aliases.", "hooks."},
		"config.reset": {"keys.", "projects.aliases."},
	} {
		raw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:" + id + ":input")
		if err != nil {
			t.Fatal(err)
		}
		for _, marker := range append([]string{`"operator": "trim_unicode_whitespace_if_prefix"`}, prefixes...) {
			if !bytes.Contains(raw, []byte(marker)) {
				t.Errorf("%s invocation schema omits conditional key normalization marker %q", id, marker)
			}
		}
	}
}

func TestInvocationSchemasPublishProfileAndAuthIdentifierGrammar(t *testing.T) {
	cachePathID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:cache.path:input"
	invocation := func(profile string) map[string]any {
		return map[string]any{"arguments": map[string]any{}, "options": map[string]any{"profile": profile}, "stdin": nil}
	}
	validateContractValue(t, cachePathID, invocation("automation"), true)
	validateContractValue(t, cachePathID, invocation("../automation"), false)

	authBindID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:auth.bind:input"
	validateContractValue(t, authBindID, map[string]any{
		"arguments": map[string]any{}, "options": map[string]any{"auth": "main"}, "stdin": nil,
	}, true)
	validateContractValue(t, authBindID, map[string]any{
		"arguments": map[string]any{}, "options": map[string]any{"auth": "../main"}, "stdin": nil,
	}, false)

	rootID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:root:input"
	validateContractValue(t, rootID, map[string]any{
		"arguments": map[string]any{}, "options": map[string]any{"profile": "../ignored", "version": true}, "stdin": nil,
	}, true)
	validateContractValue(t, rootID, map[string]any{
		"arguments": map[string]any{}, "options": map[string]any{"profile": "../effective", "version": false}, "stdin": nil,
	}, false)
}

func TestInvocationSchemasPublishQueryAndManagedFeatureSemantics(t *testing.T) {
	getID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:get:input"
	getRaw, err := schemas.ReadByID(getID)
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		`"operator": "parameter_search"`, `"operator": "canonicalize_target_selector"`,
		`"operator": "selection_composition"`,
		`"repeated_source_combination": "or"`, `"across_source_combination": "and"`,
		`"selection": "all_configured_projects_enabled_templates"`,
	} {
		if !bytes.Contains(getRaw, []byte(marker)) {
			t.Errorf("get invocation schema omits %s", marker)
		}
	}
	for _, marker := range []string{`x-fbrcm-accepts-directory`, `"const": "directory"`} {
		if bytes.Contains(getRaw, []byte(marker)) {
			t.Errorf("get invocation schema publishes experimental stdin transport marker %s", marker)
		}
	}
	semanticRaw, err := schemas.ReadByID(contract.SemanticSchemaID())
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		`"operator": "mode_prefixed_query"`, `"unqualified_target_selection": "all_configured_enabled_templates"`,
		`"unqualified_target_selection": "configured_primary_template"`, `"explicit_target_selection": "single_named_template"`,
		`"client_target_canonicalization": "unqualified_project_id"`, `"repository_aliases"`,
	} {
		if !bytes.Contains(semanticRaw, []byte(marker)) {
			t.Errorf("semantic schema omits selection marker %s", marker)
		}
	}
	for id, markers := range map[string][]string{
		"conditions.list": {`"condition_name"`},
		"draft.list":      {`"$ref": "#/$defs/draft_filter"`},
		"project.export":  {`"$ref": "#/$defs/target_positional_selector"`},
		"projects.list":   {`"project_id"`, `"display_name"`, `"repository_aliases"`},
	} {
		raw, readErr := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:" + id + ":input")
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, marker := range markers {
			if !bytes.Contains(raw, []byte(marker)) {
				t.Errorf("%s invocation schema omits selection marker %s", id, marker)
			}
		}
	}

	experimentID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:experiments.show:input"
	experiment := func(id string) map[string]any {
		return map[string]any{"arguments": map[string]any{"project": "demo", "experiment_id": id}, "options": map[string]any{}, "stdin": nil}
	}
	validateContractValue(t, experimentID, experiment("experiment-1"), true)
	validateContractValue(t, experimentID, experiment("projects/123/namespaces/firebase/experiments/experiment-1"), true)
	validateContractValue(t, experimentID, experiment("wrong/resource"), false)

	personalizationRaw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:personalizations.show:input")
	if err != nil {
		t.Fatal(err)
	}
	var personalizationSchema map[string]any
	if err := json.Unmarshal(personalizationRaw, &personalizationSchema); err != nil {
		t.Fatal(err)
	}
	personalizationArgument := personalizationSchema["properties"].(map[string]any)["arguments"].(map[string]any)["properties"].(map[string]any)["personalization_id"].(map[string]any)
	if personalizationArgument["x-fbrcm-normalization"] != nil {
		t.Fatal("personalization id falsely publishes trim normalization")
	}
}

func TestExpressionSchemaListsRegisteredPredicateFunctions(t *testing.T) {
	raw, err := schemas.ReadByID(contract.SemanticSchemaID())
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	definitions := document["$defs"].(map[string]any)
	language := definitions["expression"].(map[string]any)["x-fbrcm-language"].(map[string]any)
	functions := language["functions"].([]any)
	for _, want := range []string{"is_boolean", "is_number"} {
		if !slices.ContainsFunc(functions, func(value any) bool { return value == want }) {
			t.Errorf("expression functions omit %q: %#v", want, functions)
		}
	}
}

func TestInvocationSchemasAcceptCaseInsensitiveRuntimeEnums(t *testing.T) {
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:add:input", map[string]any{
		"arguments": map[string]any{"parameter": "flag"},
		"options":   map[string]any{"value": "true", "type": "BoOl"},
		"stdin":     nil,
	}, true)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:project.defaults:input", map[string]any{
		"arguments": map[string]any{"project": "demo"},
		"options":   map[string]any{"format": "PlIsT"},
		"stdin":     nil,
	}, true)
}

func TestInvocationSchemasMatchBooleanPresenceAndPositiveTimeoutRuntime(t *testing.T) {
	addID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:add:input"
	base := func(options map[string]any) map[string]any {
		return map[string]any{"arguments": map[string]any{"parameter": "flag"}, "options": options, "stdin": nil}
	}
	validateContractValue(t, addID, base(map[string]any{"use-in-app-default": false}), false)
	validateContractValue(t, addID, base(map[string]any{"use-in-app-default": true, "type": "string"}), true)

	capabilitiesID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:capabilities:input"
	capabilitiesInput := func(timeout string) map[string]any {
		return map[string]any{"arguments": map[string]any{}, "options": map[string]any{"timeout": timeout}, "stdin": nil}
	}
	for _, zero := range []string{"0", "+0", "0s", "+0s", ".0s", "0.s", "0h0m", "0.0ms"} {
		validateContractValue(t, capabilitiesID, capabilitiesInput(zero), false)
	}
	for _, positive := range []string{"1ms", "+1ms", ".5s", "1.s", "1μs"} {
		validateContractValue(t, capabilitiesID, capabilitiesInput(positive), true)
	}
	for _, subNanosecond := range []string{".1ns", "0.1ns", "+.9ns"} {
		validateContractValue(t, capabilitiesID, capabilitiesInput(subNanosecond), false)
	}
	raw, err := schemas.ReadByID(capabilitiesID)
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{`"operator": "parse_duration"`, `"parser": "time.ParseDuration"`, `"require_positive": true`} {
		if !bytes.Contains(raw, []byte(marker)) {
			t.Errorf("timeout invocation schema omits %s", marker)
		}
	}
}

func TestInvocationSchemasUseCommandSpecificScopeAndPrioritySemantics(t *testing.T) {
	base := func(options map[string]any) map[string]any {
		return map[string]any{"arguments": map[string]any{}, "options": options, "stdin": nil}
	}
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:config.validate:input", base(map[string]any{"scope": "all"}), true)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:config.path:input", base(map[string]any{"scope": "effective"}), false)

	conditionInput := map[string]any{
		"arguments": map[string]any{"project": "demo", "name": "beta"},
		"options":   map[string]any{"expression": "true", "priority": 0},
		"stdin":     nil,
	}
	conditionID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:conditions.add:input"
	validateContractValue(t, conditionID, conditionInput, true)
	conditionInput["options"].(map[string]any)["priority"] = 2147483647
	validateContractValue(t, conditionID, conditionInput, true)
	conditionInput["options"].(map[string]any)["priority"] = 2147483648
	validateContractValue(t, conditionID, conditionInput, false)
	conditionInput["options"].(map[string]any)["priority"] = -1
	validateContractValue(t, conditionID, conditionInput, false)

	versionsListID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:versions.list:input"
	versionsListInput := func(limit int64) map[string]any {
		return map[string]any{"arguments": map[string]any{"project": "demo"}, "options": map[string]any{"limit": limit}, "stdin": nil}
	}
	validateContractValue(t, versionsListID, versionsListInput(2147483647), true)
	validateContractValue(t, versionsListID, versionsListInput(2147483648), false)

	moveID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:conditions.move:input"
	moveInput := func(priority string) map[string]any {
		return map[string]any{"arguments": map[string]any{"project": "demo", "condition": "beta", "priority": priority}, "options": map[string]any{}, "stdin": nil}
	}
	validateContractValue(t, moveID, moveInput("1"), true)
	validateContractValue(t, moveID, moveInput("+001"), true)
	for _, priority := range []string{"0", "-1", "first"} {
		validateContractValue(t, moveID, moveInput(priority), false)
	}

	colorID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:conditions.add:input"
	colorInput := func(color string) map[string]any {
		return map[string]any{"arguments": map[string]any{"project": "demo", "name": "beta"}, "options": map[string]any{"expression": "true", "color": color}, "stdin": nil}
	}
	validateContractValue(t, colorID, colorInput("blue"), true)
	validateContractValue(t, colorID, colorInput("red"), false)
}

func TestHelpAndRootInvocationSchemasDescribeActualArgumentsAndVersion(t *testing.T) {
	helpID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:help:input"
	validateContractValue(t, helpID, map[string]any{"arguments": map[string]any{"command": []any{"projects", "aliases", "list"}}, "options": map[string]any{}, "stdin": nil}, true)
	validateContractValue(t, helpID, map[string]any{"arguments": map[string]any{"command": "projects"}, "options": map[string]any{}, "stdin": nil}, false)

	rootID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:root:input"
	validateContractValue(t, rootID, map[string]any{"arguments": map[string]any{}, "options": map[string]any{"version": true}, "stdin": nil}, true)
}

func TestRepeatableInvocationOptionsRejectUnrepresentableEmptyArrays(t *testing.T) {
	root := NewRootForContract("schema")
	for _, capability := range contract.DetailedCapabilities(root) {
		raw, err := schemas.ReadByID(capability.InvocationSchema)
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		properties := document["properties"].(map[string]any)
		options := properties["options"].(map[string]any)["properties"].(map[string]any)
		for _, flag := range capability.Flags {
			if !flag.Repeatable {
				continue
			}
			name := strings.TrimPrefix(flag.Name, "--")
			option, ok := options[name].(map[string]any)
			if !ok {
				t.Errorf("%s repeatable option %s has no input schema", capability.ID, flag.Name)
				continue
			}
			if got := option["minItems"]; got != float64(1) {
				t.Errorf("%s %s minItems = %v, want 1", capability.ID, flag.Name, got)
			}
		}
	}
}

func TestInvocationSchemasPublishCommandLocalSelectionSemantics(t *testing.T) {
	getID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:get:input"
	getInput := func(arguments, options map[string]any) map[string]any {
		return map[string]any{"arguments": arguments, "options": options, "stdin": nil}
	}
	validateContractValue(t, getID, getInput(map[string]any{"parameter": "flag"}, map[string]any{}), true)
	validateContractValue(t, getID, getInput(map[string]any{}, map[string]any{"filter": []any{"=flag"}}), true)
	validateContractValue(t, getID, getInput(map[string]any{"parameter": "flag"}, map[string]any{"filter": []any{"=flag"}}), false)
	assertSchemaMarkers(t, getID,
		`"operator": "parameter_argument_resolution"`,
		`"arguments.parameter"`,
	)

	for _, commandID := range []string{"draft.show", "draft.change-note", "draft.diff", "draft.publish", "draft.discard"} {
		id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:" + commandID + ":input"
		assertSchemaMarkers(t, id, `"$ref": "#/$defs/draft_selector"`)
	}
	assertSchemaMarkers(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:draft.list:input",
		`"$ref": "#/$defs/draft_filter"`,
	)
	assertSchemaMarkers(t, contract.SemanticSchemaID(),
		`"operator": "draft_resolution"`,
		`"existing_drafts_in_configured_enabled_templates_or_client_fallback"`,
		`"query_normalization": "preserve_argv"`,
		`"comparison": "exact_case_sensitive"`,
		`"canonicalize_positional_target_selector"`,
	)
	semanticRaw, err := schemas.ReadByID(contract.SemanticSchemaID())
	if err != nil {
		t.Fatal(err)
	}
	for _, obsolete := range [][]byte{
		[]byte(`"project_id_or_display_name_substring"`),
		[]byte(`"mode_match_project_id_or_display_name"`),
		[]byte(`"parameter_argument_exact_filter"`),
	} {
		if bytes.Contains(semanticRaw, obsolete) {
			t.Fatalf("semantic schema retains obsolete search behavior %s", obsolete)
		}
	}
	for _, commandID := range []string{"draft.publish", "draft.discard"} {
		id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:" + commandID + ":input"
		assertSchemaMarkers(t, id, `"operator": "draft_batch_selection"`, `"canonical_order": "deduplicate_then_sort_target_id"`)
	}

	assertSchemaMarkers(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:projects.diff:input",
		`"operator": "deduplicate_preserve_first"`,
		`"operator": "trim_unicode_whitespace"`,
	)
	for commandID, operator := range map[string]string{
		"conditions.show":       "condition_positional_resolution",
		"duplicate":             "duplicate_source_resolution",
		"groups.edit":           "group_name_resolution",
		"personalizations.show": "personalization_id_resolution",
		"versions.show":         "version_resolution",
	} {
		id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:" + commandID + ":input"
		assertSchemaMarkers(t, id, `"operator": "`+operator+`"`)
	}
	assertSchemaMarkers(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:conditions.add:input",
		`"operator": "condition_priority"`, `"maximum": "resolved_condition_count_plus_one"`,
	)
	assertSchemaMarkers(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:conditions.move:input",
		`"operator": "condition_priority"`, `"maximum": "resolved_condition_count"`,
	)
	assertSchemaMarkers(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:capabilities:input",
		`"operator": "command_path_resolution"`, `"reserved_root_token": "root"`, `"non_executable_result": "command.not_executable"`,
	)
	assertSchemaMarkers(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:help:input",
		`"operator": "help_path_resolution"`,
	)
	assertSchemaMarkers(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:schema.show:input",
		`"operator": "schema_id_resolution"`,
	)
	for commandID, operator := range map[string]string{
		"auth.login":              "auth_id_resolution",
		"profile.delete":          "profile_name_resolution",
		"projects.aliases.remove": "project_alias_resolution",
	} {
		assertSchemaMarkers(t, "urn:fbrcm:schema:cli:"+contract.Version+":command:"+commandID+":input",
			"\"operator\": \""+operator+"\"",
		)
	}
}

func assertSchemaMarkers(t *testing.T, schemaID string, markers ...string) {
	t.Helper()
	raw, err := schemas.ReadByID(schemaID)
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range markers {
		if !bytes.Contains(raw, []byte(marker)) {
			t.Errorf("%s omits %s", schemaID, marker)
		}
	}
}

func TestPublishedStdinSchemaDescribesRemoteConfig(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":stdin:remote_config"
	raw, err := schemas.ReadByID(id)
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	variants, ok := schema["anyOf"].([]any)
	if !ok || len(variants) != 2 {
		t.Fatalf("remote config variants = %#v", schema["anyOf"])
	}
	template, ok := variants[0].(map[string]any)
	if !ok {
		t.Fatalf("remote config template = %#v", variants[0])
	}
	properties, ok := template["properties"].(map[string]any)
	if !ok {
		t.Fatalf("remote config properties = %#v", template["properties"])
	}
	for _, name := range []string{"parameters", "parameterGroups", "conditions", "version"} {
		if _, ok := properties[name]; !ok {
			t.Errorf("remote config schema is missing %q", name)
		}
	}
	cache, ok := variants[1].(map[string]any)
	cacheProperties, propertiesOK := cache["properties"].(map[string]any)
	if !ok || !propertiesOK || cacheProperties["remote_config"] == nil {
		t.Fatalf("parameters-cache schema = %#v", variants[1])
	}
	validTemplate := map[string]any{"parameters": map[string]any{"flag": map[string]any{"defaultValue": map[string]any{"value": "on"}}}}
	validateContractValue(t, id, validTemplate, true)
	validateContractValue(t, id, map[string]any{
		"conditions": []any{map[string]any{"name": "beta", "expression": "true", "tagColor": "deep_orange"}},
		"parameters": map[string]any{"flag": map[string]any{"valueType": "BOOLEAN", "defaultValue": map[string]any{"value": "true"}}},
	}, true)
	for _, locallyAcceptedTemplate := range []any{
		map[string]any{"conditions": []any{map[string]any{}}},
		map[string]any{"conditions": []any{map[string]any{"name": "beta", "expression": "true", "tagColor": "RED"}}},
		map[string]any{"parameters": map[string]any{"flag": map[string]any{"valueType": "FLOAT", "defaultValue": map[string]any{"value": "1"}}}},
		map[string]any{"conditions": []any{map[string]any{"name": "same", "expression": "not valid"}, map[string]any{"name": "same", "expression": "also invalid"}}},
		map[string]any{"parameters": map[string]any{}, "futureRootField": map[string]any{"enabled": true}},
	} {
		validateContractValue(t, id, locallyAcceptedTemplate, true)
	}
	for _, typedNull := range []any{
		nil,
		map[string]any{"version": nil},
		map[string]any{"conditions": []any{nil}},
		map[string]any{"parameters": map[string]any{"flag": nil}},
		map[string]any{"parameterGroups": map[string]any{"checkout": nil}},
	} {
		validateContractValue(t, id, typedNull, false)
	}
	validateContractValue(t, id, map[string]any{"conditions": []any{map[string]any{"name": "beta", "expression": "true", "unsupported": true}}}, false)
	if bytes.Contains(raw, []byte(`"operator": "remote_validate"`)) || bytes.Contains(raw, []byte(`"operator": "unique_by"`)) {
		t.Fatalf("stdin schema contains Firebase-only semantic validation: %s", raw)
	}
	for _, invalidValue := range []any{
		map[string]any{},
		map[string]any{"value": "on", "useInAppDefault": true},
		map[string]any{"useInAppDefault": false},
	} {
		validateContractValue(t, id, map[string]any{"parameters": map[string]any{"flag": map[string]any{"defaultValue": invalidValue}}}, false)
	}
}

func TestProjectImportStdinSchemaPublishesLocalValidation(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":stdin:remote_config_import"
	raw, err := schemas.ReadByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte(`"operator": "local_validate"`)) || !bytes.Contains(raw, []byte(`"validator": "firebase.NormalizeRemoteConfigForUpdate"`)) {
		t.Fatal("project import stdin schema does not publish its local normalization validation")
	}
	validateContractValue(t, id, map[string]any{"parameters": map[string]any{"flag": map[string]any{"defaultValue": map[string]any{"value": "on"}}}}, true)
}

func TestInputSchemasPublishFileBeforeStdinSelection(t *testing.T) {
	for _, command := range []string{"auth.add.oauth", "auth.add.service-account", "project.import"} {
		raw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":command:" + command + ":input")
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		rules, ok := document["x-fbrcm-input-selection"].([]any)
		if !ok || len(rules) != 1 {
			t.Fatalf("%s input selection = %#v", command, document["x-fbrcm-input-selection"])
		}
		rule := rules[0].(map[string]any)
		if rule["operator"] != "first_available" || !slices.Equal(contractStrings(rule["sources"]), []string{"options.from", "stdin.document"}) || rule["on_missing"] != "interaction.required" || rule["later_sources"] != "ignored_without_consumption" {
			t.Fatalf("%s input selection = %#v", command, rule)
		}
	}
}

func TestPublishedCredentialStdinSchemaDescribesAcceptedCredentialKinds(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":stdin:credentials"
	validateContractValue(t, id, map[string]any{"installed": map[string]any{
		"client_id": "client", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token", "redirect_uris": []any{"http://localhost"},
	}}, true)
	validateContractValue(t, id, map[string]any{
		"type": "service_account", "project_id": "demo", "private_key": "key", "client_email": "svc@example.com", "token_uri": "https://accounts.example/token",
	}, true)
	validateContractValue(t, id, map[string]any{}, false)
	validateContractValue(t, id, map[string]any{"installed": map[string]any{
		"redirect_uris": []any{"http://localhost"},
	}}, false)
	validateContractValue(t, id, map[string]any{"type": "service_account"}, false)

	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", map[string]any{"installed": map[string]any{
		"client_id": "client", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token", "redirect_uris": []any{"http://localhost"},
	}}, true)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", map[string]any{"installed": map[string]any{
		"client_id": "client", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token",
	}}, false)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", map[string]any{
		"installed": map[string]any{"client_id": "client", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token", "redirect_uris": []any{"http://localhost"}},
		"web":       map[string]any{"client_id": "client", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token", "redirect_uris": []any{"http://localhost"}},
	}, true)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", map[string]any{
		"installed": map[string]any{"client_id": "client", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token", "redirect_uris": []any{"http://localhost"}},
		"web":       nil,
	}, true)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", map[string]any{
		"installed": nil,
		"web":       map[string]any{"client_id": "client", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token", "redirect_uris": []any{"http://localhost"}},
	}, true)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", map[string]any{
		"installed": map[string]any{"client_id": "client", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token"},
		"web":       map[string]any{"redirect_uris": []any{"http://localhost"}},
	}, true)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", map[string]any{"installed": map[string]any{
		"client_id": "   ", "client_secret": "secret", "auth_uri": "https://accounts.example/authorize", "token_uri": "https://accounts.example/token", "redirect_uris": []any{"http://localhost"},
	}}, false)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", map[string]any{
		"type": "service_account", "project_id": "demo", "private_key": "key", "client_email": "svc@example.com", "token_uri": "https://accounts.example/token",
	}, false)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:service_account_credentials", map[string]any{
		"type": "service_account", "project_id": "demo", "private_key": "key", "client_email": "svc@example.com", "token_uri": "https://accounts.example/token",
	}, true)
	validateContractValue(t, "urn:fbrcm:schema:cli:"+contract.Version+":stdin:service_account_credentials", map[string]any{
		"type": "service_account", "project_id": "   ", "private_key": "key", "client_email": "svc@example.com", "token_uri": "https://accounts.example/token",
	}, false)
	for _, schemaID := range []string{
		"urn:fbrcm:schema:cli:" + contract.Version + ":stdin:oauth_credentials",
		"urn:fbrcm:schema:cli:" + contract.Version + ":stdin:service_account_credentials",
	} {
		raw, err := schemas.ReadByID(schemaID)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(raw, []byte(`"operator": "parse_uri"`)) {
			t.Errorf("%s omits its runtime URI parser", schemaID)
		}
		if !bytes.Contains(raw, []byte(`"require_absolute": true`)) {
			t.Errorf("%s omits the runtime absolute-URI requirement", schemaID)
		}
	}
	serviceAccountRaw, err := schemas.ReadByID("urn:fbrcm:schema:cli:" + contract.Version + ":stdin:service_account_credentials")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(serviceAccountRaw, []byte(`"operator": "parse_email"`)) {
		t.Fatal("service-account credential schema omits its runtime email parser")
	}

	root := NewRootForContract("test")
	oauth, err := contract.FindCapability(root, []string{"auth", "add", "oauth"})
	if err != nil || oauth.StdinSchema == nil || !strings.HasSuffix(*oauth.StdinSchema, ":stdin:oauth_credentials") {
		t.Fatalf("OAuth stdin schema = %#v, %v", oauth.StdinSchema, err)
	}
	serviceAccount, err := contract.FindCapability(root, []string{"auth", "add", "service-account"})
	if err != nil || serviceAccount.StdinSchema == nil || !strings.HasSuffix(*serviceAccount.StdinSchema, ":stdin:service_account_credentials") {
		t.Fatalf("service-account stdin schema = %#v, %v", serviceAccount.StdinSchema, err)
	}
}

func TestArtifactSchemaRejectsContradictoryRepresentations(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:project.export:response"
	root := NewRootForContract("test")
	cmd, _, err := root.Find([]string{"project", "export"})
	if err != nil {
		t.Fatal(err)
	}
	target := "demo"
	artifact := contract.NewArtifact(&target, "application/json", []byte(`{"parameters":{"flag":{"defaultValue":{"value":"true"}}}}`), nil, false)
	payload, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	envelope := contract.BuildEnvelope(cmd, "test", payload, nil)
	if envelope.Outcome != "success" {
		t.Fatalf("artifact data was mistaken for a counted response DTO: %#v", envelope)
	}
	validateContractDocument(t, id, marshalEnvelope(t, envelope))

	document := structToContractValue(t, envelope).(map[string]any)
	data := document["data"].(map[string]any)
	data["encoding"] = "utf-8"
	validateContractValue(t, id, document, false)

	document = structToContractValue(t, envelope).(map[string]any)
	data = document["data"].(map[string]any)
	data["text_content"] = "contradictory inline representation"
	validateContractValue(t, id, document, false)
}

func TestCommandResponseSchemasConstrainReachableOutcomesAndWarnings(t *testing.T) {
	root := NewRootForContract("test")
	partialCommands := map[string]bool{
		"add": true, "conditions.add": true, "conditions.delete": true, "conditions.edit": true,
		"conditions.move": true, "conditions.rename": true, "delete": true, "draft.publish": true,
		"duplicate": true, "groups.add": true, "groups.delete": true, "groups.edit": true,
		"groups.rename": true, "project.import": true, "projects.promote": true, "update": true,
		"versions.restore": true, "versions.rollback": true,
	}
	postPublication := []string{"publication.cache_stale", "publication.post_publish_hook_failed"}
	warningsByCommand := map[string][]string{
		"get":               {"cache.stale"},
		"conditions.add":    postPublication,
		"conditions.delete": postPublication,
		"conditions.edit":   postPublication,
		"conditions.move":   postPublication,
		"conditions.rename": postPublication,
		"project.import":    postPublication,
		"projects.promote":  postPublication,
		"versions.restore":  postPublication,
		"versions.rollback": postPublication,
		"draft.publish":     {"publication.cache_stale", "publication.draft_cleanup_failed", "publication.non_atomic", "publication.post_publish_hook_failed"},
	}
	for _, id := range []string{"add", "delete", "duplicate", "update", "groups.add", "groups.delete", "groups.edit", "groups.rename"} {
		warningsByCommand[id] = []string{"publication.non_atomic", "publication.cache_stale", "publication.post_publish_hook_failed"}
	}
	for _, capability := range contract.DetailedCapabilities(root) {
		t.Run(capability.ID, func(t *testing.T) {
			raw, err := schemas.ReadByID(capability.ResponseSchema)
			if err != nil {
				t.Fatal(err)
			}
			var document map[string]any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			commandSchema := document["allOf"].([]any)[1].(map[string]any)
			constraints := commandSchema["allOf"].([]any)
			var gotOutcomes, gotWarnings []string
			warningsForbidden := false
			for _, rawConstraint := range constraints {
				constraint := rawConstraint.(map[string]any)
				properties, _ := constraint["properties"].(map[string]any)
				if outcome, ok := properties["outcome"].(map[string]any); ok {
					gotOutcomes = contractStrings(outcome["enum"])
				}
				if warningList, ok := properties["warnings"].(map[string]any); ok {
					if maximum, ok := warningList["maxItems"].(float64); ok && maximum == 0 {
						warningsForbidden = true
					}
					items, _ := warningList["items"].(map[string]any)
					itemProperties, _ := items["properties"].(map[string]any)
					code, _ := itemProperties["code"].(map[string]any)
					gotWarnings = contractStrings(code["enum"])
				}
			}
			wantOutcomes := []string{"success", "failure"}
			if capability.ID == "config.edit" {
				wantOutcomes = []string{"failure"}
			} else if partialCommands[capability.ID] {
				wantOutcomes = []string{"success", "partial_success", "failure"}
			}
			if !slices.Equal(gotOutcomes, wantOutcomes) {
				t.Fatalf("outcomes = %v, want %v", gotOutcomes, wantOutcomes)
			}
			wantWarnings := warningsByCommand[capability.ID]
			if len(wantWarnings) == 0 {
				if !warningsForbidden {
					t.Fatal("warnings are not forbidden")
				}
			} else if !slices.Equal(gotWarnings, wantWarnings) {
				t.Fatalf("warnings = %v, want %v", gotWarnings, wantWarnings)
			}
		})
	}
}

func contractStrings(value any) []string {
	values, _ := value.([]any)
	result := make([]string, 0, len(values))
	for _, item := range values {
		result = append(result, item.(string))
	}
	return result
}

func TestArtifactSchemasConstrainCommandReachableEncodings(t *testing.T) {
	jsonArtifact := artifactContractData("json")
	utf8Artifact := artifactContractData("utf-8")
	jsonUTF8Artifact := maps.Clone(utf8Artifact)
	jsonUTF8Artifact["media_type"] = "application/json"
	base64Artifact := artifactContractData("base64")
	destinationArtifact := artifactContractData("none")
	destinationArtifact["destination"] = "/tmp/output"

	for _, test := range []struct {
		command string
		data    map[string]any
		valid   bool
	}{
		{command: "add", data: jsonArtifact, valid: true},
		{command: "add", data: utf8Artifact, valid: false},
		{command: "add", data: destinationArtifact, valid: false},
		{command: "update", data: jsonArtifact, valid: true},
		{command: "update", data: base64Artifact, valid: false},
		{command: "delete", data: jsonArtifact, valid: true},
		{command: "delete", data: destinationArtifact, valid: false},
		{command: "project.export", data: jsonArtifact, valid: true},
		{command: "project.export", data: destinationArtifact, valid: true},
		{command: "project.export", data: base64Artifact, valid: false},
		{command: "versions.export", data: jsonArtifact, valid: true},
		{command: "versions.export", data: destinationArtifact, valid: true},
		{command: "versions.export", data: utf8Artifact, valid: false},
		{command: "project.defaults", data: jsonArtifact, valid: true},
		{command: "project.defaults", data: utf8Artifact, valid: true},
		{command: "project.defaults", data: base64Artifact, valid: true},
		{command: "project.defaults", data: destinationArtifact, valid: true},
		{command: "draft.show", data: jsonArtifact, valid: true},
		{command: "draft.show", data: jsonUTF8Artifact, valid: true},
		{command: "draft.show", data: base64Artifact, valid: true},
		{command: "draft.show", data: destinationArtifact, valid: true},
	} {
		t.Run(test.command+"/"+test.data["encoding"].(string), func(t *testing.T) {
			validateContractValue(t, contract.SchemaID(test.command), artifactEnvelopeValue(test.command, test.data), test.valid)
		})
	}

	for _, command := range []string{"add", "update", "delete", "project.export", "versions.export", "project.defaults", "draft.show"} {
		nullTarget := maps.Clone(jsonArtifact)
		nullTarget["target"] = nil
		validateContractValue(t, contract.SchemaID(command), artifactEnvelopeValue(command, nullTarget), false)
	}
	scalarExport := maps.Clone(jsonArtifact)
	scalarExport["json_content"] = "valid JSON is not decoded on the project export path"
	validateContractValue(t, contract.SchemaID("project.export"), artifactEnvelopeValue("project.export", scalarExport), true)
	scalarDefaults := maps.Clone(jsonArtifact)
	scalarDefaults["json_content"] = true
	validateContractValue(t, contract.SchemaID("project.defaults"), artifactEnvelopeValue("project.defaults", scalarDefaults), true)
	scalarDraft := maps.Clone(jsonArtifact)
	scalarDraft["json_content"] = 1
	validateContractValue(t, contract.SchemaID("draft.show"), artifactEnvelopeValue("draft.show", scalarDraft), true)
	validateContractValue(t, contract.SchemaID("versions.export"), artifactEnvelopeValue("versions.export", scalarExport), false)
	validateContractValue(t, contract.SchemaID("add"), artifactEnvelopeValue("add", scalarExport), false)
	overwritingDraft := maps.Clone(destinationArtifact)
	overwritingDraft["overwritten"] = true
	validateContractValue(t, contract.SchemaID("draft.show"), artifactEnvelopeValue("draft.show", overwritingDraft), false)
}

func artifactContractData(encoding string) map[string]any {
	data := map[string]any{
		"target": "demo", "media_type": "application/json", "encoding": encoding,
		"json_content": nil, "text_content": nil, "base64_content": nil, "destination": nil,
		"size_bytes": 2, "sha256": strings.Repeat("0", 64), "overwritten": false,
	}
	switch encoding {
	case "json":
		data["json_content"] = map[string]any{}
	case "utf-8":
		data["media_type"] = "application/xml"
		data["text_content"] = "ok"
	case "base64":
		data["base64_content"] = "AA=="
	}
	return data
}

func artifactEnvelopeValue(command string, data map[string]any) map[string]any {
	return map[string]any{
		"schema": contract.SchemaID(command), "contract_version": contract.Version,
		"command": command, "requested_command": command, "outcome": "success", "exit_code": 0,
		"producer": map[string]any{"name": "fbrcm", "version": "test"},
		"context":  map[string]any{"profile": "default", "offline": false, "dry_run": false, "draft": false},
		"data":     data, "errors": []any{}, "warnings": []any{},
	}
}

func TestResponseSchemasRejectImpossibleDTOStates(t *testing.T) {
	root := NewRootForContract("test")
	tests := []struct {
		path []string
		raw  string
		edit func(map[string]any)
	}{
		{
			path: []string{"projects", "aliases", "remove"},
			raw:  `{"alias":"prod","previous_project_id":"demo","status":"removed","changed":true,"source":"fbrcm"}`,
			edit: func(data map[string]any) { data["status"] = "not_found" },
		},
		{
			path: []string{"projects", "promote"},
			raw:  `{"source_project":"source","target_project":"target","status":"published","changed":true,"dry_run":false,"published":true,"validated":true,"validation_source":"firebase","selected":1,"change_note":null,"summary":{"conditions":{"added":0,"removed":0,"changed":0,"unchanged":0},"parameters":{"added":1,"removed":0,"changed":0,"unchanged":0},"group_descriptions":{"added":0,"removed":0,"changed":0,"unchanged":0}},"error":null}`,
			edit: func(data map[string]any) { data["published"] = false },
		},
		{
			path: []string{"project", "import"},
			raw:  `{"project_id":"demo","status":"imported","changed":true,"draft":false,"dry_run":false,"validated":true,"validation_source":"firebase","change_note":null,"published":true,"error":null}`,
			edit: func(data map[string]any) { data["published"] = false },
		},
		{
			path: []string{"draft", "publish"},
			raw:  `[{"project_id":"demo","status":"published","base_version":"1","previous_version":"2","published_version":"3","rebased":false,"changed":true,"draft_deleted":true,"dry_run":false,"validated":true,"validation_source":"firebase","error":null,"change_note":null}]`,
			edit: func(data map[string]any) {
				data["items"].([]any)[0].(map[string]any)["published_version"] = ""
			},
		},
		{
			path: []string{"draft", "publish"},
			raw:  `[{"project_id":"demo","status":"published-hook-failed","base_version":"1","previous_version":"2","published_version":"3","rebased":false,"changed":true,"draft_deleted":false,"dry_run":false,"validated":true,"validation_source":"firebase","error":{"stage":"post_publish_hook","message":"hook failed"},"change_note":null}]`,
			edit: func(data map[string]any) {
				data["items"].([]any)[0].(map[string]any)["validation_source"] = "local"
			},
		},
		{
			path: []string{"draft", "publish"},
			raw:  `[{"project_id":"demo","status":"published-cache-failed","base_version":"1","previous_version":"2","published_version":"3","rebased":false,"changed":true,"draft_deleted":false,"dry_run":false,"validated":true,"validation_source":"firebase","error":{"stage":"cache","message":"cache failed"},"change_note":null}]`,
			edit: func(data map[string]any) {
				data["items"].([]any)[0].(map[string]any)["validation_source"] = "local"
			},
		},
		{
			path: []string{"draft", "publish"},
			raw:  `[{"project_id":"demo","status":"published-cleanup-failed","base_version":"1","previous_version":"2","published_version":"3","rebased":false,"changed":true,"draft_deleted":false,"dry_run":false,"validated":true,"validation_source":"firebase","error":{"stage":"cleanup","message":"cleanup failed"},"change_note":null}]`,
			edit: func(data map[string]any) {
				data["items"].([]any)[0].(map[string]any)["validation_source"] = "local"
			},
		},
		{
			path: []string{"versions", "restore"},
			raw:  `{"project_id":"demo","operation":"restore","status":"published","previous_version":"2","source_version":"1","published_version":"3","dry_run":false,"changed":true,"validated":true,"validation_source":"firebase","change_note":null,"error":null}`,
			edit: func(data map[string]any) { data["published_version"] = nil },
		},
		{
			path: []string{"versions", "restore"},
			raw:  `{"project_id":"demo","operation":"restore","status":"published","previous_version":"2","source_version":"1","published_version":"3","dry_run":false,"changed":true,"validated":true,"validation_source":"firebase","change_note":null,"error":null}`,
			edit: func(data map[string]any) { data["operation"] = "rollback" },
		},
		{
			path: []string{"versions", "restore"},
			raw:  `{"project_id":"demo","operation":"restore","status":"published-hook-failed","previous_version":"2","source_version":"1","published_version":"3","dry_run":false,"changed":true,"validated":true,"validation_source":"firebase","change_note":null,"error":{"stage":"post_publish_hook","message":"hook failed"}}`,
			edit: func(data map[string]any) { data["validation_source"] = "local" },
		},
		{
			path: []string{"project", "open"},
			raw:  `{"project_id":"demo","url":"https://console.firebase.google.com/project/demo/config","opened":false}`,
			edit: func(data map[string]any) { data["opened"] = true },
		},
		{
			path: []string{"conditions", "validate"},
			raw:  `{"project":{"name":"Demo","project_id":"demo","auth_id":"main"},"source":"firebase","valid":true}`,
			edit: func(data map[string]any) { data["valid"] = false },
		},
		{
			path: []string{"project", "templates", "show"},
			raw:  `{"project":"Demo","project_id":"demo","templates":["client"],"primary_template":"client"}`,
			edit: func(data map[string]any) { data["primary_template"] = "server" },
		},
		{
			path: []string{"auth", "add", "oauth"},
			raw:  `{"auth_id":"main","type":"oauth","label":"Main","status":"added","paths":{"id":"main","type":"oauth","auth_config_path":"/config/auth.json","profile_config_path":"/config/profile.json","client_secret_path":"/config/client.json","token_path":"/config/token.json"}}`,
			edit: func(data map[string]any) {
				data["paths"].(map[string]any)["service_account_path"] = "/config/service.json"
			},
		},
		{
			path: []string{"config", "path"},
			raw:  `{"scope":"global","path":"/config/fbrcm.toml","exists":true}`,
			edit: func(data map[string]any) { data["scope"] = "effective" },
		},
		{
			path: []string{"config", "reset"},
			raw:  `{"scope":"global","key":"preferences","status":"reset","changed":true}`,
			edit: func(data map[string]any) { data["changed"] = false },
		},
		{
			path: []string{"auth", "bind"},
			raw:  `{"auth_id":"main","bound":1,"skipped":0,"items":[{"project_id":"demo","status":"bound","reason":null}]}`,
			edit: func(data map[string]any) { data["items"].([]any)[0].(map[string]any)["reason"] = "impossible" },
		},
		{
			path: []string{"update"},
			raw:  `[{"target":"demo","status":"unchanged","changed_item_count":0,"previous_version":null,"published_version":null,"draft":false,"dry_run":false,"validated":true,"validation_source":"local","error":null,"retry_selector":null,"change_note":null,"selection":{"default_scope":true,"resolved_target_count":1,"matched_item_count":0},"no_op_reason":"no_match"}]`,
			edit: func(data map[string]any) {
				data["items"].([]any)[0].(map[string]any)["changed_item_count"] = float64(1)
			},
		},
		{
			path: []string{"update"},
			raw:  `[{"target":"demo","status":"published","changed_item_count":1,"previous_version":"1","published_version":"2","draft":false,"dry_run":false,"validated":true,"validation_source":"firebase","error":null,"retry_selector":null,"change_note":null,"selection":{"default_scope":true,"resolved_target_count":1,"matched_item_count":1},"no_op_reason":null}]`,
			edit: func(data map[string]any) {
				data["items"].([]any)[0].(map[string]any)["validation_source"] = "local"
			},
		},
		{
			path: []string{"update"},
			raw:  `[{"target":"demo","status":"published-hook-failed","changed_item_count":1,"previous_version":"1","published_version":"2","draft":false,"dry_run":false,"validated":true,"validation_source":"firebase","error":{"stage":"post_publish_hook","message":"hook failed"},"retry_selector":null,"change_note":null,"selection":{"default_scope":true,"resolved_target_count":1,"matched_item_count":1},"no_op_reason":null}]`,
			edit: func(data map[string]any) {
				data["items"].([]any)[0].(map[string]any)["error"].(map[string]any)["stage"] = "preparation"
			},
		},
	}
	for _, test := range tests {
		cmd, remaining, err := root.Find(test.path)
		if err != nil || len(remaining) != 0 {
			t.Fatalf("find %v: remaining=%v err=%v", test.path, remaining, err)
		}
		envelope := contract.BuildEnvelope(cmd, "test", []byte(test.raw), nil)
		if envelope.Outcome != "success" {
			t.Fatalf("valid %v fixture produced %#v", test.path, envelope)
		}
		document := structToContractValue(t, envelope).(map[string]any)
		test.edit(document["data"].(map[string]any))
		validateContractValue(t, contract.SchemaID(contract.CommandID(cmd)), document, false)
	}
}

func TestProjectsResetResponseAcceptsChangedAndNoOpResults(t *testing.T) {
	root := NewRootForContract("test")
	cmd, remaining, err := root.Find([]string{"projects", "reset"})
	if err != nil || len(remaining) != 0 {
		t.Fatalf("find projects reset: remaining=%v err=%v", remaining, err)
	}
	for _, changed := range []bool{false, true} {
		raw := fmt.Appendf(nil, `{"path":"/tmp/projects-config.json","status":"reset","changed":%t}`, changed)
		envelope := contract.BuildEnvelope(cmd, "test", raw, nil)
		if envelope.Outcome != "success" {
			t.Fatalf("changed=%t envelope = %#v", changed, envelope)
		}
	}
}

func TestProfileRenameResponseAcceptsSameNameNoOp(t *testing.T) {
	root := NewRootForContract("test")
	cmd, remaining, err := root.Find([]string{"profile", "rename"})
	if err != nil || len(remaining) != 0 {
		t.Fatalf("find profile rename: remaining=%v err=%v", remaining, err)
	}
	envelope := contract.BuildEnvelope(cmd, "test", []byte(`{"old_profile":"same","new_profile":"same","effective_profile":"same","changed":false}`), nil)
	if envelope.Outcome != "success" {
		t.Fatalf("same-name profile rename envelope = %#v", envelope)
	}
	validateContractDocument(t, envelope.Schema, marshalEnvelope(t, envelope))
}

func TestSchemasPublishNonCalculableDTOInvariants(t *testing.T) {
	for _, id := range []string{
		contract.SchemaID("capabilities"),
		contract.SchemaID("schema.list"),
		contract.SchemaID("projects.forget"),
		contract.SchemaID("auth.list"),
		contract.SchemaID("profile.list"),
		contract.SchemaID("profile.switch"),
		contract.SchemaID("profile.rename"),
		contract.SchemaID("projects.aliases.import"),
		contract.SchemaID("versions.show"),
		contract.SchemaID("projects.list"),
		contract.SchemaID("auth.bind"),
		contract.SchemaID("project.export"),
	} {
		raw, err := schemas.ReadByID(id)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(raw, []byte(`"x-fbrcm-invariants"`)) {
			t.Errorf("%s does not publish normative invariants", id)
		}
	}
}

func TestSemanticSchemaDefinesStructuredExtensionLanguage(t *testing.T) {
	raw, err := schemas.ReadByID(contract.SemanticSchemaID())
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	language, ok := document["x-fbrcm-extension-language"].(map[string]any)
	if !ok || language["version"] != float64(1) {
		t.Fatalf("extension language = %#v", language)
	}
	for _, section := range []string{"validation", "normalization", "matching", "invariants"} {
		metadata, ok := language[section].(map[string]any)
		if !ok {
			t.Fatalf("extension language omits %s: %#v", section, language)
		}
		operators, ok := metadata["operators"].(map[string]any)
		if !ok || len(operators) == 0 {
			t.Fatalf("extension language %s operators = %#v", section, metadata["operators"])
		}
	}
	effectiveness, ok := language["option_effectiveness"].(map[string]any)
	if !ok || effectiveness["keyword"] != "x-fbrcm-effective" {
		t.Fatalf("extension language omits option effectiveness: %#v", language)
	}
	conditionalEffectiveness, ok := language["conditional_option_effectiveness"].(map[string]any)
	if !ok || conditionalEffectiveness["keyword"] != "x-fbrcm-effective-when" {
		t.Fatalf("extension language omits conditional option effectiveness: %#v", language)
	}
}

func TestProblemSchemasRequireKnownDetailKinds(t *testing.T) {
	base := func(code, category string, details any) map[string]any {
		return map[string]any{
			"code": code, "category": category, "message": "failure", "retryable": false,
			"target": nil, "stage": nil, "details": details, "remediation": []any{},
		}
	}
	id := contract.ErrorSchemaID()
	validateContractValue(t, id, base("profile.invalid", "profile", map[string]any{"kind": "validation", "source": "profile"}), true)
	validateContractValue(t, id, base("profile.invalid", "profile", nil), false)
	validateContractValue(t, id, base("profile.not_found", "not_found", map[string]any{"kind": "selection", "resource": "profile", "query": "missing", "candidates": []any{}}), true)
	validateContractValue(t, id, base("profile.conflict", "conflict", map[string]any{"kind": "conflict", "resource": "profile"}), true)
	validateContractValue(t, id, base("filesystem.permission_denied", "permission", map[string]any{"kind": "file", "operation": "open", "path": "/tmp/config"}), true)
	validateContractValue(t, id, base("filesystem.permission_denied", "permission", nil), false)
	validateContractValue(t, id, base("configuration.invalid", "configuration", map[string]any{"kind": "validation", "source": "configuration"}), true)
	validateContractValue(t, id, base("configuration.invalid", "configuration", nil), false)
	retryableConfiguration := base("configuration.invalid", "configuration", map[string]any{"kind": "validation", "source": "configuration"})
	retryableConfiguration["retryable"] = true
	validateContractValue(t, id, retryableConfiguration, false)
	validateContractValue(t, id, base("auth.setup_required", "auth", map[string]any{"kind": "auth"}), true)
	validateContractValue(t, id, base("auth.setup_required", "auth", nil), false)
	validateContractValue(t, id, base("auth.credentials_invalid", "auth", map[string]any{"kind": "validation", "source": "auth"}), true)
	authUnavailable := base("network.unavailable", "unavailable", map[string]any{"kind": "remote_api", "service": "google_auth", "operation": "refresh_token", "http_status": 503, "remote_status": "", "remote_code": "temporarily_unavailable", "retry_after_ms": 0})
	authUnavailable["retryable"] = true
	validateContractValue(t, id, authUnavailable, true)
	validateContractValue(t, id, base("remote_config.validation_failed", "validation", map[string]any{"kind": "validation", "source": "local"}), true)
	validateContractValue(t, id, base("remote_config.validation_failed", "validation", map[string]any{"kind": "validation", "source": "firebase"}), true)
	validateContractValue(t, id, base("remote_config.validation_failed", "validation", map[string]any{"kind": "validation", "source": "configuration"}), false)
}

func TestJSONArgumentFailureHasNoUsageOrLegacyOutput(t *testing.T) {
	envelope, raw := executeJSONContract(t, "does-not-exist", "--json")
	if envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Category != "argument" {
		t.Fatalf("envelope = %#v", envelope)
	}
	if strings.Contains(string(raw), "Usage:") || strings.Contains(string(raw), "unknown command") && !bytes.Contains(raw, []byte(`"message"`)) {
		t.Fatalf("failure output is contaminated: %s", raw)
	}
}

func TestJSONRequiredFlagFailureUsesArgumentExitStatus(t *testing.T) {
	envelope, raw := executeJSONContract(t, "projects", "aliases", "import", "--json")
	if envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Category != "argument" {
		t.Fatalf("envelope = %#v", envelope)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestJSONRejectsNonPositiveGlobalTimeout(t *testing.T) {
	envelope, raw := executeJSONContract(t, "capabilities", "--timeout", "0s", "--json")
	if envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Category != "argument" {
		t.Fatalf("envelope = %#v", envelope)
	}
	if strings.Contains(string(raw), "Usage:") {
		t.Fatalf("failure output is contaminated: %s", raw)
	}
}

func TestJSONInvalidConfigScopeIsAnArgumentFailure(t *testing.T) {
	envelope, raw := executeJSONContract(t, "config", "path", "--scope", "garbage", "--json")
	if envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" || envelope.Errors[0].Category != "argument" {
		t.Fatalf("envelope = %#v", envelope)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestJSONUnknownConfigKeyIsAnArgumentFailure(t *testing.T) {
	envelope, raw := executeJSONContract(t, "config", "show", "not-a-key", "--json")
	if envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" || envelope.Errors[0].Category != "argument" {
		t.Fatalf("envelope = %#v", envelope)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestJSONConfigShowAcceptsTrimmedNestedKey(t *testing.T) {
	envelope, raw := executeJSONContract(t, "config", "show", " keys.global.quit ", "--scope", "global", "--json")
	if envelope.Outcome != "success" || envelope.ExitCode != 0 {
		t.Fatalf("envelope = %#v", envelope)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestJSONConfigEditReturnsInteractionBeforeHumanFlagValidation(t *testing.T) {
	envelope, raw := executeJSONContract(t, "config", "edit", "--scope", "invalid", "--editor", "ignored", "--full", "--json")
	if envelope.Outcome != "failure" || envelope.ExitCode != 10 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "interaction.required" {
		t.Fatalf("envelope = %#v", envelope)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestJSONBlankUpdateParameterIsAnArgumentFailure(t *testing.T) {
	envelope, raw := executeJSONContract(t, "update", "   ", "--description", "updated", "--json")
	if envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" || envelope.Errors[0].Category != "argument" {
		t.Fatalf("envelope = %#v", envelope)
	}
	validateContractDocument(t, envelope.Schema, raw)
}

func TestRequestedTimeoutUsesLastValue(t *testing.T) {
	if got := requestedTimeout([]string{"--timeout=1s", "show", "--timeout", "2s"}); got.String() != "2s" {
		t.Fatalf("timeout = %s", got)
	}
}

func executeJSONContract(t *testing.T, args ...string) (contract.Envelope, []byte) {
	t.Helper()
	root := NewRootForContract("test")
	var captured bytes.Buffer
	root.SetOut(&captured)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(args)
	root.SilenceErrors, root.SilenceUsage = true, true
	shared.SetMachineMode(true)
	t.Cleanup(func() { shared.SetMachineMode(false) })
	executed, err := root.ExecuteC()
	envelope := contract.BuildEnvelope(executed, "test", captured.Bytes(), err)
	var output bytes.Buffer
	if writeErr := contract.Write(&output, envelope); writeErr != nil {
		t.Fatal(writeErr)
	}
	return envelope, output.Bytes()
}

func validateContractDocument(t *testing.T, schemaID string, raw []byte) {
	t.Helper()
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("decode contract document: %v", err)
	}
	validateContractValue(t, schemaID, value, true)
}

func marshalEnvelope(t *testing.T, envelope contract.Envelope) []byte {
	t.Helper()
	var output bytes.Buffer
	if err := contract.Write(&output, envelope); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

type contractSchemaRegistry struct {
	mu       sync.Mutex
	compiler *jsonschema.Compiler
	compiled map[string]*jsonschema.Schema
}

var (
	contractSchemasOnce sync.Once
	contractSchemas     *contractSchemaRegistry
	contractSchemasErr  error
)

func loadContractSchemaRegistry() {
	ids, err := schemas.List()
	if err != nil {
		contractSchemasErr = err
		return
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	for _, id := range ids {
		raw, readErr := schemas.ReadByID(id)
		if readErr != nil {
			contractSchemasErr = readErr
			return
		}
		var document any
		if unmarshalErr := json.Unmarshal(raw, &document); unmarshalErr != nil {
			contractSchemasErr = fmt.Errorf("decode schema %s: %w", id, unmarshalErr)
			return
		}
		if addErr := compiler.AddResource(id, document); addErr != nil {
			contractSchemasErr = fmt.Errorf("register schema %s: %w", id, addErr)
			return
		}
	}
	contractSchemas = &contractSchemaRegistry{
		compiler: compiler,
		compiled: make(map[string]*jsonschema.Schema),
	}
}

func compiledContractSchema(id string) (*jsonschema.Schema, error) {
	contractSchemasOnce.Do(loadContractSchemaRegistry)
	if contractSchemasErr != nil {
		return nil, contractSchemasErr
	}
	contractSchemas.mu.Lock()
	defer contractSchemas.mu.Unlock()
	if compiled, ok := contractSchemas.compiled[id]; ok {
		return compiled, nil
	}
	compiled, err := contractSchemas.compiler.Compile(id)
	if err != nil {
		return nil, fmt.Errorf("compile %s: %w", id, err)
	}
	contractSchemas.compiled[id] = compiled
	return compiled, nil
}

func validateContractValue(t *testing.T, schemaID string, value any, valid bool) {
	t.Helper()
	compiled, err := compiledContractSchema(schemaID)
	if err != nil {
		t.Fatal(err)
	}
	err = compiled.Validate(value)
	if valid && err != nil {
		t.Fatalf("document does not conform to %s: %v", schemaID, err)
	}
	if !valid && err == nil {
		t.Fatalf("document unexpectedly conforms to %s: %#v", schemaID, value)
	}
}

func structToContractValue(t *testing.T, value any) any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result any
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func hasContractFlag(root *cobra.Command, path []string, name string) bool {
	cmd, _, err := root.Find(path)
	if err != nil {
		return false
	}
	return cmd.Flags().Lookup(name) != nil || cmd.InheritedFlags().Lookup(name) != nil
}

func commandForCapability(t *testing.T, root *cobra.Command, capability contract.Capability) *cobra.Command {
	t.Helper()
	if capability.ID == "root" {
		return root
	}
	cmd, remaining, err := root.Find(capability.Path)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("find %s: error=%v remaining=%v", capability.ID, err, remaining)
	}
	return cmd
}

func responseDataSchema(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	allOf, ok := document["allOf"].([]any)
	if !ok || len(allOf) != 2 {
		t.Fatalf("response schema has invalid allOf: %s", raw)
	}
	specialization, ok := allOf[1].(map[string]any)
	if !ok {
		t.Fatalf("response specialization = %#v", allOf[1])
	}
	properties, ok := specialization["properties"].(map[string]any)
	if !ok {
		t.Fatalf("response properties = %#v", specialization["properties"])
	}
	data, ok := properties["data"].(map[string]any)
	if !ok || len(data) == 0 {
		t.Fatalf("response data schema is missing or unconstrained: %#v", properties["data"])
	}
	variants, hasVariants := data["oneOf"].([]any)
	if hasVariants && len(variants) == 2 {
		first, _ := variants[0].(map[string]any)
		if first["$ref"] == "#/$defs/success_data" {
			definitions, _ := document["$defs"].(map[string]any)
			success, _ := definitions["success_data"].(map[string]any)
			if len(success) == 0 {
				t.Fatal("response success_data definition is missing")
			}
			return map[string]any{"oneOf": []any{success, variants[1]}}
		}
	}
	return data
}

func collectionItemProperties(t *testing.T, data map[string]any) map[string]any {
	t.Helper()
	variants, _ := data["oneOf"].([]any)
	for _, variant := range variants {
		object, _ := variant.(map[string]any)
		properties, _ := object["properties"].(map[string]any)
		items, _ := properties["items"].(map[string]any)
		item, _ := items["items"].(map[string]any)
		if itemProperties, ok := item["properties"].(map[string]any); ok {
			return itemProperties
		}
	}
	t.Fatalf("data schema has no collection item DTO: %#v", data)
	return nil
}

func containsArbitraryJSONDocument(data map[string]any) bool {
	if data["description"] == "arbitrary JSON document" {
		return true
	}
	variants, _ := data["oneOf"].([]any)
	for _, variant := range variants {
		object, _ := variant.(map[string]any)
		if object["description"] == "arbitrary JSON document" {
			return true
		}
	}
	return false
}

func assertResolvablePublishedSchemaReferences(t *testing.T, command string, raw []byte) {
	t.Helper()
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	semanticRaw, err := schemas.ReadByID(contract.SemanticSchemaID())
	if err != nil {
		t.Fatal(err)
	}
	var semanticDocument any
	if err := json.Unmarshal(semanticRaw, &semanticDocument); err != nil {
		t.Fatal(err)
	}
	var visit func(any)
	visit = func(value any) {
		switch value := value.(type) {
		case map[string]any:
			for key, child := range value {
				if key == "$ref" {
					ref, _ := child.(string)
					if strings.HasPrefix(ref, contract.SemanticSchemaID()+"#/") {
						pointer := strings.TrimPrefix(ref, contract.SemanticSchemaID())
						if !resolvesLocalJSONPointer(semanticDocument, pointer) {
							t.Errorf("%s has unresolved semantic-schema reference %q", command, ref)
						}
					} else if ref == contract.EnvelopeSchemaID() || ref == contract.CapabilitySchemaID() {
						if _, err := schemas.ReadByID(ref); err != nil {
							t.Errorf("%s has unresolved envelope-schema reference %q", command, ref)
						}
					} else if strings.HasPrefix(ref, "#") && !strings.HasPrefix(ref, "#/") {
						if !resolvesLocalAnchor(document, strings.TrimPrefix(ref, "#")) {
							t.Errorf("%s has unresolved response-schema anchor %q", command, ref)
						}
					} else if !strings.HasPrefix(ref, "#/") {
						t.Errorf("%s has external response-schema reference %q", command, ref)
					} else if !resolvesLocalJSONPointer(document, ref) {
						t.Errorf("%s has unresolved response-schema reference %q", command, ref)
					}
				}
				visit(child)
			}
		case []any:
			for _, child := range value {
				visit(child)
			}
		}
	}
	visit(document)
}

func resolvesLocalAnchor(value any, anchor string) bool {
	switch value := value.(type) {
	case map[string]any:
		if value["$anchor"] == anchor {
			return true
		}
		for _, child := range value {
			if resolvesLocalAnchor(child, anchor) {
				return true
			}
		}
	case []any:
		for _, child := range value {
			if resolvesLocalAnchor(child, anchor) {
				return true
			}
		}
	}
	return false
}

func resolvesLocalJSONPointer(document any, ref string) bool {
	current := document
	for token := range strings.SplitSeq(strings.TrimPrefix(ref, "#/"), "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		object, ok := current.(map[string]any)
		if !ok {
			return false
		}
		current, ok = object[token]
		if !ok {
			return false
		}
	}
	return true
}

func walkSchemaProperties(value any, visit func(string, map[string]any)) {
	switch value := value.(type) {
	case map[string]any:
		if properties, ok := value["properties"].(map[string]any); ok {
			for name, rawSchema := range properties {
				if schema, ok := rawSchema.(map[string]any); ok {
					visit(name, schema)
				}
			}
		}
		for _, child := range value {
			walkSchemaProperties(child, visit)
		}
	case []any:
		for _, child := range value {
			walkSchemaProperties(child, visit)
		}
	}
}
