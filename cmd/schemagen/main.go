package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yumauri/fbrcm/cli/app"
	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/core/firebase"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

const draft = "https://json-schema.org/draft/2020-12/schema"

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintln(os.Stderr, "schemagen takes no arguments")
		os.Exit(2)
	}
	root := app.NewRootForContract("schema")
	index := contract.Capabilities(root)
	detailed := contract.DetailedCapabilities(root)
	stageRoot, err := os.MkdirTemp(".", ".fbrcm-schemagen-")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(stageRoot) }()
	goldenDir := filepath.Join(stageRoot, "cli", "app", "testdata")
	major, _, _ := strings.Cut(contract.Version, ".")
	goldenName := "contract_v" + major + "_capabilities.golden.json"
	detailedGoldenName := "contract_v" + major + "_capabilities_detailed.golden.json"
	write(goldenDir, goldenName, index)
	write(goldenDir, detailedGoldenName, detailed)
	outDir := filepath.Join(stageRoot, "schemas", "cli", contract.Version)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}
	removeGeneratedSchemas(outDir)
	semantic := semanticSchema()
	extensions, err := newExtensionValidators(semantic)
	if err != nil {
		panic(err)
	}
	writeSchema := func(name string, schema map[string]any) {
		if err := extensions.Validate(schema); err != nil {
			panic(fmt.Errorf("validate extensions in %s: %w", name, err))
		}
		write(outDir, name, schema)
	}
	writeSchema("envelope.schema.json", envelopeSchema())
	writeSchema("error.schema.json", errorSchema())
	writeSchema("capability.schema.json", capabilitySchema(detailed))
	writeSchema("semantic.schema.json", semantic)
	writeSchema("stdin.remote_config.schema.json", remoteConfigStdinSchema("urn:fbrcm:schema:cli:"+contract.Version+":stdin:remote_config"))
	writeSchema("stdin.remote_config_import.schema.json", remoteConfigImportStdinSchema("urn:fbrcm:schema:cli:"+contract.Version+":stdin:remote_config_import"))
	writeSchema("stdin.credentials.schema.json", credentialsSchema("urn:fbrcm:schema:cli:"+contract.Version+":stdin:credentials"))
	writeSchema("stdin.oauth_credentials.schema.json", standaloneSchema("urn:fbrcm:schema:cli:"+contract.Version+":stdin:oauth_credentials", "Google OAuth client credential JSON", oauthCredentialSchema()))
	writeSchema("stdin.service_account_credentials.schema.json", standaloneSchema("urn:fbrcm:schema:cli:"+contract.Version+":stdin:service_account_credentials", "Google service-account credential JSON", serviceAccountCredentialSchema()))
	for _, capability := range detailed {
		command := root
		if capability.ID != "root" {
			var err error
			var remaining []string
			command, remaining, err = root.Find(capability.Path)
			if err != nil || len(remaining) != 0 {
				panic(fmt.Sprintf("find command %s: error=%v remaining=%v", capability.ID, err, remaining))
			}
		}
		dataSchema, err := contract.ResponseDataSchema(command)
		if err != nil {
			panic(err)
		}
		successDataSchema, err := contract.ResponseSuccessDataSchema(command)
		if err != nil {
			panic(err)
		}
		base := strings.ReplaceAll(capability.ID, ".", "_")
		writeSchema(base+".input.schema.json", inputSchema(capability, command))
		writeSchema(base+".response.schema.json", responseSchema(capability, dataSchema, successDataSchema))
	}
	enforceContractLock(stageRoot, outDir, filepath.Join(goldenDir, goldenName), filepath.Join(goldenDir, detailedGoldenName))
	if err := publishGeneratedContract(stageRoot); err != nil {
		panic(err)
	}
}

type contractLock struct {
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Released bool   `json:"released"`
}

func enforceContractLock(stageRoot, schemaDir string, extraFiles ...string) {
	digest, err := generatedContractDigest(stageRoot, schemaDir, extraFiles...)
	if err != nil {
		panic(err)
	}
	lockPath := filepath.Join("schemas", "cli", "contract.lock.json")
	generated := contractLock{Version: contract.Version, SHA256: digest}
	raw, err := os.ReadFile(lockPath)
	if err == nil {
		var previous contractLock
		if decodeErr := json.Unmarshal(raw, &previous); decodeErr != nil {
			panic(fmt.Errorf("decode %s: %w", lockPath, decodeErr))
		}
		if validateErr := validateContractLock(previous, generated); validateErr != nil {
			panic(validateErr)
		}
		generated.Released = previous.Released && previous.Version == generated.Version
	} else if !os.IsNotExist(err) {
		panic(err)
	}
	write(filepath.Join(stageRoot, filepath.Dir(lockPath)), filepath.Base(lockPath), generated)
}

func validateContractLock(previous, generated contractLock) error {
	if previous.Released && previous.Version == generated.Version && previous.SHA256 != generated.SHA256 {
		return fmt.Errorf("CLI contract changed without a version bump from %s; bump cli/contract.Version before regenerating", generated.Version)
	}
	return nil
}

func generatedContractDigest(root, schemaDir string, extraFiles ...string) (string, error) {
	versionDirs, err := os.ReadDir(filepath.Dir(schemaDir))
	if err != nil {
		return "", err
	}
	paths := append([]string(nil), extraFiles...)
	for _, versionDir := range versionDirs {
		if !versionDir.IsDir() {
			continue
		}
		dir := filepath.Join(filepath.Dir(schemaDir), versionDir.Name())
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			return "", readErr
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".schema.json") {
				paths = append(paths, filepath.Join(dir, entry.Name()))
			}
		}
	}
	slices.Sort(paths)
	digest := sha256.New()
	for _, path := range paths {
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", readErr
		}
		logicalPath, relativeErr := filepath.Rel(root, path)
		if relativeErr != nil {
			return "", relativeErr
		}
		_, _ = digest.Write([]byte(filepath.ToSlash(logicalPath)))
		_, _ = digest.Write([]byte{0})
		_, _ = digest.Write(raw)
		_, _ = digest.Write([]byte{0})
	}
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

type publishItem struct {
	staged string
	target string
	backup string
}

func publishGeneratedContract(stageRoot string) error {
	major, _, _ := strings.Cut(contract.Version, ".")
	items := []publishItem{
		{staged: filepath.Join(stageRoot, "schemas", "cli", contract.Version), target: filepath.Join("schemas", "cli", contract.Version)},
		{staged: filepath.Join(stageRoot, "cli", "app", "testdata", "contract_v"+major+"_capabilities.golden.json"), target: filepath.Join("cli", "app", "testdata", "contract_v"+major+"_capabilities.golden.json")},
		{staged: filepath.Join(stageRoot, "cli", "app", "testdata", "contract_v"+major+"_capabilities_detailed.golden.json"), target: filepath.Join("cli", "app", "testdata", "contract_v"+major+"_capabilities_detailed.golden.json")},
		{staged: filepath.Join(stageRoot, "schemas", "cli", "contract.lock.json"), target: filepath.Join("schemas", "cli", "contract.lock.json")},
	}
	return publishTransaction(items)
}

func publishTransaction(items []publishItem) (returnErr error) {
	committed := 0
	defer func() {
		if returnErr == nil {
			for _, item := range items {
				if item.backup != "" {
					_ = os.RemoveAll(item.backup)
				}
			}
			return
		}
		for index := committed - 1; index >= 0; index-- {
			item := items[index]
			_ = os.RemoveAll(item.target)
			if item.backup != "" {
				_ = os.Rename(item.backup, item.target)
			}
		}
		for index := committed; index < len(items); index++ {
			item := items[index]
			if item.backup != "" {
				_ = os.Rename(item.backup, item.target)
			}
		}
	}()

	for index := range items {
		item := &items[index]
		if _, err := os.Stat(item.staged); err != nil {
			return fmt.Errorf("inspect staged contract asset %s: %w", item.staged, err)
		}
		parent := filepath.Dir(item.target)
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return err
		}
		if _, err := os.Lstat(item.target); err == nil {
			reserved, reserveErr := os.MkdirTemp(parent, ".schemagen-backup-")
			if reserveErr != nil {
				return reserveErr
			}
			if removeErr := os.Remove(reserved); removeErr != nil {
				return removeErr
			}
			item.backup = reserved
			if renameErr := os.Rename(item.target, item.backup); renameErr != nil {
				return renameErr
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.Rename(item.staged, item.target); err != nil {
			return fmt.Errorf("publish contract asset %s: %w", item.target, err)
		}
		committed++
	}
	return nil
}

func standaloneSchema(id, description string, schema map[string]any) map[string]any {
	schema["$schema"] = draft
	schema["$id"] = id
	schema["description"] = description
	return schema
}

func removeGeneratedSchemas(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schema.json") {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			panic(err)
		}
	}
}

func write(dir, name string, value any) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(filepath.Join(dir, name), raw, 0o644); err != nil {
		panic(err)
	}
}

func remoteConfigSchema(id string) map[string]any {
	template := contract.SchemaForDTO(firebase.RemoteConfig{})
	cache := map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"required":             []string{"remote_config"},
		"properties": map[string]any{
			"etag":          map[string]any{"type": "string"},
			"cached_at":     map[string]any{"type": "string", "format": "date-time"},
			"remote_config": template,
		},
	}
	return map[string]any{
		"$schema":     draft,
		"$id":         id,
		"description": "Firebase Remote Config template or fbrcm parameters-cache object",
		"oneOf":       []any{template, cache},
	}
}

// remoteConfigStdinSchema models the local JSON decoder rather than Firebase
// validation. Stdin transformations intentionally accept templates that may
// only become Firebase-valid after the requested mutation.
func remoteConfigStdinSchema(id string) map[string]any {
	document := remoteConfigSchema(id)
	document["description"] = "Firebase Remote Config template or fbrcm parameters-cache object accepted by the local stdin decoder"
	document["anyOf"] = document["oneOf"]
	delete(document, "oneOf")
	for _, variant := range document["anyOf"].([]any) {
		object := variant.(map[string]any)
		properties, _ := object["properties"].(map[string]any)
		if nested, ok := properties["remote_config"].(map[string]any); ok {
			loosenRemoteConfigForStdin(nested)
			continue
		}
		loosenRemoteConfigForStdin(object)
	}
	return document
}

func remoteConfigImportStdinSchema(id string) map[string]any {
	document := remoteConfigStdinSchema(id)
	document["description"] = "Firebase Remote Config template or fbrcm parameters-cache object accepted for project import after local Remote Config validation"
	for _, variant := range document["anyOf"].([]any) {
		object := variant.(map[string]any)
		properties, _ := object["properties"].(map[string]any)
		target := object
		if nested, ok := properties["remote_config"].(map[string]any); ok {
			target = nested
		}
		target["x-fbrcm-validation"] = []any{map[string]any{
			"operator": "local_validate", "validator": "firebase.NormalizeRemoteConfigForUpdate",
		}}
	}
	return document
}

func loosenRemoteConfigForStdin(schema map[string]any) {
	delete(schema, "x-fbrcm-validation")
	schema["additionalProperties"] = true
	properties, _ := schema["properties"].(map[string]any)

	if conditions, ok := properties["conditions"].(map[string]any); ok {
		delete(conditions, "uniqueItems")
		delete(conditions, "x-fbrcm-validation")
		if item, ok := conditions["items"].(map[string]any); ok {
			delete(item, "required")
			conditionProperties, _ := item["properties"].(map[string]any)
			for _, name := range []string{"name", "expression", "tagColor"} {
				conditionProperties[name] = map[string]any{"type": "string"}
			}
		}
	}

	for _, name := range []string{"parameters", "parameterGroups"} {
		if entries, ok := properties[name].(map[string]any); ok {
			delete(entries, "propertyNames")
			loosenRemoteConfigMapEntries(entries, name == "parameterGroups")
		}
	}
	if version, ok := properties["version"].(map[string]any); ok {
		loosenJSONStructForStdin(version)
	}
}

func loosenRemoteConfigMapEntries(entries map[string]any, groups bool) {
	entry, _ := entries["additionalProperties"].(map[string]any)
	if entry == nil {
		return
	}
	entry["additionalProperties"] = true
	properties, _ := entry["properties"].(map[string]any)
	for _, name := range []string{"description", "valueType"} {
		if _, ok := properties[name]; ok {
			properties[name] = map[string]any{"type": "string"}
		}
	}
	if groups {
		if parameters, ok := properties["parameters"].(map[string]any); ok {
			delete(parameters, "propertyNames")
			loosenRemoteConfigMapEntries(parameters, false)
		}
	}
}

func loosenJSONStructForStdin(schema map[string]any) {
	schema["additionalProperties"] = true
	properties, _ := schema["properties"].(map[string]any)
	for _, property := range properties {
		child, ok := property.(map[string]any)
		if !ok {
			continue
		}
		delete(child, "format")
		delete(child, "pattern")
		delete(child, "maxLength")
		if child["type"] == "object" || child["properties"] != nil {
			loosenJSONStructForStdin(child)
		}
	}
}

func credentialsSchema(id string) map[string]any {
	return map[string]any{"$schema": draft, "$id": id, "description": "Google OAuth client secret or service-account credential JSON", "anyOf": []any{oauthCredentialSchema(), serviceAccountCredentialSchema()}}
}

func oauthCredentialSchema() map[string]any {
	nonblank := map[string]any{"type": "string", "pattern": `.*\S.*`}
	uri := map[string]any{"type": "string", "x-fbrcm-validation": []any{map[string]any{"operator": "parse_uri", "parser": "net/url.ParseRequestURI", "normalization": "trim_unicode_whitespace", "require_absolute": true}}}
	client := map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"required":             []string{"client_id", "client_secret", "auth_uri", "token_uri", "redirect_uris"},
		"properties": map[string]any{
			"client_id":     maps.Clone(nonblank),
			"client_secret": maps.Clone(nonblank),
			"auth_uri":      maps.Clone(uri),
			"token_uri":     maps.Clone(uri),
			"redirect_uris": map[string]any{"type": "array", "minItems": 1, "items": uri},
		},
	}
	installedSelected := maps.Clone(client)
	installedSelected["required"] = []string{"client_id", "client_secret", "auth_uri", "token_uri"}
	installedProperties := maps.Clone(client["properties"].(map[string]any))
	installedProperties["redirect_uris"] = map[string]any{"type": "array", "items": uri}
	installedSelected["properties"] = installedProperties
	webSelected := map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"required":             []string{"redirect_uris"},
		"properties": map[string]any{
			"redirect_uris": map[string]any{"type": "array", "minItems": 1, "items": uri},
		},
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"oneOf": []any{
			map[string]any{
				"required":   []string{"installed"},
				"properties": map[string]any{"installed": client, "web": map[string]any{"type": "null"}},
			},
			map[string]any{
				"required":   []string{"web"},
				"properties": map[string]any{"installed": map[string]any{"type": "null"}, "web": client},
			},
			map[string]any{
				"required":   []string{"installed", "web"},
				"properties": map[string]any{"installed": installedSelected, "web": webSelected},
			},
		},
	}
}

func serviceAccountCredentialSchema() map[string]any {
	nonblank := map[string]any{"type": "string", "pattern": `.*\S.*`}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"required":             []string{"type", "project_id", "private_key", "client_email", "token_uri"},
		"properties": map[string]any{
			"type":        map[string]any{"const": "service_account"},
			"project_id":  maps.Clone(nonblank),
			"private_key": maps.Clone(nonblank),
			"client_email": map[string]any{
				"type": "string", "x-fbrcm-validation": []any{map[string]any{"operator": "parse_email", "parser": "net/mail.ParseAddress", "require_exact": true}},
			},
			"token_uri": map[string]any{"type": "string", "x-fbrcm-validation": []any{map[string]any{"operator": "parse_uri", "parser": "net/url.ParseRequestURI", "normalization": "trim_unicode_whitespace", "require_absolute": true}}},
		},
	}
}

