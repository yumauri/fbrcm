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
)

const managedFeatureNamespace = "firebase"

// ManagedFeatureResourceError identifies an invalid experiment or rollout ID
// before any Firebase request is created.
type ManagedFeatureResourceError struct {
	Collection string
	ItemID     string
	Err        error
}

func (e *ManagedFeatureResourceError) Error() string { return e.Err.Error() }
func (e *ManagedFeatureResourceError) Unwrap() error { return e.Err }

type Experiment struct {
	Name           string               `json:"name"`
	Definition     ExperimentDefinition `json:"definition"`
	State          string               `json:"state,omitempty"`
	StartTime      string               `json:"startTime,omitempty" contract:"format=date-time"`
	EndTime        string               `json:"endTime,omitempty" contract:"format=date-time"`
	LastUpdateTime string               `json:"lastUpdateTime,omitempty" contract:"format=date-time"`
	ETag           string               `json:"etag,omitempty"`
	raw            json.RawMessage
}

type ExperimentDefinition struct {
	DisplayName string               `json:"displayName,omitempty"`
	Description string               `json:"description,omitempty"`
	Service     string               `json:"service,omitempty"`
	Objectives  ExperimentObjectives `json:"objectives,omitzero"`
	Variants    []ExperimentVariant  `json:"variants,omitempty"`
}

type ExperimentObjectives struct {
	ActivationEvent ExperimentActivationEvent  `json:"activationEvent,omitzero"`
	EventObjectives []ExperimentEventObjective `json:"eventObjectives,omitempty"`
}

type ExperimentActivationEvent struct {
	Event string `json:"event,omitempty"`
}

type ExperimentEventObjective struct {
	IsPrimary               bool                       `json:"isPrimary,omitempty"`
	ABTOptimizationFunction string                     `json:"abtOptimizationFunction,omitempty"`
	CustomObjectiveDetails  *ExperimentCustomObjective `json:"customObjectiveDetails,omitempty"`
	SystemObjectiveDetails  *ExperimentSystemObjective `json:"systemObjectiveDetails,omitempty"`
}

type ExperimentCustomObjective struct {
	Event     string `json:"event,omitempty"`
	CountType string `json:"countType,omitempty"`
}

type ExperimentSystemObjective struct {
	Objective string `json:"objective,omitempty"`
}

