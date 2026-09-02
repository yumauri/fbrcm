package contract

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/machine"
)

const commandGroupAnnotation = "fbrcm.contract.command-group"

type Support struct {
	DryRun             bool `json:"dry_run"`
	Draft              bool `json:"draft"`
	Plan               bool `json:"plan"`
	ConfirmationBypass bool `json:"confirmation_bypass"`
	Stdin              bool `json:"stdin"`
	Stateless          bool `json:"stateless"`
}

type ArgumentCapability struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Repeated bool   `json:"repeated"`
	Schema   string `json:"schema"`
}

type FlagCapability struct {
	Name          string                    `json:"name"`
	Aliases       []string                  `json:"aliases"`
	Type          string                    `json:"type"`
	Default       any                       `json:"default"`
	Required      bool                      `json:"required"`
	Repeatable    bool                      `json:"repeatable"`
	Effective     bool                      `json:"effective"`
	EffectiveWhen []BehaviorConditionClause `json:"effective_when,omitempty"`
	Usage         string                    `json:"usage"`
}

type InteractionCapability struct {
	Mode         string `json:"mode"`
	JSONBehavior string `json:"json_behavior"`
}

type IdempotencyCondition struct {
	Idempotency string                    `json:"idempotency"`
	When        []BehaviorConditionClause `json:"when"`
}