func envelopeSchema() map[string]any {
	problem := problemObjectSchema("#/$defs/error")
	return map[string]any{
		"$schema": draft, "$id": contract.EnvelopeSchemaID(), "type": "object", "additionalProperties": false,
		"$defs":    map[string]any{"error": problem},
		"required": []string{"schema", "contract_version", "command", "requested_command", "outcome", "exit_code", "producer", "context", "data", "errors", "warnings"},
		"properties": map[string]any{
			"schema": map[string]any{"type": "string"}, "contract_version": map[string]any{"const": contract.Version}, "command": map[string]any{"type": "string", "minLength": 1}, "requested_command": map[string]any{"type": "string", "minLength": 1}, "outcome": map[string]any{"enum": []string{"success", "partial_success", "failure"}}, "exit_code": map[string]any{"type": "integer", "minimum": 0, "maximum": 255},
			"producer": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"name", "version"}, "properties": map[string]any{"name": map[string]any{"const": "fbrcm"}, "version": map[string]any{"type": "string"}}},
			"context":  map[string]any{"type": "object", "additionalProperties": false, "required": []string{"profile", "offline", "dry_run", "draft"}, "properties": map[string]any{"profile": map[string]any{"type": []string{"string", "null"}}, "offline": map[string]any{"type": "boolean"}, "dry_run": map[string]any{"type": "boolean"}, "draft": map[string]any{"type": "boolean"}}},
			"data":     true, "errors": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/$defs/error"}},
			"warnings": map[string]any{"type": "array", "items": warningObjectSchema()},
		},
		"allOf": append([]any{
			map[string]any{"if": map[string]any{"properties": map[string]any{"outcome": map[string]any{"const": "success"}}}, "then": map[string]any{"properties": map[string]any{"exit_code": map[string]any{"enum": []int{0, 1}}, "data": map[string]any{"not": map[string]any{"type": "null"}}, "errors": map[string]any{"maxItems": 0}}}},
			map[string]any{"if": map[string]any{"properties": map[string]any{"outcome": map[string]any{"const": "partial_success"}}}, "then": map[string]any{"properties": map[string]any{"exit_code": map[string]any{"const": 12}, "data": map[string]any{"not": map[string]any{"type": "null"}}, "errors": map[string]any{"minItems": 1}}}},
			map[string]any{"if": map[string]any{"properties": map[string]any{"outcome": map[string]any{"const": "failure"}}}, "then": map[string]any{"properties": map[string]any{"exit_code": map[string]any{"not": map[string]any{"const": 0}}, "errors": map[string]any{"minItems": 1}}}},
		}, failureCategoryStatusConstraints()...),
	}
}

func warningObjectSchema() map[string]any {
	object := func(required []string, properties map[string]any) map[string]any {
		return map[string]any{"type": "object", "additionalProperties": false, "required": required, "properties": properties}
	}
	detailsByCode := map[string]map[string]any{
		"cache.stale":                          object([]string{"source"}, map[string]any{"source": map[string]any{"type": "string", "minLength": 1}}),
		"publication.non_atomic":               object([]string{"target_count"}, map[string]any{"target_count": map[string]any{"type": "integer", "minimum": 2}}),
		"publication.cache_stale":              object([]string{"stage"}, map[string]any{"stage": map[string]any{"const": "cache"}}),
		"publication.draft_cleanup_failed":     object([]string{"stage"}, map[string]any{"stage": map[string]any{"const": "cleanup"}}),
		"publication.post_publish_hook_failed": object([]string{"stage"}, map[string]any{"stage": map[string]any{"const": "post_publish_hook"}}),
	}
	constraints := make([]any, 0, len(detailsByCode))
	for _, code := range contract.KnownWarningCodes() {
		details, exists := detailsByCode[code]
		if !exists {
			panic("missing warning-details schema for " + code)
		}
		constraints = append(constraints, map[string]any{
			"if":   map[string]any{"properties": map[string]any{"code": map[string]any{"const": code}}, "required": []string{"code"}},
			"then": map[string]any{"properties": map[string]any{"details": details}},
		})
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"code", "message", "target", "details", "remediation"},
		"properties": map[string]any{
			"code":        map[string]any{"type": "string", "pattern": `^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)+$`, "enum": contract.KnownWarningCodes()},
			"message":     map[string]any{"type": "string", "maxLength": 4097},
			"target":      map[string]any{"type": []string{"string", "null"}},
			"details":     map[string]any{"oneOf": []any{map[string]any{"type": "null"}, map[string]any{"type": "object"}}},
			"remediation": remediationSchema(),
		},
		"allOf": constraints,
	}
}

func failureCategoryStatusConstraints() []any {
	statuses := []struct {
		status     int
		categories []string
	}{
		{2, []string{"argument"}}, {3, []string{"configuration", "profile"}}, {4, []string{"auth"}}, {5, []string{"permission"}}, {6, []string{"not_found"}},
		{7, []string{"conflict"}}, {8, []string{"validation"}}, {9, []string{"timeout"}}, {10, []string{"interaction"}}, {11, []string{"unavailable"}},
		{12, []string{"partial_success"}}, {13, []string{"io"}}, {14, []string{"hook"}}, {15, []string{"internal"}}, {130, []string{"canceled"}},
	}
	result := make([]any, 0, len(statuses))
	for _, item := range statuses {
		allowed := []int{item.status}
		if item.status == 8 {
			allowed = append(allowed, 1)
		}
		result = append(result, map[string]any{
			"if": map[string]any{
				"properties": map[string]any{
					"outcome": map[string]any{"const": "failure"},
					"errors":  map[string]any{"prefixItems": []any{map[string]any{"properties": map[string]any{"category": map[string]any{"enum": item.categories}}, "required": []string{"category"}}}},
				},
				"required": []string{"outcome", "errors"},
			},
			"then": map[string]any{"properties": map[string]any{"exit_code": map[string]any{"enum": allowed}}},
		})
	}
	return result
}

func errorSchema() map[string]any {
	result := problemObjectSchema("#")
	result["$schema"] = draft
	result["$id"] = contract.ErrorSchemaID()
	return result
}

func problemObjectSchema(selfRef string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"code", "category", "message", "retryable", "target", "stage", "details", "remediation"},
		"properties": map[string]any{
			"code":        map[string]any{"type": "string", "pattern": `^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)+$`, "enum": contract.KnownProblemCodes()},
			"category":    map[string]any{"enum": []string{"argument", "configuration", "profile", "auth", "permission", "not_found", "conflict", "validation", "timeout", "interaction", "unavailable", "partial_success", "io", "hook", "canceled", "internal"}},
			"message":     map[string]any{"type": "string", "maxLength": 4097},
			"retryable":   map[string]any{"type": "boolean"},
			"target":      map[string]any{"type": []string{"string", "null"}},
			"stage":       map[string]any{"type": []string{"string", "null"}},
			"details":     problemDetailsSchema(selfRef),
			"remediation": remediationSchema(),
		},
		"allOf": problemCodeConstraints(),
	}
}

func problemCodeConstraints() []any {
	byCategory := map[string][]string{
		"argument":        {"argument.invalid", "argument.unknown_command", "auth.id_invalid", "command.not_executable", "condition.ambiguous", "draft.ambiguous", "parameter.ambiguous", "project.ambiguous"},
		"configuration":   {"configuration.invalid", "configuration.local_disabled", "configuration.local_not_found", "configuration.project_aliases_invalid", "hooks.not_configured"},
		"profile":         {"profile.invalid"},
		"auth":            {"auth.configuration_invalid", "auth.credentials_invalid", "auth.setup_required"},
		"permission":      {"firebase.permission_denied", "filesystem.permission_denied"},
		"not_found":       {"auth.not_found", "command.not_found", "condition.not_found", "draft.not_found", "group.not_found", "parameter.not_found", "parameters_cache.not_found", "personalization.not_found", "profile.not_found", "project.not_found", "resource.not_found", "schema.not_found", "version.not_found"},
		"conflict":        {"draft.exists", "hooks.changed", "parameter.exists", "profile.conflict", "project_alias.conflict", "project_alias.read_only", "remote_config.conflict", "resource.conflict"},
		"validation":      {"condition.invalid", "diagnostic.failed", "expression.invalid", "firebase.request_failed", "remote_config.invalid", "remote_config.validation_failed", "result.unsuccessful", "stdin.remote_config.invalid", "validation.failed"},
		"timeout":         {"command.timeout", "firebase.timeout", "network.timeout"},
		"interaction":     {"interaction.required"},
		"unavailable":     {"firebase.rate_limited", "firebase.service_unavailable", "network.offline", "network.unavailable"},
		"partial_success": {"batch.partial_success", "publication.cache_failed", "publication.hook_failed"},
		"io":              {"file.io_failed"},
		"hook":            {"hook.failed"},
		"canceled":        {"command.canceled"},
		"internal":        {"internal.contract_violation", "internal.unclassified"},
	}
	result := make([]any, 0, len(byCategory)+5)
	categories := make([]string, 0, len(byCategory))
	for category := range byCategory {
		categories = append(categories, category)
	}
	slices.Sort(categories)
	for _, category := range categories {
		codes := byCategory[category]
		result = append(result, map[string]any{
			"if":   map[string]any{"properties": map[string]any{"code": map[string]any{"enum": codes}}, "required": []string{"code"}},
			"then": map[string]any{"properties": map[string]any{"category": map[string]any{"const": category}}},
		})
	}
	result = append(result, map[string]any{
		"if":   map[string]any{"properties": map[string]any{"code": map[string]any{"enum": []string{"command.timeout", "firebase.rate_limited", "firebase.service_unavailable", "firebase.timeout", "hooks.changed", "network.offline", "network.timeout", "network.unavailable"}}}, "required": []string{"code"}},
		"then": map[string]any{"properties": map[string]any{"retryable": map[string]any{"const": true}}},
	})
	for _, rule := range []struct {
		codes []string
		kind  []string
	}{
		{[]string{"argument.unknown_command"}, []string{"invocation"}},
		{[]string{"interaction.required"}, []string{"interaction", "oauth_authorization"}},
		{[]string{"expression.invalid"}, []string{"expression"}},
		{[]string{"hook.failed"}, []string{"hook"}},
		{[]string{"firebase.permission_denied", "firebase.rate_limited", "firebase.request_failed", "firebase.service_unavailable", "firebase.timeout"}, []string{"remote_api"}},
		{[]string{"resource.not_found"}, []string{"remote_api", "selection"}},
		{[]string{"condition.ambiguous", "condition.not_found", "draft.ambiguous", "draft.not_found", "group.not_found", "parameter.ambiguous", "parameter.not_found", "parameters_cache.not_found", "personalization.not_found", "profile.not_found", "project.ambiguous", "project.not_found", "version.not_found"}, []string{"selection"}},
		{[]string{"configuration.invalid", "configuration.local_disabled", "configuration.local_not_found", "configuration.project_aliases_invalid", "hooks.not_configured", "profile.invalid", "condition.invalid", "remote_config.invalid", "remote_config.validation_failed", "stdin.remote_config.invalid"}, []string{"validation"}},
		{[]string{"auth.credentials_invalid"}, []string{"remote_api", "validation"}},
		{[]string{"auth.configuration_invalid", "auth.id_invalid", "auth.not_found", "auth.setup_required"}, []string{"auth"}},
		{[]string{"draft.exists", "hooks.changed", "parameter.exists", "profile.conflict", "project_alias.conflict", "project_alias.read_only", "resource.conflict"}, []string{"conflict"}},
		{[]string{"remote_config.conflict"}, []string{"conflict", "remote_api"}},
		{[]string{"batch.failed", "batch.partial_success"}, []string{"batch"}},
		{[]string{"file.io_failed", "filesystem.permission_denied"}, []string{"file"}},
	} {
		result = append(result, map[string]any{
			"if":   map[string]any{"properties": map[string]any{"code": map[string]any{"enum": rule.codes}}, "required": []string{"code"}},
			"then": map[string]any{"properties": map[string]any{"details": map[string]any{"type": "object", "properties": map[string]any{"kind": map[string]any{"enum": rule.kind}}, "required": []string{"kind"}}}},
		})
	}
	result = append(result, map[string]any{
		"if": map[string]any{
			"properties": map[string]any{"code": map[string]any{"const": "remote_config.validation_failed"}},
			"required":   []string{"code"},
		},
		"then": map[string]any{
			"properties": map[string]any{
				"details": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"kind":   map[string]any{"const": "validation"},
						"source": map[string]any{"enum": []string{"local", "firebase"}},
					},
					"required": []string{"kind", "source"},
				},
			},
		},
	})
	result = append(result, map[string]any{
		"if": map[string]any{
			"properties": map[string]any{"code": map[string]any{"enum": []string{
				"argument.invalid", "argument.unknown_command", "auth.configuration_invalid", "auth.credentials_invalid", "auth.id_invalid", "auth.not_found", "auth.setup_required",
				"command.canceled", "command.not_executable", "command.not_found", "condition.ambiguous", "condition.invalid", "condition.not_found",
				"configuration.invalid", "configuration.local_disabled", "configuration.local_not_found", "configuration.project_aliases_invalid", "diagnostic.failed",
				"draft.ambiguous", "draft.exists", "draft.not_found", "expression.invalid", "file.io_failed", "filesystem.permission_denied", "firebase.permission_denied", "firebase.request_failed",
				"group.not_found", "hooks.not_configured", "interaction.required", "internal.contract_violation", "internal.unclassified", "parameter.ambiguous", "parameter.exists", "parameter.not_found",
				"parameters_cache.not_found", "personalization.not_found", "profile.conflict", "profile.invalid", "profile.not_found", "project.ambiguous", "project.not_found", "project_alias.conflict", "project_alias.read_only",
				"publication.cache_failed", "publication.hook_failed", "remote_config.invalid", "remote_config.validation_failed", "resource.conflict", "resource.not_found",
				"result.unsuccessful", "schema.not_found", "stdin.remote_config.invalid", "validation.failed", "version.not_found",
			}}},
			"required": []string{"code"},
		},
		"then": map[string]any{"properties": map[string]any{"retryable": map[string]any{"const": false}}},
	})
	return result
}

func problemDetailsSchema(selfRef string) map[string]any {
	object := func(required []string, properties map[string]any) map[string]any {
		return map[string]any{"type": "object", "additionalProperties": false, "required": required, "properties": properties}
	}
	stringValue := map[string]any{"type": "string"}
	candidate := object([]string{"name", "id"}, map[string]any{"name": stringValue, "id": stringValue})
	failure := map[string]any{"$ref": selfRef}
	return map[string]any{"oneOf": []any{
		map[string]any{"type": "null"},
		object([]string{"kind", "interaction_type", "required_option", "destructive"}, map[string]any{"kind": map[string]any{"const": "interaction"}, "interaction_type": map[string]any{"enum": []string{"confirmation", "destination_conflict", "external_input", "input_required", "selection_required"}}, "required_option": map[string]any{"type": []string{"string", "null"}}, "destructive": map[string]any{"type": "boolean"}}),
		object([]string{"kind", "auth_id"}, map[string]any{"kind": map[string]any{"const": "oauth_authorization"}, "auth_id": stringValue}),
		object([]string{"kind", "expression", "context"}, map[string]any{"kind": map[string]any{"const": "expression"}, "expression": stringValue, "context": stringValue}),
		object([]string{"kind", "source"}, map[string]any{"kind": map[string]any{"const": "validation"}, "source": stringValue}),
		object([]string{"kind"}, map[string]any{"kind": map[string]any{"const": "auth"}}),
		object([]string{"kind", "resource"}, map[string]any{"kind": map[string]any{"const": "conflict"}, "resource": stringValue}),
		object([]string{"kind", "resource", "query", "candidates"}, map[string]any{"kind": map[string]any{"const": "selection"}, "resource": stringValue, "query": stringValue, "candidates": map[string]any{"type": "array", "items": candidate}}),
		object([]string{"kind", "requested_command", "resolved_command"}, map[string]any{"kind": map[string]any{"const": "invocation"}, "requested_command": stringValue, "resolved_command": stringValue}),
		object([]string{"kind", "operation", "failed_targets", "failures", "successful_target_count", "published_target_count"}, map[string]any{"kind": map[string]any{"const": "batch"}, "operation": stringValue, "failed_targets": map[string]any{"type": "array", "items": stringValue}, "failures": map[string]any{"type": "array", "items": failure}, "successful_target_count": map[string]any{"type": "integer", "minimum": 0}, "published_target_count": map[string]any{"type": "integer", "minimum": 0}}),
		object([]string{"kind", "event", "index", "exit_code", "timed_out", "output"}, map[string]any{"kind": map[string]any{"const": "hook"}, "event": stringValue, "index": map[string]any{"type": "integer", "minimum": 0}, "exit_code": map[string]any{"type": "integer"}, "timed_out": map[string]any{"type": "boolean"}, "output": stringValue}),
		object([]string{"kind", "service", "operation", "http_status", "remote_status", "remote_code", "retry_after_ms"}, map[string]any{"kind": map[string]any{"const": "remote_api"}, "service": stringValue, "operation": stringValue, "http_status": map[string]any{"type": "integer", "minimum": 100, "maximum": 599}, "remote_status": stringValue, "remote_code": stringValue, "retry_after_ms": map[string]any{"type": "integer", "minimum": 0}}),
		object([]string{"kind", "operation", "path"}, map[string]any{"kind": map[string]any{"const": "file"}, "operation": stringValue, "path": stringValue}),
	}}
}

