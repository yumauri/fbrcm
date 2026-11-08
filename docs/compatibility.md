# Compatibility guarantees from fbrcm 1.0.0

fbrcm 1.0.0 establishes the supported public interfaces below. The application
release, CLI JSON contract, MCP protocol, and persisted file formats have
separate versions. The CLI JSON contract starts at `1.0.0`; existing integer
file-format versions remain `1`. The pre-1.0 permission to break these
interfaces without migration or a version change has ended.

## 1.1.0 release boundary

Version 1.1.0 adds CLI `explain`, its structured JSON results and capture
inputs, MCP `parameters.explain`, and condition blame through CLI and MCP.
These additions follow the same compatibility policy as the 1.0.0 interfaces.

| Versioned surface | Boundary for this release |
| --- | --- |
| Application | fbrcm `1.1.0`; documented CLI behavior remains compatible throughout 1.x. |
| CLI JSON | Contract `1.1.0`. The published `1.0.0` schemas and archived release lock remain available. |
| MCP | Negotiated protocol version, with the documented public tool catalog, launch policy, schema adaptation and shared operation envelopes. The CLI lock does not independently freeze MCP adaptation. |
| Persisted formats | Existing integer versions remain `1`; versionless durable formats retain their established meaning. No encoding change or migration is introduced by this release. |

## Stable interfaces

| Interface | Guarantee within fbrcm 1.x |
| --- | --- |
| Documented CLI commands | Command paths, positional arguments, flags and aliases, accepted inputs, defaults, selector/filter/expression semantics, documented environment variables, confirmation requirements, side effects, and exit-status meanings remain compatible. |
| CLI JSON | Versioned envelopes, command IDs, schema URNs, DTO fields and meanings, problem/warning codes and details, remediation strategies, artifact contracts, capability metadata, and normative semantic annotations follow the contract-version rules below. |
| MCP | Public tool names and their operation mappings, launch options, discovered input/output contracts, wrapper defaults, result/error mapping, and write/hook/confirmation/authentication policy remain compatible for the same launch configuration and negotiated protocol capabilities. |
| Durable persisted data | Supported configuration, identities, user preferences, drafts, retained snapshots, and publication plans written by any released 1.x version remain readable and retain their meaning on upgrades to later 1.x versions, directly or through a documented migration. |

The [CLI reference](CLI.md), [JSON contract](cli-contract.md), and
[MCP reference](MCP.md) define the concrete behavior. Fixing behavior that
violates those contracts is a bug fix, not a new compatibility exemption.
Firebase permissions, quotas, retained server history, credentials expiring or
being revoked, and upstream service availability remain external conditions.

## CLI and JSON evolution

Application patch releases preserve supported invocations and behavior. Minor
releases may add compatible commands, flags, capabilities, and MCP tools.
Removing or renaming a supported interface, making a formerly optional input
required, rejecting previously valid input, changing an existing value's
meaning, or changing its documented effects or exit status requires an
application major release. Deprecation alone does not permit removal in 1.x.

The CLI JSON contract is versioned independently:

- **Major:** incompatible inputs, outputs, semantics, or guarantees. This
  includes removed/renamed fields or commands, changed field types or meanings,
  changed exit statuses, and changes to closed enums that older consumers
  cannot handle. An incompatible machine change also requires an application
  major release.
- **Minor:** compatible additions such as new commands, optional inputs, or
  optional output fields with existing meanings preserved. Adding values is
  compatible only at an explicitly open extension point; closed enums are not
  implicitly extensible.
- **Patch:** generated metadata or schema clarifications that preserve accepted
  invocations, emitted data, and semantics. Documentation corrections alone
  do not require a contract bump unless a fingerprinted artifact changes.

A published schema ID always identifies the same schema. New contract versions
use new IDs and directories; previous published schema resources are retained.
Use the `schema` and `contract_version` advertised by the installed binary and
validate with that version's schemas. A new minor version's document need not
validate against an older, closed schema. Consumers supporting a contract
major must tolerate compatible additions and use the advertised schema; pin a
binary release when an exact schema/contract version is required. The CLI has
no output-version negotiation flag and does not promise to emit older envelopes.
JSON whitespace, object-member order, human-readable messages and descriptions
are not machine interfaces. Explicit artifact byte/digest rules still apply.

## MCP boundary

MCP uses its negotiated protocol version over stdio; this is not the CLI
`contract_version`. Completed tools return the shared operation envelope in
`structuredContent` and equivalent JSON in text `content`. The operation IDs
and schema IDs remain the CLI IDs even when a tool has a different public name.
Discovery remains conditional on toolsets, write permission, and stateful or
stateless execution. Hosts must discover the schemas for their actual launch
policy; one configuration's catalog is not a promise about another's.

Schema bundling and portable representations may evolve while preserving their
validation meaning. An incompatible tool rename, input constraint, result
meaning, or authorization change requires an application major release;
incompatible shared envelopes also require a CLI contract major version.
Changes confined to MCP names, adaptation, or lifecycle are not covered by the
CLI schema fingerprint and require their own review and MCP regression tests.

