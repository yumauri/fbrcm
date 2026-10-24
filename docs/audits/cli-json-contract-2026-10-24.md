# CLI JSON contract audit — 2026-10-24

Repository revision: `3b5b59d028671cedea49331f0a919c3afccc3fed` plus the audited repair worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `130`
Input schemas: `130`
Response schemas: `130`
Shared schemas: `12`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0` unresolved (`4` found and closed)

The revision, clean starting tree, release state, contract version, and audit
standard were frozen before inspection. No acceptance criteria or versions were
changed. This audit applies the finite 51-criterion standard to the current
Cobra tree, registered DTOs, behavior manifest, schemas, capability goldens,
CLI documentation, runtime tests, and evidence matrix. Previously passing
records were rechecked against current artifacts and the complete regression
suite; changed and dependent branches were exercised again.

## Inventory

`E = C = I = R = D = B = M`, with 130 members in every set and empty symmetric
differences. Appendix A identifies every member and its complete record.
`TestEveryExecutableCommandHasCapabilityAndPublishedSchemas` checks live
command discovery, DTO registration, current response schemas and references;
`TestCapabilityGoldenCoversEveryExecutableCommand` checks compact discovery;
`TestDetailedCapabilityGoldenCoversAuthoritativeBehavior` checks the behavior
manifest; the documentation inventory and evidence matrix have separate
exhaustive tests. The checked-in input, response, documentation and evidence
sets were independently compared during report preparation.

All 272 schema files and IDs are listed in Appendix B. IDs are unique, every
schema compiles with Draft 2020-12 format assertions, and every reference
resolves. Capability paths and input/response IDs identify their matching
commands. The compact capability golden is unchanged. Detailed capabilities
and three response schemas remove only an unreachable stdin problem code.
The semantic schema and theme-import input schema publish the repaired
source rules and directory/name constraint.

The two spellings of Cobra's dynamically installed shell transport,
`__complete` and `__completeNoDesc`, were also inspected and probed. They are
not executable JSON operations: JSON invocation now stops before completion
and returns the published root `command.not_executable` failure with canonical
`requested_command: "__complete"`. This human-only transport is documented
explicitly; it cannot advertise an unpublished response schema. Ordinary
shell completion remains covered by its existing initialization-bypass test.

## Closed findings

### F1 — A validation operator had no published meaning

Criterion: ARG-07, SEL-01, DOC-01, GEN-02, GEN-03.
Commands: theme.import; shared semantic-language generation gate.
Evidence: `cmd/schemagen/main.go`, `extensionLanguageMetadata`;
`cmd/schemagen/extensions.go`; `semantic.schema.json` and
`theme_import.input.schema.json`.
Observed: `theme_import_source` was accepted by the rule AST and emitted by the
input schema, but absent from the extension language's operator definitions.
An agent could not interpret its source/ref validation boundary.
Required: publish typed semantic validation and matching rules consistently.
Remedy class: schema, generated, documentation, test.
Retest: the operator now publishes source/name/ref field paths, decoded-prefix
matching, inferred names, directory restrictions, lookup/pinning behavior, and
filesystem/provider/TOML/budget validation boundaries. Generation rejects any
`operator` alternative, including enum alternatives, without published semantics.
`TestEveryExtensionRuleOperatorPublishesSemantics` tests rejection in validation,
normalization, matching, and input selection. **Closed.**

### F2 — Repository theme input constraints differed from runtime

Criterion: ARG-02, ARG-04, DOC-01, GEN-04.
Commands: theme.import.
Evidence: `cli/commands/theme/import_repository.go`,
`core/themesource/source.go`; theme-import input schema and shared browser-URL
schema definitions; `cli/app/theme_repository_test.go`.
Observed: normalized input accepted `--name` for repository directories even
though runtime rejected it before network access. The `--ref` URL schema also
rejected percent-encoded browser markers that the runtime decodes and accepts.
Required: standard constraints and published grammar match runtime.
Remedy class: schema, generated, documentation, test.
Retest: a shared URL-schema helper accepts the runtime's literal/encoded
markers; a conditional rejects names on GitHub/GitLab directory URLs while
accepting single-file names and unnamed directories.
`TestRepositoryThemeAmbiguityAndRefSchema` validates literal and encoded input
cases; `TestRepositoryThemeInputAndOfflineFailures` executes the encoded
GitHub/GitLab parser boundaries and the directory/name rejection. **Closed.**

### F3 — Non-Remote Config stdin advertised a Remote Config error

Criterion: ERR-03, OUT-04, DOC-01, GEN-04.
Commands: auth.add.oauth, auth.add.service-account, theme.import.
Evidence: `ops/contract/problem_codes.go`, `CommandProblemCodes`;
three command response schemas and detailed capabilities;
`TestJSONStdinProblemsMatchInputDomain`.
Observed: generic stdin support added `stdin.remote_config.invalid` to each
closed problem set even though these commands parse credentials or TOML and
cannot emit that code.
Required: closed codes must be reachable and tested or explicitly reserved.
Remedy class: capability, schema, generated, documentation, test.
Retest: actual invalid stdin retains `auth.credentials_invalid`/exit 4 for
credentials and `theme.invalid`/exit 3 for TOML; every envelope validates and
the unreachable code is absent. Remote Config commands retain their existing
stdin codes. **Closed.**

### F4 — Hidden shell completion emitted an unpublished JSON schema

Criterion: INV-04, OUT-01, OUT-05, INT-01, DOC-03, GEN-04.
Commands: root rejection boundary for __complete and alias __completeNoDesc.
Evidence: real-process completion probes; `cli/app/root.go`;
`ops/contract/contract.go`, `envelopeCommandID`;
`TestJSONShellCompletionTransportUsesPublishedRootFailure`.
Observed: Cobra dynamically installed its shell protocol after registration;
JSON invocation ran completion and emitted a success envelope claiming the
nonexistent `command:__complete:response` schema.
Required: JSON mode must return a published, valid envelope and expose no
unsupported human transport as a machine operation.
Remedy class: runtime, documentation, test, generated evidence.
Retest: both spellings and prefix/suffix JSON placements return root-schema
`command.not_executable`, exit 2, null data, no completion diagnostics, and no
persisted state. Root response validation also covers the dynamic endpoint.
`TestRootCommandCompletesEmptyTokenWithoutInitialization` still passes for
ordinary shell requests. Eight additional real-process probes validate the
repaired rejection. **Closed.**

## Criterion results

Every criterion below passes. Evidence references incorporate the complete
named tests and each command's applicable 15-class evidence record.

| Criterion | Result and regression evidence |
| --- | --- |
| INV-01 | PASS — Discovery/DTO/schema/docs/evidence inventories; root/help/completion/parser/startup envelope tests and F4 shell-transport regression. |
| INV-02 | PASS — Discovery/DTO/schema/docs/evidence inventories; root/help/completion/parser/startup envelope tests and F4 shell-transport regression. |
| INV-03 | PASS — Discovery/DTO/schema/docs/evidence inventories; root/help/completion/parser/startup envelope tests and F4 shell-transport regression. |
| INV-04 | PASS — Discovery/DTO/schema/docs/evidence inventories; root/help/completion/parser/startup envelope tests and F4 shell-transport regression. |
| ARG-01 | PASS — Exhaustive Cobra arity; semantic invocation, presence/repetition, normalization, enum/boundary, effectiveness and stdin-restriction suites; F1/F2 regressions. |
| ARG-02 | PASS — Exhaustive Cobra arity; semantic invocation, presence/repetition, normalization, enum/boundary, effectiveness and stdin-restriction suites; F1/F2 regressions. |
| ARG-03 | PASS — Exhaustive Cobra arity; semantic invocation, presence/repetition, normalization, enum/boundary, effectiveness and stdin-restriction suites; F1/F2 regressions. |
| ARG-04 | PASS — Exhaustive Cobra arity; semantic invocation, presence/repetition, normalization, enum/boundary, effectiveness and stdin-restriction suites; F1/F2 regressions. |
| ARG-05 | PASS — Exhaustive Cobra arity; semantic invocation, presence/repetition, normalization, enum/boundary, effectiveness and stdin-restriction suites; F1/F2 regressions. |
| ARG-06 | PASS — Exhaustive Cobra arity; semantic invocation, presence/repetition, normalization, enum/boundary, effectiveness and stdin-restriction suites; F1/F2 regressions. |
| ARG-07 | PASS — Exhaustive Cobra arity; semantic invocation, presence/repetition, normalization, enum/boundary, effectiveness and stdin-restriction suites; F1/F2 regressions. |
| SEL-01 | PASS — Command-local selector and project-argument schema suites; workflow resolver, visibility, alias, prefix, repeated/default scope and ambiguity tests. |
| SEL-02 | PASS — Command-local selector and project-argument schema suites; workflow resolver, visibility, alias, prefix, repeated/default scope and ambiguity tests. |
| SEL-03 | PASS — Command-local selector and project-argument schema suites; workflow resolver, visibility, alias, prefix, repeated/default scope and ambiguity tests. |
| SEL-04 | PASS — Command-local selector and project-argument schema suites; workflow resolver, visibility, alias, prefix, repeated/default scope and ambiguity tests. |
| SEL-05 | PASS — Command-local selector and project-argument schema suites; workflow resolver, visibility, alias, prefix, repeated/default scope and ambiguity tests. |
| STDIN-01 | PASS — Concrete Remote Config/credential/theme/plan input schemas; file-before-stdin precedence, typed malformed-input and stdin restriction tests. |
| STDIN-02 | PASS — Concrete Remote Config/credential/theme/plan input schemas; file-before-stdin precedence, typed malformed-input and stdin restriction tests. |
| STDIN-03 | PASS — Concrete Remote Config/credential/theme/plan input schemas; file-before-stdin precedence, typed malformed-input and stdin restriction tests. |
| STDIN-04 | PASS — Concrete Remote Config/credential/theme/plan input schemas; file-before-stdin precedence, typed malformed-input and stdin restriction tests. |
| OUT-01 | PASS — Runtime envelope/registered DTO/schema conformance; empty/no-op/status/count correlations; help/version/artifact byte/digest suites; process probes. |
| OUT-02 | PASS — Runtime envelope/registered DTO/schema conformance; empty/no-op/status/count correlations; help/version/artifact byte/digest suites; process probes. |
| OUT-03 | PASS — Runtime envelope/registered DTO/schema conformance; empty/no-op/status/count correlations; help/version/artifact byte/digest suites; process probes. |
| OUT-04 | PASS — Runtime envelope/registered DTO/schema conformance; empty/no-op/status/count correlations; help/version/artifact byte/digest suites; process probes. |
| OUT-05 | PASS — Runtime envelope/registered DTO/schema conformance; empty/no-op/status/count correlations; help/version/artifact byte/digest suites; process probes. |
| OUT-06 | PASS — Runtime envelope/registered DTO/schema conformance; empty/no-op/status/count correlations; help/version/artifact byte/digest suites; process probes. |
| ERR-01 | PASS — Typed source-classification/problem-detail/retryability/category/exit/redaction tests; batch, warning and post-publication regressions; F3. |
| ERR-02 | PASS — Typed source-classification/problem-detail/retryability/category/exit/redaction tests; batch, warning and post-publication regressions; F3. |
| ERR-03 | PASS — Typed source-classification/problem-detail/retryability/category/exit/redaction tests; batch, warning and post-publication regressions; F3. |
| ERR-04 | PASS — Typed source-classification/problem-detail/retryability/category/exit/redaction tests; batch, warning and post-publication regressions; F3. |
| ERR-05 | PASS — Typed source-classification/problem-detail/retryability/category/exit/redaction tests; batch, warning and post-publication regressions; F3. |
| ERR-06 | PASS — Typed source-classification/problem-detail/retryability/category/exit/redaction tests; batch, warning and post-publication regressions; F3. |
| BEH-01 | PASS — Reviewed behavior manifest and predicate integrity; workflow effects, cache freshness, stateless/stdin suppression, hooks, visibility and theme-import suites. |
| BEH-02 | PASS — Reviewed behavior manifest and predicate integrity; workflow effects, cache freshness, stateless/stdin suppression, hooks, visibility and theme-import suites. |
| BEH-03 | PASS — Reviewed behavior manifest and predicate integrity; workflow effects, cache freshness, stateless/stdin suppression, hooks, visibility and theme-import suites. |
| BEH-04 | PASS — Reviewed behavior manifest and predicate integrity; workflow effects, cache freshness, stateless/stdin suppression, hooks, visibility and theme-import suites. |
| BEH-05 | PASS — Reviewed behavior manifest and predicate integrity; workflow effects, cache freshness, stateless/stdin suppression, hooks, visibility and theme-import suites. |
| BEH-06 | PASS — Reviewed behavior manifest and predicate integrity; workflow effects, cache freshness, stateless/stdin suppression, hooks, visibility and theme-import suites. |
| INT-01 | PASS — Machine confirmation, OAuth, selection, file-picker/editor/setup and preview tests; F4 non-interactive shell boundary. |
| INT-02 | PASS — Machine confirmation, OAuth, selection, file-picker/editor/setup and preview tests; F4 non-interactive shell boundary. |
| INT-03 | PASS — Machine confirmation, OAuth, selection, file-picker/editor/setup and preview tests; F4 non-interactive shell boundary. |
| INT-04 | PASS — Machine confirmation, OAuth, selection, file-picker/editor/setup and preview tests; F4 non-interactive shell boundary. |
| DOC-01 | PASS — Exhaustive documentation inventory; input/capability/response golden agreement and runtime example/selector/input tests; CLI and contract repairs. |
| DOC-02 | PASS — Exhaustive documentation inventory; input/capability/response golden agreement and runtime example/selector/input tests; CLI and contract repairs. |
| DOC-03 | PASS — Exhaustive documentation inventory; input/capability/response golden agreement and runtime example/selector/input tests; CLI and contract repairs. |
| DOC-04 | PASS — Exhaustive documentation inventory; input/capability/response golden agreement and runtime example/selector/input tests; CLI and contract repairs. |
| GEN-01 | PASS — Two byte-identical generation runs; all schemas/reference compilation; exhaustive command/evidence gates; tests, vet and lint for both Go modules. |
| GEN-02 | PASS — Two byte-identical generation runs; all schemas/reference compilation; exhaustive command/evidence gates; tests, vet and lint for both Go modules. |
| GEN-03 | PASS — Two byte-identical generation runs; all schemas/reference compilation; exhaustive command/evidence gates; tests, vet and lint for both Go modules. |
| GEN-04 | PASS — Two byte-identical generation runs; all schemas/reference compilation; exhaustive command/evidence gates; tests, vet and lint for both Go modules. |
| GEN-05 | PASS — Two byte-identical generation runs; all schemas/reference compilation; exhaustive command/evidence gates; tests, vet and lint for both Go modules. |

## Runtime and repository evidence

The evidence golden contains 130 command records × 15 test classes (1,950
cells) and 64 checked test-symbol references. Each applicable
cell cites known tests and/or replay scenario IDs; every inapplicable cell
contains a checked reason. The matrix test rejects missing commands/classes,
stale symbols/scenarios and incorrect applicability decisions. Appendix A links
to each complete record, including every explicit N/A.

Real-process probes used fresh home/config/cache/work directories, offline
mode, no inherited Firebase or repository credentials, no local project
configuration, explicit stdin, and per-process time limits. All 130 executable
operations were probed for minimal argv, missing profile, invalid timeout,
help, stateless mode, and unknown options. Applicable cases also checked
confirmation bypass, explicit empty string options and five stdin boundaries.
Eight additional probes covered both hidden shell transports and JSON placement.

All **982 probes passed**: exactly one JSON document and one trailing newline,
matching process and embedded exit codes, and no `internal.contract_violation`
or `internal.unclassified` fallback. An independent Draft 2020-12 compiler with
format assertions compiled all **272 schemas** and validated all **982 captured
envelopes** against their advertised response schemas. These probes supplement
the populated, empty/no-op, batch, warning, artifact, interaction and failure
scenarios in the permanent unit and replay E2E suites; they do not replace them.

| Gate | Result |
| --- | --- |
| `go test -count=1 ./...` (root module) | PASS |
| `go vet ./...` (root module) | PASS |
| `golangci-lint run` (root module) | PASS, zero issues |
| `go test -count=1 ./...` (separate `e2e/` module, replay mode) | PASS |
| `go vet ./...` (separate `e2e/` module) | PASS |
| `golangci-lint run` (separate `e2e/` module) | PASS, zero issues |
| `go run ./cmd/schemagen`, twice consecutively | PASS, 277 byte-identical generated assets |
| Every published schema/reference and captured process envelope | PASS |
| `git diff --check` and final intended-change review | PASS |

The sandbox initially denied a loopback test server and ordinary Go cache
writes. Those checks were rerun successfully with the required local access;
no product code was changed to bypass those checks. All tests used local fakes
or replayed provider/Firebase responses. No live Firebase mutation was made.

Final contract-lock SHA-256: `8ad4bbd3d590b5142f33389d7ea2476ff6cba0fea3b44fd115a62f0a17b19213`.
The sorted path-to-SHA-256 manifest digest is
`f6b9a738bc64fb6bcd85c1f9aa0a0b19a9801ae6866dbdd9061c83fc688785fd`. This manifest covers the
272 schemas, embedded capabilities, three goldens, and contract lock.

All twelve section-7 acceptance conditions pass. All four criterion-linked
findings are closed. Under the frozen audit standard 1.0.0, the verdict is
**AUTHORITATIVE: PASS** for the audited repair worktree.

## Appendix A — complete command audit records

Each of the 130 records below supplies all fourteen section-4 columns. Linked
schemas and capabilities incorporate their full field definitions, semantic
rules, options, defaults, conditions and variants. Linked evidence incorporates
the full 15-class record with its test/scenario IDs and explicit N/A reasons.
Failure semantics also use the shared [error schema](../../schemas/cli/1.0.0/error.schema.json)
and [exit-status rules](../cli-contract.md#exit-statuses).

### `add`

| Column | Audited content |
| --- | --- |
| Command | `add`; `fbrcm add`. |
| Arguments | required parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--expr`, `--group`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--type`, `--use-in-app-default`, `--value`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); [input schema](../../schemas/cli/1.0.0/add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/add.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/add/commands.go); [response schema](../../schemas/cli/1.0.0/add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L72). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L72). |
| Effects | Level 3; network `conditional`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L72). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:add:input`; `urn:fbrcm:schema:cli:1.0.0:command:add:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1379); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L72); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `apply`

| Column | Audited content |
| --- | --- |
| Command | `apply`; `fbrcm apply`. |
| Arguments | required plan. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/apply.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962). |
| Options | `--dry-run`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); [input schema](../../schemas/cli/1.0.0/apply.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/apply.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:publication_plan`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/apply.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/apply/commands.go); [response schema](../../schemas/cli/1.0.0/apply.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/apply.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `plan.source_draft_changed`, `publication.cache_stale`, `publication.draft_cleanup_failed`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/apply.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L221). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L221). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L962); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L221). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:apply:input`; `urn:fbrcm:schema:cli:1.0.0:command:apply:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1356); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apply`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L221); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `apps.config`

| Column | Audited content |
| --- | --- |
| Command | `apps.config`; `fbrcm apps config`. |
| Arguments | required app. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/apps_config.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`, `--to`, `--update`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); [input schema](../../schemas/cli/1.0.0/apps_config.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/apps_config.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.0.0/apps_config.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/apps_config.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/apps_config.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L339). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L339). |
| Effects | Level 2; network `conditional`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1600); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L339). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:apps.config:input`; `urn:fbrcm:schema:cli:1.0.0:command:apps.config:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2138); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.config`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L339); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `apps.list`

| Column | Audited content |
| --- | --- |
| Command | `apps.list`; `fbrcm apps list`. |
| Arguments | optional project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/apps_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159). |
| Options | `--cached`, `--filter`, `--include-hidden`, `--no-local-config`, `--platform`, `--profile`, `--project`, `--show-deleted`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); [input schema](../../schemas/cli/1.0.0/apps_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/apps_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.0.0/apps_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/apps_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/apps_list.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L440). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L440). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2159); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L440). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:apps.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:apps.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2097); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L440); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `apps.show`

| Column | Audited content |
| --- | --- |
| Command | `apps.show`; `fbrcm apps show`. |
| Arguments | required app. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/apps_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); [input schema](../../schemas/cli/1.0.0/apps_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/apps_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.0.0/apps_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/apps_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/apps_show.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L536). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L536). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2639); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L536). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:apps.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:apps.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2121); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `apps.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L536); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.add.gcloud`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.gcloud`; `fbrcm auth add gcloud`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_add_gcloud.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125). |
| Options | `--include-hidden`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); [input schema](../../schemas/cli/1.0.0/auth_add_gcloud.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/auth_add_gcloud.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_add_gcloud.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_add_gcloud.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L632). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3125); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L632). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3344); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.gcloud`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L632); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.add.google`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.google`; `fbrcm auth add google`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_add_google.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327). |
| Options | `--include-hidden`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); [input schema](../../schemas/cli/1.0.0/auth_add_google.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/auth_add_google.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_add_google.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_add_google.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L718). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3327); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L718). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3268); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.google`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L718); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.add.oauth`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.oauth`; `fbrcm auth add oauth`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526). |
| Options | `--from`, `--include-hidden`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); [input schema](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:oauth_credentials`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_add_oauth.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_add_oauth.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L819). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3526); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L819). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3287); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.oauth`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L819); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.add.service-account`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.service-account`; `fbrcm auth add service-account`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774). |
| Options | `--from`, `--include-hidden`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); [input schema](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:service_account_credentials`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_add_service-account.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_add_service-account.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L914). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3774); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L914). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3321); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.add.service-account`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L914); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.bind`

| Column | Audited content |
| --- | --- |
| Command | `auth.bind`; `fbrcm auth bind`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_bind.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022). |
| Options | `--auth`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); [input schema](../../schemas/cli/1.0.0/auth_bind.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/auth_bind.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_bind.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_bind.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1009). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4022); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1009). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3423); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.bind`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1009); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.delete`

| Column | Audited content |
| --- | --- |
| Command | `auth.delete`; `fbrcm auth delete`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); [input schema](../../schemas/cli/1.0.0/auth_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/auth_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1098). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4186); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1098). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3410); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1098); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.list`

