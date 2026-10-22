# CLI JSON contract audit — 2026-10-22

Repository revision: `f4c2f9b0b92886a75b7fc55f2d74c15f0d365771` plus the audited repair worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `118`
Input schemas: `118`
Response schemas: `118`
Shared schemas: `12`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0` unresolved (`7` found and closed)

The revision and standard above were frozen before inspection. No audit criteria,
release marker, or contract version were changed. Inspection used the complete
Cobra tree, DTO registrations, behavior manifest, generated schemas, both
capability goldens, documentation, and the 15-class evidence matrix. The earlier
passing command records remain useful background, but Appendix A provides a
current record for every command, with exact references to its authoritative
arguments, options, selection, results, effects, documentation, and evidence.
References incorporate their complete content; they are not omitted cells.

## Inventory and generation

`E = C = I = R = D = B = M`: each set has 118 members and every symmetric
difference is empty. Runtime discovery and DTO/behavior registration are checked
by `TestEveryExecutableCommandHasCapabilityAndPublishedSchemas`; documentation
and test-class membership have separate exhaustive tests. The compact golden is
unchanged. The detailed golden and embedded capabilities change only the repaired
command problem-code sets. Appendix B inventories all 248 schema files and IDs.
All IDs are unique; all schemas compile with Draft 2020-12 and all references
resolve. Every capability input/response ID identifies its command's schema.

Final contract-lock SHA-256: `29f1e9b274b07b8d2b92639d5da0344e4a207c5698101f5f2b2e85bf374fa219`.
Two consecutive `go run ./cmd/schemagen` invocations produced byte-identical
schemas, capability artifacts, audit evidence, and lock (253 generated assets).
The sorted path-to-SHA-256 manifest digest was
`d01d50df2efce2d6623bcb34448820e82a5e486f64a2a072523b91481d04939d`.

## Criterion and test-class results

| Criteria | Result and evidence |
| --- | --- |
| INV-01–INV-04 | PASS: complete set equality, unique paths and schema IDs, version/release agreement, root/help/completion/parsing/startup envelopes. |
| ARG-01–ARG-07 | PASS: exhaustive Cobra arity, input constraints, enums/bounds/normalization, presence/repetition, ignored and conditional options, runtime validation annotations. |
| SEL-01–SEL-05 | PASS: command-local selector rules, prefixes, exactness, precedence, repeated/default scope, not-found and ambiguity cases. No-match forget now retains its typed problem. |
| STDIN-01–STDIN-04 | PASS: concrete schemas and mode restrictions; malformed/empty/null/array/object plan inputs fail with their typed problem. |
| OUT-01–OUT-06 | PASS: registered DTO/no-data variants, clean envelopes, empty/no-op states, status/count correlations, artifacts, process/embedded statuses, and schema validation. |
| ERR-01–ERR-06 | PASS: typed source classification, closed reachable code sets, detail shapes, category/retryability/status constraints, warnings, batch aggregation, redaction and bounds. |
| BEH-01–BEH-06 | PASS: shared and command-specific effect, network, destruction, idempotency, suppression and predicate suites. |
| INT-01–INT-04 | PASS: non-interactive confirmation, selection, input, OAuth and guided-setup boundaries and preview schemas. |
| DOC-01–DOC-04 | PASS: command inventory, CLI reference, shared contract, capability goldens, examples and runtime agree after repairs. |
| GEN-01–GEN-05 | PASS: deterministic generation, reference compilation, exhaustive regression inventory, complete evidence matrix, root and E2E tests/vet/lint. |

The generated evidence golden contains 118 records × 15 classes and 56 checked
test symbols. Discovery, invocation, boundary, effectiveness, selection, stdin,
success, no-op, failure, interaction, warning, batch, effects, artifact and
determinism cells either cite executable evidence or state a justified N/A.
All catalog symbols and E2E scenario references are checked by the matrix test.

Additional real-process testing used fresh config/cache/home/work directories,
offline mode, no inherited Firebase credentials, explicit stdin, and time limits:

- 590 probes: all 118 commands with minimal argv, missing profile, invalid
  timeout, help, and stateless variants; zero unexpected internal failures.
- 96 probes: applicable confirmation bypasses, missing input files, and
  empty/null/array/object/malformed stdin; zero unexpected internal failures.
- Every probe checked a single JSON document, trailing newline, matching
  process/envelope status, and absence of contract-enforcement or unclassified
  fallbacks. The 96 additional documents were also independently validated
  against the compiled published response schemas.
- Permanent regressions cover 12 unconfigured/input failures, 15 actual plan
  parser failures, 10 typed plan/quota/app envelope cases, and four actual
  syscall error classifications. Typed invariant failures are checked before
  command execution, including exact preservation of their source message.

## Closed findings

### 1. Filesystem syscall failures classified as network failures

Criterion: ERR-01, ERR-02, OUT-05, GEN-04.
Commands: every command that reads or writes local paths; reproduced in plan
show/validate, project quota set, project templates show, and projects forget.
Evidence: `ops/contract/contract.go`, `Classify`; actual missing-file probes;
`ops/contract/contract_test.go`, `TestClassifyStableProblemCategories`;
`cli/app/audit_runtime_test.go`, unconfigured-failure cases.
Observed: `errors.As` found a nested syscall implementing `net.Error` before
recognizing its enclosing `os.PathError`. This produced network errors or a
schema-enforcement fallback, with incorrect retryability and status.
Required: preserve the typed filesystem operation and its documented semantics.
Remedy class: runtime, test, documentation.
Retest: ENOENT, wrapped ENOENT, EACCES and ENOSPC retain file details and exit 13
or permission exit 5; missing-file process cases pass. **Closed.**

### 2. Missing auth setup absent from command problem catalogs

Criterion: OUT-02, OUT-05, ERR-02, DOC-01, GEN-04.
Commands: auth.path, auth.delete, auth.quota-project.set.
Evidence: `ops/contract/problem_codes.go`; corresponding response schemas;
`TestJSONUnconfiguredAndInvalidInputFailuresConform`.
Observed: missing/empty auth registries emitted `auth.setup_required`, absent
from these closed catalogs, and were replaced by internal contract failures.
Required: publish every reachable typed failure with its actual status.
Remedy class: capability, schema, test, generated, documentation.
Retest: schema-valid `auth.setup_required`, category auth, exit 4; related
quota-unset behavior is also covered. **Closed.**

### 3. Get stdin advertised the wrong validation code

Criterion: STDIN-02, OUT-05, ERR-02, ERR-03, DOC-01, GEN-04.
Commands: get, including stateless stdin execution.
Evidence: `ops/workflows/get/stdin.go`; `get.response.schema.json`;
`ops/contract/problem_codes.go`; new runtime regressions.
Observed: runtime emitted `remote_config.invalid`, while discovery and the
response schema advertised `stdin.remote_config.invalid` for this command.
Required: schema, code catalog and actual local decoder failure must agree.
Remedy class: capability, schema, test, generated, documentation.
Retest: malformed and boundary stdin retain validation status 8; the unreachable
advertised code is replaced; schema-show E2E snapshot updated. **Closed.**

### 4. Forget no-match failure absent from its catalog

Criterion: SEL-04, OUT-05, ERR-02, DOC-01, GEN-04.
Commands: projects.forget.
Evidence: `ops/workflows/projects/commands.go`; its response schema;
new empty-registry regression.
Observed: an existing registry with zero selected projects emitted typed
`project.not_found`, absent from its response contract.
Required: preserve the documented selector failure instead of enforcement exit 15.
Remedy class: capability, schema, test, generated, documentation.
Retest: an existing empty registry produces schema-valid not_found exit 6;
a missing registry independently retains file.io_failed exit 13. **Closed.**

### 5. Shared problem schema omitted reachable details and constraints

Criterion: OUT-05, ERR-02, DOC-01, GEN-02, GEN-04.
Commands: apply, plan.show, plan.validate, quota-dependent operations and app
lookup operations; shared error/envelope resources affect every response.
Evidence: `cmd/schemagen/main.go`, `problemDetailsSchema` and
`problemCodeConstraints`; `TestJSONInvalidPublicationPlanInputsConform` and
`TestTypedPlanQuotaAndAppProblemsConform`.
Observed: publication_plan, quota_project and internal details were absent;
plan/quota/app codes lacked their fixed category/retryability constraints.
All 15 malformed-plan process probes became internal contract failures.
Required: every actual detail object must validate, and known fixed problem
semantics must reject contradictory fields.
Remedy class: schema, test, generated, documentation.
Retest: all 15 parser failures retain plan.invalid exit 8; typed integrity,
unsupported version, stale/existing plan, quota, invariant and app errors
validate; contradictory category/retryability values are rejected. **Closed.**

### 6. Typed internal startup failures recast as argument failures

Criterion: ERR-01, ERR-02, OUT-05, GEN-04.
Commands: any command reaching a typed internal invariant before RunE.
Evidence: `ops/contract/contract.go`, `classifyCommandError` and `ExitCode`;
quota-header invariant regression in `TestTypedPlanQuotaAndAppProblemsConform`.
Observed: all internal-category startup errors were recast as argument.invalid;
problem classification and exit calculation duplicated this logic.
Required: typed failures retain their source code/category and semantic status.
Remedy class: runtime, test.
Retest: a typed quota-header invariant retains internal.contract_violation,
its original message/internal details, and exit 15. Exit calculation reuses the
same command classifier; only unclassified Cobra failures use the argument
fallback. **Closed.**

### 7. E2E module lint failed on intentional process reaping

Criterion: GEN-05.
Commands: repository gate; all E2E-backed command records.
Evidence: `e2e/internal/harness/hoverfly.go`; initial errcheck diagnostics and
final module-wide lint run.
Observed: error-channel receives while reaping intentionally terminated
Hoverfly processes triggered errcheck. Blank assignments also triggered
staticcheck S1005, so they could not satisfy both enabled rules.
Required: the complete repository gate, including the separate E2E module, passes.
Remedy class: test harness.
Retest: narrow, explained errcheck annotations on intentional termination
receives preserve existing behavior; unexpected process exits retain their
existing handling. Full replay and both linters pass. **Closed.**

## Final repository gates

| Gate | Final result |
| --- | --- |
| go run ./cmd/schemagen, twice | PASS; byte-identical across all 253 assets. |
| Schema/DTO/capability/docs/behavior/evidence inventory | PASS; all seven 118-member sets equal. |
| Draft 2020-12 compilation | PASS; 248 unique schemas, no dangling references. |
| go test -count=1 ./... (root) | PASS; 100 tested packages. |
| go test -count=1 ./... (e2e) | PASS; all 176 scenarios, including 164 JSON scenarios, plus harness/readguard tests. |
| go vet ./... (root and e2e) | PASS. |
| golangci-lint run (root and e2e) | PASS; zero issues. |
| Additional isolated process probes | PASS; 686 invocations. |
| git diff --check and final scope review | PASS; intended repairs, regression/evidence, docs, generated artifacts, snapshot, audit report only. |

Sandboxed final test attempts could not bind loopback fake servers. The final
passing test runs used authorized loopback access; E2E remained in fixture replay
mode. Writable temporary Go/lint caches avoided sandbox restrictions on the
user's ordinary cache directories. No live Firebase publication was performed.

All twelve section-7 acceptance conditions pass and all findings are closed.
Under the frozen standard 1.0.0, the verdict is **AUTHORITATIVE: PASS**.

## Appendix A: complete command records

Each record has all fourteen section-4 columns. Source links incorporate the
exact current record/schema content, including all structured normalization,
matching, validation, invariants, conditions and field correlations. DTO source
links identify the static registrations checked against generated success data.
Failure semantics use the shared [problem schema](../../schemas/cli/1.0.0/error.schema.json)
and [exit-status rules](../cli-contract.md#exit-statuses). Evidence links identify
the command's complete 15-class matrix, including reasoned N/A decisions.

### `add`

| Column | Audited content |
| --- | --- |
| Command | `add`; `fbrcm add`. |
| Arguments | required parameter. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--expr`, `--group`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--type`, `--use-in-app-default`, `--value`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/add.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/add.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/add.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/add/commands.go); [response schema](../../schemas/cli/1.0.0/add.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/add.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/add.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L64). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L64). |
| Effects | Level 3; network `conditional`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L64). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:add:input`; `urn:fbrcm:schema:cli:1.0.0:command:add:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm add <parameter>](../../docs/CLI.md#L1350); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L64); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `apply`

| Column | Audited content |
| --- | --- |
| Command | `apply`; `fbrcm apply`. |
| Arguments | required plan. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/apply.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L939). |
| Options | `--dry-run`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L939); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/apply.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/apply.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L939). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:publication_plan`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/apply.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/apply/commands.go); [response schema](../../schemas/cli/1.0.0/apply.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L939); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/apply.response.schema.json). |
| Warnings | `plan.source_draft_changed`, `publication.cache_stale`, `publication.draft_cleanup_failed`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/apply.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L213). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L939), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L213). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L939), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L213). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:apply:input`; `urn:fbrcm:schema:cli:1.0.0:command:apply:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm apply <plan>](../../docs/CLI.md#L1327); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L213); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `apps.config`

| Column | Audited content |
| --- | --- |
| Command | `apps.config`; `fbrcm apps config`. |
| Arguments | required app. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/apps_config.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1554). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`, `--to`, `--update`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1554); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/apps_config.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/apps_config.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1554). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.0.0/apps_config.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1554); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/apps_config.response.schema.json). |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/apps_config.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L331). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1554), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L331). |
| Effects | Level 2; network `conditional`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L1554), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L331). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:apps.config:input`; `urn:fbrcm:schema:cli:1.0.0:command:apps.config:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm apps config <app>](../../docs/CLI.md#L2109); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L331); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `apps.list`

| Column | Audited content |
| --- | --- |
| Command | `apps.list`; `fbrcm apps list`. |
| Arguments | optional project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/apps_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2090). |
| Options | `--cached`, `--filter`, `--no-local-config`, `--platform`, `--profile`, `--project`, `--show-deleted`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2090); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/apps_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/apps_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2090). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.0.0/apps_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2090); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/apps_list.response.schema.json). |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/apps_list.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L432). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2090), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L432). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2090), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L432). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:apps.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:apps.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm apps list [project]](../../docs/CLI.md#L2068); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L432); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `apps.show`

| Column | Audited content |
| --- | --- |
| Command | `apps.show`; `fbrcm apps show`. |
| Arguments | required app. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/apps_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2547). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2547); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/apps_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/apps_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2547). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/apps/commands.go); [response schema](../../schemas/cli/1.0.0/apps_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2547); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/apps_show.response.schema.json). |
| Warnings | `cache.stale`, `cache.write_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/apps_show.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L528). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2547), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L528). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L2547), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L528). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:apps.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:apps.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm apps show <app>](../../docs/CLI.md#L2092); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L528); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.add.gcloud`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.gcloud`; `fbrcm auth add gcloud`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_add_gcloud.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3010). |
| Options | `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3010); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_add_gcloud.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_add_gcloud.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3010). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_add_gcloud.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3010); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_add_gcloud.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3010), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L624). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.gcloud:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth add gcloud <auth-id>](../../docs/CLI.md#L3199); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L624); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.add.google`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.google`; `fbrcm auth add google`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_add_google.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3202). |
| Options | `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3202); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_add_google.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_add_google.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3202). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_add_google.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3202); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_add_google.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3202), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L710). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth add google <auth-id>](../../docs/CLI.md#L3123); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L710); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.add.oauth`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.oauth`; `fbrcm auth add oauth`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3391). |
| Options | `--from`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3391); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3391). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:oauth_credentials`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/auth_add_oauth.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_add_oauth.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3391); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_add_oauth.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3391), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L811). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3391), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L811). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.oauth:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth add oauth <auth-id>](../../docs/CLI.md#L3142); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L811); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.add.service-account`

