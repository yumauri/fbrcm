package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	corelog "fbrcm/core/log"
)

type RemoteConfig struct {
	Conditions      []RemoteConfigCondition      `json:"conditions,omitempty"`
	Parameters      map[string]RemoteConfigParam `json:"parameters,omitempty"`
	ParameterGroups map[string]RemoteConfigGroup `json:"parameterGroups,omitempty"`
	Version         RemoteConfigVersion          `json:"version,omitzero"`
}

type RemoteConfigCondition struct {
	Name        string `json:"name,omitempty"`
	Expression  string `json:"expression,omitempty"`
	Description string `json:"description,omitempty"`
	TagColor    string `json:"tagColor,omitempty"`
}

type RemoteConfigGroup struct {
	Description string                       `json:"description,omitempty"`
	Parameters  map[string]RemoteConfigParam `json:"parameters,omitempty"`
}

type RemoteConfigParam struct {
	DefaultValue      *RemoteConfigValue           `json:"defaultValue,omitempty"`
	ConditionalValues map[string]RemoteConfigValue `json:"conditionalValues,omitempty"`
	Description       string                       `json:"description,omitempty"`
	ValueType         string                       `json:"valueType,omitempty"`
}

type RemoteConfigValue struct {
	Value                string          `json:"value,omitempty"`
	UseInAppDefault      bool            `json:"useInAppDefault,omitempty"`
	PersonalizationValue json.RawMessage `json:"personalizationValue,omitempty"`
	RolloutValue         json.RawMessage `json:"rolloutValue,omitempty"`
}

type RemoteConfigVersion struct {
	VersionNumber string `json:"versionNumber,omitempty"`
	UpdateTime    string `json:"updateTime,omitempty"`
	Description   string `json:"description,omitempty"`
}

type listVersionsResponse struct {
	Versions []RemoteConfigVersion `json:"versions"`
}

const notAvailableVersion = "NA"

func ParseRemoteConfig(raw json.RawMessage) (*RemoteConfig, error) {
	raw = normalizeRemoteConfigRaw(raw)
	var cfg RemoteConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("decode remote config: %w", err)
	}
	if strings.TrimSpace(cfg.Version.VersionNumber) == "" {
		cfg.Version.VersionNumber = notAvailableVersion
	}
	return &cfg, nil
}