| Column | Audited content |
| --- | --- |
| Command | `auth.list`; `fbrcm auth list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); [input schema](../../schemas/cli/1.0.0/auth_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/auth_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1225). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4408); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1225). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3254); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1225); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.login`

| Column | Audited content |
| --- | --- |
| Command | `auth.login`; `fbrcm auth login`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_login.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528). |
| Options | `--include-hidden`, `--no-local-config`, `--noopen`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); [input schema](../../schemas/cli/1.0.0/auth_login.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/auth_login.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_login.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_login.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1315). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4528); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1315). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.login:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.login:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3376); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.login`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1315); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.path`

| Column | Audited content |
| --- | --- |
| Command | `auth.path`; `fbrcm auth path`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); [input schema](../../schemas/cli/1.0.0/auth_path.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/auth_path.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1418). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4727); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1418). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.path:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3400); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1418); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.quota-project.set`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.set`; `fbrcm auth quota-project set`. |
| Arguments | required auth_id, required quota_project_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_quota-project_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); [input schema](../../schemas/cli/1.0.0/auth_quota-project_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/auth_quota-project_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_quota-project_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_quota-project_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1513). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4856); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1513). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1513); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.quota-project.show`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.show`; `fbrcm auth quota-project show`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_quota-project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); [input schema](../../schemas/cli/1.0.0/auth_quota-project_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/auth_quota-project_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_quota-project_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_quota-project_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1600). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5008); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1600). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1600); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `auth.quota-project.unset`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.unset`; `fbrcm auth quota-project unset`. |
| Arguments | required auth_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); [input schema](../../schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_quota-project_unset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/auth_quota-project_unset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1682). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5170); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1682). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `auth.quota-project.unset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1682); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `cache.clear`