// BehaviorPredicate is one atomic test in a capability behavior condition.
// Clauses combine predicates with AND; the containing clause list uses OR.
type BehaviorPredicate struct {
	Source   string `json:"source"`
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type BehaviorConditionClause struct {
	AllOf []BehaviorPredicate `json:"all_of"`
}

type SideEffectCondition struct {
	Effect string                    `json:"effect"`
	When   []BehaviorConditionClause `json:"when"`
}

type Capability struct {
	ID                 string                    `json:"id"`
	Path               []string                  `json:"path"`
	Summary            string                    `json:"summary"`
	Arguments          []ArgumentCapability      `json:"arguments"`
	Flags              []FlagCapability          `json:"flags"`
	InvocationSchema   string                    `json:"invocation_schema"`
	StdinSchema        *string                   `json:"stdin_schema"`
	ResponseSchema     string                    `json:"response_schema"`
	ErrorSchema        string                    `json:"error_schema"`
	ProblemCodes       []string                  `json:"problem_codes"`
	SideEffectLevel    int                       `json:"side_effect_level"`
	SideEffects        []string                  `json:"side_effects"`
	SideEffectWhen     []SideEffectCondition     `json:"side_effect_when"`
	NetworkAccess      string                    `json:"network_access"`
	NetworkWhen        []BehaviorConditionClause `json:"network_when"`
	Destructive        bool                      `json:"destructive"`
	DestructiveWhen    []BehaviorConditionClause `json:"destructive_when"`
	DestructiveReasons []string                  `json:"destructive_reasons"`
	Idempotency        string                    `json:"idempotency"`
	IdempotencyWhen    []IdempotencyCondition    `json:"idempotency_when"`
	Supports           Support                   `json:"supports"`
	StdinModes         []string                  `json:"stdin_modes"`
	Interaction        InteractionCapability     `json:"interaction"`
	InteractionWhen    []BehaviorConditionClause `json:"interaction_when"`
}

// CapabilitySummary is the compact discovery record returned by the capability
// index. Detailed arguments, flags, and behavior remain available through an
// exact capabilities <command...> lookup.
type CapabilitySummary struct {
	ID               string   `json:"id"`
	Path             []string `json:"path"`
	Summary          string   `json:"summary"`
	InvocationSchema string   `json:"invocation_schema"`
	ResponseSchema   string   `json:"response_schema"`
	SideEffectLevel  int      `json:"side_effect_level"`
	Destructive      bool     `json:"destructive"`
	Supports         Support  `json:"supports"`
}

type CapabilityIndex struct {
	ContractVersion string              `json:"contract_version"`
	Count           int                 `json:"count"`
	Commands        []CapabilitySummary `json:"commands"`
}

// CapabilitySchemaID identifies the strict detailed-capability schema.
func CapabilitySchemaID() string {
	return "urn:fbrcm:schema:cli:" + Version + ":capability"
}

func Capabilities(root *cobra.Command) CapabilityIndex {
	detailed := DetailedCapabilities(root)
	commands := make([]CapabilitySummary, 0, len(detailed))
	for _, capability := range detailed {
		commands = append(commands, CapabilitySummary{
			ID:               capability.ID,
			Path:             capability.Path,
			Summary:          capability.Summary,
			InvocationSchema: capability.InvocationSchema,
			ResponseSchema:   capability.ResponseSchema,
			SideEffectLevel:  capability.SideEffectLevel,
			Destructive:      capability.Destructive,
			Supports:         capability.Supports,
		})
	}
	return CapabilityIndex{ContractVersion: Version, Count: len(commands), Commands: commands}
}

// DetailedCapabilities returns the authoritative detailed records used by
// schema generation and conformance tests.
func DetailedCapabilities(root *cobra.Command) []Capability {
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()
	commands := make([]Capability, 0)
	seen := make(map[string]bool)
	walkCommands(root, func(cmd *cobra.Command) {
		capability := describe(cmd)
		if err := validateCapability(capability); err != nil {
			panic(fmt.Sprintf("invalid authoritative capability %q: %v", capability.ID, err))
		}
		seen[capability.ID] = true
		commands = append(commands, capability)
	})
	for id := range capabilityBehaviors {
		if !seen[id] {
			panic(fmt.Sprintf("authoritative capability behavior references unknown command %q", id))
		}
	}
	sort.Slice(commands, func(i, j int) bool { return commands[i].ID < commands[j].ID })
	return commands
}

func validateCapability(capability Capability) error {
	wantPath := []string{}
	if capability.ID != "root" {
		wantPath = strings.Split(capability.ID, ".")
	}
	if !slices.Equal(capability.Path, wantPath) {
		return fmt.Errorf("path %v does not match id", capability.Path)
	}
	if capability.InvocationSchema != "urn:fbrcm:schema:cli:"+Version+":command:"+capability.ID+":input" || capability.ResponseSchema != SchemaID(capability.ID) || capability.ErrorSchema != ErrorSchemaID() {
		return fmt.Errorf("schema URNs do not match id and contract version")
	}
	if capability.Supports.Stdin != (capability.StdinSchema != nil) || capability.Supports.Stdin != (len(capability.StdinModes) > 0) {
		return fmt.Errorf("stdin support, schema, and modes disagree")
	}
	if !capability.Supports.Stdin && len(capability.StdinModes) != 0 {
		return fmt.Errorf("stdin modes declared without stdin support")
	}
	if len(capability.ProblemCodes) == 0 || !sort.StringsAreSorted(capability.ProblemCodes) {
		return fmt.Errorf("problem_codes must be a non-empty sorted set")
	}
	for index, code := range capability.ProblemCodes {
		if index > 0 && capability.ProblemCodes[index-1] == code || !slices.Contains(knownProblemCodeValues, code) {
			return fmt.Errorf("problem_codes contains duplicate or unknown code %q", code)
		}
	}
	flags := make(map[string]bool, len(capability.Flags))
	for _, flag := range capability.Flags {
		flags[strings.TrimPrefix(flag.Name, "--")] = true
	}
	arguments := make(map[string]bool, len(capability.Arguments))
	for _, argument := range capability.Arguments {
		arguments[argument.Name] = true
	}
	validateClauses := func(label string, clauses []BehaviorConditionClause) error {
		for _, clause := range clauses {
			for _, item := range clause.AllOf {
				if item.Source == "argument" && !arguments[item.Name] {
					return fmt.Errorf("%s references unknown argument %s", label, item.Name)
				}
				if item.Source == "option" && !flags[item.Name] {
					return fmt.Errorf("%s references unknown option --%s", label, item.Name)
				}
				if item.Source == "stdin" && !capability.Supports.Stdin {
					return fmt.Errorf("%s references stdin for a command without stdin support", label)
				}
			}
		}
		return nil
	}
	if err := validateClauses("network_when", capability.NetworkWhen); err != nil {
		return err
	}
	if err := validateClauses("destructive_when", capability.DestructiveWhen); err != nil {
		return err
	}
	if err := validateClauses("interaction_when", capability.InteractionWhen); err != nil {
		return err
	}
	if capability.NetworkAccess == "conditional" && len(capability.NetworkWhen) == 0 || capability.NetworkAccess != "conditional" && len(capability.NetworkWhen) != 0 {
		return fmt.Errorf("network_access and network_when disagree")
	}
	if capability.Interaction.Mode == "none" && len(capability.InteractionWhen) != 0 || capability.Interaction.Mode != "none" && len(capability.InteractionWhen) == 0 {
		return fmt.Errorf("interaction mode and interaction_when disagree")
	}
	if capability.Idempotency == "conditional" && len(capability.IdempotencyWhen) == 0 || capability.Idempotency != "conditional" && len(capability.IdempotencyWhen) != 0 {
		return fmt.Errorf("idempotency and idempotency_when disagree")
	}
	if len(capability.SideEffects) != len(capability.SideEffectWhen) {
		return fmt.Errorf("side_effects and side_effect_when lengths disagree")
	}
	seenEffects := make(map[string]bool, len(capability.SideEffects))
	for index, name := range capability.SideEffects {
		if seenEffects[name] || capability.SideEffectWhen[index].Effect != name {
			return fmt.Errorf("side_effect_when must correspond one-to-one with unique side_effects")
		}
		seenEffects[name] = true
		if err := validateClauses("side_effect_when", capability.SideEffectWhen[index].When); err != nil {
			return err
		}
	}
	for _, item := range capability.IdempotencyWhen {
		if item.Idempotency == "conditional" {
			return fmt.Errorf("idempotency_when cannot recursively be conditional")
		}
		if err := validateClauses("idempotency_when", item.When); err != nil {
			return err
		}
	}
	return nil
}

func FindCapability(root *cobra.Command, path []string) (Capability, error) {
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()
	if len(path) == 1 && path[0] == "root" {
		return describe(root), nil
	}
	cmd, remaining, err := root.Find(path)
	if err != nil {
		return Capability{}, &machine.SelectionError{Resource: "command", Kind: "not_found", Query: strings.Join(path, " "), Err: err}
	}
	if len(remaining) != 0 {
		return Capability{}, &machine.SelectionError{Resource: "command", Kind: "not_found", Query: strings.Join(path, " ")}
	}
	if cmd != root && IsCommandGroup(cmd) {
		return Capability{}, &machine.ArgumentError{Code: "command.not_executable", Err: fmt.Errorf("%s is not an executable command", strings.Join(path, " "))}
	}
	return describe(cmd), nil
}

func walkCommands(cmd *cobra.Command, visit func(*cobra.Command)) {
	if cmd.Parent() == nil || !IsCommandGroup(cmd) && (cmd.RunE != nil || cmd.Run != nil) {
		visit(cmd)
	}
	for _, child := range cmd.Commands() {
		walkCommands(child, visit)
	}
}

// MarkCommandGroup records that a Cobra node is navigational rather than an
// executable contract operation, even when it has a RunE that renders help.
func MarkCommandGroup(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[commandGroupAnnotation] = "true"
}

// IsCommandGroup reports whether a command is a non-executable navigation
// node in the machine contract.
func IsCommandGroup(call invocation.Call) bool {
	cmd, ok := call.(*cobra.Command)
	return ok && cmd != nil && cmd.Annotations[commandGroupAnnotation] == "true"
}

func describe(cmd *cobra.Command) Capability {
	id := CommandID(cmd)
	path := []string{}
	if id != "root" {
		path = strings.Split(strings.TrimPrefix(cmd.CommandPath(), "fbrcm "), " ")
	}
	flags := make([]FlagCapability, 0)
	seen := map[string]bool{}
	addFlags := func(set *pflag.FlagSet) {
		set.VisitAll(func(flag *pflag.Flag) {
			if seen[flag.Name] || flag.Name == "json" {
				return
			}
			seen[flag.Name] = true
			aliases := []string{}
			if flag.Shorthand != "" {
				aliases = append(aliases, "-"+flag.Shorthand)
			}
			defaultValue := typedFlagDefault(flag)
			if flag.Name == "profile" {
				// Environment-derived process state is not a stable schema default.
				defaultValue = ""
			}
			effective := !optionIgnored(id, flag.Name)
			usage := flag.Usage
			if !effective {
				usage += "; accepted but not applied by this command"
			}
			var effectiveWhen []BehaviorConditionClause
			if id == "root" && (flag.Name == "profile" || flag.Name == "no-local-config") {
				effectiveWhen = []BehaviorConditionClause{conditionClause(predicate("option", "version", "equals", false))}
				usage += "; accepted but not applied with --version"
			}
			if SupportsStatelessCommand(id) && flag.Name == "profile" {
				effectiveWhen = []BehaviorConditionClause{conditionClause(predicate("option", "stateless", "equals", false))}
				usage += "; cannot be combined with --stateless"
			}
			if SupportsStatelessCommand(id) && flag.Name == "cached" {
				effectiveWhen = []BehaviorConditionClause{conditionClause(predicate("option", "stateless", "equals", false))}
				usage += "; cannot be combined with --stateless"
			}
			if SupportsStatelessCommand(id) && flag.Name == "draft" {
				effectiveWhen = []BehaviorConditionClause{conditionClause(predicate("option", "stateless", "equals", false))}
				usage += "; cannot be combined with --stateless"
			}
			if id == "projects.list" && flag.Name == "update" {
				effectiveWhen = []BehaviorConditionClause{conditionClause(predicate("option", "stateless", "equals", false))}
				usage += "; cannot be combined with --stateless"
			}
			if id == "projects.list" && flag.Name == "expr" {
				usage += "; with --stateless, evaluates against directly fetched client Remote Config after project filtering"
			}
			if slices.Contains([]string{
				"conditions.list", "conditions.show", "experiments.list", "experiments.show", "groups.list",
				"personalizations.list", "personalizations.show", "project.show", "rollouts.list", "rollouts.show",
			}, id) && flag.Name == "update" {
				effectiveWhen = []BehaviorConditionClause{conditionClause(predicate("option", "stateless", "equals", false))}
				usage += "; cannot be combined with --stateless"
			}
			if id == "get" && flag.Name == "update" {
				effectiveWhen = []BehaviorConditionClause{conditionClause(
					predicate("option", "stateless", "equals", false),
					predicate("stdin", "document", "absent", nil),
				)}
				usage += "; cannot be combined with --stateless or stdin"
			}
			if id == "get" && flag.Name == "project" {
				usage += "; with remote --stateless execution, exact (=) targets bypass discovery and other selectors filter remote project IDs and display names"
			}
			if flag.Name == "stateless" && !SupportsStatelessCommand(id) {
				usage += "; true is currently rejected by this command"
			}
			flags = append(flags, FlagCapability{Name: "--" + flag.Name, Aliases: aliases, Type: flag.Value.Type(), Default: defaultValue, Required: flag.Annotations != nil && len(flag.Annotations[cobra.BashCompOneRequiredFlag]) > 0, Repeatable: strings.Contains(flag.Value.Type(), "Slice") || strings.Contains(flag.Value.Type(), "Array"), Effective: effective, EffectiveWhen: effectiveWhen, Usage: usage})
		})
	}
	addFlags(cmd.InheritedFlags())
	addFlags(cmd.Flags())
	if id := CommandID(cmd); id == "root" {
		flags = append(flags, FlagCapability{Name: "--version", Aliases: []string{"-v"}, Type: "bool", Default: false, Effective: true, Usage: "Print build version"})
	}
	sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })

	arguments := parseArguments(cmd)
	behavior, ok := capabilityBehaviors[id]
	if !ok {
		panic(fmt.Sprintf("missing authoritative capability behavior for %q", id))
	}
	behavior = withMachineAuthenticationEffects(id, behavior)
	if id != "mcp" {
		behavior = withJSONEnvelopeProfileBootstrap(behavior)
	}
	if hasFlag(cmd, "plan-out") {
		behavior = withPlanOutputEffects(behavior)
	}
	support := Support{DryRun: hasFlag(cmd, "dry-run"), Draft: hasFlag(cmd, "draft"), Plan: hasFlag(cmd, "plan-out"), ConfirmationBypass: hasFlag(cmd, "yes"), Stdin: behavior.stdin, Stateless: SupportsStatelessCommand(id)}
	interaction := InteractionCapability{Mode: "none", JSONBehavior: "non_interactive"}
	interactionWhen := cloneConditions(behavior.interactionWhen)
	if support.ConfirmationBypass {
		interaction = InteractionCapability{Mode: "optional", JSONBehavior: "confirmation_required_without_bypass"}
		if !containsPredicate(interactionWhen, "option", "yes", "equals") {
			confirmation := conditionClause(
				predicate("option", "yes", "equals", false),
				predicate("runtime_state", "confirmation", "required", nil),
			)
			if support.Plan {
				confirmation.AllOf = append(confirmation.AllOf, predicate("option", "plan-out", "equals", ""))
			}
			interactionWhen = append(interactionWhen, confirmation)
		}
	}
	if support.Plan {
		for index := range interactionWhen {
			if slices.ContainsFunc(interactionWhen[index].AllOf, func(item BehaviorPredicate) bool {
				return item.Source == "option" && item.Name == "yes" && item.Operator == "equals"
			}) && !slices.ContainsFunc(interactionWhen[index].AllOf, func(item BehaviorPredicate) bool {
				return item.Source == "option" && item.Name == "plan-out" && item.Operator == "equals"
			}) {
				interactionWhen[index].AllOf = append(interactionWhen[index].AllOf, predicate("option", "plan-out", "equals", ""))
			}
		}
	}
	if behavior.interaction != nil {
		interaction = *behavior.interaction
	}
	if behavior.network != "none" && id != "auth.login" && id != "doctor" && id != "theme.import" && !quotaProjectCredentialResolutionCommand(id) {
		authWhen := conditionClause(predicate("runtime_state", "authentication", "requires_human_authorization", nil))
		if interaction.Mode == "none" {
			interaction = InteractionCapability{Mode: "optional", JSONBehavior: "oauth_authorization_returns_interaction"}
			if id == "project.open" {
				interaction.JSONBehavior = "browser_launch_suppressed_and_oauth_authorization_returns_interaction"
			}
		} else {
			interaction.Mode = "optional"
			interaction.JSONBehavior = "declared_conditions_return_interaction"
		}
		interactionWhen = append(interactionWhen, authWhen)
		if behavior.idempotency == "conditional" && !containsPredicateInIdempotency(behavior.idempotencyWhen, "runtime_state", "authentication", "requires_human_authorization") {
			behavior.idempotencyWhen = append([]IdempotencyCondition{{Idempotency: "yes", When: []BehaviorConditionClause{authWhen}}}, behavior.idempotencyWhen...)
		}
	}
	if SupportsStatelessCommand(id) {
		behavior = withStatelessCommandEffects(behavior)
		interactionWhen = withStatelessCommandInteractions(interactionWhen)
	}
	var stdinSchema *string
	if support.Stdin {
		kind := "remote_config"
		switch id {
		case "project.import":
			kind = "remote_config_import"
		case "auth.add.oauth":
			kind = "oauth_credentials"
		case "auth.add.service-account":
			kind = "service_account_credentials"
		case "theme.import":
			kind = "theme"
		case "apply", "plan.show", "plan.validate":
			kind = "publication_plan"
		}
		value := "urn:fbrcm:schema:cli:" + Version + ":stdin:" + kind
		stdinSchema = &value
	}
	capability := Capability{
		ID: id, Path: path, Summary: cmd.Short, Arguments: arguments, Flags: flags,
		InvocationSchema: "urn:fbrcm:schema:cli:" + Version + ":command:" + id + ":input",
		StdinSchema:      stdinSchema, ResponseSchema: SchemaID(id), ErrorSchema: ErrorSchemaID(),
		SideEffectLevel: behavior.level, SideEffects: behavior.effectNames(),
		SideEffectWhen: behavior.sideEffectConditions(), NetworkAccess: behavior.network,
		NetworkWhen: append([]BehaviorConditionClause{}, behavior.networkWhen...), Destructive: behavior.destructive,
		DestructiveWhen: cloneConditions(behavior.destructiveWhen), DestructiveReasons: append([]string{}, behavior.destructiveReasons...), Idempotency: behavior.idempotency,
		IdempotencyWhen: cloneIdempotencyConditions(behavior.idempotencyWhen), Supports: support, StdinModes: stdinModes(id, support.Stdin), Interaction: interaction,
		InteractionWhen: interactionWhen,
	}
	capability.ProblemCodes = CommandProblemCodes(capability)
	return capability
}

