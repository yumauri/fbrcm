package core

import (
	"context"
	"fmt"
	"maps"

	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

type directFirebaseServiceContextKey struct{}
type directFirebaseDiscoveryServiceContextKey struct{}

type directFirebaseServiceBindings map[string]*firebase.Service

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
	bindings := make(directFirebaseServiceBindings)
	if existing, ok := ctx.Value(directFirebaseServiceContextKey{}).(directFirebaseServiceBindings); ok {
		maps.Copy(bindings, existing)
	}
	bindings[target.ProjectID] = service
	return context.WithValue(ctx, directFirebaseServiceContextKey{}, bindings), nil
}

func directFirebaseServiceFromContext(ctx context.Context, projectID string) (*firebase.Service, bool, error) {
	if ctx == nil {
		return nil, false, nil
	}
	bindings, ok := ctx.Value(directFirebaseServiceContextKey{}).(directFirebaseServiceBindings)
	if !ok {
		return nil, false, nil
	}
	service, ok := bindings[projectID]
	if !ok {
		if len(bindings) == 1 {
			for boundProjectID := range bindings {
				return nil, true, fmt.Errorf("direct firebase service is bound to project %q, not %q", boundProjectID, projectID)
			}
		}
		return nil, true, fmt.Errorf("direct firebase service is not bound to project %q", projectID)
	}
	return service, true, nil
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
