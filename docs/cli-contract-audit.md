# CLI JSON contract audit standard

Standard version: `1.0.0`

This document defines the complete, finite acceptance standard for the fbrcm
CLI JSON contract. It applies to the versioned machine interface enabled by
the global `--json` flag. It is normative for contract reviews and for every
future executable CLI command or machine-contract change.

The purpose of this standard is to make an audit reproducible. An auditor does
not decide what “authoritative” means while conducting an audit. The term,
scope, criteria, evidence, and verdicts are fixed below.

Normative words `MUST`, `MUST NOT`, `REQUIRED`, `SHOULD`, and `MAY` use the
meanings defined by RFC 2119 as updated by RFC 8174 when written in uppercase.
Only `MUST`, `MUST NOT`, and `REQUIRED` criteria affect the pass/fail verdict.

## 1. Fixed definitions

### 1.1 Machine contract

The **machine contract** is the versioned interface formed by all of these
artifacts together:

- executable Cobra command paths available with global `--json`;
- detailed and compact capability records;
- normalized invocation schemas;
- response, envelope, error, capability, semantic, and stdin schemas;
- JSON-mode runtime behavior, including exit status and stdout/stderr rules;
- structured success data, problems, warnings, and remediation vectors;
- declared effects, network access, destructiveness, idempotency, stdin, and
  interaction behavior;
- the machine-interface statements in `docs/cli-contract.md` and the matching
  command descriptions in `docs/CLI.md`.

No one artifact silently overrides another. A disagreement between any two
applicable artifacts is a contract defect.

### 1.2 Normalized invocation

A **normalized invocation** is the `arguments`, `options`, and `stdin` object
described by a command input schema. It is the planner-facing representation
of documented argv and stdin, not an alternative wire protocol accepted by the
binary.

### 1.3 Valid invocation

A **valid invocation** is a normalized invocation accepted by its input schema.
It may still produce a typed environmental, selection, conflict, validation,
interaction, remote-service, or I/O result. Input-schema validity does not
promise that a requested external resource exists or that Firebase will accept
a candidate.

If a check cannot be completed before execution, the schema MUST identify the
runtime validation boundary with a documented semantic annotation. Examples
include expression compilation, duration overflow, remote candidate
validation, and resource existence.

### 1.4 Effective option

An option is **effective** when changing its supplied value can change the
machine invocation's result, effects, interaction, or selected work.

- An option accepted by Cobra but never applied in JSON mode MUST have
  `effective: false` and `x-fbrcm-effective: false`.
- An option applied only under stated conditions MUST publish identical
  `effective_when` and `x-fbrcm-effective-when` predicates.
- An option forbidden in JSON mode MUST be rejected by the input schema and
  MUST NOT simultaneously be described as effective for machine use.

Human table layout, coloring, browser opening, editor choice, file pickers, and
other human-only presentation do not make an option effective in JSON mode.

### 1.5 Selector

A **selector** is any argument or option whose value chooses projects,
templates, parameters, groups, conditions, versions, managed features,
personalizations, drafts, profiles, auth records, aliases, configuration keys,
or command paths.

Every selector MUST publish, as applicable:

- accepted grammar and normalization;
- candidate fields;
- exactness and case sensitivity;
- prefix or mode interpretation;
- precedence and ambiguity behavior;
- repeated-value and cross-selector composition;
- default selection when omitted;
- target-prefix and canonicalization behavior;
- resource-existence validation boundary and typed not-found/ambiguous errors.

A plain string schema is sufficient only for a literal, case-sensitive value
that has no normalization, fallback, ambiguity, composition, or lookup
semantics beyond direct equality. Otherwise the matching semantics MUST be
machine-readable.

### 1.6 Reachable result

A **reachable result** is a success status, item status, outcome, error code,
warning code, or response variant that current runtime control flow can emit
for at least one invocation and runtime state.

Every closed enum value in a response schema MUST be reachable and covered by
a test, or MUST be explicitly marked and documented as a reserved value.
Schemas MAY contain explicitly documented open extension points. An open
extension point is not permission to advertise known but unreachable values.

