package versions

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcdiffinput "github.com/yumauri/fbrcm/core/rc/diffinput"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
)

func newVersionsBlameCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "blame <project> <parameter>",
		Short: "Show who changed one Remote Config parameter",
		Args:  invocation.ExactArgs(2),
		RunE: func(cmd invocation.Call, args []string) error {
			return runVersionsBlame(cmd, svc, args)
		},
	}
	cmd.Flags().Int("limit", 1, "Maximum parameter changes to print")
	cmd.Flags().Bool("all", false, "Print all changes in retained Firebase history")
	cmd.Flags().String("at", "current", "Newest version to inspect")
	cmd.Flags().Bool("json", false, "Print parameter history as JSON")
	cmd.MarkFlagsMutuallyExclusive("all", "limit")
	return cmd
}

func runVersionsBlame(cmd invocation.Call, svc *core.Core, args []string) error {
	parameter := args[1]
	if strings.TrimSpace(parameter) == "" {
		return shared.InvalidArgument(fmt.Errorf("parameter must not be blank"))
	}
	if utf8.RuneCountInString(parameter) > 256 {
		return shared.InvalidArgument(fmt.Errorf("parameter %q exceeds 256 characters", parameter))
	}
	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 1 {
		return shared.InvalidArgument(fmt.Errorf("--limit must be greater than zero"))
	}
	all, _ := cmd.Flags().GetBool("all")
	at, _ := cmd.Flags().GetString("at")
	project, err := resolveVersionProject(cmd, svc, args[0], false)
	if err != nil {
		return err
	}
	ctx, err := shared.FirebaseServiceContextForExecution(shared.CommandContext(cmd), project.ProjectID)
	if err != nil {
		return err
	}
	cmd.SetContext(ctx)
	progress.Start("Scanning parameter history for " + project.ProjectID + "…")
	history, err := svc.GetRemoteConfigParameterHistory(ctx, project.ProjectID, parameter, core.ParameterHistoryOptions{At: at, Limit: limit, All: all})
	if err != nil {
		return err
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	if jsonOut {
		return shared.WriteJSON(cmd, versionBlameJSON(project, history))
	}
	text, err := renderVersionBlame(project, history, shared.TerminalWidth())
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), text)
	return err
}

type versionBlameResult struct {
	Project             core.Project             `json:"project"`
	Parameter           string                   `json:"parameter"`
	AtVersion           string                   `json:"at_version"`
	AtGroup             *string                  `json:"at_group"`
	ScannedVersionCount int                      `json:"scanned_version_count"`
	HistoryExhausted    bool                     `json:"history_exhausted"`
	Boundary            *versionBlameBoundary    `json:"boundary"`
	Changes             []versionBlameChangeJSON `json:"changes"`
}

type versionBlameBoundary struct {
	Version string  `json:"version"`
	State   string  `json:"state" contract:"enum=present|absent"`
	Group   *string `json:"group"`
}

type versionBlameChangeJSON struct {
	PreviousVersion string                    `json:"previous_version"`
	Version         string                    `json:"version"`
	UpdateTime      string                    `json:"update_time" contract:"format=date-time"`
	UpdateUser      firebase.RemoteConfigUser `json:"update_user"`
	ChangeNote      *string                   `json:"change_note"`
	UpdateOrigin    string                    `json:"update_origin"`
	UpdateType      string                    `json:"update_type"`
	RollbackSource  *string                   `json:"rollback_source"`
	Change          rcdiff.ParameterChange    `json:"change"`
}

func versionBlameJSON(project core.Project, history core.ParameterHistory) versionBlameResult {
	result := versionBlameResult{
		Project: project, Parameter: history.Parameter, AtVersion: history.AtVersion, AtGroup: history.AtGroup,
		ScannedVersionCount: history.ScannedVersionCount, HistoryExhausted: history.HistoryExhausted,
		Changes: make([]versionBlameChangeJSON, 0, len(history.Changes)),
	}
	if history.Boundary != nil {
		state := "absent"
		var group *string
		if history.Boundary.Present {
			state = "present"
			value := history.Boundary.Group
			group = &value
		}
		result.Boundary = &versionBlameBoundary{Version: history.Boundary.Version, State: state, Group: group}
	}
	for _, item := range history.Changes {
		result.Changes = append(result.Changes, versionBlameChangeJSON{
			PreviousVersion: item.PreviousVersion,
			Version:         item.Version.VersionNumber,
			UpdateTime:      item.Version.UpdateTime,
			UpdateUser:      item.Version.UpdateUser,
			ChangeNote:      optionalVersionString(item.Version.ChangeNote),
			UpdateOrigin:    item.Version.UpdateOrigin,
			UpdateType:      item.Version.UpdateType,
			RollbackSource:  optionalVersionString(item.Version.RollbackSource),
			Change:          item.Change,
		})
	}
	return result
}

