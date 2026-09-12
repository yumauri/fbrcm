# CLI JSON contract audit — setup and shared contract closure

Repository revision: `0bf46824db7c75af2a22e5efe13cd26fcc760977` plus the audited repair worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `118`
Input schemas: `118`
Response schemas: `118`
Shared schemas: `12`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0` unresolved (`4` found and closed)

The audit scope was frozen at revision
`0bf46824db7c75af2a22e5efe13cd26fcc760977`. Initial inspection and adversarial
runtime probing produced `NOT AUTHORITATIVE: FAIL`. Four criterion-linked
defects were repaired without changing audit standard 1.0.0. The final
unreleased contract lock has SHA-256
`0cbe20eecdae367e7ac11404c9cfff420a19e9e7afa2895416b1e5a1bff6b680`.

## Inventory

Exact set comparison: `E = C = I = R = D = B = M`, with 118 members in every
set and zero symmetric-difference members. The only executable command added
after the September 8 audit is `setup`.

There are 248 unique Draft 2020-12 files under `schemas/cli/1.0.0`: 118 input
schemas, 118 response schemas, and 12 shared schemas (`capability`, `envelope`,
`error`, `publication_plan`, `semantic`, six Remote Config stdin schemas, and
the theme stdin schema). All `$id` values are unique, all references compile
and resolve, and every capability points to the matching command input and
response schema. Compact and detailed capabilities, the embedded registry,
response DTO registration, the behavior manifest, documentation inventory,
and generated evidence matrix exactly match the executable tree.

## Command audit records

The 117 records in the
[September 8 audit](cli-json-contract-2026-09-08.md), including records
incorporated there by reference, are incorporated here by reference. This is
record compression, not skipped testing: all 117 were regenerated and rerun
through the current discovery, invocation, schema, runtime, behavior,
documentation, root-module, and E2E gates. The shared corrections to profile
failure closure and stateless/profile option effectiveness were reapplied to
every affected record.

The one new record follows. Together with the incorporated records, it forms
exactly one complete section-4 record for every member of `E`.

### `setup` — `fbrcm setup`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `setup`; argv path `setup`; executable leaf. |
| Arguments | No positional arguments. Too few is impossible; extra arguments are rejected by Cobra and the invocation schema. |
| Options | `--noopen` is a Boolean defaulting to false and is ineffective in JSON mode. Global `--profile` uses trim-then-safe-path-segment normalization, `--no-local-config` is Boolean, and `--timeout` is a positive duration. `--stateless=true` is rejected and marked ineffective. |
| Selection | Effective `--profile` selects one existing profile exactly; an absent explicit profile returns typed `profile.not_found`. With no effective profile, envelope initialization may bootstrap `default`. |
| Stdin | `N/A`; capability support is false and normalized stdin is null. |
| Success | `N/A` in JSON mode: after valid startup, the guided human workflow always stops with a typed interaction failure. The response schema explicitly registers no successful data. |
| Failure | Closed startup/argument/profile/filesystem/timeout/cancel/internal set plus `interaction.required`. Missing explicit profile is `profile.not_found`, category `not_found`, non-retryable, exit 6. Guided setup is category `interaction`, exit 10, with typed `guided_setup` details. |
| Warnings | `N/A`; no warning branch is reachable. |
| Interaction | After valid startup, JSON always returns `interaction.required` without prompting or opening a browser, editor, or picker. There is no JSON-mode bypass and no remediation argv; completion requires removing `--json`. Early typed startup failures remain possible. |
| Effects | No network access and no setup/auth/project mutation. The ordinary envelope profile initialization may conditionally write global/profile/cache state only when default-profile bootstrap is required. Existing profile configuration is unchanged. Non-destructive and idempotent. |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:setup:input`, `urn:fbrcm:schema:cli:1.0.0:command:setup:response`, plus the shared envelope, error, capability, and semantic schemas. |
| Documentation | `docs/CLI.md`, `fbrcm setup`; `docs/cli-contract.md`, interaction and option-effectiveness rules. |
| Evidence | Exhaustive discovery/arity/unknown-option/schema tests; setup command unit tests; `TestJSONSetupReturnsGuidedInteraction`; `TestJSONSetupDoesNotMutateAnExistingProfile`; `TestJSONMissingProfileIsAConformingSelectionFailure`; generated 15-class matrix; `setup_interaction_json` E2E scenario. |
| Verdict | **PASS.** |

## Criterion results