func capabilitySchema(published []contract.Capability) map[string]any {
	stringArray := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	scalar := map[string]any{"type": []string{"string", "boolean", "integer", "number"}}
	effectValue := map[string]any{"enum": []string{"local_state_write", "local_file_write", "local_file_delete", "local_cache_write", "local_cache_move", "local_cache_delete", "local_draft_write", "local_draft_delete", "authentication_remote_access", "firebase_remote_read", "firebase_remote_validation", "firebase_remote_write", "firebase_managed_feature_delete", "trusted_hook_execution"}}
	predicate := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"source", "name", "operator", "value"},
		"properties": map[string]any{
			"source":   map[string]any{"enum": []string{"argument", "option", "stdin", "context", "runtime_state"}},
			"name":     map[string]any{"type": "string", "minLength": 1},
			"operator": map[string]any{"enum": []string{"equals", "absent", "present", "not_usable", "configured_for_event", "executed", "not_executed", "conflicts", "write_authorized", "write_succeeded", "delete_succeeded", "available", "has_changes", "has_no_changes", "is_destructive", "accepted", "cache_write_succeeded", "sync_required", "sync_write_succeeded", "succeeded", "requires_network", "requires_human_authorization", "credentials_reused", "token_persisted", "required", "authorized_or_not_required", "persisted"}},
			"value":    map[string]any{"oneOf": []any{scalar, map[string]any{"type": "null"}}},
		},
		"allOf": []any{
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"operator": map[string]any{"const": "equals"}}, "required": []string{"operator"}},
				"then": map[string]any{"properties": map[string]any{"value": scalar}},
				"else": map[string]any{"properties": map[string]any{"value": map[string]any{"type": "null"}}},
			},
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}}, "required": []string{"source"}},
				"then": map[string]any{"properties": map[string]any{"name": map[string]any{"enum": []string{"required_cache", "remote_read", "trusted_hook", "output_destination", "credential_file", "mutation_plan", "publication", "authentication", "version_request", "external_editor", "promotion_selection", "confirmation", "profile_bootstrap", "profile_cache", "project_registry", "import_strategy", "import_merge_resolution", "draft_change_note", "diagnostic_cache_probe", "diagnostic_identity"}}}},
			},
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "context"}}, "required": []string{"source"}},
				"then": map[string]any{"properties": map[string]any{"name": map[string]any{"enum": []string{"offline"}}}},
			},
		},
	}
	rules := predicate["allOf"].([]any)
	for _, state := range []struct{ name, operator string }{
		{"required_cache", "not_usable"},
		{"publication", "accepted"},
		{"external_editor", "required"},
		{"promotion_selection", "required"},
		{"profile_bootstrap", "required"},
		{"profile_cache", "available"},
		{"import_strategy", "required"},
		{"import_merge_resolution", "required"},
		{"draft_change_note", "persisted"},
		{"diagnostic_identity", "available"},
	} {
		rules = append(rules, map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": state.name}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"const": state.operator}, "value": map[string]any{"type": "null"}}},
		})
	}
	rules = append(rules,
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "credential_file"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"const": "write_succeeded"}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "trusted_hook"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"enum": []string{"configured_for_event", "executed", "not_executed"}}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "confirmation"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"enum": []string{"required", "authorized_or_not_required"}}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "project_registry"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"enum": []string{"sync_required", "sync_write_succeeded"}}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "diagnostic_cache_probe"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"enum": []string{"write_succeeded", "delete_succeeded"}}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "mutation_plan"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"enum": []string{"has_changes", "has_no_changes", "is_destructive"}}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "remote_read"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"enum": []string{"cache_write_succeeded", "succeeded"}}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "authentication"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"enum": []string{"requires_network", "requires_human_authorization", "credentials_reused", "token_persisted"}}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "version_request"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"const": "requires_network"}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "runtime_state"}, "name": map[string]any{"const": "output_destination"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"enum": []string{"conflicts", "write_authorized"}}, "value": map[string]any{"type": "null"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "context"}, "name": map[string]any{"const": "offline"}}, "required": []string{"source", "name"}},
			"then": map[string]any{"properties": map[string]any{"operator": map[string]any{"const": "equals"}, "value": map[string]any{"type": "boolean"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"source": map[string]any{"const": "stdin"}}, "required": []string{"source"}},
			"then": map[string]any{"properties": map[string]any{"name": map[string]any{"const": "document"}, "operator": map[string]any{"enum": []string{"absent", "present"}}, "value": map[string]any{"type": "null"}}},
		},
	)
	predicate["allOf"] = rules
	conditionClauses := map[string]any{
		"type": "array",
		"items": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"all_of"},
			"properties": map[string]any{
				"all_of": map[string]any{"type": "array", "minItems": 1, "items": predicate},
			},
		},
	}
	flag := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"name", "aliases", "type", "default", "required", "repeatable", "effective", "usage"},
		"properties": map[string]any{
			"name":           map[string]any{"type": "string"},
			"aliases":        stringArray,
			"type":           map[string]any{"enum": []string{"bool", "duration", "int", "string", "stringArray", "stringSlice"}},
			"default":        map[string]any{"oneOf": []any{scalar, map[string]any{"type": "array"}}},
			"required":       map[string]any{"type": "boolean"},
			"repeatable":     map[string]any{"type": "boolean"},
			"effective":      map[string]any{"type": "boolean"},
			"effective_when": conditionClauses,
			"usage":          map[string]any{"type": "string"},
		},
		"allOf": []any{
			map[string]any{"if": map[string]any{"properties": map[string]any{"type": map[string]any{"const": "bool"}}}, "then": map[string]any{"properties": map[string]any{"default": map[string]any{"type": "boolean"}}}},
			map[string]any{"if": map[string]any{"properties": map[string]any{"type": map[string]any{"const": "int"}}}, "then": map[string]any{"properties": map[string]any{"default": map[string]any{"type": "integer"}}}},
			map[string]any{"if": map[string]any{"properties": map[string]any{"type": map[string]any{"enum": []string{"stringArray", "stringSlice"}}}}, "then": map[string]any{"properties": map[string]any{"default": map[string]any{"type": "array"}}}},
			map[string]any{"if": map[string]any{"properties": map[string]any{"type": map[string]any{"enum": []string{"duration", "string"}}}}, "then": map[string]any{"properties": map[string]any{"default": map[string]any{"type": "string"}}}},
		},
	}
	publishedRecords := make([]any, 0, len(published))
	for _, capability := range published {
		publishedRecords = append(publishedRecords, map[string]any{"const": capability})
	}
	return map[string]any{
		"$schema": draft, "$id": contract.CapabilitySchemaID(), "type": "object", "additionalProperties": false,
		"x-fbrcm-runtime-state-semantics": runtimeStatePredicateSemantics(),
		"x-fbrcm-side-effect-semantics":   sideEffectSemantics(),
		"oneOf":                           publishedRecords,
		"required":                        []string{"id", "path", "summary", "arguments", "flags", "invocation_schema", "stdin_schema", "response_schema", "error_schema", "side_effect_level", "side_effects", "side_effect_when", "network_access", "network_when", "destructive", "destructive_when", "destructive_reasons", "idempotency", "idempotency_when", "supports", "stdin_modes", "interaction", "interaction_when"},
		"properties": map[string]any{
			"id": map[string]any{"type": "string", "pattern": `^(?:root|[a-z][a-z0-9-]*(?:\.[a-z][a-z0-9-]*)*)$`}, "path": map[string]any{"type": "array", "items": map[string]any{"type": "string", "pattern": `^[a-z][a-z0-9-]*$`}}, "summary": map[string]any{"type": "string"},
			"arguments":         map[string]any{"type": "array", "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"name", "required", "repeated", "schema"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "required": map[string]any{"type": "boolean"}, "repeated": map[string]any{"type": "boolean"}, "schema": map[string]any{"type": "string"}}}},
			"flags":             map[string]any{"type": "array", "items": flag},
			"invocation_schema": map[string]any{"type": "string", "pattern": `^urn:fbrcm:schema:cli:` + regexp.QuoteMeta(contract.Version) + `:command:[a-z0-9.-]+:input$`}, "stdin_schema": map[string]any{"type": []string{"string", "null"}, "pattern": `^urn:fbrcm:schema:cli:` + regexp.QuoteMeta(contract.Version) + `:stdin:[a-z0-9_]+$`}, "response_schema": map[string]any{"type": "string", "pattern": `^urn:fbrcm:schema:cli:` + regexp.QuoteMeta(contract.Version) + `:command:[a-z0-9.-]+:response$`}, "error_schema": map[string]any{"const": contract.ErrorSchemaID()}, "side_effect_level": map[string]any{"type": "integer", "minimum": 0, "maximum": 3}, "side_effects": map[string]any{"type": "array", "uniqueItems": true, "items": effectValue},
			"side_effect_when": map[string]any{"type": "array", "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"effect", "when"}, "properties": map[string]any{"effect": effectValue, "when": conditionClauses}}},
			"network_access":   map[string]any{"enum": []string{"none", "conditional", "required"}}, "network_when": conditionClauses,
			"destructive": map[string]any{"type": "boolean"}, "destructive_when": conditionClauses, "destructive_reasons": stringArray, "idempotency": map[string]any{"enum": []string{"yes", "conditional", "no"}},
			"idempotency_when": map[string]any{"type": "array", "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"idempotency", "when"}, "properties": map[string]any{"idempotency": map[string]any{"enum": []string{"yes", "no"}}, "when": conditionClauses}}},
			"supports":         contract.CapabilitySupportSchema(),
			"stdin_modes":      map[string]any{"type": "array", "uniqueItems": true, "items": map[string]any{"const": "json_document"}},
			"interaction":      map[string]any{"type": "object", "additionalProperties": false, "required": []string{"mode", "json_behavior"}, "properties": map[string]any{"mode": map[string]any{"enum": []string{"none", "optional", "required"}}, "json_behavior": map[string]any{"enum": []string{"non_interactive", "confirmation_required_without_bypass", "destination_conflict_returns_interaction", "external_input_returns_interaction", "browser_launch_suppressed", "oauth_authorization_returns_interaction", "browser_launch_suppressed_and_oauth_authorization_returns_interaction", "declared_conditions_return_interaction", "missing_input_returns_interaction", "deletion_preview_requires_confirmation", "import_requires_explicit_strategy_and_confirmation", "promotion_requires_explicit_selection_and_confirmation"}}}},
			"interaction_when": conditionClauses,
		},
		"allOf": []any{
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"network_access": map[string]any{"const": "conditional"}}, "required": []string{"network_access"}},
				"then": map[string]any{"properties": map[string]any{"network_when": map[string]any{"minItems": 1}}},
				"else": map[string]any{"properties": map[string]any{"network_when": map[string]any{"maxItems": 0}}},
			},
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"idempotency": map[string]any{"const": "conditional"}}, "required": []string{"idempotency"}},
				"then": map[string]any{"properties": map[string]any{"idempotency_when": map[string]any{"minItems": 1}}},
				"else": map[string]any{"properties": map[string]any{"idempotency_when": map[string]any{"maxItems": 0}}},
			},
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"interaction": map[string]any{"properties": map[string]any{"mode": map[string]any{"const": "none"}}, "required": []string{"mode"}}}, "required": []string{"interaction"}},
				"then": map[string]any{"properties": map[string]any{"interaction_when": map[string]any{"maxItems": 0}}},
				"else": map[string]any{"properties": map[string]any{"interaction_when": map[string]any{"minItems": 1}}},
			},
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"supports": map[string]any{"properties": map[string]any{"stdin": map[string]any{"const": true}}, "required": []string{"stdin"}}}, "required": []string{"supports"}},
				"then": map[string]any{"properties": map[string]any{"stdin_schema": map[string]any{"type": "string"}, "stdin_modes": map[string]any{"minItems": 1}}},
				"else": map[string]any{"properties": map[string]any{"stdin_schema": map[string]any{"type": "null"}, "stdin_modes": map[string]any{"maxItems": 0}}},
			},
			map[string]any{
				"if":   map[string]any{"properties": map[string]any{"destructive": map[string]any{"const": true}}, "required": []string{"destructive"}},
				"then": map[string]any{"properties": map[string]any{"destructive_when": map[string]any{"minItems": 1}, "destructive_reasons": map[string]any{"minItems": 1}}},
				"else": map[string]any{"properties": map[string]any{"destructive_when": map[string]any{"maxItems": 0}, "destructive_reasons": map[string]any{"maxItems": 0}}},
			},
		},
	}
}

func runtimeStatePredicateSemantics() []any {
	definition := func(name, operator, semantics string) any {
		return map[string]any{"source": "runtime_state", "name": name, "operator": operator, "semantics": semantics}
	}
	return []any{
		definition("required_cache", "not_usable", "At the read decision point, the cache or registry required by the command is absent, unreadable, invalid, empty when an empty registry cannot satisfy selection, or expired under the active cache policy. A stale value used only after a failed refresh still satisfies this predicate."),
		definition("remote_read", "cache_write_succeeded", "The command completed a Firebase read and successfully persisted the returned Remote Config data or immutable version snapshot in its command-specific local cache."),
		definition("remote_read", "succeeded", "The command completed the Firebase read successfully, independently of any subsequent local persistence."),
		definition("trusted_hook", "configured_for_event", "At least one locally trusted hook command is configured for the publication event reached by this invocation."),
		definition("trusted_hook", "executed", "At least one locally trusted hook command was executed during this invocation, including a pre_publish hook reached during dry-run."),
		definition("trusted_hook", "not_executed", "No locally trusted hook command was executed during this invocation."),
		definition("output_destination", "conflicts", "The requested destination already exists at the command's pre-write conflict check and the invocation has not supplied the command's overwrite authorization."),
		definition("output_destination", "write_authorized", "The command reached its destination write and either no conflicting destination existed or overwrite was explicitly authorized."),
		definition("mutation_plan", "has_changes", "The completed command-specific plan differs from its comparison baseline and would perform the declared mutation when not previewed."),
		definition("mutation_plan", "has_no_changes", "The completed command-specific plan is equal to its comparison baseline and requires no content mutation."),
		definition("mutation_plan", "is_destructive", "The completed plan contains at least one removal, replacement, overwrite, or other command-specific destructive operation described by destructive_reasons."),
		definition("publication", "accepted", "Firebase accepted the publication before any subsequent hook, cache, or local cleanup step."),
		definition("authentication", "requires_network", "The selected authentication flow cannot reuse sufficient local credentials and must contact its identity provider or metadata service."),
		definition("authentication", "requires_human_authorization", "The selected OAuth identity has no reusable authorization and requires a human browser/device authorization step, which JSON mode suppresses."),
		definition("authentication", "credentials_reused", "Authentication completed using existing locally available credentials without persisting a new token."),
		definition("authentication", "token_persisted", "Authentication obtained and successfully persisted a new or refreshed local token."),
		definition("version_request", "requires_network", "After applying the selector and --cached policy, at least one requested immutable version cannot be resolved from local snapshots and must be fetched from Firebase."),
		definition("external_editor", "required", "Completing the requested command requires launching an external editor, which JSON mode does not launch."),
		definition("promotion_selection", "required", "The promotion plan contains eligible items but the invocation supplies neither an explicit non-interactive selection nor an authorization that selects them."),
		definition("confirmation", "required", "The completed plan requires confirmation and the invocation has not supplied the command's confirmation-bypass option."),
		definition("confirmation", "authorized_or_not_required", "The command reached a confirmation-guarded effect because the invocation supplied the bypass option or the completed plan did not require confirmation."),
		definition("profile_bootstrap", "required", "No explicit or persisted effective profile can provide envelope context, so final envelope construction attempts to create the default profile directories and global configuration."),
		definition("profile_cache", "available", "The old profile has a local cache directory and the destination profile has no cache directory, so profile rename moves the old cache tree."),
		definition("project_registry", "sync_required", "A non-cached project resolution reached a missing or empty live project registry and therefore had to contact Firebase before resolving the requested project."),
		definition("project_registry", "sync_write_succeeded", "A missing or empty live project registry was synchronized from Firebase and the synchronized registry was successfully persisted locally."),
		definition("import_strategy", "required", "The import target contains current Remote Config content and the invocation supplies neither --merge nor --override, so JSON mode must request an explicit strategy."),
		definition("import_merge_resolution", "required", "An explicitly selected merge produced at least one unresolved conflict and the invocation supplies no --merge-resolve value."),
		definition("draft_change_note", "persisted", "draft publish received an explicit change note and successfully persisted it to the local draft before continuing publication."),
		definition("diagnostic_cache_probe", "write_succeeded", "Doctor successfully created its temporary cache probe file."),
		definition("diagnostic_cache_probe", "delete_succeeded", "Doctor successfully removed the temporary cache probe file it created."),
		definition("diagnostic_identity", "available", "Doctor resolved a locally usable configured identity with which it can attempt its Firebase diagnostic read."),
		definition("credential_file", "write_succeeded", "The command successfully created or replaced the OAuth client-secret or service-account credential file."),
	}
}

func sideEffectSemantics() map[string]any {
	return map[string]any{
		"local_state_write":               "Creates or changes local application configuration or registry state.",
		"local_file_write":                "Creates or replaces a user-visible local file outside the command-specific cache and draft stores.",
		"local_file_delete":               "Deletes a local configuration, credential, registry, or user-visible file.",
		"local_cache_write":               "Creates or updates command-specific cached Remote Config data or version snapshots.",
		"local_cache_move":                "Moves an existing local cache tree to a different application-managed path without changing its cached content.",
		"local_cache_delete":              "Deletes command-specific cached Remote Config data or version snapshots.",
		"local_draft_write":               "Creates or updates a local Remote Config draft.",
		"local_draft_delete":              "Deletes a local Remote Config draft.",
		"authentication_remote_access":    "Contacts an identity provider or metadata service for authentication.",
		"firebase_remote_read":            "Reads Firebase project, Remote Config, version, or managed-feature state.",
		"firebase_remote_validation":      "Requests Firebase validation without publishing the candidate template.",
		"firebase_remote_write":           "Publishes or rolls back Firebase Remote Config.",
		"firebase_managed_feature_delete": "Deletes a Firebase Remote Config experiment or rollout.",
		"trusted_hook_execution":          "Executes locally configured hook commands that passed the trust check.",
	}
}