### 1.7 Authoritative

For this project, the CLI JSON contract is **authoritative for agents** if and
only if all acceptance criteria in section 7 pass for the complete command
inventory and there are no unresolved findings under this standard.

“Authoritative” means all of the following, no more and no less:

1. An agent can discover every executable machine operation.
2. An agent can construct every supported invocation without scraping help or
   source code.
3. The schemas reject combinations the runtime rejects before environmental or
   explicitly annotated runtime validation.
4. The schemas and capabilities describe selection, effects, interaction, and
   results accurately enough to plan safely.
5. Every emitted JSON document validates against the advertised schema and its
   process status has the documented meaning.
6. Documentation, generated artifacts, and runtime agree.

This is a conformance statement about the repository state audited against this
standard version. It is not a claim that Firebase, operating systems, or future
code cannot change. A later contract-affecting change must pass the gate again.

## 2. Frozen audit scope

An audit MUST record the repository revision and this standard version before
inspection begins. Its scope is then frozen.

### 2.1 Included

The audit includes:

- the root operation and every executable Cobra descendant, including help and
  completion operations;
- all positional, local, persistent, inherited, and synthetic machine options;
- global JSON envelope behavior for success and every failure stage;
- normalized JSON-document stdin explicitly published by capabilities;
- all success, empty, no-op, dry-run, draft, partial, interaction, warning,
  artifact, and typed-failure branches that apply to a command;
- all filesystem, authentication, network, Firebase, hook, cache, draft,
  registry, and profile effects reachable in JSON mode;
- generated schemas and both capability goldens;
- machine-contract documentation.

### 2.2 Excluded

The audit does not include:

- TUI behavior;
- human-output wording, styling, column sizing, or terminal interaction except
  where it leaks into or contradicts JSON mode;
- undocumented experimental human-only transports, including directory stdin;
- Firebase behavior outside the boundary fbrcm publishes;
- proposals for new commands, fields, conveniences, or guarantees;
- compatibility with an unreleased older draft of the same contract version.

Excluded behavior MUST NOT appear in capability or schema metadata. It MAY be
documented as human-only if that documentation cannot be mistaken for a
machine-contract feature.

### 2.3 No criteria changes during an audit

An audit finding MUST cite a criterion ID from section 5 that existed when the
audit began. A desirable property not required by the frozen standard is a
**proposal**, not a finding, and MUST NOT block acceptance.

Changing this standard requires a separate, deliberate documentation change
with a standard-version increment and rationale. A new criterion does not
retroactively fail an audit already in progress. A subsequent audit may select
the revised standard.

## 3. Required audit inventory

Before judging behavior, the auditor MUST produce these sets from the current
repository:

- `E`: executable Cobra command IDs and argv paths;
- `C`: detailed capability command IDs;
- `I`: command input-schema IDs;
- `R`: command response-schema IDs;
- `D`: commands with registered successful data DTOs or an explicit no-data
  registration;
- `B`: command IDs present in the exhaustive behavior manifest;
- `M`: command sections in `docs/CLI.md` or an explicit shared documentation
  entry.

The inventory passes only when:

```text
E = C = I = R = D = B = M
```

Navigational command groups are not members of `E`. The root operation is.
Shared schemas such as envelope, error, capability, semantic, and stdin schemas
are inventoried separately and MUST have unique IDs. File counts are evidence,
not acceptance by themselves; set equality is REQUIRED.

The inventory report MUST also record:

- contract version and release state;
- every schema file and `$id`;
- duplicate or dangling schema IDs and references;
- capability paths and referenced schema IDs;
- dirty-worktree files present before the audit.

## 4. Command audit record

Every member of `E` MUST have exactly one audit record. Records use this fixed
matrix; no cell may be silently omitted.