## Persisted-format boundary

The upgrade guarantee covers:

- global `config.toml` and repository `.fbrcm.toml`, including configuration
  keys, project aliases, hooks, and custom key configurations;
- profile layout and identity references, `auth-config.json` and managed
  credential/token files, `projects-config.json` and its user-owned settings,
  and `project-visibility.json`;
- installed theme TOML and hook trust records;
- draft JSON, retained Remote Config version snapshots, and saved publication
  plan documents, even when they live under a cache root.

Auth, project registry, visibility, draft, application-cache, and publication
plan formats currently use integer version `1`. TOML configuration, themes,
hook trust records, and Remote Config snapshot wrappers have no separate
format-version field; that does not exempt their supported data from upgrade
compatibility. A release number alone never increments a file-format version.
Version 1.1.0 keeps the existing encodings, including the
publication-plan artifact schema `urn:fbrcm:schema:cli:1.0.0:publication_plan` and its integer
`format_version: 1`. Capture inputs for `explain` are not a new persisted format;
the operation does not store captures.

An incompatible encoding needs a distinct format version where one exists and
a documented migration before replacing durable data. Versionless formats need
an unambiguous migration boundary rather than silent reinterpretation. Preserve
user data and provide recovery for migration failures. Unsupported future
versions must be rejected without overwriting them. No migrations are needed
for the unchanged encodings in 1.1.0.

Transient application inventories, freshness metadata, discovery results, and
other rebuildable caches may be refreshed or invalidated. They are not a backup
or permanent offline-retention promise. Drafts and retained version snapshots
are not disposable caches for this purpose. Explicit user operations such as
cache clearing, draft discard, profile deletion, and credential replacement
retain their documented destructive behavior. Token refresh and hook trust
revalidation may still be necessary after an upgrade.

The guarantee covers data from released 1.0.0 and subsequent released 1.x
versions when upgrading to later 1.x versions. It does not cover
arbitrary pre-1.0 files, files from newer versions than the reading binary, or
downgrades. Formatting, JSON member order,
comments, and byte-for-byte round trips are not promised unless an artifact
contract explicitly requires them. Directly editing internal cache or token
representations is not a supported integration; use the CLI and returned paths.

## Interfaces outside the stability promise

**Human table layouts are not stable interfaces.** Column order, widths,
wrapping, labels, borders, colors, progress, logs, help wording, and prompts may
change in patch or minor releases. Use `--json` or MCP for automation. Documented
command behavior and raw/artifact output contracts remain covered.

**Default TUI bindings and interaction layouts are not stable interfaces.**
Shortcuts, panels, mouse flows, and action-palette presentation may evolve.
The supported custom key-configuration syntax and stored action identifiers
are durable configuration: existing valid custom bindings must remain usable
or receive a documented migration. Default shortcuts are not an automation API.

**Exported Go packages are not a supported library API.** This repository ships
an application. Exported identifiers in `core`, `ops`, `cli`, `mcp`, `schemas`,
and other packages may change in any release; Go export visibility does not
extend the application's compatibility promise to source or binary consumers.
Integrate through the documented CLI, JSON, or MCP interfaces.

Explicitly documented experimental human-only transports, such as theme
directory stdin, remain outside the promise. Historical audit reports record
their original release state; the current release status is recorded in
`schemas/cli/contract.lock.json`.

## Release lock

The released `1.0.0` fingerprint and marker are archived unchanged in
`schemas/cli/contract-1.0.0.lock.json`. `schemas/cli/contract.lock.json` records
the active contract version, fingerprint, and release marker. A lock with
`released: false` allows in-place regeneration during release preparation.
A released SHA-256 freezes the generated CLI schemas, compact/detailed capabilities,
and audit-evidence golden. The embedded capability catalog is regenerated and
tested against those records. This lock is a generation gate, not a fingerprint
of every runtime implementation, MCP adaptation, or persisted format.

Run `go run ./cmd/schemagen` before releasing. It must preserve the released
marker and digest at the same version, and reject a changed fingerprint before
publishing any generated files. After publication, generated changes need the
appropriate contract-version bump and reviewed schemas, capabilities, evidence, tests, and
documentation. After publication, even a test-evidence-only fingerprint change
requires a version bump; it does not authorize changing runtime behavior. Never clear the released
marker or replace a released fingerprint to bypass this gate.

A new contract version is staged as unreleased until its review is complete;
then set its release marker after generation. The release workflow requires a
released lock before publishing binaries. Persisted-format and MCP compatibility
still require review and regression evidence in addition to this CLI gate.
When advancing beyond 1.1.0, retain its released lock as
`schemas/cli/contract-1.1.0.lock.json` before replacing the current lock with
the next contract version, as was done for 1.0.0.


## Condition blame compatibility

Condition blame is available since CLI JSON contract `1.1.0`. Parameter blame
retains its invocation and response fields. The MCP `versions.blame` tool
accepts the same optional condition selector.
