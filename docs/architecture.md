# fbrcm module map

This is a navigation aid for maintainers. It describes the current package
boundaries and important data-flow rules.

## Top-level flow

```text
main.go
├── no arguments ────> tui/
├── mcp command ─────> mcp/
└── other arguments ─> cli/

cli/ ─┐
      ├──> ops/ ──> core/ ──> Firebase and local storage
mcp/ ─┘            ↑
tui/ ──────────────┘
```

`main.go` initializes logging, constructs `core.Core`, and selects CLI, TUI, or
MCP mode. Each frontend owns its startup and lifecycle. `core` never imports a
frontend; `ops` and `mcp` never import `cli`. Architecture tests enforce these
dependency boundaries.

## `core/`

The root `core` package is the domain facade used by the frontends and shared
workflows. Its files expose authentication, project discovery, Remote Config
reads and publication, parameter and condition trees, imports, drafts, history,
and cross-project promotion.

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

The facade keeps presentation and workflow callers out of storage and wire
packages. Thin delegating methods in root `core` files are intentional.

## `cli/`

`cli/app` builds the Cobra root, selects the process profile, initializes
offline mode, and maps command errors to exit codes. `cli/commands` contains
CLI-specific management commands and adapters for shared workflows.
`cli/operation` adapts operation definitions to Cobra parsing and human output;
`cli/commands/testutil` supplies command-test helpers.

The `mcp` descriptor in the CLI tree provides help, completion, and capability
metadata. `main.go` routes server execution to the `mcp` frontend.

## `ops/`

`ops/workflows` owns reusable operation definitions, defaults, typed results,
and workflow handlers. `ops/invocation` carries context, options, and I/O
between a frontend adapter and a workflow. `ops.Registry` binds structured
input to a fresh operation and builds its machine envelope. CLI and MCP share
publication policy; TUI workflows call `core` directly.

`ops/contract` owns the versioned JSON envelope, typed problem
classification, artifact wrapping, and capability/schema metadata.
`schemas` embeds generated Draft 2020-12 schemas and capability metadata;
`cmd/schemagen` regenerates both from the same definitions as CLI discovery.

Notable packages:

- `ops/workflows/project/import` owns the interactive/non-interactive import
  flow;
- `ops/workflows/get/table` owns the responsive parameter table.

`ops/shared` contains behavior used across workflows: target resolution,
flags, text and expression filtering, confirmation, terminal sizing, JSON
output, parameter search, prompt input, and batch-result helpers. Its machine
mode prevents confirmations and selection prompts from running under the
global JSON contract.

### Remote Config workflow pipeline

`ops/shared/rc` centralizes the read/transform/diff/validate/publish pipeline:

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

- `ops/shared/diffview` renders static side-by-side property diffs;
- `ops/shared/fileoutput` provides overwrite-safe private file output;
- `internal/terminal` owns reusable terminal progress and styles;
- `internal/jsonscan` supports order-aware JSON processing without becoming
  a public package.

## `mcp/`

The MCP frontend uses the Go MCP SDK for protocol sessions and host interactions.
`mcp/command.go` parses launch options; `mcp/init.go` owns stdio lifecycle and
signal handling. Tool calls use `ops.Registry` directly, with fresh options,
context, results, and warning state for each invocation.

`mcp/server/catalog.go` maps shared operation IDs to public tool names, toolsets,
and mutation policy. The launch configuration determines which tools and options
are exposed. `schema.go` and `defaults.go` specialize operation schemas and
normalize input; see [MCP schema adaptation](../schemas/README.md#mcp-schema-adaptation).

`mcp/server/server.go` owns approval, OAuth interaction, cancellation, and
continuations. Continuation handles are bound to the operation, normalized
input, and connection. Completed interactive results are retained for one minute
to avoid replaying a mutation. The server retains at most 64 pending or recently
completed operations and serializes workflow execution because profile/cache
configuration is process-scoped.

OAuth recovery uses a temporary localhost callback listener, state validation,
and PKCE. The host offers the authorization URL after the listener is ready;
fbrcm exchanges the code directly with Google. Cancellation closes the listener.
For user-facing setup and recovery, see the [MCP guide](MCP.md).

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
- `core/styles` owns the runtime palette shared by CLI, logging, About, and the
  TUI; `tui/styles` owns reusable TUI render styles. See
  [theming.md](theming.md);
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
