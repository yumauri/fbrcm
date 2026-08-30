package core

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func (s *Core) firebaseServiceForProject(ctx context.Context, projectID string) (*firebase.Service, error) {
	target, err := rctarget.Parse(projectID)
	if err != nil {
		return nil, err
	}
	if fb, direct, err := directFirebaseServiceFromContext(ctx, target.ProjectID); direct {
		return fb, err
	}
	if err := requireLocalStateRead(ctx, "configured Firebase service resolution"); err != nil {
		return nil, err
	}
	project, err := s.ProjectByID(target.ProjectID)
	if err != nil {
		return nil, err
	}
	if project.Disabled {
		return nil, fmt.Errorf("project %q is disabled; run projects update to rediscover it", projectID)
	}
	fb, err := s.firebaseServiceForAuth(ctx, project.AuthID)
	if err != nil {
		return nil, err
	}
	return fb.WithQuotaProjectOverride(project.QuotaProjectID), nil
}

func (s *Core) firebaseServiceForAuth(ctx context.Context, authID string) (*firebase.Service, error) {
	clientKey := firebaseClientKey(authID)
	s.firebaseMu.Lock()
	if fb, ok := s.firebase[clientKey]; ok {
		s.firebaseMu.Unlock()
		return fb, nil
	}
	s.firebaseMu.Unlock()

	serviceCtx := s.ctx
	if ctx != nil {
		serviceCtx = ctx
	}
	serviceCtx = s.WithFirebaseRequestController(serviceCtx)

	result, err, _ := s.firebaseInit.Do(clientKey, func() (any, error) {
		s.firebaseMu.Lock()
		if fb, ok := s.firebase[clientKey]; ok {
			s.firebaseMu.Unlock()
			return fb, nil
		}
		s.firebaseMu.Unlock()

		auth, err := s.authEntry(authID)
		if err != nil {
			return nil, err
		}
		authCtx, autoOpen := s.oauthAuthorizationContext(serviceCtx, authID, true)
		fb, err := firebase.NewServiceForAuth(authCtx, auth, autoOpen, s.googleOAuthCredentials)
		if err != nil {
			if errors.Is(err, firebase.ErrOAuthInteractionRequired) {
				return nil, &OAuthInteractionError{AuthID: authID, Err: err}
			}
			return nil, withAuthFailureID(authID, err)
		}
		s.firebaseMu.Lock()
		s.firebase[clientKey] = fb
		s.firebaseMu.Unlock()
		return fb, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*firebase.Service), nil
}

func withAuthFailureID(authID string, err error) error {
	if _, ok := errors.AsType[*firebase.QuotaProjectRequiredError](err); ok {
		return &AuthError{Kind: "quota_project_required", AuthID: authID, Err: err}
	}
	if authentication, ok := errors.AsType[*firebase.AuthenticationError](err); ok {
		return &AuthError{Kind: authentication.Kind, AuthID: authID, Err: err}
	}
	if quotaProject, ok := errors.AsType[*firebase.QuotaProjectError](err); ok {
		if quotaProject.Source == firebase.QuotaProjectSourceCredentials {
			authentication := &firebase.AuthenticationError{
				Kind:      firebase.AuthenticationCredentialsInvalid,
				AuthType:  "gcloud",
				Operation: "load_credentials",
				Err:       err,
			}
			return &AuthError{Kind: authentication.Kind, AuthID: authID, Err: authentication}
		}
		return &AuthError{Kind: "configuration", AuthID: authID, Err: err}
	}
	return err
}

func (s *Core) authEntry(authID string) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, &AuthError{Kind: "invalid_argument", AuthID: authID, Err: err}
	}
	authFile, err := loadAuthWithSetupHint()
	if err != nil {
		return config.AuthEntry{}, err
	}
	auth, ok := authFile.FindAuth(authID)
	if !ok {
		return config.AuthEntry{}, authNotConfiguredError(authFile, authID)
	}
	return auth, nil
}

func (s *Core) dropFirebaseService(authID string) {
	s.firebaseMu.Lock()
	delete(s.firebase, firebaseClientKey(authID))
	s.firebaseMu.Unlock()
}

// InjectFirebaseService replaces the cached firebase client for authID.
// It is intended for tests that stub Firebase HTTP responses.
func (s *Core) InjectFirebaseService(authID string, fb *firebase.Service) {
	s.firebaseMu.Lock()
	s.firebase[firebaseClientKey(authID)] = fb
	s.firebaseMu.Unlock()
}

func firebaseClientKey(authID string) string {
	return config.GetActiveProfileName() + "\x00" + authID
}

func removeAuthFiles(auth config.AuthEntry) error {
	switch auth.Type {
	case config.AuthTypeOAuth:
		if err := removeFileIfPresent(config.OAuthClientSecretPath(auth)); err != nil {
			return fmt.Errorf("remove client secret: %w", err)
		}
		if err := removeFileIfPresent(config.OAuthTokenPath(auth)); err != nil {
			return fmt.Errorf("remove token: %w", err)
		}
	case config.AuthTypeGoogle:
		if err := removeFileIfPresent(config.OAuthTokenPath(auth)); err != nil {
			return fmt.Errorf("remove token: %w", err)
		}
	case config.AuthTypeServiceAccount:
		if err := removeFileIfPresent(config.ServiceAccountKeyPath(auth)); err != nil {
			return fmt.Errorf("remove service account key: %w", err)
		}
	}
	return nil
}

func removeFileIfPresent(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
