package core

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func (s *Core) firebaseServiceForProject(ctx context.Context, projectID string) (*firebase.Service, error) {
	project, err := s.ProjectByID(projectID)
	if err != nil {
		return nil, err
	}
	return s.firebaseServiceForAuth(ctx, project.AuthID)
}

func (s *Core) firebaseServiceForAuth(ctx context.Context, authID string) (*firebase.Service, error) {
	s.firebaseMu.Lock()
	if fb, ok := s.firebase[authID]; ok {
		s.firebaseMu.Unlock()
		return fb, nil
	}
	s.firebaseMu.Unlock()

	serviceCtx := s.ctx
	if ctx != nil {
		serviceCtx = ctx
	}

	result, err, _ := s.firebaseInit.Do(authID, func() (any, error) {
		s.firebaseMu.Lock()
		if fb, ok := s.firebase[authID]; ok {
			s.firebaseMu.Unlock()
			return fb, nil
		}
		s.firebaseMu.Unlock()

		auth, err := s.authEntry(authID)
		if err != nil {
			return nil, err
		}
		fb, err := firebase.NewServiceForAuth(serviceCtx, auth, true)
		if err != nil {
			return nil, err
		}
		s.firebaseMu.Lock()
		s.firebase[authID] = fb
		s.firebaseMu.Unlock()
		return fb, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*firebase.Service), nil
}

func (s *Core) authEntry(authID string) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := config.LoadAuth()
	if err != nil {
		return config.AuthEntry{}, err
	}
	auth, ok := authFile.FindAuth(authID)
	if !ok {
		return config.AuthEntry{}, fmt.Errorf("auth %q is not configured", authID)
	}
	return auth, nil
}

func (s *Core) dropFirebaseService(authID string) {
	s.firebaseMu.Lock()
	delete(s.firebase, authID)
	s.firebaseMu.Unlock()
}

// InjectFirebaseService replaces the cached firebase client for authID.
// It is intended for tests that stub Firebase HTTP responses.
func (s *Core) InjectFirebaseService(authID string, fb *firebase.Service) {
	s.firebaseMu.Lock()
	s.firebase[authID] = fb
	s.firebaseMu.Unlock()
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