| Column | Audited content |
| --- | --- |
| Command | `auth.add.service-account`; `fbrcm auth add service-account`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3630). |
| Options | `--from`, `--label`, `--no-local-config`, `--profile`, `--quota-project`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3630); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3630). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:service_account_credentials`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/auth_add_service-account.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_add_service-account.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3630); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_add_service-account.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3630), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L904). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `no`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3630), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L904). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.service-account:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth add service-account <auth-id>](../../docs/CLI.md#L3176); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L904); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.bind`

| Column | Audited content |
| --- | --- |
| Command | `auth.bind`; `fbrcm auth bind`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_bind.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3869). |
| Options | `--auth`, `--no-local-config`, `--profile`, `--project`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3869); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_bind.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_bind.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3869). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_bind.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3869); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_bind.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L3869), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L997). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.bind:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth bind](../../docs/CLI.md#L3278); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L997); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.delete`

| Column | Audited content |
| --- | --- |
| Command | `auth.delete`; `fbrcm auth delete`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4023). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4023); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_delete.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_delete.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4023). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_delete.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4023); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_delete.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4023), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1086). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4023), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1086). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.delete:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth delete <auth-id>](../../docs/CLI.md#L3265); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1086); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.list`

| Column | Audited content |
| --- | --- |
| Command | `auth.list`; `fbrcm auth list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4235). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4235); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4235). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4235); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4235), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1213). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth list](../../docs/CLI.md#L3109); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1213); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.login`

| Column | Audited content |
| --- | --- |
| Command | `auth.login`; `fbrcm auth login`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_login.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4345). |
| Options | `--no-local-config`, `--noopen`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4345); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_login.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_login.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4345). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_login.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4345); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_login.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4345), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1303). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4345), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1303). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.login:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.login:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth login <auth-id>](../../docs/CLI.md#L3231); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1303); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.path`

| Column | Audited content |
| --- | --- |
| Command | `auth.path`; `fbrcm auth path`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4534). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4534); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_path.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_path.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4534). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_path.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4534); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_path.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4534), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1406). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.path:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth path <auth-id>](../../docs/CLI.md#L3255); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1406); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.quota-project.set`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.set`; `fbrcm auth quota-project set`. |
| Arguments | required auth_id, required quota_project_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_quota-project_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4653). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4653); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_quota-project_set.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_quota-project_set.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4653). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_quota-project_set.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4653); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_quota-project_set.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4653), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1501). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.set:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth quota-project show / set / unset](../../docs/CLI.md#L3221); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1501); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.quota-project.show`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.show`; `fbrcm auth quota-project show`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_quota-project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4795). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4795); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_quota-project_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_quota-project_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4795). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_quota-project_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4795); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_quota-project_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4795), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1588). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth quota-project show / set / unset](../../docs/CLI.md#L3221); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1588); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `auth.quota-project.unset`

| Column | Audited content |
| --- | --- |
| Command | `auth.quota-project.unset`; `fbrcm auth quota-project unset`. |
| Arguments | required auth_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4947). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4947); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/auth_quota-project_unset.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4947). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/auth/commands.go); [response schema](../../schemas/cli/1.0.0/auth_quota-project_unset.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4947); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/auth_quota-project_unset.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L4947), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1670). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.quota-project.unset:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm auth quota-project show / set / unset](../../docs/CLI.md#L3221); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1670); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `cache.clear`

| Column | Audited content |
| --- | --- |
| Command | `cache.clear`; `fbrcm cache clear`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/cache_clear.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5115). |
| Options | `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5115); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/cache_clear.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/cache_clear.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5115). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.0.0/cache_clear.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5115); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/cache_clear.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5115), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1757). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5115), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1757). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:input`; `urn:fbrcm:schema:cli:1.0.0:command:cache.clear:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm cache clear](../../docs/CLI.md#L2863); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1757); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `cache.list`

| Column | Audited content |
| --- | --- |
| Command | `cache.list`; `fbrcm cache list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/cache_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5311). |
| Options | `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5311); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/cache_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/cache_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5311). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.0.0/cache_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5311); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/cache_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5311), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1847). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:cache.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:cache.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm cache list](../../docs/CLI.md#L2839); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1847); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `cache.path`

| Column | Audited content |
| --- | --- |
| Command | `cache.path`; `fbrcm cache path`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/cache_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5430). |
| Options | `--kind`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5430); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/cache_path.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/cache_path.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5430). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/cache/commands.go); [response schema](../../schemas/cli/1.0.0/cache_path.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5430); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/cache_path.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5430), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1937). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:cache.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:cache.path:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm cache path](../../docs/CLI.md#L2852); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L1937); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `capabilities`

| Column | Audited content |
| --- | --- |
| Command | `capabilities`; `fbrcm capabilities`. |
| Arguments | optional command (repeated). Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/capabilities.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5549). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5549); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/capabilities.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/capabilities.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5549). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.0.0/capabilities.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5549); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/capabilities.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5549), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2019). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:capabilities:input`; `urn:fbrcm:schema:cli:1.0.0:command:capabilities:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm capabilities [command...]](../../docs/CLI.md#L1180); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2019); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `completion.bash`

| Column | Audited content |
| --- | --- |
| Command | `completion.bash`; `fbrcm completion bash`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/completion_bash.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5660). |
| Options | `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5660); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/completion_bash.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/completion_bash.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5660). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/completion_bash.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5660); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/completion_bash.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5660), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2104). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:input`; `urn:fbrcm:schema:cli:1.0.0:command:completion.bash:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm completion](../../docs/CLI.md#L3461); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2104); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `completion.fish`

| Column | Audited content |
| --- | --- |
| Command | `completion.fish`; `fbrcm completion fish`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/completion_fish.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5773). |
| Options | `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5773); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/completion_fish.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/completion_fish.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5773). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/completion_fish.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5773); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/completion_fish.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5773), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2179). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:input`; `urn:fbrcm:schema:cli:1.0.0:command:completion.fish:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm completion](../../docs/CLI.md#L3461); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2179); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `completion.powershell`

| Column | Audited content |
| --- | --- |
| Command | `completion.powershell`; `fbrcm completion powershell`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/completion_powershell.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5886). |
| Options | `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5886); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/completion_powershell.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/completion_powershell.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5886). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/completion_powershell.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5886); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/completion_powershell.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5886), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2254). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:input`; `urn:fbrcm:schema:cli:1.0.0:command:completion.powershell:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm completion](../../docs/CLI.md#L3461); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2254); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `completion.zsh`

| Column | Audited content |
| --- | --- |
| Command | `completion.zsh`; `fbrcm completion zsh`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/completion_zsh.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5999). |
| Options | `--no-descriptions`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5999); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/completion_zsh.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/completion_zsh.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5999). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/completion_zsh.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5999); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/completion_zsh.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L5999), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2329). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:input`; `urn:fbrcm:schema:cli:1.0.0:command:completion.zsh:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm completion](../../docs/CLI.md#L3461); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2329); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `conditions.add`

| Column | Audited content |
| --- | --- |
| Command | `conditions.add`; `fbrcm conditions add`. |
| Arguments | optional project, required name. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/conditions_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6112). |
| Options | `--change-note`, `--color`, `--draft`, `--dry-run`, `--expression`, `--no-local-config`, `--plan-out`, `--priority`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6112); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/conditions_add.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/conditions_add.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6112). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_add.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6112); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/conditions_add.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/conditions_add.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2404). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6112), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2404). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6112), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2404). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.add:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Condition mutations](../../docs/CLI.md#L1598); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2404); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `conditions.delete`

| Column | Audited content |
| --- | --- |
| Command | `conditions.delete`; `fbrcm conditions delete`. |
| Arguments | optional project, optional condition. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/conditions_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6942). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6942); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/conditions_delete.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/conditions_delete.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6942). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_delete.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6942); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/conditions_delete.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/conditions_delete.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2532). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6942), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2532). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L6942), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2532). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.delete:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Condition mutations](../../docs/CLI.md#L1598); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2532); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `conditions.edit`

| Column | Audited content |
| --- | --- |
| Command | `conditions.edit`; `fbrcm conditions edit`. |
| Arguments | required project, required condition. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/conditions_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7795). |
| Options | `--change-note`, `--color`, `--draft`, `--dry-run`, `--expression`, `--no-color`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7795); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/conditions_edit.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/conditions_edit.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7795). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_edit.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7795); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/conditions_edit.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/conditions_edit.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2660). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7795), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2660). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L7795), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2660). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.edit:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Condition mutations](../../docs/CLI.md#L1598); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2660); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `conditions.list`

| Column | Audited content |
| --- | --- |
| Command | `conditions.list`; `fbrcm conditions list`. |
| Arguments | optional project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/conditions_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8614). |
| Options | `--cached`, `--expr`, `--filter`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8614); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/conditions_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/conditions_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8614). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8614); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/conditions_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8614), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2788). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L8614), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2788). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm conditions list [project]](../../docs/CLI.md#L1556); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2788); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `conditions.move`

| Column | Audited content |
| --- | --- |
| Command | `conditions.move`; `fbrcm conditions move`. |
| Arguments | required project, required condition, required priority. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/conditions_move.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9072). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9072); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/conditions_move.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/conditions_move.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9072). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_move.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9072); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/conditions_move.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/conditions_move.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2888). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9072), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2888). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9072), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2888). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.move:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Condition mutations](../../docs/CLI.md#L1598); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L2888); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `conditions.rename`

| Column | Audited content |
| --- | --- |
| Command | `conditions.rename`; `fbrcm conditions rename`. |
| Arguments | required project, required condition, required new_name. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/conditions_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9867). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9867); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/conditions_rename.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/conditions_rename.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9867). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_rename.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9867); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/conditions_rename.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/conditions_rename.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3016). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9867), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3016). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L9867), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3016). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.rename:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Condition mutations](../../docs/CLI.md#L1598); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3016); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `conditions.show`

| Column | Audited content |
| --- | --- |
| Command | `conditions.show`; `fbrcm conditions show`. |
| Arguments | required project, required condition. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/conditions_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10662). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10662); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/conditions_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/conditions_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10662). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10662); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/conditions_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10662), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3144). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L10662), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3144). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm conditions show <project> <condition>](../../docs/CLI.md#L1582); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3144); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `conditions.validate`

| Column | Audited content |
| --- | --- |
| Command | `conditions.validate`; `fbrcm conditions validate`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/conditions_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11082). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11082); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/conditions_validate.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/conditions_validate.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11082). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/conditions/commands.go); [response schema](../../schemas/cli/1.0.0/conditions_validate.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11082); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/conditions_validate.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11082), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3244). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11082), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3244). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:input`; `urn:fbrcm:schema:cli:1.0.0:command:conditions.validate:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm conditions validate <project>](../../docs/CLI.md#L1688); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3244); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `config.edit`

| Column | Audited content |
| --- | --- |
| Command | `config.edit`; `fbrcm config edit`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/config_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11343). |
| Options | `--editor`, `--full`, `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11343); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/config_edit.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/config_edit.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11343). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | N/A; explicitly registered without successful JSON data. [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_edit.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11343); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/config_edit.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `required`; `external_input_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11343), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3344). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11343), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3344). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.edit:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.edit:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm config edit](../../docs/CLI.md#L3033); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3344); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `config.path`

| Column | Audited content |
| --- | --- |
| Command | `config.path`; `fbrcm config path`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/config_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11491). |
| Options | `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11491); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/config_path.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/config_path.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11491). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_path.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11491); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/config_path.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11491), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3420). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.path:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm config path](../../docs/CLI.md#L2879); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3420); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `config.reset`

| Column | Audited content |
| --- | --- |
| Command | `config.reset`; `fbrcm config reset`. |
| Arguments | optional key. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/config_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11610). |
| Options | `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11610); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/config_reset.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/config_reset.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11610). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_reset.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11610); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/config_reset.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11610), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3502). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11610), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3502). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.reset:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.reset:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm config reset [key]](../../docs/CLI.md#L2989); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3502); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `config.set`

| Column | Audited content |
| --- | --- |
| Command | `config.set`; `fbrcm config set`. |
| Arguments | required key, required value (repeated). Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/config_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11807). |
| Options | `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11807); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/config_set.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/config_set.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11807). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_set.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11807); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/config_set.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11807), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3592). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.set:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm config set <key> <value>...](../../docs/CLI.md#L2934); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3592); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `config.show`

| Column | Audited content |
| --- | --- |
| Command | `config.show`; `fbrcm config show`. |
| Arguments | optional key. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/config_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11955). |
| Options | `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11955); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/config_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/config_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11955). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11955); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/config_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L11955), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3687). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm config show [key]](../../docs/CLI.md#L2897); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3687); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `config.validate`

| Column | Audited content |
| --- | --- |
| Command | `config.validate`; `fbrcm config validate`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/config_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12081). |
| Options | `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12081); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/config_validate.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/config_validate.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12081). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/config/commands.go); [response schema](../../schemas/cli/1.0.0/config_validate.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12081); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/config_validate.response.schema.json). |
| Warnings | `duplicate_binding`, `empty_binding`, `invalid_binding`, `invalid_profile`, `keybinding_conflict`, `legacy_bindings`, `missing_profile`, `project_alias_source`, `repository_scope_required`, `toml_decode`, `unknown_action`, `unknown_block`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/config_validate.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3769). |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12081), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3769). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:config.validate:input`; `urn:fbrcm:schema:cli:1.0.0:command:config.validate:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm config validate](../../docs/CLI.md#L3012); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3769); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `delete`

| Column | Audited content |
| --- | --- |
| Command | `delete`; `fbrcm delete`. |
| Arguments | optional parameter. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12201). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12201); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/delete.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/delete.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12201). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/delete.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/delete/commands.go); [response schema](../../schemas/cli/1.0.0/delete.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12201); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/delete.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/delete.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3859). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12201), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3859). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L12201), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3859). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:delete:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm delete [parameter]](../../docs/CLI.md#L1520); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L3859); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `doctor`

| Column | Audited content |
| --- | --- |
| Command | `doctor`; `fbrcm doctor`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/doctor.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13129). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13129); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/doctor.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/doctor.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13129). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/doctor/commands.go); [response schema](../../schemas/cli/1.0.0/doctor.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13129); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/doctor.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13129), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4023). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:doctor:input`; `urn:fbrcm:schema:cli:1.0.0:command:doctor:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm doctor](../../docs/CLI.md#L2820); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4023); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `draft.change-note`

| Column | Audited content |
| --- | --- |
| Command | `draft.change-note`; `fbrcm draft change-note`. |
| Arguments | required project, optional text. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/draft_change-note.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13338). |
| Options | `--clear`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13338); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/draft_change-note.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/draft_change-note.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13338). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_change-note.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13338); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/draft_change-note.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13338), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4105). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.change-note:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm draft change-note <project> [text]](../../docs/CLI.md#L1797); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4105); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `draft.diff`

| Column | Audited content |
| --- | --- |
| Command | `draft.diff`; `fbrcm draft diff`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/draft_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13552). |
| Options | `--against`, `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13552); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/draft_diff.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/draft_diff.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13552). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_diff.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13552); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/draft_diff.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13552), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4191). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13552), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4191). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.diff:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm draft diff <project>](../../docs/CLI.md#L1808); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4191); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `draft.discard`

| Column | Audited content |
| --- | --- |
| Command | `draft.discard`; `fbrcm draft discard`. |
| Arguments | optional project (repeated). Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/draft_discard.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13880). |
| Options | `--all`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13880); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/draft_discard.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/draft_discard.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13880). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_discard.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13880); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/draft_discard.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13880), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4279). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L13880), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4279). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.discard:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm draft discard [project...]](../../docs/CLI.md#L1864); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4279); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `draft.list`

| Column | Audited content |
| --- | --- |
| Command | `draft.list`; `fbrcm draft list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/draft_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14087). |
| Options | `--filter`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14087); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/draft_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/draft_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14087). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14087); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/draft_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14087), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4372). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm draft list](../../docs/CLI.md#L1747); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4372); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `draft.path`

| Column | Audited content |
| --- | --- |
| Command | `draft.path`; `fbrcm draft path`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/draft_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14208). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14208); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/draft_path.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/draft_path.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14208). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_path.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14208); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/draft_path.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14208), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4466). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.path:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm draft path](../../docs/CLI.md#L1774); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4466); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `draft.publish`

| Column | Audited content |
| --- | --- |
| Command | `draft.publish`; `fbrcm draft publish`. |
| Arguments | optional project (repeated). Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/draft_publish.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14317). |
| Options | `--all`, `--change-note`, `--dry-run`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14317); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/draft_publish.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/draft_publish.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14317). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_publish.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14317); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/draft_publish.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.draft_cleanup_failed`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/draft_publish.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4548). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14317), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4548). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L14317), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4548). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.publish:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm draft publish [project...]](../../docs/CLI.md#L1836); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4548); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `draft.show`

| Column | Audited content |
| --- | --- |
| Command | `draft.show`; `fbrcm draft show`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/draft_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15030). |
| Options | `--no-local-config`, `--profile`, `--raw`, `--stateless`, `--timeout`, `--to`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15030); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/draft_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/draft_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15030). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/draft/commands.go); [response schema](../../schemas/cli/1.0.0/draft_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15030); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/draft_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `destination_conflict_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15030), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4647). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15030), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4647). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:draft.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:draft.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm draft show <project>](../../docs/CLI.md#L1784); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4647); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `duplicate`

| Column | Audited content |
| --- | --- |
| Command | `duplicate`; `fbrcm duplicate`. |
| Arguments | required source, required target. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/duplicate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15198). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15198); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/duplicate.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/duplicate.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15198). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/duplicate/commands.go); [response schema](../../schemas/cli/1.0.0/duplicate.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15198); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/duplicate.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/duplicate.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4736). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15198), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4736). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L15198), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4736). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:duplicate:input`; `urn:fbrcm:schema:cli:1.0.0:command:duplicate:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm duplicate <source> <target>](../../docs/CLI.md#L1399); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4736); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `experiments.delete`

| Column | Audited content |
| --- | --- |
| Command | `experiments.delete`; `fbrcm experiments delete`. |
| Arguments | required project, required experiment_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/experiments_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16010). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16010); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/experiments_delete.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/experiments_delete.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16010). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/experiments_delete.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16010); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/experiments_delete.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16010), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4878). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16010), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4878). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:experiments.delete:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm experiments delete <project> <experiment-id>](../../docs/CLI.md#L2191); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4878); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `experiments.list`

| Column | Audited content |
| --- | --- |
| Command | `experiments.list`; `fbrcm experiments list`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/experiments_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16358). |
| Options | `--cached`, `--filter`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16358); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/experiments_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/experiments_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16358). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/experiments_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16358); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/experiments_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16358), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4970). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16358), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4970). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:experiments.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm experiments list <project>](../../docs/CLI.md#L2151); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L4970); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `experiments.show`

| Column | Audited content |
| --- | --- |
| Command | `experiments.show`; `fbrcm experiments show`. |
| Arguments | required project, required experiment_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/experiments_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16697). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16697); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/experiments_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/experiments_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16697). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/experiments_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16697); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/experiments_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16697), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5070). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L16697), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5070). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:experiments.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm experiments show <project> <experiment-id>](../../docs/CLI.md#L2172); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5070); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `get`

| Column | Audited content |
| --- | --- |
| Command | `get`; `fbrcm get`. |
| Arguments | optional parameter. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/get.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17030). |
| Options | `--all`, `--cached`, `--expr`, `--filter`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17030); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/get.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/get.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17030). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/get.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/get/commands.go); [response schema](../../schemas/cli/1.0.0/get.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17030); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/get.response.schema.json). |
| Warnings | `cache.stale`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/get.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5166). |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17030), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5166). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17030), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5166). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:get:input`; `urn:fbrcm:schema:cli:1.0.0:command:get:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm get [parameter]](../../docs/CLI.md#L1439); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5166); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `groups.add`

| Column | Audited content |
| --- | --- |
| Command | `groups.add`; `fbrcm groups add`. |
| Arguments | required name. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/groups_add.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17562). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17562); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/groups_add.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/groups_add.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17562). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_add.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17562); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/groups_add.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/groups_add.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5349). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17562), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5349). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L17562), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5349). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.add:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.add:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Group mutations](../../docs/CLI.md#L1722); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5349); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `groups.delete`

| Column | Audited content |
| --- | --- |
| Command | `groups.delete`; `fbrcm groups delete`. |
| Arguments | required group. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/groups_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18366). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18366); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/groups_delete.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/groups_delete.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18366). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_delete.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18366); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/groups_delete.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/groups_delete.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5477). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18366), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5477). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L18366), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5477). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.delete:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Group mutations](../../docs/CLI.md#L1722); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5477); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `groups.edit`

| Column | Audited content |
| --- | --- |
| Command | `groups.edit`; `fbrcm groups edit`. |
| Arguments | required group. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/groups_edit.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19179). |
| Options | `--change-note`, `--description`, `--draft`, `--dry-run`, `--no-description`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19179); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/groups_edit.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/groups_edit.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19179). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_edit.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19179); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/groups_edit.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/groups_edit.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5605). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19179), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5605). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19179), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5605). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.edit:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Group mutations](../../docs/CLI.md#L1722); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5605); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `groups.list`

| Column | Audited content |
| --- | --- |
| Command | `groups.list`; `fbrcm groups list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/groups_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19993). |
| Options | `--cached`, `--filter`, `--no-local-config`, `--profile`, `--project`, `--search`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19993); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/groups_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/groups_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19993). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19993); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/groups_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19993), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5733). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L19993), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5733). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm groups list](../../docs/CLI.md#L1702); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5733); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `groups.rename`

| Column | Audited content |
| --- | --- |
| Command | `groups.rename`; `fbrcm groups rename`. |
| Arguments | required group, required new_name. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/groups_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20433). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20433); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/groups_rename.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/groups_rename.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20433). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/groups/commands.go); [response schema](../../schemas/cli/1.0.0/groups_rename.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20433); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/groups_rename.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/groups_rename.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5833). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20433), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5833). |
| Effects | Level 3; network `required`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L20433), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5833). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:input`; `urn:fbrcm:schema:cli:1.0.0:command:groups.rename:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [Group mutations](../../docs/CLI.md#L1722); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5833); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `help`

| Column | Audited content |
| --- | --- |
| Command | `help`; `fbrcm help`. |
| Arguments | optional command (repeated). Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/help.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21233). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21233); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/help.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/help.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21233). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/help.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21233); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/help.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21233), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5961). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:help:input`; `urn:fbrcm:schema:cli:1.0.0:command:help:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm help [command]](../../docs/CLI.md#L3493); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L5961); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `hooks.fingerprint`

| Column | Audited content |
| --- | --- |
| Command | `hooks.fingerprint`; `fbrcm hooks fingerprint`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/hooks_fingerprint.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21342). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21342); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/hooks_fingerprint.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/hooks_fingerprint.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21342). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.0.0/hooks_fingerprint.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21342); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/hooks_fingerprint.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21342), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6055). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:input`; `urn:fbrcm:schema:cli:1.0.0:command:hooks.fingerprint:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm hooks fingerprint](../../docs/CLI.md#L3078); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6055); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `hooks.status`

| Column | Audited content |
| --- | --- |
| Command | `hooks.status`; `fbrcm hooks status`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/hooks_status.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21450). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21450); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/hooks_status.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/hooks_status.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21450). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.0.0/hooks_status.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21450); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/hooks_status.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21450), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6137). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:input`; `urn:fbrcm:schema:cli:1.0.0:command:hooks.status:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm hooks status](../../docs/CLI.md#L3067); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6137); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `hooks.trust`

| Column | Audited content |
| --- | --- |
| Command | `hooks.trust`; `fbrcm hooks trust`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/hooks_trust.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21557). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21557); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/hooks_trust.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/hooks_trust.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21557). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.0.0/hooks_trust.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21557); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/hooks_trust.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21557), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6219). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21557), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6219). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:input`; `urn:fbrcm:schema:cli:1.0.0:command:hooks.trust:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm hooks trust](../../docs/CLI.md#L3084); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6219); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `hooks.untrust`

| Column | Audited content |
| --- | --- |
| Command | `hooks.untrust`; `fbrcm hooks untrust`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/hooks_untrust.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21724). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21724); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/hooks_untrust.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/hooks_untrust.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21724). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/hooks/commands.go); [response schema](../../schemas/cli/1.0.0/hooks_untrust.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21724); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/hooks_untrust.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21724), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6309). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:input`; `urn:fbrcm:schema:cli:1.0.0:command:hooks.untrust:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm hooks untrust](../../docs/CLI.md#L3098); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6309); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `mcp`

| Column | Audited content |
| --- | --- |
| Command | `mcp`; `fbrcm mcp`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/mcp.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21860). |
| Options | `--allow-hooks`, `--allow-writes`, `--auth-timeout`, `--browser-auth`, `--confirmation`, `--no-local-config`, `--profile`, `--request-timeout`, `--stateless`, `--timeout`, `--toolsets`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21860); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/mcp.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/mcp.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21860). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | N/A; explicitly registered without successful JSON data. [DTO registrations](../../cli/app/mcp.go); [response schema](../../schemas/cli/1.0.0/mcp.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21860); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/mcp.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 0; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L21860), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6395). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:mcp:input`; `urn:fbrcm:schema:cli:1.0.0:command:mcp:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [MCP server](../../docs/CLI.md#L945); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6395); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `personalizations.list`

| Column | Audited content |
| --- | --- |
| Command | `personalizations.list`; `fbrcm personalizations list`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/personalizations_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22017). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22017); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/personalizations_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/personalizations_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22017). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/personalizations_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22017); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/personalizations_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22017), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6482). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22017), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6482). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:personalizations.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm personalizations list <project>](../../docs/CLI.md#L2259); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6482); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `personalizations.show`

| Column | Audited content |
| --- | --- |
| Command | `personalizations.show`; `fbrcm personalizations show`. |
| Arguments | required project, required personalization_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/personalizations_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22430). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22430); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/personalizations_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/personalizations_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22430). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/personalizations_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22430); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/personalizations_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22430), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6578). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22430), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6578). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:personalizations.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm personalizations show <project> <personalization-id>](../../docs/CLI.md#L2273); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6578); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `plan.show`

| Column | Audited content |
| --- | --- |
| Command | `plan.show`; `fbrcm plan show`. |
| Arguments | required plan. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/plan_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22850). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22850); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/plan_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/plan_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22850). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:publication_plan`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/plan_show.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/plan/commands.go); [response schema](../../schemas/cli/1.0.0/plan_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22850); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/plan_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22850), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6678). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:plan.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:plan.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm plan show <plan>](../../docs/CLI.md#L1309); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6678); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `plan.validate`

| Column | Audited content |
| --- | --- |
| Command | `plan.validate`; `fbrcm plan validate`. |
| Arguments | required plan. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/plan_validate.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22968). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22968); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/plan_validate.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/plan_validate.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22968). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:publication_plan`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/plan_validate.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/plan/commands.go); [response schema](../../schemas/cli/1.0.0/plan_validate.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22968); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/plan_validate.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L22968), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6767). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:plan.validate:input`; `urn:fbrcm:schema:cli:1.0.0:command:plan.validate:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm plan validate <plan>](../../docs/CLI.md#L1318); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6767); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `profile`

| Column | Audited content |
| --- | --- |
| Command | `profile`; `fbrcm profile`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/profile.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23086). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23086); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/profile.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/profile.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23086). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23086); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/profile.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23086), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6856). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm profile](../../docs/CLI.md#L3291); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6856); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `profile.delete`

| Column | Audited content |
| --- | --- |
| Command | `profile.delete`; `fbrcm profile delete`. |
| Arguments | required profile. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/profile_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23193). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23193); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/profile_delete.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/profile_delete.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23193). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_delete.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23193); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/profile_delete.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23193), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6934). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23193), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6934). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.delete:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm profile delete <profile>](../../docs/CLI.md#L3335); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L6934); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `profile.list`

| Column | Audited content |
| --- | --- |
| Command | `profile.list`; `fbrcm profile list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/profile_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23431). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23431); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/profile_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/profile_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23431). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23431); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/profile_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23431), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7027). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm profile list](../../docs/CLI.md#L3295); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7027); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `profile.path`

| Column | Audited content |
| --- | --- |
| Command | `profile.path`; `fbrcm profile path`. |
| Arguments | required profile. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/profile_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23539). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23539); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/profile_path.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/profile_path.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23539). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_path.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23539); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/profile_path.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23539), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7109). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.path:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm profile path <profile>](../../docs/CLI.md#L3325); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7109); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `profile.rename`

| Column | Audited content |
| --- | --- |
| Command | `profile.rename`; `fbrcm profile rename`. |
| Arguments | required old_name, required new_name. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/profile_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23655). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23655); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/profile_rename.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/profile_rename.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23655). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_rename.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23655); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/profile_rename.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23655), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7191). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.rename:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm profile rename <old-name> <new-name>](../../docs/CLI.md#L3315); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7191); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `profile.switch`

| Column | Audited content |
| --- | --- |
| Command | `profile.switch`; `fbrcm profile switch`. |
| Arguments | required name. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/profile_switch.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23816). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23816); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/profile_switch.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/profile_switch.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23816). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/profile/commands.go); [response schema](../../schemas/cli/1.0.0/profile_switch.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23816); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/profile_switch.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23816), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7280). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:input`; `urn:fbrcm:schema:cli:1.0.0:command:profile.switch:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm profile switch <name>](../../docs/CLI.md#L3305); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7280); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.defaults`

| Column | Audited content |
| --- | --- |
| Command | `project.defaults`; `fbrcm project defaults`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_defaults.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23920). |
| Options | `--format`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23920); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_defaults.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_defaults.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23920). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_defaults.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23920); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_defaults.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23920), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7366). |
| Effects | Level 2; network `required`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L23920), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7366). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.defaults:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project defaults <project>](../../docs/CLI.md#L1973); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7366); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.export`

| Column | Audited content |
| --- | --- |
| Command | `project.export`; `fbrcm project export`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_export.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24236). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24236); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_export.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_export.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24236). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_export.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24236); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_export.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24236), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7505). |
| Effects | Level 2; network `required`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24236), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7505). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.export:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.export:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project export <project>](../../docs/CLI.md#L1953); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7505); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.import`

| Column | Audited content |
| --- | --- |
| Command | `project.import`; `fbrcm project import`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24542). |
| Options | `--change-note`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--from`, `--group`, `--keep-portable-conditions-only`, `--merge`, `--merge-resolve`, `--no-local-config`, `--override`, `--plan-out`, `--profile`, `--remove-all-conditions`, `--search`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24542); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_import.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_import.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24542). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config_import`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/project_import.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_import.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24542); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_import.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/project_import.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7644). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24542), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7644). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L24542), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7644). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.import:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.import:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project import <project>](../../docs/CLI.md#L1989); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7644); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.open`

| Column | Audited content |
| --- | --- |
| Command | `project.open`; `fbrcm project open`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_open.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25491). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25491); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_open.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_open.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25491). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_open.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25491); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_open.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `browser_launch_suppressed_and_oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25491), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7763). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25491), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7763). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.open:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.open:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project open <project>](../../docs/CLI.md#L1945); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7763); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.quota-project.set`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.set`; `fbrcm project quota-project set`. |
| Arguments | required project, required quota_project_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_quota-project_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25746). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25746); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_quota-project_set.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_quota-project_set.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25746). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_quota-project_set.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25746); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_quota-project_set.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25746), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7859). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.set:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project quota-project show / set / unset](../../docs/CLI.md#L1935); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7859); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.quota-project.show`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.show`; `fbrcm project quota-project show`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_quota-project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25887). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25887); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_quota-project_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_quota-project_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25887). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_quota-project_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25887); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_quota-project_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L25887), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7946). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project quota-project show / set / unset](../../docs/CLI.md#L1935); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L7946); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.quota-project.unset`

