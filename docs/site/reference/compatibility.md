# Compatibility from 1.0.0

fbrcm 1.0.0 establishes a stable application interface and a released CLI JSON
contract `1.0.0`. The release, JSON contract, negotiated MCP protocol, and file
formats have separate versions. Existing integer file-format versions stay at
`1`; the pre-1.0 in-place breaking-change exception has ended.

Version 1.1.0 is under development and has not been published. Its CLI JSON
contract `1.1.0` has `released: false` and may be regenerated in place before
release. It includes condition blame in the CLI and MCP. Published `1.0.0` schemas and the original
released lock remain unchanged. Persisted encodings stay the same, so this
release needs no migration.

| Surface | Stability promise |
| --- | --- |
| Documented CLI syntax and behavior | Compatible throughout fbrcm 1.x: commands, arguments, flags, defaults, selection, effects, confirmations, environment variables, and exit meanings |
| JSON envelopes and schemas | Versioned compatibility for DTOs, codes, details, capability metadata, semantic rules, and artifacts |
| MCP tools | Stable public names, operation mappings, launch policy, discovered contracts, results, and authorization behavior for the same server configuration |
| Durable data | Configuration, identities, preferences, custom key configurations, themes, hook trust, drafts, retained snapshots, and plans from every released 1.x version remain readable on later 1.x upgrades, directly or through a documented migration |

Breaking a supported interface requires an application major release.
Incompatible JSON changes also require a contract major version. Compatible
additions use minor versions; fingerprinted clarifications with no validation
or semantic change may use a contract patch version. Closed enums are not
implicitly extensible.

Use the installed binary's advertised `schema` and `contract_version` with its
schemas. Published schema IDs are immutable. There is no CLI output-version
negotiation flag; pin a binary release if you require an exact contract version.
MCP protocol negotiation is separate from the shared JSON envelope version.

## What is not a stable integration interface

- **Human tables:** columns, widths, labels, borders, colors, logs, progress,
  help wording, and prompts may change. Use JSON or MCP for automation.
- **Default TUI bindings and layouts:** shortcuts, panels, and mouse flows may
  evolve. Supported custom key-configuration syntax and saved action
  identifiers remain durable configuration and require compatibility or a
  documented migration.
- **Exported Go packages:** fbrcm is an application, not a supported Go library.
  Exported symbols may change in any release.

Transient rebuildable caches may be invalidated. Drafts and retained version
snapshots are protected durable data even under a cache root. The promise covers
upgrades from released 1.0.0 and subsequent released 1.x data to later 1.x binaries,
not arbitrary pre-1.0 data, files newer than the reading binary, or downgrades.
Explicit user deletion and credential replacement keep their documented effects.

The released CLI lock rejects a changed generated fingerprint at the same
contract version, and the release workflow refuses an unreleased lock. That
fingerprint does not replace MCP or persisted-format compatibility review.
The publication-plan artifact retains its original `1.0.0` schema ID and
`format_version: 1`; an application or JSON-contract version change alone never
changes a durable file-format version. Versionless formats also require an
explicit migration boundary for incompatible changes.
See the full [compatibility policy](https://github.com/yumauri/fbrcm/blob/main/docs/compatibility.md)
for versioning, migration, and release requirements.


Condition blame through `versions blame --condition <name>` is part of
unreleased contract `1.1.0`. Parameter invocations and responses remain unchanged.
The existing MCP `versions.blame` tool accepts the same optional selector.
