package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// RemoteConfigPublisher validates and publishes Remote Config payloads.
type RemoteConfigPublisher interface {
	ValidateRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) error
	PublishRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error)
}

// ValidateAndPublishRemoteConfig validates, publishes, and reports whether callers should retry.
func ValidateAndPublishRemoteConfig(ctx context.Context, publisher RemoteConfigPublisher, projectID string, raw json.RawMessage, etag, operation string, errOut io.Writer) (bool, error) {
	if err := publisher.ValidateRemoteConfigWithETag(ctx, projectID, raw, etag); err != nil {
		if IsRemoteConfigConflict(err) {
			writeRemoteConfigRetry(errOut, operation, projectID)
			return true, nil
		}
		return false, err
	}
	if _, _, err := publisher.PublishRemoteConfigWithETag(ctx, projectID, raw, etag); err != nil {
		if IsRemoteConfigConflict(err) {
			writeRemoteConfigRetry(errOut, operation, projectID)
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func writeRemoteConfigRetry(out io.Writer, operation, projectID string) {
	if out == nil {
		return
	}
	_, _ = fmt.Fprintf(out, "remote config changed during %s; restarting project %s\n", operation, projectID)
}
