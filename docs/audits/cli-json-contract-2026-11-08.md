# CLI JSON contract audit — 2026-11-08

```text
Repository revision: 43e9fcb904d825cac0b167a7828351f8141cfdd9 plus the audited repair worktree
Contract version: 1.1.0
Contract released: false
Audit standard: docs/cli-contract-audit.md version 1.0.0
Executable commands: 131
Input schemas: 131
Response schemas: 131
Shared schemas: 12 current, 12 retained
Pre-existing dirty files: none
Verdict: AUTHORITATIVE: PASS
Findings: 0 unresolved (3 found and closed)
```

## Inventory and audit scope

The starting revision, clean worktree, release state and audit standard were
recorded on November 8. Final validation completed on November 9. The audit
applied all 51 criteria in standard
1.0.0. It excludes the TUI and live Firebase behavior outside fbrcm's published
boundary. Replay and runtime tests use local fixtures and pre-recorded responses.

`E = C = I = R = D = B = M = 131`. All sets have unique IDs and
empty symmetric differences. The live executable tree includes root, help and
the four shell script generators. Cobra's private completion protocol is
excluded from machine discovery and its JSON rejection is tested. DTO
registration lookup succeeds for each executable operation, including explicit
no-data registration. Detailed discovery requires each behavior entry and rejects
unused entries. The documentation and 15-class evidence inventories independently
match the live command IDs.

The current contract has 131 input, 131 response and 12 shared schemas. The
retained contract has 272 files, including the publication-plan artifact with its
original ID. All 546 files have unique IDs and compile with Draft
2020-12 format assertions. The capability paths and input, response, error and
stdin references resolve. Appendix A contains the command records. Appendix B
lists every schema file and ID.

The [November 6 audit](cli-json-contract-2026-11-06.md) supplies the previously
verified command-specific behavior. The current records were checked against the
live inventory, generated artifacts, changes since that audit, and a fresh
run of the complete suites. Recent changes to `explain` and the configuration
key surface received direct runtime and schema checks. Fifteen additional
isolated CLI calls exercised setting, showing and resetting the Context keys
with whitespace around their selectors.

The compact and detailed capability goldens and embedded discovery records are
byte-identical to the starting tree. Generation updates the semantic schema,
the audit-evidence matrix and the active development lock. The released 1.0.0
schemas match the release tag, and the archived released lock matches the
active lock from that tag. Contract 1.1.0 remains unreleased.

## Findings and closure

### F1 — Process launch misreads option values as flags

- **Criterion** INV-04, ARG-05, OUT-01, OUT-05, DOC-01, GEN-04.
- **Commands** all CLI JSON commands, including MCP launch rejection.
- **Evidence** mode detection in `ops/contract/contract.go:119` and deadline detection in `cli/app/root.go:378` at the starting revision. [Subprocess regression](../../main_test.go#L68)
- **Observed** `fbrcm --json schema show missing --profile --json=false` emits usage text because the launch scan reads the profile value as a flag. Conversely, `fbrcm --profile --json` chooses JSON without a JSON flag. The deadline scan similarly reads option values that resemble `--timeout`.
- **Required** apply the executable command's flag grammar and preserve the JSON envelope on argument failures.
- **Remedy class** runtime, documentation, test, generated evidence.
- **Retest** option values, local shorthand values, repeats, end-of-options, unknown and malformed flags, and CLI/MCP launch boundaries.

**Closed.** Launch inspection now uses the selected command's Cobra/pflag definitions. The new subprocess regressions verify CLI and MCP behavior, and every executable command was retested with flag-like values and parse errors.

### F2 — Explain HAR timestamps can break response encoding

- **Criterion** ARG-07, STDIN-02, OUT-01, ERR-02, DOC-01, GEN-04.
- **Commands** explain and its shared capture stdin schema.
- **Evidence** HAR decoding in `core/rc/evaluate/input.go` and time bounds in `core/rc/evaluate/capture_context.go`. [Explain response schema](../../schemas/cli/1.1.0/explain.response.schema.json), [capture regressions](../../cli/app/explain_audit_test.go#L187)
- **Observed** a candidate with `startedDateTime` ending in `+24:00` decodes as a Go time, then triggers `internal.contract_violation` at exit 15 when an ambiguous-capture response encodes the candidate. A derived timestamp bound can also overflow the RFC3339 year range.
- **Required** enforce the typed capture validation boundary before reporting candidates or returning success.
- **Remedy class** runtime, schema semantics, documentation, test, generated.
- **Retest** invalid candidate offset and derived bounds, valid noncanonical input timestamp, and schema validation of the resulting envelopes.

**Closed.** The HAR parser validates candidate timestamps and computed time bounds before producing an exchange. The semantic schema publishes the boundary. The ambiguous-candidate and out-of-range-bound regressions both validate the typed failure, and the accepted single-digit-hour case still passes.

### F3 — Explain without redirected stdin reports an unrelated interaction

- **Criterion** STDIN-01, ERR-02, INT-03, DOC-01, GEN-04.
- **Commands** explain.
- **Evidence** `readExchange` in `ops/workflows/explain/commands.go` and the `explain_capture_source` rule. [Semantic schema](../../schemas/cli/1.1.0/semantic.schema.json), [capture regressions](../../cli/app/explain_audit_test.go#L187)
- **Observed** `explain flag --from - --json` without redirected stdin reads the non-interactive prompt guard and returns an input interaction at exit 10. The declared boundary is `capture.invalid` at exit 8.
- **Required** missing stdin must produce the declared typed capture failure without reading an interactive stream.
- **Remedy class** runtime, documentation, test.
- **Retest** missing and empty explicit stdin, valid redirected capture, file precedence, and the ambiguous-capture interaction.

**Closed.** The explain workflow checks for redirected input before attempting the explicit stdin read. The subprocess and workflow regressions assert the typed capture failure. File precedence and the ambiguous-capture interaction still pass.

## Criterion results

| Criterion | Result and evidence |
| --- | --- |
| INV-01 | PASS. Live discovery, exact inventory sets, published schema compilation, goldens and process envelopes. F1 covers startup mode and flag errors. |
| INV-02 | PASS. Live discovery, exact inventory sets, published schema compilation, goldens and process envelopes. F1 covers startup mode and flag errors. |
| INV-03 | PASS. Live discovery, exact inventory sets, published schema compilation, goldens and process envelopes. F1 covers startup mode and flag errors. |
| INV-04 | PASS. Live discovery, exact inventory sets, published schema compilation, goldens and process envelopes. F1 covers startup mode and flag errors. |
| ARG-01 | PASS. Cobra arity, invocation schema grammar, option dependencies and effectiveness tests. F1 covers literal values, repeats and invalid flags. F2 covers the capture timestamp boundary. |
| ARG-02 | PASS. Cobra arity, invocation schema grammar, option dependencies and effectiveness tests. F1 covers literal values, repeats and invalid flags. F2 covers the capture timestamp boundary. |
| ARG-03 | PASS. Cobra arity, invocation schema grammar, option dependencies and effectiveness tests. F1 covers literal values, repeats and invalid flags. F2 covers the capture timestamp boundary. |
| ARG-04 | PASS. Cobra arity, invocation schema grammar, option dependencies and effectiveness tests. F1 covers literal values, repeats and invalid flags. F2 covers the capture timestamp boundary. |
| ARG-05 | PASS. Cobra arity, invocation schema grammar, option dependencies and effectiveness tests. F1 covers literal values, repeats and invalid flags. F2 covers the capture timestamp boundary. |
| ARG-06 | PASS. Cobra arity, invocation schema grammar, option dependencies and effectiveness tests. F1 covers literal values, repeats and invalid flags. F2 covers the capture timestamp boundary. |
| ARG-07 | PASS. Cobra arity, invocation schema grammar, option dependencies and effectiveness tests. F1 covers literal values, repeats and invalid flags. F2 covers the capture timestamp boundary. |
| SEL-01 | PASS. Command-specific selection and composition records, fixture-based tests, capture-entry selection, and direct config key probes. |
| SEL-02 | PASS. Command-specific selection and composition records, fixture-based tests, capture-entry selection, and direct config key probes. |
| SEL-03 | PASS. Command-specific selection and composition records, fixture-based tests, capture-entry selection, and direct config key probes. |
| SEL-04 | PASS. Command-specific selection and composition records, fixture-based tests, capture-entry selection, and direct config key probes. |
| SEL-05 | PASS. Command-specific selection and composition records, fixture-based tests, capture-entry selection, and direct config key probes. |
| STDIN-01 | PASS. Concrete decoder fixtures, stdin restrictions, absent and malformed input process cases. F2 and F3 cover timestamp validation and missing capture stdin. |
| STDIN-02 | PASS. Concrete decoder fixtures, stdin restrictions, absent and malformed input process cases. F2 and F3 cover timestamp validation and missing capture stdin. |
| STDIN-03 | PASS. Concrete decoder fixtures, stdin restrictions, absent and malformed input process cases. F2 and F3 cover timestamp validation and missing capture stdin. |
| STDIN-04 | PASS. Concrete decoder fixtures, stdin restrictions, absent and malformed input process cases. F2 and F3 cover timestamp validation and missing capture stdin. |
| OUT-01 | PASS. Actual process envelopes validate with process status and newline checks. Registered DTO, empty/no-op, batch, artifact and runtime conformance tests cover the applicable branches. |
| OUT-02 | PASS. Actual process envelopes validate with process status and newline checks. Registered DTO, empty/no-op, batch, artifact and runtime conformance tests cover the applicable branches. |
| OUT-03 | PASS. Actual process envelopes validate with process status and newline checks. Registered DTO, empty/no-op, batch, artifact and runtime conformance tests cover the applicable branches. |
| OUT-04 | PASS. Actual process envelopes validate with process status and newline checks. Registered DTO, empty/no-op, batch, artifact and runtime conformance tests cover the applicable branches. |
| OUT-05 | PASS. Actual process envelopes validate with process status and newline checks. Registered DTO, empty/no-op, batch, artifact and runtime conformance tests cover the applicable branches. |
| OUT-06 | PASS. Actual process envelopes validate with process status and newline checks. Registered DTO, empty/no-op, batch, artifact and runtime conformance tests cover the applicable branches. |
| ERR-01 | PASS. Typed problem, warning, redaction, aggregation and exit-status tests. F2 and F3 validate the expected capture problem and exit status. |
| ERR-02 | PASS. Typed problem, warning, redaction, aggregation and exit-status tests. F2 and F3 validate the expected capture problem and exit status. |
| ERR-03 | PASS. Typed problem, warning, redaction, aggregation and exit-status tests. F2 and F3 validate the expected capture problem and exit status. |
| ERR-04 | PASS. Typed problem, warning, redaction, aggregation and exit-status tests. F2 and F3 validate the expected capture problem and exit status. |
| ERR-05 | PASS. Typed problem, warning, redaction, aggregation and exit-status tests. F2 and F3 validate the expected capture problem and exit status. |
| ERR-06 | PASS. Typed problem, warning, redaction, aggregation and exit-status tests. F2 and F3 validate the expected capture problem and exit status. |
| BEH-01 | PASS. Exhaustive behavior manifest, predicate validation, fake-service side-effect tests, and replayed network journals. Applicable effects and retry-safety evidence are linked per command. |
| BEH-02 | PASS. Exhaustive behavior manifest, predicate validation, fake-service side-effect tests, and replayed network journals. Applicable effects and retry-safety evidence are linked per command. |
| BEH-03 | PASS. Exhaustive behavior manifest, predicate validation, fake-service side-effect tests, and replayed network journals. Applicable effects and retry-safety evidence are linked per command. |
| BEH-04 | PASS. Exhaustive behavior manifest, predicate validation, fake-service side-effect tests, and replayed network journals. Applicable effects and retry-safety evidence are linked per command. |
| BEH-05 | PASS. Exhaustive behavior manifest, predicate validation, fake-service side-effect tests, and replayed network journals. Applicable effects and retry-safety evidence are linked per command. |
| BEH-06 | PASS. Exhaustive behavior manifest, predicate validation, fake-service side-effect tests, and replayed network journals. Applicable effects and retry-safety evidence are linked per command. |
| INT-01 | PASS. Interaction-required and bypass tests, preview conformance, and no-prompt/no-launch evidence. F3 retests the distinct missing-stdin and candidate-choice branches. |
| INT-02 | PASS. Interaction-required and bypass tests, preview conformance, and no-prompt/no-launch evidence. F3 retests the distinct missing-stdin and candidate-choice branches. |
| INT-03 | PASS. Interaction-required and bypass tests, preview conformance, and no-prompt/no-launch evidence. F3 retests the distinct missing-stdin and candidate-choice branches. |
| INT-04 | PASS. Interaction-required and bypass tests, preview conformance, and no-prompt/no-launch evidence. F3 retests the distinct missing-stdin and candidate-choice branches. |
| DOC-01 | PASS. CLI command sections, shared rules, example tests and current schemas. F1 through F3 update the affected rules and runtime boundaries. |
| DOC-02 | PASS. CLI command sections, shared rules, example tests and current schemas. F1 through F3 update the affected rules and runtime boundaries. |
| DOC-03 | PASS. CLI command sections, shared rules, example tests and current schemas. F1 through F3 update the affected rules and runtime boundaries. |
| DOC-04 | PASS. CLI command sections, shared rules, example tests and current schemas. F1 through F3 update the affected rules and runtime boundaries. |
| GEN-01 | PASS. Two byte-identical generation passes, complete schema/reference compilation, current goldens, all 15 evidence classes and the repository gates below. |
| GEN-02 | PASS. Two byte-identical generation passes, complete schema/reference compilation, current goldens, all 15 evidence classes and the repository gates below. |
| GEN-03 | PASS. Two byte-identical generation passes, complete schema/reference compilation, current goldens, all 15 evidence classes and the repository gates below. |
| GEN-04 | PASS. Two byte-identical generation passes, complete schema/reference compilation, current goldens, all 15 evidence classes and the repository gates below. |
| GEN-05 | PASS. Two byte-identical generation passes, complete schema/reference compilation, current goldens, all 15 evidence classes and the repository gates below. |

## Final validation

The process audit compiled all 546 schemas and checked
**1,505 envelopes** across 131 commands. Every
invocation used an isolated home, configuration and cache, explicit stdin
handling, offline mode and a process limit. Cases included minimal and bare
argv, help, unknown and malformed flags, invalid JSON mode values, flag-like
option values, timeout, missing profile, stateless boundaries, malformed and
empty stdin documents, and both hidden shell transport spellings. Every JSON
result used a published response schema, contained one document and one trailing
newline, and matched the process exit status. The probes rejected unexpected
internal fallbacks. Fixture-based and replay tests exercise populated, batch,
warning, interaction and artifact branches.

| Gate | Result |
| --- | --- |
| `go run ./cmd/schemagen`, twice | PASS, 552 byte-identical generated assets |
| `go test -count=1 ./...` | PASS |
| `go vet ./...` | PASS |
| `golangci-lint run` | PASS, zero issues |
| Replay module tests, vet and lint | PASS, zero lint issues |
| Inventory, capability and evidence goldens | PASS |
| All current and retained schemas and real process envelopes | PASS |
| Released 1.0.0 schemas and archived lock | PASS, byte-identical to release |
| Final intended diff and `git diff --check` | PASS |

The local loopback test servers required expanded execution access. The audit
used isolated state and replayed network responses. It made no live Firebase
mutations.

- Active lock file SHA-256 `2bdfee3dedbd9a6fd4bf43390478e43a8b3fbf099a24eaab144177d9e03a3a8b`
- Sorted generated-asset digest manifest SHA-256 `d63d6b5e0f8a1ca41f61c8c7cd2e5e3a910c5bb79583526834b941d563bf1677`
- Captured process evidence SHA-256 `f06d48b51d8c9ae46218886628f4c3050ed420ce48e4fe6fdaaeef37d4785647`

All twelve section-7 conditions pass for the audited worktree.

## Appendix A — complete command audit records

Each record links to the current input and response schemas, capability,
documentation and applicable test classes. All commands share the global
JSON-mode launch rules and the process regression `TestCLILaunchFlags`.
Each record's failure cell incorporates the shared structured problem schema
and semantic process status rules.

### `add`

| Column | Audited content |
| --- | --- |
| Command | `add`; `fbrcm add`. |
| Arguments | required parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--expr`, `--group`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--type`, `--use-in-app-default`, `--value`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); [input schema](../../schemas/cli/1.1.0/add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:remote_config`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/add.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/add/commands.go); [response schema](../../schemas/cli/1.1.0/add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L81). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L81). |
| Effects | Level 3; network `conditional`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L81). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:add:input`; `urn:fbrcm:schema:cli:1.1.0:command:add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1408); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L81); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `apply`

| Column | Audited content |
| --- | --- |
| Command | `apply`; `fbrcm apply`. |
| Arguments | required plan. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/apply.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962). |
| Options | `--dry-run`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); [input schema](../../schemas/cli/1.1.0/apply.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/apply.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:publication_plan`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/apply.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/apply/commands.go); [response schema](../../schemas/cli/1.1.0/apply.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/apply.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `plan.source_draft_changed`, `publication.cache_stale`, `publication.draft_cleanup_failed`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/apply.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L232). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L232). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L232). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:apply:input`; `urn:fbrcm:schema:cli:1.1.0:command:apply:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1385); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apply`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L232); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `apps.config`

| Column | Audited content |
| --- | --- |
| Command | `apps.config`; `fbrcm apps config`. |
| Arguments | required app. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/apps_config.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`, `--to`, `--update`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); [input schema](../../schemas/cli/1.1.0/apps_config.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/apps_config.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.1.0/apps_config.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/apps_config.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/apps_config.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L352). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L352). |
| Effects | Level 2; network `conditional`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L352). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:apps.config:input`; `urn:fbrcm:schema:cli:1.1.0:command:apps.config:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2329); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.config`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L352); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `apps.list`

| Column | Audited content |
| --- | --- |
| Command | `apps.list`; `fbrcm apps list`. |
| Arguments | optional project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/apps_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159). |
| Options | `--cached`, `--filter`, `--include-hidden`, `--no-local-config`, `--platform`, `--profile`, `--project`, `--show-deleted`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); [input schema](../../schemas/cli/1.1.0/apps_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/apps_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.1.0/apps_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/apps_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/apps_list.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L455). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L455). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L455). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:apps.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:apps.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2288); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L455); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `apps.show`

| Column | Audited content |
| --- | --- |
| Command | `apps.show`; `fbrcm apps show`. |
| Arguments | required app. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/apps_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); [input schema](../../schemas/cli/1.1.0/apps_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/apps_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.1.0/apps_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/apps_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/apps_show.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L553). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L553). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L553). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:apps.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:apps.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2312); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L553); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.add.gcloud`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.gcloud`; `fbrcm auth add gcloud`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_add_gcloud.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125). |
| Options | `--include-hidden`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); [input schema](../../schemas/cli/1.1.0/auth_add_gcloud.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/auth_add_gcloud.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_add_gcloud.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_add_gcloud.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L651). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L651). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.gcloud:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.add.gcloud:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3570); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.gcloud`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L651); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.add.google`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.google`; `fbrcm auth add google`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_add_google.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327). |
| Options | `--include-hidden`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); [input schema](../../schemas/cli/1.1.0/auth_add_google.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/auth_add_google.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_add_google.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_add_google.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L739). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L739). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.google:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.add.google:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3494); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.google`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L739); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.add.oauth`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.oauth`; `fbrcm auth add oauth`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_add_oauth.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526). |
| Options | `--from`, `--include-hidden`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); [input schema](../../schemas/cli/1.1.0/auth_add_oauth.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/auth_add_oauth.input.schema.json). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:oauth_credentials`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/auth_add_oauth.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_add_oauth.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_add_oauth.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L842). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L842). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.oauth:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.add.oauth:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3513); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.oauth`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L842); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.add.service-account`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.service-account`; `fbrcm auth add service-account`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_add_service-account.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774). |
| Options | `--from`, `--include-hidden`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); [input schema](../../schemas/cli/1.1.0/auth_add_service-account.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/auth_add_service-account.input.schema.json). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:service_account_credentials`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/auth_add_service-account.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_add_service-account.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_add_service-account.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L939). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L939). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.service-account:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.add.service-account:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3547); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.service-account`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L939); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.bind`