func profileOptionIgnored(id string) bool {
	return id == "help" || id == "capabilities" || strings.HasPrefix(id, "schema.") ||
		strings.HasPrefix(id, "plan.") ||
		strings.HasPrefix(id, "config.") || strings.HasPrefix(id, "hooks.") ||
		strings.HasPrefix(id, "projects.aliases.") || strings.HasPrefix(id, "theme")
}

func optionIgnored(id, name string) bool {
	if name == "profile" {
		return profileOptionIgnored(id)
	}
	if id == "auth.login" && name == "noopen" {
		return true
	}
	return id == "config.edit" && slices.Contains([]string{"editor", "full", "scope"}, name)
}

func containsPredicate(conditions []BehaviorConditionClause, source, name, operator string) bool {
	for _, clause := range conditions {
		for _, item := range clause.AllOf {
			if item.Source == source && item.Name == name && item.Operator == operator {
				return true
			}
		}
	}
	return false
}

func containsPredicateInIdempotency(conditions []IdempotencyCondition, source, name, operator string) bool {
	for _, condition := range conditions {
		if containsPredicate(condition.When, source, name, operator) {
			return true
		}
	}
	return false
}

func stdinModes(id string, supported bool) []string {
	if !supported {
		return []string{}
	}
	if id == "theme.import" {
		return []string{"toml_document"}
	}
	return []string{"json_document"}
}

