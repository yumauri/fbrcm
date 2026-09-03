# Built-in MCP server

`fbrcm mcp` exposes Remote Config workflows over stdio JSON-RPC using the official
Go MCP SDK v1.7.0. It supports the 2026-07-28 multi-round-trip interaction flow;
the SDK adapts interactions for older supported protocol versions. Browser
recovery requires URL elicitation (introduced in MCP 2025-11-25); mutation
confirmation requires form elicitation. Host support and UI vary.

## Architecture

MCP is a separate application mode alongside CLI and TUI. `main` routes
`fbrcm mcp` directly to the top-level `mcp` package, which owns startup,
protocol sessions, cancellation, and host interaction. The CLI retains a small
launch descriptor for help, completion, and capability discovery.

CLI and MCP invoke the same workflows in `ops/workflows`. Their shared
definitions contain option defaults, argument constraints, response DTOs, and
handlers. `cli/operation` adapts them to Cobra; MCP binds structured input
directly through `ops.Registry`. It does not construct command argv,
execute Cobra commands per tool call, or pass through CLI startup. Each call
gets fresh option values, context, output, and warning state.

The versioned contract lives in `ops/contract`; its public schema URNs
and envelopes are unchanged. `go run ./cmd/schemagen` also publishes the embedded
capability catalog used by MCP, so protocol startup needs no CLI command tree.

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

Every MCP tool has a namespaced name. MCP names are separate from CLI/shared
operation IDs: `parameters.get`, `parameters.add`, `parameters.update`,
`parameters.delete`, and `parameters.duplicate` invoke `get`, `add`, `update`,
`delete`, and `duplicate`; `plan.apply` invokes `apply`, and `diagnostics.doctor`
invokes `doctor`. Other names match their operation IDs. Only the names listed
below are exposed; there are no unqualified aliases. Toolset membership is
independent of the name prefix: `plan.apply` belongs to `publish`, not `plans`.
Response `command`, `requested_command`, and schema IDs retain the original
operation IDs, so CLI commands and the shared JSON contract are unchanged.

| Group | MCP tool names |
| --- | --- |
| `inspect` | `parameters.get`; `projects.list`, `projects.diff`; `project.show`, `project.defaults`; `groups.list`; `conditions.list`, `conditions.show`, `conditions.validate`; `versions.list`, `versions.show`, `versions.diff`; `experiments.list`, `experiments.show`; `rollouts.list`, `rollouts.show`; `personalizations.list`, `personalizations.show` |
| `edit` | `parameters.add`, `parameters.update`, `parameters.delete`, `parameters.duplicate`; `groups.add`, `groups.edit`, `groups.rename`, `groups.delete`; `conditions.add`, `conditions.edit`, `conditions.rename`, `conditions.move`, `conditions.delete` |
| `drafts` | `draft.list`, `draft.show`, `draft.diff`, `draft.change-note`, `draft.discard` |
| `plans` | `plan.show`, `plan.validate` |
| `publish` | `plan.apply`, `draft.publish`, `project.import`, `project.export`, `projects.promote`, `versions.export`, `versions.rollback`, `versions.restore`, `experiments.delete`, `rollouts.delete` |
| `diagnostics` | `diagnostics.doctor` (opt-in) |

Without `--allow-writes`, mutation tools and explicit `to`/`plan-out` options
are unavailable. Inspection is not filesystem-pure: it can refresh caches,
synchronize the registry, and persist credentials. Doctor can create and remove
diagnostic probe files. Stateless mode disables application-managed local state.

Each tool publishes semantic input and output schemas. Output schemas explicitly
declare `"type": "object"` at the root for compatibility with legacy MCP clients.
At the MCP boundary, nullable/multi-type schemas use `anyOf` with scalar types,
forbidden schemas use `{"not":{}}`, and unrestricted values explicitly allow all
JSON types. These portability rewrites preserve validation semantics; they do
not allow forbidden fields or restrict free-form values to objects. Published
CLI schemas and response envelopes are unchanged.

MCP additionally allows omission of empty invocation wrappers where valid:
`arguments` and `options` default to `{}`, and `stdin` defaults to `null`.
The tool schema advertises these defaults for form-based clients such as
Inspector; the server also applies them, so agents need not supply boilerplate.
Required inputs stay required: for example, `conditions.list` still needs
`arguments.project`, and `parameters.add` still needs its parameter and
value/type options.
Conditional requirements and exclusions also apply to omitted fields exactly as
they do to explicit defaults. Explicit nulls are not replaced with empty objects,
and no individual flag, selector, file path, or document content is synthesized.

