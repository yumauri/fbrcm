// Package publication defines immutable, self-contained Remote Config
// publication plans. It deliberately contains no authentication, filesystem,
// or Firebase client logic so plans can be inspected offline.
package publication

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
)

const (
	Kind          = "fbrcm-publication-plan"
	FormatVersion = 1
	IDPrefix      = "sha256:"
	SchemaID      = "urn:fbrcm:schema:cli:1.0.0:publication_plan"
)

type Action string

const (
	ActionNone    Action = "none"
	ActionPublish Action = "publish"
)

type Plan struct {
	Kind          string    `json:"kind"`
	FormatVersion int       `json:"format_version"`
	PlanID        string    `json:"plan_id"`
	CreatedAt     time.Time `json:"created_at"`
	Producer      Producer  `json:"producer"`
	Operation     Operation `json:"operation"`
	Execution     Execution `json:"execution"`
	Targets       []Target  `json:"targets"`
}

type Producer struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Operation struct {
	CommandID  string          `json:"command_id"`
	ChangeNote *string         `json:"change_note"`
	Selection  json.RawMessage `json:"selection,omitempty"`
}

type Execution struct {
	Policy               string `json:"policy"`
	HooksEnabled         bool   `json:"hooks_enabled"`
	HookDefinitionSHA256 string `json:"hook_definition_sha256,omitempty"`
}

type Target struct {
	Target     string     `json:"target"`
	ProjectID  string     `json:"project_id"`
	Template   string     `json:"template_kind"`
	Action     Action     `json:"action"`
	ChangeNote *string    `json:"change_note,omitempty"`
	Base       Snapshot   `json:"base"`
	Candidate  Snapshot   `json:"candidate"`
	Validation Validation `json:"validation"`
	Source     Source     `json:"source"`
}

type Snapshot struct {
	Version      string          `json:"version,omitempty"`
	ETag         string          `json:"etag,omitempty"`
	SHA256       string          `json:"sha256"`
	RemoteConfig json.RawMessage `json:"remote_config"`
}

type Validation struct {
	Source      string    `json:"source"`
	ValidatedAt time.Time `json:"validated_at"`
}

type Source struct {
	Kind        string `json:"kind"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// FileArtifact describes the exact private plan bytes written to disk without
// embedding the sensitive plan document in a machine response.
type FileArtifact struct {
	MediaType   string `json:"media_type"`
	Encoding    string `json:"encoding"`
	Destination string `json:"destination"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	Overwritten bool   `json:"overwritten"`
}

// NewFileArtifact returns metadata for one exclusively created publication
// plan file.
func NewFileArtifact(destination string, raw []byte) FileArtifact {
	digest := sha256.Sum256(raw)
	return FileArtifact{
		MediaType:   "application/vnd.fbrcm.publication-plan+json",
		Encoding:    "none",
		Destination: destination,
		SizeBytes:   int64(len(raw)),
		SHA256:      hex.EncodeToString(digest[:]),
		Overwritten: false,
	}
}

type ErrorKind string

const (
	ErrorInvalid            ErrorKind = "invalid"
	ErrorIntegrity          ErrorKind = "integrity"
	ErrorUnsupportedVersion ErrorKind = "unsupported_version"
)

