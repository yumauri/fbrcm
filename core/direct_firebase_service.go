package core

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

type directFirebaseServiceContextKey struct{}
type directFirebaseDiscoveryServiceContextKey struct{}

type directFirebaseServiceBinding struct {
	projectID string
	service   *firebase.Service
}

// WithDirectFirebaseService binds a Firebase service to one physical project
// for project operations made with the returned context. The binding bypasses
// the persisted project and authentication registries.
func WithDirectFirebaseService(ctx context.Context, projectID string, service *firebase.Service) (context.Context, error) {
	target, err := rctarget.Parse(projectID)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, fmt.Errorf("firebase service must not be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	binding := directFirebaseServiceBinding{projectID: target.ProjectID, service: service}
	return context.WithValue(ctx, directFirebaseServiceContextKey{}, binding), nil
}

func directFirebaseServiceFromContext(ctx context.Context, projectID string) (*firebase.Service, bool, error) {
	if ctx == nil {
		return nil, false, nil
	}
	binding, ok := ctx.Value(directFirebaseServiceContextKey{}).(directFirebaseServiceBinding)
	if !ok {
		return nil, false, nil
	}
	if binding.projectID != projectID {
		return nil, true, fmt.Errorf("direct firebase service is bound to project %q, not %q", binding.projectID, projectID)
	}
	return binding.service, true, nil
}

// WithDirectFirebaseDiscoveryService binds an in-memory Firebase service for
// project discovery without consulting persisted authentication or project
// registries.
func WithDirectFirebaseDiscoveryService(ctx context.Context, service *firebase.Service) (context.Context, error) {
	if service == nil {
		return nil, fmt.Errorf("firebase service must not be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, directFirebaseDiscoveryServiceContextKey{}, service), nil
}

func directFirebaseDiscoveryServiceFromContext(ctx context.Context) (*firebase.Service, bool) {
	if ctx == nil {
		return nil, false
	}
	service, ok := ctx.Value(directFirebaseDiscoveryServiceContextKey{}).(*firebase.Service)
	return service, ok
}