| Column | Audited content |
| --- | --- |
| Command | `auth.bind`; `fbrcm auth bind`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_bind.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022). |
| Options | `--auth`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); [input schema](../../schemas/cli/1.1.0/auth_bind.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/auth_bind.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_bind.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_bind.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1036). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1036). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.bind:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.bind:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3660); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.bind`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1036); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.delete`

| Column | Audited content |
| --- | --- |
| Command | `auth.delete`; `fbrcm auth delete`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); [input schema](../../schemas/cli/1.1.0/auth_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/auth_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1127). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1127). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3647); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1127); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.list`

| Column | Audited content |
| --- | --- |
| Command | `auth.list`; `fbrcm auth list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); [input schema](../../schemas/cli/1.1.0/auth_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/auth_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1256). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1256). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3480); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1256); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.login`

| Column | Audited content |
| --- | --- |
| Command | `auth.login`; `fbrcm auth login`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_login.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528). |
| Options | `--include-hidden`, `--no-local-config`, `--noopen`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); [input schema](../../schemas/cli/1.1.0/auth_login.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/auth_login.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_login.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_login.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1348). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1348). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.login:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.login:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3613); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.login`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1348); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.path`

| Column | Audited content |
| --- | --- |
| Command | `auth.path`; `fbrcm auth path`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); [input schema](../../schemas/cli/1.1.0/auth_path.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/auth_path.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1453). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1453). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3637); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1453); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.quota-project.set`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.set`; `fbrcm auth quota-project set`. |
| Arguments | required auth_id, required quota_project_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_quota-project_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); [input schema](../../schemas/cli/1.1.0/auth_quota-project_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/auth_quota-project_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_quota-project_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_quota-project_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1550). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1550). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3603); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1550); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.quota-project.show`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.show`; `fbrcm auth quota-project show`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_quota-project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); [input schema](../../schemas/cli/1.1.0/auth_quota-project_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/auth_quota-project_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_quota-project_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_quota-project_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1639). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1639). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3598); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1639); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.quota-project.unset`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.unset`; `fbrcm auth quota-project unset`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/auth_quota-project_unset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); [input schema](../../schemas/cli/1.1.0/auth_quota-project_unset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/auth_quota-project_unset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.1.0/auth_quota-project_unset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/auth_quota-project_unset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1723). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1723). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.unset:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.unset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3607); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.unset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1723); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `cache.clear`

| Column | Audited content |
| --- | --- |
| Command | `cache.clear`; `fbrcm cache clear`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/cache_clear.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348). |
| Options | `--include-hidden`, `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); [input schema](../../schemas/cli/1.1.0/cache_clear.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/cache_clear.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.1.0/cache_clear.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/cache_clear.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1812). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1812). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:cache.clear:input`; `urn:fbrcm:schema:cli:1.1.0:command:cache.clear:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3230); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1812); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `cache.list`

| Column | Audited content |
| --- | --- |
| Command | `cache.list`; `fbrcm cache list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/cache_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554). |
| Options | `--include-hidden`, `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); [input schema](../../schemas/cli/1.1.0/cache_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/cache_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.1.0/cache_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/cache_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1904). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1904). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:cache.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:cache.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3206); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1904); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `cache.path`

| Column | Audited content |
| --- | --- |
| Command | `cache.path`; `fbrcm cache path`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/cache_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683). |
| Options | `--include-hidden`, `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); [input schema](../../schemas/cli/1.1.0/cache_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/cache_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.1.0/cache_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/cache_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1996). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1996). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:cache.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:cache.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3219); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1996); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `capabilities`

| Column | Audited content |
| --- | --- |
| Command | `capabilities`; `fbrcm capabilities`. |
| Arguments | optional command (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/capabilities.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); [input schema](../../schemas/cli/1.1.0/capabilities.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/capabilities.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.1.0/capabilities.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/capabilities.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2080). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2080). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:capabilities:input`; `urn:fbrcm:schema:cli:1.1.0:command:capabilities:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1238); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `capabilities`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2080); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `completion.bash`

| Column | Audited content |
| --- | --- |
| Command | `completion.bash`; `fbrcm completion bash`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/completion_bash.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933). |
| Options | `--include-hidden`, `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); [input schema](../../schemas/cli/1.1.0/completion_bash.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/completion_bash.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.1.0/completion_bash.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/completion_bash.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2167). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2167). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:completion.bash:input`; `urn:fbrcm:schema:cli:1.1.0:command:completion.bash:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3922); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.bash`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2167); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `completion.fish`

| Column | Audited content |
| --- | --- |
| Command | `completion.fish`; `fbrcm completion fish`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/completion_fish.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056). |
| Options | `--include-hidden`, `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); [input schema](../../schemas/cli/1.1.0/completion_fish.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/completion_fish.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.1.0/completion_fish.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/completion_fish.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2244). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2244). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:completion.fish:input`; `urn:fbrcm:schema:cli:1.1.0:command:completion.fish:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3930); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.fish`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2244); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `completion.powershell`

| Column | Audited content |
| --- | --- |
| Command | `completion.powershell`; `fbrcm completion powershell`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/completion_powershell.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179). |
| Options | `--include-hidden`, `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); [input schema](../../schemas/cli/1.1.0/completion_powershell.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/completion_powershell.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.1.0/completion_powershell.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/completion_powershell.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2321). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2321). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:completion.powershell:input`; `urn:fbrcm:schema:cli:1.1.0:command:completion.powershell:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3938); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.powershell`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2321); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `completion.zsh`

| Column | Audited content |
| --- | --- |
| Command | `completion.zsh`; `fbrcm completion zsh`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/completion_zsh.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302). |
| Options | `--include-hidden`, `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); [input schema](../../schemas/cli/1.1.0/completion_zsh.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/completion_zsh.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.1.0/completion_zsh.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/completion_zsh.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2398). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2398). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:completion.zsh:input`; `urn:fbrcm:schema:cli:1.1.0:command:completion.zsh:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3946); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.zsh`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2398); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.add`

| Column | Audited content |
| --- | --- |
| Command | `conditions.add`; `fbrcm conditions add`. |
| Arguments | optional project, required name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/conditions_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425). |
| Options | `--change-note`, `--color`, `--draft`, `--dry-run`, `--expression`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--priority`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); [input schema](../../schemas/cli/1.1.0/conditions_add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/conditions_add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.1.0/conditions_add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/conditions_add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2475). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2475). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2475). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.add:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1805); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2475); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.delete`

| Column | Audited content |
| --- | --- |
| Command | `conditions.delete`; `fbrcm conditions delete`. |
| Arguments | optional project, optional condition. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/conditions_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); [input schema](../../schemas/cli/1.1.0/conditions_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/conditions_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.1.0/conditions_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/conditions_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2605). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2605). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2605). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1858); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2605); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.edit`

| Column | Audited content |
| --- | --- |
| Command | `conditions.edit`; `fbrcm conditions edit`. |
| Arguments | required project, required condition. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/conditions_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154). |
| Options | `--change-note`, `--color`, `--draft`, `--dry-run`, `--expression`, `--include-hidden`, `--no-color`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); [input schema](../../schemas/cli/1.1.0/conditions_edit.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/conditions_edit.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.1.0/conditions_edit.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/conditions_edit.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_edit.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2735). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2735). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2735). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.edit:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.edit:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1828); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2735); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.list`

| Column | Audited content |
| --- | --- |
| Command | `conditions.list`; `fbrcm conditions list`. |
| Arguments | optional project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/conditions_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996). |
| Options | `--cached`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); [input schema](../../schemas/cli/1.1.0/conditions_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/conditions_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.1.0/conditions_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/conditions_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2865). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2865). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1730); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2865); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.move`

| Column | Audited content |
| --- | --- |
| Command | `conditions.move`; `fbrcm conditions move`. |
| Arguments | required project, required condition, required priority. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/conditions_move.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); [input schema](../../schemas/cli/1.1.0/conditions_move.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/conditions_move.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.1.0/conditions_move.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/conditions_move.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_move.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2967). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2967). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2967). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.move:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.move:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1851); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.move`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2967); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.rename`

| Column | Audited content |
| --- | --- |
| Command | `conditions.rename`; `fbrcm conditions rename`. |
| Arguments | required project, required condition, required new_name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/conditions_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); [input schema](../../schemas/cli/1.1.0/conditions_rename.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/conditions_rename.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.1.0/conditions_rename.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/conditions_rename.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_rename.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3097). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3097). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3097). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.rename:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.rename:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1847); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3097); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.show`

| Column | Audited content |
| --- | --- |
| Command | `conditions.show`; `fbrcm conditions show`. |
| Arguments | required project, required condition. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/conditions_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); [input schema](../../schemas/cli/1.1.0/conditions_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/conditions_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.1.0/conditions_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/conditions_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3227). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3227). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1756); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3227); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.validate`

| Column | Audited content |
| --- | --- |
| Command | `conditions.validate`; `fbrcm conditions validate`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/conditions_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); [input schema](../../schemas/cli/1.1.0/conditions_validate.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/conditions_validate.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.1.0/conditions_validate.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/conditions_validate.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3329). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3329). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.validate:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.validate:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1882); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3329); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.edit`

| Column | Audited content |
| --- | --- |
| Command | `config.edit`; `fbrcm config edit`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840). |
| Options | `--editor`, `--full`, `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); [input schema](../../schemas/cli/1.1.0/config_edit.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/config_edit.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_edit.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_edit.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `required`; `external_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3431). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3431). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.edit:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.edit:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3404); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3431); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.path`

| Column | Audited content |
| --- | --- |
| Command | `config.path`; `fbrcm config path`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); [input schema](../../schemas/cli/1.1.0/config_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/config_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3509). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3509). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3246); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3509); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.reset`

| Column | Audited content |
| --- | --- |
| Command | `config.reset`; `fbrcm config reset`. |
| Arguments | optional key. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [input schema](../../schemas/cli/1.1.0/config_reset.input.schema.json). |
| Selection | Case-sensitive key or subtree lookup. Supported nested prefixes are trimmed before matching. Omitting the key resets the selected scope. Unknown keys return `argument.invalid`. [Input schema](../../schemas/cli/1.1.0/config_reset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_reset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_reset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3593). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3593). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.reset:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.reset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3360); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3593); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.set`

| Column | Audited content |
| --- | --- |
| Command | `config.set`; `fbrcm config set`. |
| Arguments | required key, required value (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [input schema](../../schemas/cli/1.1.0/config_set.input.schema.json). |
| Selection | Case-sensitive supported key lookup. Nested prefixes are trimmed before the closed grammar is applied. Local project aliases require local scope. Unknown keys return `argument.invalid`. [Input schema](../../schemas/cli/1.1.0/config_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3685). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3685). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3301); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3685); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.show`

| Column | Audited content |
| --- | --- |
| Command | `config.show`; `fbrcm config show`. |
| Arguments | optional key. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [input schema](../../schemas/cli/1.1.0/config_show.input.schema.json). |
| Selection | Case-sensitive key or subtree lookup. Supported nested prefixes, including hooks, are trimmed before matching. Omitting the key shows the selected scope. Unknown keys return `argument.invalid`. [Input schema](../../schemas/cli/1.1.0/config_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3782). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3782). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3264); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3782); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.validate`

| Column | Audited content |
| --- | --- |
| Command | `config.validate`; `fbrcm config validate`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); [input schema](../../schemas/cli/1.1.0/config_validate.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/config_validate.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_validate.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_validate.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3866). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3866). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.validate:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.validate:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3383); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3866); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `delete`

| Column | Audited content |
| --- | --- |
| Command | `delete`; `fbrcm delete`. |
| Arguments | optional parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); [input schema](../../schemas/cli/1.1.0/delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:remote_config`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/delete.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/delete/commands.go); [response schema](../../schemas/cli/1.1.0/delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3958). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3958). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3958). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1694); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3958); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `doctor`

| Column | Audited content |
| --- | --- |
| Command | `doctor`; `fbrcm doctor`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/doctor.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); [input schema](../../schemas/cli/1.1.0/doctor.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/doctor.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/doctor/commands.go); [response schema](../../schemas/cli/1.1.0/doctor.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/doctor.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4124). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4124). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:doctor:input`; `urn:fbrcm:schema:cli:1.1.0:command:doctor:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3187); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `doctor`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4124); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.change-note`

| Column | Audited content |
| --- | --- |
| Command | `draft.change-note`; `fbrcm draft change-note`. |
| Arguments | required project, optional text. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/draft_change-note.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928). |
| Options | `--clear`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); [input schema](../../schemas/cli/1.1.0/draft_change-note.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/draft_change-note.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.1.0/draft_change-note.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/draft_change-note.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4208). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4208). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.change-note:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.change-note:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2004); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.change-note`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4208); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.diff`

| Column | Audited content |
| --- | --- |
| Command | `draft.diff`; `fbrcm draft diff`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/draft_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165). |
| Options | `--against`, `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); [input schema](../../schemas/cli/1.1.0/draft_diff.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/draft_diff.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.1.0/draft_diff.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/draft_diff.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4296). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4296). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.diff:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.diff:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2015); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4296); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.discard`

| Column | Audited content |
| --- | --- |
| Command | `draft.discard`; `fbrcm draft discard`. |
| Arguments | optional project (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/draft_discard.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516). |
| Options | `--all`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); [input schema](../../schemas/cli/1.1.0/draft_discard.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/draft_discard.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.1.0/draft_discard.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/draft_discard.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4386). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4386). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.discard:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.discard:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2071); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.discard`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4386); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.list`

| Column | Audited content |
| --- | --- |
| Command | `draft.list`; `fbrcm draft list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/draft_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); [input schema](../../schemas/cli/1.1.0/draft_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/draft_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.1.0/draft_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/draft_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4481). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4481). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1954); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4481); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.path`

| Column | Audited content |
| --- | --- |
| Command | `draft.path`; `fbrcm draft path`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/draft_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); [input schema](../../schemas/cli/1.1.0/draft_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/draft_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.1.0/draft_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/draft_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4577). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4577). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1981); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4577); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.publish`

