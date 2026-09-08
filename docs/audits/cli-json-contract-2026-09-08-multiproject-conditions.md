# CLI JSON contract audit — multi-project condition mutations

Repository revision: `64e236f122b6ff5971cf078b4ece1e29af359681` plus this audit-report worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `117`
Input schemas: `117`
Response schemas: `117`
Shared schemas: `12`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0`

The audit scope was frozen before inspection at revision
`64e236f122b6ff5971cf078b4ece1e29af359681`. The landed change makes
`conditions.add` and `conditions.delete` multi-target operations and changes
the shared normalized-input positional compaction used by operation callers.
All twelve acceptance conditions in audit standard 1.0.0 pass, with no
criterion-linked findings. The final unreleased contract lock has SHA-256
`4523aec3259d1e70de0ffa193f5f719ad5a4ebffd67020c98d6c67b6565113e6`.

## Inventory

Exact set comparison: `E = C = I = R = D = B = M`, with 117 members in each
set and zero symmetric-difference members. The executable inventory did not
gain or lose a command; the changed records are `conditions.add` and
`conditions.delete`.

There are 246 unique Draft 2020-12 schema files under
`schemas/cli/1.0.0`: 117 command input schemas, 117 command response schemas,
and 12 shared schemas. All `$id` values are unique, all references compile and
resolve, and each capability references the matching command schemas. The
compact and detailed capability sets, DTO registry, exhaustive behavior
manifest, documentation inventory, and generated evidence matrix all contain
the same 117 command IDs and paths. Contract version `1.0.0` and release state
`false` agree across runtime, lock, schemas, capabilities, and documentation.

## Command audit records

The 115 unchanged command records from the
[preceding September 8 audit](cli-json-contract-2026-09-08.md) are incorporated
by reference. They were not accepted from stale output: their discovery,
invocation, schema, response, behavior, documentation, generated-evidence,
root-test, and E2E gates were rebuilt and rerun at the frozen revision. The
shared positional-input change was also checked across the complete argument
inventory. `conditions.add` is the only optional-before-required argument
sequence; the remaining optional argument forms retain their former binding.
The two records below supersede the corresponding records incorporated by the
preceding report. Together these records cover every member of `E` exactly
once.

### `conditions.add` — `fbrcm conditions add [project] <name>`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.add`; executable leaf; one-argument bulk form and two-argument positional-project compatibility form. |
| Arguments | `name` is required, trimmed, nonblank, and at most 100 Unicode code points. Optional `project` precedes it in the normalized model; arity compacts an omitted project to one argv argument without allowing an optional positional gap. |
| Options | Required nonblank `--expression`; normalized color enum; integer priority from zero through the portable integer ceiling with a per-target runtime upper bound; repeatable `--project/-p`; dry-run, draft, change note, plan output, confirmation bypass, and applicable global options. Positional project and project filters are schema-enforced alternatives; stateless draft and all plan exclusions are enforced. |
| Selection | Positional project uses exact-then-filtered stateful target resolution and literal stateless target resolution. Repeated project filters are ORed, target-aware, deduplicated, and sorted. Omission selects all configured enabled templates. Stateless exact filters bypass discovery; other filters use live discovery. Project not-found and ambiguity failures are typed. |
| Stdin | N/A; normalized stdin is `null`. |
| Success | Shared remote-mutation DTO with one ordered item per resolved target, including target breadth, one matched item per prepared add, status, validation provenance, versions, dry-run/draft markers, and no-op provenance. Publication-plan output is the registered artifact variant. |
| Failure | Typed argument, condition validation, project/profile/configuration, auth/network/Firebase, draft conflict, plan/file, timeout/cancel, interaction, publication, batch, and internal failures. Per-target preparation, validation, conflict, publication, cache, and hook outcomes retain their reached stage. |
| Warnings | Multi-target live publication emits structured `publication.non_atomic`; accepted publications with cache or hook aftermath emit their structured post-publication warnings and safe remediation. |
| Interaction | A changed target without `--yes` returns typed interaction information in JSON rather than prompting. Plan output suppresses confirmation; configured OAuth authorization remains an explicitly declared stateful branch. |
| Effects | Conditional Firebase reads, validation, writes, local cache/draft/state/file writes, auth access, and trusted hooks match the shared mutation predicates. Stateless execution suppresses stateful persistence; dry-run, draft, plan, no-change, confirmation, and hook conditions are published. Non-destructive; retry safety is conditional because targets publish independently. |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:input`, matching response schema, and shared envelope/error/capability/semantic/publication-plan schemas. Dynamic condition priority is annotated as per-target Firebase-state validation. |
| Documentation | `docs/CLI.md`, Condition mutations; `docs/cli-contract.md`, Multi-project operations and direct Remote Config mutation results. |
| Evidence | Input-schema scalar/bulk/stateless/exclusion cases; positional compaction tests; stateless HTTP tests for positional, exact-filter, and repeated-filter forms; shared batch/no-op/warning/plan/interaction suites; response-invariant, generated matrix, root, and E2E tests. |
| Verdict | **PASS.** |

### `conditions.delete` — `fbrcm conditions delete [project] [condition]`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `conditions.delete`; executable leaf; zero-, one-, and two-argument forms. One argument is always the project; the compatibility condition argument requires it. |
| Arguments | Optional project uses the shared target positional grammar. Optional condition is nonblank and exact case-sensitive; the schema rejects a condition without a project, matching argv representability. |
| Options | Repeatable project and name filters, normalized substring search, compiled condition-context expression, and the shared mutation/plan/global options. Positional project conflicts with project filters; positional condition conflicts with name filters; stateless draft and plan exclusions are enforced. |
| Selection | Repeated project filters and repeated name filters are ORed within their sources; project, name, search, expression, and positional-condition sources compose with AND. Omitted project sources select all configured enabled templates; omitted condition sources match every condition. Positional condition is exact and case-sensitive with typed `condition.not_found`; filter misses yield per-target `no_match`. Targets and conditions preserve deterministic order. |
| Stdin | N/A; normalized stdin is `null`. |
| Success | Shared ordered per-target mutation DTO. Each result preserves default scope, resolved target count, matched condition count, changed count, status, validation source, versions, and `no_match` versus `already_applied`. Multiple matching conditions are removed in one candidate. Publication-plan output is the registered artifact variant. |
| Failure | Closed typed set includes expression and condition selection failures plus the shared argument, project, auth, network, Firebase, draft, plan/file, interaction, target-stage, batch, timeout/cancel, and internal failures. Mixed and all-failed batches retain typed target failures and retry selectors. |
| Warnings | Structured non-atomic publication warning for multi-target live runs and the reachable cache/hook post-publication warnings, with bounded messages and safe remediation. |
| Interaction | Destructive changed targets require confirmation unless bypassed; JSON returns typed interaction data and never prompts. Plan output suppresses confirmation; stateful OAuth authorization is separately declared. |
| Effects | Destructive exactly when a non-plan mutation removes conditions. Firebase/local/auth/hook effects and their stateless, draft, dry-run, plan, no-match, confirmation, and publication predicates match runtime. Retry safety is conditional across independently published targets. Empty and description-only groups are preserved. |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:input`, matching response schema, and shared envelope/error/capability/semantic/publication-plan schemas. Expression compilation and resource lookup boundaries are typed. |
| Documentation | `docs/CLI.md`, Condition mutations; `docs/cli-contract.md`, Multi-project operations and direct Remote Config mutation results. All published examples validate against the normalized contract and reach the documented argv form. |
| Evidence | Zero/one/two positional schema cases; project/condition-filter exclusions; exact matching and expression composition tests; stateless HTTP exact and expression deletion; shared multi-target batch continuation/ordering/aggregation, no-op, warning, interaction, and plan suites; response-invariant, generated matrix, root, and E2E tests. |
| Verdict | **PASS.** |

