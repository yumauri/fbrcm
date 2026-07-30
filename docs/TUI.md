# fbrcm TUI

Run `fbrcm` without arguments to open the interactive interface:

```sh
fbrcm
```

The minimum supported terminal size is 80 columns by 20 rows. A wider terminal
is recommended when comparing values or promoting changes.

## Mental model

The workspace has four main areas:

- **Projects** selects the Firebase projects and template targets in view.
- **Parameters**, **Conditions**, **History**, **A/B Tests**,
  **Personalizations**, and **Rollouts** are tabs over the main data panel.
- **Details** opens the selected parameter, group, condition, value, or managed
  feature.
- **Logs** shows live application activity and errors.

Promotion temporarily replaces the normal data tabs with a dedicated
source-to-target workspace.

The active profile appears in the upper-right corner. Press `?` anywhere to open
the searchable action palette. It uses the effective key configuration, hides
irrelevant actions, and explains why an action is unavailable.

## Setup and authentication

When the active profile has neither authentication nor cached projects, the TUI
opens guided setup instead of an empty workspace. It can:

1. import an OAuth Desktop app client JSON and complete browser authorization;
2. import a service-account JSON key;
3. validate existing [Google Cloud CLI](https://cloud.google.com/cli)
   Application Default Credentials;
4. discover the projects available to the selected identity.

OAuth authorization is cancellable. The authorization dialog shows the complete
URL and offers **Open Browser**, **Copy Link**, and **Cancel**. It closes after a
successful callback.

If valid credentials return no projects, setup offers a retry, another identity,
or an empty workspace. Cached projects can still open when no usable identity is
currently configured.

From the workspace:

| Default key | Action |
| --- | --- |
| `Ctrl+A` | Open Accounts |
| `Ctrl+P` | Open Profiles |
| `?` | Open the action palette |

Accounts can add, validate, and remove identities. They also show how many
cached projects use each identity. Profiles separate authentication, projects,
caches, and drafts; they can be created, switched, and renamed from the Profiles
tab. An inactive profile can also be removed; the active profile cannot.

If `FBRCM_PROFILE` selected the current profile, the Profiles tab treats it as
pinned. Restart without the environment variable to switch profiles
interactively.

### OAuth Desktop app

Create a Desktop app OAuth client in Google Cloud:

1. Open [Google Cloud OAuth clients](https://console.cloud.google.com/auth/clients).
2. Select or create the Google Cloud project that will own the OAuth client.
3. Choose **Create client**, then the **Desktop app** application type.
4. Download the client JSON and import it during setup.

Authorization opens a browser and waits on a local loopback callback. If the
browser cannot open, copy the complete URL from the TUI dialog.

The equivalent CLI flow is:

```sh
fbrcm auth add oauth default --from /path/to/client-secret.json
fbrcm auth login default
```

### Service account

Import a service-account key during setup, or use:

```sh
fbrcm auth add service-account production \
  --from /path/to/service-account.json
```

The account needs the relevant project discovery and Remote Config permissions.

### Google Cloud CLI Application Default Credentials

Create ADC first, then select the gcloud identity during setup:

```sh
gcloud auth application-default login
fbrcm auth add gcloud default
```

## Navigation

These defaults work throughout the workspace:

| Default key | Action |
| --- | --- |
| `1` | Focus Projects |
| `2` | Focus Parameters |
| `3` | Focus Conditions |
| `4` | Focus History |
| `5` | Focus Details |
| `6` | Focus A/B Tests |
| `7` | Focus Personalizations |
| `8` | Focus Rollouts |
| `9` | Focus the open Promote workspace |
| `0` | Focus Logs |
| `Tab` | Focus the next panel |
| `\` | Open the hidden workspace tabs menu |
| `q` | Quit |
| `Ctrl+C` | Force quit |

Most lists use the arrow keys or `j`/`k`. `Home` and `End` jump to the first and
last item. Panel-specific alternatives are shown in the footer and action
palette.

When the data panel is wide enough, its header shows every workspace tab. At
narrower widths, it keeps a sequence of complete titles on the left, places a
`\≡` overflow button before one trailing title, and hides the remaining titles
instead of shortening them to individual letters. The trailing title is
normally the next title after the left-hand sequence. Selecting a later hidden
tab moves its complete active title into that position; if necessary, another
left-hand title is hidden to make room. The current-profile badge width is also
reserved so it never covers a workspace title. Compact fitting reserves the
widest title that can occupy the trailing slot, so opening the menu cannot
suddenly hide another leading title.

Press `\` or click the overflow button to open the collapsed titles as one
ordered vertical stack. Its current title occupies the panel-border row and
later titles appear below it. `Down`/`j` previews the next title in the
panel-border row and shifts the stack upward; `Up`/`k` previews the previous
title and shifts it downward. Previewing does not switch the active panel or
recalculate which leading titles fit. The popup's complete border moves with
the stack: after rows move above the terminal, its vertical sides continue at
the top edge and the cursor row has no internal separator. A `▸` in the
overflow-glyph column marks the title that Enter will select and inherits the
active or inactive border color. The actually active workspace title keeps its
highlighted background wherever it appears in the visible stack. The menu keeps
two spaces between its longest title and the right border.
Consequently, opening the menu with Rollouts in the trailing slot starts at the
maximum upward shift, with Rollouts in the border row. `Home` and `End` move the
first and last collapsed titles into that row. `Enter` activates the border-row
title and closes the menu; clicking a visible title activates it immediately.
`Esc` or `\` closes the menu, and the numbered workspace shortcuts continue to
work while it is open. `q` follows the normal quit flow while the menu is open.

`q` quits immediately unless an open Details form has unsaved edits. In that
case, fbrcm asks before discarding them. `Ctrl+C` always exits immediately.
Accounts and Profiles cannot open while Details contains unsaved edits.

## Projects

Selecting one project shows its parameters. Marking several projects combines
their trees for comparison.

| Default key | Action |
| --- | --- |
| `Enter` | Select the focused project |
| `Space` | Mark or unmark a project |
| `u` | Synchronize projects |
| `o` | Open Remote Config in the Firebase console |
| `b` | Bind the focused or marked projects to another identity |
| `t` | Toggle the focused project's template rows |
| `p` | Make the focused template primary |
| `i` | Import Remote Config |
| `e` | Export Remote Config |
| `d` | Download application defaults |
| `v` | Promote to another project or template |
| `x` | Forget the focused or marked projects locally |

Forgetting a project removes its local registry entry, cached client and server
templates, cached versions, and drafts. It does not delete or modify the
Firebase project.

### Client and server templates

Projects initially use the client template. Press `t` to show both the client
row and the `server@` row. Press `p` to choose the primary template. When both
rows are visible, pressing `t` keeps only the focused template and makes it
primary.

The selection is local and controls unqualified CLI project targets as well as
the default TUI row. It does not create or delete Firebase templates. Explicit
CLI targets such as `client@project-id` and `server@project-id` always override
the saved preference.

## Parameters and groups

The Parameters tree keeps groups, parameters, default values, and conditional
values visible in their hierarchy. Empty and description-only groups remain
first-class items.

| Default key | Action |
| --- | --- |
| `Enter` | Open Details |
| `Space` | Expand or collapse the selected item |
| `Right` / `l` | Expand |
| `Left` / `h` | Collapse |
| `z` | Maximize or restore the data workspace |
| `a` | Add a parameter |
| `A` | Add an empty parameter group |
| `c` | Duplicate a parameter |
| `e` | Edit the selected value |
| `r` | Rename the selected parameter or group |
| `m` | Move the selected parameter or group |
| `x` | Delete the selected item |
| `u` / `U` | Refresh the current project / all projects |
| `p` / `P` | Publish the current draft / all drafts |
| `d` / `D` | Discard the current draft / all drafts |
| `y` / `Y` | Copy the selected name / full path |

Group deletion is explicit and removes the group with all of its parameters.
Parameter mutations, imports, draft merges, and promotions otherwise preserve
empty groups and group descriptions.

## Conditions

Conditions are shown in Firebase evaluation order. Details include the raw
expression, display color, priority, and every parameter value that refers to
the condition.

| Default key | Action |
| --- | --- |
| `Enter` | Open condition Details |
| `a` | Add a condition |
| `r` | Rename a condition and its references |
| `e` | Edit the raw Firebase expression |
| `c` | Change the display color |
| `m` | Move evaluation priority |
| `x` | Delete the condition and its conditional values |
| `p` / `P` | Publish the current draft / all drafts |
| `d` / `D` | Discard the current draft / all drafts |

Condition order affects value resolution, so moves are reviewed as Remote
Config changes rather than treated as cosmetic sorting.

## A/B Tests, Personalizations, and Rollouts

The three managed-feature tabs are read-only and show entities grouped by
selected Firebase project. Press `Enter`, or double-click an entity, to open its
Details. A tab loads each selected project the first time it is shown and keeps
that content when focus moves between Projects and the tab. Use `u` to update
the project under the cursor or `U` to update every project in the active tab.
Newly selected projects are loaded when the tab is next shown.

These tabs apply only to selected client Remote Config templates. A
`server@project` template is omitted because Firebase exposes managed features
under the client `firebase` namespace, not the `firebase-server` namespace.

- **A/B Tests** lists experiment IDs, display names, states, start times,
  last-updated times, descriptions, and published-template binding counts.
  Its `~`, `^`, `/`, and `=` filters match display names only, not descriptions.
  Opening
  Details reuses that list metadata immediately, then lazily refreshes it and
  loads the selected experiment's variants and objectives from its exact
  resource. Details also shows affected parameters, exposure, and each
  template variant value or no-change marker.
- **Personalizations** shows personalization IDs and the parameter-value
  bindings found in the published Remote Config template. Firebase does not
  expose value candidates or result metrics through this API.
- **Rollouts** combines rollout metadata from the public rollout API with the
  parameter-value bindings found in the published template.

An explicit zero percentage is rendered as `0%`, and an explicit empty managed
value is rendered as `""`; both remain distinct from missing data. If a project
load fails, the project row shows the error and the tab retries it when
activated again. Reloading a list refreshes matching open Details without
changing its scroll position. Navigating to a managed feature while an editable
Details form has unsaved changes uses the same save-or-discard flow as other
Details navigation.

The lists use `Up`/`Down` or `j`/`k`, `Page Up`/`Page Down` or `h`/`l`, and
`Home`/`End`. `u` updates the current project, `U` updates all projects, and `z`
maximizes or restores the workspace.

## Details and editors

Details is both an inspector and an edit form. The available actions depend on
the selected entity. Managed-feature Details are read-only and use `Esc` to
close.

| Default key | Action |
| --- | --- |
| `Esc` | Close Details or cancel the current editor |
| `Right` / `e` | Edit the selected value or condition expression |
| `a` | Add a conditional value |
| `r` | Rename the selected entity |
| `m` | Move the selected entity or condition priority |
| `x` | Delete the selected entity |
| `d` | Toggle the selected value's in-app-default source |
| `Ctrl+Enter` | Save Details |
| `Ctrl+Y` | Copy the selected value or expression |

Inside a form, `Enter` finishes the active field or invokes that row's
contextual action. If no project draft exists, saving opens a
publish-or-draft confirmation. Once the project has a draft, later edits stage
into it immediately.

A value using Firebase's in-app-default source cannot be edited as a remote
value. Press `d` again to return it to a neutral remote value for its type:
`""`, `false`, `0`, or `{}`.

The JSON editor supports `Ctrl+F` to format and `Ctrl+S` or `Ctrl+Enter` to save.
The expanded string editor uses `Ctrl+E` to switch size and `Ctrl+S` or
`Ctrl+Enter` to save.

## Draft workflow

Drafts are local, profile-scoped, and independent for client and server
templates.

The first mutation to a project can be published immediately or saved as a
draft. Once a draft exists, subsequent TUI edits compose onto it. The project
row indicates that the displayed template includes draft state.

Publishing a draft:

1. fetches current Firebase state;
2. performs a three-way merge of the draft base, local intent, and current
   template;
3. shows the exact candidate that would be sent;
4. validates and publishes that candidate with ETag protection.

Conflicts and failed validation preserve the draft. Publishing several drafts
is non-atomic: every project is reviewed and attempted independently, and a
results dialog reports all outcomes.

## Import, export, and defaults

The Projects-panel cursor—not marked-project state—determines the target of
`i`, `e`, and `d`.

### Import

The import wizard accepts raw Firebase Remote Config JSON or an fbrcm parameter
cache file. It supports:

- merge or replacement;
- group and parameter selection;
- text and expression filtering;
- condition cleanup;
- per-conflict keep-current or use-import choices.

The resulting diff is always reviewed. A new import can be published or saved
as a draft. If the target already has a draft, the import updates that draft
and leaves publication to the normal draft action.

### Export

When a draft exists, export asks whether to use published Firebase state or the
local draft. Output uses stable JSON normalization and private file
permissions. Replacing an existing file requires confirmation.

### Application defaults

Defaults can be downloaded as JSON for Web, XML for Android, or plist for Apple.
The downloaded file contains backend default values, not the complete Remote
Config template.

## History

History compares two published versions of each selected project. The initial
pair is the previous and current version.

| Default key | Action |
| --- | --- |
| `Enter` | Open the selected property's diff |
| `c` | Show only changes |
| `,` / `.` | Move both sides to an older / newer pair |
| `v` | Open the version picker |

In the picker, `Tab` switches sides, arrows choose versions, `r` restores the
default previous/current pair, and `R` begins a native Firebase rollback to the
active historical version. Rollback is blocked while the project has an
unpublished draft, shows a diff, and requires confirmation.

Firebase history is authoritative, but fbrcm also retains immutable templates
it encounters. The CLI can restore a cached snapshot that Firebase no longer
retains; see [version history](CLI.md#remote-config-version-history).

## Promotion

Press `v` on a project or template row, then choose a target. Promote shows a
clear source-to-target direction and a selectable list of parameter, condition,
and group-description changes.

| Default key | Action |
| --- | --- |
| `Space` | Select or clear the focused change |
| `a` / `n` | Select all visible changes / clear visible changes |
| `Enter` | Open the focused change in the diff viewer |
| `s` | Swap source and target |
| `x` | Include or exclude target-only removals |
| `u` | Use the source draft or published source |
| `d` | Save the selected result as a target draft |
| `p` | Review and publish the selected result |
| `Esc` | Close Promote |

Selecting a parameter automatically includes any condition definition or
updated group description required by it. Target-only items remain unchanged
unless pruning is enabled.

The source uses its local draft by default when one exists; `u` switches between
draft and published source. The target always starts from its effective local
state, including an existing draft. Cached-only reviews can be inspected and
saved as drafts, but publication requires a verified Firebase snapshot.

Before publishing, fbrcm saves the exact selected candidate as a target draft.
If validation or publication fails, that draft remains available for recovery.

## Filtering

Projects, Parameters, Conditions, History, A/B Tests, and Promote support live
text filtering:

| Prefix or key | Mode |
| --- | --- |
| `~` | Fuzzy match |
| `^` | Starts with |
| `/` | Contains |
| `=` | Exact match |

Projects, Parameters, Conditions, and History also support expression filtering
with `:`. Promote and A/B Tests use only the four text modes above. A/B Tests
matches experiment display names only; it does not search descriptions or
resource IDs.

Text and expression filters remember separate input. While an expression is
temporarily invalid, the input turns red, a compact compiler error appears on
the panel border, and the last valid result remains visible.

Expression context depends on the panel:

- Projects uses project context;
- Parameters and History use parameter context;
- Conditions uses condition context;

History evaluates the newer state, or the older state for a removed parameter.
See [Expression filtering](EXPR.md) for fields, typing, and helpers.

## Logs

Press `0` to focus Logs.

| Default key | Action |
| --- | --- |
| `c` | Collapse or expand |
| `[` / `]` | Lower / raise the log threshold |
| `-` / `+` | Shrink / grow the panel |
| `Enter` | Insert a visual blank line |

Mouse reporting is disabled while Logs is active and no interactive popup is
open, so normal terminal text selection continues to work.

## Mouse support

Selectable rows support left-click selection. Double-clicking the same item
performs its Enter action. Visible buttons activate on a single left click.
Scrolling and popup hit regions follow their rendered positions.

## Configuration

The global TUI configuration is stored in `config.toml` under the fbrcm config
root. Use the local-only `config` commands to inspect and edit it:

```sh
fbrcm config path
fbrcm config show
fbrcm config validate
fbrcm config edit
```

`config edit` resolves the editor in this order:

1. `--editor`;
2. `FBRCM_EDITOR`;
3. `VISUAL`;
4. `EDITOR`;
5. `vi` on Unix-like systems or `notepad.exe` on Windows.

Editor commands may include arguments. GUI editors should wait for the file to
close, for example:

```sh
FBRCM_EDITOR="code --wait" fbrcm config edit
```

If the edited file is invalid, the original remains unchanged and fbrcm reports
the staged recovery file.

Every binding shown by `fbrcm config show` can be replaced. For example:

```sh
fbrcm config set keys.projects.refresh u ctrl+r
fbrcm config reset keys.projects.refresh
```

Configurable Powerline separators are enabled by default. Disable them when the
terminal font does not provide compatible Powerline glyphs:

```sh
fbrcm config set powerline_glyphs false
```

Configuration commands do not require a profile or network connection. See the
[CLI configuration reference](CLI.md#fbrcm-config-path) for validation and JSON
output contracts.
