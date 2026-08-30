# CLI JSON contract audit — built-in Google OAuth

Repository revision: `5331bd60af2a5217c62071b28cb8ea0ec0a98594` plus the audited repair worktree
Contract version: `1.0.0`
Contract released: `false`
Audit standard: `docs/cli-contract-audit.md` version `1.0.0`
Executable commands: `109`
Input schemas: `109`
Response schemas: `109`
Shared schemas: `10`
Pre-existing dirty files: none
Verdict: **AUTHORITATIVE: PASS**
Findings: `0` unresolved (`5` found and closed)

Availability note: the command contract is implemented, but the built-in Google
sign-in method remains unavailable until Google completes verification of
fbrcm's OAuth application.

The audit scope was frozen at branch revision `5331bd60af2a5217c62071b28cb8ea0ec0a98594`
before repair. The generated contract lock is unreleased and has SHA-256
`3024c7a03609f20b0d0cf04de354c8b55dc33e95c9064a5af9ce9c68cabadcb9`.

## Inventory

Exact set comparison: `E = C = I = R = D = B = M`, with 109 members in every
set and zero symmetric-difference members. The new member is
`auth.add.google`; every other member is listed in the 2026-08-28 full audit.
The executable tree, DTO registry, behavior manifest, detailed and compact
capabilities, documentation inventory, input schemas, and response schemas are
compared by the repository contract tests and generated goldens.

There are 228 unique Draft 2020-12 schema files under `schemas/cli/1.0.0`:
109 input schemas, 109 response schemas, and the same 10 shared schema IDs
listed in the [2026-08-28 audit](cli-json-contract-2026-08-28.md). All IDs and
references compile and resolve. Capability paths reference the matching input,
response, and error schema IDs; no duplicate or dangling IDs were found.

## Command audit records

The 101 unchanged records from the
[2026-08-28 full audit](cli-json-contract-2026-08-28.md) are incorporated by
reference. Their runtime, schemas, capabilities, documentation, and evidence
were revalidated by the complete inventory, schema, root-module, and E2E gates.
The seven affected existing records are amended below, followed by the one new
record. Together these are the complete 109-command record set.

### Affected existing records

| Command | Complete audited amendment |
| --- | --- |
| `auth.add.gcloud` | Arguments and selection remain a literal required `auth_id`; stdin and interaction remain N/A. `--quota-project` now explicitly publishes Unicode trim normalization followed by the physical-project-ID grammar, and runtime reports invalid values as typed `argument.invalid`. Success remains the registered `gcloud` `authMutationResult` with status `added`; warnings remain N/A. Effects remain conditional `local_state_write` and `local_file_delete`, network `none`, destructive replacement, idempotency `no`. Input/response IDs remain command-specific. Documentation is `docs/CLI.md § fbrcm auth add gcloud` plus the shared quota rule. Schema, boundary, invocation, failure, effect, and generation tests pass. Verdict: **PASS**. |
| `auth.add.oauth` | The prior record remains complete for arguments, imported-credential stdin, success, failure, warnings, interaction, effects, and schemas. Its `--quota-project` option now publishes the same trim plus physical-ID rule and typed invalid-argument runtime classification. Imported-credential stdin still takes precedence exactly as documented. Updated input-schema and runtime boundary tests pass. Verdict: **PASS**. |
| `auth.add.service-account` | The prior record remains complete for arguments, service-account stdin, success, failure, warnings, interaction, effects, and schemas. Its `--quota-project` option now publishes the same trim plus physical-ID rule and typed invalid-argument runtime classification. Updated input-schema and runtime boundary tests pass. Verdict: **PASS**. |
| `auth.list` | Arguments/options/selection/stdin/interaction/effects remain unchanged and applicable N/A cells remain explicit in the prior record. Success type now includes the reachable `google` item shape: a required nonempty token path with no client-secret or service-account path. `auth_list_google_json` validates populated Google output and the existing empty/list scenarios remain valid. Schemas and docs agree. Verdict: **PASS**. |
| `auth.path` | Literal auth-ID selection, failures, warnings, interaction, and effects remain as previously recorded. Success now includes the `google` path shape with exactly the managed token path and no imported client-secret/key path. `auth_path_google_json` validates the runtime envelope. Verdict: **PASS**. |
| `auth.login` | The prior arguments/options/network/authentication-effect record remains complete. Success now includes the `google` path/type variant, covered by `auth_login_google_json`. Missing reusable authorization in JSON mode returns typed `interaction.required` before offline network suppression, listener creation, or browser launch; `auth_login_google_interaction_json` covers that branch. Service-account and gcloud behavior remains unchanged. Verdict: **PASS**. |
| `auth.delete` | The prior literal selection, confirmation interaction, destructive effects, failures, and warnings remain complete. Success adds `google` and now correlates `type` with `deleted_paths`: Google and service-account return one nonempty path, imported OAuth returns two unique paths, and gcloud returns `null`. All four types have E2E envelopes; impossible cross-type path counts are schema-rejected. Verdict: **PASS**. |

