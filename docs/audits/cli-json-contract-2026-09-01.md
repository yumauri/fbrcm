# CLI JSON contract audit — publication plans

Repository revision: `340d4a7de5e24fead1f23e75fc9a3386aa3c9acf` plus the audited repair worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `112`
Input schemas: `112`
Response schemas: `112`
Shared schemas: `12`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0` unresolved (`6` found and closed)

The audit scope was frozen at branch `plans`, revision
`340d4a7de5e24fead1f23e75fc9a3386aa3c9acf`, before repair. The generated
contract lock is unreleased and has SHA-256
`c22e7cbd8a7d4712931927b45e7fc17bc61e2083c5a4314eb9dd9f4aba688be2`.
The first audit pass produced `NOT AUTHORITATIVE: FAIL`; the criterion-linked
findings below were repaired without changing audit standard 1.0.0, and this
report records the complete repair audit and final passing verdict.

## Inventory

Exact set comparison: `E = C = I = R = D = B = M`, with 112 members in every
set and zero symmetric-difference members. The three members added since the
prior full audit are `apply`, `plan.show`, and `plan.validate`. The executable
tree, DTO registry, exhaustive behavior manifest, compact and detailed
capabilities, documentation inventory, input schemas, and response schemas are
compared by repository contract tests and generated goldens.

There are 236 unique Draft 2020-12 files under `schemas/cli/1.0.0`: 112 input
schemas, 112 response schemas, and these 12 shared schemas: `capability`,
`envelope`, `error`, `publication_plan`, `semantic`, `stdin:credentials`,
`stdin:oauth_credentials`, `stdin:publication_plan`, `stdin:remote_config`,
`stdin:remote_config_import`, `stdin:service_account_credentials`, and
`stdin:theme`. All `$id` values and references compile and resolve. Capability
paths reference their exact input and response schema IDs; no duplicate,
dangling, or mismatched member was found. The E2E capability and schema-list
snapshots independently report 112 and 236.

## Command audit records

The 92 unaffected records from the
[2026-08-30 full repair audit](cli-json-contract-2026-08-30.md), including the
records it incorporates from the
[2026-08-28 full audit](cli-json-contract-2026-08-28.md), are incorporated by
reference. Every cell was revalidated by the current 112-command generated
evidence matrix and the complete schema, root-module, and E2E gates. The 17
affected existing records and three new records follow. Together they form
exactly one complete record for every member of `E`.

### Affected existing records

All 17 commands retain their previously audited argv path, argument arity,
selector semantics, non-plan stdin behavior, ordinary mutation DTOs, typed
failures and warnings, and command-specific documentation. Each now also has
the following audited plan branch: `--plan-out` is an effective, Unicode-trimmed,
nonblank private path other than `-`; it excludes `--dry-run`, `--draft` where
present, and `--yes`; it creates rather than overwrites a plan; it performs the
command's normal resolution, preparation, Firebase validation, and applicable
trusted pre-publish hooks without publishing or changing drafts; confirmation
is suppressed; `plan.exists` is the only plan-production problem; retry safety
is conditional on whether a trusted hook executed; and success may be the
registered `PlanCreatedResult` with exact artifact metadata and correlated
target counts. Input/response IDs remain command-specific and reference the
shared semantic and envelope schemas.

| Command | Complete audited amendment and verdict |
| --- | --- |
| `add` | Required parameter argument, value-source constraints, target selectors, stdin transform branch, direct/draft/dry-run publication variants, and warnings remain as previously recorded. Plan production preserves the resolved selection and candidate. `docs/CLI.md § fbrcm add` plus Publication plans; runtime/schema/artifact/effect tests pass. **PASS.** |
| `delete` | Parameter/filter/search/expression selection, destructive removal semantics, stdin transform, no-match, draft/dry-run/live results, and typed failures remain complete. Plan creation records exact removal candidates without deleting or publishing. Shared plan docs and all applicable evidence classes pass. **PASS.** |
| `duplicate` | Exact source-name resolution, destination/value behavior, multi-target selection, no-match/ambiguous failures, and publication variants remain complete. Plan mode records the same candidate and selection without mutation. **PASS.** |
| `update` | Positional parameter plus option selectors, value/condition/group transformations, composition and no-op provenance, stdin mode, publication variants, and failures remain complete. Plan mode records the exact selected update. **PASS.** |
| `conditions.add` | Required name/expression, priority bounds, exact project resolution, validation/publication variants, and failures remain complete. Plan mode preserves the candidate condition ordering and validation provenance. **PASS.** |
| `conditions.delete` | Exact condition selection, destructive behavior, no-match handling, and mutation results remain complete. Plan mode records but does not perform the deletion. **PASS.** |
| `conditions.edit` | Exact condition selection, editor/input behavior outside JSON, JSON non-interaction, validation, and mutation results remain complete. Plan mode records the edited candidate. **PASS.** |
| `conditions.move` | Exact condition selection and runtime-dependent priority bounds remain complete. Plan mode records the reordered candidate and its validated snapshot. **PASS.** |
| `conditions.rename` | Exact old-name selection, new-name validation/conflicts, reference rewrite, and mutation results remain complete. Plan mode records the rename without publication. **PASS.** |
| `groups.add` | Group-name/description input, project selection, conflict and mutation variants remain complete. Plan mode records exact group creation. **PASS.** |
| `groups.delete` | Exact group selection, destructive group-level removal, empty/description-only preservation rules elsewhere, and mutation variants remain complete. Plan mode records the group removal without applying it. **PASS.** |
| `groups.edit` | Exact group selection, description transformation, no-op and publication variants remain complete. Plan mode records the edited candidate. **PASS.** |
| `groups.rename` | Exact old-name selection, destination conflict, contained-parameter preservation, and publication variants remain complete. Plan mode records the rename candidate. **PASS.** |
| `draft.publish` | Batch draft selection, all/explicit composition, stale/rebase behavior, dry-run/live/already-applied/post-publication statuses, warnings, cleanup, partial failure, and interaction remain complete. Plan mode converts the selected drafts into immutable targets without deleting drafts or publishing; artifact and hook retry boundaries pass. **PASS.** |
| `project.import` | File-before-stdin selection, local import validation, strategy/merge interaction, selection, draft/dry-run/live variants, warnings, and failures remain complete. Plan mode retains strategy output and exact candidate, suppresses confirmation, and writes only the private artifact. **PASS.** |
| `projects.promote` | Source/target selection, diff filters, explicit promotion selection, validation, dry-run/live/no-op results, warnings, and failures remain complete. Plan mode records the selected promotion candidate without target publication. **PASS.** |
| `versions.restore` | Target and version-selector grammar, history reads, exact restore candidate, dry-run/live/no-op/post-publication statuses, and failures remain complete. Plan mode records the ETag-protected restore candidate; native `versions.rollback` remains excluded. **PASS.** |

### `plan.show`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `plan.show`; argv `fbrcm plan show <plan>`. |
| Arguments | Exactly one required, non-repeated `plan` path; raw `-` selects stdin, every other nonblank value selects a file. |
| Options | `--json` plus inherited global options; no command-local mutation option. Machine option types/defaults and profileless execution agree with Cobra and capability metadata. |
| Selection | N/A. The path selects one document; target order is validated as canonical rather than queried. |
| Stdin | `supports.stdin=true`, mode `json_document`, schema `stdin:publication_plan`; `path_or_stdin_document` consumes one document only for `-` and otherwise leaves stdin unconsumed. |
| Success | Registered non-secret `summary`: plan identity, producer/operation/execution metadata, nonzero target count, publish count, and ordered target summaries. Counts correlate with the target array and publish actions. Human mode additionally renders each complete base-to-candidate diff. |
| Failure | Closed plan parsing/integrity/version problems plus typed file, argument, startup, cancellation, timeout, and internal failures with documented categories and exit statuses. |
| Warnings | N/A; successful inspection emits no warning branch. |
| Interaction | N/A; no prompt, editor, browser, picker, or confirmation in JSON mode. |
| Effects | Profileless and offline for plan content; no Firebase, cache, draft, hook, or artifact-write effect. JSON envelope construction remains covered by the shared profile-bootstrap predicate where applicable. Idempotency `yes`. |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:plan.show:input`; matching response ID; shared `publication_plan`, `stdin:publication_plan`, semantic, envelope, and error schemas. |
| Documentation | `docs/CLI.md § Publication plans` and `§ fbrcm plan show`; stdin, validation, and response rules in `docs/cli-contract.md`. |
| Evidence | Generated 15-class record; `plan_show_json`; `TestPlanShowAndValidateJSON`; plan integrity/input-selection schema tests; malformed/tampered plan tests; response invariant, arity, unknown-option, envelope, and determinism suites. Selection, warning, batch, artifact-write, and interaction cells are explicitly N/A. |
| Verdict | **PASS.** |

