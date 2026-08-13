package machine

import (
	"context"
	"sync"
)

type Warning struct {
	Code        string
	Message     string
	Target      string
	Details     any
	Remediation []Remediation
}

type State struct {
	mu       sync.Mutex
	warnings []Warning
	started  bool
}

type stateKey struct{}

func WithState(ctx context.Context) context.Context {
	return context.WithValue(ctx, stateKey{}, &State{})
}

func FromContext(ctx context.Context) *State {
	if ctx == nil {
		return nil
	}
	state, _ := ctx.Value(stateKey{}).(*State)
	return state
}

func (s *State) MarkRun() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
}

func (s *State) RunStarted() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

func (s *State) AddWarning(warning Warning) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.warnings = append(s.warnings, warning)
	s.mu.Unlock()
}

func (s *State) Warnings() []Warning {
	if s == nil {
		return []Warning{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Warning(nil), s.warnings...)
}