| Column | Audited content |
| --- | --- |
| Command | `cache.clear`; `fbrcm cache clear`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/cache_clear.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348). |
| Options | `--include-hidden`, `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); [input schema](../../schemas/cli/1.0.0/cache_clear.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/cache_clear.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.0.0/cache_clear.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/cache_clear.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1769). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5348); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1769). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:input`; `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3008); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1769); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `cache.list`

| Column | Audited content |
| --- | --- |
| Command | `cache.list`; `fbrcm cache list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/cache_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554). |
| Options | `--include-hidden`, `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); [input schema](../../schemas/cli/1.0.0/cache_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/cache_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.0.0/cache_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/cache_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1859). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5554); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1859). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:cache.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:cache.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2984); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1859); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `cache.path`

| Column | Audited content |
| --- | --- |
| Command | `cache.path`; `fbrcm cache path`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/cache_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683). |
| Options | `--include-hidden`, `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); [input schema](../../schemas/cli/1.0.0/cache_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/cache_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.0.0/cache_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/cache_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1949). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5683); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1949). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:cache.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:cache.path:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2997); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `cache.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1949); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `capabilities`

| Column | Audited content |
| --- | --- |
| Command | `capabilities`; `fbrcm capabilities`. |
| Arguments | optional command (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/capabilities.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); [input schema](../../schemas/cli/1.0.0/capabilities.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/capabilities.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.0.0/capabilities.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/capabilities.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2031). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5812); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2031). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:capabilities:input`; `urn:fbrcm:schema:cli:1.0.0:command:capabilities:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1209); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `capabilities`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2031); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `completion.bash`

| Column | Audited content |
| --- | --- |
| Command | `completion.bash`; `fbrcm completion bash`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/completion_bash.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933). |
| Options | `--include-hidden`, `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); [input schema](../../schemas/cli/1.0.0/completion_bash.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/completion_bash.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/completion_bash.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/completion_bash.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2116). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5933); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2116). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:input`; `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.bash`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2116); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `completion.fish`

| Column | Audited content |
| --- | --- |
| Command | `completion.fish`; `fbrcm completion fish`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/completion_fish.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056). |
| Options | `--include-hidden`, `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); [input schema](../../schemas/cli/1.0.0/completion_fish.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/completion_fish.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/completion_fish.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/completion_fish.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2191). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6056); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2191). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:input`; `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.fish`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2191); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `completion.powershell`

| Column | Audited content |
| --- | --- |
| Command | `completion.powershell`; `fbrcm completion powershell`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/completion_powershell.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179). |
| Options | `--include-hidden`, `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); [input schema](../../schemas/cli/1.0.0/completion_powershell.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/completion_powershell.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/completion_powershell.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/completion_powershell.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2266). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6179); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2266). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:input`; `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.powershell`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2266); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `completion.zsh`

| Column | Audited content |
| --- | --- |
| Command | `completion.zsh`; `fbrcm completion zsh`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/completion_zsh.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302). |
| Options | `--include-hidden`, `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); [input schema](../../schemas/cli/1.0.0/completion_zsh.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/completion_zsh.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/completion_zsh.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/completion_zsh.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2341). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6302); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2341). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:input`; `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `completion.zsh`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2341); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.add`

| Column | Audited content |
| --- | --- |
| Command | `conditions.add`; `fbrcm conditions add`. |
| Arguments | optional project, required name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/conditions_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425). |
| Options | `--change-note`, `--color`, `--draft`, `--dry-run`, `--expression`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--priority`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); [input schema](../../schemas/cli/1.0.0/conditions_add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/conditions_add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/conditions_add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/conditions_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2416). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2416). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6425); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2416). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2416); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.delete`

| Column | Audited content |
| --- | --- |
| Command | `conditions.delete`; `fbrcm conditions delete`. |
| Arguments | optional project, optional condition. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/conditions_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); [input schema](../../schemas/cli/1.0.0/conditions_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/conditions_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/conditions_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/conditions_delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2544). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2544). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7278); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2544). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2544); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.edit`

| Column | Audited content |
| --- | --- |
| Command | `conditions.edit`; `fbrcm conditions edit`. |
| Arguments | required project, required condition. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/conditions_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154). |
| Options | `--change-note`, `--color`, `--draft`, `--dry-run`, `--expression`, `--include-hidden`, `--no-color`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); [input schema](../../schemas/cli/1.0.0/conditions_edit.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/conditions_edit.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_edit.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/conditions_edit.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/conditions_edit.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2672). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2672). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8154); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2672). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2672); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.list`

| Column | Audited content |
| --- | --- |
| Command | `conditions.list`; `fbrcm conditions list`. |
| Arguments | optional project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/conditions_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996). |
| Options | `--cached`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); [input schema](../../schemas/cli/1.0.0/conditions_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/conditions_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/conditions_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2800). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8996); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2800). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1585); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2800); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.move`

| Column | Audited content |
| --- | --- |
| Command | `conditions.move`; `fbrcm conditions move`. |
| Arguments | required project, required condition, required priority. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/conditions_move.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); [input schema](../../schemas/cli/1.0.0/conditions_move.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/conditions_move.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_move.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/conditions_move.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/conditions_move.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2900). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2900). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9477); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2900). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.move`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2900); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.rename`

| Column | Audited content |
| --- | --- |
| Command | `conditions.rename`; `fbrcm conditions rename`. |
| Arguments | required project, required condition, required new_name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/conditions_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); [input schema](../../schemas/cli/1.0.0/conditions_rename.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/conditions_rename.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_rename.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/conditions_rename.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/conditions_rename.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3028). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3028). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10295); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3028). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3028); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.show`

| Column | Audited content |
| --- | --- |
| Command | `conditions.show`; `fbrcm conditions show`. |
| Arguments | required project, required condition. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/conditions_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); [input schema](../../schemas/cli/1.0.0/conditions_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/conditions_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/conditions_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3156). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11113); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3156). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1611); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3156); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `conditions.validate`

| Column | Audited content |
| --- | --- |
| Command | `conditions.validate`; `fbrcm conditions validate`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/conditions_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); [input schema](../../schemas/cli/1.0.0/conditions_validate.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/conditions_validate.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_validate.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/conditions_validate.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3256). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11556); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3256). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1717); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `conditions.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3256); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.edit`

| Column | Audited content |
| --- | --- |
| Command | `config.edit`; `fbrcm config edit`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/config_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840). |
| Options | `--editor`, `--full`, `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); [input schema](../../schemas/cli/1.0.0/config_edit.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/config_edit.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_edit.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/config_edit.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `required`; `external_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3356). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11840); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3356). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.edit:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.edit:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3178); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3356); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.path`

| Column | Audited content |
| --- | --- |
| Command | `config.path`; `fbrcm config path`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/config_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); [input schema](../../schemas/cli/1.0.0/config_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/config_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/config_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3432). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11998); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3432). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.path:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3024); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3432); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.reset`

| Column | Audited content |
| --- | --- |
| Command | `config.reset`; `fbrcm config reset`. |
| Arguments | optional key. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/config_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [input schema](../../schemas/cli/1.0.0/config_reset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/config_reset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_reset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/config_reset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3514). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12127); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3514). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.reset:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.reset:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3134); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3514); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.set`

| Column | Audited content |
| --- | --- |
| Command | `config.set`; `fbrcm config set`. |
| Arguments | required key, required value (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/config_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [input schema](../../schemas/cli/1.0.0/config_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/config_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/config_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3604). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12334); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3604). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.set:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3079); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3604); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.show`

| Column | Audited content |
| --- | --- |
| Command | `config.show`; `fbrcm config show`. |
| Arguments | optional key. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/config_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [input schema](../../schemas/cli/1.0.0/config_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/config_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/config_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3699). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12492); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3699). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3042); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3699); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `config.validate`

| Column | Audited content |
| --- | --- |
| Command | `config.validate`; `fbrcm config validate`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/config_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); [input schema](../../schemas/cli/1.0.0/config_validate.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/config_validate.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_validate.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/config_validate.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3781). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12628); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3781). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.validate:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.validate:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3157); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `config.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3781); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `delete`

| Column | Audited content |
| --- | --- |
| Command | `delete`; `fbrcm delete`. |
| Arguments | optional parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); [input schema](../../schemas/cli/1.0.0/delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/delete.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/delete/commands.go); [response schema](../../schemas/cli/1.0.0/delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3871). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3871). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12758); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3871). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:delete:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1549); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3871); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `doctor`

| Column | Audited content |
| --- | --- |
| Command | `doctor`; `fbrcm doctor`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/doctor.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); [input schema](../../schemas/cli/1.0.0/doctor.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/doctor.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/doctor/commands.go); [response schema](../../schemas/cli/1.0.0/doctor.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/doctor.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4035). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13709); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4035). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:doctor:input`; `urn:fbrcm:schema:cli:1.0.0:command:doctor:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2965); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `doctor`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4035); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.change-note`

| Column | Audited content |
| --- | --- |
| Command | `draft.change-note`; `fbrcm draft change-note`. |
| Arguments | required project, optional text. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/draft_change-note.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928). |
| Options | `--clear`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); [input schema](../../schemas/cli/1.0.0/draft_change-note.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/draft_change-note.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_change-note.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/draft_change-note.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4117). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13928); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4117). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1826); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.change-note`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4117); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.diff`

| Column | Audited content |
| --- | --- |
| Command | `draft.diff`; `fbrcm draft diff`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/draft_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165). |
| Options | `--against`, `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); [input schema](../../schemas/cli/1.0.0/draft_diff.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/draft_diff.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_diff.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/draft_diff.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4203). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14165); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4203). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1837); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4203); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.discard`

| Column | Audited content |
| --- | --- |
| Command | `draft.discard`; `fbrcm draft discard`. |
| Arguments | optional project (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/draft_discard.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516). |
| Options | `--all`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); [input schema](../../schemas/cli/1.0.0/draft_discard.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/draft_discard.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_discard.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/draft_discard.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4291). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14516); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4291). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1893); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.discard`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4291); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.list`

| Column | Audited content |
| --- | --- |
| Command | `draft.list`; `fbrcm draft list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/draft_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); [input schema](../../schemas/cli/1.0.0/draft_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/draft_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/draft_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4384). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14746); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4384). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1776); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4384); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.path`

| Column | Audited content |
| --- | --- |
| Command | `draft.path`; `fbrcm draft path`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/draft_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); [input schema](../../schemas/cli/1.0.0/draft_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/draft_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/draft_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4478). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14890); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4478). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.path:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1803); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4478); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.publish`

| Column | Audited content |
| --- | --- |
| Command | `draft.publish`; `fbrcm draft publish`. |
| Arguments | optional project (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/draft_publish.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009). |
| Options | `--all`, `--change-note`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); [input schema](../../schemas/cli/1.0.0/draft_publish.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/draft_publish.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_publish.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/draft_publish.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.draft_cleanup_failed`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/draft_publish.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4560). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4560). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15009); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4560). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1865); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.publish`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4560); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `draft.show`

| Column | Audited content |
| --- | --- |
| Command | `draft.show`; `fbrcm draft show`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/draft_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--raw`, `--stateless`, `--timeout`, `--to`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); [input schema](../../schemas/cli/1.0.0/draft_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/draft_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/draft_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `destination_conflict_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4659). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15745); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4659). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1813); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `draft.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4659); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `duplicate`

| Column | Audited content |
| --- | --- |
| Command | `duplicate`; `fbrcm duplicate`. |
| Arguments | required source, required target. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/duplicate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); [input schema](../../schemas/cli/1.0.0/duplicate.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/duplicate.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/duplicate/commands.go); [response schema](../../schemas/cli/1.0.0/duplicate.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/duplicate.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/duplicate.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4748). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4748). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15936); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4748). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:duplicate:input`; `urn:fbrcm:schema:cli:1.0.0:command:duplicate:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1428); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `duplicate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4748); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `experiments.delete`

| Column | Audited content |
| --- | --- |
| Command | `experiments.delete`; `fbrcm experiments delete`. |
| Arguments | required project, required experiment_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/experiments_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); [input schema](../../schemas/cli/1.0.0/experiments_delete.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/experiments_delete.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/experiments_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/experiments_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4890). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16771); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4890). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2220); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4890); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `experiments.list`

| Column | Audited content |
| --- | --- |
| Command | `experiments.list`; `fbrcm experiments list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/experiments_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142). |
| Options | `--cached`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); [input schema](../../schemas/cli/1.0.0/experiments_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/experiments_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/experiments_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/experiments_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4982). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17142); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4982). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2180); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4982); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `experiments.show`

| Column | Audited content |
| --- | --- |
| Command | `experiments.show`; `fbrcm experiments show`. |
| Arguments | required project, required experiment_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/experiments_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); [input schema](../../schemas/cli/1.0.0/experiments_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/experiments_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/experiments_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/experiments_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5082). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17504); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5082). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2201); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `experiments.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5082); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `get`

