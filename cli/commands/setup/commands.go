package setup

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
)

const (
	choiceRetry      = "retry"
	choiceAddAnother = "add-another"
	choiceFinish     = "finish"
)

type setupService interface {
	GoogleAuthAvailable() bool
	InspectStartupState() (core.StartupState, error)
	AddGoogleAuthWithQuotaProject(authID, label, quotaProjectID string) (config.AuthEntry, error)
	AddOAuthAuthWithQuotaProject(authID, label string, secret []byte, quotaProjectID string) (config.AuthEntry, error)
	AddServiceAccountAuthWithQuotaProject(authID, label string, key []byte, quotaProjectID string) (config.AuthEntry, error)
	AddGCloudAuthWithQuotaProject(authID, label, quotaProjectID string) (config.AuthEntry, error)
	SetAuthQuotaProject(authID, quotaProjectID string) (config.AuthEntry, string, bool, error)
	EnsureAuthLogin(ctx context.Context, authID string, noOpen bool) error
	SyncProjectsForAuth(ctx context.Context, authID string) ([]core.Project, string, error)
}

// New constructs the guided CLI setup command.
func New(svc *core.Core) *cobra.Command {
	return newCommand(svc, nil)
}

func newCommand(svc setupService, prompts prompter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactively configure authentication and discover projects",
		Args:  cobra.NoArgs,
	}
	cmd.Flags().Bool("noopen", false, "Do not open the OAuth authorization URL automatically")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if shared.MachineMode(cmd) {
			return shared.InteractionRequiredWithArguments(
				"guided setup requires an interactive terminal; run `fbrcm setup` without --json",
				"guided_setup",
				false,
				"",
			)
		}
		noOpen, err := cmd.Flags().GetBool("noopen")
		if err != nil {
			return err
		}
		if svc == nil {
			return fmt.Errorf("setup service is unavailable")
		}
		if prompts == nil {
			prompts = newTerminalPrompter(cmd)
		}
		progress.Stop()
		return (&runner{cmd: cmd, svc: svc, prompts: prompts, noOpen: noOpen}).run(shared.CommandContext(cmd))
	}
	contract.RegisterNoData(cmd)
	return cmd
}

type runner struct {
	cmd     *cobra.Command
	svc     setupService
	prompts prompter
	noOpen  bool
}

func (r *runner) run(ctx context.Context) error {
	state, err := r.svc.InspectStartupState()
	if err != nil {
		return fmt.Errorf("inspect setup state: %w", err)
	}
	_, _ = fmt.Fprintf(r.cmd.ErrOrStderr(), "Setting up profile: %s\n\n", profileName(state.Profile))

	if len(state.Auth) > 0 && len(state.Projects) > 0 {
		return r.printComplete(state)
	}

	var auth config.AuthEntry
	if len(state.Auth) == 0 {
		if len(state.Projects) > 0 {
			_, _ = fmt.Fprintf(r.cmd.ErrOrStderr(), "The profile has %s, but no authentication identity for live Firebase access.\n\n", rcdisplay.FormatCount(len(state.Projects), "cached project", "cached projects"))
		}
		auth, err = r.addIdentity(state.Auth)
	} else {
		auth, err = r.chooseIdentity(state.Auth, state.DefaultAuthID)
		if err == nil && auth.QuotaProjectID == "" {
			auth, err = r.configureQuotaProject(auth)
		}
	}
	if err != nil {
		return err
	}

	for {
		discovered, err := r.authenticateAndDiscover(ctx, auth)
		if err != nil {
			return err
		}
		if discovered > 0 {
			return r.printDiscoveredComplete(state.Profile, auth, discovered)
		}

		_, _ = fmt.Fprintln(r.cmd.ErrOrStderr(), "✓ Authentication is valid, but it did not return any accessible Firebase projects.")
		choice, err := r.prompts.selectOne("What would you like to do?", []promptChoice{
			{Label: "Try project discovery again", Value: choiceRetry},
			{Label: "Add another authentication identity", Value: choiceAddAnother},
			{Label: "Finish without projects", Value: choiceFinish},
		})
		if err != nil {
			return err
		}
		switch choice {
		case choiceRetry:
			continue
		case choiceAddAnother:
			state, err = r.svc.InspectStartupState()
			if err != nil {
				return fmt.Errorf("refresh setup state: %w", err)
			}
			auth, err = r.addIdentity(state.Auth)
			if err != nil {
				return err
			}
		case choiceFinish:
			return r.printNoProjectsComplete(state.Profile, auth)
		default:
			return fmt.Errorf("unsupported setup choice %q", choice)
		}
	}
}

