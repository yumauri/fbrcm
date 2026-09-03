# CLI JSON contract audit — MCP launch boundary

Repository revision: `8d0bb2f28c950776c3d123e43ce6c2f870279c12` plus the audited repair worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `113`
Input schemas: `113`
Response schemas: `113`
Shared schemas: `12`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0` unresolved (`2` found and closed)

The audit scope was frozen on branch `mcp` at revision
`8d0bb2f28c950776c3d123e43ce6c2f870279c12`. The generated contract lock is
unreleased and has SHA-256
`da3e5605463c5034d317514f8771521c2e3cde7495630c77160e648f27720557`.
The initial audit pass produced `NOT AUTHORITATIVE: FAIL` for the new `mcp`
command's one-shot JSON descriptor. Both criterion-linked findings were repaired
without changing audit standard 1.0.0. This report records the complete repair
audit and final verdict.

## Inventory

Exact set comparison: `E = C = I = R = D = B = M`, with 113 members in every
set and zero symmetric-difference members. The sole command added after the
September 1 plan audit is `mcp`. Repository tests compare the executable Cobra
tree, response registration, behavior manifest, compact and detailed
capabilities, documentation inventory, generated schemas, and audit-evidence
matrix by exact command ID.

There are 238 unique Draft 2020-12 files under `schemas/cli/1.0.0`: 113 input
schemas, 113 response schemas, and these 12 shared schemas: `capability`,
`envelope`, `error`, `publication_plan`, `semantic`, `stdin:credentials`,
`stdin:oauth_credentials`, `stdin:publication_plan`, `stdin:remote_config`,
`stdin:remote_config_import`, `stdin:service_account_credentials`, and
`stdin:theme`. Every `$id` is unique, all references compile and resolve, and
each capability points to the matching command input and response IDs.

The current compact capability snapshot and E2E discovery output both report
113 commands. `schemas/capabilities.json`, the detailed capability golden, the
standalone capability schema, the audit matrix, and the embedded schema
registry are generated from the same current command tree.

## Command audit records

The 112 records in the
[September 1 full repair audit](cli-json-contract-2026-09-01.md), including the
earlier records it incorporates, are incorporated here by reference. This is
narrative compression, not a skipped retest: all 112 were rebuilt through the
current shared-operation adapters and revalidated by the 113-command inventory,
schema, capability, response-registration, arity, effectiveness, selector,
stdin, outcome, problem-code, interaction, effect, audit-evidence, root-module,
and E2E gates. Their machine-visible arguments, options, selection, stdin,
results, problems, warnings, interaction, effects, schema IDs, and documented
behavior are unchanged by the package refactor.

The one new record follows. Together with the incorporated records it forms
exactly one complete section-4 record for every member of `E`.

### `mcp` — `fbrcm mcp`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `mcp`; argv path `mcp`; executable Cobra leaf with `cobra.NoArgs`. In ordinary mode it launches a streaming frontend, but the CLI JSON operation is only the early transport-incompatibility failure described here. |
| Arguments | No positional arguments. Too many arguments and unknown options return typed `argument.invalid`; no launch or profile bootstrap occurs. |
| Options | Accepted non-JSON launch flags are `--allow-hooks` (bool, default false), `--allow-writes` (bool, false), `--auth-timeout` (duration, `2m`), `--browser-auth` (string, `auto`), `--confirmation` (string, `host`), `--no-local-config` (bool, false), `--profile` (string, environment-derived runtime default represented as empty), `--request-timeout` (duration, `5m`), `--stateless` (bool, false), global `--timeout` (duration, `0s`), and repeatable `--toolsets` (string slice, default `inspect,edit,drafts,plans,publish`). None is required and none has an alias. Every flag is accepted but marked `effective: false` for CLI JSON execution. Basic Cobra type and shared nonblank parsing remain represented; server-only enum, positivity, nonempty-toolset, profile/stateless, and hooks/stateless rules are deliberately not applied after `--json` selects the early rejection. Omitted, false, empty list, repeated, zero, and negative parsed values retain their distinct normalized forms while producing the same rejection. |
| Selection | N/A. The JSON branch resolves no profile, project, toolset, tool, or other resource. |
| Stdin | N/A. `supports.stdin` is false, `stdin_schema` is null, and normalized `stdin` must be null. The non-JSON protocol stream is outside the one-shot CLI JSON contract and is never consumed on this branch. |
| Success | N/A. The command registers explicit no-success data; its response schema makes success and partial success unsatisfiable and requires `data: null`. `mcp --help --json` is the separate generated help operation and uses the help response schema. |
| Failure | Closed set: `argument.invalid`, category `argument`, retryable false, exit status 2. The ordinary `mcp --json` error reports transport incompatibility. Cobra parsing, arity, unknown-option, and shared nonblank failures remain the same typed code. The rejected branch cannot reach configuration, profile, filesystem, timeout, cancellation, server-startup, or unclassified failures. |
| Warnings | N/A. The response schema requires an empty warnings array. |
| Interaction | N/A. No prompt, browser, editor, file picker, MCP elicitation, or interaction response is reachable. `interaction.mode` is `none`. |
| Effects | Level 0; no local, authentication, network, Firebase, hook, cache, profile, or protocol effect; `network_access: none`; non-destructive; idempotent. `supports.stateless` is false because the one-shot JSON operation never enters stateless server mode. |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:mcp:input`, `urn:fbrcm:schema:cli:1.0.0:command:mcp:response`, shared `error`, `envelope`, `capability`, and `semantic` schemas. |
| Documentation | `docs/CLI.md` command tree and “MCP server”; `docs/cli-contract.md` “MCP launch exception”. Streaming setup is separately documented in `docs/MCP.md` and is not presented as a CLI JSON success branch. |
| Evidence | `cli/app/mcp_test.go#TestMCPJSONFailureDoesNotBootstrapProfile`; `cli/app/mcp_test.go#TestMCPJSONContractDescribesOnlyEarlyTransportRejection`; shared discovery, arity, unknown-option, response, problem-code, effectiveness, and schema tests in `cli/app/contract_test.go`; E2E scenario `mcp_json`; exact per-class matrix record in `cli/app/testdata/contract_v1_audit_evidence.golden.json`. The non-JSON `mcp_stateless_eof` scenario is explicitly excluded from CLI JSON stateless and audit-evidence coverage. |
| Verdict | **PASS.** |