### `plan.validate`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `plan.validate`; argv `fbrcm plan validate <plan>`. |
| Arguments | Exactly one required, non-repeated path with the same literal `-` stdin rule as `plan.show`. |
| Options | `--json` plus inherited global options; types, defaults, and effectiveness agree with Cobra. |
| Selection | N/A; it validates one supplied document. |
| Stdin | Concrete `stdin:publication_plan` JSON document selected only by path `-`; non-stdin path leaves stdin untouched. |
| Success | Registered `validationResult` with verified `plan_id`, `valid: true`, and nonzero `target_count`. No invalid document is represented as a successful `valid: false` variant. |
| Failure | `plan.invalid`, `plan.integrity_failed`, and `plan.unsupported_version`, plus applicable typed file/argument/startup/cancellation/timeout/internal failures. Integrity checks cover nonempty/canonical/unique targets, digests, `none` equality, action/provenance, and content-derived ID. |
| Warnings | N/A. |
| Interaction | N/A; JSON validation is non-interactive. |
| Effects | Offline/profileless document validation; no network or persisted mutation; idempotency `yes`. Embedded Remote Config uses the local `firebase.PrepareRemoteConfigUpdate` boundary, not remote validation. |
| Schemas | Command input/response IDs plus shared `publication_plan`, `stdin:publication_plan`, semantic, envelope, and error IDs. |
| Documentation | `docs/CLI.md § fbrcm plan validate`; shared stdin/integrity rules in `docs/cli-contract.md`. |
| Evidence | Generated matrix; `plan_validate_json`; valid, empty-target, malformed, unknown-field, trailing-document, digest-tamper, provenance, unsupported-version, input-selection, response-constant, and deterministic-generation tests. Non-applicable classes are explicitly marked N/A. |
| Verdict | **PASS.** |

