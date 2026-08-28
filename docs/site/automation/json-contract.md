# JSON contract

With `--json`, every invocation writes one JSON document and one trailing
newline to stdout. This includes argument and startup failures. Parse stdout
only. Explicitly enabled logs and trusted hooks may write to stderr.

## Envelope

Every response contains the same top-level fields:

```json
{
  "schema": "urn:fbrcm:schema:cli:1.0.0:command:projects.list:response",
  "contract_version": "1.0.0",
  "command": "projects.list",
  "requested_command": "projects.list",
  "outcome": "success",
  "exit_code": 0,
  "producer": { "name": "fbrcm", "version": "0.20.0" },
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
  urn:fbrcm:schema:cli:1.0.0:command:update:response \
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