func inputSchema(capability contract.Capability, command *cobra.Command) map[string]any {
	definitions := semanticDefinitions(capability.ID)
	arguments := map[string]any{}
	required := make([]string, 0)
	for _, argument := range capability.Arguments {
		argumentSchema := argumentSchema(capability.ID, argument.Name)
		if argument.Repeated {
			argumentSchema = map[string]any{"type": "array", "items": argumentSchema, "minItems": 1}
		}
		if slices.Contains([]string{"capabilities", "help"}, capability.ID) && argument.Name == "command" {
			argumentSchema["x-fbrcm-grammar"] = "exact executable command argv path components"
			rule := map[string]any{
				"operator": "command_path_resolution", "comparison": "exact_case_sensitive_argv_components",
				"unknown_result": "command.not_found",
			}
			if capability.ID == "capabilities" {
				rule["candidate_source"] = "executable_commands"
				rule["omitted_result"] = "capability_index"
				rule["reserved_root_token"] = "root"
				rule["non_executable_result"] = "command.not_executable"
				addMatchingRule(argumentSchema, rule)
			} else {
				argumentSchema["x-fbrcm-grammar"] = "longest exact existing command argv path prefix; unmatched suffix components are ignored"
				addMatchingRule(argumentSchema, map[string]any{"operator": "help_path_resolution"})
			}
		}
		arguments[argument.Name] = argumentSchema
		if argument.Required {
			required = append(required, argument.Name)
		}
	}
	options := map[string]any{}
	requiredOptions := make([]string, 0)
	for _, flag := range capability.Flags {
		name := strings.TrimPrefix(flag.Name, "--")
		options[name] = flagSchema(capability.ID, flag)
		if flag.Required {
			requiredOptions = append(requiredOptions, name)
		}
	}
	stdin := map[string]any{"type": "null"}
	switch {
	case slices.Contains([]string{"get", "add", "update", "delete", "project.import"}, capability.ID):
		remote := remoteConfigStdinSchema("")
		if capability.ID == "project.import" {
			remote = remoteConfigImportStdinSchema("")
		}
		delete(remote, "$schema")
		delete(remote, "$id")
		definitions["remote_config"] = remote
		stdin = map[string]any{"oneOf": []any{map[string]any{"type": "null"}, map[string]any{"$ref": "#/$defs/remote_config"}}}
	case strings.HasPrefix(capability.ID, "auth.add.") && capability.ID != "auth.add.gcloud":
		credentials := oauthCredentialSchema()
		if capability.ID == "auth.add.service-account" {
			credentials = serviceAccountCredentialSchema()
		}
		definitions["credentials"] = credentials
		stdin = map[string]any{"oneOf": []any{map[string]any{"type": "null"}, map[string]any{"$ref": "#/$defs/credentials"}}}
	}
	result := map[string]any{"$schema": draft, "$id": capability.InvocationSchema, "$defs": definitions, "type": "object", "additionalProperties": false, "required": []string{"arguments", "options", "stdin"}, "properties": map[string]any{"arguments": map[string]any{"type": "object", "additionalProperties": false, "required": required, "properties": arguments}, "options": map[string]any{"type": "object", "additionalProperties": false, "required": requiredOptions, "properties": options}, "stdin": stdin}}
	if slices.Contains([]string{"auth.add.oauth", "auth.add.service-account", "project.import"}, capability.ID) {
		result["x-fbrcm-input-selection"] = []any{map[string]any{
			"operator": "first_available", "sources": []any{"options.from", "stdin.document"},
			"on_missing": "interaction.required", "later_sources": "ignored_without_consumption",
		}}
	}
	if matching := selectionComposition(capability.ID, arguments, options); matching != nil {
		result["x-fbrcm-matching"] = []any{matching}
	}
	if slices.Contains([]string{"draft.publish", "draft.discard"}, capability.ID) {
		result["x-fbrcm-matching"] = []any{map[string]any{
			"operator": "draft_batch_selection", "project_source": "arguments.project", "all_source": "options.all",
			"composition": "exactly_one_source", "canonical_order": "deduplicate_then_sort_target_id",
		}}
	}
	constraints := optionConstraints(capability.ID, command, options)
	if slices.Contains([]string{"add", "update", "delete"}, capability.ID) {
		constraints = append(constraints, map[string]any{
			"if": map[string]any{"properties": map[string]any{"stdin": map[string]any{"not": map[string]any{"type": "null"}}}, "required": []string{"stdin"}},
			"then": map[string]any{"properties": map[string]any{"options": map[string]any{
				"not": map[string]any{"anyOf": []any{
					map[string]any{"required": []string{"draft"}},
					map[string]any{"required": []string{"change-note"}},
					map[string]any{"required": []string{"project"}},
					map[string]any{"required": []string{"dry-run"}},
				}},
			}}},
		})
	}
	if len(constraints) > 0 {
		result["allOf"] = constraints
	}
	pruneUnusedDefinitions(result)
	return result
}

func selectionComposition(commandID string, arguments, options map[string]any) map[string]any {
	sources := make([]any, 0, 6)
	for _, name := range []string{"project", "filter", "search", "expr"} {
		if _, ok := options[name]; ok {
			sources = append(sources, "options."+name)
		}
	}
	if _, ok := options["group"]; ok && slices.Contains([]string{"draft.diff", "project.import", "projects.diff", "projects.promote", "versions.diff"}, commandID) {
		sources = append(sources, "options.group")
	}
	if _, ok := arguments["parameter"]; ok && slices.Contains([]string{"delete", "get", "update"}, commandID) {
		sources = append(sources, "arguments.parameter")
	}
	if len(sources) == 0 {
		return nil
	}
	targetDefaults := []any{}
	if commandID != "auth.bind" && options["project"] != nil {
		targetDefaults = append(targetDefaults, map[string]any{"source": "options.project", "selection": "all_configured_projects_enabled_templates"})
	}
	return map[string]any{
		"operator": "selection_composition", "sources": sources,
		"repeated_source_combination": "or", "across_source_combination": "and", "absent_source_behavior": "match_all",
		"target_defaults": targetDefaults,
	}
}

func pruneUnusedDefinitions(schema map[string]any) {
	definitions, ok := schema["$defs"].(map[string]any)
	if !ok {
		return
	}
	used := make(map[string]bool)
	var collect func(any)
	collect = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			for key, child := range typed {
				if key == "$defs" {
					continue
				}
				if key == "$ref" {
					if ref, ok := child.(string); ok {
						if name, found := strings.CutPrefix(ref, "#/$defs/"); found {
							used[name] = true
						}
					}
				}
				collect(child)
			}
		case []any:
			for _, child := range typed {
				collect(child)
			}
		}
	}
	collect(schema)
	for previous := -1; previous != len(used); {
		previous = len(used)
		for name := range used {
			collect(definitions[name])
		}
	}
	if len(used) == 0 {
		delete(schema, "$defs")
		return
	}
	pruned := make(map[string]any, len(used))
	for name := range used {
		definition, exists := definitions[name]
		if !exists {
			panic("missing local schema definition " + name)
		}
		pruned[name] = definition
	}
	schema["$defs"] = pruned
}

func flagSchema(commandID string, flag contract.FlagCapability) map[string]any {
	result := map[string]any{"description": flag.Usage}
	if !flag.Effective {
		result["x-fbrcm-effective"] = false
	}
	if len(flag.EffectiveWhen) > 0 {
		result["x-fbrcm-effective-when"] = flag.EffectiveWhen
	}
	name := strings.TrimPrefix(flag.Name, "--")
	switch flag.Type {
	case "bool":
		result["type"] = "boolean"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		result["type"] = "integer"
	case "float32", "float64":
		result["type"] = "number"
	case "stringArray", "stringSlice", "intSlice":
		itemType := "string"
		if flag.Type == "intSlice" {
			itemType = "integer"
		}
		result["type"] = "array"
		result["minItems"] = 1
		result["items"] = map[string]any{"type": itemType}
		if itemType == "string" {
			result["items"].(map[string]any)["pattern"] = `.*\S.*`
		}
	case "duration":
		result["type"] = "string"
		result["pattern"] = `^[+-]?(?:0|(?:(?:\d+(?:\.\d*)?|\.\d+)(?:ns|us|µs|μs|ms|s|m|h))+)$`
		result["x-fbrcm-grammar"] = "go-duration"
	default:
		result["type"] = "string"
		if slices.Contains([]string{"change-note", "description", "group", "label", "value"}, name) {
			result["pattern"] = `^(?:$|.*\S.*)$`
		} else {
			result["pattern"] = `.*\S.*`
		}
	}
	if name == "stateless" && !contract.SupportsStatelessCommand(commandID) {
		result["const"] = false
	}
	if flag.Effective {
		applyFlagSemantics(result, commandID, name)
	}
	if name == "profile" && flag.Effective {
		addNormalization(result, "trim_unicode_whitespace")
		if len(flag.EffectiveWhen) == 0 {
			delete(result, "pattern")
			result["allOf"] = []any{map[string]any{"$ref": "#/$defs/path_segment"}}
		}
	}
	return result
}

func applyFlagSemantics(schema map[string]any, commandID, name string) {
	if name == "expr" {
		schema["$ref"] = "#/$defs/expression"
		delete(schema, "type")
	}
	if name == "filter" {
		if commandID == "draft.list" {
			schema["items"] = map[string]any{"$ref": "#/$defs/draft_filter"}
		} else {
			schema["items"] = map[string]any{
				"allOf":            []any{map[string]any{"$ref": "#/$defs/filter_query"}},
				"x-fbrcm-matching": []any{modePrefixedMatching(filterMatchingFields(commandID), false)},
			}
		}
	}
	if name == "project" {
		ref := "#/$defs/target_selector"
		if commandID == "auth.bind" {
			ref = "#/$defs/project_filter"
		}
		schema["items"] = map[string]any{"$ref": ref}
	}
	if commandID == "auth.bind" && name == "auth" {
		delete(schema, "pattern")
		schema["allOf"] = []any{map[string]any{"$ref": "#/$defs/path_segment"}}
	}
	if values := flagEnum(commandID, name); len(values) > 0 {
		target := schema
		if items, ok := schema["items"].(map[string]any); ok {
			target = items
		}
		if caseInsensitiveFlag(name) {
			target["pattern"] = caseInsensitiveEnumPattern(values)
			target["x-fbrcm-values"] = values
			caseOperator := "lowercase"
			if name == "color" {
				caseOperator = "uppercase"
			}
			addNormalization(target, "trim_unicode_whitespace", caseOperator)
		} else {
			target["enum"] = values
		}
	}
	switch name {
	case "limit":
		schema["minimum"] = 1
		schema["maximum"] = 2147483647
	case "priority":
		minimum := 1
		if commandID == "conditions.add" {
			minimum = 0
		}
		schema["minimum"] = minimum
		schema["maximum"] = 2147483647
		if commandID == "conditions.add" {
			addValidationRule(schema, map[string]any{
				"operator": "condition_priority", "operation": "add", "project_argument": "arguments.project",
				"maximum": "resolved_condition_count_plus_one", "zero_behavior": "append",
			})
		}
	case "templates":
		schema["minItems"] = 1
		addNormalization(schema, "deduplicate_preserve_first")
		addNormalizationRule(schema, map[string]any{
			"operator": "sort_by_declared_order", "source": "argv", "target": "normalized_invocation", "order": []any{"client", "server"},
		})
	case "condition", "name":
		if commandID == "update" {
			schema["pattern"] = `.*\S.*`
			if name == "name" {
				schema["maxLength"] = 256
			}
		}
	case "remove-conditional-value":
		if commandID == "update" {
			schema["items"] = map[string]any{"type": "string", "pattern": `.*\S.*`}
			addNormalization(schema["items"].(map[string]any), "trim_unicode_whitespace")
			schema["uniqueItems"] = true
			addNormalization(schema, "deduplicate_preserve_first")
		}
	case "timeout":
		// Cobra rejects every zero-valued duration spelling, not only the
		// literal "0" (for example 0s and 0h0m).
		schema["pattern"] = `^\+?(?:(?:0+(?:\.0*)?|\.0+)(?:ns|us|µs|μs|ms|s|m|h))*(?:(?:\d*[1-9]\d*(?:\.\d*)?|0*\.\d*[1-9]\d*)(?:ns|us|µs|μs|ms|s|m|h))(?:(?:\d+(?:\.\d*)?|\.\d+)(?:ns|us|µs|μs|ms|s|m|h))*$`
		schema["not"] = map[string]any{"pattern": `^\+?(?:0*\.)\d+ns$`}
		addValidationRule(schema, map[string]any{"operator": "parse_duration", "parser": "time.ParseDuration", "require_positive": true})
	}
	if name == "change-note" {
		schema["pattern"] = `^[^\x00-\x1f\x7f-\x9f]*$`
		addNormalization(schema, "trim_unicode_whitespace")
		addValidationRule(schema, map[string]any{"operator": "reject_raw_whitespace_only", "allow_empty": true})
	}
	if name == "expr" {
		addNormalization(schema, "trim_unicode_whitespace")
	}
	if name == "search" {
		switch commandID {
		case "conditions.list":
			addNormalization(schema, "trim_unicode_whitespace", "lowercase")
			addMatchingRule(schema, caseInsensitiveSubstringMatching("name", "expression"))
		case "groups.list":
			addNormalization(schema, "trim_unicode_whitespace", "lowercase")
			addMatchingRule(schema, caseInsensitiveSubstringMatching("name", "description"))
		default:
			addMatchingRule(schema, parameterSearchMatching())
		}
	}
	if name == "group" && slices.Contains([]string{"add", "update", "project.import", "projects.diff", "projects.promote"}, commandID) {
		target := schema
		if items, ok := schema["items"].(map[string]any); ok {
			target = items
		}
		addNormalization(target, "trim_unicode_whitespace")
		if items, ok := schema["items"].(map[string]any); ok {
			items["pattern"] = `.*\S.*`
			addNormalization(schema, "deduplicate_preserve_first")
		} else {
			addValidationRule(schema, map[string]any{"operator": "reject_raw_whitespace_only", "allow_empty": true})
		}
	}
	if name == "description" && slices.Contains([]string{"groups.add", "groups.edit"}, commandID) {
		addNormalization(schema, "trim_unicode_whitespace")
		addValidationRule(schema, map[string]any{"operator": "reject_raw_whitespace_only", "allow_empty": true})
	}
	if name == "expression" && slices.Contains([]string{"conditions.add", "conditions.edit"}, commandID) {
		addNormalization(schema, "trim_unicode_whitespace")
	}
	if commandID == "update" && slices.Contains([]string{"condition", "name"}, name) {
		addNormalization(schema, "trim_unicode_whitespace")
	}
	if commandID == "update" && name == "condition" {
		addMatchingRule(schema, map[string]any{"operator": "condition_name_resolution"})
	}
	if name == "description" && slices.Contains([]string{"add", "update", "groups.add", "groups.edit"}, commandID) {
		schema["maxLength"] = 256
	}
	if name == "expression" && slices.Contains([]string{"conditions.add", "conditions.edit"}, commandID) {
		schema["pattern"] = `.*\S.*`
		schema["x-fbrcm-validation"] = []any{map[string]any{
			"operator": "remote_validate", "service": "firebase_remote_config", "grammar": "condition_expression",
		}}
	}
	if commandID == "versions.list" && name == "before" {
		schema["pattern"] = `^[1-9]\d*$`
		schema["x-fbrcm-grammar"] = "positive Firebase version number"
	}
	if slices.Contains([]string{"add", "update"}, commandID) && name == "value" {
		schema["x-fbrcm-value-type-option"] = "type"
		schema["x-fbrcm-validation"] = []any{map[string]any{
			"operator": "dispatch_by_field", "field": "type", "normalized_case": "lowercase",
			"cases": map[string]any{
				"boolean": map[string]any{"operator": "match_schema_pattern"},
				"bool":    map[string]any{"operator": "match_schema_pattern"},
				"number":  map[string]any{"operator": "match_schema_pattern"},
				"json":    map[string]any{"operator": "parse_json", "specification": "RFC 8259", "consume": "entire_string"},
				"string":  map[string]any{"operator": "accept"},
			},
		}}
	}
	if name == "since" || name == "until" {
		schema["x-fbrcm-validation"] = []any{map[string]any{"operator": "parse_time", "specification": "Go time.RFC3339"}}
	}
}

func addNormalization(schema map[string]any, operators ...string) {
	rules, _ := schema["x-fbrcm-normalization"].([]any)
	for _, operator := range operators {
		rules = append(rules, map[string]any{"operator": operator, "source": "argv", "target": "normalized_invocation"})
	}
	schema["x-fbrcm-normalization"] = rules
}