### `apply`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `apply`; argv `fbrcm apply <plan>`. |
| Arguments | Exactly one required plan path; `-` consumes one stdin plan document and any other nonblank path reads one file. |
| Options | `--dry-run=false`, `--yes|-y=false`, `--json=false`, and inherited global options. Dry-run validates without publication/cleanup; yes bypasses only the aggregate confirmation; stateful/stateless policy must equal the plan. |
| Selection | N/A for user selectors. The immutable, canonically ordered plan targets are the complete execution set; no runtime broadening or replanning occurs. |
| Stdin | Strict `stdin:publication_plan`; valid document, malformed/null/empty, locally invalid template, tamper, unsupported version, and file-versus-stdin branches are schema/runtime covered. |
| Success | Registered `result` with plan ID, dry-run marker, accepted publication count, and at least one ordered target item. Statuses are `unchanged`, `would-publish`, `published`, `already-applied`, `conflict`, `publish-failed`, `published-hook-failed`, and `published-cache-failed`; dry-run/live reachability, validation source, error stage, and publication-version presence are correlated. `published_count` counts the three Firebase-accepted statuses. |
| Failure | Plan parsing/integrity/version/staleness, interaction, typed authentication/network/Firebase/validation/publication, batch, file/startup/cancellation/timeout/internal failures. `draft.exists` is not advertised because apply cleans or preserves sources rather than rejecting them. Mixed target failures preserve typed batch details and partial-success status. |
| Warnings | `publication.non_atomic` for multiple live publish targets; `publication.post_publish_hook_failed`; `publication.cache_stale` with exact refresh selector; `publication.draft_cleanup_failed`; and `plan.source_draft_changed`. Details, target, safe message bounds, remediation argv/strategy where declared, and non-fatal/partial envelopes are covered. |
| Interaction | One aggregate confirmation only when a changed target exists, dry-run is false, and `--yes` is false. JSON mode returns typed `interaction.required` and never prompts. No confirmation occurs for an all-`none` plan. |
| Effects | Network and Firebase reads occur iff the plan has a publish action. Firebase validation/hooks occur for prepared changed targets; writes require live authorized changes. Local cache/auth/profile effects retain stateful predicates. Matching draft deletion occurs only after successful live publication or already-applied convergence, with exact source fingerprint and stateful policy. All-`none` apply is offline. Retry is safe before authorization, for all-`none`, and for publish paths where no trusted hook executed; it is not declared safe after a trusted hook ran. Multi-target publication is destructive and non-atomic. |
| Schemas | Command input/response IDs plus shared publication-plan/stdin, capability, semantic, envelope, and error IDs. |
| Documentation | `docs/CLI.md § Publication plans` and `§ fbrcm apply`; behavior, stdin, warning, response, retry, and stateless rules in `docs/cli-contract.md`. |
| Evidence | Generated 15-class record; `apply_stateless_json`; no-Firebase no-op runtime; every preflight and publication status; every apply warning; exact draft cleanup/drift/failure; impossible response states; plan input/integrity schema; interaction/effect predicates; failure/outcome/batch/redaction suites; deterministic generation. |
| Verdict | **PASS.** |