func cloneIdempotencyConditions(items []IdempotencyCondition) []IdempotencyCondition {
	result := make([]IdempotencyCondition, 0, len(items))
	for _, item := range items {
		result = append(result, IdempotencyCondition{Idempotency: item.Idempotency, When: cloneConditions(item.When)})
	}
	return result
}

func predicate(source, name, operator string, value any) BehaviorPredicate {
	return BehaviorPredicate{Source: source, Name: name, Operator: operator, Value: value}
}

func conditionClause(predicates ...BehaviorPredicate) BehaviorConditionClause {
	return BehaviorConditionClause{AllOf: append([]BehaviorPredicate{}, predicates...)}
}

func cloneConditions(conditions []BehaviorConditionClause) []BehaviorConditionClause {
	result := make([]BehaviorConditionClause, 0, len(conditions))
	for _, clause := range conditions {
		result = append(result, conditionClause(clause.AllOf...))
	}
	return result
}

func typedFlagDefault(flag *pflag.Flag) any {
	switch flag.Value.Type() {
	case "bool":
		value, err := strconv.ParseBool(flag.DefValue)
		if err == nil {
			return value
		}
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		value, err := strconv.ParseInt(flag.DefValue, 10, 64)
		if err == nil {
			return value
		}
	case "float32", "float64":
		value, err := strconv.ParseFloat(flag.DefValue, 64)
		if err == nil {
			return value
		}
	case "stringArray", "stringSlice", "intSlice":
		var value any
		if err := json.Unmarshal([]byte(flag.DefValue), &value); err == nil {
			return value
		}
		if flag.Value.Type() != "intSlice" {
			// pflag renders nonempty string slices as bracketed CSV, not JSON.
			if values, err := csv.NewReader(strings.NewReader(strings.TrimSuffix(strings.TrimPrefix(flag.DefValue, "["), "]"))).Read(); err == nil {
				return values
			}
		}
	}
	return flag.DefValue
}