func addNormalizationRule(schema map[string]any, rule map[string]any) {
	rules, _ := schema["x-fbrcm-normalization"].([]any)
	schema["x-fbrcm-normalization"] = append(rules, rule)
}

func addValidationRule(schema map[string]any, rule map[string]any) {
	rules, _ := schema["x-fbrcm-validation"].([]any)
	schema["x-fbrcm-validation"] = append(rules, rule)
}

func addMatchingRule(schema map[string]any, rule map[string]any) {
	rules, _ := schema["x-fbrcm-matching"].([]any)
	schema["x-fbrcm-matching"] = append(rules, rule)
}

func stringValues(values ...string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func modePrefixedMatching(fields []string, targetAware bool) map[string]any {
	rule := map[string]any{
		"operator": "mode_prefixed_query", "query_normalization": "trim_unicode_whitespace",
		"default_mode": "fuzzy", "comparison": "unicode_case_insensitive",
		"mode_prefixes": map[string]any{"~": "fuzzy", "^": "starts-with", "/": "includes", "=": "exact"},
		"fields":        stringValues(fields...),
	}
	if targetAware {
		rule["target_prefixes"] = []any{"client", "server"}
		rule["unqualified_target_selection"] = "all_configured_enabled_templates"
		rule["explicit_target_selection"] = "single_named_template"
		rule["client_target_canonicalization"] = "unqualified_project_id"
	}
	return rule
}

func statelessGetProjectMatching() map[string]any {
	rule := modePrefixedMatching([]string{"project_id", "display_name"}, true)
	rule["unqualified_target_selection"] = "client_template"
	return rule
}

func draftFilterMatching() map[string]any {
	rule := modePrefixedMatching([]string{"project_id", "display_name", "repository_aliases"}, true)
	rule["unqualified_target_selection"] = "existing_drafts_in_configured_enabled_templates_or_client_fallback"
	return rule
}

func draftResolutionMatching() map[string]any {
	return map[string]any{
		"operator": "draft_resolution", "candidate_source": "local_draft_target_ids",
		"fields":              []any{"project_id", "display_name", "repository_aliases"},
		"query_normalization": "preserve_argv",
		"comparison":          "exact_case_sensitive",
		"precedence":          []any{"exact_draft_project_id", "exact_repository_alias", "exact_display_name"},
		"target_prefixes":     []any{"client", "server"}, "unqualified_target_selection": "configured_primary_template_or_client_fallback",
		"explicit_target_selection": "single_named_template", "client_target_canonicalization": "unqualified_project_id",
		"zero_result": "draft.not_found", "multiple_result": "draft.ambiguous",
	}
}

func projectPositionalMatching(targetAware bool) map[string]any {
	rule := map[string]any{
		"operator":            "project_positional_resolution",
		"fields":              stringValues("project_id", "display_name", "repository_aliases"),
		"query_normalization": "preserve_argv",
		"comparison":          "exact_case_sensitive",
		"precedence":          []any{"exact_project_id", "exact_repository_alias", "exact_display_name"},
	}
	if targetAware {
		rule["target_prefixes"] = []any{"client", "server"}
		rule["unqualified_target_selection"] = "configured_primary_template"
		rule["explicit_target_selection"] = "single_named_template"
		rule["client_target_canonicalization"] = "unqualified_project_id"
	}
	return rule
}

func filterMatchingFields(commandID string) []string {
	switch commandID {
	case "conditions.list":
		return []string{"condition_name"}
	case "experiments.list":
		return []string{"display_name"}
	case "groups.list":
		return []string{"group_name"}
	case "projects.forget", "projects.list", "projects.update":
		return []string{"project_id", "display_name", "repository_aliases"}
	default:
		return []string{"parameter_key"}
	}
}

func caseInsensitiveSubstringMatching(fields ...string) map[string]any {
	publishedFields := make([]any, len(fields))
	for index, field := range fields {
		publishedFields[index] = field
	}
	return map[string]any{
		"operator": "case_insensitive_substring", "fields": publishedFields,
		"query_normalization": "trim_then_unicode_lowercase", "haystack_normalization": "unicode_lowercase", "separator": "\n",
	}
}

func parameterSearchMatching() map[string]any {
	return map[string]any{
		"operator":          "parameter_search",
		"normalized_fields": []any{"name", "description", "condition_names"},
		"raw_fields":        []any{"default_value", "conditional_values", "condition_expressions"},
		"normalized_query":  "lowercase_alphanumeric_words", "raw_query": "collapse_unicode_whitespace",
		"match": "substring", "combination": "or",
	}
}

func flagEnum(commandID, name string) []string {
	switch name {
	case "color":
		if strings.HasPrefix(commandID, "conditions.") {
			return []string{"CONDITION_DISPLAY_COLOR_UNSPECIFIED", "BLUE", "BROWN", "CYAN", "DEEP_ORANGE", "GREEN", "INDIGO", "LIME", "ORANGE", "PINK", "PURPLE", "TEAL"}
		}
	case "type":
		return []string{"string", "boolean", "bool", "number", "json"}
	case "against":
		return []string{"base", "current"}
	case "merge-resolve":
		return []string{"current", "import"}
	case "primary":
		return []string{"client", "server"}
	case "templates":
		return []string{"client", "server"}
	case "format":
		return []string{"json", "xml", "plist"}
	case "conflict":
		return []string{"error", "keep", "overwrite"}
	case "scope":
		switch commandID {
		case "config.validate":
			return []string{"all", "effective", "global", "local"}
		case "config.show":
			return []string{"effective", "global", "local"}
		default:
			return []string{"global", "local"}
		}
	}
	return nil
}

func caseInsensitiveFlag(name string) bool {
	return slices.Contains([]string{"type", "merge-resolve", "primary", "templates", "format", "scope", "color"}, name)
}

func caseInsensitiveEnumPattern(values []string) string {
	var pattern strings.Builder
	pattern.WriteString("^(?:")
	for index, value := range values {
		if index > 0 {
			pattern.WriteByte('|')
		}
		for _, char := range value {
			switch {
			case char >= 'a' && char <= 'z':
				pattern.WriteByte('[')
				pattern.WriteRune(char)
				pattern.WriteRune(char - ('a' - 'A'))
				pattern.WriteByte(']')
			case char >= 'A' && char <= 'Z':
				pattern.WriteByte('[')
				pattern.WriteRune(char + ('a' - 'A'))
				pattern.WriteRune(char)
				pattern.WriteByte(']')
			default:
				pattern.WriteString(regexp.QuoteMeta(string(char)))
			}
		}
	}
	pattern.WriteString(")$")
	return pattern.String()
}

func argumentSchema(commandID, name string) map[string]any {
	schema := map[string]any{"type": "string", "minLength": 1, "pattern": `.*\S.*`}
	if commandID == "draft.change-note" && name == "text" {
		schema["pattern"] = `^[^\x00-\x1f\x7f-\x9f]*$`
		schema["not"] = map[string]any{"pattern": `^\s*$`}
	}
	if name == "parameter" && slices.Contains([]string{"delete", "get", "update"}, commandID) {
		addMatchingRule(schema, map[string]any{"operator": "parameter_argument_resolution"})
	}
	if commandID == "duplicate" && name == "source" {
		addMatchingRule(schema, map[string]any{"operator": "duplicate_source_resolution"})
	}
	if strings.HasPrefix(commandID, "conditions.") && name == "condition" {
		addMatchingRule(schema, map[string]any{"operator": "condition_positional_resolution"})
	}
	if commandID == "personalizations.show" && name == "personalization_id" {
		addMatchingRule(schema, map[string]any{"operator": "personalization_id_resolution"})
	}
	if slices.Contains([]string{"groups.delete", "groups.edit", "groups.rename"}, commandID) && name == "group" {
		addMatchingRule(schema, map[string]any{"operator": "group_name_resolution"})
	}
	if commandID == "config.show" && name == "key" {
		return configKeySchema(true)
	}
	if commandID == "config.reset" && name == "key" {
		return configKeySchema(false)
	}
	if name == "parameter" && slices.Contains([]string{"add", "delete", "update"}, commandID) || commandID == "duplicate" && slices.Contains([]string{"source", "target"}, name) {
		schema["maxLength"] = 256
		schema["pattern"] = `.*\S.*`
	}
	if argumentIsTrimmed(commandID, name) {
		addNormalization(schema, "trim_unicode_whitespace")
	}
	if parameterOrGroupName(commandID, name) {
		schema["maxLength"] = 256
		schema["pattern"] = `.*\S.*`
	}
	if (commandID == "conditions.add" && name == "name") || (commandID == "conditions.rename" && name == "new_name") {
		schema["maxLength"] = 100
		schema["pattern"] = `.*\S.*`
	}
	if commandID == "projects.aliases.set" && name == "project_id" {
		return map[string]any{"$ref": "#/$defs/physical_project_id"}
	}
	if strings.HasPrefix(commandID, "projects.aliases.") && name == "alias" {
		result := map[string]any{"$ref": "#/$defs/project_alias"}
		if commandID == "projects.aliases.remove" {
			addMatchingRule(result, map[string]any{"operator": "project_alias_resolution"})
		}
		return result
	}
	if strings.HasPrefix(commandID, "auth.") && name == "auth_id" {
		result := map[string]any{"$ref": "#/$defs/path_segment"}
		if slices.Contains([]string{"auth.login", "auth.path", "auth.delete"}, commandID) {
			addMatchingRule(result, map[string]any{"operator": "auth_id_resolution"})
		}
		return result
	}
	if strings.HasPrefix(commandID, "profile.") && slices.Contains([]string{"profile", "name", "old_name", "new_name"}, name) {
		result := map[string]any{"$ref": "#/$defs/path_segment"}
		if commandID == "profile.delete" && name == "profile" || commandID == "profile.rename" && name == "old_name" {
			addMatchingRule(result, map[string]any{"operator": "profile_name_resolution"})
		}
		return result
	}
	if slices.Contains([]string{"experiment_id", "rollout_id"}, name) {
		schema["pattern"] = `.*\S.*`
		collection := strings.TrimSuffix(strings.Split(commandID, ".")[0], "s") + "s"
		schema["anyOf"] = []any{
			map[string]any{"pattern": `^[^/]+$`},
			map[string]any{"pattern": `^projects/[^/]+/namespaces/firebase/` + collection + `/[^/]+$`},
		}
		addValidationRule(schema, map[string]any{
			"operator": "managed_feature_id", "collection": collection, "project_argument": "project",
			"accepted_forms": []any{"bare_id", "resolved_project_resource_name"},
		})
	}
	if name == "personalization_id" {
		schema["pattern"] = `.*\S.*`
	}
	if name == "project" || strings.HasSuffix(name, "_project") || strings.Contains(name, "project") && !strings.Contains(name, "profile") {
		if strings.HasPrefix(commandID, "draft.") {
			return map[string]any{"$ref": "#/$defs/draft_selector"}
		}
		if literalProjectPositional(commandID, name) {
			return map[string]any{"$ref": "#/$defs/project_positional_selector"}
		}
		if prefixRejectingProjectPositional(commandID, name) {
			return map[string]any{"$ref": "#/$defs/physical_project_selector"}
		}
		return map[string]any{"$ref": "#/$defs/target_positional_selector"}
	}
	if name == "schema_id" {
		schema["pattern"] = `^urn:fbrcm:schema:cli:`
		addMatchingRule(schema, map[string]any{"operator": "schema_id_resolution"})
	}
	if commandID == "conditions.move" && name == "priority" {
		schema["pattern"] = `^\+?0*[1-9]\d*$`
		schema["x-fbrcm-grammar"] = "positive condition priority accepted by strconv.Atoi, including an optional plus sign and leading zeroes"
		schema["x-fbrcm-validation"] = []any{
			map[string]any{"operator": "parse_positive_integer", "parser": "strconv.Atoi", "minimum": 1},
			map[string]any{
				"operator": "condition_priority", "operation": "move", "project_argument": "arguments.project",
				"maximum": "resolved_condition_count",
			},
		}
	}
	if slices.Contains([]string{"versions.show", "versions.diff", "versions.export", "versions.rollback", "versions.restore"}, commandID) {
		if strings.Contains(name, "version") {
			return map[string]any{"$ref": "#/$defs/version_selector"}
		}
	}
	if commandID == "versions.diff" && slices.Contains([]string{"from", "to"}, name) {
		return map[string]any{"$ref": "#/$defs/version_selector"}
	}
	return schema
}

func argumentIsTrimmed(commandID, name string) bool {
	if commandID == "add" && name == "parameter" || commandID == "duplicate" && name == "target" {
		return true
	}
	if commandID == "groups.add" && name == "name" || commandID == "groups.rename" && name == "new_name" {
		return true
	}
	if commandID == "conditions.add" && name == "name" || commandID == "conditions.rename" && name == "new_name" {
		return true
	}
	return commandID == "draft.change-note" && name == "text"
}

func literalProjectPositional(commandID, name string) bool {
	return name == "project" && slices.Contains([]string{"project.open", "project.show"}, commandID)
}

func prefixRejectingProjectPositional(commandID, name string) bool {
	if name != "project" {
		return false
	}
	return slices.Contains([]string{
		"project.templates.show", "project.templates.set",
		"experiments.list", "experiments.show", "experiments.delete",
		"rollouts.list", "rollouts.show", "rollouts.delete",
		"personalizations.list", "personalizations.show",
	}, commandID)
}

func parameterOrGroupName(commandID, name string) bool {
	if name == "parameter" && slices.Contains([]string{"add", "delete", "update"}, commandID) {
		return true
	}
	if name == "name" && commandID == "groups.add" {
		return true
	}
	if name == "new_name" && commandID == "groups.rename" {
		return true
	}
	return false
}

func configKeySchema(show bool) map[string]any {
	values := []string{"powerline_glyphs", "keys"}
	if show {
		values = append(values, "profile", "hooks", "hooks.timeout", "hooks.pre_publish", "hooks.post_publish", "projects")
	}
	values = append(values, "projects.aliases")
	values = append(values, keyBindingConfigNames(true)...)
	slices.Sort(values)
	schema := map[string]any{
		"type": "string",
		"anyOf": []any{
			map[string]any{"enum": values},
			map[string]any{"pattern": `^projects\.aliases\.[a-z][a-z0-9_-]{0,62}$`},
		},
	}
	prefixes := []any{"keys.", "projects.aliases."}
	if show {
		prefixes = append(prefixes, "hooks.")
	}
	addNormalizationRule(schema, map[string]any{
		"operator": "trim_unicode_whitespace_if_prefix", "source": "argv", "target": "normalized_invocation", "prefixes": prefixes,
	})
	return schema
}

func keyBindingConfigNames(includeBlocks bool) []string {
	keyMap := tuiconfig.DefaultKeyMap()
	blocks := make([]string, 0, len(keyMap))
	for block := range keyMap {
		blocks = append(blocks, string(block))
	}
	slices.Sort(blocks)
	names := make([]string, 0)
	for _, block := range blocks {
		if includeBlocks {
			names = append(names, "keys."+block)
		}
		actions := make([]string, 0, len(keyMap[tuiconfig.Block(block)]))
		for action := range keyMap[tuiconfig.Block(block)] {
			actions = append(actions, string(action))
		}
		slices.Sort(actions)
		for _, action := range actions {
			names = append(names, "keys."+block+"."+action)
		}
	}
	return names
}

func semanticDefinitions(commandID string) map[string]any {
	context := expressionContext(commandID)
	return map[string]any{
		"target_selector": map[string]any{"allOf": []any{map[string]any{"$ref": contract.SemanticRef("target_selector")}}, "x-fbrcm-normalization": []any{map[string]any{"operator": "canonicalize_target_selector", "source": "argv", "target": "normalized_invocation"}}},
		"draft_filter":    map[string]any{"allOf": []any{map[string]any{"$ref": contract.SemanticRef("draft_filter")}}, "x-fbrcm-normalization": []any{map[string]any{"operator": "canonicalize_target_selector", "source": "argv", "target": "normalized_invocation"}}},
		"draft_selector":  map[string]any{"allOf": []any{map[string]any{"$ref": contract.SemanticRef("draft_selector")}}, "x-fbrcm-normalization": []any{map[string]any{"operator": "canonicalize_positional_target_selector", "source": "argv", "target": "normalized_invocation"}}},
		"filter_mode":     map[string]any{"$ref": contract.SemanticRef("filter_mode")},
		"filter_query": map[string]any{
			"allOf":                 []any{map[string]any{"$ref": contract.SemanticRef("filter_query")}},
			"x-fbrcm-normalization": []any{map[string]any{"operator": "trim_unicode_whitespace", "source": "argv", "target": "normalized_invocation"}},
		},
		"project_filter": map[string]any{"allOf": []any{
			map[string]any{"$ref": contract.SemanticRef("filter_query")},
			map[string]any{"not": map[string]any{"pattern": `^(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@`}},
		}, "x-fbrcm-normalization": []any{map[string]any{"operator": "trim_unicode_whitespace", "source": "argv", "target": "normalized_invocation"}}, "x-fbrcm-matching": []any{modePrefixedMatching([]string{"project_id", "display_name", "repository_aliases"}, false)}},
		"target_positional_selector":  map[string]any{"allOf": []any{map[string]any{"$ref": contract.SemanticRef("target_positional_selector")}}, "x-fbrcm-normalization": []any{map[string]any{"operator": "canonicalize_positional_target_selector", "source": "argv", "target": "normalized_invocation"}}},
		"project_positional_selector": map[string]any{"$ref": contract.SemanticRef("project_positional_selector")},
		"physical_project_selector":   map[string]any{"$ref": contract.SemanticRef("physical_project_selector")},
		"physical_project_id": map[string]any{
			"$ref": contract.SemanticRef("physical_project_id"),
		},
		"stateless_target_selector": map[string]any{
			"type":            "string",
			"pattern":         `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?[^=^/~@\s][^@\s]*$`,
			"x-fbrcm-grammar": "[client@|server@]literal physical project ID; an omitted template prefix selects client",
			"x-fbrcm-matching": []any{map[string]any{
				"operator": "literal_project_id", "comparison": "exact_case_sensitive", "default_template": "client", "lookup": false,
			}},
		},
		"stateless_get_project_selector": map[string]any{
			"oneOf": []any{
				map[string]any{
					"type":            "string",
					"pattern":         `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?=[^=^/~@\s][^@\s]*$`,
					"x-fbrcm-grammar": "[client@|server@]=literal physical project ID; exact mode bypasses project discovery",
					"x-fbrcm-matching": []any{map[string]any{
						"operator": "literal_project_id", "comparison": "exact_case_sensitive", "default_template": "client", "lookup": false,
					}},
				},
				map[string]any{
					"allOf": []any{
						map[string]any{"$ref": "#/$defs/target_selector"},
						map[string]any{"not": map[string]any{"pattern": `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?=`}},
					},
					"x-fbrcm-matching": []any{statelessGetProjectMatching()},
				},
			},
		},
		"project_alias": map[string]any{"$ref": contract.SemanticRef("project_alias")},
		"path_segment":  map[string]any{"$ref": contract.SemanticRef("path_segment")},
		"version_selector": map[string]any{"allOf": []any{map[string]any{
			"$ref": contract.SemanticRef("version_selector"),
		}}, "x-fbrcm-matching": []any{map[string]any{"operator": "version_resolution"}}},
		"expression_context": map[string]any{
			"$ref": contract.SemanticRef("expression_context"),
		},
		"expression": map[string]any{"allOf": []any{map[string]any{"$ref": contract.SemanticRef("expression")}}, "x-fbrcm-expression-context": context},
	}
}

func expressionContext(commandID string) string {
	switch {
	case strings.HasPrefix(commandID, "conditions."):
		return "condition"
	case slices.Contains([]string{"add", "duplicate", "projects.list", "projects.update", "projects.forget", "projects.reset"}, commandID):
		return "project"
	default:
		return "parameter"
	}
}

func semanticSchema() map[string]any {
	definitions := map[string]any{
		"remote_mutation_status": map[string]any{"type": "string", "enum": []string{"unchanged", "preparation-failed", "published", "validation-failed", "conflict", "publish-failed", "published-cache-failed", "published-hook-failed", "drafted", "would-draft", "would-publish", "draft-failed"}},
		"no_op_reason":           map[string]any{"type": "string", "enum": []string{"no_match", "already_applied"}},
		"validation_source":      map[string]any{"type": "string", "enum": []string{"", "local", "firebase"}},
		"artifact_encoding":      map[string]any{"type": "string", "enum": []string{"none", "json", "utf-8", "base64"}},
	}
	maps.Copy(definitions, extensionSchemaDefinitions())
	definitions["target_selector"] = map[string]any{
		"type": "string", "minLength": 1,
		"pattern":          `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?(?:[=^/~].+|.+)$`,
		"not":              map[string]any{"pattern": `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?[=^/~]?\s*$`},
		"x-fbrcm-grammar":  "[client@|server@][mode-prefix]project-query",
		"x-fbrcm-matching": []any{modePrefixedMatching([]string{"project_id", "display_name", "repository_aliases"}, true)},
	}
	definitions["draft_filter"] = map[string]any{
		"type": "string", "minLength": 1,
		"pattern":          `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?(?:[=^/~].+|.+)$`,
		"not":              map[string]any{"pattern": `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?[=^/~]?\s*$`},
		"x-fbrcm-grammar":  "[client@|server@][mode-prefix]draft-project-query",
		"x-fbrcm-matching": []any{draftFilterMatching()},
	}
	definitions["draft_selector"] = map[string]any{
		"type": "string", "minLength": 1,
		"pattern":          `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?.+$`,
		"not":              map[string]any{"pattern": `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?\s*$`},
		"x-fbrcm-grammar":  "[client@|server@](exact draft project ID, exact repository alias, or exact display name); resource matching is case-sensitive and =, ^, /, and ~ are literal characters",
		"x-fbrcm-matching": []any{draftResolutionMatching()},
	}
	definitions["target_positional_selector"] = map[string]any{
		"type": "string", "minLength": 1,
		"pattern":          `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?.+$`,
		"not":              map[string]any{"pattern": `^(?:(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@)?\s*$`},
		"x-fbrcm-grammar":  "[client@|server@](exact project ID, exact repository alias, or exact display name); resource matching is case-sensitive and =, ^, /, and ~ are literal characters",
		"x-fbrcm-matching": []any{projectPositionalMatching(true)},
	}
	definitions["project_positional_selector"] = map[string]any{
		"type": "string", "minLength": 1, "pattern": `.*\S.*`,
		"x-fbrcm-grammar":  "exact case-sensitive project ID, repository alias, or display name without template-target parsing; argv whitespace is preserved and all characters are literal",
		"x-fbrcm-matching": []any{projectPositionalMatching(false)},
	}
	definitions["physical_project_selector"] = map[string]any{
		"type": "string", "minLength": 1, "pattern": `.*\S.*`,
		"not":              map[string]any{"pattern": `^(?:[cC][lL][iI][eE][nN][tT]|[sS][eE][rR][vV][eE][rR])@`},
		"x-fbrcm-grammar":  "exact case-sensitive physical project ID, repository alias, or display name without a client@ or server@ template-target prefix; argv whitespace is preserved",
		"x-fbrcm-matching": []any{projectPositionalMatching(false)},
	}
	definitions["physical_project_id"] = map[string]any{"type": "string", "pattern": `^[^=^/~@\s][^@\s]*$`, "x-fbrcm-grammar": "literal physical project ID without target-selector syntax"}
	definitions["project_alias"] = map[string]any{"type": "string", "pattern": `^[a-z][a-z0-9_-]{0,62}$`}
	definitions["path_segment"] = map[string]any{"type": "string", "pattern": `^[^/\\\s](?:[^/\\]*[^/\\\s])?$`, "not": map[string]any{"enum": []string{".", ".."}}, "x-fbrcm-grammar": "nonempty single filesystem-safe path segment"}
	definitions["version_selector"] = map[string]any{
		"type": "string", "pattern": `^(?:\+?0*[1-9]\d*|current|latest|previous|(?:current|latest)~\+?0*(?:[1-9]\d?|[12]\d{2}))$`,
		"x-fbrcm-grammar": "exact case-sensitive positive version number accepted by strconv.ParseInt, current, latest, previous, current~N, or latest~N; argv whitespace is not trimmed; numeric components accept an optional plus sign and leading zeroes",
		"x-fbrcm-validation": []any{map[string]any{
			"operator": "parse_version_selector", "absolute_parser": "strconv.ParseInt base 10 bitSize 64; require result > 0", "relative_parser": "strconv.ParseInt base 10 bitSize 32; require result > 0", "maximum_relative_distance": 299,
		}},
	}
	definitions["filter_mode"] = map[string]any{"type": "string", "enum": []string{"fuzzy", "starts-with", "includes", "exact"}}
	definitions["filter_query"] = map[string]any{
		"type": "string", "minLength": 1, "pattern": `^(?:[=^/~])?.+$`,
		"not":             map[string]any{"pattern": `^[=^/~]?\s*$`},
		"x-fbrcm-grammar": "optional mode prefix (= exact, ^ starts-with, / includes, ~ fuzzy) followed by query",
	}
	definitions["expression_context"] = map[string]any{"type": "string", "enum": []string{"project", "parameter", "condition"}}
	definitions["expression"] = map[string]any{
		"type": "string", "minLength": 1, "pattern": `.*\S.*`,
		"x-fbrcm-grammar": "expr-lang v1.17.8 expression that compiles and evaluates to boolean",
		"x-fbrcm-validation": []any{map[string]any{
			"operator": "compile_expression", "language": "expr-lang", "version": "1.17.8", "result_type": "boolean",
		}},
		"x-fbrcm-language": expressionLanguageMetadata(),
	}
	return map[string]any{
		"$schema":                    draft,
		"$id":                        contract.SemanticSchemaID(),
		"$defs":                      definitions,
		"x-fbrcm-extension-language": extensionLanguageMetadata(),
	}
}

func extensionLanguageMetadata() map[string]any {
	operation := func(operands []string, result, semantics string) map[string]any {
		return map[string]any{"operands": operands, "result": result, "semantics": semantics}
	}
	return map[string]any{
		"version": 1,
		"option_effectiveness": map[string]any{
			"keyword":   "x-fbrcm-effective",
			"semantics": "False means argv accepts the option but the selected command does not apply its value; absence means effective.",
		},
		"conditional_option_effectiveness": map[string]any{
			"keyword":   "x-fbrcm-effective-when",
			"semantics": "When present, argv accepts the option but applies it only when at least one listed condition clause is true; clauses use the same OR-of-AND predicate language as capability effective_when.",
		},
		"input_selection": map[string]any{
			"keyword":   "x-fbrcm-input-selection",
			"schema":    contract.SemanticRef("input_selection_rules"),
			"semantics": "Evaluate sources in their declared order and select the first available source. Do not consume or apply later sources. If none is available, return the declared structured result.",
			"operators": map[string]any{
				"first_available": operation([]string{"sources", "on_missing", "later_sources"}, "selected_input_source_or_interaction", "Select options.from when its nonempty supplied value is present; otherwise select stdin.document when a redirected JSON document is present; otherwise return interaction.required. A selected earlier source leaves every later source unconsumed and ineffective."),
			},
		},
		"validation": map[string]any{
			"keyword":   "x-fbrcm-validation",
			"schema":    contract.SemanticRef("validation_rules"),
			"semantics": "Every listed rule must accept the instance after x-fbrcm-normalization and standard Draft 2020-12 validation, except reject_raw_whitespace_only, which evaluates the original argv value before normalization.",
			"operators": map[string]any{
				"accept":                     operation(nil, "accept", "Accept the instance without an additional check."),
				"compile_expression":         operation([]string{"language", "version", "result_type"}, "accept_or_reject", "Compile the complete string with the named language and require the declared result type."),
				"dispatch_by_field":          operation([]string{"field", "normalized_case", "cases"}, "accept_or_reject", "Read the sibling field, normalize its case as declared, and evaluate the matching case rule."),
				"fields_differ":              operation([]string{"fields", "comparison"}, "accept_or_reject", "Read both named fields and require unequal values after the declared comparison normalization."),
				"match_schema_pattern":       operation(nil, "accept_or_reject", "Apply the pattern in the active conditional schema to the instance."),
				"parse_email":                operation([]string{"parser", "require_exact"}, "accept_or_reject", "Parse the complete address with the named parser and require the parsed address to equal the input when require_exact is true."),
				"parse_duration":             operation([]string{"parser", "require_positive"}, "accept_or_reject", "Parse the complete duration with the named runtime parser, reject syntax and overflow errors, and require a result greater than zero when require_positive is true."),
				"parse_json":                 operation([]string{"specification", "consume"}, "accept_or_reject", "Parse the string as JSON and, when consume is entire_string, reject trailing non-whitespace input."),
				"parse_positive_integer":     operation([]string{"parser", "minimum"}, "accept_or_reject", "Parse the entire decimal string with the named runtime parser and require a result at least minimum; parser overflow is rejection."),
				"condition_priority":         operation([]string{"operation", "project_argument", "maximum", "zero_behavior?"}, "accept_or_reject", "After resolving the project template and loading its effective Remote Config, enforce the declared condition-count-dependent upper bound. Add accepts zero as append and otherwise accepts 1 through count plus one; move accepts 1 through count. Out-of-range values return condition.invalid."),
				"parse_time":                 operation([]string{"specification"}, "accept_or_reject", "Parse the complete timestamp using the named runtime layout and parser semantics."),
				"parse_uri":                  operation([]string{"parser", "normalization", "require_absolute"}, "accept_or_reject", "Normalize the string as declared, parse it with the named parser, and require an absolute URI with a nonempty scheme when require_absolute is true."),
				"parse_version_selector":     operation([]string{"absolute_parser", "relative_parser", "maximum_relative_distance"}, "accept_or_reject", "Apply the named runtime integer parsers to absolute and relative numeric components, require positive results, and reject a relative distance above the declared maximum."),
				"reject_raw_whitespace_only": operation([]string{"allow_empty"}, "accept_or_reject", "Inspect the original argv string before normalization; accept an exact empty string when allow_empty is true, but reject every nonempty string made only of Unicode whitespace."),
				"local_validate":             operation([]string{"validator"}, "accept_or_reject", "Decode a clone of the Remote Config template and require firebase.NormalizeRemoteConfigForUpdate to succeed before import transformations are applied."),
				"managed_feature_id":         operation([]string{"collection", "project_argument", "accepted_forms"}, "accept_or_reject", "Without trimming argv, accept either a slash-free ID exactly as supplied or the exact case-sensitive Firebase resource name for the resolved project and declared managed-feature collection. Remote lookup succeeds only for a canonical existing ID."),
				"remote_validate":            operation([]string{"service", "grammar"}, "accept_or_reject", "Require the named remote service to accept the instance under the declared grammar."),
				"unique_by":                  operation([]string{"field"}, "accept_or_reject", "Require the named field to be present and pairwise unique across all array members."),
				"unique_tokens":              operation([]string{"separator", "range"}, "accept_or_reject", "Split the string by separator and require tokens in the selected range to be pairwise distinct."),
			},
		},
		"normalization": map[string]any{
			"keyword":   "x-fbrcm-normalization",
			"schema":    contract.SemanticRef("normalization_rules"),
			"semantics": "Apply the listed operations in order when converting raw argv to the normalized invocation object, before validating that object.",
			"operators": map[string]any{
				"trim_unicode_whitespace":                 operation([]string{"source", "target"}, "string", "Remove the leading and trailing Unicode White_Space code points from the source string."),
				"trim_unicode_whitespace_if_prefix":       operation([]string{"source", "target", "prefixes"}, "string", "Trim outer Unicode whitespace only when the trimmed value starts with one of the declared prefixes; otherwise preserve the source string."),
				"lowercase":                               operation([]string{"source", "target"}, "string", "Convert the source string to lowercase."),
				"uppercase":                               operation([]string{"source", "target"}, "string", "Convert the source string to uppercase."),
				"canonicalize_target_selector":            operation([]string{"source", "target"}, "string", "Trim outer Unicode whitespace, recognize client@ and server@ case-insensitively, lowercase an explicit target prefix, trim Unicode whitespace around the remaining project query, and preserve whether client@ was explicit."),
				"canonicalize_positional_target_selector": operation([]string{"source", "target"}, "string", "Recognize client@ and server@ case-insensitively, lowercase an explicit target prefix, omit an explicit client@ prefix from the canonical target identity, and preserve the project selector exactly without trimming whitespace."),
				"deduplicate_preserve_first":              operation([]string{"source", "target"}, "array", "Remove repeated array values after item normalization while preserving the first occurrence and its order."),
				"sort_by_declared_order":                  operation([]string{"source", "target", "order"}, "array", "Sort array values according to the complete declared order after item normalization and deduplication."),
			},
		},
		"matching": map[string]any{
			"keyword":   "x-fbrcm-matching",
			"schema":    contract.SemanticRef("matching_rules"),
			"semantics": "Apply the declared query preparation and matching algorithm when predicting local selection. Matching metadata does not mutate the normalized invocation value unless a separate normalization rule says so.",
			"operators": map[string]any{
				"auth_id_resolution":              operation(nil, "selection", "Compare the positional auth ID exactly and case-sensitively against configured canonical auth IDs. No match returns auth.not_found; IDs are unique."),
				"literal_project_id":              operation([]string{"comparison", "default_template", "lookup"}, "selection", "Use the supplied physical Firebase project ID exactly without project-registry, display-name, or repository-alias lookup. An omitted template prefix selects the client template; client@ and server@ select one named template."),
				"mode_prefixed_query":             operation([]string{"fields", "query_normalization", "default_mode", "mode_prefixes", "comparison", "target_prefixes?", "unqualified_target_selection?", "explicit_target_selection?", "client_target_canonicalization?"}, "boolean_or_target_selection", "Match the declared resource fields after optional target-prefix parsing, using the first query rune as a declared mode prefix or the default mode. Fuzzy matches query runes as an ordered subsequence; starts-with, includes, and exact have their literal meanings. Target-aware rules additionally declare unqualified expansion, explicit single-template selection, and client target canonicalization."),
				"project_positional_resolution":   operation([]string{"fields", "query_normalization", "comparison", "precedence", "target_prefixes?", "unqualified_target_selection?", "explicit_target_selection?", "client_target_canonicalization?"}, "selection", "Resolve the untrimmed literal project selector exactly and case-sensitively in precedence order: project ID, repository alias, then display name. Substrings and mode prefixes have no search meaning. Target-aware rules additionally declare primary-template resolution, explicit single-template selection, and client target canonicalization."),
				"draft_resolution":                operation([]string{"candidate_source", "fields", "query_normalization", "comparison", "precedence", "target_prefixes", "unqualified_target_selection", "explicit_target_selection", "client_target_canonicalization", "zero_result", "multiple_result"}, "selection", "Resolve only existing local draft target IDs. Parse an optional template target without trimming its project selector, then compare exactly and case-sensitively in precedence order: draft physical project ID, repository alias, then display name. Unqualified queries select the configured primary template. Substrings and mode prefixes have no search meaning. Zero and multiple matches return the declared typed draft problems."),
				"draft_batch_selection":           operation([]string{"project_source", "all_source", "composition", "canonical_order"}, "selection", "Require exactly one of explicit project selectors or the truthy all option. Resolve each project selector with draft_resolution, deduplicate canonical target IDs, and sort them lexicographically; all selects every existing local draft target ID."),
				"condition_name_resolution":       operation(nil, "selection", "Resolve a condition by exact name first, then by Unicode case-insensitive equality in stored condition order. This operator is used by the non-positional update condition option; no match returns condition.not_found."),
				"condition_positional_resolution": operation(nil, "selection", "Compare the untrimmed positional condition selector exactly and case-sensitively against canonical condition names. No match returns condition.not_found; duplicate exact names resolve to the first stored condition."),
				"duplicate_source_resolution":     operation(nil, "selection", "Compare the untrimmed positional source name exactly and case-sensitively against parameter keys across root and all groups, skip a target on zero matches, and return parameter.ambiguous with candidates on multiple exact matches."),
				"group_name_resolution":           operation(nil, "selection", "Compare the untrimmed positional group selector exactly and case-sensitively against canonical group map keys. Zero matches skip that target; multiple matches are impossible because group keys are unique."),
				"command_path_resolution":         operation([]string{"candidate_source", "comparison", "omitted_result", "reserved_root_token", "unknown_result", "non_executable_result"}, "selection", "Resolve the supplied argv path components exactly and case-sensitively against the declared executable command inventory. Omission returns the capability index, the reserved single root token returns the executable root operation, unknown paths return the declared not-found problem, and navigational command groups return the declared non-executable problem."),
				"help_path_resolution":            operation(nil, "selection", "Resolve the longest exact, case-sensitive existing command prefix, including navigational groups, and render its help. Ignore unmatched suffix components. With no existing first component or no components, render root help; this selector does not return not-found or ambiguity problems."),
				"parameter_argument_resolution":   operation(nil, "selection", "Compare the untrimmed positional parameter selector exactly and case-sensitively against canonical parameter keys across root and groups. It conflicts with options.filter and composes with the other selector sources declared by selection_composition; zero matches produce the command's documented empty or no-op result."),
				"personalization_id_resolution":   operation(nil, "selection", "Compare the untrimmed positional personalization ID exactly and case-sensitively against canonical IDs and return personalization.not_found on zero matches."),
				"profile_name_resolution":         operation(nil, "selection", "Compare the positional existing-profile name exactly and case-sensitively against canonical profile directory names. No match returns profile.not_found; profile names are unique."),
				"project_alias_resolution":        operation(nil, "selection", "Compare the positional existing project-alias name exactly and case-sensitively against canonical repository alias keys. No match is a successful unchanged result because alias removal is idempotent; alias keys are unique."),
				"schema_id_resolution":            operation(nil, "selection", "Look up the complete schema ID by exact case-sensitive equality in the embedded schema registry and return schema.not_found on zero matches; multiple matches are impossible because schema IDs are unique."),
				"version_resolution":              operation(nil, "selection", "Resolve the untrimmed selector exactly and case-sensitively: a positive number directly; current and latest as the current publication; previous as one publication before current; current~N and latest~N as N publications before current. Live mode uses Firebase history, cached mode uses local snapshot numbers, unavailable results return version.not_found, and versions.diff defaults an omitted to argument to current."),
				"selection_composition":           operation([]string{"sources", "repeated_source_combination", "across_source_combination", "absent_source_behavior", "target_defaults"}, "selection", "Combine values within every repeated selector source as declared, combine distinct present selector sources as declared, let absent sources match all candidates, and apply each declared target default when its source is absent."),
				"case_insensitive_substring":      operation([]string{"fields", "query_normalization", "haystack_normalization", "separator"}, "boolean", "Join the declared fields with separator, normalize the query and haystack as declared, and test whether the haystack contains the query."),
				"parameter_search":                operation([]string{"normalized_fields", "raw_fields", "normalized_query", "raw_query", "match", "combination"}, "boolean", "Build both query variants. The normalized variant lowercases letters and digits, replaces every other rune with a space, and collapses Unicode whitespace; the raw variant only collapses Unicode whitespace. Match each against the corresponding joined fields and combine the results as declared."),
			},
		},
		"invariants": map[string]any{
			"keyword":   "x-fbrcm-invariants",
			"schema":    contract.SemanticRef("invariant_rules"),
			"semantics": "Evaluate each expression against the annotated DTO and require true. Fields are relative JSON member paths; item is bound to each array member by count_where.",
			"operators": map[string]any{
				"and":           operation([]string{"values"}, "boolean", "True exactly when every expression in values is true."),
				"byte_length":   operation([]string{"value"}, "integer", "Number of bytes in the byte-string operand."),
				"count_where":   operation([]string{"collection", "where"}, "integer", "Number of collection members for which where evaluates true."),
				"eq":            operation([]string{"left", "right"}, "boolean", "JSON-semantic equality of left and right."),
				"gt":            operation([]string{"left", "right"}, "boolean", "Numeric left is greater than numeric right."),
				"iff":           operation([]string{"left", "right"}, "boolean", "Left and right have the same Boolean truth value."),
				"implies":       operation([]string{"if", "then"}, "boolean", "False only when if is true and then is false."),
				"in":            operation([]string{"value", "set"}, "boolean", "Value equals at least one member of set."),
				"is_non_null":   operation([]string{"value"}, "boolean", "The value exists and is not JSON null."),
				"length":        operation([]string{"value"}, "integer", "Number of members in an array."),
				"lowercase_hex": operation([]string{"value"}, "string", "Lowercase hexadecimal encoding of the byte-string operand."),
				"sha256":        operation([]string{"value"}, "bytes", "SHA-256 digest of the byte-string operand."),
				"sum":           operation([]string{"values"}, "number", "Arithmetic sum of the numeric operands."),
			},
			"term_forms": map[string]any{
				"const":  "The literal JSON value in const.",
				"field":  "The value at the dot-separated field path in the current evaluation scope.",
				"symbol": "A contract-defined external value named by symbol.",
			},
			"symbols": map[string]any{
				"canonical_artifact_bytes":      "Canonical bytes defined by the artifact encoding and media-type rules in docs/cli-contract.md.",
				"existing_destination_replaced": "True exactly when an authorized artifact write replaced an existing destination.",
			},
		},
	}
}

func expressionLanguageMetadata() map[string]any {
	field := func(name, valueType, description string) map[string]any {
		return map[string]any{"name": name, "type": valueType, "description": description}
	}
	return map[string]any{
		"name": "expr-lang", "version": "1.17.8", "result_type": "boolean",
		"operators": []string{"==", "!=", "<", "<=", ">", ">=", "&&", "||", "!", "in", "contains", "startsWith", "endsWith", "matches", "|"},
		"functions": []string{"all", "any", "filter", "float", "int", "is_boolean", "is_empty", "is_json", "is_number", "is_string", "jq", "keys", "len", "map", "none", "string", "type", "values"},
		"contexts": map[string]any{
			"project": []any{
				field("project_id", "string", "Firebase project ID"), field("project", "string", "Firebase project display name"),
				field("conditions", "list<string>", "sorted condition names"), field("groups", "list<string>", "sorted parameter-group names"),
				field("parameters", "map<string,parameter>", "parameters keyed by parameter name"),
			},
			"parameter": []any{
				field("project_id", "string", "Firebase project ID"), field("project", "string", "Firebase project display name"),
				field("conditions", "list<string>", "sorted condition names"), field("groups", "list<string>", "sorted parameter-group names"),
				field("parameters", "map<string,parameter>", "parameters keyed by parameter name"), field("name", "string", "current parameter key"),
				field("group", "string|null", "current group; root compares equal to null and (root)"), field("default", "any", "typed default value"),
				field("value", "any", "typed default or conditional value"), field("conditionals", "map<string,any>", "typed conditional values by condition"),
			},
			"condition": []any{
				field("project_id", "string", "Firebase project ID"), field("project", "string", "Firebase project display name"),
				field("name", "string", "current condition name"), field("priority", "integer", "one-based evaluation priority"),
				field("expression", "string", "Firebase condition expression"), field("color", "string", "condition tag color or empty string"),
				field("usage_count", "integer", "number of conditional-value usages"), field("usages", "list<usage>", "usages with group, parameter, value, and value_type"),
			},
		},
		"value_typing":  map[string]any{"BOOLEAN": "bool", "NUMBER": "float", "STRING": "string", "JSON": "string"},
		"documentation": "docs/EXPR.md",
	}
}

func optionConstraints(commandID string, command *cobra.Command, publishedOptions map[string]any) []any {
	constraints := make([]any, 0)
	if contract.SupportsStatelessCommand(commandID) {
		statelessProjectSchema := "stateless_target_selector"
		if slices.Contains([]string{
			"experiments.delete", "experiments.list", "experiments.show", "personalizations.list", "personalizations.show",
			"project.open", "project.show", "rollouts.delete", "rollouts.list", "rollouts.show",
		}, commandID) {
			statelessProjectSchema = "physical_project_id"
		}
		statelessOptions := map[string]any{
			"not": map[string]any{"required": []string{"profile"}},
		}
		statelessOptionProperties := map[string]any{}
		if statelessMutationRejectsDraft(commandID) {
			statelessOptionProperties["draft"] = map[string]any{"const": false}
		}
		if slices.Contains([]string{"projects.diff", "versions.diff", "versions.export", "versions.list", "versions.show"}, commandID) {
			statelessOptionProperties["cached"] = map[string]any{"const": false}
		}
		if slices.Contains([]string{
			"conditions.list", "conditions.show", "experiments.list", "experiments.show", "groups.list",
			"personalizations.list", "personalizations.show", "project.show", "rollouts.list", "rollouts.show",
		}, commandID) {
			statelessOptionProperties["update"] = map[string]any{"const": false}
		}
		if commandID == "get" {
			statelessOptionProperties["update"] = map[string]any{"const": false}
		}
		if commandID == "projects.list" {
			statelessOptionProperties["update"] = map[string]any{"const": false}
		}
		if len(statelessOptionProperties) > 0 {
			statelessOptions["properties"] = statelessOptionProperties
		}
		statelessConstraint := map[string]any{
			"if": map[string]any{
				"properties": map[string]any{"stateless": map[string]any{"const": true}},
				"required":   []string{"stateless"},
			},
			"then": statelessOptions,
		}
		if commandID == "projects.list" {
			statelessConstraint["else"] = map[string]any{
				"properties": map[string]any{
					"profile": map[string]any{"allOf": []any{map[string]any{"$ref": "#/$defs/path_segment"}}},
				},
			}
		}
		constraints = append(constraints, optionsConstraint(statelessConstraint))
		if commandID != "projects.list" && !usesStatelessProjectOption(commandID) && !usesStatelessDualTargetArguments(commandID) {
			constraints = append(constraints, map[string]any{
				"if": map[string]any{
					"properties": map[string]any{
						"options": map[string]any{
							"properties": map[string]any{"stateless": map[string]any{"const": true}},
							"required":   []string{"stateless"},
						},
					},
					"required": []string{"options"},
				},
				"then": map[string]any{
					"properties": map[string]any{
						"arguments": map[string]any{
							"properties": map[string]any{"project": map[string]any{"$ref": "#/$defs/" + statelessProjectSchema}},
						},
					},
				},
				"else": map[string]any{
					"properties": map[string]any{
						"options": map[string]any{
							"properties": map[string]any{
								"profile": map[string]any{"allOf": []any{map[string]any{"$ref": "#/$defs/path_segment"}}},
							},
						},
					},
				},
			})
		}
		if usesStatelessProjectOption(commandID) {
			constraints = append(constraints, map[string]any{
				"if": map[string]any{
					"properties": map[string]any{
						"options": map[string]any{
							"properties": map[string]any{"stateless": map[string]any{"const": true}},
							"required":   []string{"stateless"},
						},
						"stdin": map[string]any{"type": "null"},
					},
					"required": []string{"options", "stdin"},
				},
				"then": map[string]any{
					"properties": map[string]any{
						"options": map[string]any{
							"properties": map[string]any{
								"project": map[string]any{
									"type":  "array",
									"items": map[string]any{"$ref": "#/$defs/stateless_get_project_selector"},
								},
							},
						},
					},
				},
			})
		}
		if usesStatelessDualTargetArguments(commandID) {
			constraints = append(constraints, map[string]any{
				"if": map[string]any{
					"properties": map[string]any{
						"options": map[string]any{
							"properties": map[string]any{"stateless": map[string]any{"const": true}},
							"required":   []string{"stateless"},
						},
					},
					"required": []string{"options"},
				},
				"then": map[string]any{
					"properties": map[string]any{
						"arguments": map[string]any{
							"properties": map[string]any{
								"source_project": map[string]any{"$ref": "#/$defs/stateless_target_selector"},
								"target_project": map[string]any{"$ref": "#/$defs/stateless_target_selector"},
							},
						},
					},
				},
			})
		}
	}
	for _, group := range annotationGroups(command.Flags(), "cobra_annotation_mutually_exclusive") {
		if slices.Contains(group, "json") {
			for _, name := range group {
				if name != "json" {
					if _, published := publishedOptions[name]; published {
						constraints = append(constraints, optionsConstraint(map[string]any{"not": map[string]any{"required": []string{name}}}))
					}
				}
			}
			continue
		}
		group = slices.DeleteFunc(group, func(name string) bool {
			_, published := publishedOptions[name]
			return !published
		})
		for left := range len(group) {
			for right := left + 1; right < len(group); right++ {
				constraints = append(constraints, optionsConstraint(map[string]any{"not": map[string]any{"required": []string{group[left], group[right]}}}))
			}
		}
	}
	requireWith := func(trigger string, required ...string) {
		constraints = append(constraints, optionsConstraint(map[string]any{"if": map[string]any{"required": []string{trigger}}, "then": map[string]any{"required": required}}))
	}
	requireTrueWith := func(trigger string, required ...string) {
		constraints = append(constraints, optionsConstraint(map[string]any{"if": map[string]any{"properties": map[string]any{trigger: map[string]any{"const": true}}, "required": []string{trigger}}, "then": map[string]any{"required": required}}))
	}
	requireValueSourceWithType := func() {
		constraints = append(constraints,
			optionsConstraint(map[string]any{
				"if": map[string]any{"required": []string{"type"}},
				"then": map[string]any{"anyOf": []any{
					map[string]any{"required": []string{"value"}},
					map[string]any{"properties": map[string]any{"use-in-app-default": map[string]any{"const": true}}, "required": []string{"use-in-app-default"}},
				}},
			}),
			typedValueConstraint([]string{"boolean", "bool"}, map[string]any{"pattern": `^(?:true|false)$`}),
			typedValueConstraint([]string{"number"}, map[string]any{
				"pattern":         `^(?:[+-]?(?:(?:\d(?:_?\d)*(?:\.\d(?:_?\d)*)?|\d(?:_?\d)*\.|\.\d(?:_?\d)*)(?:[eE][+-]?\d(?:_?\d)*)?|0[xX]_?(?:[0-9a-fA-F](?:_?[0-9a-fA-F])*(?:\.[0-9a-fA-F](?:_?[0-9a-fA-F])*)?|[0-9a-fA-F](?:_?[0-9a-fA-F])*\.|\.[0-9a-fA-F](?:_?[0-9a-fA-F])*)[pP][+-]?\d(?:_?\d)*|[iI][nN][fF](?:[iI][nN][iI][tT][yY])?)|[nN][aA][nN])$`,
				"x-fbrcm-grammar": "strconv.ParseFloat 64-bit number",
			}),
			typedValueConstraint([]string{"json"}, map[string]any{
				"contentMediaType": "application/json",
				"contentSchema":    map[string]any{},
				"x-fbrcm-validation": []any{map[string]any{
					"operator": "parse_json", "specification": "RFC 8259", "consume": "entire_string",
				}},
			}),
		)
	}
	switch commandID {
	case "root":
		constraints = append(constraints, map[string]any{
			"if": map[string]any{
				"properties": map[string]any{
					"options": map[string]any{
						"properties": map[string]any{"version": map[string]any{"const": false}},
						"required":   []string{"profile"},
					},
				},
				"required": []string{"options"},
			},
			"then": map[string]any{
				"properties": map[string]any{
					"options": map[string]any{
						"properties": map[string]any{
							"profile": map[string]any{"allOf": []any{map[string]any{"$ref": "#/$defs/path_segment"}}},
						},
					},
				},
			},
		})
	case "add":
		constraints = append(constraints, optionsConstraint(map[string]any{"anyOf": []any{map[string]any{"required": []string{"value"}}, map[string]any{"properties": map[string]any{"use-in-app-default": map[string]any{"const": true}}, "required": []string{"use-in-app-default"}}}}))
		requireWith("value", "type")
		requireTrueWith("use-in-app-default", "type")
		requireValueSourceWithType()
	case "update":
		requireWith("value", "type")
		requireTrueWith("use-in-app-default", "type")
		requireValueSourceWithType()
		constraints = append(constraints, optionsConstraint(map[string]any{"if": map[string]any{"required": []string{"condition"}}, "then": map[string]any{"anyOf": []any{map[string]any{"required": []string{"value"}}, map[string]any{"properties": map[string]any{"use-in-app-default": map[string]any{"const": true}}, "required": []string{"use-in-app-default"}}}}}))
		constraints = append(constraints, argumentOptionMutualExclusion("parameter", "filter"))
	case "delete", "get":
		constraints = append(constraints, argumentOptionMutualExclusion("parameter", "filter"))
		if commandID == "get" {
			constraints = append(constraints, map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"stdin": map[string]any{"not": map[string]any{"type": "null"}}},
					"required":   []string{"stdin"},
				},
				"then": map[string]any{
					"properties": map[string]any{
						"options": map[string]any{"properties": map[string]any{"update": map[string]any{"const": false}}},
					},
				},
			})
		}
	case "duplicate":
		constraints = append(constraints, map[string]any{"x-fbrcm-validation": []any{map[string]any{"operator": "fields_differ", "fields": []any{"arguments.source", "arguments.target"}, "comparison": "exact_codepoint"}}})
	case "config.set":
		constraints = append(constraints, configSetConstraint())
	case "project.import":
		constraints = append(constraints, optionsConstraint(map[string]any{
			"if": map[string]any{"required": []string{"merge-resolve"}},
			"then": map[string]any{
				"properties": map[string]any{"merge": map[string]any{"const": true}},
				"required":   []string{"merge"},
			},
		}))
	case "draft.diff":
		constraints = append(constraints, optionsConstraint(map[string]any{"if": map[string]any{"properties": map[string]any{"cached": map[string]any{"const": true}}, "required": []string{"cached"}}, "then": map[string]any{"properties": map[string]any{"against": map[string]any{"const": "current"}}, "required": []string{"against"}}}))
	case "conditions.edit":
		constraints = append(constraints,
			optionsConstraint(map[string]any{"anyOf": []any{
				map[string]any{"required": []string{"expression"}},
				map[string]any{"required": []string{"color"}},
				map[string]any{"properties": map[string]any{"no-color": map[string]any{"const": true}}, "required": []string{"no-color"}},
			}}),
			optionsConstraint(map[string]any{"not": map[string]any{
				"properties": map[string]any{"no-color": map[string]any{"const": true}},
				"required":   []string{"color", "no-color"},
			}}),
		)
	case "groups.edit":
		constraints = append(constraints,
			optionsConstraint(map[string]any{"anyOf": []any{
				map[string]any{"required": []string{"description"}},
				map[string]any{"properties": map[string]any{"no-description": map[string]any{"const": true}}, "required": []string{"no-description"}},
			}}),
			optionsConstraint(map[string]any{"not": map[string]any{
				"properties": map[string]any{"no-description": map[string]any{"const": true}},
				"required":   []string{"description", "no-description"},
			}}),
		)
	case "project.templates.set":
		constraints = append(constraints, optionsConstraint(map[string]any{"anyOf": []any{
			map[string]any{"required": []string{"templates"}},
			map[string]any{"required": []string{"primary"}},
		}}))
		for _, value := range []string{"client", "server"} {
			constraints = append(constraints, optionsConstraint(map[string]any{
				"if":   map[string]any{"properties": map[string]any{"primary": map[string]any{"pattern": caseInsensitiveEnumPattern([]string{value})}}, "required": []string{"primary", "templates"}},
				"then": map[string]any{"properties": map[string]any{"templates": map[string]any{"contains": map[string]any{"pattern": caseInsensitiveEnumPattern([]string{value})}}}},
			}))
		}
	case "draft.publish", "draft.discard":
		constraints = append(constraints,
			map[string]any{"anyOf": []any{
				map[string]any{"properties": map[string]any{"arguments": map[string]any{"required": []string{"project"}}}},
				map[string]any{"properties": map[string]any{"options": map[string]any{"properties": map[string]any{"all": map[string]any{"const": true}}, "required": []string{"all"}}}},
			}},
			map[string]any{"not": map[string]any{"properties": map[string]any{
				"arguments": map[string]any{"required": []string{"project"}},
				"options":   map[string]any{"properties": map[string]any{"all": map[string]any{"const": true}}, "required": []string{"all"}},
			}}},
		)
	case "draft.change-note":
		constraints = append(constraints, map[string]any{"not": map[string]any{"properties": map[string]any{
			"arguments": map[string]any{"required": []string{"text"}},
			"options":   map[string]any{"properties": map[string]any{"clear": map[string]any{"const": true}}, "required": []string{"clear"}},
		}}})
	}
	return constraints
}