### `auth.add.google`

| Column | Audited content |
| --- | --- |
| Command | Stable ID `auth.add.google`; argv `fbrcm auth add google`. |
| Arguments | One required, non-repeated literal `auth_id`; exact arity; `path_segment` grammar; no normalization or lookup fallback. |
| Options | `--label` string default `""`; `--quota-project` string default `""`, effective when supplied, Unicode-trimmed then validated as a literal physical project ID; inherited `--profile`, `--no-local-config`, `--stateless=false`, and positive `--timeout` semantics match Cobra and the input schema. |
| Selection | N/A. `auth_id` names the exact registry key to upsert; it has no fuzzy, prefix, ambiguity, or remote lookup behavior. |
| Stdin | N/A. `supports.stdin=false`, `stdin_schema=null`, no modes, and normalized `stdin` is null-only. |
| Success | Registered `authMutationResult`; outcome `success`; status `added`; type const `google`; paths require auth/profile config paths and one token path while forbidding client-secret and service-account paths. New-identity and imported-OAuth replacement branches both validate. |
| Failure | Closed codes: `argument.invalid`, `auth.configuration_invalid`, `command.canceled`, `command.timeout`, `configuration.invalid`, `file.io_failed`, `filesystem.permission_denied`, `internal.contract_violation`, `internal.unclassified`, and `profile.invalid`. Missing build credentials are a typed auth configuration failure with exit 4 before persistence; invalid quota input is an argument failure with exit 2. |
| Warnings | N/A; response schema requires an empty array. |
| Interaction | N/A; JSON behavior `non_interactive`; the add operation neither starts OAuth authorization nor opens a browser. |
| Effects | Level 1; conditional `local_state_write` and `local_file_delete`; network `none`; destructive only when replacing an identity whose managed credential/token files are removed; idempotency `no` because a stopped replacement can have partially persisted local effects. |
| Schemas | `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:input`; `urn:fbrcm:schema:cli:1.0.0:command:auth.add.google:response`; shared envelope and error schemas. |
| Documentation | `docs/CLI.md § fbrcm auth add google`; quota normalization, auth effects, OAuth interaction, and response rules in `docs/cli-contract.md`. |
| Evidence | Exact 15-class record in `contract_v1_audit_evidence.golden.json`; discovery/arity/unknown-option/schema suites; command-specific quota and missing-build-client failures; `auth_add_google_json`; `auth_add_google_replace_oauth_json`; deterministic generator test. Artifact, batch, selection, stdin, interaction, and warning classes are explicitly N/A. |
| Verdict | **PASS.** |

## Criterion results