## Criterion results

| Criteria | Result |
| --- | --- |
| `INV-01`–`INV-04` | PASS: exact 113-member inventory equality; unique paths, IDs, and references; lock, runtime, schemas, capabilities, and documentation agree on unreleased 1.0.0; root, help, completion, parsing, startup, and MCP rejection envelopes retain coverage. |
| `ARG-01`–`ARG-07` | PASS: all Cobra arities, flag types/defaults/repeatability, grammars, normalization, dependencies, exclusions, omitted/empty/repeated distinctions, effectiveness, and runtime validation boundaries agree. MCP launch options are now explicitly ineffective and its input schema models only parsing before the early JSON rejection. |
| `SEL-01`–`SEL-05` | PASS: all 112 existing selector records were revalidated; the MCP schema no longer imports a spurious stateless target selector and explicitly records selection as N/A. |
| `STDIN-01`–`STDIN-04` | PASS: every published stdin command retains exact mode/schema/runtime agreement; MCP normalized stdin is null and its non-JSON protocol stream is excluded. |
| `OUT-01`–`OUT-06` | PASS: one clean schema-valid envelope and trailing newline; static DTO/no-data registration; empty, nullable, status, count, item, and artifact invariants; MCP failure data is exactly null. |
| `ERR-01`–`ERR-06` | PASS: typed problems and warnings, details, retryability, remediation, aggregation, redaction, outcomes, and statuses remain closed and tested. MCP now advertises only its reachable `argument.invalid` code. |
| `BEH-01`–`BEH-06` | PASS: effects, network access, destructiveness, idempotency, suppression branches, and predicate types match runtime. MCP JSON rejects before every server effect and declares level 0. |
| `INT-01`–`INT-04` | PASS: ordinary JSON execution never prompts or launches external interaction; every declared interaction and preview branch remains tested; MCP JSON cannot enter host elicitation or browser authorization. |
| `DOC-01`–`DOC-04` | PASS: CLI reference, shared contract, examples, schemas, capabilities, and runtime agree; streaming MCP behavior is clearly separated from the rejected one-shot operation. |
| `GEN-01`–`GEN-05` | PASS: two final generations were byte-identical; all 238 schemas compile and resolve; exhaustive inventories and the generated 113-by-15 evidence matrix pass; root/E2E tests, vet, and lint pass. |

## Test-class evidence

`cli/app/testdata/contract_v1_audit_evidence.golden.json` contains 113 command
records and all 15 fixed class cells per command. Every applicable cell cites a
checked test symbol and relevant JSON E2E scenarios; every N/A cell has a
nonblank reason. The catalog's file/symbol references and E2E command IDs are
validated automatically.

