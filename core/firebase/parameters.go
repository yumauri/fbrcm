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

	corelog "github.com/yumauri/fbrcm/core/log"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

type RemoteConfig struct {
	Conditions      []RemoteConfigCondition      `json:"conditions,omitempty"`
	Parameters      map[string]RemoteConfigParam `json:"parameters,omitempty"`
	ParameterGroups map[string]RemoteConfigGroup `json:"parameterGroups,omitempty"`
	Version         RemoteConfigVersion          `json:"version,omitzero"`
}

// UnmarshalJSON keeps the permissive forward-field behavior of encoding/json
// while rejecting null where the published Remote Config contract requires a
// concrete object.
func (c *RemoteConfig) UnmarshalJSON(data []byte) error {
	if err := rejectNullJSONFields(data, "remote config", "version"); err != nil {
		return err
	}
	type wireRemoteConfig RemoteConfig
	var decoded wireRemoteConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*c = RemoteConfig(decoded)
	return nil
}

type RemoteConfigCondition struct {
	Name       string `json:"name,omitempty"`
	Expression string `json:"expression,omitempty"`
	TagColor   string `json:"tagColor,omitempty"`
}

// ConditionTagColorsByName indexes Remote Config condition display colors.
func ConditionTagColorsByName(conditions []RemoteConfigCondition) map[string]string {
	colors := make(map[string]string, len(conditions))
	for _, condition := range conditions {
		colors[condition.Name] = condition.TagColor
	}
	return colors
}

// UnmarshalJSON rejects fields outside Firebase's condition schema instead of
// silently discarding unsupported condition metadata.
func (c *RemoteConfigCondition) UnmarshalJSON(data []byte) error {
	if err := rejectNullJSONFields(data, "remote config condition", "name", "expression", "tagColor"); err != nil {
		return err
	}
	type wireCondition RemoteConfigCondition
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded wireCondition
	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("decode condition: %w", err)
	}
	*c = RemoteConfigCondition(decoded)
	return nil
}

type RemoteConfigGroup struct {
	Description string                       `json:"description,omitempty"`
	Parameters  map[string]RemoteConfigParam `json:"parameters,omitempty"`
}

func (g *RemoteConfigGroup) UnmarshalJSON(data []byte) error {
	if err := rejectNullJSONFields(data, "remote config parameter group", "description"); err != nil {
		return err
	}
	type wireGroup RemoteConfigGroup
	var decoded wireGroup
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*g = RemoteConfigGroup(decoded)
	return nil
}

type RemoteConfigParam struct {
	DefaultValue      *RemoteConfigValue           `json:"defaultValue,omitempty"`
	ConditionalValues map[string]RemoteConfigValue `json:"conditionalValues,omitempty"`
	Description       string                       `json:"description,omitempty"`
	ValueType         string                       `json:"valueType,omitempty"`
}

func (p *RemoteConfigParam) UnmarshalJSON(data []byte) error {
	if err := rejectNullJSONFields(data, "remote config parameter", "description", "valueType"); err != nil {
		return err
	}
	type wireParam RemoteConfigParam
	var decoded wireParam
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = RemoteConfigParam(decoded)
	return nil
}

type RemoteConfigValue struct {
	Value                string          `json:"-"`
	UseInAppDefault      bool            `json:"-"`
	PersonalizationValue json.RawMessage `json:"-"`
	ExperimentValue      json.RawMessage `json:"-"`
	RolloutValue         json.RawMessage `json:"-"`
	UnknownValueOption   string          `json:"-"`
	UnknownValue         json.RawMessage `json:"-"`
}

type RemoteConfigVersion struct {
	VersionNumber  string           `json:"versionNumber,omitempty"`
	UpdateTime     string           `json:"updateTime,omitempty" contract:"format=date-time"`
	UpdateUser     RemoteConfigUser `json:"updateUser,omitzero"`
	ChangeNote     string           `json:"description,omitempty"`
	UpdateOrigin   string           `json:"updateOrigin,omitempty"`
	UpdateType     string           `json:"updateType,omitempty"`
	RollbackSource string           `json:"rollbackSource,omitempty"`
	IsLegacy       bool             `json:"isLegacy,omitempty"`
}