| Column | Audited content |
| --- | --- |
| Command | `get`; `fbrcm get`. |
| Arguments | optional parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/get.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860). |
| Options | `--all`, `--cached`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); [input schema](../../schemas/cli/1.0.0/get.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/get.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/get.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/get/commands.go); [response schema](../../schemas/cli/1.0.0/get.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/get.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `cache.stale`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/get.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5178). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5178). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17860); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5178). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:get:input`; `urn:fbrcm:schema:cli:1.0.0:command:get:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1468); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `get`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5178); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.add`

| Column | Audited content |
| --- | --- |
| Command | `groups.add`; `fbrcm groups add`. |
| Arguments | required name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/groups_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18421). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18421); [input schema](../../schemas/cli/1.0.0/groups_add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/groups_add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18421). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18421); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/groups_add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/groups_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5385). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18421); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5385). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18421); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5385). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.add:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.add:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5385); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.delete`

| Column | Audited content |
| --- | --- |
| Command | `groups.delete`; `fbrcm groups delete`. |
| Arguments | required group. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/groups_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19248). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19248); [input schema](../../schemas/cli/1.0.0/groups_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/groups_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19248). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19248); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/groups_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/groups_delete.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5513). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19248); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5513). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19248); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5513). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5513); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.edit`

| Column | Audited content |
| --- | --- |
| Command | `groups.edit`; `fbrcm groups edit`. |
| Arguments | required group. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/groups_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20084). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--include-hidden`, `--no-description`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20084); [input schema](../../schemas/cli/1.0.0/groups_edit.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/groups_edit.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20084). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_edit.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20084); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/groups_edit.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/groups_edit.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5641). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20084); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5641). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20084); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5641). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.edit`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5641); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.list`

| Column | Audited content |
| --- | --- |
| Command | `groups.list`; `fbrcm groups list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/groups_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20921). |
| Options | `--cached`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20921); [input schema](../../schemas/cli/1.0.0/groups_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/groups_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20921). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20921); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/groups_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20921); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5769). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20921); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5769). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1731); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5769); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `groups.rename`

| Column | Audited content |
| --- | --- |
| Command | `groups.rename`; `fbrcm groups rename`. |
| Arguments | required group, required new_name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/groups_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21384). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21384); [input schema](../../schemas/cli/1.0.0/groups_rename.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/groups_rename.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21384). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_rename.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21384); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/groups_rename.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/groups_rename.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5869). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21384); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5869). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21384); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5869). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `groups.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5869); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `help`

| Column | Audited content |
| --- | --- |
| Command | `help`; `fbrcm help`. |
| Arguments | optional command (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/help.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22207). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22207); [input schema](../../schemas/cli/1.0.0/help.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/help.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22207). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/help.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22207); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/help.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22207); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5997). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22207); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5997). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:help:input`; `urn:fbrcm:schema:cli:1.0.0:command:help:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3705); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `help`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5997); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `hooks.fingerprint`

| Column | Audited content |
| --- | --- |
| Command | `hooks.fingerprint`; `fbrcm hooks fingerprint`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/hooks_fingerprint.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22326). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22326); [input schema](../../schemas/cli/1.0.0/hooks_fingerprint.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/hooks_fingerprint.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.0.0/hooks_fingerprint.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22326); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/hooks_fingerprint.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22326); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6091). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22326); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6091). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:input`; `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3223); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.fingerprint`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6091); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `hooks.status`

| Column | Audited content |
| --- | --- |
| Command | `hooks.status`; `fbrcm hooks status`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/hooks_status.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22444). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22444); [input schema](../../schemas/cli/1.0.0/hooks_status.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/hooks_status.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.0.0/hooks_status.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22444); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/hooks_status.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22444); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6173). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22444); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6173). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:input`; `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3212); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.status`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6173); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `hooks.trust`

| Column | Audited content |
| --- | --- |
| Command | `hooks.trust`; `fbrcm hooks trust`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/hooks_trust.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22561). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22561); [input schema](../../schemas/cli/1.0.0/hooks_trust.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/hooks_trust.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.0.0/hooks_trust.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22561); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/hooks_trust.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22561); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6255). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22561); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6255). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:input`; `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3229); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.trust`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6255); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `hooks.untrust`

| Column | Audited content |
| --- | --- |
| Command | `hooks.untrust`; `fbrcm hooks untrust`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/hooks_untrust.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22738). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22738); [input schema](../../schemas/cli/1.0.0/hooks_untrust.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/hooks_untrust.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.0.0/hooks_untrust.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22738); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/hooks_untrust.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22738); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6345). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22738); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6345). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:input`; `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3243); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `hooks.untrust`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6345); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `mcp`

| Column | Audited content |
| --- | --- |
| Command | `mcp`; `fbrcm mcp`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/mcp.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22884). |
| Options | `--allow-hooks`, `--allow-writes`, `--auth-timeout`, `--browser-auth`, `--confirmation`, `--include-hidden`, `--no-local-config`, `--profile`, `--request-timeout`, `--stateless`, `--timeout`, `--toolsets`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22884); [input schema](../../schemas/cli/1.0.0/mcp.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/mcp.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/mcp.go); [response schema](../../schemas/cli/1.0.0/mcp.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22884); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/mcp.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22884); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6431). |
| Effects | Level 0; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22884); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6431). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:mcp:input`; `urn:fbrcm:schema:cli:1.0.0:command:mcp:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `mcp`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6431); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `personalizations.list`

| Column | Audited content |
| --- | --- |
| Command | `personalizations.list`; `fbrcm personalizations list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/personalizations_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23051). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23051); [input schema](../../schemas/cli/1.0.0/personalizations_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/personalizations_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/personalizations_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23051); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/personalizations_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23051); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6518). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23051); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6518). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2288); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `personalizations.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6518); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `personalizations.show`

| Column | Audited content |
| --- | --- |
| Command | `personalizations.show`; `fbrcm personalizations show`. |
| Arguments | required project, required personalization_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/personalizations_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23487). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23487); [input schema](../../schemas/cli/1.0.0/personalizations_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/personalizations_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23487). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/personalizations_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23487); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/personalizations_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23487); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6614). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23487); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6614). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2302); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `personalizations.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6614); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `plan.show`

| Column | Audited content |
| --- | --- |
| Command | `plan.show`; `fbrcm plan show`. |
| Arguments | required plan. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/plan_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23930). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23930); [input schema](../../schemas/cli/1.0.0/plan_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/plan_show.input.schema.json). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:publication_plan`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/plan_show.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/plan/commands.go); [response schema](../../schemas/cli/1.0.0/plan_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23930); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/plan_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23930); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6714). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23930); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6714). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:plan.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:plan.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1338); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `plan.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6714); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `plan.validate`

| Column | Audited content |
| --- | --- |
| Command | `plan.validate`; `fbrcm plan validate`. |
| Arguments | required plan. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/plan_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24058). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24058); [input schema](../../schemas/cli/1.0.0/plan_validate.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/plan_validate.input.schema.json). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:publication_plan`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/plan_validate.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/plan/commands.go); [response schema](../../schemas/cli/1.0.0/plan_validate.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24058); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/plan_validate.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24058); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6803). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24058); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6803). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:plan.validate:input`; `urn:fbrcm:schema:cli:1.0.0:command:plan.validate:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1347); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `plan.validate`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6803); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile`

| Column | Audited content |
| --- | --- |
| Command | `profile`; `fbrcm profile`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/profile.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24186). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24186); [input schema](../../schemas/cli/1.0.0/profile.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/profile.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24186); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/profile.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24186); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6892). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24186); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6892). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3436); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6892); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.delete`

| Column | Audited content |
| --- | --- |
| Command | `profile.delete`; `fbrcm profile delete`. |
| Arguments | required profile. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/profile_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24303). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24303); [input schema](../../schemas/cli/1.0.0/profile_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/profile_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24303). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24303); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/profile_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24303); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6970). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24303); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6970). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3480); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6970); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.list`

| Column | Audited content |
| --- | --- |
| Command | `profile.list`; `fbrcm profile list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/profile_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24551). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24551); [input schema](../../schemas/cli/1.0.0/profile_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/profile_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24551); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/profile_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24551); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7063). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24551); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7063). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3440); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7063); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.path`

| Column | Audited content |
| --- | --- |
| Command | `profile.path`; `fbrcm profile path`. |
| Arguments | required profile. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/profile_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24669). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24669); [input schema](../../schemas/cli/1.0.0/profile_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/profile_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24669); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/profile_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24669); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7145). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24669); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7145). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.path:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3470); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7145); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.rename`

| Column | Audited content |
| --- | --- |
| Command | `profile.rename`; `fbrcm profile rename`. |
| Arguments | required old_name, required new_name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/profile_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24795). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24795); [input schema](../../schemas/cli/1.0.0/profile_rename.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/profile_rename.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24795). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_rename.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24795); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/profile_rename.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24795); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7227). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24795); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7227). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3460); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7227); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `profile.switch`

| Column | Audited content |
| --- | --- |
| Command | `profile.switch`; `fbrcm profile switch`. |
| Arguments | required name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/profile_switch.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24966). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24966); [input schema](../../schemas/cli/1.0.0/profile_switch.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/profile_switch.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_switch.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24966); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/profile_switch.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24966); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7316). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24966); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7316). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3450); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `profile.switch`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7316); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.defaults`

| Column | Audited content |
| --- | --- |
| Command | `project.defaults`; `fbrcm project defaults`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_defaults.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25080). |
| Options | `--format`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25080); [input schema](../../schemas/cli/1.0.0/project_defaults.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/project_defaults.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25080). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_defaults.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25080); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_defaults.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25080); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7402). |
| Effects | Level 2; network `required`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25080); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7402). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2002); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.defaults`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7402); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.export`

| Column | Audited content |
| --- | --- |
| Command | `project.export`; `fbrcm project export`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_export.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25419). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25419); [input schema](../../schemas/cli/1.0.0/project_export.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/project_export.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25419). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_export.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25419); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_export.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25419); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7541). |
| Effects | Level 2; network `required`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25419); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7541). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.export:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.export:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1982); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.export`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7541); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.import`

| Column | Audited content |
| --- | --- |
| Command | `project.import`; `fbrcm project import`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25748). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--from`, `--group`, `--include-hidden`, `--keep-portable-conditions-only`, `--merge`, `--merge-resolve`, `--no-local-config`, `--override`, `--plan-out`, `--profile`, `--remove-all-conditions`, `--search`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25748); [input schema](../../schemas/cli/1.0.0/project_import.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/project_import.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25748). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config_import`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/project_import.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_import.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25748); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_import.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/project_import.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7680). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25748); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7680). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25748); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7680). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.import:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.import:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2018); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7680); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.open`

| Column | Audited content |
| --- | --- |
| Command | `project.open`; `fbrcm project open`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_open.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26720). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26720); [input schema](../../schemas/cli/1.0.0/project_open.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/project_open.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_open.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26720); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_open.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `browser_launch_suppressed_and_oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26720); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7799). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26720); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7799). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.open:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.open:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1974); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.open`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7799); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.quota-project.set`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.set`; `fbrcm project quota-project set`. |
| Arguments | required project, required quota_project_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_quota-project_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26998). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26998); [input schema](../../schemas/cli/1.0.0/project_quota-project_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/project_quota-project_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_quota-project_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26998); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_quota-project_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26998); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7895). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26998); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7895). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7895); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.quota-project.show`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.show`; `fbrcm project quota-project show`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_quota-project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27162). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27162); [input schema](../../schemas/cli/1.0.0/project_quota-project_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/project_quota-project_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_quota-project_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27162); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_quota-project_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27162); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7982). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27162); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7982). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7982); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.quota-project.unset`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.unset`; `fbrcm project quota-project unset`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_quota-project_unset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27339). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27339); [input schema](../../schemas/cli/1.0.0/project_quota-project_unset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/project_quota-project_unset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_quota-project_unset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27339); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_quota-project_unset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27339); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8064). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27339); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8064). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.quota-project.unset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8064); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.show`

