package firebase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	corelog "github.com/yumauri/fbrcm/core/log"
)

// DefaultsFormat is a Firebase Remote Config defaults download format.
type DefaultsFormat string

const (
	DefaultsFormatJSON  DefaultsFormat = "JSON"
	DefaultsFormatXML   DefaultsFormat = "XML"
	DefaultsFormatPlist DefaultsFormat = "PLIST"
)

// ParseDefaultsFormat parses a case-insensitive defaults format name.
func ParseDefaultsFormat(value string) (DefaultsFormat, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case string(DefaultsFormatJSON):
		return DefaultsFormatJSON, nil
	case string(DefaultsFormatXML):
		return DefaultsFormatXML, nil
	case string(DefaultsFormatPlist):
		return DefaultsFormatPlist, nil
	default:
		return "", fmt.Errorf("unsupported defaults format %q (allowed: json, xml, plist)", value)
	}
}

// DownloadRemoteConfigDefaults downloads the client-side template defaults.
func (s *Service) DownloadRemoteConfigDefaults(ctx context.Context, projectID string, format DefaultsFormat) ([]byte, error) {
	logger := corelog.For("firebase")
	parsedFormat, err := ParseDefaultsFormat(string(format))
	if err != nil {
		return nil, err
	}
	logger.Info("download remote config defaults", "project_id", projectID, "format", parsedFormat)

	endpoint := fmt.Sprintf("https://firebaseremoteconfig.googleapis.com/v1/projects/%s/remoteConfig:downloadDefaults", projectID)
	endpoint += "?" + url.Values{"format": []string{string(parsedFormat)}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create remote config defaults request: %w", err)
	}
	s.setQuotaProject(req, projectID)
	logHTTPRequest(logger.With("project_id", projectID), req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("remote config defaults request failed", "project_id", projectID, "format", parsedFormat, "err", err)
		return nil, fmt.Errorf("download remote config defaults: %w", err)
	}
	logHTTPResponse(logger.With("project_id", projectID), req, resp)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read remote config defaults response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("remote config defaults api returned non-200", "project_id", projectID, "format", parsedFormat, "status", resp.Status)
		return nil, fmt.Errorf("remote config defaults api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	logger.Info("remote config defaults downloaded", "project_id", projectID, "format", parsedFormat, "bytes", len(body))
	return body, nil
}
