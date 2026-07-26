package config

import (
	"testing"

	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func TestNormalizeTemplatePreferencesDefaultsAndValidatesPrimary(t *testing.T) {
	project := Project{ProjectID: "demo"}
	if err := project.NormalizeTemplatePreferences(); err != nil {
		t.Fatal(err)
	}
	if len(project.Templates) != 1 || project.Templates[0] != rctarget.Client || project.PrimaryTemplate != rctarget.Client {
		t.Fatalf("default preferences = %#v", project)
	}

	project = Project{
		ProjectID:       "demo",
		Templates:       []rctarget.Kind{rctarget.Client},
		PrimaryTemplate: rctarget.Server,
	}
	if err := project.NormalizeTemplatePreferences(); err == nil {
		t.Fatal("hidden primary template was accepted")
	}
}

func TestTemplateKindsPlacesPrimaryFirst(t *testing.T) {
	project := Project{
		ProjectID:       "demo",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Server,
	}
	got := project.TemplateKinds()
	if len(got) != 2 || got[0] != rctarget.Server || got[1] != rctarget.Client {
		t.Fatalf("TemplateKinds = %v, want server then client", got)
	}
}