## Criterion results

| Criteria | Result |
| --- | --- |
| `INV-01`–`INV-04` | PASS: exact 112-member equality; unique identities/paths/references; runtime, schemas, lock, capabilities, and docs agree on unreleased 1.0.0; all executable and failure-stage JSON paths retain envelope coverage. |
| `ARG-01`–`ARG-07` | PASS: Cobra arity and flags agree; `--plan-out` publishes Unicode trimming, nonblank/exclusive destination and `-` rejection; plan integrity/local-validation deferral has typed semantic operators and failures. |
| `SEL-01`–`SEL-05` | PASS: all previous selector records revalidated; plan targets are fixed input rather than an undeclared selector and plan producers preserve their command-specific selection provenance. |
| `STDIN-01`–`STDIN-04` | PASS: apply/show/validate advertise the concrete publication-plan document, exact path-versus-stdin selection, forward-compatible local decoder shape, local preparation boundary, and no human-only transport. |
| `OUT-01`–`OUT-06` | PASS: one clean envelope; statically registered DTOs; target/count/status correlations; exact private plan artifact media, encoding, destination, overwrite, size, and SHA metadata; sensitive plan bytes absent from stdout. |
| `ERR-01`–`ERR-06` | PASS: plan errors are typed; producer-only unreachable parse/integrity codes and apply-only `draft.exists` were removed; apply failures/warnings, safe text, remediation, batch aggregation, and exit/outcome behavior are closed and tested. |
| `BEH-01`–`BEH-06` | PASS: conditional apply network, reads, writes, hooks, caches, and draft cleanup match runtime; dry-run/all-none/already-applied/stateless suppression is explicit; confirmation and trusted-hook stop points partition retry safety; all predicates have defined typed semantics. |
| `INT-01`–`INT-04` | PASS: plan creation never confirms; apply confirmation metadata matches the live changed branch and bypass; JSON mode performs no prompt/launch/picker/editor; applicable interaction previews remain schema-valid. |
| `DOC-01`–`DOC-04` | PASS: command tree, detailed sections, shared plan/stdin/artifact/retry rules, examples, runtime, capabilities, and schemas agree. |
| `GEN-01`–`GEN-05` | PASS: two final generations are byte-identical; all 236 schemas compile and resolve; exhaustive inventories and generated evidence matrix pass; root and E2E tests, vet, and lint pass. |

## Test-class evidence