| Class | Passing evidence |
| --- | --- |
| Discovery | Exact executable/capability/schema/DTO/behavior/docs equality, reference compilation, compact and detailed goldens, embedded capabilities, and E2E discovery snapshot. |
| Invocation | Cobra arity, unknown option/command, normalized wrapper shape, every shared dependency/exclusion, and the MCP early-rejection option matrix. |
| Boundary | Shared enum, numeric, duration, whitespace, malformed, overflow, stdin, selector, plan, and MCP parsed-versus-server-only boundaries. |
| Effectiveness | Detailed behavior golden, machine-ignored option checks, conditional predicates, and all MCP flags asserted ineffective in capability and input schema. |
| Selection | Existing zero/one/multiple, mode, precedence, repetition, composition, and default tests; MCP is justified N/A. |
| Stdin | Remote Config, credentials, theme, and publication-plan valid/malformed/boundary/restriction suites; MCP is justified N/A. |
| Success | Populated, empty, status, summary, completion, root/help, and adapter-envelope coverage for every success-capable command; MCP is justified no-success. |
| No-op | No-match, already-applied, unchanged, and zero-change mutation coverage. |
| Failure | Closed command problem sets, typed classifications/details/statuses, actual schema-valid failures, unknown options, and MCP's sole reachable code. |
| Interaction | Confirmation, OAuth, external input, destination conflict, selection, preview, and suppression branches; MCP JSON is justified N/A. |
| Warning | Reachable warning codes, details, remediation, redaction, successful/partial outcomes, and schema validation. |
| Batch | All-success, mixed, all-failed, ordering, continuation, and typed aggregation evidence. |
| Effects | Instrumented remote/local/cache/draft/hook/profile/artifact effect and suppression branches; MCP JSON is justified N/A. |
| Artifact | Inline/destination media and encoding variants, overwrite interaction, exact bytes, size, and SHA-256. |
| Determinism | Generator determinism unit test plus the two final repository generation passes below. |

## Closed findings

### Finding 1

Criterion: `ARG-06`, `BEH-05`, `DOC-01`

Commands: `mcp`

Evidence: initial `schemas/capabilities.json` and
`cli/app/testdata/contract_v1_capabilities_detailed.golden.json` `mcp` records;
`mcp/command.go` early `contract.Enabled(cmd)` branch; missing command-specific
effectiveness regression.

Observed: all 11 server-launch flags were marked effective and
`supports.stateless` was true even though `mcp --json` returned before applying
any launch option or entering stateless mode.

Required: accepted but unapplied JSON options must be marked ineffective, and
support metadata must describe the one-shot operation rather than an excluded
transport mode.

Remedy class: capability, schema, documentation, test, generated.

Retest: inspect every flag's `effective` and `effective_when`; validate compact
and detailed capabilities; execute JSON with stateful/stateless and deliberately
server-invalid launch values; prove null profile, zero effects, and the same
typed rejection. **Closed.**

### Finding 2

Criterion: `ARG-02`, `ARG-04`, `ARG-05`, `ERR-03`, `DOC-01`, `GEN-04`

Commands: `mcp`

Evidence: initial `schemas/cli/1.0.0/mcp.input.schema.json` and
`mcp.response.schema.json`; `mcp/command.go`; `mcp/server/catalog.go`; initial
audit-evidence matrix.

Observed: the one-shot input schema enforced positive server timeouts, launch
enums, nonempty toolsets, and stateless launch exclusions that runtime never
reached after `--json`; it imported a nonexistent selector and rejected an
explicit empty toolset list. The response and capability advertised
configuration, profile, path, timeout, cancellation, contract, and unclassified
problems that the early failure could not emit. Evidence also counted a
non-JSON stdio scenario as CLI JSON coverage.

Required: schemas must accept and distinguish exactly the parsed normalized
values reaching the rejection, closed problem sets may contain only reachable
codes, N/A selector/stdin/effect cells must be explicit, and evidence must stay
inside the frozen JSON scope.

Remedy class: schema, capability, documentation, test, generated.

Retest: arbitrary nonblank launch enums, zero/negative parsed durations, empty
and repeated toolsets, profile plus stateless, hooks plus stateless, blank and
malformed parse failures, exact `argument.invalid` response set, selector N/A,
JSON-only E2E evidence, schema compilation, and generated goldens. **Closed.**

## Determinism and repository gates

| Check | Result |
| --- | --- |
| `go run ./cmd/schemagen` twice | PASS; aggregate generated-artifact SHA-256 was `9f0c6badd435e02d51b6204eb3b6786c3c446c59753dae4e93027ce6f3b7f41d` after both runs. |
| Schema compilation/reference tests | PASS; 238 unique Draft 2020-12 files and no duplicate or unresolved IDs/references. |
| Generated audit matrix | PASS; 113 commands, 15 class cells each, no blank cells, missing symbols, mismatched E2E command IDs, or out-of-scope streaming MCP scenario. |
| `go test -count=1 ./...` (root) | PASS. |
| `go test -count=1 ./...` (`e2e`) | PASS. An earlier concurrently loaded baseline run had one Hoverfly fixture timeout; the isolated final gate and its harness tests passed. |
| `go vet ./...` (root and `e2e`) | PASS. |
| `golangci-lint run` (root and `e2e`) | PASS; zero issues. |
| Final diff | PASS; only the intended MCP JSON metadata/generator/runtime classification, tests, generated schemas/goldens/lock, documentation, E2E snapshot/coverage distinction, and this audit report changed. No pre-existing user changes were present. |

All twelve section-7 acceptance conditions pass, every applicable criterion and
test class passes, and there are zero unresolved findings. Under audit standard
1.0.0, the required verdict is **AUTHORITATIVE: PASS**.
