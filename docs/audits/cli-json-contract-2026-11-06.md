# CLI JSON contract audit — 2026-11-06

```text
Repository revision: acc2540013a5120252def0ddc7f99e29078a5374 plus the audited repair worktree
Contract version: 1.1.0
Contract released: false
Audit standard: docs/cli-contract-audit.md version 1.0.0
Executable commands: 131
Input schemas: 131
Response schemas: 131
Shared schemas: 12 current, 12 retained
Pre-existing dirty files: untracked test.har (untouched)
Verdict: AUTHORITATIVE: PASS
Findings: 0 unresolved (4 found and closed)
```

The revision, starting worktree, contract version, release state and audit-standard
version were frozen before inspection. The 51 criteria were not changed. Scope
includes the complete CLI JSON interface, current and retained schemas, all
options/selectors/stdin/effects/results, generated capabilities/evidence, and
machine documentation. Existing passing command records from the 2026-10-24
audit were rechecked against current artifacts and the complete regression
suites; changed and dependent records were retested. This follows the record
reuse rule in section 12 of the frozen standard.

The repairs correct unreleased 1.1.0 metadata and schemas to describe existing
runtime behavior. They introduce no executable command or runtime behavior change.
The released 1.0.0 schemas and archived release lock remain byte-for-byte unchanged;
1.1.0 remains unreleased. The audit standard and compatibility boundary remain intact.

## Inventory and identity

`E = C = I = R = D = B = M = 131`, with no duplicates and empty symmetric
differences. The live Cobra traversal excludes explicitly marked navigation
nodes and includes root, help and all four completion operations. DTO lookup
succeeds for every executable operation, including explicit no-data registration.
Detailed discovery requires a behavior entry for every executable command and
rejects every unused behavior entry, proving the behavior-manifest equality.
The evidence-matrix and documentation inventories independently contain the same
131 IDs. Appendix A records every ID and argv path with all fourteen required cells.

There are 274 current schema files (131 inputs, 131 responses, 12 shared) and
272 retained 1.0.0 files: **546 unique schema IDs** in total. Appendix B lists
all files and IDs. All compile with Draft 2020-12 format assertions; every
reference resolves. Capability IDs/paths/input/response/error/stdin references
match the command catalog. The durable publication-plan schema keeps its
original 1.0.0 ID, as required by compatibility.

The compact capability golden is unchanged. Detailed/embedded capabilities and
the capability schema remove only the unreachable explain stateless branch.
The explain input, shared semantic language and capture stdin schemas contain
the selector/parser repairs. The generated evidence matrix now includes the
new regressions plus the existing explain and retained-schema replay scenarios
that had been added since its last generation.

Cobra's hidden `__complete`/`__completeNoDesc` shell protocol remains human-only:
JSON mode returns the published root `command.not_executable` failure, exit 2,
without running completion or advertising a private schema. Both names and JSON
flag placement are covered by permanent tests and process probes.

## Findings and closure

### F1 — Explain project selection was not usable from its schema