Example `parameters.get` input:

```json
{
  "options": {"project": ["=acme-staging"]}
}
```

This is equivalent to supplying `"arguments": {}` and `"stdin": null` explicitly.
Defaults are applied before validation and before matching a suspended operation,
so switching between omitted wrappers and explicit defaults does not replay a
confirmed mutation. The CLI's normalized invocation contract is unchanged.

Argument names come from the schema. Options use long flag names without `--`;
repeatable values are arrays, booleans are JSON booleans. Supplied `stdin` is a
JSON document rather than a JSON-encoded string. Inputs are limited to 16 MiB.
Selection, filtering, expressions, validation, dry runs, plans, drafts, and
non-atomic batch behavior remain consistent with CLI execution.

Completed results contain the unchanged CLI envelope as `structuredContent`
and identical JSON in text content. Inspect `outcome`, `errors`, `warnings`, and
`data`, not just `exit_code`: changed diffs succeed with exit `1`. Partial success
is an error result retaining completed/failed target data.

## Manual smoke tests with Inspector

No agent or Firebase credentials are needed for discovery and an in-memory
`parameters.get` call. With Node.js 22.19 or newer, run from the repository root:

```sh
npx @modelcontextprotocol/inspector go run . mcp -- --stateless --toolsets inspect
```

Open the local URL printed by Inspector and connect. Inspector launches fbrcm;
do not start a separate server. The tools list should contain `parameters.get`
and other inspection tools, with no malformed-entry warnings or mutation tools.
Call `parameters.get` with this tool input (the `stdin` property is a JSON
object, not a string):

```json
{
  "stdin": {
    "parameters": {
      "smoke_test": {"defaultValue": {"value": "hello"}}
    }
  }
}
```

Expect `structuredContent.outcome` to be `success`, `exit_code` to be `0`, and
`data.count` to be `1`. Set `options` to `{"search":"missing"}` to check an empty
success, then restore `{}` to get one item again. Supplying `{"yes":true}` must
fail validation; a subsequent valid call must still work without reconnecting.
These calls do not contact Firebase or use fbrcm profiles and caches.

In Inspector's generated form, empty optional wrappers initialize from the
advertised defaults. For the offline smoke test, enter the Remote Config object
above in the `stdin` field; do not paste the enclosing tool input there. In
"Edit as JSON", paste the complete example. Tools with required arguments still
need those values filled in before execution. Both editing modes call the same
server-side validation and operations.

Also run Inspector's strict portability check; successfully listing or calling a
tool alone does not catch schema-portability errors or warnings:

```sh
npx @modelcontextprotocol/inspector --cli go run . mcp --stateless --toolsets inspect -- --method tools/list --strict --format json
```

Expect exit `0` and no `schemaFindings` in the result. Strict mode fails on
portability errors, but warnings must also be checked. In Inspector's CLI mode,
arguments before `--` belong to fbrcm and options after it belong to Inspector.

Inspector defaults to the **Legacy** protocol era, which uses the `initialize`
handshake (normally MCP `2025-11-25`). This is not an error or a consequence of
fbrcm's `--stateless` flag. In Inspector's server settings, choose **Auto** or
**Modern** and reconnect to exercise MCP `2026-07-28` discovery instead. Both
protocol eras support the smoke test above. See the
[Inspector configuration reference](https://github.com/modelcontextprotocol/inspector/blob/main/docs/mcp-server-configuration.md)
for protocol-era settings and launch-argument handling.

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
logs use stderr as plain text, without ANSI colors, decorations, or hyperlink
escape sequences. This is automatic in MCP mode; no `NO_COLOR` setting is needed.
`FBRCM_LOG_LEVEL` and `FBRCM_LOG_NO_TIMESTAMP` still control verbosity and
timestamps. Set `FBRCM_LOG_PLAIN=1` to enable the same escape-free mode for CLI
or TUI log entries, for example in CI. Any non-empty value enables it; an empty
or unset value preserves normal CLI/TUI log styling but does not disable plain
MCP logs. External hook output is not sanitized by this logger setting.
`--json` is incompatible with the streaming command.
