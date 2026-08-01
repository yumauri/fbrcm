package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

type Metadata struct {
	Operation  string
	Target     string
	Current    json.RawMessage
	Candidate  json.RawMessage
	DryRun     bool
	ChangeNote string
}

type ContextFile struct {
	SchemaVersion int     `json:"schema_version"`
	Event         Event   `json:"event"`
	Operation     string  `json:"operation"`
	Target        string  `json:"target"`
	ProjectID     string  `json:"project_id"`
	TemplateKind  string  `json:"template_kind"`
	Profile       string  `json:"profile"`
	DryRun        bool    `json:"dry_run"`
	ChangeNote    string  `json:"change_note"`
	CurrentFile   string  `json:"current_file"`
	CandidateFile string  `json:"candidate_file"`
	PublishedFile *string `json:"published_file"`
}

type Error struct {
	Event    Event
	Index    int
	Command  string
	ExitCode int
	TimedOut bool
	Err      error
	Output   string
}

func (e *Error) Error() string {
	position := fmt.Sprintf("%s hook %d", e.Event, e.Index+1)
	message := ""
	if e.TimedOut {
		message = fmt.Sprintf("%s timed out: %s", position, e.Command)
	} else if e.ExitCode >= 0 {
		message = fmt.Sprintf("%s failed with exit code %d: %s", position, e.ExitCode, e.Command)
	} else {
		message = fmt.Sprintf("%s failed: %s: %v", position, e.Command, e.Err)
	}
	if strings.TrimSpace(e.Output) != "" {
		message += "\nHook output:\n" + strings.TrimSpace(e.Output)
	}
	return message
}

func (e *Error) Unwrap() error { return e.Err }

type Session struct {
	resolution    Resolution
	metadata      Metadata
	tempDir       string
	currentFile   string
	candidateFile string
	contextFile   string
	publishedFile string
	output        io.Writer
}

func Prepare(metadata Metadata, output io.Writer) (*Session, error) {
	resolution, err := Resolve()
	if err != nil {
		return nil, fmt.Errorf("resolve hooks: %w", err)
	}
	session := &Session{resolution: resolution, metadata: metadata, output: output}
	if len(resolution.Hooks.PrePublish) == 0 && len(resolution.Hooks.PostPublish) == 0 {
		return session, nil
	}
	if err := resolution.RequireTrust(PrePublish); err != nil {
		return nil, err
	}
	if err := resolution.RequireTrust(PostPublish); err != nil {
		return nil, err
	}
	tempDir, err := os.MkdirTemp("", "fbrcm-hooks-*")
	if err != nil {
		return nil, fmt.Errorf("create hook context directory: %w", err)
	}
	session.tempDir = tempDir
	if err := os.Chmod(tempDir, config.PrivateDirMode); err != nil {
		session.Close()
		return nil, err
	}
	session.currentFile = filepath.Join(tempDir, "current.json")
	session.candidateFile = filepath.Join(tempDir, "candidate.json")
	session.contextFile = filepath.Join(tempDir, "context.json")
	session.publishedFile = filepath.Join(tempDir, "published.json")
	current := metadata.Current
	if len(current) == 0 {
		current = json.RawMessage("{}")
	}
	if err := writeReadOnlyJSON(session.currentFile, current); err != nil {
		session.Close()
		return nil, err
	}
	if err := writeReadOnlyJSON(session.candidateFile, metadata.Candidate); err != nil {
		session.Close()
		return nil, err
	}
	return session, nil
}

func (s *Session) Close() {
	if s != nil && s.tempDir != "" {
		_ = os.RemoveAll(s.tempDir)
	}
}

func (s *Session) Run(ctx context.Context, event Event, published json.RawMessage) error {
	if s == nil {
		return nil
	}
	commands := s.resolution.Commands(event)
	if len(commands) == 0 {
		return nil
	}
	if err := s.resolution.RequireTrust(event); err != nil {
		return err
	}
	var publishedPointer *string
	if event == PostPublish {
		if err := writeReadOnlyJSON(s.publishedFile, published); err != nil {
			return err
		}
		publishedPointer = &s.publishedFile
	}
	target, err := rctarget.Parse(s.metadata.Target)
	if err != nil {
		return err
	}
	contextData := ContextFile{
		SchemaVersion: 1, Event: event, Operation: s.metadata.Operation,
		Target: target.String(), ProjectID: target.ProjectID, TemplateKind: string(target.Kind),
		Profile: config.GetActiveProfileName(), DryRun: s.metadata.DryRun, ChangeNote: s.metadata.ChangeNote,
		CurrentFile: s.currentFile, CandidateFile: s.candidateFile, PublishedFile: publishedPointer,
	}
	contextRaw, err := json.MarshalIndent(contextData, "", "  ")
	if err != nil {
		return err
	}
	if err := writePrivateJSON(s.contextFile, contextRaw); err != nil {
		return err
	}
	source, _ := s.resolution.Source(event)
	workingDir := filepath.Dir(source)
	for index, command := range commands {
		if s.output != nil {
			_, _ = fmt.Fprintf(s.output, "Running %s hook %d/%d: %s\n", event, index+1, len(commands), command)
		} else {
			corelog.For("hooks").Info("running hook", "event", event, "index", index+1, "count", len(commands), "command", command)
		}
		hookCtx, cancel := context.WithTimeout(ctx, s.resolution.Timeout)
		cmd := shellCommand(hookCtx, command)
		cmd.Dir = workingDir
		cmd.Env = overlayEnvironment(os.Environ(), s.environment(contextData, source))
		tail := &tailWriter{limit: 64 * 1024}
		if s.output != nil {
			writer := io.MultiWriter(s.output, tail)
			cmd.Stdout, cmd.Stderr = writer, writer
		} else {
			writer := io.MultiWriter(&logWriter{}, tail)
			cmd.Stdout, cmd.Stderr = writer, writer
		}
		err := cmd.Run()
		timedOut := errors.Is(hookCtx.Err(), context.DeadlineExceeded)
		cancel()
		if err != nil {
			exitCode := -1
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				exitCode = exitErr.ExitCode()
			}
			return &Error{Event: event, Index: index, Command: command, ExitCode: exitCode, TimedOut: timedOut, Err: err, Output: tail.String()}
		}
	}
	return nil
}