| Column | Audited content |
| --- | --- |
| Command | `project.show`; `fbrcm project show`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27532). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27532); [input schema](../../schemas/cli/1.0.0/project_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/project_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27532); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27532); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8150). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27532); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8150). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1909); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8150); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.templates.set`

| Column | Audited content |
| --- | --- |
| Command | `project.templates.set`; `fbrcm project templates set`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_templates_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27896). |
| Options | `--include-hidden`, `--no-local-config`, `--primary`, `--profile`, `--stateless`, `--templates`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27896); [input schema](../../schemas/cli/1.0.0/project_templates_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/project_templates_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_templates_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27896); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_templates_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27896); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8246). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27896); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8246). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1940); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.templates.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8246); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `project.templates.show`

| Column | Audited content |
| --- | --- |
| Command | `project.templates.show`; `fbrcm project templates show`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/project_templates_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28074). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28074); [input schema](../../schemas/cli/1.0.0/project_templates_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/project_templates_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_templates_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28074); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/project_templates_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28074); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8328). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28074); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8328). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1930); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `project.templates.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8328); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.aliases.import`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.import`; `fbrcm projects aliases import`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_aliases_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28216). |
| Options | `--conflict`, `--dry-run`, `--from`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28216); [input schema](../../schemas/cli/1.0.0/projects_aliases_import.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_aliases_import.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_aliases_import.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28216); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_aliases_import.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28216); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8419). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28216); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8419). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2872); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8419); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.aliases.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.list`; `fbrcm projects aliases list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_aliases_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28445). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28445); [input schema](../../schemas/cli/1.0.0/projects_aliases_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_aliases_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_aliases_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28445); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_aliases_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28445); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8509). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28445); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8509). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2821); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8509); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.aliases.remove`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.remove`; `fbrcm projects aliases remove`. |
| Arguments | required alias. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_aliases_remove.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28565). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28565); [input schema](../../schemas/cli/1.0.0/projects_aliases_remove.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_aliases_remove.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28565). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_aliases_remove.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28565); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_aliases_remove.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28565); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8599). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28565); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8599). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2855); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8599); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.aliases.set`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.set`; `fbrcm projects aliases set`. |
| Arguments | required alias, required project_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_aliases_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28765). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28765); [input schema](../../schemas/cli/1.0.0/projects_aliases_set.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_aliases_set.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_aliases_set.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28765); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_aliases_set.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28765); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8692). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28765); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8692). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2836); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.aliases.set`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8692); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.add`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.add`; `fbrcm projects blacklist add`. |
| Arguments | optional project_id (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_blacklist_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28971). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28971); [input schema](../../schemas/cli/1.0.0/projects_blacklist_add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_blacklist_add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28971). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_blacklist_add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28971); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_blacklist_add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28971); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8782). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28971); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8782). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.add:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.add:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2660); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8782); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.clear`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.clear`; `fbrcm projects blacklist clear`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_blacklist_clear.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29126). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29126); [input schema](../../schemas/cli/1.0.0/projects_blacklist_clear.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_blacklist_clear.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_blacklist_clear.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29126); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_blacklist_clear.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29126); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8901). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29126); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8901). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.clear:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.clear:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2668); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8901); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.disable`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.disable`; `fbrcm projects blacklist disable`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_blacklist_disable.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29262). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29262); [input schema](../../schemas/cli/1.0.0/projects_blacklist_disable.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_blacklist_disable.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_blacklist_disable.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29262); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_blacklist_disable.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29262); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9014). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29262); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9014). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.disable:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.disable:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2676); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.disable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9014); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.enable`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.enable`; `fbrcm projects blacklist enable`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_blacklist_enable.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29398). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29398); [input schema](../../schemas/cli/1.0.0/projects_blacklist_enable.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_blacklist_enable.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_blacklist_enable.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29398); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_blacklist_enable.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29398); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9127). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29398); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9127). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.enable:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.enable:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2672); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.enable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9127); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.list`; `fbrcm projects blacklist list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_blacklist_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29534). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29534); [input schema](../../schemas/cli/1.0.0/projects_blacklist_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_blacklist_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_blacklist_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29534); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_blacklist_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29534); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9240). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29534); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9240). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2656); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9240); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.blacklist.remove`

| Column | Audited content |
| --- | --- |
| Command | `projects.blacklist.remove`; `fbrcm projects blacklist remove`. |
| Arguments | optional project_id (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_blacklist_remove.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29654). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29654); [input schema](../../schemas/cli/1.0.0/projects_blacklist_remove.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_blacklist_remove.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29654). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_blacklist_remove.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29654); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_blacklist_remove.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29654); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9346). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29654); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9346). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.remove:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.remove:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2664); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.blacklist.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9346); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.diff`

| Column | Audited content |
| --- | --- |
| Command | `projects.diff`; `fbrcm projects diff`. |
| Arguments | required source_project, required target_project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29809). |
| Options | `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29809); [input schema](../../schemas/cli/1.0.0/projects_diff.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_diff.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29809). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_diff.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29809); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_diff.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29809); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9465). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29809); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9465). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2753); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9465); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.forget`

| Column | Audited content |
| --- | --- |
| Command | `projects.forget`; `fbrcm projects forget`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_forget.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30179). |
| Options | `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30179); [input schema](../../schemas/cli/1.0.0/projects_forget.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_forget.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30179). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_forget.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30179); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_forget.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30179); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9565). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30179); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9565). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2741); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.forget`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9565); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.list`; `fbrcm projects list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30437). |
| Options | `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`, `--url`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30437); [input schema](../../schemas/cli/1.0.0/projects_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30437). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30437); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30437); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9659). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30437); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9659). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2704); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9659); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.path`

| Column | Audited content |
| --- | --- |
| Command | `projects.path`; `fbrcm projects path`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30825). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30825); [input schema](../../schemas/cli/1.0.0/projects_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30825); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30825); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9789). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30825); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9789). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.path:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2896); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9789); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.promote`

| Column | Audited content |
| --- | --- |
| Command | `projects.promote`; `fbrcm projects promote`. |
| Arguments | required source_project, required target_project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_promote.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30944). |
| Options | `--all`, `--change-note`, `--conditions`, `--dry-run`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--interactive`, `--no-local-config`, `--parameters`, `--plan-out`, `--profile`, `--prune`, `--search`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30944); [input schema](../../schemas/cli/1.0.0/projects_promote.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_promote.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30944). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_promote.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30944); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_promote.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/projects_promote.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9871). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30944); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9871). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30944); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9871). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2780); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.promote`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9871); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.reset`

| Column | Audited content |
| --- | --- |
| Command | `projects.reset`; `fbrcm projects reset`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31758). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31758); [input schema](../../schemas/cli/1.0.0/projects_reset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_reset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_reset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31758); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_reset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31758); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9985). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31758); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9985). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2906); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9985); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.update`

| Column | Audited content |
| --- | --- |
| Command | `projects.update`; `fbrcm projects update`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_update.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31970). |
| Options | `--auth`, `--expr`, `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--url`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31970); [input schema](../../schemas/cli/1.0.0/projects_update.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_update.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31970). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_update.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31970); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_update.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31970); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10075). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `no`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31970); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10075). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.update:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.update:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2725); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.update`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10075); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.add`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.add`; `fbrcm projects whitelist add`. |
| Arguments | optional project_id (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_whitelist_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32218). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32218); [input schema](../../schemas/cli/1.0.0/projects_whitelist_add.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_whitelist_add.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32218). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_whitelist_add.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32218); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_whitelist_add.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `project.remains_hidden`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/projects_whitelist_add.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10167). |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32218); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10167). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32218); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10167). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.add:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.add:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2684); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.add`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10167); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.clear`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.clear`; `fbrcm projects whitelist clear`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_whitelist_clear.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32373). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32373); [input schema](../../schemas/cli/1.0.0/projects_whitelist_clear.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_whitelist_clear.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_whitelist_clear.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32373); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_whitelist_clear.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32373); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10293). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32373); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10293). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.clear:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.clear:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2692); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.clear`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10293); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.disable`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.disable`; `fbrcm projects whitelist disable`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_whitelist_disable.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32509). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32509); [input schema](../../schemas/cli/1.0.0/projects_whitelist_disable.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_whitelist_disable.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_whitelist_disable.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32509); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_whitelist_disable.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32509); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10406). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32509); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10406). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.disable:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.disable:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2700); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.disable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10406); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.enable`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.enable`; `fbrcm projects whitelist enable`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_whitelist_enable.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32645). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32645); [input schema](../../schemas/cli/1.0.0/projects_whitelist_enable.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_whitelist_enable.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_whitelist_enable.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32645); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_whitelist_enable.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32645); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10519). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32645); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10519). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.enable:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.enable:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2696); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.enable`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10519); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.list`; `fbrcm projects whitelist list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_whitelist_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32781). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32781); [input schema](../../schemas/cli/1.0.0/projects_whitelist_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/projects_whitelist_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_whitelist_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32781); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_whitelist_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32781); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10632). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32781); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10632). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2680); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10632); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `projects.whitelist.remove`

| Column | Audited content |
| --- | --- |
| Command | `projects.whitelist.remove`; `fbrcm projects whitelist remove`. |
| Arguments | optional project_id (repeatable). Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/projects_whitelist_remove.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32901). |
| Options | `--filter`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32901); [input schema](../../schemas/cli/1.0.0/projects_whitelist_remove.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/projects_whitelist_remove.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32901). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/projects/visibility.go); [response schema](../../schemas/cli/1.0.0/projects_whitelist_remove.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32901); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/projects_whitelist_remove.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32901); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10738). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32901); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10738). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.remove:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.remove:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2688); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `projects.whitelist.remove`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10738); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `rollouts.delete`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.delete`; `fbrcm rollouts delete`. |
| Arguments | required project, required rollout_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/rollouts_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33056). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33056); [input schema](../../schemas/cli/1.0.0/rollouts_delete.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/rollouts_delete.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/rollouts_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33056); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/rollouts_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33056); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10857). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33056); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10857). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2272); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10857); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `rollouts.list`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.list`; `fbrcm rollouts list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/rollouts_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33427). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33427); [input schema](../../schemas/cli/1.0.0/rollouts_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/rollouts_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/rollouts_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33427); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/rollouts_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33427); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10949). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33427); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10949). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2238); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10949); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `rollouts.show`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.show`; `fbrcm rollouts show`. |
| Arguments | required project, required rollout_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/rollouts_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33777). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33777); [input schema](../../schemas/cli/1.0.0/rollouts_show.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/rollouts_show.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/rollouts_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33777); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/rollouts_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33777); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11045). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33777); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11045). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2255); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `rollouts.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11045); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `root`

| Column | Audited content |
| --- | --- |
| Command | `root`; `fbrcm`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/root.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34133). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--version`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34133); [input schema](../../schemas/cli/1.0.0/root.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/root.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/root.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34133); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/root.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34133); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11141). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34133); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11141). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:root:input`; `urn:fbrcm:schema:cli:1.0.0:command:root:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1171); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `root`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11141); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `schema.list`

| Column | Audited content |
| --- | --- |
| Command | `schema.list`; `fbrcm schema list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/schema_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34286). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34286); [input schema](../../schemas/cli/1.0.0/schema_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/schema_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.0.0/schema_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34286); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/schema_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34286); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11233). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34286); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11233). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:schema.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:schema.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1300); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `schema.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11233); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `schema.show`

| Column | Audited content |
| --- | --- |
| Command | `schema.show`; `fbrcm schema show`. |
| Arguments | required schema_id. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/schema_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34399). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34399); [input schema](../../schemas/cli/1.0.0/schema_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/schema_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34399). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.0.0/schema_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34399); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/schema_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34399); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11315). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34399); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11315). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:schema.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:schema.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1305); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `schema.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11315); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `setup`

| Column | Audited content |
| --- | --- |
| Command | `setup`; `fbrcm setup`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/setup.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34520). |
| Options | `--include-hidden`, `--no-local-config`, `--noopen`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34520); [input schema](../../schemas/cli/1.0.0/setup.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/setup.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/setup/commands.go); [response schema](../../schemas/cli/1.0.0/setup.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34520); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/setup.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `required`; `guided_setup_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34520); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11400). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34520); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11400). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:setup:input`; `urn:fbrcm:schema:cli:1.0.0:command:setup:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2918); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `setup`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11400); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme`