func (v *RemoteConfigVersion) UnmarshalJSON(data []byte) error {
	if err := rejectNullJSONFields(data, "remote config version", "versionNumber", "updateTime", "updateUser", "description", "updateOrigin", "updateType", "rollbackSource", "isLegacy"); err != nil {
		return err
	}
	type wireVersion RemoteConfigVersion
	var decoded wireVersion
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*v = RemoteConfigVersion(decoded)
	return nil
}

type RemoteConfigUser struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty" contract:"format=email"`
	ImageURL string `json:"imageUrl,omitempty" contract:"format=uri"`
}

func (u *RemoteConfigUser) UnmarshalJSON(data []byte) error {
	if err := rejectNullJSONFields(data, "remote config user", "name", "email", "imageUrl"); err != nil {
		return err
	}
	type wireUser RemoteConfigUser
	var decoded wireUser
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*u = RemoteConfigUser(decoded)
	return nil
}

func rejectNullJSONFields(data []byte, objectName string, fields ...string) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("decode %s: expected object, got null", objectName)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return fmt.Errorf("decode %s: %w", objectName, err)
	}
	if object == nil {
		return fmt.Errorf("decode %s: expected object", objectName)
	}
	for _, field := range fields {
		raw, ok := object[field]
		if ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("decode %s: field %q cannot be null", objectName, field)
		}
	}
	return nil
}

type listVersionsResponse struct {
	Versions      []RemoteConfigVersion `json:"versions"`
	NextPageToken string                `json:"nextPageToken,omitempty"`
}

type ListVersionsOptions struct {
	PageSize         int
	PageToken        string
	EndVersionNumber string
	StartTime        string
	EndTime          string
}