`cli/app/testdata/contract_v1_audit_evidence.golden.json` contains 112 command
records and all 15 fixed test-class cells per command. Each applicable cell
cites a checked test symbol and any E2E scenario for that command; every N/A
cell contains a reason. Catalog references are themselves verified against the
test source tree.

| Class | Passing evidence |
| --- | --- |
| Discovery | Compact/detailed capability equality, published schema lookup/reference compilation, DTO/behavior/docs inventory, capability and schema-list E2E snapshots. |
| Invocation | Exact Cobra arity, unknown option/command, dependency/exclusion, semantic input schema, and new plan path/stdin rules. |
| Boundary | Shared typed-value matrices plus plan-out whitespace/`-`, nonempty targets, object-only Remote Config, provenance, digest, ordering, and tamper cases. |
| Effectiveness | Detailed behavior golden, machine-ignored option checks, plan confirmation suppression, and hook-dependent plan retry conditions. |
| Selection | Existing selector composition suites and plan-producer runtime selection evidence; N/A for the three fixed-document commands. |
| Stdin | Remote Config/credential/theme suites plus publication-plan valid, malformed, null/empty, local-invalid, file-precedence, and unconsumed-stdin semantics. |
| Success | Runtime envelopes for all 112 commands through 169 E2E scenarios or audited unit fallbacks; apply and plan status/summary variants included. |
| No-op | Shared no-match/already-applied evidence plus apply all-`none`, prepared, already-applied, and stale preflight classification. |
| Failure | Closed capability codes, typed runtime failures, schema validation, exit mapping, unknown option, plan parse/integrity/stale/version cases. |
| Interaction | Confirmation/OAuth/input/selection suites; plan producer suppression and apply trigger/bypass conditions; plan inspection explicitly N/A. |
| Warning | Reachable warning-code schemas and runtime; all five apply warnings include details and declared remediation. |
| Batch | All-success/mixed/all-failed aggregation suites; apply publication result classification and typed batch construction. |
| Effects | Detailed predicates plus instrumented network/cache/draft/hook/write suppression and execution; apply all-none and cleanup branches added. |
| Artifact | Existing inline/destination encodings and exact byte invariants plus private publication-plan creation, mode, exclusivity, destination, byte count, and SHA. |
| Determinism | Generator determinism unit test and the two final repository generation passes reported below. |

## Closed findings

### Finding 1

Criterion: `BEH-01`, `BEH-02`, `BEH-04`, `BEH-05`, `INT-03`

Commands: `apply` and all 17 `--plan-out` producers

Evidence: `cli/commands/apply/commands.go`, `cli/shared/rc/plan.go`, initial
detailed capability golden, and missing branch regressions.

Observed: apply advertised unconditional network/read behavior, omitted source
draft deletion, and under-specified retry stop points; plan creation inherited
ordinary confirmation metadata and unconditional retry safety.

Required: exact effects, suppression, interaction, and retry conditions for
every reachable stop point.

Remedy class: capability, documentation, test, generated.

Retest: all-none/publish plans; stateful/stateless and dry-run branches;
confirmation/bypass; hook executed/not-executed; matching/changed/broken draft;
detailed capability schema and golden. **Closed.**

### Finding 2

Criterion: `ARG-02`, `ARG-03`, `ARG-07`, `ERR-03`

Commands: all 17 `--plan-out` producers and `apply`

Evidence: `cli/shared/planflag/planflag.go`, generated producer input schemas,
and generated problem-code sets.

Observed: runtime trimmed the destination and rejected `-`, but schemas omitted
those semantics; producers advertised unreachable plan parse/integrity errors,
and apply advertised unreachable `draft.exists`.

Required: runtime-exact grammar/normalization and closed reachable problems.

Remedy class: schema, capability, documentation, test, generated.

Retest: omitted, whitespace, trimmed path, `-`, exclusive destination, producer
and apply closed problem sets. **Closed.**

### Finding 3

Criterion: `STDIN-01`, `STDIN-02`, `ARG-07`, `DOC-01`

Commands: `apply`, `plan.show`, `plan.validate`

