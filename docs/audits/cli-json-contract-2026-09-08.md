# CLI JSON contract audit — applications, blame, and shared exactness

Repository revision: `dff1ce72e448e5a2ec3a25f541b2b3fb8041480f` plus the audited repair worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `117`
Input schemas: `117`
Response schemas: `117`
Shared schemas: `12`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0` unresolved (`4` found and closed)

The audit scope was frozen at revision
`dff1ce72e448e5a2ec3a25f541b2b3fb8041480f`. The initial inspection produced
`NOT AUTHORITATIVE: FAIL`. Four criterion-linked defects were repaired without
changing audit standard 1.0.0. The final unreleased contract lock has SHA-256
`e66be71eec6c7614bdb7f61b1f3028c1689e07ab6e5496e8e42c280cd335f660`.

## Inventory

Exact set comparison: `E = C = I = R = D = B = M`, with 117 members in every
set and zero symmetric-difference members. The four commands added after the
September 3 audit are `apps.config`, `apps.list`, `apps.show`, and
`versions.blame`.

There are 246 unique Draft 2020-12 files under `schemas/cli/1.0.0`: 117 input
schemas, 117 response schemas, and 12 shared schemas (`capability`, `envelope`,
`error`, `publication_plan`, `semantic`, six stdin schemas, and `theme` stdin).
All `$id` values are unique, all references compile and resolve, and every
capability points to the matching command input and response schema. Compact
and detailed capabilities, the embedded registry, response DTO registration,
the behavior manifest, documentation inventory, and generated evidence matrix
all derive from and exactly match the executable tree.

## Command audit records

The 113 records in the
[September 3 audit](cli-json-contract-2026-09-03.md), including the earlier
records incorporated there, are incorporated here by reference. This is record
compression, not skipped testing: all 113 were rebuilt and rerun through the
current discovery, invocation, schema, runtime, behavior, documentation, root,
and E2E gates. Shared changes to confirmation effectiveness, stateless effect
predicates, collection invariants, and cache DTOs were reapplied to every
affected record.

The four new records follow. Together with the incorporated records, they form
exactly one section-4 record for every member of `E`.

### `apps.list` — `fbrcm apps list [project]`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `apps.list`; optional project argument; executable leaf. |
| Arguments | Zero or one project. Stateful resolution uses the shared exact-then-filtered project selector; stateless positional selection is a literal physical project ID. |
| Options | Repeatable `--project/-p` and `--filter/-f`; `--platform` enum; `--show-deleted`; mutually exclusive `--cached`/`--update`; global profile, local-config, timeout, stateless, and JSON options. Cache flags and profile publish stateless applicability. |
| Selection | Repeated project selectors are ORed and deduplicated. App filters use exact, starts-with, contains, or fuzzy modes over documented app fields. Platform and deleted-state filtering are local. |
| Stdin | N/A; normalized stdin is null. |
| Success | Closed counted DTO. Every item carries project identity, app summary, source, and a required non-null `cached_at` for cache-backed sources; count equals item length. Empty is reachable. |
| Failure | Closed typed set for arguments, project selection, profile/configuration, auth, Firebase/network, filesystem/cache, interaction, timeout/cancel, and internal contract enforcement. |
| Warnings | `cache.stale` and `cache.write_failed`; cache error details are redacted and bounded. |
| Interaction | Only configured-auth authorization is reachable; JSON returns typed interaction metadata. |
| Effects | Conditional Firebase read and cache write; stateful-only registry/profile/auth persistence; no mutation, non-destructive, idempotent. Stateless reads live and persists nothing. |
| Schemas/docs/evidence | Command input/response schemas; `docs/CLI.md` Apps section; shared contract Apps section; unit, exhaustive contract, stateless multi-project, generated matrix, and E2E coverage. |
| Verdict | **PASS.** |

### `apps.show` — `fbrcm apps show <app>`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `apps.show`; exactly one app selector; executable leaf. |
| Arguments | App selector is nonblank. A complete Firebase App ID may infer its project; all other forms require `--project`. |
| Options | `--project/-p`; mutually exclusive `--cached`/`--update`; global profile, local-config, timeout, stateless, and JSON options, with stateful applicability published. |
| Selection | Project resolution is exact-then-filtered in stateful mode and literal in stateless mode. App resolution is exact, case-sensitive across App ID, resource name, namespace, and display name, with typed not-found/ambiguity candidates. |
| Stdin | N/A. |
| Success | Closed platform-specific details DTO. Android, iOS, and Web shapes reject fields impossible for that platform. Cache-backed sources require `cached_at`. |
| Failure | Closed typed set including `app.not_found` and `app.ambiguous` plus the shared read/auth/network/configuration failures. |
| Warnings | `cache.stale` and safe bounded `cache.write_failed`. |
| Interaction | Configured-auth authorization only; represented as typed JSON interaction. |
| Effects | Conditional Firebase/cache read, stateful-only cache/registry/profile/auth writes; non-destructive and idempotent. |
| Schemas/docs/evidence | Command schemas; Apps documentation; runtime app-selection/cache tests, response-invariant tests, matrix and E2E evidence. |
| Verdict | **PASS.** |

### `apps.config` — `fbrcm apps config <app>`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `apps.config`; exactly one app selector; executable leaf. |
| Arguments | Same project inference and exact app-selection boundary as `apps.show`. |
| Options | `--project/-p`, `--to`, `--yes/-y`, mutually exclusive `--cached`/`--update`, and global options. `--yes` is effective exactly when a destination conflict requires confirmation. |
| Selection | Same project/app rules as `apps.show`; platform selects the Firebase SDK-config representation, not another resource. |
| Stdin | N/A. |
| Success | Closed DTO containing app metadata, suggested filename, provenance, and an artifact. Artifact target is a nonempty Firebase resource name; platform/media-type, inline/destination, encoding, size, digest, and overwrite rules are constrained. |
| Failure | Closed app/project/auth/network/configuration/file/interaction/internal problem set. Existing destination without bypass returns `interaction.required`. |
| Warnings | `cache.stale` and safe bounded `cache.write_failed`; no implicit stale SDK-config fallback. |
| Interaction | Destination overwrite and configured-auth authorization are explicit conditional branches. |
| Effects | Conditional remote/cache reads and cache writes; explicit destination write in either state mode; stateful-only registry/profile/auth persistence. Conditionally destructive only for an authorized overwrite; idempotent. |
| Schemas/docs/evidence | Command schemas; Apps/artifact documentation; exact-byte/private-file, warning-safety, response-invariant, interaction, matrix, and E2E tests. |
| Verdict | **PASS.** |

### `versions.blame` — `fbrcm versions blame <project> <parameter>`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `versions.blame`; exactly two arguments; executable leaf. |
| Arguments | Project uses the shared target selector. Parameter is nonblank, exact, case-sensitive, and at most 256 Unicode code points. |
| Options | `--at` uses the current/relative/canonical-version grammar; positive `--limit` defaults to 1; `--all` and an explicit limit are mutually exclusive; global state/timeout/JSON options apply. |
| Selection | Project ambiguity/not-found and version not-found are typed. Parameter lookup spans retained history and returns `parameter.not_found` only when never present. |
| Stdin | N/A. |
| Success | Closed history DTO. At least one version is scanned; `changes` is always an array and excludes `unchanged`; `history_exhausted` is equivalent to a non-null boundary; present/absent boundary group shapes are correlated. |
| Failure | Closed typed set for argument, project/version/parameter selection, profile/configuration, auth, Firebase/network, filesystem, interaction, timeout/cancel, and internal enforcement. |
| Warnings | N/A; warning array is empty. |
| Interaction | Configured-auth authorization only; stateless token mode is non-interactive. |
| Effects | Firebase history and template reads are required. Stateful execution may cache immutable snapshots and persist registry/profile/auth state; stateless execution performs no local writes. Non-destructive and idempotent. |
| Schemas/docs/evidence | Command schemas; Versions documentation; core pagination/boundary tests, CLI JSON success and response-rejection tests, generated matrix, and stateless E2E scenario. |
| Verdict | **PASS.** |

## Criterion results

| Criteria | Result |
| --- | --- |
| `INV-01`–`INV-04` | PASS: exact 117-member equality, unique paths/IDs/references, unreleased 1.0.0 agreement, complete root/help/completion/startup inventory. |
| `ARG-01`–`ARG-07` | PASS: arity, types, defaults, repetition, enums/bounds/grammars, dependencies/exclusions, normalization, omitted/empty distinctions, and effectiveness agree. Every `--yes` flag now publishes its runtime confirmation condition. |
| `SEL-01`–`SEL-05` | PASS: all literal, mode-prefixed, repeated, default, app/project/version/parameter, and stateless selector rules and typed ambiguity/not-found branches agree. |
| `STDIN-01`–`STDIN-04` | PASS: every stdin-capable command retains exact modes and concrete schemas; the four new commands explicitly have none. |
| `OUT-01`–`OUT-06` | PASS: clean envelopes, typed DTO registration, empty/null/count/status/platform/history/cache/artifact invariants, and reachable variants agree. Duplicate collection invariants and stale cache-clear field constraints were removed. |
| `ERR-01`–`ERR-06` | PASS: closed typed problems/warnings, details, retryability, remediation, aggregation, statuses, redaction, and bounds agree. App cache-write details are now safe text. |
| `BEH-01`–`BEH-06` | PASS: effects, network, destruction, idempotency, suppression, and predicates match runtime. No clause has duplicate or conflicting equality predicates; stateless persistence clauses are stateful-only. |
| `INT-01`–`INT-04` | PASS: JSON never prompts or launches an external UI; confirmation, OAuth, destination, and selection interactions are typed and their bypass/suppression branches are covered. |
| `DOC-01`–`DOC-04` | PASS: CLI reference, contract guide, examples, schemas, capabilities, runtime, and new cache/app/blame exactness rules agree. |
| `GEN-01`–`GEN-05` | PASS: two final generations were byte-identical; 246 schemas compile; exhaustive inventories and the 117-by-15 matrix pass; root/E2E test, vet, and lint gates pass. |

## Test-class evidence

`cli/app/testdata/contract_v1_audit_evidence.golden.json` has 117 command
records, exactly 15 class cells per command, and 50 checked evidence symbols.
Every applicable cell cites a valid symbol and relevant JSON E2E scenario;
every N/A cell has a nonblank reason.

| Class | Passing evidence |
| --- | --- |
| Discovery | Exact executable/capability/schema/DTO/behavior/docs equality, reference compilation, goldens, embedded capabilities, and E2E discovery. |
| Invocation/boundary/effectiveness | Exhaustive arity and option failures, normalized schemas, enums/bounds/grammars, dependency/exclusion tests, ignored/conditional options, and confirmation applicability. |
| Selection/stdin | Shared and command-local selector semantics plus all concrete stdin valid, malformed, boundary, and restriction suites. |
| Success/no-op/failure/warning | Runtime envelopes, impossible-state rejection, status/count correlations, closed codes, classifications, warning detail schemas, safety, and redaction. |
| Interaction/batch/effects | Confirmation/OAuth/destination branches, mixed batch aggregation, instrumented remote/local/cache/draft/hook/profile effects, stateless suppression, and predicate-integrity tests. |
| Artifact/determinism | Inline/destination representations, exact bytes/size/digest/overwrite behavior, generator unit determinism, and the two repository generation passes. |

## Closed findings

### Finding 1 — contradictory stateless effects and incomplete option applicability

Criteria: `ARG-06`, `BEH-01`, `BEH-05`, `BEH-06`, `GEN-04`.

The shared stateless decorator retained a cache-write clause requiring
`stateless: true`, then appended `stateless: false`; already-stateful clauses
also received duplicate false predicates. Separately, confirmation-bypass flags
were advertised as unconditionally effective. The decorator now removes
stateless-only persistence clauses, adds one stateful predicate to the remaining
clauses, safely restricts unconditional persistence, and avoids duplicates.
Every `--yes` flag publishes `runtime_state.confirmation required`. Generated
capabilities are regression-tested for duplicate and conflicting predicates.
**Closed.**

### Finding 2 — unreachable `versions.blame` response states

Criteria: `OUT-02`, `OUT-03`, `OUT-04`, `OUT-05`, `DOC-01`.

The response schema accepted null changes, zero scanned versions, an
`unchanged` attributed entry, contradictory exhaustion/boundary combinations,
and invalid boundary group states. Runtime cannot emit them. The DTO schema now
encodes every correlation and rejection fixtures cover each former false
positive. Documentation records the same invariants. **Closed.**

### Finding 3 — incomplete application provenance, artifact, and warning safety

Criteria: `OUT-03`, `OUT-06`, `ERR-04`, `ERR-05`, `DOC-01`.

App schemas allowed cache-backed results without timestamps and config
artifacts with null or empty targets. Cache-write warning details copied raw
error text without the contract's secret redaction and length bound. Cache
metadata now rejects missing timestamps; response schemas require them for
cache sources; config artifact targets are nonempty; warning details use safe
error text and publish the matching bound. Runtime and schema rejection tests
cover the repaired behavior. **Closed.**

### Finding 4 — stale cache DTO constraints and duplicate collection rules

Criteria: `OUT-03`, `OUT-04`, `GEN-02`, `GEN-03`, `DOC-01`.

`cache.clear` status constraints still targeted the removed
`snapshots_deleted` property, so impossible counts validated. Mixed cache-list
items also admitted kind/resource/version combinations runtime never emits, and
all generated collection DTOs repeated the same count invariant. The generator
now constrains `entries_deleted`, correlates all clear counts, closes cache item
shapes, validates listed app-cache metadata, and emits one count invariant.
The schema-show E2E golden was refreshed to the corrected representation.
**Closed.**

## Determinism and repository gates

| Check | Result |
| --- | --- |
| `go run ./cmd/schemagen` twice | PASS; byte-identical generated trees; aggregate generated-artifact SHA-256 `fd56dd58935e884703e1f1bd5febb94dd9c5ac4dbd55d0a8c8c2feb6d8efba06`. |
| Schema compilation/reference tests | PASS; 246 unique Draft 2020-12 files, no duplicate/dangling IDs or references. |
| Generated audit matrix | PASS; 117 commands × 15 classes, 50 checked symbols, no blank or invalid cells. |
| `go test -count=1 ./...` (root) | PASS. |
| `go test -count=1 ./...` (`e2e`) | PASS. |
| `go vet ./...` (root and `e2e`) | PASS. |
| `golangci-lint run` (root and `e2e`) | PASS; zero issues. |
| Final diff | PASS; only contract/runtime repairs, tests, documentation, generated artifacts/goldens/lock, the schema-show E2E snapshot, and this report changed. No pre-existing user changes existed. |

An earlier E2E run under concurrent repository test and lint load passed the
CLI scenarios but timed out in the standalone Hoverfly harness integration
test. The isolated final E2E gate above passed all three packages.

All twelve section-7 acceptance conditions pass, every applicable criterion
and test class passes, and there are zero unresolved findings. Under audit
standard 1.0.0, the required verdict is **AUTHORITATIVE: PASS**.