| Column | Required content |
| --- | --- |
| Command | Stable ID and argv path. |
| Arguments | Arity, required/repeated state, grammar, normalization, and constraints. |
| Options | Type, default, aliases, repeatability, required state, effectiveness, and dependencies. |
| Selection | Every selector's matching, precedence, composition, default scope, and typed lookup failures, or `N/A`. |
| Stdin | Supported mode and concrete schema, or `N/A`. |
| Success | DTO type and every reachable success/empty/no-op/status variant. |
| Failure | Applicable typed problems, details, retryability, and exit statuses. |
| Warnings | Reachable warning branches and remediation, or `N/A`. |
| Interaction | Trigger, required option or bypass, and absence of prompts/launches, or `N/A`. |
| Effects | Effects, network, destructiveness, conditions, and idempotency. |
| Schemas | Input, response, and shared schema IDs. |
| Documentation | Exact section or shared rule covering the command. |
| Evidence | Runtime and schema tests covering every applicable branch. |
| Verdict | `PASS` or criterion-linked findings. |

`N/A` is an explicit audited result. A blank cell is a failure of audit
completion.

## 5. Finite audit matrix

These 51 criteria are the only criteria that produce findings under standard
`1.0.0`.

### INV — Inventory and identity

- **INV-01**: The sets in section 3 MUST be equal.
- **INV-02**: Command IDs, argv paths, schema IDs, filenames, and schema
  references MUST be unique and mutually consistent.
- **INV-03**: The runtime envelope, lock, generated schemas, capabilities, and
  documentation MUST publish the same contract version and release state.
- **INV-04**: Every executable command MUST return an envelope in JSON mode,
  including root, help, completion, parsing failures, and startup failures.

### ARG — Arguments, options, and normalization

- **ARG-01**: Argument arity, required/repeated state, option type, default,
  alias, repeatability, and required state MUST match Cobra and runtime parsing.
- **ARG-02**: Accepted grammars, enums, formats, bounds, length limits, and
  whitespace rules MUST match runtime validation.
- **ARG-03**: Normalization order and raw-versus-normalized validation MUST be
  published for every value runtime normalizes.
- **ARG-04**: Mutual exclusions, dependencies, conditional requirements, and
  at-least-one requirements MUST reject exactly the combinations rejected by
  pre-execution runtime validation.
- **ARG-05**: Omitted, explicitly false, explicitly empty, null, and repeated
  values MUST retain their distinct runtime meanings.
- **ARG-06**: Machine option effectiveness MUST satisfy definition 1.4.
- **ARG-07**: Validation intentionally deferred to runtime MUST have a typed,
  documented semantic annotation and a structured failure.

### SEL — Selection and composition

- **SEL-01**: Every selector MUST satisfy definition 1.5.
- **SEL-02**: Positional shorthand transformations MUST be published and their
  conflicts with explicit filters MUST be schema-enforced.
- **SEL-03**: Repeated selectors, different selector sources, omitted
  selectors, and target prefixes MUST publish their exact composition.
- **SEL-04**: Zero, one, and multiple candidate results MUST produce the
  documented selection or typed not-found/ambiguous problem.
- **SEL-05**: Selector schemas and shared documentation MUST describe the same
  algorithm used by the specific command; a shared selector MUST NOT be reused
  when command-local behavior differs.

### STDIN — Standard input

- **STDIN-01**: `supports.stdin`, `stdin_modes`, `stdin_schema`, and the input
  schema's `stdin` property MUST agree with runtime JSON-document handling.
- **STDIN-02**: Each published stdin shape MUST cover exactly the locally
  accepted decoder shape and distinguish later Firebase validation.
- **STDIN-03**: Options unavailable or ineffective with stdin MUST be rejected
  or conditionally marked ineffective, matching runtime.
- **STDIN-04**: Unpublished experimental human-only stdin behavior MUST remain
  absent from machine schemas and capabilities.

### OUT — Envelope and success data

- **OUT-01**: JSON stdout MUST contain exactly one schema-valid envelope and one
  trailing newline, with no prompt, table, usage, progress, or raw artifact.
- **OUT-02**: Every successful data DTO MUST be statically registered and its
  actual JSON MUST validate against the advertised response schema.