## Criterion results

| Criteria | Result |
| --- | --- |
| `INV-01`–`INV-04` | PASS: exact 117-member equality, unique IDs/paths/references, unreleased 1.0.0 agreement, and full JSON envelope coverage. |
| `ARG-01`–`ARG-07` | PASS: the new arities, compact optional-before-required binding, types, defaults, exclusions, normalization, dynamic priority annotation, option effectiveness, and stateless restrictions match runtime. |
| `SEL-01`–`SEL-05` | PASS: scalar and repeated target selection, exact condition compatibility selection, list-style condition filters, cross-source composition, omitted defaults, deterministic ordering, and typed zero/ambiguous/not-found behavior agree. |
| `STDIN-01`–`STDIN-04` | PASS: neither changed command publishes stdin; all previously stdin-capable records remain unchanged and passing. |
| `OUT-01`–`OUT-06` | PASS: actual scalar and multi-target JSON validates; target breadth, matched counts, no-op/status/provenance fields, plan artifacts, and envelope invariants agree. |
| `ERR-01`–`ERR-06` | PASS: expression and condition errors, per-target failures, batch aggregation, exit statuses, warnings, retry selectors/remediation, redaction, and bounds remain typed and schema-valid. |
| `BEH-01`–`BEH-06` | PASS: every remote/local/auth/hook/destructive/idempotency branch is declared, with dry-run/draft/plan/stateless/no-op suppression and valid predicates. |
| `INT-01`–`INT-04` | PASS: JSON mode never prompts or launches UI; confirmation and OAuth branches remain typed, schema-valid, and correctly suppressed or bypassed. |
| `DOC-01`–`DOC-04` | PASS: command tree, details, examples, multi-project contract rules, generated artifacts, capabilities, and runtime agree. |
| `GEN-01`–`GEN-05` | PASS: generation is deterministic, 246 schemas compile and resolve, exhaustive inventories and evidence pass, and all final repository gates are clean. |

