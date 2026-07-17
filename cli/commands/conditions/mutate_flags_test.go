package conditions

import "testing"

func TestConditionDescriptionMutationFlags(t *testing.T) {
	cmd := New(nil)
	add, _, err := cmd.Find([]string{"add"})
	if err != nil || add.Flags().Lookup("description") == nil {
		t.Fatalf("conditions add --description missing: %v", err)
	}
	edit, _, err := cmd.Find([]string{"edit"})
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"description", "no-description"} {
		if edit.Flags().Lookup(flag) == nil {
			t.Errorf("conditions edit missing --%s", flag)
		}
	}
}
