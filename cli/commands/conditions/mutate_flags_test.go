package conditions

import "testing"

func TestConditionDescriptionMutationFlagsAreNotExposed(t *testing.T) {
	cmd := New(nil)
	add, _, err := cmd.Find([]string{"add"})
	if err != nil {
		t.Fatal(err)
	}
	if add.Flags().Lookup("description") != nil {
		t.Fatal("conditions add still exposes --description")
	}
	edit, _, err := cmd.Find([]string{"edit"})
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"description", "no-description"} {
		if edit.Flags().Lookup(flag) != nil {
			t.Errorf("conditions edit still exposes --%s", flag)
		}
	}
}