type ExperimentVariant struct {
	Name   string `json:"name,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

type ExperimentsPage struct {
	Experiments   []Experiment `json:"experiments"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

type Rollout struct {
	Name           string            `json:"name"`
	Definition     RolloutDefinition `json:"definition"`
	State          string            `json:"state,omitempty"`
	CreateTime     string            `json:"createTime,omitempty" contract:"format=date-time"`
	StartTime      string            `json:"startTime,omitempty" contract:"format=date-time"`
	EndTime        string            `json:"endTime,omitempty" contract:"format=date-time"`
	LastUpdateTime string            `json:"lastUpdateTime,omitempty" contract:"format=date-time"`
	ETag           string            `json:"etag,omitempty"`
	raw            json.RawMessage
}

type RolloutDefinition struct {
	DisplayName    string            `json:"displayName,omitempty"`
	Description    string            `json:"description,omitempty"`
	ControlVariant ExperimentVariant `json:"controlVariant,omitzero"`
	EnabledVariant ExperimentVariant `json:"enabledVariant,omitzero"`
}

type RolloutsPage struct {
	Rollouts      []Rollout `json:"rollouts"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

type ListManagedFeaturesOptions struct {
	PageSize  int
	PageToken string
}

type experimentWire Experiment

func (e *Experiment) UnmarshalJSON(data []byte) error {
	var decoded experimentWire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*e = Experiment(decoded)
	e.raw = append(e.raw[:0], data...)
	return nil
}

func (e Experiment) MarshalJSON() ([]byte, error) {
	if len(e.raw) > 0 {
		return append([]byte(nil), e.raw...), nil
	}
	return json.Marshal(experimentWire(e))
}

type rolloutWire Rollout

func (r *Rollout) UnmarshalJSON(data []byte) error {
	var decoded rolloutWire
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*r = Rollout(decoded)
	r.raw = append(r.raw[:0], data...)
	return nil
}

func (r Rollout) MarshalJSON() ([]byte, error) {
	if len(r.raw) > 0 {
		return append([]byte(nil), r.raw...), nil
	}
	return json.Marshal(rolloutWire(r))
}

func (s *Service) ListExperiments(ctx context.Context, projectIdentifier, quotaProjectID string, opts ListManagedFeaturesOptions) (ExperimentsPage, error) {
	var page ExperimentsPage
	if err := s.getManagedFeatureJSON(ctx, projectIdentifier, quotaProjectID, "experiments", "", opts, &page); err != nil {
		return ExperimentsPage{}, err
	}
	return page, nil
}

func (s *Service) GetExperiment(ctx context.Context, projectIdentifier, quotaProjectID, experimentID string) (Experiment, error) {
	var experiment Experiment
	if err := s.getManagedFeatureJSON(ctx, projectIdentifier, quotaProjectID, "experiments", experimentID, ListManagedFeaturesOptions{}, &experiment); err != nil {
		return Experiment{}, err
	}
	return experiment, nil
}

func (s *Service) DeleteExperiment(ctx context.Context, projectIdentifier, quotaProjectID, experimentID string) error {
	return s.deleteManagedFeature(ctx, projectIdentifier, quotaProjectID, "experiments", experimentID)
}

func (s *Service) ListRollouts(ctx context.Context, projectIdentifier, quotaProjectID string, opts ListManagedFeaturesOptions) (RolloutsPage, error) {
	var page RolloutsPage
	if err := s.getManagedFeatureJSON(ctx, projectIdentifier, quotaProjectID, "rollouts", "", opts, &page); err != nil {
		return RolloutsPage{}, err
	}
	return page, nil
}

func (s *Service) GetRollout(ctx context.Context, projectIdentifier, quotaProjectID, rolloutID string) (Rollout, error) {
	var rollout Rollout
	if err := s.getManagedFeatureJSON(ctx, projectIdentifier, quotaProjectID, "rollouts", rolloutID, ListManagedFeaturesOptions{}, &rollout); err != nil {
		return Rollout{}, err
	}
	return rollout, nil
}

func (s *Service) DeleteRollout(ctx context.Context, projectIdentifier, quotaProjectID, rolloutID string) error {
	return s.deleteManagedFeature(ctx, projectIdentifier, quotaProjectID, "rollouts", rolloutID)
}

func (s *Service) getManagedFeatureJSON(ctx context.Context, projectIdentifier, quotaProjectID, collection, itemID string, opts ListManagedFeaturesOptions, destination any) error {
	logger := corelog.For("firebase")
	resource, err := managedFeatureResource(projectIdentifier, collection, itemID)
	if err != nil {
		return err
	}
	values := url.Values{}
	if itemID == "" {
		if opts.PageSize > 0 {
			values.Set("pageSize", fmt.Sprint(opts.PageSize))
		}
		if strings.TrimSpace(opts.PageToken) != "" {
			values.Set("pageToken", opts.PageToken)
		}
	}
	endpoint := "https://firebaseremoteconfig.googleapis.com/v1/" + resource
	if encoded := values.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create Remote Config %s request: %w", collection, err)
	}
	s.setQuotaProject(req, quotaProjectID)
	logHTTPRequest(logger.With("project_id", quotaProjectID, "resource", resource), req)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("get Remote Config %s: %w", collection, err)
	}
	logHTTPResponse(logger.With("project_id", quotaProjectID, "resource", resource), req, resp)
	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return fmt.Errorf("read Remote Config %s response: %w", collection, err)
	}
	if resp.StatusCode != http.StatusOK {
		return newAPIError("firebase_remote_config", "get_"+collection, resp, body)
	}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(destination); err != nil {
		return fmt.Errorf("decode Remote Config %s response: %w", collection, err)
	}
	return nil
}

func (s *Service) deleteManagedFeature(ctx context.Context, projectIdentifier, quotaProjectID, collection, itemID string) error {
	logger := corelog.For("firebase")
	resource, err := managedFeatureResource(projectIdentifier, collection, itemID)
	if err != nil {
		return err
	}
	endpoint := "https://firebaseremoteconfig.googleapis.com/v1/" + resource
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create Remote Config %s delete request: %w", collection, err)
	}
	s.setQuotaProject(req, quotaProjectID)
	logHTTPRequest(logger.With("project_id", quotaProjectID, "resource", resource), req)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete Remote Config %s: %w", collection, err)
	}
	logHTTPResponse(logger.With("project_id", quotaProjectID, "resource", resource), req, resp)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read Remote Config %s delete response: %w", collection, err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return newAPIError("firebase_remote_config", "delete_"+collection, resp, body)
	}
	return nil
}

func managedFeatureResource(projectIdentifier, collection, itemID string) (string, error) {
	projectIdentifier = strings.TrimSpace(projectIdentifier)
	if projectIdentifier == "" || strings.Contains(projectIdentifier, "/") {
		return "", fmt.Errorf("invalid Firebase project identifier %q", projectIdentifier)
	}
	if collection != "experiments" && collection != "rollouts" {
		return "", fmt.Errorf("unsupported Remote Config managed feature %q", collection)
	}
	parent := "projects/" + url.PathEscape(projectIdentifier) + "/namespaces/" + managedFeatureNamespace + "/" + collection
	if itemID == "" {
		return parent, nil
	}
	if strings.Contains(itemID, "/") {
		prefix := parent + "/"
		if !strings.HasPrefix(itemID, prefix) || strings.TrimPrefix(itemID, prefix) == "" || strings.Contains(strings.TrimPrefix(itemID, prefix), "/") {
			return "", &ManagedFeatureResourceError{
				Collection: collection,
				ItemID:     itemID,
				Err:        fmt.Errorf("invalid %s resource name %q", collection, itemID),
			}
		}
		return itemID, nil
	}
	return parent + "/" + url.PathEscape(itemID), nil
}

func ManagedFeatureID(resourceName string) string {
	resourceName = strings.TrimSpace(resourceName)
	if index := strings.LastIndex(resourceName, "/"); index >= 0 {
		return resourceName[index+1:]
	}
	return resourceName
}