Evidence: `core/rc/publication/plan.go`, initial publication-plan/stdin schemas,
and missing input-selection annotation.

Observed: schemas did not state path-versus-stdin consumption or the complete
integrity boundary, claimed remote validation on offline embedded templates,
rejected forward-compatible fields accepted by runtime, and admitted empty
targets and JSON `null` Remote Config snapshots.

Required: concrete locally accepted stdin shape, exact source selection, and
typed local versus remote validation boundaries.

Remedy class: runtime, schema, documentation, test, generated.

Retest: file/`-` selection, unused stdin, malformed/null/array/unknown fields,
empty targets, local template preparation, action/provenance, digests/order/ID.
**Closed.**

### Finding 4

Criterion: `OUT-03`, `OUT-06`

Commands: all 17 `--plan-out` producers

Evidence: initial `PlanCreatedResult`, written plan bytes, and generated
producer response schemas.

Observed: success exposed only path and counts, so an agent could not verify
the media, encoding, exact bytes, overwrite result, size, or digest of the
written artifact.

Required: exact non-secret artifact metadata and structured invariants matching
the bytes written.

Remedy class: runtime, schema, documentation, test, generated.

Retest: private exclusive write, no replacement, fixed media/encoding,
destination/path equality, exact size and SHA, count sum, response schemas.
**Closed.**

### Finding 5

Criterion: `OUT-03`, `OUT-04`, `ERR-02`, `ERR-06`

Commands: `apply`, `plan.show`, `plan.validate`

Evidence: initial response schemas and apply status/warning code paths.

Observed: schemas admitted impossible target counts, validity, dry-run/status,
validation-source, version, error-stage, and accepted-publication-count states;
new apply statuses and warnings lacked command-specific runtime evidence.

Required: every reachable result variant and warning must be closed,
correlated, safe, schema-valid, and tested.

Remedy class: schema, runtime refactor, documentation, test, generated.

Retest: every preflight/publication status; dry-run/live sets; hook/cache/
conflict/generic failure; non-atomic, hook, cache, cleanup, and changed-source
warnings; impossible DTO rejection; summary/validation count rules. **Closed.**

### Finding 6

Criterion: `GEN-01`, `GEN-04`, `GEN-05`

Commands: all 112 commands, with changed evidence for the 20 plan/apply records

Evidence: stale generated audit evidence, the two initially reported E2E
snapshot mismatches, and missing catalog references for the new runtime tests.

Observed: generated evidence did not cite the new plan E2E/runtime branches and
the reported capability/schema-list snapshots expected the previous inventory.

Required: current deterministic artifacts, complete applicable test-class
evidence, and passing repository/E2E gates.

Remedy class: test, generated.

Retest: evidence-catalog symbol validation, 112-by-15 matrix, capability count
112, schema count 236, full root/E2E tests, vet, lint, and two generation
passes. **Closed.**

## Determinism and repository gates

| Check | Result |
| --- | --- |
| `go run ./cmd/schemagen` twice | PASS; aggregate generated-artifact SHA-256 remained `e125c596fdbeb23fb5758e1680f2a7bfc9c144f29af23def0b6ae9e4e2392a5b`. |
| Schema compilation/reference tests | PASS; 236 unique Draft 2020-12 files and no unresolved references. |
| Generated audit matrix | PASS; 112 commands, 15 class cells each, no blank cell or missing test symbol. |
| `go test -count=1 ./...` (root) | PASS. |
| `go test -count=1 ./...` (`e2e`) | PASS; 169 scenarios plus harness/readguard tests. |
| `go vet ./...` (root and `e2e`) | PASS. |
| `golangci-lint run` (root and `e2e`) | PASS; zero issues after the sole staticcheck capitalization finding was corrected and retested. |
| Final diff | PASS; only intended runtime, contract, generator, tests, generated schemas/goldens, documentation, and this audit report changed. |

All twelve section-7 acceptance conditions pass, every applicable criterion
and test class passes, and there are zero unresolved findings. Under audit
standard 1.0.0, the required verdict is **AUTHORITATIVE: PASS**.