Criterion: ARG-02, ARG-03, ARG-07, SEL-01, SEL-03, SEL-05, DOC-01, GEN-04.
Commands: explain.
Evidence: [scalar selector generation](../../cmd/schemagen/explain.go#L33), [runtime project resolver](../../ops/workflows/explain/commands.go#L257);
[generated explain input schema](../../schemas/cli/1.1.0/explain.input.schema.json#L165);
`TestExplainAuditInputBoundaries` and
`TestExplainAuditProjectSelectionAndSourcePrecedence`.
Observed: string `options.project` placed a target-filter reference under
array-only `items`. It accepted empty mode queries such as `=` and published
neither its single-project resolution nor numeric project-number/capture rules.
Required: effective constraints and the actual command-local matching,
normalization, defaults, composition, ambiguity and lookup boundaries.
Remedy class: schema, documentation, test, generated.
Retest: the scalar references the existing positional-target definition;
`explain_project_resolution` and `explain_capture_entry_resolution` publish
number/ID/alias/display/fallback rules, visibility, exact parameter lookup,
original HAR indices and captured namespace/project agreement. Empty target
queries and entry overflow are rejected; zero/one/multiple lookup, exact key
case, wrong-project and source-precedence envelopes conform. **Closed.**

### F2 — Capture schemas rejected supported decoder inputs

Criterion: ARG-02, ARG-07, STDIN-01, STDIN-02, STDIN-03, DOC-01, GEN-04.
Commands: explain; shared stdin:capture resource.
Evidence: [HAR decoder](../../core/rc/evaluate/input.go#L104), [cURL decoder](../../core/rc/evaluate/curl.go#L11),
[capture schema generation](../../cmd/schemagen/explain.go#L9) and
[generated capture stdin schema](../../schemas/cli/1.1.0/stdin.capture.schema.json#L2);
`TestExplainAuditInputBoundaries` and
`TestExplainAuditCaptureFailuresAndEntryBounds`.
Observed: a leading-token pattern rejected a quoted curl executable; HAR
schemas required omitted decoder fields, rejected nullable zero values,
required absolute URLs and fetch-only status/encoding rules on unrelated
traffic, and imposed strict date-time syntax narrower than Go's decoder
(for example, a single-digit hour). Case-insensitive HAR field names and the
capture parser/runtime byte boundary were not modeled adequately. Source
precedence was described in prose without a machine-readable selection rule.
Required: concrete stdin shapes and parser/selection boundaries agree with the
actual decoder and leave ignored input ineffective.
Remedy class: schema, documentation, test, generated.
Retest: concrete nullable/optional HAR fields describe the decoder;
`capture_document` publishes format detection, case-insensitive Go decoding,
HAR timestamp decoding, selected-fetch body/status validation, candidate existence and
32 MiB byte/gzip budgets. Strings are syntax-checked at that annotated runtime
boundary. `explain_capture_source` publishes file/stdin precedence. Accepted
quoted cURL, omitted/null/case-variant HAR, unrelated traffic and Go timestamp
cases conform; malformed/null/array/no-candidate/bad-body cases produce typed
capture.invalid, exit 8. File input leaves piped stdin unconsumed. **Closed.**

### F3 — Indexed override whitespace differed from the evaluator lexer

Criterion: ARG-02, ARG-03, DOC-01, GEN-04.
Commands: explain.
Evidence: [field parser](../../core/rc/evaluate/answers.go#L11), [lexer](../../core/rc/evaluate/expression.go#L194),
[override schema pattern](../../cmd/schemagen/explain.go#L43);
[generated assignment validator](../../schemas/cli/1.1.0/explain.input.schema.json#L178);
`TestExplainAuditInputBoundaries`.
Observed: supported leading whitespace on indexed assignments was rejected
by the anchored schema pattern; the pattern's whitespace class also excluded
vertical tabs accepted by the lexer.
Required: accepted assignment grammar matches runtime, with complete semantic
validation and no silent trimming of scalar fields or values.
Remedy class: schema, documentation, test, generated.
Retest: the lexical pattern accepts indexed whitespace including vertical tabs.
The existing `evaluate_input_assignment` rule retains the complete field/value,
canonical-key, marker, finite-percentile, audience, timestamp and repeated-value
validation. Escaped indexed keys are included as a preserved-behavior regression.
Runtime success and normalized input both validate. **Closed.**

### F4 — Explain advertised a stateless-only network path it cannot execute

Criterion: BEH-01, BEH-02, BEH-05, BEH-06, DOC-01, GEN-04.
Commands: explain; cross-command metadata gate.
Evidence: [cache-read behavior builder](../../ops/contract/capability_behavior.go#L103),
[detailed explain capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860) and
[generated capability schema](../../schemas/cli/1.1.0/capability.schema.json#L2);
`TestContractAuditMetadataHasNoInapplicableBranches`.
Observed: the shared app-cache capability builder added `stateless=true`
network/Firebase-read predicates to explain despite `supports.stateless=false`
and an input schema constrained to stateless=false.
Required: reachable effect/network predicates agree with supported invocations.
Remedy class: capability, test, generated.
Retest: the shared builder explicitly receives stateless support. Existing app
commands retain their behavior; explain no longer advertises the unreachable
branch. The permanent gate checks every command's network, side-effect,
interaction, destructiveness, option-effectiveness and idempotency predicates,
and rejects scalar schemas hiding constraints in `items`. **Closed.**

## Criterion results

Every one of the frozen 51 criteria passes. Each command's applicable evidence
classes and N/A reasons are linked in Appendix A; the catalog resolves every
named test and replay scenario. The following evidence groups incorporate those
complete per-command records, rather than treating schema file counts as proof.

| Criterion | Result and regression evidence |
| --- | --- |
| INV-01 | PASS — Live command/DTO/behavior discovery, compact and detailed goldens, documentation/evidence inventories, schema lookup and 981 process envelopes. `TestEveryExecutableCommandHasCapabilityAndPublishedSchemas`; `TestEveryExecutableCommandHasDocumentationInventoryEntry`. |
| INV-02 | PASS — Live command/DTO/behavior discovery, compact and detailed goldens, documentation/evidence inventories, schema lookup and 981 process envelopes. `TestEveryExecutableCommandHasCapabilityAndPublishedSchemas`; `TestEveryExecutableCommandHasDocumentationInventoryEntry`. |
| INV-03 | PASS — Live command/DTO/behavior discovery, compact and detailed goldens, documentation/evidence inventories, schema lookup and 981 process envelopes. `TestEveryExecutableCommandHasCapabilityAndPublishedSchemas`; `TestEveryExecutableCommandHasDocumentationInventoryEntry`. |
| INV-04 | PASS — Live command/DTO/behavior discovery, compact and detailed goldens, documentation/evidence inventories, schema lookup and 981 process envelopes. `TestEveryExecutableCommandHasCapabilityAndPublishedSchemas`; `TestEveryExecutableCommandHasDocumentationInventoryEntry`. |
| ARG-01 | PASS — Cobra arity, unknown options, normalized invocation/dependency/boundary/effectiveness suites; F1/F2/F3 regressions; runtime entry overflow and presence handling. |
| ARG-02 | PASS — Cobra arity, unknown options, normalized invocation/dependency/boundary/effectiveness suites; F1/F2/F3 regressions; runtime entry overflow and presence handling. |
| ARG-03 | PASS — Cobra arity, unknown options, normalized invocation/dependency/boundary/effectiveness suites; F1/F2/F3 regressions; runtime entry overflow and presence handling. |
| ARG-04 | PASS — Cobra arity, unknown options, normalized invocation/dependency/boundary/effectiveness suites; F1/F2/F3 regressions; runtime entry overflow and presence handling. |
| ARG-05 | PASS — Cobra arity, unknown options, normalized invocation/dependency/boundary/effectiveness suites; F1/F2/F3 regressions; runtime entry overflow and presence handling. |
| ARG-06 | PASS — Cobra arity, unknown options, normalized invocation/dependency/boundary/effectiveness suites; F1/F2/F3 regressions; runtime entry overflow and presence handling. |
| ARG-07 | PASS — Cobra arity, unknown options, normalized invocation/dependency/boundary/effectiveness suites; F1/F2/F3 regressions; runtime entry overflow and presence handling. |
| SEL-01 | PASS — Command-local selector and composition suites, literal no-trim and target-prefix tests, visibility regressions, and F1 project/number/ambiguity/capture consistency cases. |
| SEL-02 | PASS — Command-local selector and composition suites, literal no-trim and target-prefix tests, visibility regressions, and F1 project/number/ambiguity/capture consistency cases. |
| SEL-03 | PASS — Command-local selector and composition suites, literal no-trim and target-prefix tests, visibility regressions, and F1 project/number/ambiguity/capture consistency cases. |
| SEL-04 | PASS — Command-local selector and composition suites, literal no-trim and target-prefix tests, visibility regressions, and F1 project/number/ambiguity/capture consistency cases. |
| SEL-05 | PASS — Command-local selector and composition suites, literal no-trim and target-prefix tests, visibility regressions, and F1 project/number/ambiguity/capture consistency cases. |
| STDIN-01 | PASS — Published concrete stdin/domain/restriction suites; F2 nullable/omitted/case-insensitive HAR, unrelated traffic, quoted cURL and parser-boundary regressions; file-source precedence and typed failures. |
| STDIN-02 | PASS — Published concrete stdin/domain/restriction suites; F2 nullable/omitted/case-insensitive HAR, unrelated traffic, quoted cURL and parser-boundary regressions; file-source precedence and typed failures. |
| STDIN-03 | PASS — Published concrete stdin/domain/restriction suites; F2 nullable/omitted/case-insensitive HAR, unrelated traffic, quoted cURL and parser-boundary regressions; file-source precedence and typed failures. |
| STDIN-04 | PASS — Published concrete stdin/domain/restriction suites; F2 nullable/omitted/case-insensitive HAR, unrelated traffic, quoted cURL and parser-boundary regressions; file-source precedence and typed failures. |
| OUT-01 | PASS — Actual response DTO/schema conformance, envelope invariants, reachable outcome/status tests, empty/no-op/partial and artifact suites; all replay responses and 981 process envelopes. |
| OUT-02 | PASS — Actual response DTO/schema conformance, envelope invariants, reachable outcome/status tests, empty/no-op/partial and artifact suites; all replay responses and 981 process envelopes. |
| OUT-03 | PASS — Actual response DTO/schema conformance, envelope invariants, reachable outcome/status tests, empty/no-op/partial and artifact suites; all replay responses and 981 process envelopes. |
| OUT-04 | PASS — Actual response DTO/schema conformance, envelope invariants, reachable outcome/status tests, empty/no-op/partial and artifact suites; all replay responses and 981 process envelopes. |
| OUT-05 | PASS — Actual response DTO/schema conformance, envelope invariants, reachable outcome/status tests, empty/no-op/partial and artifact suites; all replay responses and 981 process envelopes. |
| OUT-06 | PASS — Actual response DTO/schema conformance, envelope invariants, reachable outcome/status tests, empty/no-op/partial and artifact suites; all replay responses and 981 process envelopes. |
| ERR-01 | PASS — Typed classification and message-independence tests; batch aggregation, warning details/remediation, safe-text redaction/bounds, source-typed failure and exit-status suites; capture failures. |
| ERR-02 | PASS — Typed classification and message-independence tests; batch aggregation, warning details/remediation, safe-text redaction/bounds, source-typed failure and exit-status suites; capture failures. |
| ERR-03 | PASS — Typed classification and message-independence tests; batch aggregation, warning details/remediation, safe-text redaction/bounds, source-typed failure and exit-status suites; capture failures. |
| ERR-04 | PASS — Typed classification and message-independence tests; batch aggregation, warning details/remediation, safe-text redaction/bounds, source-typed failure and exit-status suites; capture failures. |
| ERR-05 | PASS — Typed classification and message-independence tests; batch aggregation, warning details/remediation, safe-text redaction/bounds, source-typed failure and exit-status suites; capture failures. |
| ERR-06 | PASS — Typed classification and message-independence tests; batch aggregation, warning details/remediation, safe-text redaction/bounds, source-typed failure and exit-status suites; capture failures. |
| BEH-01 | PASS — Exhaustive behavior manifest and predicate/confirmation/stateless/plan tests; cache/network/replay and mutation/publication/hook fixtures; F4 removes unreachable stateless paths. |
| BEH-02 | PASS — Exhaustive behavior manifest and predicate/confirmation/stateless/plan tests; cache/network/replay and mutation/publication/hook fixtures; F4 removes unreachable stateless paths. |
| BEH-03 | PASS — Exhaustive behavior manifest and predicate/confirmation/stateless/plan tests; cache/network/replay and mutation/publication/hook fixtures; F4 removes unreachable stateless paths. |
| BEH-04 | PASS — Exhaustive behavior manifest and predicate/confirmation/stateless/plan tests; cache/network/replay and mutation/publication/hook fixtures; F4 removes unreachable stateless paths. |
| BEH-05 | PASS — Exhaustive behavior manifest and predicate/confirmation/stateless/plan tests; cache/network/replay and mutation/publication/hook fixtures; F4 removes unreachable stateless paths. |
| BEH-06 | PASS — Exhaustive behavior manifest and predicate/confirmation/stateless/plan tests; cache/network/replay and mutation/publication/hook fixtures; F4 removes unreachable stateless paths. |
| INT-01 | PASS — Interaction-required/preview and setup/OAuth/confirmation tests, hidden shell transport rejection, machine-mode no-launch tests, and capture candidate/entry regressions. |
| INT-02 | PASS — Interaction-required/preview and setup/OAuth/confirmation tests, hidden shell transport rejection, machine-mode no-launch tests, and capture candidate/entry regressions. |
| INT-03 | PASS — Interaction-required/preview and setup/OAuth/confirmation tests, hidden shell transport rejection, machine-mode no-launch tests, and capture candidate/entry regressions. |
| INT-04 | PASS — Interaction-required/preview and setup/OAuth/confirmation tests, hidden shell transport rejection, machine-mode no-launch tests, and capture candidate/entry regressions. |
| DOC-01 | PASS — Complete command tree and documentation inventory; command-local source and generated-schema comparison; updated explain selector, parser and precedence descriptions; documented example tests and replay scenarios. |
| DOC-02 | PASS — Complete command tree and documentation inventory; command-local source and generated-schema comparison; updated explain selector, parser and precedence descriptions; documented example tests and replay scenarios. |
| DOC-03 | PASS — Complete command tree and documentation inventory; command-local source and generated-schema comparison; updated explain selector, parser and precedence descriptions; documented example tests and replay scenarios. |
| DOC-04 | PASS — Complete command tree and documentation inventory; command-local source and generated-schema comparison; updated explain selector, parser and precedence descriptions; documented example tests and replay scenarios. |
| GEN-01 | PASS — Two consecutive byte-identical generation passes, all 546 schema compilations/references, exhaustive coverage/golden gates and 15-class matrix, root and E2E tests/vet/lint, final intended-change review. |
| GEN-02 | PASS — Two consecutive byte-identical generation passes, all 546 schema compilations/references, exhaustive coverage/golden gates and 15-class matrix, root and E2E tests/vet/lint, final intended-change review. |
| GEN-03 | PASS — Two consecutive byte-identical generation passes, all 546 schema compilations/references, exhaustive coverage/golden gates and 15-class matrix, root and E2E tests/vet/lint, final intended-change review. |
| GEN-04 | PASS — Two consecutive byte-identical generation passes, all 546 schema compilations/references, exhaustive coverage/golden gates and 15-class matrix, root and E2E tests/vet/lint, final intended-change review. |
| GEN-05 | PASS — Two consecutive byte-identical generation passes, all 546 schema compilations/references, exhaustive coverage/golden gates and 15-class matrix, root and E2E tests/vet/lint, final intended-change review. |

## Runtime, schema, generation and repository evidence

The permanent 15-class evidence matrix is complete for all 131 commands, with
explicit reasons for inapplicable classes. All catalog references exist and
all replay scenarios identify the correct command. Applicable success, empty,
no-op, partial publication, warning, artifact, typed failure, interaction,
selection, effectiveness, stdin and boundary tests pass and validate actual
envelopes. Shared cache/network behavior also has core freshness/fallback tests
(`TestReadParametersCachePreference`), with read/write instrumentation and the
full offline replay suite. Explain's evaluator, capture, guarded-outcome,
provenance, managed-value and short-circuit tests pass.

An additional real-process audit used a fresh isolated home/config/cache/work
root for every invocation, offline mode, no inherited credentials or local
repository configuration, explicit stdin, and an eight-second process limit.
It exercised every command's minimal/bare argv, help, unknown option, invalid
timeout, missing profile and stateless boundary; every stdin command also
received malformed, null, empty-object, array and empty input. Hidden shell
transport spellings were checked with prefix/suffix JSON placement.

All **981 process probes passed**: one JSON document and trailing newline,
published response schema, matching process/embedded status, and no unexpected
internal.contract_violation or internal.unclassified fallback. All **546 schemas**
compiled with Draft 2020-12 format assertions before validating the captured
process envelopes. Populated, batch, artifact and warning behavior is additionally
covered by the permanent fake-service and replay scenarios.

| Gate | Result |
| --- | --- |
| `go test -count=1 ./...` (root module) | PASS |
| `go vet ./...` (root module) | PASS |
| `golangci-lint run` (root module) | PASS, zero issues |
| `go test -count=1 ./...` (separate e2e module, replay mode) | PASS |
| `go vet ./...` (separate e2e module) | PASS |
| `golangci-lint run` (separate e2e module) | PASS, zero issues |
| `go run ./cmd/schemagen`, twice consecutively | PASS, 552 byte-identical generated assets |
| All current/retained schema IDs, references and process envelopes | PASS |
| Fourteen-cell audit records and fifteen-class evidence for every command | PASS |
| Released 1.0.0 schemas and archived lock unchanged | PASS |
| Final intended diff and `git diff --check` | PASS |

The environment initially denied loopback test servers and Go cache access.
The affected checks were rerun successfully with the required access. Tests used
local fakes and replayed responses; this audit made no live Firebase mutations.
The pre-existing untracked test.har was neither read nor modified.

Final contract-lock SHA-256: `914d9576c145a4efce5fb0ba9702012efe94ac182a5ca7846f4c73a674b53b87`.
Sorted path-to-SHA-256 manifest digest (552 generated assets): `4339f667930d7ca631f72d4ae2198d6ee98f961fcbd9facf805eb4a7954fe3fc`.
Captured process-evidence SHA-256: `feaa2f14a1001ea474571849269142e3b07dfb2812816a691fb4e8d002c23074`.

All twelve section-7 acceptance conditions pass, and all four findings are
closed. The final verdict for this audited repair worktree under standard
1.0.0 is **AUTHORITATIVE: PASS**.

## Appendix A — complete command audit records

The following 131 records each contain all fourteen required columns. Linked
schemas and capabilities incorporate complete field definitions, grammar,
normalization, options/defaults/aliases, conditions and reachable variants.
Linked evidence incorporates the full fifteen-class record, named test/scenario
references and explicit N/A reasons. Failure details and statuses also use the
current shared error schema and the machine-contract exit-status rules.

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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L77). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L77). |
| Effects | Level 3; network `conditional`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L77). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:add:input`; `urn:fbrcm:schema:cli:1.1.0:command:add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1399); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L77); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `plan.source_draft_changed`, `publication.cache_stale`, `publication.draft_cleanup_failed`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/apply.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L226). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L226). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L226). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:apply:input`; `urn:fbrcm:schema:cli:1.1.0:command:apply:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1376); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apply`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L226); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/apps_config.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L344). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L344). |
| Effects | Level 2; network `conditional`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L344). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:apps.config:input`; `urn:fbrcm:schema:cli:1.1.0:command:apps.config:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2233); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.config`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L344); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/apps_list.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L445). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L445). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L445). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:apps.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:apps.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2192); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L445); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/apps_show.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L541). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L541). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L541). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:apps.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:apps.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2216); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L541); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L637). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L637). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.gcloud:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.add.gcloud:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3439); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.gcloud`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L637); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L723). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L723). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.google:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.add.google:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3363); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.google`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L723); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L824). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L824). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.oauth:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.add.oauth:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3382); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.oauth`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L824); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L919). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L919). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.add.service-account:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.add.service-account:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3416); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.service-account`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L919); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1014). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1014). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.bind:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.bind:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3518); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.bind`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1014); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1103). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1103). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3505); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1103); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1230). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1230). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3349); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1230); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1320). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1320). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.login:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.login:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3471); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.login`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1320); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1423). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1423). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3495); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1423); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1518). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1518). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1518); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1605). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1605). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1605); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1687). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1687). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.unset:input`; `urn:fbrcm:schema:cli:1.1.0:command:auth.quota-project.unset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.unset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1687); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1774). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1774). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:cache.clear:input`; `urn:fbrcm:schema:cli:1.1.0:command:cache.clear:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3103); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1774); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1864). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1864). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:cache.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:cache.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3079); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1864); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1954). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1954). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:cache.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:cache.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3092); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1954); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2036). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2036). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:capabilities:input`; `urn:fbrcm:schema:cli:1.1.0:command:capabilities:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1229); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `capabilities`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2036); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2121). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2121). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:completion.bash:input`; `urn:fbrcm:schema:cli:1.1.0:command:completion.bash:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3763); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.bash`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2121); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2196). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2196). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:completion.fish:input`; `urn:fbrcm:schema:cli:1.1.0:command:completion.fish:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3763); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.fish`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2196); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2271). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2271). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:completion.powershell:input`; `urn:fbrcm:schema:cli:1.1.0:command:completion.powershell:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3763); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.powershell`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2271); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2346). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2346). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:completion.zsh:input`; `urn:fbrcm:schema:cli:1.1.0:command:completion.zsh:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3763); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.zsh`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2346); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2421). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2421). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2421). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.add:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2421); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2549). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2549). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2549). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2549); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_edit.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2677). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2677). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2677). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.edit:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.edit:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2677); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2805). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2805). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1680); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2805); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_move.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2905). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2905). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2905). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.move:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.move:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.move`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2905); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/conditions_rename.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3033). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3033). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3033). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.rename:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.rename:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3033); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3161). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3161). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1706); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3161); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3261). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3261). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:conditions.validate:input`; `urn:fbrcm:schema:cli:1.1.0:command:conditions.validate:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1812); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3261); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `required`; `external_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3361). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3361). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.edit:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.edit:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3273); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3361); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3437). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3437). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3119); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3437); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.reset`

| Column | Audited content |
| --- | --- |
| Command | `config.reset`; `fbrcm config reset`. |
| Arguments | optional key. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [input schema](../../schemas/cli/1.1.0/config_reset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/config_reset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_reset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_reset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3519). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3519). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.reset:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.reset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3229); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3519); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.set`

