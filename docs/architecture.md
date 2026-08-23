# fbrcm module map

This is a navigation aid for maintainers. It describes the current package
boundaries and important data-flow rules.

## Top-level flow

```text
main.go
├── no arguments ──> tui/
└── any arguments ─> cli/

cli/ ─┐
      ├──> core/ ──> Firebase and local storage
tui/ ─┘
```

`main.go` initializes logging, constructs `core.Core`, and selects CLI or TUI
mode. `cli` and `tui` depend on `core`; `core` never imports either presentation
layer.

## `core/`

The root `core` package is the facade used by both front ends. Its files expose
authentication, project discovery, Remote Config reads and publication,
parameter and condition trees, imports, drafts, history, and cross-project
promotion.

| Package | Responsibility |
| --- | --- |
| `core/config` | Global config, profiles, auth registry, project registry, caches, drafts, private file I/O, and path resolution |
| `core/firebase` | Google credentials, resilient HTTP transport, shared request pacing and 429 cooldowns, project APIs, Remote Config APIs, defaults, history, rollback, diagnostics, and offline gating |
| `core/parameters` | Parameter/group/value tree models and display values |
| `core/conditions` | Ordered condition definitions, usage indexing, validation, and mutation impact |
| `core/groups` | Explicit parameter-group add, edit, rename, and delete operations |
| `core/draft` | Draft envelopes, mutation composition, three-way merge, publish preparation, and cleanup |
| `core/filter` | Mode-prefixed text matching and expr-lang evaluation with embedded gojq |
| `core/rc/diff` | Structured and human-readable Remote Config comparison |
| `core/rc/diffinput` | Conversion of Remote Config entities into generic property diffs |
| `core/rc/display` | Shared labels, counts, and display formatting |
| `core/rc/importer` | Import selection, condition cleanup, merge planning, and conflict resolution |
| `core/rc/mutate` | Wire-level parameter slot collection and mutation |
| `core/rc/promote` | Promotion plans, dependencies, selection, pruning, and application |
| `core/rc/target` | Parsing and canonicalization of client/server template targets |
| `core/rc/value` | Parameter value validation and JSON number handling |
| `core/rootgroup` | Canonical root-group constants |
| `core/dictdiff` | Ordered generic dictionary/property comparison |
| `core/strfold` | Case-insensitive comparison, stable keys, and project sorting |
| `core/browser`, `core/env`, `core/log`, `core/styles` | Cross-cutting browser, environment, logging, and style helpers |

The facade keeps CLI/TUI callers out of storage and wire packages. Thin
delegating methods in root `core` files are intentional.

## `cli/`

`cli/app` builds the Cobra root, selects the process profile, initializes
offline mode, and maps command errors to exit codes. `cli/contract` owns the
versioned JSON envelope, typed problem classification, artifact wrapping, and
capability discovery. `schemas` embeds the generated per-command Draft 2020-12
schemas; `cmd/schemagen` regenerates those schemas and the capability golden.

`cli/commands` contains one package per top-level feature:

```text
add  auth  cache  conditions  config  delete  doctor  draft
duplicate  get  groups  profile  project  projects  update  versions
```

Notable nested packages:

- `cli/commands/project/import` owns the interactive/non-interactive import
  flow;
- `cli/commands/get/table` owns the responsive parameter table;
- `cli/commands/testutil` contains command-test helpers.

`cli/shared` contains behavior used across command packages: target resolution,
flags, text and expression filtering, confirmation, terminal sizing, JSON
output, parameter search, prompt input, and batch-result helpers. Its machine
mode prevents confirmations and selection prompts from running under the
global JSON contract.

### Remote Config CLI pipeline

`cli/shared/rc` centralizes the read/transform/diff/validate/publish pipeline:

| File | Responsibility |
| --- | --- |
| `input.go` | Read raw Remote Config or unwrap an fbrcm cache payload |
| `order.go` | Parse and marshal JSON while preserving member order |
| `normalize.go`, `normalize_conditional_reorder.go`, `normalize_conditional_scan.go` | Stable export normalization |
| `diff.go` | Shared Remote Config preview formatting |
| `publish.go`, `loop.go`, `project.go` | Revalidation, mutation, validation, ETag-protected publication, and batch flow |
| `conflict.go` | Precondition and ETag conflict classification |
| `output.go` | Order-preserving stdout for stdin transformations |

Other shared subpackages:

- `cli/shared/diffview` renders static side-by-side property diffs;
- `cli/shared/fileoutput` provides overwrite-safe private file output;
- `cli/internal/jsonscan` supports order-aware JSON processing without becoming
  a public package.

## `tui/`

`tui/app` is the Bubble Tea root model. It coordinates focus, panel layout,
popups, editors, setup, profile transitions, draft publication, history
rollback, project I/O, and promotion.

Major components under `tui/components`:

| Component | Responsibility |
| --- | --- |
| `projects` | Project/template list, marking, selection, filtering, and mouse interaction |
| `parameters` | Parameters and groups, multi-project view, drafts, and version history |
| `conditions` | Condition order, definitions, usages, and editing |
| `details` | Contextual inspector and edit form |
| `promote` | Source/target picker and selectable promotion workspace |
| `diffview` | Interactive property and Remote Config comparison |
| `projectio` | TUI import, export, and defaults workflows |
| `setup`, `authpicker` | Guided authentication and identity selection |
| `filterbox` | Text and expression filter input |
| `logs` | Live logs and level controls |
| `dialog`, `buttonbar` | Modal choices and clickable actions |
| `boolpicker`, `numberinput`, `stringinput`, `jsoninput` | Typed value editors |
| `moveparam`, `renameinput` | Move and rename editors |
| `mouseutil`, `viewutil`, `minsize`, `inputstyles`, `workspaceheader` | Shared interaction and rendering primitives |

Supporting packages:

- `tui/config` defines defaults, loads overrides, migrates old bindings, and
  validates conflicts;
- `tui/messages` contains cross-component Bubble Tea messages;
- `tui/panels` defines focus identifiers;
- `tui/styles` owns TUI colors and reusable render styles;
- `tui/testutil` contains rendering-test helpers.

## Charm stack

The UI uses the `charm.land/*` v2 packages for Bubble Tea, Bubbles, Lip Gloss,
and logging. `github.com/charmbracelet/colorprofile` handles terminal color
profile detection.

## Important invariants

- Root group representations (`""`, `__default__`, and `(root)`) are
  layer-specific. See [root-group-key.md](root-group-key.md).
- Empty and description-only parameter groups must survive filtering,
  condition cleanup, parameter mutation, import, merge, draft handling, and
  promotion. Only an explicit group operation may remove a group.
- Condition slice order is Firebase evaluation priority. Alphabetical sorting
  is never semantically safe.
- Client and `server@` targets share a physical Firebase project but have
  independent template state, caches, drafts, and version histories.
- Draft publication is candidate-stable: preview, validation, and publication
  must refer to the same merged candidate. Conflicts preserve the draft.
- Promotion selection is atomic at parameter, condition, and group-description
  level. `core/rc/promote` is the source of truth for dependencies and pruning.
- Files containing credentials, drafts, caches, or exported Remote Config use
  the private-file helpers in `core/config`.
- Application configuration is layered: the nearest repository `.fbrcm.toml`
  deeply overlays the global `config.toml`. Stored layers remain sparse;
  built-in keybindings and values are applied in memory and must never be
  materialized or copied between layers during startup.
- Human CLI tables use the shared Lip Gloss conventions and never rely on
  terminal soft wrapping.
- Selectable TUI rows support click selection and double-click activation;
  visible buttons activate on one click.

## Validation

Run the repository checks after implementation changes:

```sh
golangci-lint run
go test -race ./...
```

Add tests at the layer being changed. Rendering changes should include the
relevant narrow-width, no-color, mouse-hit-region, or view-parity regression
coverage.
