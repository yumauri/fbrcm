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
	for _, commandID := range []string{"", "versions.restore"} {
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
	base := statelessUpdatingRemoteRead()
	original := statelessUpdatingRemoteRead()
	first := withStatelessCommandEffects(base)
	second := withStatelessCommandEffects(base)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated stateless decoration accumulated predicates:\nfirst: %#v\nsecond: %#v", first, second)
	}
	if !reflect.DeepEqual(base, original) {
		t.Fatalf("stateless decoration mutated its input: %#v", base)
	}
	for _, effect := range first.effects {
		for _, clause := range effect.when {
			var stateful, stateless int
			for _, item := range clause.AllOf {
				if item.Source != "option" || item.Name != "stateless" || item.Operator != "equals" {
					continue
				}
				switch item.Value {
				case true:
					stateless++
				case false:
					stateful++
				}
			}
			if stateful > 1 || stateless > 1 || stateful+stateless > 1 {
				t.Fatalf("%s has contradictory or repeated stateless predicates: %#v", effect.name, clause)
			}
		}
	}
	for _, effect := range first.effects {
		if effect.name == "local_cache_write" && len(effect.when) != 2 {
			t.Fatalf("local cache write conditions = %#v, want only the two stateful paths", effect.when)
		}
	}
}

func TestStatelessCapabilityRestrictsUnconditionalPersistenceEffect(t *testing.T) {
	decorated := withStatelessCommandEffects(behavior(1, "none", effect("local_state_write")))
	if len(decorated.effects) != 1 || len(decorated.effects[0].when) != 1 ||
		!clauseHasPredicateValue(decorated.effects[0].when[0], "option", "stateless", "equals", false) {
		t.Fatalf("unconditional persistence effect was not restricted: %#v", decorated.effects)
	}
}