func renderVersionBlame(project core.Project, history core.ParameterHistory, terminalWidth int) (string, error) {
	var output strings.Builder
	fmt.Fprintf(&output, "Project: %s (%s)\n", project.Name, project.ProjectID)
	parameterPath := history.Parameter
	if history.AtGroup != nil {
		parameterPath = strings.TrimPrefix(rcdiffinput.ParameterEntityName(*history.AtGroup, history.Parameter), "Parameter: ")
	}
	fmt.Fprintf(&output, "Parameter: %s\n", parameterPath)

	if len(history.Changes) == 0 {
		output.WriteString("\nNo attributable change found.\n")
		if history.Boundary != nil && history.Boundary.Present {
			fmt.Fprintf(&output, "Parameter %q already exists in the oldest retained version, %s.\n", history.Parameter, history.Boundary.Version)
			output.WriteString("Its original author and creation time are unavailable.\n")
		}
		return strings.TrimRight(output.String(), "\n"), nil
	}

	output.WriteString("\n" + blameRail("┆") + "\n")
	for index, item := range history.Changes {
		if index > 0 {
			output.WriteString(blameRail("┆") + "\n" + blameRail("┆") + "\n")
		}
		block, err := renderVersionBlameChange(item, terminalWidth)
		if err != nil {
			return "", err
		}
		output.WriteString(block)
		output.WriteByte('\n')
	}
	output.WriteString(blameRail("┆"))
	if history.HistoryExhausted && history.Boundary != nil && history.Boundary.Present {
		fmt.Fprintf(&output, "\n\nHistory boundary: parameter already exists in retained version %s; its earlier origin is unavailable.", history.Boundary.Version)
	} else if !history.HistoryExhausted {
		fmt.Fprintf(&output, "\n\nShowing %s after scanning %s.",
			rcdisplay.FormatCount(len(history.Changes), "change", "changes"),
			rcdisplay.FormatCount(history.ScannedVersionCount, "retained version", "retained versions"))
		output.WriteString(" Use --all to continue through retained history.")
	}
	return output.String(), nil
}

func renderVersionBlameChange(item core.ParameterHistoryChange, terminalWidth int) (string, error) {
	var output strings.Builder
	output.WriteString(renderVersionBlameHeading(item, terminalWidth))
	output.WriteByte('\n')

	author := versionBlameAuthor(item.Version.UpdateUser)
	published := formatFirebaseVersionTime(item.Version.UpdateTime)
	metadata := author
	if published != "" {
		metadata += " on " + published
	}
	appendBlameText(&output, metadata, 0, terminalWidth)
	if item.Version.RollbackSource != "" {
		appendBlameText(&output, "Rollback source: version "+item.Version.RollbackSource, 0, terminalWidth)
	}
	if note := strings.TrimSpace(item.Version.ChangeNote); note != "" {
		output.WriteString(blameRail("│") + "\n")
		appendBlameText(&output, note, 4, terminalWidth)
	}

	diffText, err := renderVersionSideBySide(rcdiff.Result{Parameters: []rcdiff.ParameterChange{item.Change}}, max(terminalWidth-3, 1))
	if err != nil {
		return "", err
	}
	if diffText != "" {
		output.WriteString(blameRail("│") + "\n")
		for line := range strings.SplitSeq(diffText, "\n") {
			output.WriteString(blameRail("│"))
			if line != "" {
				output.WriteString("  " + line)
			}
			output.WriteByte('\n')
		}
	}
	return strings.TrimRight(output.String(), "\n"), nil
}

func renderVersionBlameHeading(item core.ParameterHistoryChange, terminalWidth int) string {
	rail := "┝━ "
	transition := "version " + item.PreviousVersion + " → "
	version := item.Version.VersionNumber
	if clistyles.NoColorEnabled() {
		return rail + transition + version
	}

	background := clistyles.ColorInactiveSelection
	headingWidth := lipgloss.Width(rail + transition + version)
	padding := strings.Repeat(" ", max(terminalWidth-headingWidth, 0))
	return lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDark).Background(background).Render(rail) +
		clistyles.PanelMuted.Background(background).Render(transition) +
		lipgloss.NewStyle().Bold(true).Foreground(clistyles.PaletteSlateBright).Background(background).Render(version) +
		lipgloss.NewStyle().Background(background).Render(padding)
}

func appendBlameText(output *strings.Builder, value string, indent, terminalWidth int) {
	contentWidth := max(terminalWidth-3-indent, 1)
	for logical := range strings.SplitSeq(value, "\n") {
		wrapped := ansi.Hardwrap(logical, contentWidth, true)
		for line := range strings.SplitSeq(wrapped, "\n") {
			output.WriteString(blameRail("│") + "  " + strings.Repeat(" ", indent) + line + "\n")
		}
	}
}

func versionBlameAuthor(user firebase.RemoteConfigUser) string {
	name, email := strings.TrimSpace(user.Name), strings.TrimSpace(user.Email)
	switch {
	case name != "" && email != "":
		return name + " <" + email + ">"
	case email != "":
		return email
	case name != "":
		return name
	default:
		return "Unknown author"
	}
}

func blameMuted(value string) string {
	if clistyles.NoColorEnabled() {
		return value
	}
	return clistyles.PanelMuted.Render(value)
}

func blameRail(value string) string {
	if clistyles.NoColorEnabled() {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDark).Render(value)
}