func parseArguments(cmd *cobra.Command) []ArgumentCapability {
	fields := strings.Fields(cmd.Use)
	arguments := make([]ArgumentCapability, 0)
	for _, field := range fields[1:] {
		required := strings.HasPrefix(field, "<")
		optional := strings.HasPrefix(field, "[")
		if !required && !optional {
			continue
		}
		name := strings.Trim(field, "<>[]")
		repeated := strings.HasSuffix(name, "...")
		name = strings.TrimSuffix(name, "...")
		name = strings.Trim(name, "<>")
		arguments = append(arguments, ArgumentCapability{Name: strings.ReplaceAll(name, "-", "_"), Required: required, Repeated: repeated, Schema: "string"})
	}
	if CommandID(cmd) == "help" && len(arguments) == 1 {
		arguments[0].Repeated = true
	}
	return arguments
}

// CommandNetworkAccess returns the reviewed network requirement for a command.
func CommandNetworkAccess(cmd *cobra.Command) string {
	behavior, ok := capabilityBehaviors[CommandID(cmd)]
	if !ok {
		return "conditional"
	}
	return behavior.network
}

func hasFlag(cmd *cobra.Command, name string) bool {
	return cmd.Flags().Lookup(name) != nil || cmd.InheritedFlags().Lookup(name) != nil
}
