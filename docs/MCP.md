# Built-in MCP server

`fbrcm mcp` exposes Remote Config workflows over stdio JSON-RPC using the official
Go MCP SDK v1.7.0. It supports the 2026-07-28 multi-round-trip interaction flow;
the SDK adapts interactions for older supported protocol versions. Browser
recovery requires URL elicitation (introduced in MCP 2025-11-25); mutation
confirmation requires form elicitation. Host support and UI vary.

## Setup and permissions

Configure your AI application with the binary path, arguments, and environment.
For hosts accepting the common `mcpServers` configuration shape:

```json
{
  "mcpServers": {
    "firebase": {
      "command": "/absolute/path/to/fbrcm",
      "args": ["mcp", "--profile", "work", "--no-local-config"]
    }
  }
}
```

The profile must already exist. The host launches the subprocess when connecting;
users do not normally launch it manually before chatting. Discovery does not
require valid Google credentials. Authentication happens when a tool needs
Firebase; existing service-account and gcloud methods remain available.

Add `--allow-writes` to enable mutations and explicit artifact writes. Default
`--confirmation=host` asks for approval of the exact tool name and inputs before
execution. Models cannot pass `yes` or change startup policy. Hosts without form
elicitation receive a typed interaction-required error. Operators can explicitly
choose `--confirmation=none` for unattended execution; this grants authority to
run enabled mutations without per-call user confirmation.

Normal local configuration and aliases apply unless `--no-local-config` is set.
Hooks are disabled unless `--allow-hooks` is present; existing trust checks are
not bypassed. No tools manage hook trust, profiles, credentials, configuration,
themes, completion, editors, or the TUI. File paths resolve on the fbrcm host
relative to its working directory. Artifact writes use the host's filesystem
permissions; MCP roots are not a filesystem sandbox.

## Tool catalog

Default groups: `inspect,edit,drafts,plans,publish`. Select a subset with
`--toolsets`. Selecting a mutation group does not itself enable writes.

| Group | Tool IDs |
| --- | --- |
| `inspect` | `get`; `projects.list`, `projects.diff`; `project.show`, `project.defaults`; `groups.list`; `conditions.list`, `conditions.show`, `conditions.validate`; `versions.list`, `versions.show`, `versions.diff`; `experiments.list`, `experiments.show`; `rollouts.list`, `rollouts.show`; `personalizations.list`, `personalizations.show` |
| `edit` | `add`, `update`, `delete`, `duplicate`; `groups.add`, `groups.edit`, `groups.rename`, `groups.delete`; `conditions.add`, `conditions.edit`, `conditions.rename`, `conditions.move`, `conditions.delete` |
| `drafts` | `draft.list`, `draft.show`, `draft.diff`, `draft.change-note`, `draft.discard` |
| `plans` | `plan.show`, `plan.validate` |
| `publish` | `apply`, `draft.publish`, `project.import`, `project.export`, `projects.promote`, `versions.export`, `versions.rollback`, `versions.restore`, `experiments.delete`, `rollouts.delete` |
| `diagnostics` | `doctor` (opt-in) |

Without `--allow-writes`, mutation tools and explicit `to`/`plan-out` options
are unavailable. Inspection is not filesystem-pure: it can refresh caches,
synchronize the registry, and persist credentials. Doctor can create and remove
diagnostic probe files. Stateless mode disables application-managed local state.

Each tool publishes semantic input and output schemas. Example `get` arguments:

```json
{
  "arguments": {},
  "options": {"project": ["=acme-staging"]},
  "stdin": null
}
```

Argument names come from the schema. Options use long flag names without `--`;
repeatable values are arrays, booleans are JSON booleans. Supplied `stdin` is a
JSON document rather than a JSON-encoded string. Inputs are limited to 16 MiB.
Selection, filtering, expressions, validation, dry runs, plans, drafts, and
non-atomic batch behavior remain consistent with CLI execution.

Completed results contain the unchanged CLI envelope as `structuredContent`
and identical JSON in text content. Inspect `outcome`, `errors`, `warnings`, and
`data`, not just `exit_code`: changed diffs succeed with exit `1`. Partial success
is an error result retaining completed/failed target data.

## Authentication recovery

Expired access tokens normally refresh silently. When Google requires a new
browser authorization and the host supports URL elicitation:

1. The existing fbrcm process generates OAuth state/PKCE values and starts a
   temporary HTTP listener on `127.0.0.1` with an available port.
2. It builds Google's URL with that callback address and asks the host to offer
   sign-in. The listener starts **before** the URL is presented. Browser navigation
   requires user consent through the host; fbrcm does not open it automatically.
3. After login, the browser redirects to localhost. fbrcm validates state and
   exchanges the authorization code directly with Google using the PKCE verifier.
4. Credentials are persisted, client initialization/recovery completes, the
   listener closes, and the same suspended operation continues.

The callback listener is not an HTTP MCP endpoint or another executable. Tokens,
codes, and PKCE secrets never pass through MCP. Accepting navigation does not
prove successful authentication, and signing in does not authorize a mutation.

Pending operations survive input-required round trips. Opaque continuation handles
are bound to the exact input and connection; they are not authorization codes.
Completed results remain available for one minute so repeating a continuation
does not repeat the operation. Expired handles are rejected. At most 64 pending
or recently completed operations are retained. Domain execution is serialized;
discovery remains responsive, and queued operations reuse newly persisted tokens
instead of starting competing sign-ins.

Cancel, timeout, failed login, or disconnect cleans up the attempt. Ordinary
tool failures leave MCP running. Mutations are not replayed as a whole after
authentication: existing conflict checks and partial-publication reporting remain.

With `--browser-auth=never` or no host URL support, follow the typed remediation:
run `fbrcm auth login <auth-id> --profile <profile>` externally, then retry. New
operations reload credentials, so reconnecting is not normally needed. Browser
callbacks require access to the fbrcm machine's localhost; SSH, containers, or
remote machines may require external login instead.

## Stateless mode

Use `["mcp", "--stateless"]` and securely provision `FBRCM_GOOGLE_ACCESS_TOKEN`
through the host environment. Do not paste tokens into chat or commit them.

Stateless execution accesses no profiles, application config, caches, drafts,
or hooks. Only compatible tools are listed; explicit artifact I/O follows normal
command rules. Remote mutations remain possible when launch and confirmation
policies permit them. Plan inspection currently remains stateful according to
its CLI capability metadata. This policy is independent of MCP's protocol-level
meaning of stateless.

Expired environment tokens must be renewed externally. Restart the subprocess
with the new environment; no fallback to profile or browser authentication occurs.

## Timeouts and lifecycle

`--timeout` limits the entire server lifetime; omit it for normal host management.
`--request-timeout` defaults to `5m` per operation, including queueing and user
interaction. `--auth-timeout` defaults to `2m` per browser attempt. Explicit
durations must be positive; the earliest deadline wins. Disconnect/interrupt
cancels work and closes temporary listeners. Stdout is reserved for MCP messages;
logs use stderr. `--json` is incompatible with the streaming command.