- **OUT-03**: Required, nullable, omitted, count, item, and artifact fields MUST
  match runtime serialization for empty and populated results.
- **OUT-04**: Status and variant enums MUST satisfy definition 1.6.
- **OUT-05**: Outcome, `data`, `errors`, `warnings`, `command`,
  `requested_command`, schema ID, and process/embedded exit code MUST satisfy
  the documented envelope invariants.
- **OUT-06**: Artifact encoding, media type, destination, overwrite, byte size,
  and digest invariants MUST match the exact returned or written bytes.

### ERR — Problems, warnings, and exit status

- **ERR-01**: Every runtime failure source MUST be classified from a typed error
  rather than message text.
- **ERR-02**: Code, category, details kind, retryability, target, stage,
  remediation strategy, outcome, and exit status MUST agree with the error and
  response schemas.
- **ERR-03**: Every documented problem and warning code MUST be reachable and
  tested, or explicitly reserved under definition 1.6.
- **ERR-04**: Batch all-failed, partial-success, and target-detail aggregation
  MUST preserve typed target failures and documented process status.
- **ERR-05**: Messages and captured output MUST obey redaction and size bounds;
  agents MUST NOT need to parse message wording.
- **ERR-06**: Every warning MUST be non-fatal, schema-valid, and accompanied by
  typed details and safe remediation when declared.

### BEH — Effects and retry safety

- **BEH-01**: Every reachable local, authentication, network, Firebase, hook,
  cache, draft, registry, profile, and artifact effect MUST appear in
  `side_effects` with an exact `side_effect_when` condition or as unconditional.
- **BEH-02**: `network_access` and `network_when` MUST describe every and only
  reachable JSON-mode network path, including authentication access.
- **BEH-03**: Destructive state, reasons, and predicates MUST match actual
  deletion, replacement, rollback, publication, and overwrite behavior.
- **BEH-04**: Idempotency and conditional idempotency MUST describe retry safety
  after every reachable stop point, including partial publication and hooks.
- **BEH-05**: Dry-run, draft, cached, offline, stdin, no-op, and already-applied
  branches MUST declare their actual changes and suppressed effects.
- **BEH-06**: Capability predicates MUST reference existing arguments, options,
  contexts, and defined runtime-state semantics with correct JSON value types.

### INT — Non-interactive operation

- **INT-01**: JSON mode MUST NOT prompt, open a browser, editor, or file picker,
  or consume stdin as an interaction response.
- **INT-02**: Every required human choice MUST return a typed interaction
  problem with the documented trigger and bypass or required option.
- **INT-03**: Interaction metadata and `interaction_when` MUST cover exactly the
  runtime branches that stop for interaction.
- **INT-04**: A preview returned with an interaction problem MUST validate as
  usable partial data under the command response schema.

### DOC — Documentation agreement

- **DOC-01**: `docs/CLI.md`, `docs/cli-contract.md`, schemas, capabilities, and
  runtime MUST agree on every machine-visible command, argument, flag, default,
  constraint, result, effect, interaction, and exit status.
- **DOC-02**: Shared rules MUST list their exceptions. Command-local behavior
  that differs from a shared rule MUST be documented locally and use a distinct
  schema definition.
- **DOC-03**: Human-only or experimental behavior MUST be labeled as such and
  MUST NOT be presented as machine-contract support.
- **DOC-04**: Examples claimed to be valid MUST validate and execute through the
  intended branch; examples claimed invalid MUST be rejected as documented.

### GEN — Generation and regression protection

- **GEN-01**: `go run ./cmd/schemagen` MUST complete successfully from a clean
  tree and a second run MUST produce byte-identical schemas and goldens.
- **GEN-02**: Generated schemas MUST be valid Draft 2020-12 documents, every
  reference MUST resolve, and representative actual envelopes MUST validate.
- **GEN-03**: Contract tests MUST fail when an executable command, response DTO,
  behavior entry, schema, capability record, or documentation inventory entry
  is missing.