type RemoteConfigVersionsPage struct {
	Versions      []RemoteConfigVersion
	NextPageToken string
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

func (s *Service) GetRemoteConfig(ctx context.Context, projectID string, versionNumber ...string) (json.RawMessage, string, error) {
	logger := corelog.For("firebase")
	resource, targetProjectID, err := remoteConfigResource(projectID)
	if err != nil {
		return nil, "", err
	}
	version := ""
	if len(versionNumber) > 0 {
		version = strings.TrimSpace(versionNumber[0])
	}
	logger.Info("get remote config", "project_id", projectID, "version", version)

	endpoint := "https://firebaseremoteconfig.googleapis.com/v1/" + resource
	if version != "" {
		endpoint += "?" + url.Values{"versionNumber": []string{version}}.Encode()
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		logger.Error("create remote config request failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("create remote config request: %w", err)
	}
	if _, err := s.setQuotaProject(req, targetProjectID); err != nil {
		return nil, "", fmt.Errorf("select remote config quota project: %w", err)
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
		return nil, "", newAPIError("firebase_remote_config", "get", resp, body)
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
	resource, targetProjectID, err := remoteConfigResource(projectID)
	if err != nil {
		return nil, "", err
	}

	body := bytes.TrimSpace(raw)
	if !json.Valid(body) {
		logger.Error("remote config update payload invalid json", "project_id", projectID)
		return nil, "", fmt.Errorf("remote config payload is not valid json")
	}

	endpoint := "https://firebaseremoteconfig.googleapis.com/v1/" + resource
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
	if _, err := s.setQuotaProject(req, targetProjectID); err != nil {
		return nil, "", fmt.Errorf("select remote config update quota project: %w", err)
	}
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
		return nil, "", newAPIError("firebase_remote_config", action, resp, respBody)
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
	page, err := s.ListRemoteConfigVersions(ctx, projectID, ListVersionsOptions{PageSize: 1})
	if err != nil {
		return RemoteConfigVersion{}, err
	}
	if len(page.Versions) == 0 {
		return RemoteConfigVersion{VersionNumber: notAvailableVersion}, nil
	}
	return page.Versions[0], nil
}

func (s *Service) ListRemoteConfigVersions(ctx context.Context, projectID string, opts ListVersionsOptions) (RemoteConfigVersionsPage, error) {
	logger := corelog.For("firebase")
	logger.Info("list remote config versions", "project_id", projectID)
	resource, targetProjectID, err := remoteConfigResource(projectID)
	if err != nil {
		return RemoteConfigVersionsPage{}, err
	}
	values := url.Values{}
	if opts.PageSize > 0 {
		values.Set("pageSize", fmt.Sprint(opts.PageSize))
	}
	if strings.TrimSpace(opts.PageToken) != "" {
		values.Set("pageToken", opts.PageToken)
	}
	if strings.TrimSpace(opts.EndVersionNumber) != "" {
		values.Set("endVersionNumber", opts.EndVersionNumber)
	}
	if strings.TrimSpace(opts.StartTime) != "" {
		values.Set("startTime", opts.StartTime)
	}
	if strings.TrimSpace(opts.EndTime) != "" {
		values.Set("endTime", opts.EndTime)
	}
	endpoint := "https://firebaseremoteconfig.googleapis.com/v1/" + resource + ":listVersions"
	if encoded := values.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		logger.Error("create remote config version request failed", "project_id", projectID, "err", err)
		return RemoteConfigVersionsPage{}, fmt.Errorf("create remote config version request: %w", err)
	}
	if _, err := s.setQuotaProject(req, targetProjectID); err != nil {
		return RemoteConfigVersionsPage{}, fmt.Errorf("select version-list quota project: %w", err)
	}
	logHTTPRequest(logger.With("project_id", projectID), req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("remote config version request failed", "project_id", projectID, "err", err)
		return RemoteConfigVersionsPage{}, fmt.Errorf("list remote config versions: %w", err)
	}
	logHTTPResponse(logger.With("project_id", projectID), req, resp)

	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		logger.Error("read remote config versions response failed", "project_id", projectID, "err", err)
		return RemoteConfigVersionsPage{}, fmt.Errorf("read remote config versions response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("remote config versions api returned non-200", "project_id", projectID, "status", resp.Status)
		return RemoteConfigVersionsPage{}, newAPIError("firebase_remote_config", "list_versions", resp, body)
	}

	var payload listVersionsResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&payload); err != nil {
		logger.Error("decode remote config versions failed", "project_id", projectID, "err", err)
		return RemoteConfigVersionsPage{}, fmt.Errorf("decode remote config versions response: %w", err)
	}
	for i := range payload.Versions {
		if strings.TrimSpace(payload.Versions[i].VersionNumber) == "" {
			payload.Versions[i].VersionNumber = notAvailableVersion
		}
	}
	return RemoteConfigVersionsPage(payload), nil
}

func (s *Service) RollbackRemoteConfig(ctx context.Context, projectID, versionNumber string) (json.RawMessage, string, error) {
	body, err := json.Marshal(map[string]string{"versionNumber": strings.TrimSpace(versionNumber)})
	if err != nil {
		return nil, "", err
	}
	resource, targetProjectID, err := remoteConfigResource(projectID)
	if err != nil {
		return nil, "", err
	}
	endpoint := "https://firebaseremoteconfig.googleapis.com/v1/" + resource + ":rollback"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("create rollback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	if _, err := s.setQuotaProject(req, targetProjectID); err != nil {
		return nil, "", fmt.Errorf("select rollback quota project: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("rollback remote config: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read rollback response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", newAPIError("firebase_remote_config", "rollback", resp, respBody)
	}
	respBody = normalizeRemoteConfigRaw(respBody)
	if !json.Valid(respBody) {
		return nil, "", fmt.Errorf("rollback remote config api returned invalid json")
	}
	return bytes.TrimSpace(respBody), strings.TrimSpace(resp.Header.Get("ETag")), nil
}

func remoteConfigResource(value string) (resource, projectID string, err error) {
	target, err := rctarget.Parse(value)
	if err != nil {
		return "", "", err
	}
	resource = "projects/" + target.ProjectID
	if namespace := target.Namespace(); namespace != "" {
		resource += "/namespaces/" + namespace
	}
	return resource + "/remoteConfig", target.ProjectID, nil
}

func normalizeRemoteConfigRaw(raw json.RawMessage) json.RawMessage {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("{}")) {
		return json.RawMessage(`{"version":{"versionNumber":"NA"}}`)
	}
	return json.RawMessage(NormalizeJSONEscapes(trimmed))
}