| Column | Audited content |
| --- | --- |
| Command | `project.quota-project.unset`; `fbrcm project quota-project unset`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_quota-project_unset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26041). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26041); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_quota-project_unset.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_quota-project_unset.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26041). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_quota-project_unset.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26041); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_quota-project_unset.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26041), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8028). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.quota-project.unset:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project quota-project show / set / unset](../../docs/CLI.md#L1935); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8028); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.show`

| Column | Audited content |
| --- | --- |
| Command | `project.show`; `fbrcm project show`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26211). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26211); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26211). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26211); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26211), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8114). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26211), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8114). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project show <project>](../../docs/CLI.md#L1880); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8114); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.templates.set`

| Column | Audited content |
| --- | --- |
| Command | `project.templates.set`; `fbrcm project templates set`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_templates_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26552). |
| Options | `--no-local-config`, `--primary`, `--profile`, `--stateless`, `--templates`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26552); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_templates_set.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_templates_set.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26552). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_templates_set.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26552); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_templates_set.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26552), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8210). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.templates.set:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project templates set <project>](../../docs/CLI.md#L1911); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8210); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `project.templates.show`

| Column | Audited content |
| --- | --- |
| Command | `project.templates.show`; `fbrcm project templates show`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/project_templates_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26707). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26707); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/project_templates_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/project_templates_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26707). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/project/commands.go); [response schema](../../schemas/cli/1.0.0/project_templates_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26707); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/project_templates_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26707), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8292). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:project.templates.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm project templates show <project>](../../docs/CLI.md#L1901); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8292); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.aliases.import`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.import`; `fbrcm projects aliases import`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_aliases_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26826). |
| Options | `--conflict`, `--dry-run`, `--from`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26826); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_aliases_import.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_aliases_import.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26826). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_aliases_import.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26826); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_aliases_import.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26826), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8375). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L26826), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8375). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.import:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects aliases import --from <path>](../../docs/CLI.md#L2727); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8375); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.aliases.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.list`; `fbrcm projects aliases list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_aliases_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27045). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27045); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_aliases_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_aliases_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27045). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_aliases_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27045); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_aliases_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27045), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8465). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects aliases list](../../docs/CLI.md#L2676); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8465); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.aliases.remove`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.remove`; `fbrcm projects aliases remove`. |
| Arguments | required alias. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_aliases_remove.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27155). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27155); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_aliases_remove.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_aliases_remove.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27155). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_aliases_remove.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27155); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_aliases_remove.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27155), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8555). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27155), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8555). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.remove:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects aliases remove <alias>](../../docs/CLI.md#L2710); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8555); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.aliases.set`

| Column | Audited content |
| --- | --- |
| Command | `projects.aliases.set`; `fbrcm projects aliases set`. |
| Arguments | required alias, required project_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_aliases_set.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27345). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27345); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_aliases_set.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_aliases_set.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27345). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_aliases_set.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27345); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_aliases_set.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27345), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8648). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27345), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8648). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.aliases.set:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects aliases set <alias> <project-id>](../../docs/CLI.md#L2691); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8648); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.diff`

| Column | Audited content |
| --- | --- |
| Command | `projects.diff`; `fbrcm projects diff`. |
| Arguments | required source_project, required target_project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27541). |
| Options | `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27541); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_diff.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_diff.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27541). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_diff.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27541); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_diff.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27541), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8738). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27541), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8738). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.diff:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects diff <source-project> <target-project>](../../docs/CLI.md#L2608); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8738); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.forget`