## Test-class evidence

The generated evidence catalog contains 117 command records with all 15
classes populated or explicitly marked N/A, and 50 checked evidence symbols.
The changed commands retain command-local invocation, boundary, selection,
success, failure, and stateless HTTP coverage and connect to the already
audited shared mutation machinery for no-op, interaction, warning, batch,
effects, plan artifact, and response validation. Fresh full-suite execution
revalidated the root/help/completion/startup, schema, DTO, typed-error,
redaction, and E2E classes for all commands.

## Determinism and repository gates

| Check | Result |
| --- | --- |
| `go run ./cmd/schemagen` twice | PASS; the second run produced no diff. Aggregate generated-artifact SHA-256: `185d9371ea5612e15dcc39dca66bbe0bdbe2d0ef5d6f55920a3ed44cfeb7473f`. |
| Schema compilation and references | PASS; 246 unique Draft 2020-12 schemas, no duplicate or dangling IDs or references. |
| Generated audit matrix | PASS; 117 commands by 15 classes, 50 checked symbols, no blank or invalid cells. |
| `go test -count=1 ./...` (root) | PASS. |
| `go test -count=1 ./...` (`e2e`, isolated final run) | PASS, including the Hoverfly harness. |
| `go vet ./...` (root and `e2e`) | PASS. |
| `golangci-lint run` (root and `e2e`) | PASS; zero issues. |
| Final diff | PASS; only this audit report was added, and no pre-existing user changes existed. |

An earlier E2E invocation run concurrently with the complete root suite passed
the CLI scenario package but hit the standalone Hoverfly harness client
timeout. The required isolated final E2E gate passed all packages; the
transient concurrent harness timeout did not expose a CLI contract finding.

## Findings and verdict

No criterion-linked findings were discovered. The landed implementation,
documentation, capabilities, schemas, goldens, runtime results, and generated
evidence agree at the frozen revision. All twelve section-7 conditions pass.
Under audit standard 1.0.0, the required verdict is
**AUTHORITATIVE: PASS**.
