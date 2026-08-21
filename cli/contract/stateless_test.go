package contract

import (
	"reflect"
	"testing"
)

func TestSupportsStatelessCommand(t *testing.T) {
	for commandID := range statelessCommands {
		if !SupportsStatelessCommand(commandID) {
			t.Errorf("SupportsStatelessCommand(%q) = false", commandID)
		}
	}
	for _, commandID := range []string{"", "conditions.add", "versions.rollback"} {
		if SupportsStatelessCommand(commandID) {
			t.Errorf("SupportsStatelessCommand(%q) = true", commandID)
		}
	}
}

func TestStatelessCommandRequiresAccessToken(t *testing.T) {
	for commandID, support := range statelessCommands {
		if support.requiresAccessToken && !StatelessCommandRequiresAccessToken(commandID) {
			t.Errorf("StatelessCommandRequiresAccessToken(%q) = false", commandID)
		}
	}
	for _, commandID := range []string{"", "project.open"} {
		if StatelessCommandRequiresAccessToken(commandID) {
			t.Errorf("StatelessCommandRequiresAccessToken(%q) = true", commandID)
		}
	}
}

func TestStatelessCapabilityEffectsDoNotMutateSharedBehavior(t *testing.T) {
	base := historicalVersionRead()
	first := withStatelessCommandEffects(base)
	second := withStatelessCommandEffects(base)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated stateless decoration accumulated predicates:\nfirst: %#v\nsecond: %#v", first, second)
	}
	if capabilityBehaviorHasPredicate(base, "option", "stateless", "equals") {
		t.Fatalf("stateless decoration mutated its input: %#v", base)
	}
}

func capabilityBehaviorHasPredicate(b capabilityBehavior, source, name, operator string) bool {
	for _, effect := range b.effects {
		if containsPredicate(effect.when, source, name, operator) {
			return true
		}
	}
	return containsPredicate(b.networkWhen, source, name, operator)
}
