# Terminal UI

Run `fbrcm` without arguments to open the interactive interface:

```sh
fbrcm
```

The minimum supported terminal size is 80 columns by 20 rows. A wider terminal
is useful when comparing projects or promoting changes.

## Workspace layout

The interface has four main areas:

- **Projects** selects the Firebase projects and template targets in view.
- **Parameters**, **Conditions**, **History**, **A/B Tests**,
  **Personalizations**, and **Rollouts** share the main data workspace.
- **Details** inspects and edits the selected item.
- **Logs** shows live activity and errors.

The active profile appears in the upper-right corner. Promotion temporarily
replaces the normal tabs with a dedicated source-to-target workspace.

## Find actions instead of memorizing keys

Press `?` anywhere to open the searchable action palette. It shows only actions
relevant to the current panel, uses your configured key bindings, and explains
why unavailable actions are disabled.

The global defaults are:

| Key | Focus |
| --- | --- |
| `1` | Projects |
| `2` | Parameters |
| `3` | Conditions |
| `4` | History |
| `5` | Details |
| `6` | A/B Tests |
| `7` | Personalizations |
| `8` | Rollouts |
| `9` | Promote workspace |
| `0` | Logs |
| `Tab` | Next panel |
| `\` | Hidden workspace tabs |

Lists use arrow keys or `j`/`k`. `Home` and `End` jump to the first and last
items. Every selectable row supports left-click selection; double-clicking it
performs the same action as `Enter`.

## Select projects and templates

Selecting one project shows its parameters. Mark several projects with `Space`
to compare their trees.

| Key | Project action |
| --- | --- |
| `Enter` | Select the focused project |
| `Space` | Mark or unmark a project |
| `u` | Synchronize projects |
| `o` | Open Remote Config in Firebase Console |
| `b` | Bind projects to another identity |
| `t` | Show or hide client/server template rows |
| `p` | Make the focused template primary |
| `i` / `e` | Import or export Remote Config |
| `d` | Download application defaults |
| `v` | Promote to another project or template |

The client template is primary by default. Expand a project with `t` to expose
its `client@` and `server@` targets, then use `p` to change the default target.

## Work offline

At startup, fbrcm performs a short proxy-aware connectivity probe. If it fails,
the TUI enters offline mode and uses available cached data. Define
`FBRCM_OFFLINE` to skip the probe and start offline immediately:

```sh
FBRCM_OFFLINE=1 fbrcm
```

Standard requests honor `HTTPS_PROXY`, `HTTP_PROXY`, and `NO_PROXY`, including
their lowercase forms.

## Accounts, profiles, and themes

| Key | Popup |
| --- | --- |
| `Ctrl+A` | Accounts |
| `Ctrl+P` | Profiles |
| `Ctrl+T` | Themes |

Accounts manage identities and their project bindings. Profiles isolate
credentials, projects, caches, and drafts. The theme picker applies each valid
theme as a live preview before you save it.

Continue with [Editing and drafts](./editing) for parameter, group, condition,
and publication workflows.
