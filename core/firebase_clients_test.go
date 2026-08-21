package core

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestWithDirectFirebaseService(t *testing.T) {
	service := firebase.NewServiceWithHTTPClient(nil)
	ctx := WithExecutionPolicy(context.Background(), StatelessExecutionPolicy())
	ctx, err := WithDirectFirebaseService(ctx, "server@demo", service)
	if err != nil {
		t.Fatalf("WithDirectFirebaseService = %v", err)
	}

	got, direct, err := directFirebaseServiceFromContext(ctx, "demo")
	if err != nil || !direct || got != service {
		t.Fatalf("directFirebaseServiceFromContext = %p, %t, %v; want %p, true, nil", got, direct, err, service)
	}
}

func TestWithDirectFirebaseDiscoveryService(t *testing.T) {
	service := firebase.NewServiceWithHTTPClient(nil)
	ctx := WithExecutionPolicy(context.Background(), StatelessExecutionPolicy())
	ctx, err := WithDirectFirebaseDiscoveryService(ctx, service)
	if err != nil {
		t.Fatalf("WithDirectFirebaseDiscoveryService = %v", err)
	}

	got, direct := directFirebaseDiscoveryServiceFromContext(ctx)
	if !direct || got != service {
		t.Fatalf("directFirebaseDiscoveryServiceFromContext = %p, %t; want %p, true", got, direct, service)
	}
	if _, err := WithDirectFirebaseDiscoveryService(context.Background(), nil); err == nil {
		t.Fatal("WithDirectFirebaseDiscoveryService accepted a nil service")
	}
}

func TestFirebaseServiceForProjectRejectsConfiguredResolutionWhenLocalReadsDisabled(t *testing.T) {
	svc := setupCoreTestEnv(t)
	ctx := WithExecutionPolicy(context.Background(), StatelessExecutionPolicy())

	_, err := svc.firebaseServiceForProject(ctx, "demo")
	var policyErr *ExecutionPolicyError
	if !errors.As(err, &policyErr) {
		t.Fatalf("firebaseServiceForProject error = %v, want ExecutionPolicyError", err)
	}
	if policyErr.Capability != "local-state reads" || policyErr.Operation != "configured Firebase service resolution" {
		t.Fatalf("policy error = %#v", policyErr)
	}
}

func TestWithDirectFirebaseServiceRejectsInvalidArguments(t *testing.T) {
	if _, err := WithDirectFirebaseService(context.Background(), "", firebase.NewServiceWithHTTPClient(nil)); err == nil {
		t.Fatal("WithDirectFirebaseService accepted an empty project")
	}
	if _, err := WithDirectFirebaseService(context.Background(), "demo", nil); err == nil {
		t.Fatal("WithDirectFirebaseService accepted a nil service")
	}
}

func TestDirectFirebaseServiceRejectsDifferentProject(t *testing.T) {
	svc := setupCoreTestEnv(t)
	direct := firebase.NewServiceWithHTTPClient(nil)
	ctx, err := WithDirectFirebaseService(context.Background(), "demo", direct)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.firebaseServiceForProject(ctx, "other")
	if err == nil || !strings.Contains(err.Error(), `direct firebase service is bound to project "demo", not "other"`) {
		t.Fatalf("firebaseServiceForProject = %v, want direct-service project mismatch", err)
	}
}

func TestWithAuthFailureIDClassifiesQuotaProjectSources(t *testing.T) {
	for _, test := range []struct {
		name     string
		source   firebase.QuotaProjectSource
		wantKind string
	}{
		{name: "environment", source: firebase.QuotaProjectSourceEnvironment, wantKind: "configuration"},
		{name: "credentials", source: firebase.QuotaProjectSourceCredentials, wantKind: firebase.AuthenticationCredentialsInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			cause := &firebase.QuotaProjectError{Source: test.source, Err: errors.New("invalid quota project")}
			err := withAuthFailureID("personal", cause)
			var authErr *AuthError
			if !errors.As(err, &authErr) || authErr.Kind != test.wantKind || authErr.AuthID != "personal" || !errors.Is(err, cause) {
				t.Fatalf("error = %#v, want auth kind %q", err, test.wantKind)
			}
		})
	}
}
