# fbrcm Module Map

This document is a navigation aid for the codebase. It describes the package
boundaries and dependency direction so refactors stay within the intended
layering. It is descriptive of the current code, not aspirational.

## Top-level layout

```
main.go            Entry point: picks CLI vs TUI mode, builds core.Service.
core/              Domain + infrastructure layer (no CLI/TUI imports).
cli/               Cobra command tree (depends on core).
tui/               Bubble Tea v2 app (depends on core).
```

Dependency direction is one-way: `cli/` and `tui/` depend on `core/`; `core/`
never imports `cli/` or `tui/`.

## core/

| Package | Responsibility |
| --- | --- |
| `core` | `Core` facade: auth registry, project sync, remote-config export/validate/publish/import, parameters cache + `ParametersTree` view model (tree types in `core/parameters`). Draft lifecycle delegates to `core/draft`. |
| `core/draft` | Draft storage, RC slot mutations, three-way merge, mutate/preview/publish pipeline. |
| `core/parameters` | Parameters view model: tree/group/entry/value types, tree building from Remote Config, display value formatting. |
| `core/rc/display` | Remote Config display formatting: summary vs diff modes, project/condition labels. |
| `core/rc/diff` | Colored Remote Config diff rendering for CLI and TUI previews. |
| `core/rc/mutate` | RC slot collection and in-memory parameter/group mutation. |
| `core/rc/value` | Parameter value validation and JSON number checks. |
| `core/config` | On-disk persistence: auth config, profiles, projects cache, parameter cache, drafts, path resolution. |
| `core/firebase` | Firebase/Google wire types and HTTP API: resilient transport, auth (oauth/gcloud/service-account/token), dry-run and offline gating, remote-config endpoints. |
| `core/filter` | Filtering: mode-prefixed fuzzy/exact/prefix/includes matching (`filter.go`) and expr-lang expression engine with jq support (`expr.go`). |
| `core/env`, `core/log`, `core/styles`, `core/browser` | Cross-cutting helpers: env overrides, logging, shared styles, browser opener. |
| `core/strfold` | Case-insensitive string compare/sort helpers and unified project sort order. |

## cli/

| Package | Responsibility |
| --- | --- |
| `cli/app` | Root command assembly and top-level error handling. |
| `cli/commands/*` | One package per command group (`add`, `auth`, `cache`, `config`, `delete`, `get`, `profile`, `project`, `projects`, `update`). |
| `cli/shared` | Reusable command plumbing: flags, project/parameter filtering, confirmation prompts, JSON input. |
| `cli/shared/rc` | Remote Config CLI pipeline: input extraction, order-preserving JSON, diff rendering, export normalization, validate/publish with ETag retry. Imported directly by RC mutation commands (`add`, `delete`, `update`, `get`, `project`). |
| `cli/styles` | CLI palette and `NO_COLOR` handling. |

### `cli/shared/rc` layout

The RC pipeline lives in one subpackage with a clear file boundary:

| File | Responsibility |
| --- | --- |
| `input.go` | Read stdin / cache payloads; extract embedded `remote_config` JSON. |
| `order.go` | Parse and marshal Remote Config JSON while preserving member order. |
| `diff.go` | Diff and conflict preview helpers (delegates to `core/rc/diff` and `core/rc/display`). |
| `normalize.go` | Stable export JSON entry points: escape normalization delegates to `core/firebase`. |
| `normalize_conditional.go` | Conditional-value key reordering and JSON scanning helpers. |
| `publish.go` | Validate/publish with ETag conflict detection and project mutation wrapper. |
| `loop.go` | Multi-project revalidate → mutate → publish loop with retry on stale ETag. |
| `project.go` | Revalidated per-project config snapshot used by the publish loop. |
| `output.go` | Order-preserving stdout writer for stdin mutation commands. |
| `conflict.go` | ETag/precondition conflict detection for publish retries. |

Commands that mutate or display Remote Config import `cli/shared/rc` for the
pipeline and `cli/shared` for shared flags, filtering, and confirmation. Display
formatting for parameter headers in prompts uses `core/rc/display` directly where
only formatting is needed.

## tui/

| Package | Responsibility |
| --- | --- |
| `tui/app` | Root Bubble Tea model that orchestrates panels, overlays, value editors, and draft dialogs. |
| `tui/components/*` | Panels and overlays: `projects`, `parameters` (`view.go`, `view_layout.go`, `view_render.go`), `details` (`model.go`, `model_fields.go`, …), value editors (`boolpicker`, `numberinput`, `stringinput`, `jsoninput`), `dialog`, `filterbox`, `logs`, `moveparam`, `renameinput`, `minsize`, and `viewutil` helpers. |
| `tui/config`, `tui/messages`, `tui/panels`, `tui/styles` | Key bindings, inter-component messages, panel identifiers, panel styles. |

## Charm stack note

The project is on the `charm.land/*` v2 line: bubbletea, bubbles, lipgloss, and
log. `core/log` uses `charm.land/log/v2` with `charm.land/lipgloss/v2` for
custom styles and `github.com/charmbracelet/colorprofile` for color profile
detection (replacing direct `termenv` usage in the logger).

## Refactor guidelines

Use these when splitting files or cleaning up structure. They describe how the
codebase is maintained today, not a future rewrite.

### Pass size and scope

- Keep refactors **behavior-preserving** unless a functional change is explicitly requested.
- Prefer **small, reviewable passes** (one concern per PR/commit): dedupe, split one oversized file, extract one helper, add tests for the touched layer.
- Target **~200–300 lines** for new files; split when a file makes keybinding, overlay, or RC pipeline changes hard to review.
- Do not split files under ~200 lines just to hit a line-count target.

### Layering and public APIs

- Respect dependency direction: `cli/` and `tui/` → `core/` only.
- Keep **public APIs stable** (`core` facade methods, exported CLI/TUI config helpers) unless the refactor requires a break — then split that break into its own migration task.
- Thin facades in `core/` (for example draft re-exports) are intentional; do not remove them during structural passes.

### Tests per layer

Run before and after every pass:

```bash
golangci-lint run ./...
go test -race ./...
```

Add or extend tests at the layer you touch:

| Layer | Expected guard |
| --- | --- |
| `core/draft`, `cli/shared/rc` | Unit tests with fixtures under `testdata/remoteconfig/` |
| `core/firebase` | Unit tests for transport, dry-run, offline, and normalize helpers |
| `tui/components/*` | `view_parity_test.go` snapshot tests when changing render output |
| `tui/config` | Keybinding tests for `Matches()` and conflict disabling |

### Invariants

- Root group key representations (`""`, `__default__`, `(root)`) must stay consistent — see [root-group-key.md](root-group-key.md).
- Private file I/O goes through `core/config.WritePrivateFile`.

### Stop criteria

Skip a refactor pass when it cannot name at least one of:

1. Duplicated logic to consolidate
2. An oversized module blocking changes
3. An untested path with real regression risk

Dependency upgrades, Charm import unification, live Firebase integration tests, and `core` facade narrowing belong in **separate migration tasks**, not mixed into hygiene passes.