- **GEN-04**: Every criterion applicable to a command MUST have automated
  regression evidence as specified in section 6.
- **GEN-05**: Repository-wide tests, vet, and lint MUST pass after generated
  artifacts are current.

## 6. Required evidence and test classes

Source inspection alone is not acceptance evidence where runtime behavior can
be exercised with local fakes. Each command record MUST cite the applicable
test classes below.

| Test class | Required when | Minimum cases |
| --- | --- | --- |
| Discovery | Every command | Compact and detailed capability entries; input and response schema lookup. |
| Invocation | Every command | Minimal valid invocation; too few/many arguments; unknown option; every dependency and exclusion. |
| Boundary | Typed strings/numbers/enums | Minimum, maximum, just outside bounds, empty, whitespace-only, case variants, overflow, and malformed values as applicable. |
| Effectiveness | Every nontrivial option | Omitted and changed value; each `effective_when` true and false branch; every ineffective option. |
| Selection | Every selector | Zero, one, and multiple matches; case/prefix/mode/precedence; repeated and cross-source composition; omitted default. |
| Stdin | Every stdin command | Valid document, malformed JSON, locally invalid shape, null/empty boundary, and every stdin-specific option restriction. |
| Success | Every command | Populated and empty success; every reachable status/variant; schema validation. |
| No-op | Every mutation | No match and already applied, distinguishing selection breadth and changed counts. |
| Failure | Every declared failure branch | Typed code/category/details/retryability/exit status and response-schema validation. |
| Interaction | Every interactive branch | Trigger without bypass, bypass where available, preview data where applicable, and proof no prompt/launch occurred. |
| Warning | Every warning branch | Typed details, remediation argv/strategy, redaction, and successful or partial outcome as applicable. |
| Batch | Every batch command | All success, mixed success/failure, all failure, continuation, ordering, and aggregation. |
| Effects | Every nonzero or conditional behavior | Instrumented proof of each effect and each condition's true/false branch; network suppression for cached/offline/stdin paths. |
| Artifact | Every artifact command | Inline and destination forms, media/encoding variants, overwrite interaction, bytes, size, and SHA-256. |
| Determinism | Entire contract | Two generation runs with no byte differences. |

“As applicable” is determined solely from the command's code paths and declared
capability branches, not auditor preference. A test class is `N/A` only when
the corresponding command-record cell explains why the branch cannot occur.

Tests MAY share fixtures and table-driven cases. The requirement is branch
coverage and schema validation, not one test function per matrix cell.

## 7. Finite acceptance criteria

The contract receives verdict **AUTHORITATIVE — PASS** only when all of these
conditions are true:

1. The audit basis records the repository revision and audit-standard version.
2. Inventory sets `E`, `C`, `I`, `R`, `D`, `B`, and `M` are equal.
3. Every executable command has a complete section-4 record with no blank
   cells.
4. Every applicable criterion `INV-01` through `GEN-05` passes.
5. Every applicable test class in section 6 has passing automated evidence.
6. There are zero unresolved criterion-linked findings.
7. Schema generation succeeds twice and the second run is byte-identical.
8. Generated artifacts and capability goldens match the audited runtime tree.
9. `go test -count=1 ./...` passes.
10. `go vet ./...` passes.
11. `golangci-lint run` reports zero issues.
12. The final workspace diff contains only intended changes and no unexplained
    generated or unrelated modifications.

Failure of any condition produces verdict **NOT AUTHORITATIVE — FAIL** with a
finite list of criterion-linked findings. There is no intermediate “probably
authoritative” verdict.

Once all twelve conditions pass, the auditor MUST call the contract
authoritative under this standard. The auditor MUST NOT withhold acceptance for
an uncodified preference, hypothetical future feature, or newly invented
criterion.

## 8. Finding format and closure

Every finding MUST contain:

