package core

import (
	"context"
	"sync"

	"github.com/yumauri/fbrcm/core/firebase"
)

// OAuthAuthorizationEvent reports an interactive OAuth flow to a host UI.
type OAuthAuthorizationEvent struct {
	FlowID uint64
	AuthID string
	URL    string
	Done   bool
	Err    error
	Cancel context.CancelFunc
}

// ConfigureOAuthAuthorization controls interactive OAuth presentation.
// CLI callers use the default browser and terminal behavior; the TUI installs
// an observer and presents authorization itself.
func (s *Core) ConfigureOAuthAuthorization(autoOpen bool, observer func(OAuthAuthorizationEvent)) {
	s.oauthMu.Lock()
	s.oauthConfigured = true
	s.oauthAutoOpen = autoOpen
	s.oauthObserver = observer
	s.oauthMu.Unlock()
}

func (s *Core) oauthAuthorizationContext(
	parent context.Context,
	authID string,
	requestedAutoOpen bool,
) (context.Context, bool) {
	s.oauthMu.RLock()
	configured := s.oauthConfigured
	autoOpen := s.oauthAutoOpen
	observer := s.oauthObserver
	s.oauthMu.RUnlock()
	if !configured {
		return parent, requestedAutoOpen
	}

	ctx := firebase.WithOAuthTerminalOutput(parent, false)
	if observer == nil {
		return ctx, autoOpen
	}
	var flowMu sync.Mutex
	var flowID uint64
	ctx = firebase.WithOAuthAuthorizationObserver(ctx, func(event firebase.OAuthAuthorizationEvent) {
		flowMu.Lock()
		if !event.Done {
			flowID = s.oauthFlowID.Add(1)
		}
		currentFlowID := flowID
		if event.Done {
			flowID = 0
		}
		flowMu.Unlock()
		observer(OAuthAuthorizationEvent{
			FlowID: currentFlowID,
			AuthID: authID,
			URL:    event.URL,
			Done:   event.Done,
			Err:    event.Err,
			Cancel: event.Cancel,
		})
	})
	return ctx, autoOpen
}