| Criteria | Result |
| --- | --- |
| `INV-01`–`INV-04` | PASS: exact 118-member equality, unique paths/IDs/references, unreleased 1.0.0 agreement, and conforming root/help/completion/startup envelopes. |
| `ARG-01`–`ARG-07` | PASS: arity, types, defaults, repetition, enums/bounds/grammars, dependencies/exclusions, normalization, omitted/empty distinctions, runtime boundaries, and option effectiveness agree. Schema-rejected `--stateless=true` values are not described as effective; completion profile values are explicitly ineffective. |
| `SEL-01`–`SEL-05` | PASS: literal, mode-prefixed, repeated, default, project/app/version/resource/profile, and stateless selection rules and typed ambiguity/not-found branches agree. |
| `STDIN-01`–`STDIN-04` | PASS: every stdin-capable command retains exact modes and concrete schemas; setup explicitly supports none. |
| `OUT-01`–`OUT-06` | PASS: clean envelopes, DTO/no-data registration, empty/null/count/status/artifact invariants, reachable variants, and actual runtime documents agree. Missing-profile output now validates as its real typed failure rather than being replaced by contract enforcement. |
| `ERR-01`–`ERR-06` | PASS: closed typed problems/warnings, details, retryability, remediation, aggregation, semantic statuses, redaction, and bounds agree. Setup interaction no longer advertises unusable machine remediation. |
| `BEH-01`–`BEH-06` | PASS: effects, network, destruction, idempotency, suppression, and predicates match runtime. Setup bootstrap effects are proven in an isolated state root, including the no-bootstrap branch. |
| `INT-01`–`INT-04` | PASS: JSON never prompts or launches external UI; confirmation, OAuth, destination, selection, and guided-setup interactions are typed and their bypass or no-bypass behavior is accurate. |
| `DOC-01`–`DOC-04` | PASS: CLI reference, contract guide, examples, schemas, capabilities, runtime, setup behavior, and shared effectiveness semantics agree. |
| `GEN-01`–`GEN-05` | PASS: two final generations were byte-identical; 248 schemas compile; exhaustive inventories and the 118-by-15 matrix pass; root/E2E test, vet, and lint gates pass. |

## Test-class evidence

`cli/app/testdata/contract_v1_audit_evidence.golden.json` has 118 command
records, exactly 15 class cells per command, and 53 checked evidence symbols.
Every applicable cell cites a valid symbol and relevant JSON E2E scenario;
every `N/A` cell has a nonblank reason.

| Class | Passing evidence |
| --- | --- |
| Discovery | Exact executable/capability/schema/DTO/behavior/docs equality, reference compilation, goldens, embedded capabilities, setup discovery, and E2E discovery. |
| Invocation/boundary/effectiveness | Exhaustive arity and option failures, normalized schemas, enums/bounds/grammars, dependency/exclusion tests, command-specific ignored options, conditional options, rejected stateless options, and completion profile applicability. |
| Selection/stdin | Shared and command-local selector semantics, missing-profile runtime conformance, and all concrete stdin valid, malformed, boundary, and restriction suites. |
| Success/no-op/failure/warning | Runtime envelopes, impossible-state rejection, no-data registration, status/count correlations, closed codes, classifications, warning schemas, safety, and redaction. |
| Interaction/batch/effects | Confirmation/OAuth/destination/guided-setup branches, no-remediation setup behavior, mixed batch aggregation, instrumented remote/local/cache/draft/hook/profile effects, default-profile bootstrap, existing-profile non-mutation, and stateless suppression. |
| Artifact/determinism | Inline/destination representations, exact bytes/size/digest/overwrite behavior, generator unit determinism, and the two repository generation passes. |

## Closed findings

### Finding 1 — reachable profile selection failure missing from response contracts

Criterion: `OUT-02`, `OUT-05`, `ERR-02`, `DOC-01`, `GEN-04`.

Commands: every command whose effective global `--profile` reaches ordinary
profile initialization, excluding `doctor` and profile-lifecycle commands that
handle absence through distinct control flow.

Evidence: `ops/contract/problem_codes.go:79`, generated command response
schemas including `schemas/cli/1.0.0/setup.response.schema.json`, and
`cli/app/contract_test.go:2918`, `cli/app/contract_test.go:3617`.

Observed: `setup --json --profile missing` reached typed `profile.not_found`,
but that code was absent from its advertised closed response schema. Runtime
response validation consequently replaced the real selection failure with
`internal.contract_violation` and exit 15.