func (r *runner) addIdentity(existing []config.AuthEntry) (config.AuthEntry, error) {
	choices := make([]promptChoice, 0, 4)
	if r.svc.GoogleAuthAvailable() {
		choices = append(choices, promptChoice{Label: "Continue with Google", Value: config.AuthTypeGoogle})
	} else {
		_, _ = fmt.Fprintln(r.cmd.ErrOrStderr(), "Continue with Google is unavailable in this build.")
	}
	choices = append(choices,
		promptChoice{Label: "OAuth desktop login", Value: config.AuthTypeOAuth},
		promptChoice{Label: "Service account", Value: config.AuthTypeServiceAccount},
		promptChoice{Label: "Existing gcloud credentials", Value: config.AuthTypeGCloud},
	)
	method, err := r.prompts.selectOne("Choose an authentication method:", choices)
	if err != nil {
		return config.AuthEntry{}, err
	}

	suggested := core.SuggestedAuthID(existing)
	authID, err := r.prompts.text("Authentication name:", suggested, func(value string) error {
		value = strings.TrimSpace(value)
		if err := config.ValidateAuthID(value); err != nil {
			return err
		}
		if slices.ContainsFunc(existing, func(entry config.AuthEntry) bool { return entry.ID == value }) {
			return fmt.Errorf("authentication identity %q already exists", value)
		}
		return nil
	})
	if err != nil {
		return config.AuthEntry{}, err
	}
	authID = strings.TrimSpace(authID)

	var credential []byte
	if method == config.AuthTypeOAuth || method == config.AuthTypeServiceAccount {
		label := "OAuth client JSON"
		if method == config.AuthTypeServiceAccount {
			label = "service-account JSON key"
		}
		path, err := r.prompts.pickJSON(label)
		if err != nil {
			return config.AuthEntry{}, err
		}
		credential, err = os.ReadFile(path)
		if err != nil {
			return config.AuthEntry{}, fmt.Errorf("read selected credential file: %w", err)
		}
		if method == config.AuthTypeOAuth {
			err = firebase.ValidateOAuthClientSecret(credential)
		} else {
			err = firebase.ValidateServiceAccountKey(credential)
		}
		if err != nil {
			return config.AuthEntry{}, &shared.ValidationError{Code: "auth.credentials_invalid", Source: "auth", Stage: "input", Target: authID, Err: err}
		}
	}

	quotaProjectID, err := r.promptQuotaProject("")
	if err != nil {
		return config.AuthEntry{}, err
	}

	var entry config.AuthEntry
	switch method {
	case config.AuthTypeGoogle:
		entry, err = r.svc.AddGoogleAuthWithQuotaProject(authID, "", quotaProjectID)
	case config.AuthTypeOAuth:
		entry, err = r.svc.AddOAuthAuthWithQuotaProject(authID, "", credential, quotaProjectID)
	case config.AuthTypeServiceAccount:
		entry, err = r.svc.AddServiceAccountAuthWithQuotaProject(authID, "", credential, quotaProjectID)
	case config.AuthTypeGCloud:
		entry, err = r.svc.AddGCloudAuthWithQuotaProject(authID, "", quotaProjectID)
	default:
		return config.AuthEntry{}, fmt.Errorf("unsupported authentication method %q", method)
	}
	if err != nil {
		return config.AuthEntry{}, fmt.Errorf("add authentication: %w", err)
	}
	_, _ = fmt.Fprintf(r.cmd.ErrOrStderr(), "✓ Authentication saved: %s\n", entry.ID)
	return entry, nil
}

func (r *runner) chooseIdentity(entries []config.AuthEntry, defaultID string) (config.AuthEntry, error) {
	if len(entries) == 1 {
		_, _ = fmt.Fprintf(r.cmd.ErrOrStderr(), "Using authentication: %s\n", entries[0].ID)
		return entries[0], nil
	}
	ordered := append([]config.AuthEntry(nil), entries...)
	slices.SortStableFunc(ordered, func(left, right config.AuthEntry) int {
		if left.ID == defaultID {
			return -1
		}
		if right.ID == defaultID {
			return 1
		}
		return strings.Compare(left.ID, right.ID)
	})
	choices := make([]promptChoice, 0, len(ordered))
	for _, entry := range ordered {
		label := fmt.Sprintf("%s (%s)", entry.ID, authTypeLabel(entry.Type))
		if entry.ID == defaultID {
			label += " — default"
		}
		choices = append(choices, promptChoice{Label: label, Value: entry.ID})
	}
	selected, err := r.prompts.selectOne("Choose an authentication identity:", choices)
	if err != nil {
		return config.AuthEntry{}, err
	}
	for _, entry := range entries {
		if entry.ID == selected {
			return entry, nil
		}
	}
	return config.AuthEntry{}, fmt.Errorf("selected authentication identity %q is unavailable", selected)
}