| Column | Audited content |
| --- | --- |
| Command | `config.set`; `fbrcm config set`. |
| Arguments | required key, required value (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [input schema](../../schemas/cli/1.1.0/config_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/config_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3609). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3609). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3174); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3609); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.show`

| Column | Audited content |
| --- | --- |
| Command | `config.show`; `fbrcm config show`. |
| Arguments | optional key. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/config_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [input schema](../../schemas/cli/1.1.0/config_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.1.0/config_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.1.0/config_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/config_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3704). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3704). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3137); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3704); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3786). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3786). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:config.validate:input`; `urn:fbrcm:schema:cli:1.1.0:command:config.validate:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3252); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3786); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3876). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3876). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3876). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1644); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3876); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4040). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4040). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:doctor:input`; `urn:fbrcm:schema:cli:1.1.0:command:doctor:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3060); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `doctor`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4040); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4122). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4122). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.change-note:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.change-note:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1921); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.change-note`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4122); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4208). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4208). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.diff:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.diff:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1932); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4208); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4296). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4296). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.discard:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.discard:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1988); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.discard`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4296); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4389). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4389). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1871); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4389); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4483). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4483). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1898); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4483); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.cache_stale`, `publication.draft_cleanup_failed`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/draft_publish.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4565). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4565). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4565). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.publish:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.publish:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1960); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.publish`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4565); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `destination_conflict_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4664). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4664). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:draft.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:draft.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1908); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4664); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/duplicate.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4753). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4753). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4753). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:duplicate:input`; `urn:fbrcm:schema:cli:1.1.0:command:duplicate:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1448); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `duplicate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4753); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4895). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4895). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:experiments.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:experiments.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2315); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4895); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4987). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4987). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:experiments.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:experiments.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2275); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4987); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5087). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5087). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:experiments.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:experiments.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2296); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5087); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `explain`

| Column | Audited content |
| --- | --- |
| Command | `explain`; `fbrcm explain <parameter>`. |
| Arguments | Exactly one required parameter key, untrimmed exact case-sensitive equality across root/groups; first exact match, or `parameter.not_found`. Grammar: [input schema](../../schemas/cli/1.1.0/explain.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860). |
| Options | `--cached`, `--entry`, `--from`, `--include-hidden`, `--no-input`, `--no-local-config`, `--profile`, `--project`, `--set`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition and effectiveness: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); dependencies and bounds: [input schema](../../schemas/cli/1.1.0/explain.input.schema.json). `--no-input` is ineffective in JSON; `--stateless=true` is rejected; `--cached`/`--update` are mutually exclusive by presence. Entry is 0 through signed 64-bit maximum, with runtime lookup. |
| Selection | `explain_project_resolution` publishes raw-number lookup, exact ID/alias/display precedence, fallback modes, visibility, primary-template defaults and capture namespace/project agreement. `explain_capture_entry_resolution` publishes original-index selection and ambiguity. No omitted all-project scope. Overrides use last canonical assignment. Full rules: [input schema](../../schemas/cli/1.1.0/explain.input.schema.json) and shared semantic schema; [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860). |
| Stdin | `json_document` structured transport accepts a capture string or HAR object; CLI stdin is raw cURL/HAR text. Concrete `urn:fbrcm:schema:cli:1.1.0:stdin:capture`. Non-`-` file source wins without consuming stdin; no capture requires project and entry=0. Capture parser/byte-budget boundary and nullable HAR shape: [input schema](../../schemas/cli/1.1.0/explain.input.schema.json). |
| Success | Registered `explanationJSON`: resolved/unresolved selection; literal/in-app-default/managed/unknown values; empty arrays; predicate traces, guarded outcomes, input provenance and nullable capture comparison. Both known and unknown results exit 0. [DTO registration](../../ops/workflows/explain/commands.go); [response schema](../../schemas/cli/1.1.0/explain.response.schema.json). |
| Failure | Closed problem catalog: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860). Includes typed capture validation (exit 8), argument failures (2), project/parameter lookup (6), and interaction (10). Categories/details/retryability and all common environment failures: [response schema](../../schemas/cli/1.1.0/explain.response.schema.json), shared error schema and exit rules. |
| Warnings | `cache.stale` and `explain.capture_version_mismatch`, typed details and non-fatal meaning: [response schema](../../schemas/cli/1.1.0/explain.response.schema.json); runtime cases: [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5183). |
| Interaction | JSON never prompts for facts or executes requests. Multiple fetches with entry=0 return safe capture candidates and `--entry`; OAuth authorization is a separate typed boundary. All triggers: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5183). |
| Effects | Level 2; conditional network and Firebase reads, cache/registry/bootstrap/token writes; non-destructive; idempotent. Reuses shared project/parameter/cache/authentication helpers. No stateless network branch, capture persistence, template mutation or hook execution. Exact predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); shared cache/runtime and replay evidence below. |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:explain:input`; `urn:fbrcm:schema:cli:1.1.0:command:explain:response`; `urn:fbrcm:schema:cli:1.1.0:error`; capture/semantic/envelope/capability resources in Appendix B. |
| Documentation | [CLI explain section](../../docs/CLI.md#L1513); [parameter explanations](../../docs/cli-contract.md#parameter-explanations); explicit documentation inventory entry `explain`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5183); `TestContractAuditMetadataHasNoInapplicableBranches`; all four permanent explain replay scenarios; shared cache-freshness tests and evaluator/capture tests. New boundary, selection, precedence and failure regressions are included in the catalog. |
| Verdict | **PASS**; F1–F4 closed, all applicable criteria and final gates pass. |

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
| Warnings | `cache.stale`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/get.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5375). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18282); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5375). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18282); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5375). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:get:input`; `urn:fbrcm:schema:cli:1.1.0:command:get:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1488); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `get`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5375); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/groups_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5582). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18843); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5582). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18843); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5582). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.add:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5582); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/groups_delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5710). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19670); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5710). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19670); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5710). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5710); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/groups_edit.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5838). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20506); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5838). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20506); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5838). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.edit:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.edit:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5838); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21343); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5966). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21343); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5966). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1826); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5966); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/groups_rename.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6066). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21806); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6066). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21806); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6066). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:groups.rename:input`; `urn:fbrcm:schema:cli:1.1.0:command:groups.rename:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6066); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22629); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6194). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22629); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6194). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:help:input`; `urn:fbrcm:schema:cli:1.1.0:command:help:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3800); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `help`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6194); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22748); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6288). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22748); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6288). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:hooks.fingerprint:input`; `urn:fbrcm:schema:cli:1.1.0:command:hooks.fingerprint:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3318); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.fingerprint`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6288); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22866); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6370). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22866); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6370). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:hooks.status:input`; `urn:fbrcm:schema:cli:1.1.0:command:hooks.status:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3307); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.status`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6370); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22983); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6452). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22983); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6452). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:hooks.trust:input`; `urn:fbrcm:schema:cli:1.1.0:command:hooks.trust:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3324); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.trust`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6452); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23160); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6542). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23160); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6542). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:hooks.untrust:input`; `urn:fbrcm:schema:cli:1.1.0:command:hooks.untrust:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3338); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.untrust`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6542); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23306); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6628). |
| Effects | Level 0; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23306); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6628). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:mcp:input`; `urn:fbrcm:schema:cli:1.1.0:command:mcp:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `mcp`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6628); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23473); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6715). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23473); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6715). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:personalizations.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:personalizations.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2383); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `personalizations.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6715); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23909); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6811). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23909); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6811). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:personalizations.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:personalizations.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2397); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `personalizations.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6811); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24352); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6911). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24352); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6911). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:plan.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:plan.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1358); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `plan.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6911); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24480); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7000). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24480); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7000). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:plan.validate:input`; `urn:fbrcm:schema:cli:1.1.0:command:plan.validate:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1367); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `plan.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7000); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24608); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7089). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24608); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7089). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3531); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7089); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24725); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7167). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24725); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7167). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3575); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7167); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24973); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7260). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24973); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7260). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3535); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7260); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25091); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7342). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25091); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7342). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3565); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7342); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25217); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7424). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25217); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7424). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.rename:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.rename:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3555); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7424); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25388); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7513). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25388); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7513). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:profile.switch:input`; `urn:fbrcm:schema:cli:1.1.0:command:profile.switch:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3545); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.switch`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7513); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25502); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7599). |
| Effects | Level 2; network `required`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25502); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7599). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.defaults:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.defaults:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2097); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.defaults`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7599); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25841); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7738). |
| Effects | Level 2; network `required`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25841); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7738). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.export:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.export:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2077); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.export`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7738); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/project_import.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7877). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7877). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7877). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.import:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.import:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2113); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7877); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `browser_launch_suppressed_and_oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7996). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7996). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.open:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.open:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2069); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.open`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7996); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27420); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8092). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27420); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8092). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8092); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27584); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8179). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27584); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8179). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8179); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27761); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8261). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27761); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8261). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.unset:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.quota-project.unset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.unset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8261); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27954); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8347). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27954); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8347). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2004); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8347); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28318); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8443). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28318); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8443). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.templates.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.templates.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2035); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.templates.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8443); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28496); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8525). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28496); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8525). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:project.templates.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:project.templates.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2025); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.templates.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8525); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28638); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8616). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28638); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8616). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.import:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.import:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2967); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8616); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28867); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8706). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28867); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8706). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2916); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8706); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28987); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8796). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28987); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8796). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.remove:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.remove:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2950); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8796); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29187); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8889). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29187); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8889). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.set:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.aliases.set:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2931); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8889); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29393); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8979). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29393); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8979). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.add:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2755); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8979); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29548); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9098). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29548); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9098). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.clear:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.clear:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2763); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9098); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29684); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9211). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29684); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9211). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.disable:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.disable:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2771); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.disable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9211); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29820); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9324). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29820); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9324). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.enable:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.enable:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2767); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.enable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9324); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29956); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9437). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29956); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9437). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2751); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9437); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30076); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9543). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30076); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9543). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.remove:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.blacklist.remove:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2759); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9543); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30231); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9662). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30231); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9662). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.diff:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.diff:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2848); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9662); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30601); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9762). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30601); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9762). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.forget:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.forget:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2836); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.forget`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9762); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30859); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9856). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30859); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9856). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2799); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9856); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31247); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9986). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31247); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9986). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2991); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9986); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/projects_promote.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10068). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31366); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10068). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31366); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10068). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.promote:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.promote:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2875); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.promote`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10068); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32180); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10182). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32180); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10182). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.reset:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.reset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3001); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10182); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32392); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10272). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32392); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10272). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.update:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.update:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2820); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.update`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10272); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `project.remains_hidden`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/projects_whitelist_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10364). |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32640); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10364). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32640); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10364). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.add:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.add:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2779); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10364); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32795); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10490). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32795); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10490). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.clear:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.clear:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2787); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10490); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32931); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10603). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32931); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10603). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.disable:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.disable:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2795); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.disable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10603); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33067); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10716). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33067); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10716). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.enable:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.enable:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2791); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.enable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10716); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33203); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10829). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33203); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10829). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2775); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10829); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33323); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10935). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33323); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10935). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.remove:input`; `urn:fbrcm:schema:cli:1.1.0:command:projects.whitelist.remove:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2783); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10935); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33478); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11054). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33478); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11054). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:rollouts.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2367); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11054); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33849); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11146). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33849); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11146). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:rollouts.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2333); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11146); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34199); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11242). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34199); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11242). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:rollouts.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:rollouts.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2350); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11242); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34555); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11338). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34555); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11338). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:root:input`; `urn:fbrcm:schema:cli:1.1.0:command:root:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1191); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `root`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11338); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34708); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11430). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34708); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11430). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:schema.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:schema.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1320); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `schema.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11430); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34821); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11512). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34821); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11512). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:schema.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:schema.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1325); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `schema.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11512); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `required`; `guided_setup_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34942); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11606). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34942); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11606). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:setup:input`; `urn:fbrcm:schema:cli:1.1.0:command:setup:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3013); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `setup`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11606); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35082); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11695). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35082); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11695). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3587); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11695); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35200); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11777). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35200); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11777). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.delete:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.delete:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3649); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11777); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `theme.already_exists`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/theme_import.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11860). |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35404); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11860). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35404); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11860). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.import:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.import:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3661); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11860); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12002). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12002). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3596); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12002); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35756); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12084). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35756); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12084). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.path:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.path:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3636); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12084); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35880); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12166). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35880); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12166). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.rename:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.rename:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3641); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12166); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36045); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12255). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36045); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12255). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.reset:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.reset:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3622); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12255); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36201); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12333). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36201); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12333). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:theme.switch:input`; `urn:fbrcm:schema:cli:1.1.0:command:theme.switch:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3607); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.switch`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12333); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/update.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12413). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36372); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12413). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36372); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12413). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:update:input`; `urn:fbrcm:schema:cli:1.1.0:command:update:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1588); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `update`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12413); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
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
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37423); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12562). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37423); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12562). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.blame:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.blame:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2460); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.blame`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12562); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.diff`