| Column | Audited content |
| --- | --- |
| Command | `draft.publish`; `fbrcm draft publish`. |
| Arguments | optional project (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/draft_publish.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009). |
| Options | `--all`, `--change-note`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); [input schema](../../schemas/cli/1.1.0/draft_publish.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/draft_publish.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.1.0/draft_publish.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/draft_publish.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.draft_cleanup_failed`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/draft_publish.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4661). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4661). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4661). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.publish:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.publish:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2043); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.publish`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4661); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.show`

| Column | Audited content |
| --- | --- |
| Command | `draft.show`; `fbrcm draft show`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/draft_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--raw`, `--stateless`, `--timeout`, `--to`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); [input schema](../../schemas/cli/1.1.0/draft_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/draft_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.1.0/draft_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/draft_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `destination_conflict_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4762). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4762). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1991); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4762); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `duplicate`

| Column | Audited content |
| --- | --- |
| Command | `duplicate`; `fbrcm duplicate`. |
| Arguments | required source, required target. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/duplicate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); [input schema](../../schemas/cli/1.1.0/duplicate.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/duplicate.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/duplicate/commands.go); [response schema](../../schemas/cli/1.1.0/duplicate.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/duplicate.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/duplicate.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4853). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4853). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4853). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:duplicate:input`; `urn:fbrcm:schema:cli:1.1.0:command:duplicate:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1457); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `duplicate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4853); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `experiments.delete`

| Column | Audited content |
| --- | --- |
| Command | `experiments.delete`; `fbrcm experiments delete`. |
| Arguments | required project, required experiment_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/experiments_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); [input schema](../../schemas/cli/1.1.0/experiments_delete.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/experiments_delete.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.1.0/experiments_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/experiments_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4997). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4997). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:experiments.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:experiments.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2411); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4997); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `experiments.list`

| Column | Audited content |
| --- | --- |
| Command | `experiments.list`; `fbrcm experiments list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/experiments_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142). |
| Options | `--cached`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); [input schema](../../schemas/cli/1.1.0/experiments_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/experiments_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.1.0/experiments_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/experiments_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5091). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5091). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:experiments.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:experiments.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2371); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5091); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `experiments.show`

| Column | Audited content |
| --- | --- |
| Command | `experiments.show`; `fbrcm experiments show`. |
| Arguments | required project, required experiment_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/experiments_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); [input schema](../../schemas/cli/1.1.0/experiments_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/experiments_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.1.0/experiments_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/experiments_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5193). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5193). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:experiments.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:experiments.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2392); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5193); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `explain`

| Column | Audited content |
| --- | --- |
| Command | `explain`; `fbrcm explain <parameter>`. |
| Arguments | Exactly one required parameter key, untrimmed exact case-sensitive equality across root/groups; first exact match, or `parameter.not_found`. Grammar: [input schema](../../schemas/cli/1.1.0/explain.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860). |
| Options | `--cached`, `--entry`, `--from`, `--include-hidden`, `--no-input`, `--no-local-config`, `--profile`, `--project`, `--set`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition and effectiveness: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); dependencies and bounds: [input schema](../../schemas/cli/1.1.0/explain.input.schema.json). `--no-input` is ineffective in JSON; `--stateless=true` is rejected; `--cached`/`--update` are mutually exclusive by presence. Entry is 0 through signed 64-bit maximum, with runtime lookup. |
| Selection | `explain_project_resolution` publishes raw-number lookup, exact ID/alias/display precedence, fallback modes, visibility, primary-template defaults and capture namespace/project agreement. `explain_capture_entry_resolution` publishes original-index selection and ambiguity. No omitted all-project scope. Overrides use last canonical assignment. Full rules: [input schema](../../schemas/cli/1.1.0/explain.input.schema.json) and shared semantic schema; [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860). |
| Stdin | `json_document` structured transport accepts a capture string or HAR object; CLI stdin is raw cURL/HAR text. Concrete `urn:fbrcm:schema:cli:1.1.0:stdin:capture`. Non-`-` file source wins without consuming stdin; no capture requires project and entry=0. Missing explicit stdin and unrepresentable capture timestamps or derived time bounds return `capture.invalid` at exit 8. Capture parser and byte-budget boundary and nullable HAR shape: [input schema](../../schemas/cli/1.1.0/explain.input.schema.json). |
| Success | Registered `explanationJSON`: resolved/unresolved selection; literal/in-app-default/managed/unknown values; empty arrays; predicate traces, guarded outcomes, input provenance and nullable capture comparison. Both known and unknown results exit 0. [DTO registration](../../ops/workflows/explain/commands.go); [response schema](../../schemas/cli/1.1.0/explain.response.schema.json). |
| Failure | Closed problem catalog: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860). Includes typed capture validation (exit 8), argument failures (2), project/parameter lookup (6), and interaction (10). Categories/details/retryability and all common environment failures: [response schema](../../schemas/cli/1.1.0/explain.response.schema.json), shared error schema and exit rules. |
| Warnings | `cache.stale` and `explain.capture_version_mismatch`, typed details and non-fatal meaning: [response schema](../../schemas/cli/1.1.0/explain.response.schema.json); runtime cases: [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5291). |
| Interaction | JSON never prompts for facts or executes requests. Multiple fetches with entry=0 return safe capture candidates and `--entry`; OAuth authorization is a separate typed boundary. All triggers: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5291). |
| Effects | Level 2; conditional network and Firebase reads, cache/registry/bootstrap/token writes; non-destructive; idempotent. Reuses shared project/parameter/cache/authentication helpers. No stateless network branch, capture persistence, template mutation or hook execution. Exact predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); shared cache/runtime and replay evidence below. |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:explain:input`; `urn:fbrcm:schema:cli:1.1.0:command:explain:response`; `urn:fbrcm:schema:cli:1.1.0:error`; capture/semantic/envelope/capability resources in Appendix B. |
| Documentation | [CLI explain section](../../docs/CLI.md#L1522); [parameter explanations](../../docs/cli-contract.md#parameter-explanations); explicit documentation inventory entry `explain`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5291); `TestContractAuditMetadataHasNoInapplicableBranches`; all four permanent explain replay scenarios; shared cache-freshness tests and evaluator/capture tests. New boundary, selection, precedence and failure regressions are included in the catalog. |
| Verdict | **PASS**. F2 and F3 are closed. |

### `get`

| Column | Audited content |
| --- | --- |
| Command | `get`; `fbrcm get`. |
| Arguments | optional parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/get.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18282). |
| Options | `--all`, `--cached`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18282); [input schema](../../schemas/cli/1.1.0/get.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/get.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18282). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:remote_config`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/get.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/get/commands.go); [response schema](../../schemas/cli/1.1.0/get.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18282); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/get.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `cache.stale`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/get.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5485). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18282); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5485). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18282); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5485). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:get:input`; `urn:fbrcm:schema:cli:1.1.0:command:get:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1497); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `get`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5485); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.add`

| Column | Audited content |
| --- | --- |
| Command | `groups.add`; `fbrcm groups add`. |
| Arguments | required name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/groups_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18843). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18843); [input schema](../../schemas/cli/1.1.0/groups_add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/groups_add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18843). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.1.0/groups_add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18843); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/groups_add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/groups_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5694). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18843); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5694). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18843); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5694). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.add:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1932); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5694); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.delete`

| Column | Audited content |
| --- | --- |
| Command | `groups.delete`; `fbrcm groups delete`. |
| Arguments | required group. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/groups_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19670). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19670); [input schema](../../schemas/cli/1.1.0/groups_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/groups_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19670). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.1.0/groups_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19670); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/groups_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/groups_delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5824). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19670); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5824). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19670); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5824). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1949); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5824); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.edit`

| Column | Audited content |
| --- | --- |
| Command | `groups.edit`; `fbrcm groups edit`. |
| Arguments | required group. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/groups_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20506). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--include-hidden`, `--no-description`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20506); [input schema](../../schemas/cli/1.1.0/groups_edit.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/groups_edit.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20506). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.1.0/groups_edit.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20506); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/groups_edit.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/groups_edit.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5954). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20506); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5954). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20506); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5954). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.edit:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.edit:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1937); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5954); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.list`

| Column | Audited content |
| --- | --- |
| Command | `groups.list`; `fbrcm groups list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/groups_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21343). |
| Options | `--cached`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21343); [input schema](../../schemas/cli/1.1.0/groups_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/groups_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21343). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.1.0/groups_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21343); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/groups_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21343); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6084). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21343); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6084). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1896); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6084); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.rename`

| Column | Audited content |
| --- | --- |
| Command | `groups.rename`; `fbrcm groups rename`. |
| Arguments | required group, required new_name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/groups_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21806). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21806); [input schema](../../schemas/cli/1.1.0/groups_rename.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/groups_rename.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21806). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.1.0/groups_rename.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21806); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/groups_rename.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/groups_rename.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6186). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21806); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6186). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21806); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6186). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.rename:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.rename:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1945); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6186); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `help`

| Column | Audited content |
| --- | --- |
| Command | `help`; `fbrcm help`. |
| Arguments | optional command (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/help.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22629). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22629); [input schema](../../schemas/cli/1.1.0/help.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/help.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22629). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.1.0/help.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22629); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/help.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22629); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6316). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22629); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6316). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:help:input`; `urn:fbrcm:schema:cli:1.1.0:command:help:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3954); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `help`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6316); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `hooks.fingerprint`

| Column | Audited content |
| --- | --- |
| Command | `hooks.fingerprint`; `fbrcm hooks fingerprint`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/hooks_fingerprint.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22748). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22748); [input schema](../../schemas/cli/1.1.0/hooks_fingerprint.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/hooks_fingerprint.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.1.0/hooks_fingerprint.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22748); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/hooks_fingerprint.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22748); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6412). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22748); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6412). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:hooks.fingerprint:input`; `urn:fbrcm:schema:cli:1.1.0:command:hooks.fingerprint:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3449); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.fingerprint`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6412); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `hooks.status`

| Column | Audited content |
| --- | --- |
| Command | `hooks.status`; `fbrcm hooks status`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/hooks_status.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22866). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22866); [input schema](../../schemas/cli/1.1.0/hooks_status.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/hooks_status.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.1.0/hooks_status.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22866); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/hooks_status.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22866); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6496). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22866); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6496). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:hooks.status:input`; `urn:fbrcm:schema:cli:1.1.0:command:hooks.status:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3438); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.status`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6496); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `hooks.trust`

| Column | Audited content |
| --- | --- |
| Command | `hooks.trust`; `fbrcm hooks trust`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/hooks_trust.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22983). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22983); [input schema](../../schemas/cli/1.1.0/hooks_trust.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/hooks_trust.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.1.0/hooks_trust.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22983); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/hooks_trust.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22983); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6580). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22983); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6580). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:hooks.trust:input`; `urn:fbrcm:schema:cli:1.1.0:command:hooks.trust:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3455); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.trust`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6580); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `hooks.untrust`

| Column | Audited content |
| --- | --- |
| Command | `hooks.untrust`; `fbrcm hooks untrust`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/hooks_untrust.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23160). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23160); [input schema](../../schemas/cli/1.1.0/hooks_untrust.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/hooks_untrust.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.1.0/hooks_untrust.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23160); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/hooks_untrust.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23160); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6672). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23160); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6672). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:hooks.untrust:input`; `urn:fbrcm:schema:cli:1.1.0:command:hooks.untrust:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3469); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.untrust`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6672); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `mcp`

| Column | Audited content |
| --- | --- |
| Command | `mcp`; `fbrcm mcp`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/mcp.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23306). |
| Options | `--allow-hooks`, `--allow-writes`, `--auth-timeout`, `--browser-auth`, `--confirmation`, `--include-hidden`, `--no-local-config`, `--profile`, `--request-timeout`, `--stateless`, `--timeout`, `--toolsets`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23306); [input schema](../../schemas/cli/1.1.0/mcp.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/mcp.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/mcp.go); [response schema](../../schemas/cli/1.1.0/mcp.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23306); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/mcp.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23306); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6760). |
| Effects | Level 0; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23306); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6760). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:mcp:input`; `urn:fbrcm:schema:cli:1.1.0:command:mcp:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L984); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `mcp`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6760); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `personalizations.list`

| Column | Audited content |
| --- | --- |
| Command | `personalizations.list`; `fbrcm personalizations list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/personalizations_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23473). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23473); [input schema](../../schemas/cli/1.1.0/personalizations_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/personalizations_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.1.0/personalizations_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23473); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/personalizations_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23473); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6849). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23473); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6849). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:personalizations.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:personalizations.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2479); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `personalizations.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6849); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `personalizations.show`

| Column | Audited content |
| --- | --- |
| Command | `personalizations.show`; `fbrcm personalizations show`. |
| Arguments | required project, required personalization_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/personalizations_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23909). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23909); [input schema](../../schemas/cli/1.1.0/personalizations_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/personalizations_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23909). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.1.0/personalizations_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23909); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/personalizations_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23909); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6947). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23909); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6947). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:personalizations.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:personalizations.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2493); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `personalizations.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6947); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `plan.show`

| Column | Audited content |
| --- | --- |
| Command | `plan.show`; `fbrcm plan show`. |
| Arguments | required plan. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/plan_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24352). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24352); [input schema](../../schemas/cli/1.1.0/plan_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/plan_show.input.schema.json). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:publication_plan`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/plan_show.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/plan/commands.go); [response schema](../../schemas/cli/1.1.0/plan_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24352); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/plan_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24352); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7049). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24352); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7049). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:plan.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:plan.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1367); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `plan.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7049); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `plan.validate`

| Column | Audited content |
| --- | --- |
| Command | `plan.validate`; `fbrcm plan validate`. |
| Arguments | required plan. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/plan_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24480). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24480); [input schema](../../schemas/cli/1.1.0/plan_validate.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/plan_validate.input.schema.json). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:publication_plan`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/plan_validate.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/plan/commands.go); [response schema](../../schemas/cli/1.1.0/plan_validate.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24480); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/plan_validate.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24480); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7140). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24480); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7140). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:plan.validate:input`; `urn:fbrcm:schema:cli:1.1.0:command:plan.validate:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1376); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `plan.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7140); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile`

| Column | Audited content |
| --- | --- |
| Command | `profile`; `fbrcm profile`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/profile.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24608). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24608); [input schema](../../schemas/cli/1.1.0/profile.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/profile.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.1.0/profile.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24608); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/profile.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24608); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7231). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24608); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7231). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3673); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7231); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.delete`

| Column | Audited content |
| --- | --- |
| Command | `profile.delete`; `fbrcm profile delete`. |
| Arguments | required profile. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/profile_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24725). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24725); [input schema](../../schemas/cli/1.1.0/profile_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/profile_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24725). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.1.0/profile_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24725); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/profile_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24725); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7311). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24725); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7311). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3717); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7311); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.list`

| Column | Audited content |
| --- | --- |
| Command | `profile.list`; `fbrcm profile list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/profile_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24973). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24973); [input schema](../../schemas/cli/1.1.0/profile_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/profile_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.1.0/profile_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24973); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/profile_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24973); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7406). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24973); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7406). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3677); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7406); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.path`

| Column | Audited content |
| --- | --- |
| Command | `profile.path`; `fbrcm profile path`. |
| Arguments | required profile. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/profile_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25091). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25091); [input schema](../../schemas/cli/1.1.0/profile_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/profile_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.1.0/profile_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25091); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/profile_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25091); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7490). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25091); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7490). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3707); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7490); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.rename`

| Column | Audited content |
| --- | --- |
| Command | `profile.rename`; `fbrcm profile rename`. |
| Arguments | required old_name, required new_name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/profile_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25217). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25217); [input schema](../../schemas/cli/1.1.0/profile_rename.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/profile_rename.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25217). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.1.0/profile_rename.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25217); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/profile_rename.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25217); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7574). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25217); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7574). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.rename:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.rename:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3697); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7574); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.switch`

| Column | Audited content |
| --- | --- |
| Command | `profile.switch`; `fbrcm profile switch`. |
| Arguments | required name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/profile_switch.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25388). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25388); [input schema](../../schemas/cli/1.1.0/profile_switch.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/profile_switch.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.1.0/profile_switch.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25388); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/profile_switch.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25388); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7665). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25388); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7665). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.switch:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.switch:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3687); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.switch`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7665); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.defaults`

| Column | Audited content |
| --- | --- |
| Command | `project.defaults`; `fbrcm project defaults`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_defaults.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25502). |
| Options | `--format`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25502); [input schema](../../schemas/cli/1.1.0/project_defaults.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/project_defaults.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25502). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_defaults.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25502); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_defaults.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25502); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7753). |
| Effects | Level 2; network `required`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25502); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7753). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.defaults:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.defaults:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2193); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.defaults`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7753); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.export`

| Column | Audited content |
| --- | --- |
| Command | `project.export`; `fbrcm project export`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_export.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25841). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25841); [input schema](../../schemas/cli/1.1.0/project_export.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/project_export.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25841). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_export.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25841); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_export.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25841); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7894). |
| Effects | Level 2; network `required`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25841); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7894). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.export:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.export:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2173); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.export`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7894); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.import`

