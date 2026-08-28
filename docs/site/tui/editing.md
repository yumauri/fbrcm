# Editing and drafts in the TUI

The Details panel shows the selected item and contains its edit form. Select a
parameter, group, condition, or value and press `Enter` to open it.

## Parameters and groups

The Parameters tree preserves the Remote Config hierarchy: groups contain
parameters, and parameters contain a default plus any conditional values.
Empty and description-only groups remain visible.

| Key | Action |
| --- | --- |
| `Space` | Expand or collapse |
| `a` | Add a parameter |
| `A` | Add an empty group |
| `c` | Duplicate a parameter |
| `e` | Edit the selected value |
| `E` | Edit text or JSON externally |
| `r` | Rename a parameter or group |
| `m` | Move a parameter or group |
| `x` | Delete the selected item |
| `y` / `Y` | Copy the name or full path |

Group deletion is an explicit group-level operation and removes every parameter
inside it. Other parameter mutations preserve empty groups and descriptions.

## Conditions

Conditions appear in Firebase evaluation order. Open one to inspect its raw
expression, display color, priority, and every parameter value that refers to
it.

| Key | Action |
| --- | --- |
| `a` | Add a condition |
| `e` | Edit the Firebase expression |
| `c` | Change the display color |
| `r` | Rename the condition and its references |
| `m` | Change evaluation priority |
| `x` | Delete the condition and its conditional values |

Priority changes are Remote Config changes, not cosmetic sorting, because
Firebase evaluates conditions in order.

## Save an edit

Inside Details:

| Key | Action |
| --- | --- |
| `Esc` | Close or cancel the current editor |
| `Right` / `e` | Edit the selected field |
| `a` | Add a conditional value |
| `d` | Toggle the in-app-default source |
| `Ctrl+Enter` | Save Details |
| `Ctrl+Y` | Copy the selected value or expression |

If the target has no draft, saving opens a choice to publish immediately or
save a draft. Once a draft exists, later edits stage into it automatically.

Every mutation review includes an optional change note. A saved draft remembers
the note; publishing sends it to Firebase as the new version description.

## Edit large text and JSON

Press `E` to pause the TUI and edit a private staged file with an external
editor. fbrcm resolves the command from `FBRCM_EDITOR`, `VISUAL`, then `EDITOR`.
GUI editors must wait for the file to close:

```sh
FBRCM_EDITOR="code --wait" fbrcm
```

After the editor exits, fbrcm validates and compacts JSON. It rejects invalid
content and keeps the recovery file so you can reopen it.

The built-in JSON editor uses `Ctrl+F` to format and `Ctrl+S` or `Ctrl+Enter` to
save. Large values automatically use the external-editor path.

## Review and publish drafts

The project row indicates when the displayed template includes draft state.

| Key | Action |
| --- | --- |
| `p` | Publish the current draft |
| `P` | Publish all drafts |
| `d` | Discard the current draft |
| `D` | Discard all drafts |

Publication fetches current Firebase state, rebases local intent with a
three-way merge, shows the exact candidate, validates it, and publishes with
ETag protection. Conflicts and validation failures keep the draft intact.

## Import, export, and promotion

In Projects, `i`, `e`, and `d` operate on the project under the cursor, not the
set of marked projects.

- Import supports merge or replacement, item selection, filtering, condition
  cleanup, and per-conflict choices.
- Export writes the effective Remote Config template.
- Application defaults downloads platform-specific defaults from Firebase.
- Promotion compares one source target with another. You select the parameters,
  groups, and conditions to transfer.

All resulting writes rejoin the same review, draft, and publication workflow.