Required: every actual typed failure must validate against the command's
advertised response schema with the documented code, category, outcome, and
exit status.

Remedy class: capability, schema, test, generated.

Retest: exhaustive problem-code closure plus an isolated missing-profile setup
invocation now emits schema-valid `profile.not_found`, category `not_found`,
exit 6. **Closed.**

### Finding 2 — rejected stateless option advertised as effective

Criterion: `ARG-06`, `DOC-01`, `GEN-04`.

Commands: every command with `supports.stateless: false`; `mcp` retains its
documented whole-operation early-rejection exception.

Evidence: `ops/contract/capability.go:337`, affected generated input schemas
including `schemas/cli/1.0.0/setup.input.schema.json`, and
`cli/app/contract_test.go:2125`.

Observed: input schemas rejected `--stateless=true`, but detailed capability
flags still reported `effective: true`, directly contradicting definition 1.4.

Required: an option forbidden in JSON mode must be schema-rejected and must
not simultaneously be described as effective for machine use.

Remedy class: capability, schema, documentation, test, generated.

Retest: every non-stateless capability now reports the flag ineffective; its
input schema publishes `x-fbrcm-effective: false` and rejects true where the
command-level invocation is otherwise supported. **Closed.**

### Finding 3 — completion profile option advertised as effective

Criterion: `ARG-06`, `DOC-01`, `GEN-04`.

Commands: `completion.bash`, `completion.fish`, `completion.powershell`, and
`completion.zsh`.

Evidence: `ops/contract/capability.go:507`, generated completion input schemas,
and the property-specific regression at `cli/app/contract_test.go:2090`.

Observed: completion execution bypasses argv profile application, but the
capabilities and invocation schemas advertised `--profile` as effective.

Required: an argv option accepted but never applied in JSON mode must publish
`effective: false` and `x-fbrcm-effective: false`.

Remedy class: capability, schema, documentation, test, generated.

Retest: all four completion records and schemas now mark profile ineffective;
the shared regression inspects the profile property itself. **Closed.**

### Finding 4 — guided setup advertised unusable machine remediation

Criterion: `ERR-02`, `INT-02`, `DOC-01`, `GEN-04`.

Commands: `setup`.

Evidence: `cli/commands/setup/commands.go:52`,
`cli/commands/setup/commands_test.go:310`,
`cli/app/contract_test.go:3585`, and
`e2e/testdata/scenarios/setup_interaction_json/stdout.golden`.

Observed: guided setup returned `retry_with_arguments` containing `setup`, even
though no JSON-mode argument can bypass the interaction; completing the flow
requires a human terminal invocation without global `--json`.

Required: interaction metadata must state the actual trigger and bypass or
required option, and remediation strategy/argv must accurately represent a
machine-usable retry.

Remedy class: runtime, documentation, test, generated.

Retest: setup now returns the typed `guided_setup` interaction with empty
remediation, performs no prompt or launch, and leaves an existing profile
unchanged. Unit, schema-valid runtime, and E2E snapshots pass. **Closed.**

## Determinism and repository gates

| Check | Result |
| --- | --- |
| `go run ./cmd/schemagen` twice | PASS; byte-identical generated trees; aggregate generated-diff SHA-256 before and after the final run `403a780c25aa22eb6d9b54ee1693bb6e68d753d5083b3849b85dda17aa1dcff5`. |
| Schema compilation/reference tests | PASS; 248 unique Draft 2020-12 files, no duplicate/dangling IDs or references. |
| Generated audit matrix | PASS; 118 commands × 15 classes, 53 checked symbols, no blank or invalid cells. |
| `go test -count=1 ./...` (root) | PASS. |
| `go test -count=1 ./...` (`e2e`) | PASS. |
| `go vet ./...` (root and `e2e`) | PASS. |
| `golangci-lint run` (root and `e2e`) | PASS; zero issues. |
| Final diff | PASS; only contract/runtime repairs, tests, documentation, generated artifacts/goldens/lock, E2E snapshots, and this audit report changed. No pre-existing user changes existed. |

The first root and E2E test attempts inside the restricted sandbox could not
bind their local fake servers. The final authorized runs used loopback access
and passed. One earlier concurrent E2E run timed out in the standalone
Hoverfly harness after all CLI scenarios passed; the isolated final E2E gate
passed all packages.

All twelve section-7 acceptance conditions pass, every applicable criterion
and test class passes, and there are zero unresolved findings. Under audit
standard 1.0.0, the required verdict is **AUTHORITATIVE: PASS**.