| Column | Audited content |
| --- | --- |
| Command | `theme`; `fbrcm theme`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/theme.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34660). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34660); [input schema](../../schemas/cli/1.0.0/theme.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/theme.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34660); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/theme.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34660); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11489). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34660); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11489). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3492); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11489); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.delete`

| Column | Audited content |
| --- | --- |
| Command | `theme.delete`; `fbrcm theme delete`. |
| Arguments | required theme. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/theme_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34778). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34778); [input schema](../../schemas/cli/1.0.0/theme_delete.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/theme_delete.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34778). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_delete.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34778); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/theme_delete.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34778); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11571). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34778); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11571). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3554); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.delete`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11571); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.import`

| Column | Audited content |
| --- | --- |
| Command | `theme.import`; `fbrcm theme import`. |
| Arguments | optional source. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/theme_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34982). |
| Options | `--include-hidden`, `--name`, `--no-local-config`, `--profile`, `--ref`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34982); [input schema](../../schemas/cli/1.0.0/theme_import.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/theme_import.input.schema.json). |
| Stdin | `toml_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:theme`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/theme_import.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_import.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34982); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/theme_import.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `theme.already_exists`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/theme_import.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11654). |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34982); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11654). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34982); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11654). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.import:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.import:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3566); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.import`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11654); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.list`

| Column | Audited content |
| --- | --- |
| Command | `theme.list`; `fbrcm theme list`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/theme_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35217). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35217); [input schema](../../schemas/cli/1.0.0/theme_list.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/theme_list.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35217); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/theme_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35217); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11796). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35217); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11796). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3501); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11796); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.path`

| Column | Audited content |
| --- | --- |
| Command | `theme.path`; `fbrcm theme path`. |
| Arguments | required theme. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/theme_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35334). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35334); [input schema](../../schemas/cli/1.0.0/theme_path.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/theme_path.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_path.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35334); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/theme_path.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35334); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11878). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35334); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11878). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.path:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3541); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.path`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11878); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.rename`

| Column | Audited content |
| --- | --- |
| Command | `theme.rename`; `fbrcm theme rename`. |
| Arguments | required old_name, required new_name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/theme_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35458). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35458); [input schema](../../schemas/cli/1.0.0/theme_rename.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/theme_rename.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35458). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_rename.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35458); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/theme_rename.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35458); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11960). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35458); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11960). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3546); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.rename`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11960); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.reset`

| Column | Audited content |
| --- | --- |
| Command | `theme.reset`; `fbrcm theme reset`. |
| Arguments | No positional arguments. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/theme_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35623). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35623); [input schema](../../schemas/cli/1.0.0/theme_reset.input.schema.json). |
| Selection | N/A; no matching/lookup selector. Literal argument rules remain in [input schema](../../schemas/cli/1.0.0/theme_reset.input.schema.json). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_reset.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35623); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/theme_reset.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35623); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12049). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35623); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12049). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3527); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.reset`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12049); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `theme.switch`

| Column | Audited content |
| --- | --- |
| Command | `theme.switch`; `fbrcm theme switch`. |
| Arguments | required name. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/theme_switch.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35779). |
| Options | `--include-hidden`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35779); [input schema](../../schemas/cli/1.0.0/theme_switch.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/theme_switch.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35779). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_switch.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35779); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/theme_switch.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | N/A; no interactive JSON branch. No-prompt behavior: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35779); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12127). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35779); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12127). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L3512); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `theme.switch`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12127); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `update`

| Column | Audited content |
| --- | --- |
| Command | `update`; `fbrcm update`. |
| Arguments | optional parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/update.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35950). |
| Options | `--change-note`, `--condition`, `--description`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--name`, `--no-group`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--remove-all-conditional-values`, `--remove-conditional-value`, `--search`, `--stateless`, `--timeout`, `--type`, `--use-in-app-default`, `--value`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35950); [input schema](../../schemas/cli/1.0.0/update.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/update.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35950). |
| Stdin | `json_document`; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; precedence and option restrictions: [input schema](../../schemas/cli/1.0.0/update.input.schema.json). |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/update/commands.go); [response schema](../../schemas/cli/1.0.0/update.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35950); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/update.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.non_atomic`, `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/update.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12207). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35950); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12207). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35950); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12207). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:update:input`; `urn:fbrcm:schema:cli:1.0.0:command:update:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L1493); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `update`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12207); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.blame`

| Column | Audited content |
| --- | --- |
| Command | `versions.blame`; `fbrcm versions blame`. |
| Arguments | required project, required parameter. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/versions_blame.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37001). |
| Options | `--all`, `--at`, `--include-hidden`, `--limit`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37001); [input schema](../../schemas/cli/1.0.0/versions_blame.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/versions_blame.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37001). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_blame.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37001); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/versions_blame.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37001); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12356). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37001); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12356). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.blame:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.blame:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2365); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.blame`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12356); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.diff`

| Column | Audited content |
| --- | --- |
| Command | `versions.diff`; `fbrcm versions diff`. |
| Arguments | required project, required from, optional to. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/versions_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37317). |
| Options | `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--include-hidden`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--side-by-side`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37317); [input schema](../../schemas/cli/1.0.0/versions_diff.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/versions_diff.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37317). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_diff.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37317); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/versions_diff.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37317); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12447). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37317); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12447). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2473); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.diff`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12447); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.export`

| Column | Audited content |
| --- | --- |
| Command | `versions.export`; `fbrcm versions export`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/versions_export.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37793). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37793); [input schema](../../schemas/cli/1.0.0/versions_export.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/versions_export.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37793). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_export.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37793); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/versions_export.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37793); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12557). |
| Effects | Level 2; network `conditional`; destructive `true`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L37793); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12557). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.export:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.export:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2501); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.export`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12557); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.list`

| Column | Audited content |
| --- | --- |
| Command | `versions.list`; `fbrcm versions list`. |
| Arguments | required project. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/versions_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38263). |
| Options | `--all`, `--before`, `--cached`, `--include-hidden`, `--limit`, `--no-local-config`, `--profile`, `--since`, `--stateless`, `--timeout`, `--until`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38263); [input schema](../../schemas/cli/1.0.0/versions_list.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/versions_list.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38263). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_list.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38263); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/versions_list.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38263); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12663). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38263); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12663). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.list:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2428); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.list`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12663); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.restore`

| Column | Audited content |
| --- | --- |
| Command | `versions.restore`; `fbrcm versions restore`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/versions_restore.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38614). |
| Options | `--change-note`, `--dry-run`, `--include-hidden`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38614); [input schema](../../schemas/cli/1.0.0/versions_restore.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/versions_restore.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38614). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_restore.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38614); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/versions_restore.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/versions_restore.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12763). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38614); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12763). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L38614); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12763). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2556); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.restore`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12763); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.rollback`

| Column | Audited content |
| --- | --- |
| Command | `versions.rollback`; `fbrcm versions rollback`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/versions_rollback.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39265). |
| Options | `--dry-run`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39265); [input schema](../../schemas/cli/1.0.0/versions_rollback.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/versions_rollback.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39265). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_rollback.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39265); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/versions_rollback.response.schema.json), shared error schema and exit-status rules. |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachable branches: [response schema](../../schemas/cli/1.0.0/versions_rollback.response.schema.json); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12862). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39265); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12862). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39265); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12862). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2517); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.rollback`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12862); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

### `versions.show`