func statelessMutationRejectsDraft(commandID string) bool {
	return slices.Contains([]string{
		"add", "delete", "duplicate", "update",
		"conditions.add", "conditions.delete", "conditions.edit", "conditions.move", "conditions.rename",
		"groups.add", "groups.delete", "groups.edit", "groups.rename",
		"project.import",
	}, commandID)
}

func usesStatelessProjectOption(commandID string) bool {
	return slices.Contains([]string{
		"add", "delete", "duplicate", "get", "update",
		"groups.add", "groups.delete", "groups.edit", "groups.list", "groups.rename",
	}, commandID)
}

func usesStatelessDualTargetArguments(commandID string) bool {
	return slices.Contains([]string{"projects.diff", "projects.promote"}, commandID)
}

func typedValueConstraint(typeNames []string, valueSchema map[string]any) map[string]any {
	return optionsConstraint(map[string]any{
		"if": map[string]any{
			"properties": map[string]any{"type": map[string]any{"pattern": caseInsensitiveEnumPattern(typeNames)}},
			"required":   []string{"type", "value"},
		},
		"then": map[string]any{"properties": map[string]any{"value": valueSchema}},
	})
}

func configSetConstraint() map[string]any {
	argumentVariant := func(key map[string]any, values map[string]any, normalize bool) map[string]any {
		if normalize {
			key["x-fbrcm-normalization"] = []any{map[string]any{
				"operator": "trim_unicode_whitespace", "source": "argv", "target": "normalized_invocation",
			}}
		}
		return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"arguments": map[string]any{
					"properties": map[string]any{"key": key, "value": values},
					"required":   []string{"key", "value"},
				},
			},
			"required": []string{"arguments"},
		}
	}
	oneBoolean := map[string]any{"type": "array", "minItems": 1, "maxItems": 1, "items": map[string]any{"enum": []string{"true", "false"}}}
	variants := []any{argumentVariant(map[string]any{"const": "powerline_glyphs"}, oneBoolean, false)}

	keyNames := keyBindingConfigNames(false)
	for index := range keyNames {
		keyNames[index] = regexp.QuoteMeta(keyNames[index])
	}
	variants = append(variants, argumentVariant(
		map[string]any{"type": "string", "pattern": "^(?:" + strings.Join(keyNames, "|") + ")$"},
		map[string]any{"type": "array", "minItems": 1, "items": map[string]any{
			"type": "string", "minLength": 1,
			"pattern": keyBindingPattern(),
			"x-fbrcm-validation": []any{map[string]any{
				"operator": "unique_tokens", "separator": "+", "range": "all_but_last",
			}},
		}},
		true,
	))
	aliasVariant := argumentVariant(
		map[string]any{"type": "string", "pattern": `^projects\.aliases\.[a-z][a-z0-9_-]{0,62}$`},
		map[string]any{"type": "array", "minItems": 1, "maxItems": 1, "items": map[string]any{"$ref": contract.SemanticRef("physical_project_id")}},
		true,
	)
	aliasVariant["properties"].(map[string]any)["options"] = map[string]any{
		"properties": map[string]any{"scope": map[string]any{"const": "local"}},
		"required":   []string{"scope"},
	}
	variants = append(variants, aliasVariant)
	return map[string]any{"oneOf": variants}
}