```text
Criterion: SEL-05
Commands: draft.show, draft.diff
Evidence: file/line, generated schema path, and failing or missing regression
Observed: exact current runtime/schema behavior
Required: exact requirement already stated by the criterion
Remedy class: schema | capability | documentation | runtime | test | generated
Retest: exact matrix cases that close the finding
```

A statement without a criterion and evidence is not an audit finding. Severity
may prioritize work but does not change pass/fail.

A finding is closed only when:

- the disagreement is corrected;
- its regression evidence is added or updated;
- affected schemas and goldens are regenerated;
- the command record passes on retest; and
- all section-7 repository checks pass.

If fixing a finding would require runtime behavior to change, the audit report
MUST identify that remedy class before implementation. Authorization and
product choice are workflow concerns; they do not change the criterion.

## 9. Gate for future commands and contract changes

Before merging a new executable command or changing a machine-visible command,
the change MUST:

1. add or update the response DTO registration and behavior manifest;
2. add or update capability arguments, options, effectiveness, selectors,
   stdin, effects, interaction, and retry-safety metadata;
3. add or update input and response schema generation;
4. update `docs/CLI.md` and, when a shared rule changes,
   `docs/cli-contract.md`;
5. add the command's complete audit record or update the maintained generated
   audit inventory when that record becomes automated;
6. add every applicable section-6 test class;
7. regenerate and review schemas and capability goldens; and
8. satisfy all section-7 acceptance criteria.

A released contract follows the compatibility and versioning rules in
`docs/cli-contract.md`. The current unreleased v1 contract may be corrected in
place, but it MUST still pass this audit standard before release.

## 10. Canonical audit report header

Every full audit report starts with this block:

```text
Repository revision: <commit or working-tree identifier>
Contract version: <version>
Contract released: <true|false>
Audit standard: docs/cli-contract-audit.md version <version>
Executable commands: <count>
Input schemas: <count>
Response schemas: <count>
Shared schemas: <count>
Pre-existing dirty files: <list or none>
Verdict: AUTHORITATIVE — PASS | NOT AUTHORITATIVE — FAIL
Findings: <count>
```

The report then contains the inventory equality result, the command audit
records, criterion results, test evidence, deterministic-generation result,
and the finite finding list. This report is the only basis for the verdict.

## 11. Audit-standard versioning

This standard has its own version and does not change the CLI contract version.
Its version follows these rules:

- patch: clarification that neither adds a criterion nor changes pass/fail;
- minor: a new criterion, included artifact, test class, or acceptance
  condition, applicable only to audits that start with that revision;
- major: a changed definition of authoritative, removed criterion, or changed
  interpretation of an existing criterion.

Every standard-version change MUST include a rationale in the same commit.
Auditors MUST use the version recorded at audit start; later revisions cannot
alter that audit's verdict.

## 12. Canonical audit procedure

A full audit follows these steps once, in order:

1. Freeze and record the repository revision, dirty files, contract version,
   release state, and audit-standard version.
2. Build sets `E`, `C`, `I`, `R`, `D`, `B`, and `M`; stop and record `INV`
   findings if they differ, but retain the union so missing members remain in
   the report.
3. Produce one section-4 record for every member of `E` and map every cell to
   the applicable section-5 criteria and section-6 evidence.
4. Run static schema, reference, capability, documentation, and source-to-
   metadata checks for all records.
5. Run the applicable runtime test classes with local fakes and validate every
   captured envelope against its advertised schema.
6. Record only criterion-linked findings. Fix them without changing the frozen
   standard; if product authorization is required, identify the runtime remedy
   and pause that finding without creating substitute criteria.
7. Regenerate twice, rerun every affected record, then run the complete test,
   vet, lint, and diff checks.
8. Evaluate the twelve section-7 conditions and issue exactly one of the two
   permitted verdicts.

If the verdict is `NOT AUTHORITATIVE — FAIL`, a later repair audit starts from
the same standard version unless the standard was separately revised. It may
reuse passing records whose runtime, schemas, capabilities, documentation, and
evidence did not change, and MUST retest every changed or dependent record.
It does not reopen passing records to search for requirements outside the
frozen matrix.