| Column | Audited content |
| --- | --- |
| Command | `versions.diff`; `fbrcm versions diff`. |
| Arguments | required project, required from, optional to. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37739). |
| Options | `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--side-by-side`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37739); [input schema](../../schemas/cli/1.1.0/versions_diff.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_diff.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37739). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_diff.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37739); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_diff.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37739); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12653). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37739); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12653). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.diff:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.diff:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2568); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12653); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.export`

| Column | Audited content |
| --- | --- |
| Command | `versions.export`; `fbrcm versions export`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_export.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38215). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38215); [input schema](../../schemas/cli/1.1.0/versions_export.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_export.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38215). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_export.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38215); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_export.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38215); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12763). |
| Effects | Level 2; network `conditional`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38215); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12763). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.export:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.export:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2596); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.export`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12763); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.list`

| Column | Audited content |
| --- | --- |
| Command | `versions.list`; `fbrcm versions list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38685). |
| Options | `--all`, `--before`, `--cached`, `--include-hidden`, `--limit`, `--no-local-config`, `--profile`, `--since`, `--stateless`, `--timeout`, `--until`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38685); [input schema](../../schemas/cli/1.1.0/versions_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38685). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38685); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38685); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12869). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38685); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12869). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.list:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.list:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2523); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12869); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.restore`

| Column | Audited content |
| --- | --- |
| Command | `versions.restore`; `fbrcm versions restore`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_restore.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39036). |
| Options | `--change-note`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39036); [input schema](../../schemas/cli/1.1.0/versions_restore.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_restore.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39036). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_restore.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39036); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_restore.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/versions_restore.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12969). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39036); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12969). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39036); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12969). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.restore:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.restore:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2651); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.restore`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12969); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.rollback`

| Column | Audited content |
| --- | --- |
| Command | `versions.rollback`; `fbrcm versions rollback`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_rollback.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39687). |
| Options | `--dry-run`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39687); [input schema](../../schemas/cli/1.1.0/versions_rollback.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_rollback.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39687). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_rollback.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39687); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_rollback.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.1.0/versions_rollback.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13068). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39687); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13068). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39687); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13068). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.rollback:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.rollback:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2612); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.rollback`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13068); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.show`

| Column | Audited content |
| --- | --- |
| Command | `versions.show`; `fbrcm versions show`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.1.0/versions_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40289). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40289); [input schema](../../schemas/cli/1.1.0/versions_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.1.0/versions_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40289). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.1.0/versions_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40289); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.1.0/versions_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40289); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13171). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L40289); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13171). |
| Schemas | `urn:fbrcm:schema:cli:1.1.0:command:versions.show:input`; `urn:fbrcm:schema:cli:1.1.0:command:versions.show:response`; `urn:fbrcm:schema:cli:1.1.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2553); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L13171); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

## Appendix B — complete schema file and ID inventory

IDs are unique across both retained and current resources. All references resolve
and every schema compiles. Commands and capability schema references are recorded
in Appendix A; historical schema resources are retained for compatibility.

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
