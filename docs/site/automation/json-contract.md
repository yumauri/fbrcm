# JSON contract

The examples on this page use the locked contract `1.1.0`, which adds `explain` and
`versions blame --condition` while preserving parameter blame. The published
`1.0.0` schemas and release lock remain available. See the
[compatibility guarantees](/reference/compatibility) for schema versioning,
upgrade rules, and the exclusion of human layouts from automation interfaces.

With `--json`, each one-shot CLI invocation writes one JSON document and one
trailing newline to stdout. This includes argument and startup failures. Parse
stdout only. Explicitly enabled logs and trusted hooks may write to stderr.

For protocol-based tool calls, use the [MCP server](./mcp#read-tool-results)
without `--json`.

## Envelope

Every response contains the same top-level fields:

```json
{
  "schema": "urn:fbrcm:schema:cli:1.1.0:command:projects.list:response",
  "contract_version": "1.1.0",
  "command": "projects.list",
  "requested_command": "projects.list",
  "outcome": "success",
  "exit_code": 0,
  "producer": { "name": "fbrcm", "version": "1.1.0" },
  "context": {
    "profile": "default",
    "offline": false,
    "dry_run": false,
    "draft": false
  },
  "data": { "count": 0, "items": [] },
  "errors": [],
  "warnings": []
}
```

`data` is the command DTO, an artifact DTO, or `null` when no usable result
exists. Collections use `{ "count", "items" }`; singular resources use an
object. `producer.version` is the installed fbrcm release and therefore differs
from the example as new versions are published.

## Outcomes and exit statuses

| Status | Meaning |
| ---: | --- |
| `0` | Success; no changes for a diff |
| `1` | Differences found, invalid validation report, or failed diagnostics |
| `2` | Invalid arguments, flags, or command path |
| `3` | Configuration or profile failure |
| `4` | Authentication failure |
| `5` | Permission denied |
| `6` | Project or other resource not found |
| `7` | Conflict, including ETag failures |
| `8` | Input or Remote Config validation failure |
| `9` | Deadline exceeded |
| `10` | Explicit interaction required |
| `11` | Network, offline, rate-limit, or service unavailable |
| `12` | Partial batch success |
| `13` | Local file or stream I/O failure |
| `14` | Publication hook failure |
| `15` | Internal or contract-encoding failure |
| `130` | Interrupted or canceled |

Status 1 is not necessarily failure. Always read `outcome` as well as the
process status.

## Structured problems

Each error has stable `code`, `category`, `retryable`, `target`, `stage`,
`details`, and `remediation` fields in addition to its message. Branch on
`code` or `category`, never on message text.

Each remediation includes one of these strategies:

- `retry_with_arguments` augments the original invocation;
- `replace_selector` replaces an ambiguous selector; and
- `run_command` is a complete fbrcm subcommand argument list.

A remediation describes a technically valid recovery. Callers must still
check scope, side effects, and authorization before executing it.

## Parameter explanations

[`explain`](/cli/explain) returns one envelope for one parameter in a published
template. Inspect `data.result.selection` for the winning condition or default.
If selection is unresolved, `data.possible_outcomes` lists the remaining
branches in priority order with their guards, and `data.needed_inputs` lists
inputs that can still affect selection. The value can be known even when the
winning branch is unresolved, such as when every remaining branch has the same
literal value. `data.conditions` contains rule results and references to the
inputs they used.

When a HAR provides response evidence, inspect
`data.capture.response.comparison` separately from the predicted result. Missing
facts, unsupported expressions, and mismatches with a captured value can all
appear in a successful response with exit `0`. Check `warnings` for
`explain.capture_version_mismatch` when the versions differ and `cache.stale`
when the template read fell back to a stale cache.

| Problem code | Exit | Meaning |
| --- | ---: | --- |
| `argument.invalid` | `2` | An invalid override, entry selection, or conflicting capture target |
| `parameter.not_found` | `6` | The exact key is absent from the selected template |
| `capture.invalid` | `8` | The capture could not be parsed or validated |
| `interaction.required` | `10` | Several Firebase fetches are present. Choose an original one-based HAR entry from the returned candidates and retry with `--entry`. |

Use `fbrcm capabilities explain --json` to find the invocation and response
schemas for the installed version. The same result is available through the
stateful [MCP `parameters.explain` tool](/automation/mcp#choose-tools-and-permissions).

## Artifacts

Commands returning Remote Config, defaults, drafts, or exports use an artifact
DTO. It describes media type, encoding, destination, byte size, SHA-256 digest,
and whether an existing destination was overwritten. Content is either inline
or written to the destination, never both.

## Discover schemas

```sh
fbrcm capabilities update --json
fbrcm schema list --json
fbrcm schema show \
  urn:fbrcm:schema:cli:1.1.0:command:update:response \
  --json
```

Schemas use JSON Schema Draft 2020-12 and include fbrcm annotations for
selection, side effects, effective flags, and invariants that JSON Schema
cannot calculate by itself.

## Time limits

Use the global timeout for a complete operation:

```sh
fbrcm projects list --json --timeout 30s
```

It covers target resolution, authentication, network pacing and retries,
validation, hooks, and local persistence. Expiration returns status 9; an
interrupt returns 130.