| Column | Audited content |
| --- | --- |
| Command | `project.import`; `fbrcm project import`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26170). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--from`, `--group`, `--include-hidden`, `--keep-portable-conditions-only`, `--merge`, `--merge-resolve`, `--no-local-config`, `--override`, `--plan-out`, `--profile`, `--remove-all-conditions`, `--search`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26170); [input schema](../../schemas/cli/1.1.0/project_import.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/project_import.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26170). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:remote_config_import`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/project_import.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_import.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26170); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_import.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/project_import.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8035). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8035). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8035). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.import:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.import:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2209); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8035); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.open`

| Column | Audited content |
| --- | --- |
| Command | `project.open`; `fbrcm project open`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_open.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27142). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27142); [input schema](../../schemas/cli/1.1.0/project_open.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/project_open.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_open.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27142); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_open.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `browser_launch_suppressed_and_oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8156). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8156). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.open:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.open:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2165); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.open`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8156); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.quota-project.set`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.set`; `fbrcm project quota-project set`. |
| Arguments | required project, required quota_project_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_quota-project_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27420). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27420); [input schema](../../schemas/cli/1.1.0/project_quota-project_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/project_quota-project_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_quota-project_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27420); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_quota-project_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27420); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8254). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27420); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8254). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2155); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8254); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.quota-project.show`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.show`; `fbrcm project quota-project show`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_quota-project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27584). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27584); [input schema](../../schemas/cli/1.1.0/project_quota-project_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/project_quota-project_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_quota-project_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27584); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_quota-project_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27584); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8343). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27584); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8343). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2150); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8343); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.quota-project.unset`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.unset`; `fbrcm project quota-project unset`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_quota-project_unset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27761). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27761); [input schema](../../schemas/cli/1.1.0/project_quota-project_unset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/project_quota-project_unset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_quota-project_unset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27761); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_quota-project_unset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27761); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8427). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27761); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8427). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.unset:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.unset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2159); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.unset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8427); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.show`

| Column | Audited content |
| --- | --- |
| Command | `project.show`; `fbrcm project show`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27954). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27954); [input schema](../../schemas/cli/1.1.0/project_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/project_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27954); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27954); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8515). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27954); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8515). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2087); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8515); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.templates.set`

| Column | Audited content |
| --- | --- |
| Command | `project.templates.set`; `fbrcm project templates set`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_templates_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28318). |
| Options | `--include-hidden`, `--no-local-config`, `--primary`, `--profile`, `--stateless`, `--templates`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28318); [input schema](../../schemas/cli/1.1.0/project_templates_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/project_templates_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_templates_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28318); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_templates_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28318); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8613). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28318); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8613). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.templates.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.templates.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2118); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.templates.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8613); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.templates.show`

| Column | Audited content |
| --- | --- |
| Command | `project.templates.show`; `fbrcm project templates show`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/project_templates_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28496). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28496); [input schema](../../schemas/cli/1.1.0/project_templates_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/project_templates_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.1.0/project_templates_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28496); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/project_templates_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28496); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8697). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28496); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8697). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.templates.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.templates.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2108); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.templates.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8697); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.aliases.import`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.import`; `fbrcm projects aliases import`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_aliases_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28638). |
| Options | `--conflict`, `--dry-run`, `--from`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28638); [input schema](../../schemas/cli/1.1.0/projects_aliases_import.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_aliases_import.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_aliases_import.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28638); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_aliases_import.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28638); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8790). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28638); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8790). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.import:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.import:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3059); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8790); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.aliases.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.list`; `fbrcm projects aliases list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_aliases_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28867). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28867); [input schema](../../schemas/cli/1.1.0/projects_aliases_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_aliases_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_aliases_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28867); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_aliases_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28867); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8882). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28867); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8882). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3043); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8882); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.aliases.remove`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.remove`; `fbrcm projects aliases remove`. |
| Arguments | required alias. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_aliases_remove.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28987). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28987); [input schema](../../schemas/cli/1.1.0/projects_aliases_remove.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_aliases_remove.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28987). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_aliases_remove.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28987); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_aliases_remove.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28987); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8974). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28987); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8974). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.remove:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.remove:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3077); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8974); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.aliases.set`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.set`; `fbrcm projects aliases set`. |
| Arguments | required alias, required project_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_aliases_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29187). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29187); [input schema](../../schemas/cli/1.1.0/projects_aliases_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_aliases_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_aliases_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29187); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_aliases_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29187); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9069). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29187); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9069). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3058); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9069); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.add`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.add`; `fbrcm projects blacklist add`. |
| Arguments | optional project_id (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_blacklist_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29393). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29393); [input schema](../../schemas/cli/1.1.0/projects_blacklist_add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_blacklist_add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29393). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_blacklist_add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29393); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_blacklist_add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29393); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9161). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29393); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9161). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.add:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2882); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9161); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.clear`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.clear`; `fbrcm projects blacklist clear`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_blacklist_clear.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29548). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29548); [input schema](../../schemas/cli/1.1.0/projects_blacklist_clear.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_blacklist_clear.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_blacklist_clear.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29548); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_blacklist_clear.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29548); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9282). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29548); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9282). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.clear:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.clear:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2890); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9282); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.disable`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.disable`; `fbrcm projects blacklist disable`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_blacklist_disable.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29684). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29684); [input schema](../../schemas/cli/1.1.0/projects_blacklist_disable.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_blacklist_disable.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_blacklist_disable.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29684); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_blacklist_disable.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29684); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9397). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29684); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9397). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.disable:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.disable:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2898); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.disable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9397); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.enable`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.enable`; `fbrcm projects blacklist enable`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_blacklist_enable.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29820). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29820); [input schema](../../schemas/cli/1.1.0/projects_blacklist_enable.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_blacklist_enable.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_blacklist_enable.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29820); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_blacklist_enable.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29820); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9512). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29820); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9512). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.enable:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.enable:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2894); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.enable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9512); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.list`; `fbrcm projects blacklist list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_blacklist_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29956). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29956); [input schema](../../schemas/cli/1.1.0/projects_blacklist_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_blacklist_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_blacklist_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29956); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_blacklist_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29956); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9627). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29956); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9627). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2878); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9627); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.remove`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.remove`; `fbrcm projects blacklist remove`. |
| Arguments | optional project_id (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_blacklist_remove.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30076). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30076); [input schema](../../schemas/cli/1.1.0/projects_blacklist_remove.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_blacklist_remove.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30076). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_blacklist_remove.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30076); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_blacklist_remove.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30076); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9735). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30076); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9735). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.remove:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.remove:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2886); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9735); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.diff`

| Column | Audited content |
| --- | --- |
| Command | `projects.diff`; `fbrcm projects diff`. |
| Arguments | required source_project, required target_project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30231). |
| Options | `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30231); [input schema](../../schemas/cli/1.1.0/projects_diff.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_diff.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30231). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_diff.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30231); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_diff.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30231); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9856). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30231); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9856). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.diff:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.diff:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2975); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9856); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.forget`

| Column | Audited content |
| --- | --- |
| Command | `projects.forget`; `fbrcm projects forget`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_forget.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30601). |
| Options | `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30601); [input schema](../../schemas/cli/1.1.0/projects_forget.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_forget.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30601). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_forget.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30601); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_forget.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30601); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9958). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30601); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9958). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.forget:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.forget:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2963); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.forget`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9958); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.list`; `fbrcm projects list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30859). |
| Options | `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`, `--url`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30859); [input schema](../../schemas/cli/1.1.0/projects_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30859). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30859); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30859); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10054). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30859); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10054). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2926); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10054); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.path`

| Column | Audited content |
| --- | --- |
| Command | `projects.path`; `fbrcm projects path`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31247). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31247); [input schema](../../schemas/cli/1.1.0/projects_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31247); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31247); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10186). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31247); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10186). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3118); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10186); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.promote`

| Column | Audited content |
| --- | --- |
| Command | `projects.promote`; `fbrcm projects promote`. |
| Arguments | required source_project, required target_project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_promote.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31366). |
| Options | `--all`, `--change-note`, `--conditions`, `--dry-run`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--interactive`, `--no-local-config`, `--parameters`, `--plan-out`, `--profile`, `--prune`, `--search`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31366); [input schema](../../schemas/cli/1.1.0/projects_promote.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_promote.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31366). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_promote.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31366); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_promote.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/projects_promote.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10270). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31366); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10270). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31366); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10270). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.promote:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.promote:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3002); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.promote`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10270); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.reset`

| Column | Audited content |
| --- | --- |
| Command | `projects.reset`; `fbrcm projects reset`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32180). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32180); [input schema](../../schemas/cli/1.1.0/projects_reset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_reset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_reset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32180); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_reset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32180); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10386). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32180); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10386). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.reset:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.reset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3128); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10386); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.update`

| Column | Audited content |
| --- | --- |
| Command | `projects.update`; `fbrcm projects update`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_update.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32392). |
| Options | `--auth`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--url`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32392); [input schema](../../schemas/cli/1.1.0/projects_update.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_update.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32392). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.1.0/projects_update.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32392); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_update.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32392); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10478). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32392); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10478). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.update:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.update:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2947); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.update`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10478); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.add`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.add`; `fbrcm projects whitelist add`. |
| Arguments | optional project_id (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_whitelist_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32640). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32640); [input schema](../../schemas/cli/1.1.0/projects_whitelist_add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_whitelist_add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32640). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_whitelist_add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32640); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_whitelist_add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `project.remains_hidden`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/projects_whitelist_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10572). |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32640); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10572). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32640); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10572). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.add:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2906); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10572); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.clear`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.clear`; `fbrcm projects whitelist clear`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_whitelist_clear.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32795). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32795); [input schema](../../schemas/cli/1.1.0/projects_whitelist_clear.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_whitelist_clear.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_whitelist_clear.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32795); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_whitelist_clear.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32795); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10700). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32795); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10700). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.clear:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.clear:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2914); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10700); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.disable`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.disable`; `fbrcm projects whitelist disable`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_whitelist_disable.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32931). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32931); [input schema](../../schemas/cli/1.1.0/projects_whitelist_disable.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_whitelist_disable.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_whitelist_disable.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32931); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_whitelist_disable.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32931); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10815). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32931); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10815). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.disable:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.disable:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2922); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.disable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10815); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.enable`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.enable`; `fbrcm projects whitelist enable`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_whitelist_enable.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33067). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33067); [input schema](../../schemas/cli/1.1.0/projects_whitelist_enable.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_whitelist_enable.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_whitelist_enable.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33067); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_whitelist_enable.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33067); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10930). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33067); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10930). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.enable:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.enable:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2918); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.enable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10930); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.list`; `fbrcm projects whitelist list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_whitelist_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33203). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33203); [input schema](../../schemas/cli/1.1.0/projects_whitelist_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/projects_whitelist_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_whitelist_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33203); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_whitelist_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33203); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11045). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33203); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11045). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2902); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11045); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.remove`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.remove`; `fbrcm projects whitelist remove`. |
| Arguments | optional project_id (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/projects_whitelist_remove.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33323). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33323); [input schema](../../schemas/cli/1.1.0/projects_whitelist_remove.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/projects_whitelist_remove.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33323). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.1.0/projects_whitelist_remove.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33323); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/projects_whitelist_remove.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33323); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11153). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33323); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11153). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.remove:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.remove:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2910); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11153); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `rollouts.delete`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.delete`; `fbrcm rollouts delete`. |
| Arguments | required project, required rollout_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/rollouts_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33478). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33478); [input schema](../../schemas/cli/1.1.0/rollouts_delete.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/rollouts_delete.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.1.0/rollouts_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33478); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/rollouts_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33478); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11274). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33478); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11274). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:rollouts.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2463); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11274); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `rollouts.list`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.list`; `fbrcm rollouts list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/rollouts_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33849). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33849); [input schema](../../schemas/cli/1.1.0/rollouts_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/rollouts_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.1.0/rollouts_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33849); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/rollouts_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33849); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11368). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33849); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11368). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:rollouts.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2429); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11368); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `rollouts.show`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.show`; `fbrcm rollouts show`. |
| Arguments | required project, required rollout_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/rollouts_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34199). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34199); [input schema](../../schemas/cli/1.1.0/rollouts_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/rollouts_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.1.0/rollouts_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34199); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/rollouts_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34199); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11466). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34199); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11466). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:rollouts.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2446); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11466); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `root`

| Column | Audited content |
| --- | --- |
| Command | `root`; `fbrcm`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/root.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34555). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--version`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34555); [input schema](../../schemas/cli/1.1.0/root.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/root.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.1.0/root.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34555); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/root.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34555); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11564). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34555); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11564). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:root:input`; `urn:fbrcm:schema:cli:1.1.0:command:root:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1200); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `root`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11564); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `schema.list`

| Column | Audited content |
| --- | --- |
| Command | `schema.list`; `fbrcm schema list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/schema_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34708). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34708); [input schema](../../schemas/cli/1.1.0/schema_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/schema_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.1.0/schema_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34708); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/schema_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34708); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11658). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34708); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11658). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:schema.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:schema.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1329); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `schema.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11658); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `schema.show`

| Column | Audited content |
| --- | --- |
| Command | `schema.show`; `fbrcm schema show`. |
| Arguments | required schema_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/schema_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34821). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34821); [input schema](../../schemas/cli/1.1.0/schema_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/schema_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34821). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.1.0/schema_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34821); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/schema_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34821); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11742). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34821); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11742). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:schema.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:schema.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1334); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `schema.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11742); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `setup`

| Column | Audited content |
| --- | --- |
| Command | `setup`; `fbrcm setup`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/setup.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34942). |
| Options | `--include-hidden`, `--no-local-config`, `--noopen`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34942); [input schema](../../schemas/cli/1.1.0/setup.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/setup.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/setup/commands.go); [response schema](../../schemas/cli/1.1.0/setup.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34942); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/setup.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `required`; `guided_setup_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34942); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11838). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34942); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11838). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:setup:input`; `urn:fbrcm:schema:cli:1.1.0:command:setup:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3140); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `setup`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11838); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme`

| Column | Audited content |
| --- | --- |
| Command | `theme`; `fbrcm theme`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/theme.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35082). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35082); [input schema](../../schemas/cli/1.1.0/theme.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/theme.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.1.0/theme.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35082); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/theme.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35082); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11929). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35082); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11929). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3729); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11929); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.delete`

| Column | Audited content |
| --- | --- |
| Command | `theme.delete`; `fbrcm theme delete`. |
| Arguments | required theme. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/theme_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35200). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35200); [input schema](../../schemas/cli/1.1.0/theme_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/theme_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35200). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.1.0/theme_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35200); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/theme_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35200); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12013). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35200); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12013). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3791); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12013); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.import`

| Column | Audited content |
| --- | --- |
| Command | `theme.import`; `fbrcm theme import`. |
| Arguments | optional source. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/theme_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35404). |
| Options | `--include-hidden`, `--name`, `--no-local-config`, `--profile`, `--ref`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35404); [input schema](../../schemas/cli/1.1.0/theme_import.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/theme_import.input.schema.json). |
| Stdin | `toml_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:theme`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/theme_import.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.1.0/theme_import.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35404); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/theme_import.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `theme.already_exists`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/theme_import.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12098). |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35404); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12098). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35404); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12098). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.import:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.import:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3803); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12098); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.list`

| Column | Audited content |
| --- | --- |
| Command | `theme.list`; `fbrcm theme list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/theme_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35639). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35639); [input schema](../../schemas/cli/1.1.0/theme_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/theme_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.1.0/theme_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35639); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/theme_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12242). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12242). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3738); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12242); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.path`

| Column | Audited content |
| --- | --- |
| Command | `theme.path`; `fbrcm theme path`. |
| Arguments | required theme. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/theme_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35756). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35756); [input schema](../../schemas/cli/1.1.0/theme_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/theme_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.1.0/theme_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35756); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/theme_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35756); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12326). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35756); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12326). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3778); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12326); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.rename`

| Column | Audited content |
| --- | --- |
| Command | `theme.rename`; `fbrcm theme rename`. |
| Arguments | required old_name, required new_name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/theme_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35880). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35880); [input schema](../../schemas/cli/1.1.0/theme_rename.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/theme_rename.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35880). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.1.0/theme_rename.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35880); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/theme_rename.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35880); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12410). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35880); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12410). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.rename:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.rename:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3783); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12410); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.reset`

| Column | Audited content |
| --- | --- |
| Command | `theme.reset`; `fbrcm theme reset`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/theme_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36045). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36045); [input schema](../../schemas/cli/1.1.0/theme_reset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/theme_reset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.1.0/theme_reset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36045); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/theme_reset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36045); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12501). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36045); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12501). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.reset:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.reset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3764); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12501); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.switch`

| Column | Audited content |
| --- | --- |
| Command | `theme.switch`; `fbrcm theme switch`. |
| Arguments | required name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/theme_switch.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36201). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36201); [input schema](../../schemas/cli/1.1.0/theme_switch.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/theme_switch.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36201). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.1.0/theme_switch.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36201); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/theme_switch.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36201); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12581). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36201); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12581). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.switch:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.switch:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3749); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.switch`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12581); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `update`

| Column | Audited content |
| --- | --- |
| Command | `update`; `fbrcm update`. |
| Arguments | optional parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/update.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36372). |
| Options | `--change-note`, `--condition`, `--description`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--name`, `--no-group`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--remove-all-conditional-values`, `--remove-conditional-value`, `--search`, `--stateless`, `--timeout`, `--type`, `--use-in-app-default`, `--value`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36372); [input schema](../../schemas/cli/1.1.0/update.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/update.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36372). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.1.0:stdin:remote_config`; precedence and option restrictions: [input schema](../../schemas/cli/1.1.0/update.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/update/commands.go); [response schema](../../schemas/cli/1.1.0/update.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36372); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/update.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/update.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12663). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36372); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12663). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36372); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12663). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:update:input`; `urn:fbrcm:schema:cli:1.1.0:command:update:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1638); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `update`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12663); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.blame`

