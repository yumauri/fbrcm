# CLI JSON Schemas

`cli/1.0.0/` is generated and checked in. It contains Draft 2020-12 input and
response schemas for every executable CLI command plus the shared envelope,
problem, capability, semantic-definition, Remote Config stdin, combined
credential stdin, OAuth credential stdin, and service-account credential stdin
schemas.

Each command response schema includes the exact registered success-data DTO.
It references the published envelope schema, which contains the problem
definition, and the semantic schema for shared enums and grammars. Top-level
slices are represented by the envelope as `{count, items}`, and the schema
describes both fields and the
full item DTO. The only intentionally arbitrary response document is `schema
show`, because its data is the selected schema itself.

Register the envelope and semantic documents by their `$id` URNs with a JSON
Schema validator before compiling a command schema.

Input schemas describe normalized `arguments`, supplied `options`, and `stdin`.
They include case-insensitive accepted values where runtime parsing is
case-insensitive, bounds, duration patterns, mutual exclusions, semantic and
conditional requirements, selector/filter/expression grammars, and concrete
stdin shapes. The semantic schema defines and validates the normative
`x-fbrcm-validation`, `x-fbrcm-normalization`, `x-fbrcm-matching`, and
`x-fbrcm-invariants` rule languages used for behavior that standard Draft
2020-12 keywords cannot express. The capability schema likewise embeds exact
runtime-state predicate and side-effect semantics.

The shared problem schema requires each remediation to declare a `strategy`
(`retry_with_arguments`, `replace_selector`, or `run_command`) and a non-empty
`argv`, so consumers never have to infer whether the vector is a full command
or an edit to the original invocation. The envelope schema constrains success,
partial success, and failure to their valid combinations of exit status, data,
and errors, including the required mapping from the first problem category to
the process status. Problem `details` objects are discriminated by `kind`, and
the code property publishes the current catalog under
`x-fbrcm-known-values` while remaining forward-compatible with new codes.

Artifact schemas enforce the relationship between encoding, inline content,
destination, and overwrite state. Remote Config stdin schemas enforce the
exactly-one-option value union. Invocation schemas include the root version
option, multi-segment help paths, condition colors and move priorities, and do
not accept positive duration strings that Go rounds down to zero.

After changing a command path, argument, flag, side effect, interaction rule,
DTO, error, or exit behavior, regenerate from the repository root:

```sh
go run ./cmd/schemagen
```

Review both the schemas and
`cli/app/testdata/contract_v1_capabilities.golden.json`. The
`cli/contract.lock.json` fingerprint makes generation fail when the machine
surface changes without a contract-version bump. Do not edit generated schema
files or the lock by hand.