func argumentOptionMutualExclusion(argument, option string) map[string]any {
	return map[string]any{"not": map[string]any{
		"properties": map[string]any{
			"arguments": map[string]any{"required": []string{argument}},
			"options":   map[string]any{"required": []string{option}},
		},
		"required": []string{"arguments", "options"},
	}}
}

func keyBindingPattern() string {
	modifiers := `(?:ctrl|alt|shift|meta|hyper|super|capslock|scrolllock|numlock)`
	named := `(?:enter|esc|escape|tab|backspace|delete|insert|up|down|left|right|pgup|pgdown|home|end|space|begin|find|select|kpenter|kpequal|kpmul|kpplus|kpcomma|kpminus|kpperiod|kpdiv|kp[0-9]|kpsep|kpup|kpdown|kpleft|kpright|kppgup|kppgdown|kphome|kpend|kpinsert|kpdelete|kpbegin|capslock|scrolllock|numlock|printscreen|pause|menu|mediaplay|mediapause|mediaplaypause|mediareverse|mediastop|mediafastforward|mediarewind|medianext|mediaprev|mediarecord|lowervol|raisevol|mute|leftshift|leftalt|leftctrl|leftsuper|lefthyper|leftmeta|rightshift|rightalt|rightctrl|rightsuper|righthyper|rightmeta|isolevel3shift|isolevel5shift)`
	function := `f(?:[1-9]|[1-5]\d|6[0-3])`
	printable := `[\p{L}\p{M}\p{N}\p{P}\p{S}]`
	base := "(?:" + printable + "|" + function + "|" + named + ")"
	return "^(?: |" + base + "|(?:" + modifiers + `\+)+` + base + ")$"
}