func (r *runner) configureQuotaProject(auth config.AuthEntry) (config.AuthEntry, error) {
	_, _ = fmt.Fprintf(r.cmd.ErrOrStderr(), "Authentication %s has no persisted quota project.\n", auth.ID)
	quotaProjectID, err := r.promptQuotaProject("")
	if err != nil {
		return config.AuthEntry{}, err
	}
	entry, _, _, err := r.svc.SetAuthQuotaProject(auth.ID, quotaProjectID)
	if err != nil {
		return config.AuthEntry{}, fmt.Errorf("set authentication quota project: %w", err)
	}
	return entry, nil
}

func (r *runner) promptQuotaProject(initial string) (string, error) {
	_, _ = fmt.Fprintln(r.cmd.ErrOrStderr(), "The quota project is used for Google API quota and billing attribution.")
	value, err := r.prompts.text("Google Cloud quota project:", initial, func(value string) error {
		return config.ValidateQuotaProjectID(strings.TrimSpace(value))
	})
	return strings.TrimSpace(value), err
}

func (r *runner) authenticateAndDiscover(ctx context.Context, auth config.AuthEntry) (int, error) {
	progress.Start("Authenticating…")
	err := r.svc.EnsureAuthLogin(ctx, auth.ID, r.noOpen)
	progress.Stop()
	if err != nil {
		return 0, fmt.Errorf("authenticate %s: %w", auth.ID, err)
	}
	_, _ = fmt.Fprintf(r.cmd.ErrOrStderr(), "✓ Authentication verified: %s\n", auth.ID)

	progress.Start("Discovering Firebase projects…")
	projects, _, err := r.svc.SyncProjectsForAuth(ctx, auth.ID)
	progress.Stop()
	if err != nil {
		return 0, fmt.Errorf("discover Firebase projects with %s: %w", auth.ID, err)
	}
	discovered := 0
	for _, project := range projects {
		if slices.Contains(project.DiscoveredBy, auth.ID) {
			discovered++
		}
	}
	return discovered, nil
}

func (r *runner) printComplete(state core.StartupState) error {
	_, err := fmt.Fprintf(r.cmd.OutOrStdout(), "✅ setup already complete\nprofile: %s\nauthentication: %s\nprojects: %s\n", profileName(state.Profile), rcdisplay.FormatCount(len(state.Auth), "identity", "identities"), rcdisplay.FormatCount(len(state.Projects), "project", "projects"))
	return err
}

func (r *runner) printDiscoveredComplete(profile string, auth config.AuthEntry, count int) error {
	_, err := fmt.Fprintf(r.cmd.OutOrStdout(), "✅ setup complete\nprofile: %s\nauthentication: %s (%s)\nprojects discovered: %s\nnext: fbrcm projects list\n", profileName(profile), auth.ID, authTypeLabel(auth.Type), rcdisplay.FormatCount(count, "project", "projects"))
	return err
}

func (r *runner) printNoProjectsComplete(profile string, auth config.AuthEntry) error {
	_, err := fmt.Fprintf(r.cmd.OutOrStdout(), "✅ authentication setup complete\nprofile: %s\nauthentication: %s (%s)\nprojects discovered: %s\nnext: check project access, then run fbrcm projects update --auth %s\n", profileName(profile), auth.ID, authTypeLabel(auth.Type), rcdisplay.FormatCount(0, "project", "projects"), auth.ID)
	return err
}

func profileName(value string) string {
	if strings.TrimSpace(value) == "" {
		return config.DefaultProfileName
	}
	return value
}

func authTypeLabel(value string) string {
	switch value {
	case config.AuthTypeGoogle:
		return "Google OAuth"
	case config.AuthTypeOAuth:
		return "OAuth"
	case config.AuthTypeServiceAccount:
		return "service account"
	case config.AuthTypeGCloud:
		return "gcloud ADC"
	default:
		return value
	}
}