| Column | Audited content |
| --- | --- |
| Command | `versions.blame`; `fbrcm versions blame`. |
| Arguments | required project, required parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_blame.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37423). |
| Options | `--all`, `--at`, `--include-hidden`, `--limit`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37423); [input schema](../../schemas/cli/1.1.0/versions_blame.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_blame.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37423). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_blame.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37423); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_blame.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37423); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12814). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37423); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12814). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.blame:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.blame:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2556); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.blame`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12814); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.diff`

| Column | Audited content |
| --- | --- |
| Command | `versions.diff`; `fbrcm versions diff`. |
| Arguments | required project, required from, optional to. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37750). |
| Options | `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--side-by-side`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37750); [input schema](../../schemas/cli/1.1.0/versions_diff.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_diff.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37750). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_diff.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37750); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_diff.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37750); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12924). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37750); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12924). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.diff:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.diff:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2695); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12924); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.export`

| Column | Audited content |
| --- | --- |
| Command | `versions.export`; `fbrcm versions export`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_export.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38226). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38226); [input schema](../../schemas/cli/1.1.0/versions_export.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_export.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38226). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_export.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38226); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_export.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38226); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13036). |
| Effects | Level 2; network `conditional`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38226); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13036). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.export:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.export:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2723); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.export`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13036); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.list`

| Column | Audited content |
| --- | --- |
| Command | `versions.list`; `fbrcm versions list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38696). |
| Options | `--all`, `--before`, `--cached`, `--include-hidden`, `--limit`, `--no-local-config`, `--profile`, `--since`, `--stateless`, `--timeout`, `--until`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38696); [input schema](../../schemas/cli/1.1.0/versions_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38696). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38696); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38696); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13144). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38696); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13144). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2650); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13144); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.restore`

| Column | Audited content |
| --- | --- |
| Command | `versions.restore`; `fbrcm versions restore`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_restore.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39047). |
| Options | `--change-note`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39047); [input schema](../../schemas/cli/1.1.0/versions_restore.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_restore.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39047). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_restore.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39047); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_restore.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/versions_restore.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13246). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39047); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13246). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39047); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13246). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.restore:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.restore:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2778); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.restore`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13246); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.rollback`

| Column | Audited content |
| --- | --- |
| Command | `versions.rollback`; `fbrcm versions rollback`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_rollback.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39698). |
| Options | `--dry-run`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39698); [input schema](../../schemas/cli/1.1.0/versions_rollback.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_rollback.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39698). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_rollback.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39698); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_rollback.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/versions_rollback.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13347). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39698); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13347). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39698); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13347). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.rollback:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.rollback:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2739); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.rollback`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13347); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.show`

| Column | Audited content |
| --- | --- |
| Command | `versions.show`; `fbrcm versions show`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40300). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40300); [input schema](../../schemas/cli/1.1.0/versions_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40300). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40300); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40300); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13452). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40300); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13452). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2680); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13452); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

## Appendix B — complete schema file and ID inventory

| Schema file | `$id` |
| --- | --- |
| [1.0.0/add.input.schema.json](../../schemas/cli/1.0.0/add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:add:input` |
| [1.0.0/add.response.schema.json](../../schemas/cli/1.0.0/add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:add:response` |
| [1.0.0/apply.input.schema.json](../../schemas/cli/1.0.0/apply.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apply:input` |
| [1.0.0/apply.response.schema.json](../../schemas/cli/1.0.0/apply.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apply:response` |
| [1.0.0/apps_config.input.schema.json](../../schemas/cli/1.0.0/apps_config.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.config:input` |
| [1.0.0/apps_config.response.schema.json](../../schemas/cli/1.0.0/apps_config.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.config:response` |
| [1.0.0/apps_list.input.schema.json](../../schemas/cli/1.0.0/apps_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.list:input` |
| [1.0.0/apps_list.response.schema.json](../../schemas/cli/1.0.0/apps_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.list:response` |
| [1.0.0/apps_show.input.schema.json](../../schemas/cli/1.0.0/apps_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.show:input` |
| [1.0.0/apps_show.response.schema.json](../../schemas/cli/1.0.0/apps_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.show:response` |
| [1.0.0/auth_add_gcloud.input.schema.json](../../schemas/cli/1.0.0/auth_add_gcloud.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:input` |
| [1.0.0/auth_add_gcloud.response.schema.json](../../schemas/cli/1.0.0/auth_add_gcloud.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:response` |
| [1.0.0/auth_add_google.input.schema.json](../../schemas/cli/1.0.0/auth_add_google.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:input` |
| [1.0.0/auth_add_google.response.schema.json](../../schemas/cli/1.0.0/auth_add_google.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:response` |
| [1.0.0/auth_add_oauth.input.schema.json](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:input` |
| [1.0.0/auth_add_oauth.response.schema.json](../../schemas/cli/1.0.0/auth_add_oauth.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:response` |
| [1.0.0/auth_add_service-account.input.schema.json](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:input` |
| [1.0.0/auth_add_service-account.response.schema.json](../../schemas/cli/1.0.0/auth_add_service-account.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:response` |
| [1.0.0/auth_bind.input.schema.json](../../schemas/cli/1.0.0/auth_bind.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:input` |
| [1.0.0/auth_bind.response.schema.json](../../schemas/cli/1.0.0/auth_bind.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:response` |
| [1.0.0/auth_delete.input.schema.json](../../schemas/cli/1.0.0/auth_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:input` |
| [1.0.0/auth_delete.response.schema.json](../../schemas/cli/1.0.0/auth_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:response` |
| [1.0.0/auth_list.input.schema.json](../../schemas/cli/1.0.0/auth_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.list:input` |
| [1.0.0/auth_list.response.schema.json](../../schemas/cli/1.0.0/auth_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.list:response` |
| [1.0.0/auth_login.input.schema.json](../../schemas/cli/1.0.0/auth_login.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.login:input` |
| [1.0.0/auth_login.response.schema.json](../../schemas/cli/1.0.0/auth_login.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.login:response` |
| [1.0.0/auth_path.input.schema.json](../../schemas/cli/1.0.0/auth_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.path:input` |
| [1.0.0/auth_path.response.schema.json](../../schemas/cli/1.0.0/auth_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.path:response` |
| [1.0.0/auth_quota-project_set.input.schema.json](../../schemas/cli/1.0.0/auth_quota-project_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:input` |
| [1.0.0/auth_quota-project_set.response.schema.json](../../schemas/cli/1.0.0/auth_quota-project_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:response` |
| [1.0.0/auth_quota-project_show.input.schema.json](../../schemas/cli/1.0.0/auth_quota-project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:input` |
| [1.0.0/auth_quota-project_show.response.schema.json](../../schemas/cli/1.0.0/auth_quota-project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:response` |
| [1.0.0/auth_quota-project_unset.input.schema.json](../../schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:input` |
| [1.0.0/auth_quota-project_unset.response.schema.json](../../schemas/cli/1.0.0/auth_quota-project_unset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:response` |
| [1.0.0/cache_clear.input.schema.json](../../schemas/cli/1.0.0/cache_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:input` |
| [1.0.0/cache_clear.response.schema.json](../../schemas/cli/1.0.0/cache_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:response` |
| [1.0.0/cache_list.input.schema.json](../../schemas/cli/1.0.0/cache_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.list:input` |
| [1.0.0/cache_list.response.schema.json](../../schemas/cli/1.0.0/cache_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.list:response` |
| [1.0.0/cache_path.input.schema.json](../../schemas/cli/1.0.0/cache_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.path:input` |
| [1.0.0/cache_path.response.schema.json](../../schemas/cli/1.0.0/cache_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.path:response` |
| [1.0.0/capabilities.input.schema.json](../../schemas/cli/1.0.0/capabilities.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:capabilities:input` |
| [1.0.0/capabilities.response.schema.json](../../schemas/cli/1.0.0/capabilities.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:capabilities:response` |
| [1.0.0/capability.schema.json](../../schemas/cli/1.0.0/capability.schema.json) | `urn:fbrcm:schema:cli:1.0.0:capability` |
| [1.0.0/completion_bash.input.schema.json](../../schemas/cli/1.0.0/completion_bash.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:input` |
| [1.0.0/completion_bash.response.schema.json](../../schemas/cli/1.0.0/completion_bash.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:response` |
| [1.0.0/completion_fish.input.schema.json](../../schemas/cli/1.0.0/completion_fish.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:input` |
| [1.0.0/completion_fish.response.schema.json](../../schemas/cli/1.0.0/completion_fish.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:response` |
| [1.0.0/completion_powershell.input.schema.json](../../schemas/cli/1.0.0/completion_powershell.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:input` |
| [1.0.0/completion_powershell.response.schema.json](../../schemas/cli/1.0.0/completion_powershell.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:response` |
| [1.0.0/completion_zsh.input.schema.json](../../schemas/cli/1.0.0/completion_zsh.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:input` |
| [1.0.0/completion_zsh.response.schema.json](../../schemas/cli/1.0.0/completion_zsh.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:response` |
| [1.0.0/conditions_add.input.schema.json](../../schemas/cli/1.0.0/conditions_add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:input` |
| [1.0.0/conditions_add.response.schema.json](../../schemas/cli/1.0.0/conditions_add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:response` |
| [1.0.0/conditions_delete.input.schema.json](../../schemas/cli/1.0.0/conditions_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:input` |
| [1.0.0/conditions_delete.response.schema.json](../../schemas/cli/1.0.0/conditions_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:response` |
| [1.0.0/conditions_edit.input.schema.json](../../schemas/cli/1.0.0/conditions_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:input` |
| [1.0.0/conditions_edit.response.schema.json](../../schemas/cli/1.0.0/conditions_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:response` |
| [1.0.0/conditions_list.input.schema.json](../../schemas/cli/1.0.0/conditions_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:input` |
| [1.0.0/conditions_list.response.schema.json](../../schemas/cli/1.0.0/conditions_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:response` |
| [1.0.0/conditions_move.input.schema.json](../../schemas/cli/1.0.0/conditions_move.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:input` |
| [1.0.0/conditions_move.response.schema.json](../../schemas/cli/1.0.0/conditions_move.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:response` |
| [1.0.0/conditions_rename.input.schema.json](../../schemas/cli/1.0.0/conditions_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:input` |
| [1.0.0/conditions_rename.response.schema.json](../../schemas/cli/1.0.0/conditions_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:response` |
| [1.0.0/conditions_show.input.schema.json](../../schemas/cli/1.0.0/conditions_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:input` |
| [1.0.0/conditions_show.response.schema.json](../../schemas/cli/1.0.0/conditions_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:response` |
| [1.0.0/conditions_validate.input.schema.json](../../schemas/cli/1.0.0/conditions_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:input` |
| [1.0.0/conditions_validate.response.schema.json](../../schemas/cli/1.0.0/conditions_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:response` |
| [1.0.0/config_edit.input.schema.json](../../schemas/cli/1.0.0/config_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.edit:input` |
| [1.0.0/config_edit.response.schema.json](../../schemas/cli/1.0.0/config_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.edit:response` |
| [1.0.0/config_path.input.schema.json](../../schemas/cli/1.0.0/config_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.path:input` |
| [1.0.0/config_path.response.schema.json](../../schemas/cli/1.0.0/config_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.path:response` |
| [1.0.0/config_reset.input.schema.json](../../schemas/cli/1.0.0/config_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.reset:input` |
| [1.0.0/config_reset.response.schema.json](../../schemas/cli/1.0.0/config_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.reset:response` |
| [1.0.0/config_set.input.schema.json](../../schemas/cli/1.0.0/config_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.set:input` |
| [1.0.0/config_set.response.schema.json](../../schemas/cli/1.0.0/config_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.set:response` |
| [1.0.0/config_show.input.schema.json](../../schemas/cli/1.0.0/config_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.show:input` |
| [1.0.0/config_show.response.schema.json](../../schemas/cli/1.0.0/config_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.show:response` |
| [1.0.0/config_validate.input.schema.json](../../schemas/cli/1.0.0/config_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.validate:input` |
| [1.0.0/config_validate.response.schema.json](../../schemas/cli/1.0.0/config_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.validate:response` |
| [1.0.0/delete.input.schema.json](../../schemas/cli/1.0.0/delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:delete:input` |
| [1.0.0/delete.response.schema.json](../../schemas/cli/1.0.0/delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:delete:response` |
| [1.0.0/doctor.input.schema.json](../../schemas/cli/1.0.0/doctor.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:doctor:input` |
| [1.0.0/doctor.response.schema.json](../../schemas/cli/1.0.0/doctor.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:doctor:response` |
| [1.0.0/draft_change-note.input.schema.json](../../schemas/cli/1.0.0/draft_change-note.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:input` |
| [1.0.0/draft_change-note.response.schema.json](../../schemas/cli/1.0.0/draft_change-note.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:response` |
| [1.0.0/draft_diff.input.schema.json](../../schemas/cli/1.0.0/draft_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:input` |
| [1.0.0/draft_diff.response.schema.json](../../schemas/cli/1.0.0/draft_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:response` |
| [1.0.0/draft_discard.input.schema.json](../../schemas/cli/1.0.0/draft_discard.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:input` |
| [1.0.0/draft_discard.response.schema.json](../../schemas/cli/1.0.0/draft_discard.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:response` |
| [1.0.0/draft_list.input.schema.json](../../schemas/cli/1.0.0/draft_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.list:input` |
| [1.0.0/draft_list.response.schema.json](../../schemas/cli/1.0.0/draft_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.list:response` |
| [1.0.0/draft_path.input.schema.json](../../schemas/cli/1.0.0/draft_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.path:input` |
| [1.0.0/draft_path.response.schema.json](../../schemas/cli/1.0.0/draft_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.path:response` |
| [1.0.0/draft_publish.input.schema.json](../../schemas/cli/1.0.0/draft_publish.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:input` |
| [1.0.0/draft_publish.response.schema.json](../../schemas/cli/1.0.0/draft_publish.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:response` |
| [1.0.0/draft_show.input.schema.json](../../schemas/cli/1.0.0/draft_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.show:input` |
| [1.0.0/draft_show.response.schema.json](../../schemas/cli/1.0.0/draft_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.show:response` |
| [1.0.0/duplicate.input.schema.json](../../schemas/cli/1.0.0/duplicate.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:duplicate:input` |
| [1.0.0/duplicate.response.schema.json](../../schemas/cli/1.0.0/duplicate.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:duplicate:response` |
| [1.0.0/envelope.schema.json](../../schemas/cli/1.0.0/envelope.schema.json) | `urn:fbrcm:schema:cli:1.0.0:envelope` |
| [1.0.0/error.schema.json](../../schemas/cli/1.0.0/error.schema.json) | `urn:fbrcm:schema:cli:1.0.0:error` |
| [1.0.0/experiments_delete.input.schema.json](../../schemas/cli/1.0.0/experiments_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:input` |
| [1.0.0/experiments_delete.response.schema.json](../../schemas/cli/1.0.0/experiments_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:response` |
| [1.0.0/experiments_list.input.schema.json](../../schemas/cli/1.0.0/experiments_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:input` |
| [1.0.0/experiments_list.response.schema.json](../../schemas/cli/1.0.0/experiments_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:response` |
| [1.0.0/experiments_show.input.schema.json](../../schemas/cli/1.0.0/experiments_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:input` |
| [1.0.0/experiments_show.response.schema.json](../../schemas/cli/1.0.0/experiments_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:response` |
| [1.0.0/get.input.schema.json](../../schemas/cli/1.0.0/get.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:get:input` |
| [1.0.0/get.response.schema.json](../../schemas/cli/1.0.0/get.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:get:response` |
| [1.0.0/groups_add.input.schema.json](../../schemas/cli/1.0.0/groups_add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.add:input` |
| [1.0.0/groups_add.response.schema.json](../../schemas/cli/1.0.0/groups_add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.add:response` |
| [1.0.0/groups_delete.input.schema.json](../../schemas/cli/1.0.0/groups_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:input` |
| [1.0.0/groups_delete.response.schema.json](../../schemas/cli/1.0.0/groups_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:response` |
| [1.0.0/groups_edit.input.schema.json](../../schemas/cli/1.0.0/groups_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:input` |
| [1.0.0/groups_edit.response.schema.json](../../schemas/cli/1.0.0/groups_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:response` |
| [1.0.0/groups_list.input.schema.json](../../schemas/cli/1.0.0/groups_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.list:input` |
| [1.0.0/groups_list.response.schema.json](../../schemas/cli/1.0.0/groups_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.list:response` |
| [1.0.0/groups_rename.input.schema.json](../../schemas/cli/1.0.0/groups_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:input` |
| [1.0.0/groups_rename.response.schema.json](../../schemas/cli/1.0.0/groups_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:response` |
| [1.0.0/help.input.schema.json](../../schemas/cli/1.0.0/help.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:help:input` |
| [1.0.0/help.response.schema.json](../../schemas/cli/1.0.0/help.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:help:response` |
| [1.0.0/hooks_fingerprint.input.schema.json](../../schemas/cli/1.0.0/hooks_fingerprint.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:input` |
| [1.0.0/hooks_fingerprint.response.schema.json](../../schemas/cli/1.0.0/hooks_fingerprint.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:response` |
| [1.0.0/hooks_status.input.schema.json](../../schemas/cli/1.0.0/hooks_status.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:input` |
| [1.0.0/hooks_status.response.schema.json](../../schemas/cli/1.0.0/hooks_status.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:response` |
| [1.0.0/hooks_trust.input.schema.json](../../schemas/cli/1.0.0/hooks_trust.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:input` |
| [1.0.0/hooks_trust.response.schema.json](../../schemas/cli/1.0.0/hooks_trust.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:response` |
| [1.0.0/hooks_untrust.input.schema.json](../../schemas/cli/1.0.0/hooks_untrust.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:input` |
| [1.0.0/hooks_untrust.response.schema.json](../../schemas/cli/1.0.0/hooks_untrust.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:response` |
| [1.0.0/mcp.input.schema.json](../../schemas/cli/1.0.0/mcp.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:mcp:input` |
| [1.0.0/mcp.response.schema.json](../../schemas/cli/1.0.0/mcp.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:mcp:response` |
| [1.0.0/personalizations_list.input.schema.json](../../schemas/cli/1.0.0/personalizations_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:input` |
| [1.0.0/personalizations_list.response.schema.json](../../schemas/cli/1.0.0/personalizations_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:response` |
| [1.0.0/personalizations_show.input.schema.json](../../schemas/cli/1.0.0/personalizations_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:input` |
| [1.0.0/personalizations_show.response.schema.json](../../schemas/cli/1.0.0/personalizations_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:response` |
| [1.0.0/plan_show.input.schema.json](../../schemas/cli/1.0.0/plan_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:plan.show:input` |
| [1.0.0/plan_show.response.schema.json](../../schemas/cli/1.0.0/plan_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:plan.show:response` |
| [1.0.0/plan_validate.input.schema.json](../../schemas/cli/1.0.0/plan_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:plan.validate:input` |
| [1.0.0/plan_validate.response.schema.json](../../schemas/cli/1.0.0/plan_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:plan.validate:response` |
| [1.0.0/profile.input.schema.json](../../schemas/cli/1.0.0/profile.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile:input` |
| [1.0.0/profile.response.schema.json](../../schemas/cli/1.0.0/profile.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile:response` |
| [1.0.0/profile_delete.input.schema.json](../../schemas/cli/1.0.0/profile_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:input` |
| [1.0.0/profile_delete.response.schema.json](../../schemas/cli/1.0.0/profile_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:response` |
| [1.0.0/profile_list.input.schema.json](../../schemas/cli/1.0.0/profile_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.list:input` |
| [1.0.0/profile_list.response.schema.json](../../schemas/cli/1.0.0/profile_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.list:response` |
| [1.0.0/profile_path.input.schema.json](../../schemas/cli/1.0.0/profile_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.path:input` |
| [1.0.0/profile_path.response.schema.json](../../schemas/cli/1.0.0/profile_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.path:response` |
| [1.0.0/profile_rename.input.schema.json](../../schemas/cli/1.0.0/profile_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:input` |
| [1.0.0/profile_rename.response.schema.json](../../schemas/cli/1.0.0/profile_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:response` |
| [1.0.0/profile_switch.input.schema.json](../../schemas/cli/1.0.0/profile_switch.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:input` |
| [1.0.0/profile_switch.response.schema.json](../../schemas/cli/1.0.0/profile_switch.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:response` |
| [1.0.0/project_defaults.input.schema.json](../../schemas/cli/1.0.0/project_defaults.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:input` |
| [1.0.0/project_defaults.response.schema.json](../../schemas/cli/1.0.0/project_defaults.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:response` |
| [1.0.0/project_export.input.schema.json](../../schemas/cli/1.0.0/project_export.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.export:input` |
| [1.0.0/project_export.response.schema.json](../../schemas/cli/1.0.0/project_export.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.export:response` |
| [1.0.0/project_import.input.schema.json](../../schemas/cli/1.0.0/project_import.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.import:input` |
| [1.0.0/project_import.response.schema.json](../../schemas/cli/1.0.0/project_import.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.import:response` |
| [1.0.0/project_open.input.schema.json](../../schemas/cli/1.0.0/project_open.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.open:input` |
| [1.0.0/project_open.response.schema.json](../../schemas/cli/1.0.0/project_open.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.open:response` |
| [1.0.0/project_quota-project_set.input.schema.json](../../schemas/cli/1.0.0/project_quota-project_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:input` |
| [1.0.0/project_quota-project_set.response.schema.json](../../schemas/cli/1.0.0/project_quota-project_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:response` |
| [1.0.0/project_quota-project_show.input.schema.json](../../schemas/cli/1.0.0/project_quota-project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:input` |
| [1.0.0/project_quota-project_show.response.schema.json](../../schemas/cli/1.0.0/project_quota-project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:response` |
| [1.0.0/project_quota-project_unset.input.schema.json](../../schemas/cli/1.0.0/project_quota-project_unset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:input` |
| [1.0.0/project_quota-project_unset.response.schema.json](../../schemas/cli/1.0.0/project_quota-project_unset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:response` |
| [1.0.0/project_show.input.schema.json](../../schemas/cli/1.0.0/project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.show:input` |
| [1.0.0/project_show.response.schema.json](../../schemas/cli/1.0.0/project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.show:response` |
| [1.0.0/project_templates_set.input.schema.json](../../schemas/cli/1.0.0/project_templates_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:input` |
| [1.0.0/project_templates_set.response.schema.json](../../schemas/cli/1.0.0/project_templates_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:response` |
| [1.0.0/project_templates_show.input.schema.json](../../schemas/cli/1.0.0/project_templates_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:input` |
| [1.0.0/project_templates_show.response.schema.json](../../schemas/cli/1.0.0/project_templates_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:response` |
| [1.0.0/projects_aliases_import.input.schema.json](../../schemas/cli/1.0.0/projects_aliases_import.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:input` |
| [1.0.0/projects_aliases_import.response.schema.json](../../schemas/cli/1.0.0/projects_aliases_import.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:response` |
| [1.0.0/projects_aliases_list.input.schema.json](../../schemas/cli/1.0.0/projects_aliases_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:input` |
| [1.0.0/projects_aliases_list.response.schema.json](../../schemas/cli/1.0.0/projects_aliases_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:response` |
| [1.0.0/projects_aliases_remove.input.schema.json](../../schemas/cli/1.0.0/projects_aliases_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:input` |
| [1.0.0/projects_aliases_remove.response.schema.json](../../schemas/cli/1.0.0/projects_aliases_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:response` |
| [1.0.0/projects_aliases_set.input.schema.json](../../schemas/cli/1.0.0/projects_aliases_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:input` |
| [1.0.0/projects_aliases_set.response.schema.json](../../schemas/cli/1.0.0/projects_aliases_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:response` |
| [1.0.0/projects_blacklist_add.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.add:input` |
| [1.0.0/projects_blacklist_add.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.add:response` |
| [1.0.0/projects_blacklist_clear.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.clear:input` |
| [1.0.0/projects_blacklist_clear.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.clear:response` |
| [1.0.0/projects_blacklist_disable.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_disable.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.disable:input` |
| [1.0.0/projects_blacklist_disable.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_disable.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.disable:response` |
| [1.0.0/projects_blacklist_enable.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_enable.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.enable:input` |
| [1.0.0/projects_blacklist_enable.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_enable.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.enable:response` |
| [1.0.0/projects_blacklist_list.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.list:input` |
| [1.0.0/projects_blacklist_list.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.list:response` |
| [1.0.0/projects_blacklist_remove.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.remove:input` |
| [1.0.0/projects_blacklist_remove.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.remove:response` |
| [1.0.0/projects_diff.input.schema.json](../../schemas/cli/1.0.0/projects_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:input` |
| [1.0.0/projects_diff.response.schema.json](../../schemas/cli/1.0.0/projects_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:response` |
| [1.0.0/projects_forget.input.schema.json](../../schemas/cli/1.0.0/projects_forget.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:input` |
| [1.0.0/projects_forget.response.schema.json](../../schemas/cli/1.0.0/projects_forget.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:response` |
| [1.0.0/projects_list.input.schema.json](../../schemas/cli/1.0.0/projects_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.list:input` |
| [1.0.0/projects_list.response.schema.json](../../schemas/cli/1.0.0/projects_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.list:response` |
| [1.0.0/projects_path.input.schema.json](../../schemas/cli/1.0.0/projects_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.path:input` |
| [1.0.0/projects_path.response.schema.json](../../schemas/cli/1.0.0/projects_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.path:response` |
| [1.0.0/projects_promote.input.schema.json](../../schemas/cli/1.0.0/projects_promote.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:input` |
| [1.0.0/projects_promote.response.schema.json](../../schemas/cli/1.0.0/projects_promote.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:response` |
| [1.0.0/projects_reset.input.schema.json](../../schemas/cli/1.0.0/projects_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:input` |
| [1.0.0/projects_reset.response.schema.json](../../schemas/cli/1.0.0/projects_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:response` |
| [1.0.0/projects_update.input.schema.json](../../schemas/cli/1.0.0/projects_update.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.update:input` |
| [1.0.0/projects_update.response.schema.json](../../schemas/cli/1.0.0/projects_update.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.update:response` |
| [1.0.0/projects_whitelist_add.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.add:input` |
| [1.0.0/projects_whitelist_add.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.add:response` |
| [1.0.0/projects_whitelist_clear.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.clear:input` |
| [1.0.0/projects_whitelist_clear.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.clear:response` |
| [1.0.0/projects_whitelist_disable.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_disable.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.disable:input` |
| [1.0.0/projects_whitelist_disable.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_disable.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.disable:response` |
| [1.0.0/projects_whitelist_enable.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_enable.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.enable:input` |
| [1.0.0/projects_whitelist_enable.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_enable.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.enable:response` |
| [1.0.0/projects_whitelist_list.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.list:input` |
| [1.0.0/projects_whitelist_list.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.list:response` |
| [1.0.0/projects_whitelist_remove.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.remove:input` |
| [1.0.0/projects_whitelist_remove.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.remove:response` |
| [1.0.0/publication_plan.schema.json](../../schemas/cli/1.0.0/publication_plan.schema.json) | `urn:fbrcm:schema:cli:1.0.0:publication_plan` |
| [1.0.0/rollouts_delete.input.schema.json](../../schemas/cli/1.0.0/rollouts_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:input` |
| [1.0.0/rollouts_delete.response.schema.json](../../schemas/cli/1.0.0/rollouts_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:response` |
| [1.0.0/rollouts_list.input.schema.json](../../schemas/cli/1.0.0/rollouts_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:input` |
| [1.0.0/rollouts_list.response.schema.json](../../schemas/cli/1.0.0/rollouts_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:response` |
| [1.0.0/rollouts_show.input.schema.json](../../schemas/cli/1.0.0/rollouts_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:input` |
| [1.0.0/rollouts_show.response.schema.json](../../schemas/cli/1.0.0/rollouts_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:response` |
| [1.0.0/root.input.schema.json](../../schemas/cli/1.0.0/root.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:root:input` |
| [1.0.0/root.response.schema.json](../../schemas/cli/1.0.0/root.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:root:response` |
| [1.0.0/schema_list.input.schema.json](../../schemas/cli/1.0.0/schema_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:schema.list:input` |
| [1.0.0/schema_list.response.schema.json](../../schemas/cli/1.0.0/schema_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:schema.list:response` |
| [1.0.0/schema_show.input.schema.json](../../schemas/cli/1.0.0/schema_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:schema.show:input` |
| [1.0.0/schema_show.response.schema.json](../../schemas/cli/1.0.0/schema_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:schema.show:response` |
| [1.0.0/semantic.schema.json](../../schemas/cli/1.0.0/semantic.schema.json) | `urn:fbrcm:schema:cli:1.0.0:semantic` |
| [1.0.0/setup.input.schema.json](../../schemas/cli/1.0.0/setup.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:setup:input` |
| [1.0.0/setup.response.schema.json](../../schemas/cli/1.0.0/setup.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:setup:response` |
| [1.0.0/stdin.credentials.schema.json](../../schemas/cli/1.0.0/stdin.credentials.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:credentials` |
| [1.0.0/stdin.oauth_credentials.schema.json](../../schemas/cli/1.0.0/stdin.oauth_credentials.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:oauth_credentials` |
| [1.0.0/stdin.publication_plan.schema.json](../../schemas/cli/1.0.0/stdin.publication_plan.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:publication_plan` |
| [1.0.0/stdin.remote_config.schema.json](../../schemas/cli/1.0.0/stdin.remote_config.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config` |
| [1.0.0/stdin.remote_config_import.schema.json](../../schemas/cli/1.0.0/stdin.remote_config_import.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config_import` |
| [1.0.0/stdin.service_account_credentials.schema.json](../../schemas/cli/1.0.0/stdin.service_account_credentials.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:service_account_credentials` |
| [1.0.0/stdin.theme.schema.json](../../schemas/cli/1.0.0/stdin.theme.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:theme` |
| [1.0.0/theme.input.schema.json](../../schemas/cli/1.0.0/theme.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme:input` |
| [1.0.0/theme.response.schema.json](../../schemas/cli/1.0.0/theme.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme:response` |
| [1.0.0/theme_delete.input.schema.json](../../schemas/cli/1.0.0/theme_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:input` |
| [1.0.0/theme_delete.response.schema.json](../../schemas/cli/1.0.0/theme_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:response` |
| [1.0.0/theme_import.input.schema.json](../../schemas/cli/1.0.0/theme_import.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.import:input` |
| [1.0.0/theme_import.response.schema.json](../../schemas/cli/1.0.0/theme_import.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.import:response` |
| [1.0.0/theme_list.input.schema.json](../../schemas/cli/1.0.0/theme_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.list:input` |
| [1.0.0/theme_list.response.schema.json](../../schemas/cli/1.0.0/theme_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.list:response` |
| [1.0.0/theme_path.input.schema.json](../../schemas/cli/1.0.0/theme_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.path:input` |
| [1.0.0/theme_path.response.schema.json](../../schemas/cli/1.0.0/theme_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.path:response` |
| [1.0.0/theme_rename.input.schema.json](../../schemas/cli/1.0.0/theme_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:input` |
| [1.0.0/theme_rename.response.schema.json](../../schemas/cli/1.0.0/theme_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:response` |
| [1.0.0/theme_reset.input.schema.json](../../schemas/cli/1.0.0/theme_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:input` |
| [1.0.0/theme_reset.response.schema.json](../../schemas/cli/1.0.0/theme_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:response` |
| [1.0.0/theme_switch.input.schema.json](../../schemas/cli/1.0.0/theme_switch.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:input` |
| [1.0.0/theme_switch.response.schema.json](../../schemas/cli/1.0.0/theme_switch.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:response` |
| [1.0.0/update.input.schema.json](../../schemas/cli/1.0.0/update.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:update:input` |
| [1.0.0/update.response.schema.json](../../schemas/cli/1.0.0/update.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:update:response` |
| [1.0.0/versions_blame.input.schema.json](../../schemas/cli/1.0.0/versions_blame.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.blame:input` |
| [1.0.0/versions_blame.response.schema.json](../../schemas/cli/1.0.0/versions_blame.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.blame:response` |
| [1.0.0/versions_diff.input.schema.json](../../schemas/cli/1.0.0/versions_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:input` |
| [1.0.0/versions_diff.response.schema.json](../../schemas/cli/1.0.0/versions_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:response` |
| [1.0.0/versions_export.input.schema.json](../../schemas/cli/1.0.0/versions_export.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.export:input` |
| [1.0.0/versions_export.response.schema.json](../../schemas/cli/1.0.0/versions_export.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.export:response` |
| [1.0.0/versions_list.input.schema.json](../../schemas/cli/1.0.0/versions_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.list:input` |
| [1.0.0/versions_list.response.schema.json](../../schemas/cli/1.0.0/versions_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.list:response` |
| [1.0.0/versions_restore.input.schema.json](../../schemas/cli/1.0.0/versions_restore.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:input` |
| [1.0.0/versions_restore.response.schema.json](../../schemas/cli/1.0.0/versions_restore.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:response` |
| [1.0.0/versions_rollback.input.schema.json](../../schemas/cli/1.0.0/versions_rollback.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:input` |
| [1.0.0/versions_rollback.response.schema.json](../../schemas/cli/1.0.0/versions_rollback.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:response` |
| [1.0.0/versions_show.input.schema.json](../../schemas/cli/1.0.0/versions_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.show:input` |
| [1.0.0/versions_show.response.schema.json](../../schemas/cli/1.0.0/versions_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.show:response` |
| [1.1.0/add.input.schema.json](../../schemas/cli/1.1.0/add.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:add:input` |
| [1.1.0/add.response.schema.json](../../schemas/cli/1.1.0/add.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:add:response` |
| [1.1.0/apply.input.schema.json](../../schemas/cli/1.1.0/apply.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:apply:input` |
| [1.1.0/apply.response.schema.json](../../schemas/cli/1.1.0/apply.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:apply:response` |
| [1.1.0/apps_config.input.schema.json](../../schemas/cli/1.1.0/apps_config.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:apps.config:input` |
| [1.1.0/apps_config.response.schema.json](../../schemas/cli/1.1.0/apps_config.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:apps.config:response` |
| [1.1.0/apps_list.input.schema.json](../../schemas/cli/1.1.0/apps_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:apps.list:input` |
| [1.1.0/apps_list.response.schema.json](../../schemas/cli/1.1.0/apps_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:apps.list:response` |
| [1.1.0/apps_show.input.schema.json](../../schemas/cli/1.1.0/apps_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:apps.show:input` |
| [1.1.0/apps_show.response.schema.json](../../schemas/cli/1.1.0/apps_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:apps.show:response` |
| [1.1.0/auth_add_gcloud.input.schema.json](../../schemas/cli/1.1.0/auth_add_gcloud.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.gcloud:input` |
| [1.1.0/auth_add_gcloud.response.schema.json](../../schemas/cli/1.1.0/auth_add_gcloud.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.gcloud:response` |
| [1.1.0/auth_add_google.input.schema.json](../../schemas/cli/1.1.0/auth_add_google.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.google:input` |
| [1.1.0/auth_add_google.response.schema.json](../../schemas/cli/1.1.0/auth_add_google.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.google:response` |
| [1.1.0/auth_add_oauth.input.schema.json](../../schemas/cli/1.1.0/auth_add_oauth.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.oauth:input` |
| [1.1.0/auth_add_oauth.response.schema.json](../../schemas/cli/1.1.0/auth_add_oauth.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.oauth:response` |
| [1.1.0/auth_add_service-account.input.schema.json](../../schemas/cli/1.1.0/auth_add_service-account.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.service-account:input` |
| [1.1.0/auth_add_service-account.response.schema.json](../../schemas/cli/1.1.0/auth_add_service-account.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.service-account:response` |
| [1.1.0/auth_bind.input.schema.json](../../schemas/cli/1.1.0/auth_bind.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.bind:input` |
| [1.1.0/auth_bind.response.schema.json](../../schemas/cli/1.1.0/auth_bind.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.bind:response` |
| [1.1.0/auth_delete.input.schema.json](../../schemas/cli/1.1.0/auth_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.delete:input` |
| [1.1.0/auth_delete.response.schema.json](../../schemas/cli/1.1.0/auth_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.delete:response` |
| [1.1.0/auth_list.input.schema.json](../../schemas/cli/1.1.0/auth_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.list:input` |
| [1.1.0/auth_list.response.schema.json](../../schemas/cli/1.1.0/auth_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.list:response` |
| [1.1.0/auth_login.input.schema.json](../../schemas/cli/1.1.0/auth_login.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.login:input` |
| [1.1.0/auth_login.response.schema.json](../../schemas/cli/1.1.0/auth_login.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.login:response` |
| [1.1.0/auth_path.input.schema.json](../../schemas/cli/1.1.0/auth_path.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.path:input` |
| [1.1.0/auth_path.response.schema.json](../../schemas/cli/1.1.0/auth_path.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.path:response` |
| [1.1.0/auth_quota-project_set.input.schema.json](../../schemas/cli/1.1.0/auth_quota-project_set.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.set:input` |
| [1.1.0/auth_quota-project_set.response.schema.json](../../schemas/cli/1.1.0/auth_quota-project_set.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.set:response` |
| [1.1.0/auth_quota-project_show.input.schema.json](../../schemas/cli/1.1.0/auth_quota-project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.show:input` |
| [1.1.0/auth_quota-project_show.response.schema.json](../../schemas/cli/1.1.0/auth_quota-project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.show:response` |
| [1.1.0/auth_quota-project_unset.input.schema.json](../../schemas/cli/1.1.0/auth_quota-project_unset.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.unset:input` |
| [1.1.0/auth_quota-project_unset.response.schema.json](../../schemas/cli/1.1.0/auth_quota-project_unset.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.unset:response` |
| [1.1.0/cache_clear.input.schema.json](../../schemas/cli/1.1.0/cache_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:cache.clear:input` |
| [1.1.0/cache_clear.response.schema.json](../../schemas/cli/1.1.0/cache_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:cache.clear:response` |
| [1.1.0/cache_list.input.schema.json](../../schemas/cli/1.1.0/cache_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:cache.list:input` |
| [1.1.0/cache_list.response.schema.json](../../schemas/cli/1.1.0/cache_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:cache.list:response` |
| [1.1.0/cache_path.input.schema.json](../../schemas/cli/1.1.0/cache_path.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:cache.path:input` |
| [1.1.0/cache_path.response.schema.json](../../schemas/cli/1.1.0/cache_path.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:cache.path:response` |
| [1.1.0/capabilities.input.schema.json](../../schemas/cli/1.1.0/capabilities.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:capabilities:input` |
| [1.1.0/capabilities.response.schema.json](../../schemas/cli/1.1.0/capabilities.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:capabilities:response` |
| [1.1.0/capability.schema.json](../../schemas/cli/1.1.0/capability.schema.json) | `urn:fbrcm:schema:cli:1.1.0:capability` |
| [1.1.0/completion_bash.input.schema.json](../../schemas/cli/1.1.0/completion_bash.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:completion.bash:input` |
| [1.1.0/completion_bash.response.schema.json](../../schemas/cli/1.1.0/completion_bash.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:completion.bash:response` |
| [1.1.0/completion_fish.input.schema.json](../../schemas/cli/1.1.0/completion_fish.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:completion.fish:input` |
| [1.1.0/completion_fish.response.schema.json](../../schemas/cli/1.1.0/completion_fish.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:completion.fish:response` |
| [1.1.0/completion_powershell.input.schema.json](../../schemas/cli/1.1.0/completion_powershell.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:completion.powershell:input` |
| [1.1.0/completion_powershell.response.schema.json](../../schemas/cli/1.1.0/completion_powershell.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:completion.powershell:response` |
| [1.1.0/completion_zsh.input.schema.json](../../schemas/cli/1.1.0/completion_zsh.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:completion.zsh:input` |
| [1.1.0/completion_zsh.response.schema.json](../../schemas/cli/1.1.0/completion_zsh.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:completion.zsh:response` |
| [1.1.0/conditions_add.input.schema.json](../../schemas/cli/1.1.0/conditions_add.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.add:input` |
| [1.1.0/conditions_add.response.schema.json](../../schemas/cli/1.1.0/conditions_add.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.add:response` |
| [1.1.0/conditions_delete.input.schema.json](../../schemas/cli/1.1.0/conditions_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.delete:input` |
| [1.1.0/conditions_delete.response.schema.json](../../schemas/cli/1.1.0/conditions_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.delete:response` |
| [1.1.0/conditions_edit.input.schema.json](../../schemas/cli/1.1.0/conditions_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.edit:input` |
| [1.1.0/conditions_edit.response.schema.json](../../schemas/cli/1.1.0/conditions_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.edit:response` |
| [1.1.0/conditions_list.input.schema.json](../../schemas/cli/1.1.0/conditions_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.list:input` |
| [1.1.0/conditions_list.response.schema.json](../../schemas/cli/1.1.0/conditions_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.list:response` |
| [1.1.0/conditions_move.input.schema.json](../../schemas/cli/1.1.0/conditions_move.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.move:input` |
| [1.1.0/conditions_move.response.schema.json](../../schemas/cli/1.1.0/conditions_move.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.move:response` |
| [1.1.0/conditions_rename.input.schema.json](../../schemas/cli/1.1.0/conditions_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.rename:input` |
| [1.1.0/conditions_rename.response.schema.json](../../schemas/cli/1.1.0/conditions_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.rename:response` |
| [1.1.0/conditions_show.input.schema.json](../../schemas/cli/1.1.0/conditions_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.show:input` |
| [1.1.0/conditions_show.response.schema.json](../../schemas/cli/1.1.0/conditions_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.show:response` |
| [1.1.0/conditions_validate.input.schema.json](../../schemas/cli/1.1.0/conditions_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.validate:input` |
| [1.1.0/conditions_validate.response.schema.json](../../schemas/cli/1.1.0/conditions_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:conditions.validate:response` |
| [1.1.0/config_edit.input.schema.json](../../schemas/cli/1.1.0/config_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.edit:input` |
| [1.1.0/config_edit.response.schema.json](../../schemas/cli/1.1.0/config_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.edit:response` |
| [1.1.0/config_path.input.schema.json](../../schemas/cli/1.1.0/config_path.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.path:input` |
| [1.1.0/config_path.response.schema.json](../../schemas/cli/1.1.0/config_path.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.path:response` |
| [1.1.0/config_reset.input.schema.json](../../schemas/cli/1.1.0/config_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.reset:input` |
| [1.1.0/config_reset.response.schema.json](../../schemas/cli/1.1.0/config_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.reset:response` |
| [1.1.0/config_set.input.schema.json](../../schemas/cli/1.1.0/config_set.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.set:input` |
| [1.1.0/config_set.response.schema.json](../../schemas/cli/1.1.0/config_set.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.set:response` |
| [1.1.0/config_show.input.schema.json](../../schemas/cli/1.1.0/config_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.show:input` |
| [1.1.0/config_show.response.schema.json](../../schemas/cli/1.1.0/config_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.show:response` |
| [1.1.0/config_validate.input.schema.json](../../schemas/cli/1.1.0/config_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.validate:input` |
| [1.1.0/config_validate.response.schema.json](../../schemas/cli/1.1.0/config_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:config.validate:response` |
| [1.1.0/delete.input.schema.json](../../schemas/cli/1.1.0/delete.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:delete:input` |
| [1.1.0/delete.response.schema.json](../../schemas/cli/1.1.0/delete.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:delete:response` |
| [1.1.0/doctor.input.schema.json](../../schemas/cli/1.1.0/doctor.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:doctor:input` |
| [1.1.0/doctor.response.schema.json](../../schemas/cli/1.1.0/doctor.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:doctor:response` |
| [1.1.0/draft_change-note.input.schema.json](../../schemas/cli/1.1.0/draft_change-note.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.change-note:input` |
| [1.1.0/draft_change-note.response.schema.json](../../schemas/cli/1.1.0/draft_change-note.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.change-note:response` |
| [1.1.0/draft_diff.input.schema.json](../../schemas/cli/1.1.0/draft_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.diff:input` |
| [1.1.0/draft_diff.response.schema.json](../../schemas/cli/1.1.0/draft_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.diff:response` |
| [1.1.0/draft_discard.input.schema.json](../../schemas/cli/1.1.0/draft_discard.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.discard:input` |
| [1.1.0/draft_discard.response.schema.json](../../schemas/cli/1.1.0/draft_discard.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.discard:response` |
| [1.1.0/draft_list.input.schema.json](../../schemas/cli/1.1.0/draft_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.list:input` |
| [1.1.0/draft_list.response.schema.json](../../schemas/cli/1.1.0/draft_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.list:response` |
| [1.1.0/draft_path.input.schema.json](../../schemas/cli/1.1.0/draft_path.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.path:input` |
| [1.1.0/draft_path.response.schema.json](../../schemas/cli/1.1.0/draft_path.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.path:response` |
| [1.1.0/draft_publish.input.schema.json](../../schemas/cli/1.1.0/draft_publish.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.publish:input` |
| [1.1.0/draft_publish.response.schema.json](../../schemas/cli/1.1.0/draft_publish.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.publish:response` |
| [1.1.0/draft_show.input.schema.json](../../schemas/cli/1.1.0/draft_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.show:input` |
| [1.1.0/draft_show.response.schema.json](../../schemas/cli/1.1.0/draft_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:draft.show:response` |
| [1.1.0/duplicate.input.schema.json](../../schemas/cli/1.1.0/duplicate.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:duplicate:input` |
| [1.1.0/duplicate.response.schema.json](../../schemas/cli/1.1.0/duplicate.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:duplicate:response` |
| [1.1.0/envelope.schema.json](../../schemas/cli/1.1.0/envelope.schema.json) | `urn:fbrcm:schema:cli:1.1.0:envelope` |
| [1.1.0/error.schema.json](../../schemas/cli/1.1.0/error.schema.json) | `urn:fbrcm:schema:cli:1.1.0:error` |
| [1.1.0/experiments_delete.input.schema.json](../../schemas/cli/1.1.0/experiments_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:experiments.delete:input` |
| [1.1.0/experiments_delete.response.schema.json](../../schemas/cli/1.1.0/experiments_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:experiments.delete:response` |
| [1.1.0/experiments_list.input.schema.json](../../schemas/cli/1.1.0/experiments_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:experiments.list:input` |
| [1.1.0/experiments_list.response.schema.json](../../schemas/cli/1.1.0/experiments_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:experiments.list:response` |
| [1.1.0/experiments_show.input.schema.json](../../schemas/cli/1.1.0/experiments_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:experiments.show:input` |
| [1.1.0/experiments_show.response.schema.json](../../schemas/cli/1.1.0/experiments_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:experiments.show:response` |
| [1.1.0/explain.input.schema.json](../../schemas/cli/1.1.0/explain.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:explain:input` |
| [1.1.0/explain.response.schema.json](../../schemas/cli/1.1.0/explain.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:explain:response` |
| [1.1.0/get.input.schema.json](../../schemas/cli/1.1.0/get.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:get:input` |
| [1.1.0/get.response.schema.json](../../schemas/cli/1.1.0/get.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:get:response` |
| [1.1.0/groups_add.input.schema.json](../../schemas/cli/1.1.0/groups_add.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.add:input` |
| [1.1.0/groups_add.response.schema.json](../../schemas/cli/1.1.0/groups_add.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.add:response` |
| [1.1.0/groups_delete.input.schema.json](../../schemas/cli/1.1.0/groups_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.delete:input` |
| [1.1.0/groups_delete.response.schema.json](../../schemas/cli/1.1.0/groups_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.delete:response` |
| [1.1.0/groups_edit.input.schema.json](../../schemas/cli/1.1.0/groups_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.edit:input` |
| [1.1.0/groups_edit.response.schema.json](../../schemas/cli/1.1.0/groups_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.edit:response` |
| [1.1.0/groups_list.input.schema.json](../../schemas/cli/1.1.0/groups_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.list:input` |
| [1.1.0/groups_list.response.schema.json](../../schemas/cli/1.1.0/groups_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.list:response` |
| [1.1.0/groups_rename.input.schema.json](../../schemas/cli/1.1.0/groups_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.rename:input` |
| [1.1.0/groups_rename.response.schema.json](../../schemas/cli/1.1.0/groups_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:groups.rename:response` |
| [1.1.0/help.input.schema.json](../../schemas/cli/1.1.0/help.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:help:input` |
| [1.1.0/help.response.schema.json](../../schemas/cli/1.1.0/help.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:help:response` |
| [1.1.0/hooks_fingerprint.input.schema.json](../../schemas/cli/1.1.0/hooks_fingerprint.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:hooks.fingerprint:input` |
| [1.1.0/hooks_fingerprint.response.schema.json](../../schemas/cli/1.1.0/hooks_fingerprint.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:hooks.fingerprint:response` |
| [1.1.0/hooks_status.input.schema.json](../../schemas/cli/1.1.0/hooks_status.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:hooks.status:input` |
| [1.1.0/hooks_status.response.schema.json](../../schemas/cli/1.1.0/hooks_status.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:hooks.status:response` |
| [1.1.0/hooks_trust.input.schema.json](../../schemas/cli/1.1.0/hooks_trust.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:hooks.trust:input` |
| [1.1.0/hooks_trust.response.schema.json](../../schemas/cli/1.1.0/hooks_trust.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:hooks.trust:response` |
| [1.1.0/hooks_untrust.input.schema.json](../../schemas/cli/1.1.0/hooks_untrust.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:hooks.untrust:input` |
| [1.1.0/hooks_untrust.response.schema.json](../../schemas/cli/1.1.0/hooks_untrust.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:hooks.untrust:response` |
| [1.1.0/mcp.input.schema.json](../../schemas/cli/1.1.0/mcp.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:mcp:input` |
| [1.1.0/mcp.response.schema.json](../../schemas/cli/1.1.0/mcp.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:mcp:response` |
| [1.1.0/personalizations_list.input.schema.json](../../schemas/cli/1.1.0/personalizations_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:personalizations.list:input` |
| [1.1.0/personalizations_list.response.schema.json](../../schemas/cli/1.1.0/personalizations_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:personalizations.list:response` |
| [1.1.0/personalizations_show.input.schema.json](../../schemas/cli/1.1.0/personalizations_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:personalizations.show:input` |
| [1.1.0/personalizations_show.response.schema.json](../../schemas/cli/1.1.0/personalizations_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:personalizations.show:response` |
| [1.1.0/plan_show.input.schema.json](../../schemas/cli/1.1.0/plan_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:plan.show:input` |
| [1.1.0/plan_show.response.schema.json](../../schemas/cli/1.1.0/plan_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:plan.show:response` |
| [1.1.0/plan_validate.input.schema.json](../../schemas/cli/1.1.0/plan_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:plan.validate:input` |
| [1.1.0/plan_validate.response.schema.json](../../schemas/cli/1.1.0/plan_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:plan.validate:response` |
| [1.1.0/profile.input.schema.json](../../schemas/cli/1.1.0/profile.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile:input` |
| [1.1.0/profile.response.schema.json](../../schemas/cli/1.1.0/profile.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile:response` |
| [1.1.0/profile_delete.input.schema.json](../../schemas/cli/1.1.0/profile_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.delete:input` |
| [1.1.0/profile_delete.response.schema.json](../../schemas/cli/1.1.0/profile_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.delete:response` |
| [1.1.0/profile_list.input.schema.json](../../schemas/cli/1.1.0/profile_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.list:input` |
| [1.1.0/profile_list.response.schema.json](../../schemas/cli/1.1.0/profile_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.list:response` |
| [1.1.0/profile_path.input.schema.json](../../schemas/cli/1.1.0/profile_path.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.path:input` |
| [1.1.0/profile_path.response.schema.json](../../schemas/cli/1.1.0/profile_path.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.path:response` |
| [1.1.0/profile_rename.input.schema.json](../../schemas/cli/1.1.0/profile_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.rename:input` |
| [1.1.0/profile_rename.response.schema.json](../../schemas/cli/1.1.0/profile_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.rename:response` |
| [1.1.0/profile_switch.input.schema.json](../../schemas/cli/1.1.0/profile_switch.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.switch:input` |
| [1.1.0/profile_switch.response.schema.json](../../schemas/cli/1.1.0/profile_switch.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:profile.switch:response` |
| [1.1.0/project_defaults.input.schema.json](../../schemas/cli/1.1.0/project_defaults.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.defaults:input` |
| [1.1.0/project_defaults.response.schema.json](../../schemas/cli/1.1.0/project_defaults.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.defaults:response` |
| [1.1.0/project_export.input.schema.json](../../schemas/cli/1.1.0/project_export.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.export:input` |
| [1.1.0/project_export.response.schema.json](../../schemas/cli/1.1.0/project_export.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.export:response` |
| [1.1.0/project_import.input.schema.json](../../schemas/cli/1.1.0/project_import.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.import:input` |
| [1.1.0/project_import.response.schema.json](../../schemas/cli/1.1.0/project_import.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.import:response` |
| [1.1.0/project_open.input.schema.json](../../schemas/cli/1.1.0/project_open.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.open:input` |
| [1.1.0/project_open.response.schema.json](../../schemas/cli/1.1.0/project_open.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.open:response` |
| [1.1.0/project_quota-project_set.input.schema.json](../../schemas/cli/1.1.0/project_quota-project_set.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.set:input` |
| [1.1.0/project_quota-project_set.response.schema.json](../../schemas/cli/1.1.0/project_quota-project_set.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.set:response` |
| [1.1.0/project_quota-project_show.input.schema.json](../../schemas/cli/1.1.0/project_quota-project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.show:input` |
| [1.1.0/project_quota-project_show.response.schema.json](../../schemas/cli/1.1.0/project_quota-project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.show:response` |
| [1.1.0/project_quota-project_unset.input.schema.json](../../schemas/cli/1.1.0/project_quota-project_unset.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.unset:input` |
| [1.1.0/project_quota-project_unset.response.schema.json](../../schemas/cli/1.1.0/project_quota-project_unset.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.unset:response` |
| [1.1.0/project_show.input.schema.json](../../schemas/cli/1.1.0/project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.show:input` |
| [1.1.0/project_show.response.schema.json](../../schemas/cli/1.1.0/project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.show:response` |
| [1.1.0/project_templates_set.input.schema.json](../../schemas/cli/1.1.0/project_templates_set.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.templates.set:input` |
| [1.1.0/project_templates_set.response.schema.json](../../schemas/cli/1.1.0/project_templates_set.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.templates.set:response` |
| [1.1.0/project_templates_show.input.schema.json](../../schemas/cli/1.1.0/project_templates_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.templates.show:input` |
| [1.1.0/project_templates_show.response.schema.json](../../schemas/cli/1.1.0/project_templates_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:project.templates.show:response` |
| [1.1.0/projects_aliases_import.input.schema.json](../../schemas/cli/1.1.0/projects_aliases_import.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.import:input` |
| [1.1.0/projects_aliases_import.response.schema.json](../../schemas/cli/1.1.0/projects_aliases_import.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.import:response` |
| [1.1.0/projects_aliases_list.input.schema.json](../../schemas/cli/1.1.0/projects_aliases_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.list:input` |
| [1.1.0/projects_aliases_list.response.schema.json](../../schemas/cli/1.1.0/projects_aliases_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.list:response` |
| [1.1.0/projects_aliases_remove.input.schema.json](../../schemas/cli/1.1.0/projects_aliases_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.remove:input` |
| [1.1.0/projects_aliases_remove.response.schema.json](../../schemas/cli/1.1.0/projects_aliases_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.remove:response` |
| [1.1.0/projects_aliases_set.input.schema.json](../../schemas/cli/1.1.0/projects_aliases_set.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.set:input` |
| [1.1.0/projects_aliases_set.response.schema.json](../../schemas/cli/1.1.0/projects_aliases_set.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.set:response` |
| [1.1.0/projects_blacklist_add.input.schema.json](../../schemas/cli/1.1.0/projects_blacklist_add.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.add:input` |
| [1.1.0/projects_blacklist_add.response.schema.json](../../schemas/cli/1.1.0/projects_blacklist_add.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.add:response` |
| [1.1.0/projects_blacklist_clear.input.schema.json](../../schemas/cli/1.1.0/projects_blacklist_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.clear:input` |
| [1.1.0/projects_blacklist_clear.response.schema.json](../../schemas/cli/1.1.0/projects_blacklist_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.clear:response` |
| [1.1.0/projects_blacklist_disable.input.schema.json](../../schemas/cli/1.1.0/projects_blacklist_disable.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.disable:input` |
| [1.1.0/projects_blacklist_disable.response.schema.json](../../schemas/cli/1.1.0/projects_blacklist_disable.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.disable:response` |
| [1.1.0/projects_blacklist_enable.input.schema.json](../../schemas/cli/1.1.0/projects_blacklist_enable.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.enable:input` |
| [1.1.0/projects_blacklist_enable.response.schema.json](../../schemas/cli/1.1.0/projects_blacklist_enable.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.enable:response` |
| [1.1.0/projects_blacklist_list.input.schema.json](../../schemas/cli/1.1.0/projects_blacklist_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.list:input` |
| [1.1.0/projects_blacklist_list.response.schema.json](../../schemas/cli/1.1.0/projects_blacklist_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.list:response` |
| [1.1.0/projects_blacklist_remove.input.schema.json](../../schemas/cli/1.1.0/projects_blacklist_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.remove:input` |
| [1.1.0/projects_blacklist_remove.response.schema.json](../../schemas/cli/1.1.0/projects_blacklist_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.remove:response` |
| [1.1.0/projects_diff.input.schema.json](../../schemas/cli/1.1.0/projects_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.diff:input` |
| [1.1.0/projects_diff.response.schema.json](../../schemas/cli/1.1.0/projects_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.diff:response` |
| [1.1.0/projects_forget.input.schema.json](../../schemas/cli/1.1.0/projects_forget.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.forget:input` |
| [1.1.0/projects_forget.response.schema.json](../../schemas/cli/1.1.0/projects_forget.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.forget:response` |
| [1.1.0/projects_list.input.schema.json](../../schemas/cli/1.1.0/projects_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.list:input` |
| [1.1.0/projects_list.response.schema.json](../../schemas/cli/1.1.0/projects_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.list:response` |
| [1.1.0/projects_path.input.schema.json](../../schemas/cli/1.1.0/projects_path.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.path:input` |
| [1.1.0/projects_path.response.schema.json](../../schemas/cli/1.1.0/projects_path.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.path:response` |
| [1.1.0/projects_promote.input.schema.json](../../schemas/cli/1.1.0/projects_promote.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.promote:input` |
| [1.1.0/projects_promote.response.schema.json](../../schemas/cli/1.1.0/projects_promote.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.promote:response` |
| [1.1.0/projects_reset.input.schema.json](../../schemas/cli/1.1.0/projects_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.reset:input` |
| [1.1.0/projects_reset.response.schema.json](../../schemas/cli/1.1.0/projects_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.reset:response` |
| [1.1.0/projects_update.input.schema.json](../../schemas/cli/1.1.0/projects_update.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.update:input` |
| [1.1.0/projects_update.response.schema.json](../../schemas/cli/1.1.0/projects_update.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.update:response` |
| [1.1.0/projects_whitelist_add.input.schema.json](../../schemas/cli/1.1.0/projects_whitelist_add.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.add:input` |
| [1.1.0/projects_whitelist_add.response.schema.json](../../schemas/cli/1.1.0/projects_whitelist_add.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.add:response` |
| [1.1.0/projects_whitelist_clear.input.schema.json](../../schemas/cli/1.1.0/projects_whitelist_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.clear:input` |
| [1.1.0/projects_whitelist_clear.response.schema.json](../../schemas/cli/1.1.0/projects_whitelist_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.clear:response` |
| [1.1.0/projects_whitelist_disable.input.schema.json](../../schemas/cli/1.1.0/projects_whitelist_disable.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.disable:input` |
| [1.1.0/projects_whitelist_disable.response.schema.json](../../schemas/cli/1.1.0/projects_whitelist_disable.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.disable:response` |
| [1.1.0/projects_whitelist_enable.input.schema.json](../../schemas/cli/1.1.0/projects_whitelist_enable.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.enable:input` |
| [1.1.0/projects_whitelist_enable.response.schema.json](../../schemas/cli/1.1.0/projects_whitelist_enable.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.enable:response` |
| [1.1.0/projects_whitelist_list.input.schema.json](../../schemas/cli/1.1.0/projects_whitelist_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.list:input` |
| [1.1.0/projects_whitelist_list.response.schema.json](../../schemas/cli/1.1.0/projects_whitelist_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.list:response` |
| [1.1.0/projects_whitelist_remove.input.schema.json](../../schemas/cli/1.1.0/projects_whitelist_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.remove:input` |
| [1.1.0/projects_whitelist_remove.response.schema.json](../../schemas/cli/1.1.0/projects_whitelist_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.remove:response` |
| [1.1.0/rollouts_delete.input.schema.json](../../schemas/cli/1.1.0/rollouts_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.delete:input` |
| [1.1.0/rollouts_delete.response.schema.json](../../schemas/cli/1.1.0/rollouts_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.delete:response` |
| [1.1.0/rollouts_list.input.schema.json](../../schemas/cli/1.1.0/rollouts_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.list:input` |
| [1.1.0/rollouts_list.response.schema.json](../../schemas/cli/1.1.0/rollouts_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.list:response` |
| [1.1.0/rollouts_show.input.schema.json](../../schemas/cli/1.1.0/rollouts_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.show:input` |
| [1.1.0/rollouts_show.response.schema.json](../../schemas/cli/1.1.0/rollouts_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.show:response` |
| [1.1.0/root.input.schema.json](../../schemas/cli/1.1.0/root.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:root:input` |
| [1.1.0/root.response.schema.json](../../schemas/cli/1.1.0/root.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:root:response` |
| [1.1.0/schema_list.input.schema.json](../../schemas/cli/1.1.0/schema_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:schema.list:input` |
| [1.1.0/schema_list.response.schema.json](../../schemas/cli/1.1.0/schema_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:schema.list:response` |
| [1.1.0/schema_show.input.schema.json](../../schemas/cli/1.1.0/schema_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:schema.show:input` |
| [1.1.0/schema_show.response.schema.json](../../schemas/cli/1.1.0/schema_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:schema.show:response` |
| [1.1.0/semantic.schema.json](../../schemas/cli/1.1.0/semantic.schema.json) | `urn:fbrcm:schema:cli:1.1.0:semantic` |
| [1.1.0/setup.input.schema.json](../../schemas/cli/1.1.0/setup.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:setup:input` |
| [1.1.0/setup.response.schema.json](../../schemas/cli/1.1.0/setup.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:setup:response` |
| [1.1.0/stdin.capture.schema.json](../../schemas/cli/1.1.0/stdin.capture.schema.json) | `urn:fbrcm:schema:cli:1.1.0:stdin:capture` |
| [1.1.0/stdin.credentials.schema.json](../../schemas/cli/1.1.0/stdin.credentials.schema.json) | `urn:fbrcm:schema:cli:1.1.0:stdin:credentials` |
| [1.1.0/stdin.oauth_credentials.schema.json](../../schemas/cli/1.1.0/stdin.oauth_credentials.schema.json) | `urn:fbrcm:schema:cli:1.1.0:stdin:oauth_credentials` |
| [1.1.0/stdin.publication_plan.schema.json](../../schemas/cli/1.1.0/stdin.publication_plan.schema.json) | `urn:fbrcm:schema:cli:1.1.0:stdin:publication_plan` |
| [1.1.0/stdin.remote_config.schema.json](../../schemas/cli/1.1.0/stdin.remote_config.schema.json) | `urn:fbrcm:schema:cli:1.1.0:stdin:remote_config` |
| [1.1.0/stdin.remote_config_import.schema.json](../../schemas/cli/1.1.0/stdin.remote_config_import.schema.json) | `urn:fbrcm:schema:cli:1.1.0:stdin:remote_config_import` |
| [1.1.0/stdin.service_account_credentials.schema.json](../../schemas/cli/1.1.0/stdin.service_account_credentials.schema.json) | `urn:fbrcm:schema:cli:1.1.0:stdin:service_account_credentials` |
| [1.1.0/stdin.theme.schema.json](../../schemas/cli/1.1.0/stdin.theme.schema.json) | `urn:fbrcm:schema:cli:1.1.0:stdin:theme` |
| [1.1.0/theme.input.schema.json](../../schemas/cli/1.1.0/theme.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme:input` |
| [1.1.0/theme.response.schema.json](../../schemas/cli/1.1.0/theme.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme:response` |
| [1.1.0/theme_delete.input.schema.json](../../schemas/cli/1.1.0/theme_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.delete:input` |
| [1.1.0/theme_delete.response.schema.json](../../schemas/cli/1.1.0/theme_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.delete:response` |
| [1.1.0/theme_import.input.schema.json](../../schemas/cli/1.1.0/theme_import.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.import:input` |
| [1.1.0/theme_import.response.schema.json](../../schemas/cli/1.1.0/theme_import.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.import:response` |
| [1.1.0/theme_list.input.schema.json](../../schemas/cli/1.1.0/theme_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.list:input` |
| [1.1.0/theme_list.response.schema.json](../../schemas/cli/1.1.0/theme_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.list:response` |
| [1.1.0/theme_path.input.schema.json](../../schemas/cli/1.1.0/theme_path.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.path:input` |
| [1.1.0/theme_path.response.schema.json](../../schemas/cli/1.1.0/theme_path.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.path:response` |
| [1.1.0/theme_rename.input.schema.json](../../schemas/cli/1.1.0/theme_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.rename:input` |
| [1.1.0/theme_rename.response.schema.json](../../schemas/cli/1.1.0/theme_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.rename:response` |
| [1.1.0/theme_reset.input.schema.json](../../schemas/cli/1.1.0/theme_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.reset:input` |
| [1.1.0/theme_reset.response.schema.json](../../schemas/cli/1.1.0/theme_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.reset:response` |
| [1.1.0/theme_switch.input.schema.json](../../schemas/cli/1.1.0/theme_switch.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.switch:input` |
| [1.1.0/theme_switch.response.schema.json](../../schemas/cli/1.1.0/theme_switch.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:theme.switch:response` |
| [1.1.0/update.input.schema.json](../../schemas/cli/1.1.0/update.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:update:input` |
| [1.1.0/update.response.schema.json](../../schemas/cli/1.1.0/update.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:update:response` |
| [1.1.0/versions_blame.input.schema.json](../../schemas/cli/1.1.0/versions_blame.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.blame:input` |
| [1.1.0/versions_blame.response.schema.json](../../schemas/cli/1.1.0/versions_blame.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.blame:response` |
| [1.1.0/versions_diff.input.schema.json](../../schemas/cli/1.1.0/versions_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.diff:input` |
| [1.1.0/versions_diff.response.schema.json](../../schemas/cli/1.1.0/versions_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.diff:response` |
| [1.1.0/versions_export.input.schema.json](../../schemas/cli/1.1.0/versions_export.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.export:input` |
| [1.1.0/versions_export.response.schema.json](../../schemas/cli/1.1.0/versions_export.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.export:response` |
| [1.1.0/versions_list.input.schema.json](../../schemas/cli/1.1.0/versions_list.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.list:input` |
| [1.1.0/versions_list.response.schema.json](../../schemas/cli/1.1.0/versions_list.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.list:response` |
| [1.1.0/versions_restore.input.schema.json](../../schemas/cli/1.1.0/versions_restore.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.restore:input` |
| [1.1.0/versions_restore.response.schema.json](../../schemas/cli/1.1.0/versions_restore.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.restore:response` |
| [1.1.0/versions_rollback.input.schema.json](../../schemas/cli/1.1.0/versions_rollback.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.rollback:input` |
| [1.1.0/versions_rollback.response.schema.json](../../schemas/cli/1.1.0/versions_rollback.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.rollback:response` |
| [1.1.0/versions_show.input.schema.json](../../schemas/cli/1.1.0/versions_show.input.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.show:input` |
| [1.1.0/versions_show.response.schema.json](../../schemas/cli/1.1.0/versions_show.response.schema.json) | `urn:fbrcm:schema:cli:1.1.0:command:versions.show:response` |