| Criteria | Result |
| --- | --- |
| `INV-01`–`INV-04` | PASS: exact 109-member inventory equality; unique identities and references; version/release agreement; all executable and failure-stage JSON envelopes covered. |
| `ARG-01`–`ARG-07` | PASS: Cobra arity/options agree; Google auth ID and quota grammars, normalization, empty/omitted distinctions, effectiveness, and typed runtime boundaries are published and tested. |
| `SEL-01`–`SEL-05` | PASS: unchanged selectors revalidated; the new command's exact upsert key is correctly recorded as literal rather than lookup selection. |
| `STDIN-01`–`STDIN-04` | PASS: Google add is null-only/no-stdin; imported OAuth and service-account stdin contracts remain distinct and conforming. |
| `OUT-01`–`OUT-06` | PASS: new success/failure envelopes are single schema-valid documents; Google path variants and auth-delete type/path correlation are closed and runtime-covered. |
| `ERR-01`–`ERR-06` | PASS: invalid quota and missing build credentials are typed; unreachable Google-add problem codes were removed; no new warnings/batches apply. |
| `BEH-01`–`BEH-06` | PASS: registry writes and both credential-deletion branches match capability predicates; add has no network path; authenticated commands retain accurate OAuth network/token effects. |
| `INT-01`–`INT-04` | PASS: add is non-interactive; JSON login prioritizes the required-human-choice result and performs no listener/browser launch. |
| `DOC-01`–`DOC-04` | PASS: command tree, detailed sections, shared rules, examples, capabilities, and schemas agree. |
| `GEN-01`–`GEN-05` | PASS: two generation runs were byte-identical; 228 schemas compile; inventory/evidence regressions pass; tests, vet, and lint pass. |

## Closed findings

1. **ERR-03 — `auth.add.google` unreachable problem codes.** Evidence was the
   detailed capability and response schema advertising credential parsing,
   auth-ID source classification, and resource-conflict branches absent from
   runtime. Remedy: capability/schema generation. Retest: closed problem-code
   set plus the missing-build-client envelope. **Closed.**
2. **ARG-02, ARG-03, ERR-02 — quota option semantics.** Evidence was runtime
   trim/physical-ID validation without matching normalized schema annotations,
   with invalid values able to fall through as unclassified command errors.
   Remedy: shared auth-add runtime validation, input schemas, docs, and tests.
   Retest: valid, trimmed, prefixed-invalid, empty, and omitted branches.
   **Closed.**
3. **BEH-01, OUT-04, GEN-04 — missing Google effect and response evidence.**
   Evidence lacked a credential-removing replacement and runtime envelopes for
   Google list/path/login/delete variants. Remedy: E2E fixtures/scenarios and
   generated evidence matrix. Retest: add-new, replace-imported-OAuth, list,
   path, cached login, and delete. **Closed.**
4. **INT-02, DOC-01 — offline interaction precedence.** Evidence showed JSON
   login without a token returning `network.offline` before the documented
   required OAuth choice. Remedy: runtime ordering plus unit and E2E coverage.
   Retest: offline/non-interactive missing-token branch. **Closed.**
5. **OUT-03, OUT-04 — auth-delete path variants underconstrained.** Evidence
   showed the response schema accepting impossible path counts for each auth
   type. Remedy: type-correlated response constraints, docs, impossible-state
   test, and all-type E2E envelopes. **Closed.**

## Determinism and repository gates

| Check | Result |
| --- | --- |
| `go run ./cmd/schemagen` twice | PASS; aggregate generated-artifact SHA-256 remained `41cddbb314332a3f87a3dc22ab68ff1e9cfae212425e723cce8ab65f87e865b6`. |
| Schema compilation/reference tests | PASS; 228 unique Draft 2020-12 files. |
| `go test -count=1 ./...` (root) | PASS. |
| `go test -count=1 ./...` (`e2e`) | PASS; 166 scenarios plus harness tests. |
| `go vet ./...` (root and `e2e`) | PASS. |
| `golangci-lint run` (root and `e2e`) | PASS; zero issues. |
| Final diff | PASS; only intended runtime, contract, generated artifacts, docs, fixtures, and audit evidence/report changes. |

All twelve section-7 acceptance conditions pass. The required verdict is
**AUTHORITATIVE: PASS**.