func (s *Service) GetRemoteConfig(ctx context.Context, projectID string) (json.RawMessage, string, error) {
	logger := corelog.For("firebase")
	logger.Info("get remote config", "project_id", projectID)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://firebaseremoteconfig.googleapis.com/v1/projects/%s/remoteConfig", projectID),
		nil,
	)
	if err != nil {
		logger.Error("create remote config request failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("create remote config request: %w", err)
	}
	logHTTPRequest(logger.With("project_id", projectID), req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("remote config request failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("get remote config: %w", err)
	}
	logHTTPResponse(logger.With("project_id", projectID), req, resp)

	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		logger.Error("read remote config response failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("read remote config response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("remote config api returned non-200", "project_id", projectID, "status", resp.Status)
		return nil, "", fmt.Errorf("remote config api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	body = normalizeRemoteConfigRaw(body)
	if !json.Valid(body) {
		logger.Error("remote config response invalid json", "project_id", projectID)
		return nil, "", fmt.Errorf("remote config api returned invalid json")
	}

	etag := strings.TrimSpace(resp.Header.Get("ETag"))
	logger.Info("remote config loaded", "project_id", projectID, "etag", etag, "bytes", len(body))
	return bytes.TrimSpace(body), etag, nil
}

func (s *Service) ValidateRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage, etag string) error {
	_, _, err := s.updateRemoteConfig(ctx, projectID, raw, etag, true)
	return err
}

func (s *Service) UpdateRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error) {
	return s.updateRemoteConfig(ctx, projectID, raw, etag, false)
}

func (s *Service) updateRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage, etag string, validateOnly bool) (json.RawMessage, string, error) {
	logger := corelog.For("firebase")
	logger.Info("update remote config", "project_id", projectID, "validate_only", validateOnly)

	body := bytes.TrimSpace(raw)
	if !json.Valid(body) {
		logger.Error("remote config update payload invalid json", "project_id", projectID)
		return nil, "", fmt.Errorf("remote config payload is not valid json")
	}

	endpoint := fmt.Sprintf("https://firebaseremoteconfig.googleapis.com/v1/projects/%s/remoteConfig", projectID)
	if validateOnly {
		endpoint += "?" + url.Values{"validateOnly": []string{"true"}}.Encode()
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		logger.Error("create update remote config request failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("create update remote config request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("If-Match", strings.TrimSpace(etag))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	logHTTPRequest(logger.With("project_id", projectID), req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("update remote config request failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("update remote config: %w", err)
	}
	logHTTPResponse(logger.With("project_id", projectID), req, resp)

	respBody, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		logger.Error("read update remote config response failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("read update remote config response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("update remote config api returned non-200", "project_id", projectID, "status", resp.Status, "validate_only", validateOnly)
		action := "update"
		if validateOnly {
			action = "validate"
		}
		return nil, "", fmt.Errorf("%s remote config api returned %s: %s", action, resp.Status, strings.TrimSpace(string(respBody)))
	}

	respBody = normalizeRemoteConfigRaw(respBody)
	if !json.Valid(respBody) {
		logger.Error("update remote config response invalid json", "project_id", projectID)
		return nil, "", fmt.Errorf("update remote config api returned invalid json")
	}

	nextETag := strings.TrimSpace(resp.Header.Get("ETag"))
	logger.Info("remote config updated", "project_id", projectID, "etag", nextETag, "bytes", len(respBody), "validate_only", validateOnly)
	return bytes.TrimSpace(respBody), nextETag, nil
}

func (s *Service) GetLatestRemoteConfigVersion(ctx context.Context, projectID string) (RemoteConfigVersion, error) {
	logger := corelog.For("firebase")
	logger.Info("get latest remote config version", "project_id", projectID)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://firebaseremoteconfig.googleapis.com/v1/projects/%s/remoteConfig:listVersions?pageSize=1", projectID),
		nil,
	)
	if err != nil {
		logger.Error("create remote config version request failed", "project_id", projectID, "err", err)
		return RemoteConfigVersion{}, fmt.Errorf("create remote config version request: %w", err)
	}
	logHTTPRequest(logger.With("project_id", projectID), req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("remote config version request failed", "project_id", projectID, "err", err)
		return RemoteConfigVersion{}, fmt.Errorf("list remote config versions: %w", err)
	}
	logHTTPResponse(logger.With("project_id", projectID), req, resp)

	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		logger.Error("read remote config versions response failed", "project_id", projectID, "err", err)
		return RemoteConfigVersion{}, fmt.Errorf("read remote config versions response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("remote config versions api returned non-200", "project_id", projectID, "status", resp.Status)
		return RemoteConfigVersion{}, fmt.Errorf("remote config versions api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload listVersionsResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&payload); err != nil {
		logger.Error("decode remote config versions failed", "project_id", projectID, "err", err)
		return RemoteConfigVersion{}, fmt.Errorf("decode remote config versions response: %w", err)
	}
	if len(payload.Versions) == 0 {
		logger.Info("remote config versions empty; using NA", "project_id", projectID)
		return RemoteConfigVersion{VersionNumber: notAvailableVersion}, nil
	}
	if strings.TrimSpace(payload.Versions[0].VersionNumber) == "" {
		payload.Versions[0].VersionNumber = notAvailableVersion
	}

	logger.Info("latest remote config version loaded", "project_id", projectID, "version", payload.Versions[0].VersionNumber)
	return payload.Versions[0], nil
}

func normalizeRemoteConfigRaw(raw json.RawMessage) json.RawMessage {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("{}")) {
		return json.RawMessage(`{"version":{"versionNumber":"NA"}}`)
	}
	trimmed = bytes.ReplaceAll(trimmed, []byte(`\u003c`), []byte("<"))
	trimmed = bytes.ReplaceAll(trimmed, []byte(`\u003e`), []byte(">"))
	trimmed = bytes.ReplaceAll(trimmed, []byte(`\u0026`), []byte("&"))
	trimmed = bytes.ReplaceAll(trimmed, []byte(`\u003C`), []byte("<"))
	trimmed = bytes.ReplaceAll(trimmed, []byte(`\u003E`), []byte(">"))
	trimmed = bytes.ReplaceAll(trimmed, []byte(`\u0026`), []byte("&"))
	return trimmed
}