func (s *Session) environment(data ContextFile, source string) []string {
	published := ""
	if data.PublishedFile != nil {
		published = *data.PublishedFile
	}
	return []string{
		"FBRCM_HOOK_EVENT=" + string(data.Event),
		"FBRCM_OPERATION=" + data.Operation,
		"FBRCM_TARGET=" + data.Target,
		"FBRCM_PROJECT_ID=" + data.ProjectID,
		"FBRCM_TEMPLATE_KIND=" + data.TemplateKind,
		"FBRCM_PROFILE=" + data.Profile,
		"FBRCM_DRY_RUN=" + strconv.FormatBool(data.DryRun),
		"FBRCM_CHANGE_NOTE=" + data.ChangeNote,
		"FBRCM_CONFIG_FILE=" + source,
		"FBRCM_PROJECT_DIR=" + filepath.Dir(source),
		"FBRCM_CURRENT_FILE=" + data.CurrentFile,
		"FBRCM_CANDIDATE_FILE=" + data.CandidateFile,
		"FBRCM_CONTEXT_FILE=" + s.contextFile,
		"FBRCM_PUBLISHED_FILE=" + published,
		"GCLOUD_PROJECT=" + data.ProjectID,
		"PROJECT_DIR=" + filepath.Dir(source),
	}
}

func overlayEnvironment(base, overrides []string) []string {
	keys := make(map[string]struct{}, len(overrides))
	for _, entry := range overrides {
		key, _, _ := strings.Cut(entry, "=")
		if runtime.GOOS == "windows" {
			key = strings.ToUpper(key)
		}
		keys[key] = struct{}{}
	}
	result := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		if runtime.GOOS == "windows" {
			key = strings.ToUpper(key)
		}
		if _, replaced := keys[key]; !replaced {
			result = append(result, entry)
		}
	}
	return append(result, overrides...)
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd.exe", "/S", "/C", command)
	}
	return exec.CommandContext(ctx, "/bin/sh", "-c", command)
}

func writeReadOnlyJSON(path string, raw []byte) error {
	if !json.Valid(raw) {
		return fmt.Errorf("hook input %s is not valid JSON", filepath.Base(path))
	}
	if err := os.WriteFile(path, raw, 0o400); err != nil {
		return err
	}
	return os.Chmod(path, 0o400)
}

func writePrivateJSON(path string, raw []byte) error {
	if err := os.WriteFile(path, append(raw, '\n'), config.PrivateFileMode); err != nil {
		return err
	}
	return os.Chmod(path, config.PrivateFileMode)
}

type logWriter struct{}

func (*logWriter) Write(raw []byte) (int, error) {
	text := strings.TrimSpace(string(raw))
	if text != "" {
		corelog.For("hooks").Info("hook output", "output", text)
	}
	return len(raw), nil
}

// MetadataFromContext fills the common publication metadata.
func MetadataFromContext(ctx context.Context, target string, current, candidate json.RawMessage) Metadata {
	changeNote, _ := firebase.ChangeNoteFromContext(ctx)
	return Metadata{Operation: OperationFromContext(ctx), Target: target, Current: current, Candidate: candidate, DryRun: firebase.IsDryRun(ctx), ChangeNote: changeNote}
}

var _ io.Writer = (*logWriter)(nil)

type tailWriter struct {
	limit int
	data  []byte
}

func (w *tailWriter) Write(raw []byte) (int, error) {
	w.data = append(w.data, raw...)
	if len(w.data) > w.limit {
		w.data = append([]byte(nil), w.data[len(w.data)-w.limit:]...)
	}
	return len(raw), nil
}

func (w *tailWriter) String() string { return string(bytes.TrimSpace(w.data)) }