| Column | Audited content |
| --- | --- |
| Command | `projects.forget`; `fbrcm projects forget`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_forget.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27888). |
| Options | `--expr`, `--filter`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27888); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_forget.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_forget.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27888). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_forget.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27888); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_forget.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27888), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8838). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L27888), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8838). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.forget:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects forget](../../docs/CLI.md#L2596); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8838); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.list`

| Column | Audited content |
| --- | --- |
| Command | `projects.list`; `fbrcm projects list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28136). |
| Options | `--expr`, `--filter`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`, `--url`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28136); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28136). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28136); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28136), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8932). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28136), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8932). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects list](../../docs/CLI.md#L2559); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L8932); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.path`

| Column | Audited content |
| --- | --- |
| Command | `projects.path`; `fbrcm projects path`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28501). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28501); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_path.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_path.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28501). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_path.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28501); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_path.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28501), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9042). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.path:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects path](../../docs/CLI.md#L2751); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9042); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.promote`

| Column | Audited content |
| --- | --- |
| Command | `projects.promote`; `fbrcm projects promote`. |
| Arguments | required source_project, required target_project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_promote.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28610). |
| Options | `--all`, `--change-note`, `--conditions`, `--dry-run`, `--expr`, `--filter`, `--group`, `--interactive`, `--no-local-config`, `--parameters`, `--plan-out`, `--profile`, `--prune`, `--search`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28610); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_promote.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_promote.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28610). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_promote.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28610); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_promote.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/projects_promote.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9124). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28610), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9124). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L28610), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9124). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.promote:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects promote <source-project> <target-project>](../../docs/CLI.md#L2635); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9124); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.reset`

| Column | Audited content |
| --- | --- |
| Command | `projects.reset`; `fbrcm projects reset`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29401). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29401); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_reset.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_reset.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29401). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_reset.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29401); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_reset.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29401), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9238). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29401), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9238). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.reset:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects reset](../../docs/CLI.md#L2761); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9238); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `projects.update`

| Column | Audited content |
| --- | --- |
| Command | `projects.update`; `fbrcm projects update`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/projects_update.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29603). |
| Options | `--auth`, `--expr`, `--filter`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--url`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29603); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/projects_update.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/projects_update.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29603). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/projects/commands.go); [response schema](../../schemas/cli/1.0.0/projects_update.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29603); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/projects_update.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29603), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9328). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `no`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29603), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9328). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:projects.update:input`; `urn:fbrcm:schema:cli:1.0.0:command:projects.update:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm projects update](../../docs/CLI.md#L2580); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9328); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `rollouts.delete`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.delete`; `fbrcm rollouts delete`. |
| Arguments | required project, required rollout_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/rollouts_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29828). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29828); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/rollouts_delete.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/rollouts_delete.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29828). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/rollouts_delete.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29828); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/rollouts_delete.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29828), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9420). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L29828), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9420). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:rollouts.delete:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm rollouts delete <project> <rollout-id>](../../docs/CLI.md#L2243); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9420); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `rollouts.list`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.list`; `fbrcm rollouts list`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/rollouts_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30176). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30176); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/rollouts_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/rollouts_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30176). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/rollouts_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30176); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/rollouts_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30176), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9512). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30176), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9512). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:rollouts.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm rollouts list <project>](../../docs/CLI.md#L2209); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9512); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `rollouts.show`

| Column | Audited content |
| --- | --- |
| Command | `rollouts.show`; `fbrcm rollouts show`. |
| Arguments | required project, required rollout_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/rollouts_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30503). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--update`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30503); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/rollouts_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/rollouts_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30503). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/managedfeatures/commands.go); [response schema](../../schemas/cli/1.0.0/rollouts_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30503); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/rollouts_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30503), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9608). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30503), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9608). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:rollouts.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm rollouts show <project> <rollout-id>](../../docs/CLI.md#L2226); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9608); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `root`

| Column | Audited content |
| --- | --- |
| Command | `root`; `fbrcm`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/root.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30836). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--version`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30836); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/root.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/root.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30836). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/app/root.go); [response schema](../../schemas/cli/1.0.0/root.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30836); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/root.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30836), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9704). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:root:input`; `urn:fbrcm:schema:cli:1.0.0:command:root:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm](../../docs/CLI.md#L1142); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9704); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `schema.list`

| Column | Audited content |
| --- | --- |
| Command | `schema.list`; `fbrcm schema list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/schema_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30979). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30979); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/schema_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/schema_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30979). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.0.0/schema_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30979); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/schema_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L30979), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9794). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:schema.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:schema.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm schema list](../../docs/CLI.md#L1271); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9794); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `schema.show`

| Column | Audited content |
| --- | --- |
| Command | `schema.show`; `fbrcm schema show`. |
| Arguments | required schema_id. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/schema_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31082). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31082); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/schema_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/schema_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31082). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/meta/commands.go); [response schema](../../schemas/cli/1.0.0/schema_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31082); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/schema_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31082), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9876). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:schema.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:schema.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm schema show <schema-id>](../../docs/CLI.md#L1276); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9876); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `setup`

| Column | Audited content |
| --- | --- |
| Command | `setup`; `fbrcm setup`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/setup.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31193). |
| Options | `--no-local-config`, `--noopen`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31193); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/setup.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/setup.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31193). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | N/A; explicitly registered without successful JSON data. [DTO registrations](../../cli/commands/setup/commands.go); [response schema](../../schemas/cli/1.0.0/setup.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31193); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/setup.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `required`; `guided_setup_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31193), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9961). |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31193), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9961). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:setup:input`; `urn:fbrcm:schema:cli:1.0.0:command:setup:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm setup](../../docs/CLI.md#L2773); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L9961); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `theme`

| Column | Audited content |
| --- | --- |
| Command | `theme`; `fbrcm theme`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/theme.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31323). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31323); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/theme.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/theme.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31323). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31323); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/theme.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31323), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10050). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm theme](../../docs/CLI.md#L3347); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10050); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `theme.delete`

| Column | Audited content |
| --- | --- |
| Command | `theme.delete`; `fbrcm theme delete`. |
| Arguments | required theme. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/theme_delete.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31431). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31431); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/theme_delete.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/theme_delete.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31431). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_delete.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31431); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/theme_delete.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `confirmation_required_without_bypass`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31431), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10132). |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31431), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10132). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.delete:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm theme delete <theme>](../../docs/CLI.md#L3409); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10132); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `theme.import`

| Column | Audited content |
| --- | --- |
| Command | `theme.import`; `fbrcm theme import`. |
| Arguments | optional source. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/theme_import.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31625). |
| Options | `--name`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31625); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/theme_import.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/theme_import.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31625). Selectors absent from those properties are N/A. |
| Stdin | toml_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:theme`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/theme_import.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_import.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31625); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/theme_import.response.schema.json). |
| Warnings | `theme.already_exists`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/theme_import.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10215). |
| Interaction | Mode `optional`; `missing_input_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31625), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10215). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31625), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10215). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.import:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.import:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm theme import [source]](../../docs/CLI.md#L3421); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10215); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `theme.list`

| Column | Audited content |
| --- | --- |
| Command | `theme.list`; `fbrcm theme list`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/theme_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31834). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31834); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/theme_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/theme_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31834). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31834); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/theme_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31834), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10324). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm theme list](../../docs/CLI.md#L3356); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10324); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `theme.path`

| Column | Audited content |
| --- | --- |
| Command | `theme.path`; `fbrcm theme path`. |
| Arguments | required theme. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/theme_path.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31941). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31941); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/theme_path.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/theme_path.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31941). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_path.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31941); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/theme_path.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L31941), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10406). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.path:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.path:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm theme path <theme>](../../docs/CLI.md#L3396); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10406); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `theme.rename`

| Column | Audited content |
| --- | --- |
| Command | `theme.rename`; `fbrcm theme rename`. |
| Arguments | required old_name, required new_name. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/theme_rename.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32055). |
| Options | `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32055); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/theme_rename.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/theme_rename.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32055). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_rename.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32055); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/theme_rename.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32055), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10488). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.rename:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm theme rename <old-name> <new-name>](../../docs/CLI.md#L3401); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10488); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `theme.reset`

| Column | Audited content |
| --- | --- |
| Command | `theme.reset`; `fbrcm theme reset`. |
| Arguments | No positional arguments. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/theme_reset.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32210). |
| Options | `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32210); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/theme_reset.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/theme_reset.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32210). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_reset.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32210); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/theme_reset.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32210), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10577). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.reset:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm theme reset](../../docs/CLI.md#L3382); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10577); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `theme.switch`

| Column | Audited content |
| --- | --- |
| Command | `theme.switch`; `fbrcm theme switch`. |
| Arguments | required name. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/theme_switch.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32356). |
| Options | `--no-local-config`, `--profile`, `--scope`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32356); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/theme_switch.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/theme_switch.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32356). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../cli/commands/theme/commands.go); [response schema](../../schemas/cli/1.0.0/theme_switch.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32356); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/theme_switch.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | N/A; no interactive branch or launch. |
| Effects | Level 1; network `none`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32356), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10655). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:input`; `urn:fbrcm:schema:cli:1.0.0:command:theme.switch:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm theme switch <name>](../../docs/CLI.md#L3367); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10655); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `update`

| Column | Audited content |
| --- | --- |
| Command | `update`; `fbrcm update`. |
| Arguments | optional parameter. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/update.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32517). |
| Options | `--change-note`, `--condition`, `--description`, `--draft`, `--dry-run`, `--expr`, `--filter`, `--group`, `--name`, `--no-group`, `--no-local-config`, `--plan-out`, `--profile`, `--project`, `--remove-all-conditional-values`, `--remove-conditional-value`, `--search`, `--stateless`, `--timeout`, `--type`, `--use-in-app-default`, `--value`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32517); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/update.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/update.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32517). Selectors absent from those properties are N/A. |
| Stdin | json_document; concrete schema `urn:fbrcm:schema:cli:1.0.0:stdin:remote_config`; source precedence/restrictions: [input schema](../../schemas/cli/1.0.0/update.input.schema.json). |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/update/commands.go); [response schema](../../schemas/cli/1.0.0/update.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32517); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/update.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.non_atomic`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/update.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10735). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32517), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10735). |
| Effects | Level 3; network `conditional`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L32517), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10735). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:update:input`; `urn:fbrcm:schema:cli:1.0.0:command:update:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm update [parameter]](../../docs/CLI.md#L1464); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10735); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `versions.blame`

| Column | Audited content |
| --- | --- |
| Command | `versions.blame`; `fbrcm versions blame`. |
| Arguments | required project, required parameter. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/versions_blame.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33545). |
| Options | `--all`, `--at`, `--limit`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33545); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/versions_blame.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/versions_blame.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33545). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_blame.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33545); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/versions_blame.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33545), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10884). |
| Effects | Level 2; network `required`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33545), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10884). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.blame:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.blame:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm versions blame <project> <parameter>](../../docs/CLI.md#L2336); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10884); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `versions.diff`

| Column | Audited content |
| --- | --- |
| Command | `versions.diff`; `fbrcm versions diff`. |
| Arguments | required project, required from, optional to. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/versions_diff.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33838). |
| Options | `--cached`, `--conditions`, `--expr`, `--filter`, `--group`, `--no-local-config`, `--parameters`, `--profile`, `--search`, `--side-by-side`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33838); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/versions_diff.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/versions_diff.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33838). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_diff.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33838); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/versions_diff.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33838), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10975). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L33838), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10975). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.diff:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm versions diff <project> <from> [<to>]](../../docs/CLI.md#L2444); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L10975); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `versions.export`

| Column | Audited content |
| --- | --- |
| Command | `versions.export`; `fbrcm versions export`. |
| Arguments | required project, required version. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/versions_export.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34291). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--to`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34291); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/versions_export.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/versions_export.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34291). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_export.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34291); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/versions_export.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34291), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11085). |
| Effects | Level 2; network `conditional`; destructive `true`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34291), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11085). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.export:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.export:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm versions export <project> <version>](../../docs/CLI.md#L2472); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11085); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `versions.list`

| Column | Audited content |
| --- | --- |
| Command | `versions.list`; `fbrcm versions list`. |
| Arguments | required project. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/versions_list.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34738). |
| Options | `--all`, `--before`, `--cached`, `--limit`, `--no-local-config`, `--profile`, `--since`, `--stateless`, `--timeout`, `--until`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34738); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/versions_list.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/versions_list.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34738). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_list.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34738); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/versions_list.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34738), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11191). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L34738), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11191). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.list:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.list:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm versions list <project>](../../docs/CLI.md#L2399); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11191); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `versions.restore`

| Column | Audited content |
| --- | --- |
| Command | `versions.restore`; `fbrcm versions restore`. |
| Arguments | required project, required version. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/versions_restore.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35066). |
| Options | `--change-note`, `--dry-run`, `--no-local-config`, `--plan-out`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35066); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/versions_restore.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/versions_restore.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35066). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_restore.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35066); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/versions_restore.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/versions_restore.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11291). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35066), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11291). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35066), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11291). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.restore:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm versions restore <project> <version>](../../docs/CLI.md#L2527); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11291); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `versions.rollback`

| Column | Audited content |
| --- | --- |
| Command | `versions.rollback`; `fbrcm versions rollback`. |
| Arguments | required project, required version. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/versions_rollback.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35694). |
| Options | `--dry-run`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`, `--yes`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35694); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/versions_rollback.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/versions_rollback.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35694). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_rollback.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35694); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/versions_rollback.response.schema.json). |
| Warnings | `publication.cache_stale`, `publication.post_publish_hook_failed`; typed details/remediation and reachability: [response schema](../../schemas/cli/1.0.0/versions_rollback.response.schema.json), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11390). |
| Interaction | Mode `optional`; `declared_conditions_return_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35694), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11390). |
| Effects | Level 3; network `required`; destructive `true`; idempotency `conditional`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L35694), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11390). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.rollback:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm versions rollback <project> <version>](../../docs/CLI.md#L2488); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11390); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

### `versions.show`

| Column | Audited content |
| --- | --- |
| Command | `versions.show`; `fbrcm versions show`. |
| Arguments | required project, required version. Arity, normalization, grammar and constraints: [input schema](../../schemas/cli/1.0.0/versions_show.input.schema.json); [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36273). |
| Options | `--cached`, `--no-local-config`, `--profile`, `--stateless`, `--timeout`. Exact types, defaults, aliases, repetition, required/effective predicates: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36273); dependencies and exclusions: [input schema](../../schemas/cli/1.0.0/versions_show.input.schema.json). |
| Selection | All literal/mode/default/cross-source matching, precedence and typed lookup rules: [input schema](../../schemas/cli/1.0.0/versions_show.input.schema.json), its referenced semantic definitions and [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36273). Selectors absent from those properties are N/A. |
| Stdin | N/A; no JSON-document stdin support. |
| Success | Every registered DTO and success/empty/no-op/status variant: [DTO registrations](../../ops/workflows/versions/commands.go); [response schema](../../schemas/cli/1.0.0/versions_show.response.schema.json). |
| Failure | Closed reachable codes: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36273); typed categories/details, retryability and statuses: shared problem schema and [response schema](../../schemas/cli/1.0.0/versions_show.response.schema.json). |
| Warnings | N/A; warnings must be empty. |
| Interaction | Mode `optional`; `oauth_authorization_returns_interaction`. Exact triggers/bypasses and no-prompt/launch proof: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36273), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11493). |
| Effects | Level 2; network `conditional`; destructive `false`; idempotency `yes`. All effects, conditions and exceptions: [capability](../../cli/app/testdata/contract_v1_capabilities_detailed.golden.json#L36273), [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11493). |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:versions.show:input`; `urn:fbrcm:schema:cli:1.0.0:command:versions.show:response`. Shared transitive resources are inventoried in Appendix B. |
| Documentation | [fbrcm versions show <project> <version>](../../docs/CLI.md#L2429); [shared machine contract](../../docs/cli-contract.md). |
| Evidence | [15-class evidence](../../cli/app/testdata/contract_v1_audit_evidence.golden.json#L11493); exhaustive discovery/arity/unknown-option/DTO/schema/docs gates and applicable command-local tests and replay scenarios referenced there. |
| Verdict | **PASS.** Shared findings above are closed and final gates pass. |

## Appendix B: schema file and ID inventory

| File | Unique `$id` |
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