func optionsConstraint(constraint map[string]any) map[string]any {
	return map[string]any{"properties": map[string]any{"options": constraint}}
}

func annotationGroups(flags *pflag.FlagSet, key string) [][]string {
	seen := make(map[string]bool)
	groups := make([][]string, 0)
	flags.VisitAll(func(flag *pflag.Flag) {
		for _, raw := range flag.Annotations[key] {
			if seen[raw] {
				continue
			}
			seen[raw] = true
			groups = append(groups, strings.Fields(raw))
		}
	})
	return groups
}

func remediationSchema() map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"description", "strategy", "argv"}, "properties": map[string]any{
		"description": map[string]any{"type": "string"},
		"strategy": map[string]any{"enum": []string{
			machine.RemediationRetryWithArguments,
			machine.RemediationReplaceSelector,
			machine.RemediationRunCommand,
		}},
		"argv": map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"type": "string"}},
	}}}
}

func responseSchema(capability contract.Capability, dataSchema, successDataSchema map[string]any) map[string]any {
	publishedDataSchema := map[string]any{"oneOf": []any{map[string]any{"$ref": "#/$defs/success_data"}, map[string]any{"type": "null"}}}
	if dataSchema["type"] == "null" {
		publishedDataSchema = dataSchema
	}
	isDiffCommand := slices.Contains([]string{"draft.diff", "projects.diff", "versions.diff"}, capability.ID)
	constraints := []any{
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"outcome": map[string]any{"enum": []string{"success", "partial_success"}}}, "required": []string{"outcome"}},
			"then": map[string]any{"properties": map[string]any{"data": map[string]any{"$ref": "#/$defs/success_data"}}},
		},
	}
	reachableOutcomes := []string{"success", "failure"}
	if slices.Contains(capability.SideEffects, "firebase_remote_write") {
		reachableOutcomes = []string{"success", "partial_success", "failure"}
	}
	if _, impossible := successDataSchema["not"]; impossible && len(successDataSchema) == 1 {
		reachableOutcomes = []string{"failure"}
	}
	constraints = append(constraints, map[string]any{
		"properties": map[string]any{"outcome": map[string]any{"enum": reachableOutcomes}},
	})
	warnings := commandWarningCodes(capability.ID)
	warningConstraint := map[string]any{"type": "array", "maxItems": 0}
	if len(warnings) > 0 {
		warningConstraint = map[string]any{
			"type": "array",
			"items": map[string]any{
				"properties": map[string]any{"code": map[string]any{"enum": warnings}},
				"required":   []string{"code"},
			},
		}
	}
	constraints = append(constraints, map[string]any{
		"properties": map[string]any{"warnings": warningConstraint},
	})
	if isDiffCommand {
		for _, result := range []struct {
			changed  bool
			exitCode int
		}{{changed: false, exitCode: 0}, {changed: true, exitCode: 1}} {
			constraints = append(constraints, map[string]any{
				"if": map[string]any{
					"properties": map[string]any{
						"outcome": map[string]any{"const": "success"},
						"data": map[string]any{
							"properties": map[string]any{"changed": map[string]any{"const": result.changed}},
							"required":   []string{"changed"},
						},
					},
					"required": []string{"outcome", "data"},
				},
				"then": map[string]any{"properties": map[string]any{"exit_code": map[string]any{"const": result.exitCode}}},
			})
		}
	} else {
		constraints = append(constraints, map[string]any{
			"if":   map[string]any{"properties": map[string]any{"outcome": map[string]any{"const": "success"}}, "required": []string{"outcome"}},
			"then": map[string]any{"properties": map[string]any{"exit_code": map[string]any{"enum": []int{0}}}},
		})
	}
	if slices.Contains([]string{"experiments.delete", "rollouts.delete"}, capability.ID) {
		constraints = append(constraints,
			map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"data": map[string]any{"type": "object", "properties": map[string]any{"status": map[string]any{"const": "would-delete"}}, "required": []string{"status"}}},
					"required":   []string{"data"},
				},
				"then": map[string]any{"properties": map[string]any{
					"outcome": map[string]any{"const": "failure"},
					"errors":  map[string]any{"contains": map[string]any{"properties": map[string]any{"code": map[string]any{"const": "interaction.required"}}, "required": []string{"code"}}, "minContains": 1},
				}},
			},
			map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"data": map[string]any{"type": "object", "properties": map[string]any{"status": map[string]any{"const": "deleted"}}, "required": []string{"status"}}},
					"required":   []string{"data"},
				},
				"then": map[string]any{"properties": map[string]any{"outcome": map[string]any{"const": "success"}}},
			},
		)
	}
	if !slices.Contains([]string{"config.validate", "doctor"}, capability.ID) {
		constraints = append(constraints, map[string]any{
			"if":   map[string]any{"properties": map[string]any{"outcome": map[string]any{"const": "failure"}}, "required": []string{"outcome"}},
			"then": map[string]any{"properties": map[string]any{"exit_code": map[string]any{"not": map[string]any{"const": 1}}}},
		})
	}
	return map[string]any{
		"$schema": draft,
		"$id":     capability.ResponseSchema,
		"$defs":   map[string]any{"success_data": successDataSchema},
		"allOf": []any{
			map[string]any{"$ref": contract.EnvelopeSchemaID()},
			map[string]any{
				"type":     "object",
				"required": []string{"schema", "command", "data"},
				"properties": map[string]any{
					"schema":  map[string]any{"const": capability.ResponseSchema},
					"command": map[string]any{"const": capability.ID},
					"data":    publishedDataSchema,
				},
				"allOf": constraints,
			},
		},
	}
}

func commandWarningCodes(commandID string) []string {
	postPublication := []string{"publication.cache_stale", "publication.post_publish_hook_failed"}
	switch commandID {
	case "get":
		return []string{"cache.stale"}
	case "add", "delete", "duplicate", "update",
		"groups.add", "groups.delete", "groups.edit", "groups.rename":
		return append([]string{"publication.non_atomic"}, postPublication...)
	case "conditions.add", "conditions.delete", "conditions.edit", "conditions.move", "conditions.rename",
		"project.import", "projects.promote", "versions.restore", "versions.rollback":
		return postPublication
	case "draft.publish":
		return []string{
			"publication.cache_stale",
			"publication.draft_cleanup_failed",
			"publication.non_atomic",
			"publication.post_publish_hook_failed",
		}
	default:
		return nil
	}
}