type Error struct {
	Kind ErrorKind
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

// New creates an unsealed plan. Seal must be called after all targets have
// been appended.
func New(version, commandID, policy string, changeNote *string) *Plan {
	return &Plan{
		Kind:          Kind,
		FormatVersion: FormatVersion,
		CreatedAt:     time.Now().UTC(),
		Producer:      Producer{Name: "fbrcm", Version: version},
		Operation:     Operation{CommandID: commandID, ChangeNote: changeNote},
		Execution:     Execution{Policy: policy},
		Targets:       []Target{},
	}
}

// Seal normalizes target order, fills snapshot hashes, and computes the plan
// identifier over every other serialized field.
func Seal(plan *Plan) error {
	if plan == nil {
		return invalid(errors.New("plan is nil"))
	}
	for index := range plan.Targets {
		target := &plan.Targets[index]
		baseHash, err := RemoteConfigDigest(target.Base.RemoteConfig)
		if err != nil {
			return invalid(fmt.Errorf("target %s base: %w", target.Target, err))
		}
		candidateHash, err := RemoteConfigDigest(target.Candidate.RemoteConfig)
		if err != nil {
			return invalid(fmt.Errorf("target %s candidate: %w", target.Target, err))
		}
		target.Base.SHA256 = baseHash
		target.Candidate.SHA256 = candidateHash
	}
	slices.SortFunc(plan.Targets, func(a, b Target) int { return strings.Compare(a.Target, b.Target) })
	plan.PlanID = ""
	digest, err := planDigest(plan)
	if err != nil {
		return err
	}
	plan.PlanID = IDPrefix + digest
	return Validate(plan)
}

// Validate checks structural, template, snapshot, and plan integrity.
func Validate(plan *Plan) error {
	if plan == nil {
		return invalid(errors.New("plan is nil"))
	}
	if plan.Kind != Kind {
		return invalid(fmt.Errorf("unexpected plan kind %q", plan.Kind))
	}
	if plan.FormatVersion != FormatVersion {
		return &Error{Kind: ErrorUnsupportedVersion, Err: fmt.Errorf("unsupported publication plan format version %d", plan.FormatVersion)}
	}
	if plan.CreatedAt.IsZero() {
		return invalid(errors.New("plan created_at is required"))
	}
	if strings.TrimSpace(plan.Producer.Name) == "" || strings.TrimSpace(plan.Operation.CommandID) == "" {
		return invalid(errors.New("plan producer name and operation command_id are required"))
	}
	if plan.Execution.Policy != "stateful" && plan.Execution.Policy != "stateless" {
		return invalid(fmt.Errorf("unsupported execution policy %q", plan.Execution.Policy))
	}
	if plan.Execution.HooksEnabled && strings.TrimSpace(plan.Execution.HookDefinitionSHA256) == "" {
		return invalid(errors.New("hook_definition_sha256 is required when hooks are enabled"))
	}
	if plan.Execution.HookDefinitionSHA256 != "" && !validSHA256(plan.Execution.HookDefinitionSHA256) {
		return invalid(errors.New("hook_definition_sha256 must be a lowercase SHA-256 digest"))
	}
	if len(plan.Targets) == 0 {
		return invalid(errors.New("publication plan must contain at least one target"))
	}
	seen := make(map[string]struct{}, len(plan.Targets))
	previous := ""
	for _, target := range plan.Targets {
		if strings.TrimSpace(target.Target) == "" || strings.TrimSpace(target.ProjectID) == "" {
			return invalid(errors.New("plan target and project_id are required"))
		}
		if target.Template != "client" && target.Template != "server" {
			return invalid(fmt.Errorf("target %s has unsupported template kind %q", target.Target, target.Template))
		}
		if target.Action != ActionNone && target.Action != ActionPublish {
			return invalid(fmt.Errorf("target %s has unsupported action %q", target.Target, target.Action))
		}
		if _, ok := seen[target.Target]; ok {
			return invalid(fmt.Errorf("duplicate plan target %s", target.Target))
		}
		seen[target.Target] = struct{}{}
		if previous != "" && previous > target.Target {
			return invalid(errors.New("plan targets are not in canonical order"))
		}
		previous = target.Target
		if err := validateSnapshot(target.Target, "base", target.Base); err != nil {
			return err
		}
		if err := validateSnapshot(target.Target, "candidate", target.Candidate); err != nil {
			return err
		}
		if target.Action == ActionNone && target.Base.SHA256 != target.Candidate.SHA256 {
			return invalid(fmt.Errorf("unchanged target %s has different base and candidate templates", target.Target))
		}
		if target.Action == ActionPublish && strings.TrimSpace(target.Base.ETag) == "" {
			return invalid(fmt.Errorf("publish target %s is missing its base ETag", target.Target))
		}
		if target.Validation.Source != "local" && target.Validation.Source != "firebase" {
			return invalid(fmt.Errorf("target %s has unsupported validation source %q", target.Target, target.Validation.Source))
		}
		if target.Action == ActionNone && target.Validation.Source != "local" {
			return invalid(fmt.Errorf("unchanged target %s must have local validation provenance", target.Target))
		}
		if target.Action == ActionPublish && target.Validation.Source != "firebase" {
			return invalid(fmt.Errorf("publish target %s must have Firebase validation provenance", target.Target))
		}
		if target.Validation.ValidatedAt.IsZero() {
			return invalid(fmt.Errorf("target %s validation provenance is incomplete", target.Target))
		}
		if strings.TrimSpace(target.Source.Kind) == "" {
			return invalid(fmt.Errorf("target %s source kind is required", target.Target))
		}
	}
	if !strings.HasPrefix(plan.PlanID, IDPrefix) {
		return &Error{Kind: ErrorIntegrity, Err: errors.New("publication plan ID is missing or malformed")}
	}
	want := plan.PlanID
	clone := *plan
	clone.PlanID = ""
	digest, err := planDigest(&clone)
	if err != nil {
		return err
	}
	if want != IDPrefix+digest {
		return &Error{Kind: ErrorIntegrity, Err: errors.New("publication plan integrity check failed")}
	}
	return nil
}

func validateSnapshot(target, label string, snapshot Snapshot) error {
	if strings.TrimSpace(snapshot.SHA256) == "" || len(snapshot.RemoteConfig) == 0 {
		return invalid(fmt.Errorf("target %s %s snapshot is incomplete", target, label))
	}
	digest, err := RemoteConfigDigest(snapshot.RemoteConfig)
	if err != nil {
		return invalid(fmt.Errorf("target %s %s: %w", target, label, err))
	}
	if digest != snapshot.SHA256 {
		return &Error{Kind: ErrorIntegrity, Err: fmt.Errorf("target %s %s template digest does not match", target, label)}
	}
	return nil
}

func invalid(err error) error { return &Error{Kind: ErrorInvalid, Err: err} }

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

// Marshal serializes a validated plan with stable indentation and a newline.
func Marshal(plan *Plan) ([]byte, error) {
	if err := Validate(plan); err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return nil, invalid(fmt.Errorf("encode publication plan: %w", err))
	}
	return append(raw, '\n'), nil
}

// Parse strictly decodes and validates one publication plan document.
func Parse(raw []byte) (*Plan, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var plan Plan
	if err := decoder.Decode(&plan); err != nil {
		return nil, invalid(fmt.Errorf("decode publication plan: %w", err))
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("unexpected trailing JSON value")
		}
		return nil, invalid(fmt.Errorf("decode publication plan: %w", err))
	}
	if err := Validate(&plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

// RemoteConfigDigest returns the SHA-256 of the canonical normalized template.
func RemoteConfigDigest(raw json.RawMessage) (string, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		if err == nil {
			err = errors.New("remote config must be a JSON object")
		}
		return "", fmt.Errorf("normalize Remote Config: %w", err)
	}
	normalized, err := firebase.PrepareRemoteConfigUpdate(raw)
	if err != nil {
		return "", fmt.Errorf("normalize Remote Config: %w", err)
	}
	digest := sha256.Sum256(normalized)
	return hex.EncodeToString(digest[:]), nil
}

func planDigest(plan *Plan) (string, error) {
	raw, err := json.Marshal(plan)
	if err != nil {
		return "", invalid(fmt.Errorf("encode publication plan for digest: %w", err))
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}