| Column | Audited content |
| --- | --- |
| Command | `versions.show`; `fbrcm versions show`. |
| Arguments | required project, required version. Arity, grammar, normalization and constraints: [input schema](../../schemas/cli/1.0.0/versions_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39867). |
| Options | `--cached`, `--include-hidden`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Types, aliases, defaults, repetition, presence, effectiveness and dependencies: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39867); [input schema](../../schemas/cli/1.0.0/versions_show.input.schema.json). |
| Selection | Complete matching, normalization, precedence, composition, default selection and typed lookup failures: [input schema](../../schemas/cli/1.0.0/versions_show.input.schema.json), referenced semantic definitions, and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39867). |
| Stdin | N/A; no normalized stdin document. |
| Success | Every registered DTO, populated/empty/no-op/status variant and serialization constraint: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_show.response.schema.json). |
| Failure | Closed reachable code set: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39867); typed category/details/retryability and envelope/process statuses: [response schema](../../schemas/cli/1.0.0/versions_show.response.schema.json), shared error schema and exit-status rules. |
| Warnings | N/A; no reachable warning branch; response requires an empty warnings array. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Triggers, bypass/required options, preview and no-prompt/launch evidence: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39867); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12965). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. Complete effect list and exact predicates, exceptions and retry conditions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L39867); [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12965). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.show:response`; `urn:fbrcm:schema:cli:1.0.0:error`; referenced shared semantic/envelope/capability/stdin resources in Appendix B. |
| Documentation | [CLI reference](../../docs/CLI.md#L2458); [shared machine contract](../../docs/cli-contract.md); explicit documentation inventory entry `versions.show`. |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L12965); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and all applicable permanent test/scenario references therein. |
| Verdict | **PASS**; all applicable criteria and final gates pass. |

## Appendix B — schema file and ID inventory

All IDs are unique and resolve during Draft 2020-12 compilation.

| File | `$id` |
| --- | --- |
| [add.input.schema.json](../../schemas/cli/1.0.0/add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:add:input` |
| [add.response.schema.json](../../schemas/cli/1.0.0/add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:add:response` |
| [apply.input.schema.json](../../schemas/cli/1.0.0/apply.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apply:input` |
| [apply.response.schema.json](../../schemas/cli/1.0.0/apply.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apply:response` |
| [apps_config.input.schema.json](../../schemas/cli/1.0.0/apps_config.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.config:input` |
| [apps_config.response.schema.json](../../schemas/cli/1.0.0/apps_config.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.config:response` |
| [apps_list.input.schema.json](../../schemas/cli/1.0.0/apps_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.list:input` |
| [apps_list.response.schema.json](../../schemas/cli/1.0.0/apps_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.list:response` |
| [apps_show.input.schema.json](../../schemas/cli/1.0.0/apps_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.show:input` |
| [apps_show.response.schema.json](../../schemas/cli/1.0.0/apps_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:apps.show:response` |
| [auth_add_gcloud.input.schema.json](../../schemas/cli/1.0.0/auth_add_gcloud.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:input` |
| [auth_add_gcloud.response.schema.json](../../schemas/cli/1.0.0/auth_add_gcloud.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:response` |
| [auth_add_google.input.schema.json](../../schemas/cli/1.0.0/auth_add_google.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:input` |
| [auth_add_google.response.schema.json](../../schemas/cli/1.0.0/auth_add_google.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:response` |
| [auth_add_oauth.input.schema.json](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:input` |
| [auth_add_oauth.response.schema.json](../../schemas/cli/1.0.0/auth_add_oauth.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:response` |
| [auth_add_service-account.input.schema.json](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:input` |
| [auth_add_service-account.response.schema.json](../../schemas/cli/1.0.0/auth_add_service-account.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:response` |
| [auth_bind.input.schema.json](../../schemas/cli/1.0.0/auth_bind.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:input` |
| [auth_bind.response.schema.json](../../schemas/cli/1.0.0/auth_bind.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:response` |
| [auth_delete.input.schema.json](../../schemas/cli/1.0.0/auth_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:input` |
| [auth_delete.response.schema.json](../../schemas/cli/1.0.0/auth_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:response` |
| [auth_list.input.schema.json](../../schemas/cli/1.0.0/auth_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.list:input` |
| [auth_list.response.schema.json](../../schemas/cli/1.0.0/auth_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.list:response` |
| [auth_login.input.schema.json](../../schemas/cli/1.0.0/auth_login.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.login:input` |
| [auth_login.response.schema.json](../../schemas/cli/1.0.0/auth_login.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.login:response` |
| [auth_path.input.schema.json](../../schemas/cli/1.0.0/auth_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.path:input` |
| [auth_path.response.schema.json](../../schemas/cli/1.0.0/auth_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.path:response` |
| [auth_quota-project_set.input.schema.json](../../schemas/cli/1.0.0/auth_quota-project_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:input` |
| [auth_quota-project_set.response.schema.json](../../schemas/cli/1.0.0/auth_quota-project_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:response` |
| [auth_quota-project_show.input.schema.json](../../schemas/cli/1.0.0/auth_quota-project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:input` |
| [auth_quota-project_show.response.schema.json](../../schemas/cli/1.0.0/auth_quota-project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:response` |
| [auth_quota-project_unset.input.schema.json](../../schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:input` |
| [auth_quota-project_unset.response.schema.json](../../schemas/cli/1.0.0/auth_quota-project_unset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:response` |
| [cache_clear.input.schema.json](../../schemas/cli/1.0.0/cache_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:input` |
| [cache_clear.response.schema.json](../../schemas/cli/1.0.0/cache_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:response` |
| [cache_list.input.schema.json](../../schemas/cli/1.0.0/cache_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.list:input` |
| [cache_list.response.schema.json](../../schemas/cli/1.0.0/cache_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.list:response` |
| [cache_path.input.schema.json](../../schemas/cli/1.0.0/cache_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.path:input` |
| [cache_path.response.schema.json](../../schemas/cli/1.0.0/cache_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:cache.path:response` |
| [capabilities.input.schema.json](../../schemas/cli/1.0.0/capabilities.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:capabilities:input` |
| [capabilities.response.schema.json](../../schemas/cli/1.0.0/capabilities.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:capabilities:response` |
| [capability.schema.json](../../schemas/cli/1.0.0/capability.schema.json) | `urn:fbrcm:schema:cli:1.0.0:capability` |
| [completion_bash.input.schema.json](../../schemas/cli/1.0.0/completion_bash.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:input` |
| [completion_bash.response.schema.json](../../schemas/cli/1.0.0/completion_bash.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:response` |
| [completion_fish.input.schema.json](../../schemas/cli/1.0.0/completion_fish.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:input` |
| [completion_fish.response.schema.json](../../schemas/cli/1.0.0/completion_fish.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:response` |
| [completion_powershell.input.schema.json](../../schemas/cli/1.0.0/completion_powershell.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:input` |
| [completion_powershell.response.schema.json](../../schemas/cli/1.0.0/completion_powershell.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:response` |
| [completion_zsh.input.schema.json](../../schemas/cli/1.0.0/completion_zsh.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:input` |
| [completion_zsh.response.schema.json](../../schemas/cli/1.0.0/completion_zsh.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:response` |
| [conditions_add.input.schema.json](../../schemas/cli/1.0.0/conditions_add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:input` |
| [conditions_add.response.schema.json](../../schemas/cli/1.0.0/conditions_add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:response` |
| [conditions_delete.input.schema.json](../../schemas/cli/1.0.0/conditions_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:input` |
| [conditions_delete.response.schema.json](../../schemas/cli/1.0.0/conditions_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:response` |
| [conditions_edit.input.schema.json](../../schemas/cli/1.0.0/conditions_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:input` |
| [conditions_edit.response.schema.json](../../schemas/cli/1.0.0/conditions_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:response` |
| [conditions_list.input.schema.json](../../schemas/cli/1.0.0/conditions_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:input` |
| [conditions_list.response.schema.json](../../schemas/cli/1.0.0/conditions_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:response` |
| [conditions_move.input.schema.json](../../schemas/cli/1.0.0/conditions_move.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:input` |
| [conditions_move.response.schema.json](../../schemas/cli/1.0.0/conditions_move.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:response` |
| [conditions_rename.input.schema.json](../../schemas/cli/1.0.0/conditions_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:input` |
| [conditions_rename.response.schema.json](../../schemas/cli/1.0.0/conditions_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:response` |
| [conditions_show.input.schema.json](../../schemas/cli/1.0.0/conditions_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:input` |
| [conditions_show.response.schema.json](../../schemas/cli/1.0.0/conditions_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:response` |
| [conditions_validate.input.schema.json](../../schemas/cli/1.0.0/conditions_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:input` |
| [conditions_validate.response.schema.json](../../schemas/cli/1.0.0/conditions_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:response` |
| [config_edit.input.schema.json](../../schemas/cli/1.0.0/config_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.edit:input` |
| [config_edit.response.schema.json](../../schemas/cli/1.0.0/config_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.edit:response` |
| [config_path.input.schema.json](../../schemas/cli/1.0.0/config_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.path:input` |
| [config_path.response.schema.json](../../schemas/cli/1.0.0/config_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.path:response` |
| [config_reset.input.schema.json](../../schemas/cli/1.0.0/config_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.reset:input` |
| [config_reset.response.schema.json](../../schemas/cli/1.0.0/config_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.reset:response` |
| [config_set.input.schema.json](../../schemas/cli/1.0.0/config_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.set:input` |
| [config_set.response.schema.json](../../schemas/cli/1.0.0/config_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.set:response` |
| [config_show.input.schema.json](../../schemas/cli/1.0.0/config_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.show:input` |
| [config_show.response.schema.json](../../schemas/cli/1.0.0/config_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.show:response` |
| [config_validate.input.schema.json](../../schemas/cli/1.0.0/config_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.validate:input` |
| [config_validate.response.schema.json](../../schemas/cli/1.0.0/config_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:config.validate:response` |
| [delete.input.schema.json](../../schemas/cli/1.0.0/delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:delete:input` |
| [delete.response.schema.json](../../schemas/cli/1.0.0/delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:delete:response` |
| [doctor.input.schema.json](../../schemas/cli/1.0.0/doctor.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:doctor:input` |
| [doctor.response.schema.json](../../schemas/cli/1.0.0/doctor.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:doctor:response` |
| [draft_change-note.input.schema.json](../../schemas/cli/1.0.0/draft_change-note.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:input` |
| [draft_change-note.response.schema.json](../../schemas/cli/1.0.0/draft_change-note.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:response` |
| [draft_diff.input.schema.json](../../schemas/cli/1.0.0/draft_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:input` |
| [draft_diff.response.schema.json](../../schemas/cli/1.0.0/draft_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:response` |
| [draft_discard.input.schema.json](../../schemas/cli/1.0.0/draft_discard.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:input` |
| [draft_discard.response.schema.json](../../schemas/cli/1.0.0/draft_discard.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:response` |
| [draft_list.input.schema.json](../../schemas/cli/1.0.0/draft_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.list:input` |
| [draft_list.response.schema.json](../../schemas/cli/1.0.0/draft_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.list:response` |
| [draft_path.input.schema.json](../../schemas/cli/1.0.0/draft_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.path:input` |
| [draft_path.response.schema.json](../../schemas/cli/1.0.0/draft_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.path:response` |
| [draft_publish.input.schema.json](../../schemas/cli/1.0.0/draft_publish.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:input` |
| [draft_publish.response.schema.json](../../schemas/cli/1.0.0/draft_publish.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:response` |
| [draft_show.input.schema.json](../../schemas/cli/1.0.0/draft_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.show:input` |
| [draft_show.response.schema.json](../../schemas/cli/1.0.0/draft_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:draft.show:response` |
| [duplicate.input.schema.json](../../schemas/cli/1.0.0/duplicate.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:duplicate:input` |
| [duplicate.response.schema.json](../../schemas/cli/1.0.0/duplicate.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:duplicate:response` |
| [envelope.schema.json](../../schemas/cli/1.0.0/envelope.schema.json) | `urn:fbrcm:schema:cli:1.0.0:envelope` |
| [error.schema.json](../../schemas/cli/1.0.0/error.schema.json) | `urn:fbrcm:schema:cli:1.0.0:error` |
| [experiments_delete.input.schema.json](../../schemas/cli/1.0.0/experiments_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:input` |
| [experiments_delete.response.schema.json](../../schemas/cli/1.0.0/experiments_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:response` |
| [experiments_list.input.schema.json](../../schemas/cli/1.0.0/experiments_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:input` |
| [experiments_list.response.schema.json](../../schemas/cli/1.0.0/experiments_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:response` |
| [experiments_show.input.schema.json](../../schemas/cli/1.0.0/experiments_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:input` |
| [experiments_show.response.schema.json](../../schemas/cli/1.0.0/experiments_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:response` |
| [get.input.schema.json](../../schemas/cli/1.0.0/get.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:get:input` |
| [get.response.schema.json](../../schemas/cli/1.0.0/get.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:get:response` |
| [groups_add.input.schema.json](../../schemas/cli/1.0.0/groups_add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.add:input` |
| [groups_add.response.schema.json](../../schemas/cli/1.0.0/groups_add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.add:response` |
| [groups_delete.input.schema.json](../../schemas/cli/1.0.0/groups_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:input` |
| [groups_delete.response.schema.json](../../schemas/cli/1.0.0/groups_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:response` |
| [groups_edit.input.schema.json](../../schemas/cli/1.0.0/groups_edit.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:input` |
| [groups_edit.response.schema.json](../../schemas/cli/1.0.0/groups_edit.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:response` |
| [groups_list.input.schema.json](../../schemas/cli/1.0.0/groups_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.list:input` |
| [groups_list.response.schema.json](../../schemas/cli/1.0.0/groups_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.list:response` |
| [groups_rename.input.schema.json](../../schemas/cli/1.0.0/groups_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:input` |
| [groups_rename.response.schema.json](../../schemas/cli/1.0.0/groups_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:response` |
| [help.input.schema.json](../../schemas/cli/1.0.0/help.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:help:input` |
| [help.response.schema.json](../../schemas/cli/1.0.0/help.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:help:response` |
| [hooks_fingerprint.input.schema.json](../../schemas/cli/1.0.0/hooks_fingerprint.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:input` |
| [hooks_fingerprint.response.schema.json](../../schemas/cli/1.0.0/hooks_fingerprint.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:response` |
| [hooks_status.input.schema.json](../../schemas/cli/1.0.0/hooks_status.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:input` |
| [hooks_status.response.schema.json](../../schemas/cli/1.0.0/hooks_status.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:response` |
| [hooks_trust.input.schema.json](../../schemas/cli/1.0.0/hooks_trust.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:input` |
| [hooks_trust.response.schema.json](../../schemas/cli/1.0.0/hooks_trust.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:response` |
| [hooks_untrust.input.schema.json](../../schemas/cli/1.0.0/hooks_untrust.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:input` |
| [hooks_untrust.response.schema.json](../../schemas/cli/1.0.0/hooks_untrust.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:response` |
| [mcp.input.schema.json](../../schemas/cli/1.0.0/mcp.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:mcp:input` |
| [mcp.response.schema.json](../../schemas/cli/1.0.0/mcp.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:mcp:response` |
| [personalizations_list.input.schema.json](../../schemas/cli/1.0.0/personalizations_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:input` |
| [personalizations_list.response.schema.json](../../schemas/cli/1.0.0/personalizations_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:response` |
| [personalizations_show.input.schema.json](../../schemas/cli/1.0.0/personalizations_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:input` |
| [personalizations_show.response.schema.json](../../schemas/cli/1.0.0/personalizations_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:response` |
| [plan_show.input.schema.json](../../schemas/cli/1.0.0/plan_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:plan.show:input` |
| [plan_show.response.schema.json](../../schemas/cli/1.0.0/plan_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:plan.show:response` |
| [plan_validate.input.schema.json](../../schemas/cli/1.0.0/plan_validate.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:plan.validate:input` |
| [plan_validate.response.schema.json](../../schemas/cli/1.0.0/plan_validate.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:plan.validate:response` |
| [profile.input.schema.json](../../schemas/cli/1.0.0/profile.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile:input` |
| [profile.response.schema.json](../../schemas/cli/1.0.0/profile.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile:response` |
| [profile_delete.input.schema.json](../../schemas/cli/1.0.0/profile_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:input` |
| [profile_delete.response.schema.json](../../schemas/cli/1.0.0/profile_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:response` |
| [profile_list.input.schema.json](../../schemas/cli/1.0.0/profile_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.list:input` |
| [profile_list.response.schema.json](../../schemas/cli/1.0.0/profile_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.list:response` |
| [profile_path.input.schema.json](../../schemas/cli/1.0.0/profile_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.path:input` |
| [profile_path.response.schema.json](../../schemas/cli/1.0.0/profile_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.path:response` |
| [profile_rename.input.schema.json](../../schemas/cli/1.0.0/profile_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:input` |
| [profile_rename.response.schema.json](../../schemas/cli/1.0.0/profile_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:response` |
| [profile_switch.input.schema.json](../../schemas/cli/1.0.0/profile_switch.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:input` |
| [profile_switch.response.schema.json](../../schemas/cli/1.0.0/profile_switch.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:response` |
| [project_defaults.input.schema.json](../../schemas/cli/1.0.0/project_defaults.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:input` |
| [project_defaults.response.schema.json](../../schemas/cli/1.0.0/project_defaults.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:response` |
| [project_export.input.schema.json](../../schemas/cli/1.0.0/project_export.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.export:input` |
| [project_export.response.schema.json](../../schemas/cli/1.0.0/project_export.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.export:response` |
| [project_import.input.schema.json](../../schemas/cli/1.0.0/project_import.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.import:input` |
| [project_import.response.schema.json](../../schemas/cli/1.0.0/project_import.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.import:response` |
| [project_open.input.schema.json](../../schemas/cli/1.0.0/project_open.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.open:input` |
| [project_open.response.schema.json](../../schemas/cli/1.0.0/project_open.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.open:response` |
| [project_quota-project_set.input.schema.json](../../schemas/cli/1.0.0/project_quota-project_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:input` |
| [project_quota-project_set.response.schema.json](../../schemas/cli/1.0.0/project_quota-project_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:response` |
| [project_quota-project_show.input.schema.json](../../schemas/cli/1.0.0/project_quota-project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:input` |
| [project_quota-project_show.response.schema.json](../../schemas/cli/1.0.0/project_quota-project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:response` |
| [project_quota-project_unset.input.schema.json](../../schemas/cli/1.0.0/project_quota-project_unset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:input` |
| [project_quota-project_unset.response.schema.json](../../schemas/cli/1.0.0/project_quota-project_unset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:response` |
| [project_show.input.schema.json](../../schemas/cli/1.0.0/project_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.show:input` |
| [project_show.response.schema.json](../../schemas/cli/1.0.0/project_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.show:response` |
| [project_templates_set.input.schema.json](../../schemas/cli/1.0.0/project_templates_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:input` |
| [project_templates_set.response.schema.json](../../schemas/cli/1.0.0/project_templates_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:response` |
| [project_templates_show.input.schema.json](../../schemas/cli/1.0.0/project_templates_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:input` |
| [project_templates_show.response.schema.json](../../schemas/cli/1.0.0/project_templates_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:response` |
| [projects_aliases_import.input.schema.json](../../schemas/cli/1.0.0/projects_aliases_import.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:input` |
| [projects_aliases_import.response.schema.json](../../schemas/cli/1.0.0/projects_aliases_import.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:response` |
| [projects_aliases_list.input.schema.json](../../schemas/cli/1.0.0/projects_aliases_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:input` |
| [projects_aliases_list.response.schema.json](../../schemas/cli/1.0.0/projects_aliases_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:response` |
| [projects_aliases_remove.input.schema.json](../../schemas/cli/1.0.0/projects_aliases_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:input` |
| [projects_aliases_remove.response.schema.json](../../schemas/cli/1.0.0/projects_aliases_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:response` |
| [projects_aliases_set.input.schema.json](../../schemas/cli/1.0.0/projects_aliases_set.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:input` |
| [projects_aliases_set.response.schema.json](../../schemas/cli/1.0.0/projects_aliases_set.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:response` |
| [projects_blacklist_add.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.add:input` |
| [projects_blacklist_add.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.add:response` |
| [projects_blacklist_clear.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.clear:input` |
| [projects_blacklist_clear.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.clear:response` |
| [projects_blacklist_disable.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_disable.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.disable:input` |
| [projects_blacklist_disable.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_disable.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.disable:response` |
| [projects_blacklist_enable.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_enable.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.enable:input` |
| [projects_blacklist_enable.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_enable.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.enable:response` |
| [projects_blacklist_list.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.list:input` |
| [projects_blacklist_list.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.list:response` |
| [projects_blacklist_remove.input.schema.json](../../schemas/cli/1.0.0/projects_blacklist_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.remove:input` |
| [projects_blacklist_remove.response.schema.json](../../schemas/cli/1.0.0/projects_blacklist_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.blacklist.remove:response` |
| [projects_diff.input.schema.json](../../schemas/cli/1.0.0/projects_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:input` |
| [projects_diff.response.schema.json](../../schemas/cli/1.0.0/projects_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:response` |
| [projects_forget.input.schema.json](../../schemas/cli/1.0.0/projects_forget.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:input` |
| [projects_forget.response.schema.json](../../schemas/cli/1.0.0/projects_forget.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:response` |
| [projects_list.input.schema.json](../../schemas/cli/1.0.0/projects_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.list:input` |
| [projects_list.response.schema.json](../../schemas/cli/1.0.0/projects_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.list:response` |
| [projects_path.input.schema.json](../../schemas/cli/1.0.0/projects_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.path:input` |
| [projects_path.response.schema.json](../../schemas/cli/1.0.0/projects_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.path:response` |
| [projects_promote.input.schema.json](../../schemas/cli/1.0.0/projects_promote.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:input` |
| [projects_promote.response.schema.json](../../schemas/cli/1.0.0/projects_promote.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:response` |
| [projects_reset.input.schema.json](../../schemas/cli/1.0.0/projects_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:input` |
| [projects_reset.response.schema.json](../../schemas/cli/1.0.0/projects_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:response` |
| [projects_update.input.schema.json](../../schemas/cli/1.0.0/projects_update.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.update:input` |
| [projects_update.response.schema.json](../../schemas/cli/1.0.0/projects_update.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.update:response` |
| [projects_whitelist_add.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_add.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.add:input` |
| [projects_whitelist_add.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_add.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.add:response` |
| [projects_whitelist_clear.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_clear.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.clear:input` |
| [projects_whitelist_clear.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_clear.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.clear:response` |
| [projects_whitelist_disable.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_disable.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.disable:input` |
| [projects_whitelist_disable.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_disable.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.disable:response` |
| [projects_whitelist_enable.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_enable.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.enable:input` |
| [projects_whitelist_enable.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_enable.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.enable:response` |
| [projects_whitelist_list.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.list:input` |
| [projects_whitelist_list.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.list:response` |
| [projects_whitelist_remove.input.schema.json](../../schemas/cli/1.0.0/projects_whitelist_remove.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.remove:input` |
| [projects_whitelist_remove.response.schema.json](../../schemas/cli/1.0.0/projects_whitelist_remove.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:projects.whitelist.remove:response` |
| [publication_plan.schema.json](../../schemas/cli/1.0.0/publication_plan.schema.json) | `urn:fbrcm:schema:cli:1.0.0:publication_plan` |
| [rollouts_delete.input.schema.json](../../schemas/cli/1.0.0/rollouts_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:input` |
| [rollouts_delete.response.schema.json](../../schemas/cli/1.0.0/rollouts_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:response` |
| [rollouts_list.input.schema.json](../../schemas/cli/1.0.0/rollouts_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:input` |
| [rollouts_list.response.schema.json](../../schemas/cli/1.0.0/rollouts_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:response` |
| [rollouts_show.input.schema.json](../../schemas/cli/1.0.0/rollouts_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:input` |
| [rollouts_show.response.schema.json](../../schemas/cli/1.0.0/rollouts_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:response` |
| [root.input.schema.json](../../schemas/cli/1.0.0/root.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:root:input` |
| [root.response.schema.json](../../schemas/cli/1.0.0/root.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:root:response` |
| [schema_list.input.schema.json](../../schemas/cli/1.0.0/schema_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:schema.list:input` |
| [schema_list.response.schema.json](../../schemas/cli/1.0.0/schema_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:schema.list:response` |
| [schema_show.input.schema.json](../../schemas/cli/1.0.0/schema_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:schema.show:input` |
| [schema_show.response.schema.json](../../schemas/cli/1.0.0/schema_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:schema.show:response` |
| [semantic.schema.json](../../schemas/cli/1.0.0/semantic.schema.json) | `urn:fbrcm:schema:cli:1.0.0:semantic` |
| [setup.input.schema.json](../../schemas/cli/1.0.0/setup.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:setup:input` |
| [setup.response.schema.json](../../schemas/cli/1.0.0/setup.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:setup:response` |
| [stdin.credentials.schema.json](../../schemas/cli/1.0.0/stdin.credentials.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:credentials` |
| [stdin.oauth_credentials.schema.json](../../schemas/cli/1.0.0/stdin.oauth_credentials.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:oauth_credentials` |
| [stdin.publication_plan.schema.json](../../schemas/cli/1.0.0/stdin.publication_plan.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:publication_plan` |
| [stdin.remote_config.schema.json](../../schemas/cli/1.0.0/stdin.remote_config.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config` |
| [stdin.remote_config_import.schema.json](../../schemas/cli/1.0.0/stdin.remote_config_import.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config_import` |
| [stdin.service_account_credentials.schema.json](../../schemas/cli/1.0.0/stdin.service_account_credentials.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:service_account_credentials` |
| [stdin.theme.schema.json](../../schemas/cli/1.0.0/stdin.theme.schema.json) | `urn:fbrcm:schema:cli:1.0.0:stdin:theme` |
| [theme.input.schema.json](../../schemas/cli/1.0.0/theme.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme:input` |
| [theme.response.schema.json](../../schemas/cli/1.0.0/theme.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme:response` |
| [theme_delete.input.schema.json](../../schemas/cli/1.0.0/theme_delete.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:input` |
| [theme_delete.response.schema.json](../../schemas/cli/1.0.0/theme_delete.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:response` |
| [theme_import.input.schema.json](../../schemas/cli/1.0.0/theme_import.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.import:input` |
| [theme_import.response.schema.json](../../schemas/cli/1.0.0/theme_import.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.import:response` |
| [theme_list.input.schema.json](../../schemas/cli/1.0.0/theme_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.list:input` |
| [theme_list.response.schema.json](../../schemas/cli/1.0.0/theme_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.list:response` |
| [theme_path.input.schema.json](../../schemas/cli/1.0.0/theme_path.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.path:input` |
| [theme_path.response.schema.json](../../schemas/cli/1.0.0/theme_path.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.path:response` |
| [theme_rename.input.schema.json](../../schemas/cli/1.0.0/theme_rename.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:input` |
| [theme_rename.response.schema.json](../../schemas/cli/1.0.0/theme_rename.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:response` |
| [theme_reset.input.schema.json](../../schemas/cli/1.0.0/theme_reset.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:input` |
| [theme_reset.response.schema.json](../../schemas/cli/1.0.0/theme_reset.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:response` |
| [theme_switch.input.schema.json](../../schemas/cli/1.0.0/theme_switch.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:input` |
| [theme_switch.response.schema.json](../../schemas/cli/1.0.0/theme_switch.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:response` |
| [update.input.schema.json](../../schemas/cli/1.0.0/update.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:update:input` |
| [update.response.schema.json](../../schemas/cli/1.0.0/update.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:update:response` |
| [versions_blame.input.schema.json](../../schemas/cli/1.0.0/versions_blame.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.blame:input` |
| [versions_blame.response.schema.json](../../schemas/cli/1.0.0/versions_blame.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.blame:response` |
| [versions_diff.input.schema.json](../../schemas/cli/1.0.0/versions_diff.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:input` |
| [versions_diff.response.schema.json](../../schemas/cli/1.0.0/versions_diff.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:response` |
| [versions_export.input.schema.json](../../schemas/cli/1.0.0/versions_export.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.export:input` |
| [versions_export.response.schema.json](../../schemas/cli/1.0.0/versions_export.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.export:response` |
| [versions_list.input.schema.json](../../schemas/cli/1.0.0/versions_list.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.list:input` |
| [versions_list.response.schema.json](../../schemas/cli/1.0.0/versions_list.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.list:response` |
| [versions_restore.input.schema.json](../../schemas/cli/1.0.0/versions_restore.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:input` |
| [versions_restore.response.schema.json](../../schemas/cli/1.0.0/versions_restore.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:response` |
| [versions_rollback.input.schema.json](../../schemas/cli/1.0.0/versions_rollback.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:input` |
| [versions_rollback.response.schema.json](../../schemas/cli/1.0.0/versions_rollback.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:response` |
| [versions_show.input.schema.json](../../schemas/cli/1.0.0/versions_show.input.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.show:input` |
| [versions_show.response.schema.json](../../schemas/cli/1.0.0/versions_show.response.schema.json) | `urn:fbrcm:schema:cli:1.0.0:command:versions.show:response` |
