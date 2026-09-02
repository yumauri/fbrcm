package machine

import (
	"context"
	"testing"
)

func TestProfilelessContextIsExplicitAndPreservesMachineState(t *testing.T) {
	ctx := WithState(context.Background())
	state := FromContext(ctx)
	if Profileless(ctx) {
		t.Fatal("ordinary machine context is profileless")
	}

	profileless := WithProfileless(ctx)
	if !Profileless(profileless) {
		t.Fatal("profileless marker is missing")
	}
	if FromContext(profileless) != state {
		t.Fatal("profileless context replaced machine state")
	}
	if Profileless(ctx) {
		t.Fatal("profileless marker mutated its parent context")
	}
}
