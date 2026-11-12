<script setup>
import { VPImage } from 'vitepress/theme'
</script>

<Badge type="info" text="Since 1.1.0" />

# Evaluate a request with Context

Context previews how the Remote Config templates open in the TUI respond to a
set of request inputs. It can evaluate a local draft already in view. Inputs
and scope are kept for the current TUI session. Changes to either affect the
evaluation view.

## Add inputs or import a request

Press `9` or click **Context** below Projects to focus the panel. `Ctrl+X`
opens the Context popup. To try a web request, press `a`, choose **Platform**,
and complete the prefilled assignment as `device.os=web`.

**User property** and **Custom signal** suggest names from the conditions used
by parameters in the selected templates. Choose **Enter another name…** to
type your own. Choose **Other field…** to enter another supported `field=value`
assignment. Use `/unknown` for an unresolved fact and `/unset` for a field that
is absent. See [CLI input examples](/cli/explain#try-a-scenario-without-a-capture).

To use a real request, press `i` in Context to open **Import Context**. Paste
the contents of a **Copy as cURL** export or HAR directly into the dialog, or
choose **Choose File** to read a saved capture. If there are several Firebase
Remote Config fetches, pick the request by its original one-based HAR entry
number. Import replaces the current inputs, resumes evaluation, and sets
**Template scope** to the captured project and template. fbrcm parses the
capture locally and never replays it.

## Scope the preview

With no scope, Context evaluates all selected templates. To limit it, press
`a` and choose **Template scope**. Enter a physical Firebase project ID or
project number for a client template, or prefix it with `server@` for a server
template. Edit or remove the scope row to change where the same inputs apply.
Removing the scope leaves the inputs in place.

## Inspect the outcome

Context hides conditions that are definitely false in **Conditions** and hides
their values in **Parameters** and **Details**. Once a parameter has a definite
winner, Parameters and Details show only that value. An unresolved
higher-priority condition can leave other values in view. The complete template
remains available to the editor.

Focus a parameter in **Parameters** or **Details** and press `Ctrl+E` to open a
scrollable explanation for that parameter. Context must be active and apply
to that template. The report shows the rule trace, condition priorities,
fallback, and any response evidence from a capture of that same project and
template type. From Details, the report also previews pending edits to the
parameter. Use the arrow keys or `Page Up` / `Page Down` to scroll and `Esc` to
close it.

<VPImage
  :image="{
    light: '/tui-context-light.png',
    dark: '/tui-context-dark.png',
    alt: 'Context explanation showing the Android beta condition enabling passkey signup for a request in the fifth percentile'
  }"
/>

| Default key in Context | Action |
| --- | --- |
| `a` | Add an input or set the template scope |
| `i` | Import cURL or HAR capture data |
| `Enter` or double-click | Edit the selected input or scope |
| `Delete` or `x` | Remove the selected input or scope |
| `H` | Show or hide excluded rows and values with their outcome notes |
| `Space` | Pause or resume evaluation while keeping the inputs |
| `X` | Clear the inputs, scope, and imported capture |
| `c` | Collapse or expand the Projects and Context sidebar from either panel |

Context shortcuts can be customized under `keys.context.*`,
`keys.global.open_context`, and `keys.global.explain_context`. Use `?` to find
the current bindings in the action palette.

For a scriptable report against the published template, use [CLI `explain`](/cli/explain).
