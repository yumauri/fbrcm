package core

import (
	"context"
	"encoding/json"
	"io"

	corehooks "github.com/yumauri/fbrcm/core/hooks"
)

// SetHookOutput selects the stream used for hook announcements and process
// output. A nil writer sends output through the core logger, which is suitable
// for the TUI Logs panel.
func (s *Core) SetHookOutput(output io.Writer) {
	s.hookMu.Lock()
	s.hookOutput = output
	s.hookMu.Unlock()
}

func (s *Core) hookOutputWriter() io.Writer {
	s.hookMu.RLock()
	defer s.hookMu.RUnlock()
	return s.hookOutput
}

func (s *Core) preparePublicationHooks(ctx context.Context, target string, current, candidate json.RawMessage) (*corehooks.Session, error) {
	if !ExecutionPolicyFromContext(ctx).RunHooks {
		return nil, nil
	}
	return corehooks.Prepare(corehooks.MetadataFromContext(ctx, target, current, candidate), s.hookOutputWriter())
}
