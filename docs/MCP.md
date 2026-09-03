# Built-in MCP server

`fbrcm mcp` lets an MCP-compatible AI application work with Firebase Remote
Config. The application launches fbrcm as a subprocess and communicates over
stdio JSON-RPC. The server exposes tools; it has no HTTP MCP endpoint,
resources, prompts, or background-task API.

For launch flags and defaults, see the [command reference](CLI.md#mcp-server).
For package boundaries and implementation details, see
[architecture](architecture.md#mcp).

## Setup

Configure your AI application with the binary path, arguments, and environment.
For hosts accepting the common `mcpServers` configuration format:

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

The selected profile must already exist. Set up profiles and authentication
through the CLI or TUI. The host starts the server when connecting; you do not
need to launch a separate process before chatting. Do not pass `--json`. Stdout
is reserved for the MCP protocol.

Tool discovery does not require Google credentials. Authentication happens when
an operation needs Firebase, using the selected profile's configured Google
OAuth, service-account, or gcloud identities.

## Permissions and scope

The launch configuration fixes the profile, execution mode, toolsets, and
permissions for the server process. Tool calls cannot change these settings.

- Writes are disabled by default. `--allow-writes` exposes mutation tools and
  explicit artifact-output options. Mutation tools require this permission even
  for dry runs.
- `--confirmation=host` asks the user to approve the exact tool name and inputs
  before a mutation or explicit artifact write. The host must support MCP form
  elicitation so it can present an approval form. Without that support, the call
  returns an interaction-required error. Tool callers cannot pass `yes`.
- `--confirmation=none` authorizes enabled mutations without per-call approval;
  use it only when unattended execution is intended.
- Hooks are disabled unless `--allow-hooks` is set. Configured hooks must also
  pass the normal trust checks.

Stateful execution uses the profile's configuration and aliases. Repository
`.fbrcm.toml` settings apply unless `--no-local-config` is set. Inspection can
refresh caches, synchronize the project registry, and persist refreshed
credentials; `diagnostics.doctor` can create and remove probe files.

There are no tools for managing profiles, credentials, application configuration,
aliases, project registry settings, hook trust, themes, or shell completion.
External editors, terminal pickers, and the TUI are not available through tool
calls.

File paths refer to the machine running fbrcm and resolve relative to its working
directory. File access uses that process's filesystem permissions; the MCP
connection does not provide a filesystem sandbox.

## Tool catalog

`--toolsets` selects groups, with `inspect,edit,drafts,plans,publish` enabled by
default. `diagnostics` is opt-in. Discovery lists only tools permitted by the
selected groups, write permission, and execution mode.

| Group | MCP tool names |
| --- | --- |
| `inspect` | `parameters.get`; `projects.list`, `projects.diff`; `project.show`, `project.defaults`; `groups.list`; `conditions.list`, `conditions.show`, `conditions.validate`; `versions.list`, `versions.show`, `versions.diff`; `experiments.list`, `experiments.show`; `rollouts.list`, `rollouts.show`; `personalizations.list`, `personalizations.show` |
| `edit` | `parameters.add`, `parameters.update`, `parameters.delete`, `parameters.duplicate`; `groups.add`, `groups.edit`, `groups.rename`, `groups.delete`; `conditions.add`, `conditions.edit`, `conditions.rename`, `conditions.move`, `conditions.delete` |
| `drafts` | `draft.list`, `draft.show`, `draft.diff`, `draft.change-note`, `draft.discard` |
| `plans` | `plan.show`, `plan.validate` |
| `publish` | `plan.apply`, `draft.publish`, `project.import`, `project.export`, `projects.promote`, `versions.export`, `versions.rollback`, `versions.restore`, `experiments.delete`, `rollouts.delete` |
| `diagnostics` | `diagnostics.doctor` |

All `edit` and `publish` tools, plus `draft.change-note` and `draft.discard`,
require `--allow-writes`. The `to` and `plan-out` options also require it.
A tool's name prefix does not determine its toolset. For example, `plan.apply`
belongs to `publish`, not `plans`.

Tools support the same selection, filtering, expressions, validation, dry runs,
drafts, and plans as their corresponding CLI operations, subject to launch
policy. Multi-target publication is not atomic; inspect each target's result.

## Tool input and results

Use the input schema returned by discovery to construct each call. Inputs have
three properties:

- `arguments`: named positional arguments, such as `parameter` or `project`.
- `options`: long CLI flag names without `--`. Repeatable options are JSON
  arrays, booleans are JSON booleans, and scalar values use their declared types.
- `stdin`: a JSON document for operations that accept document input, not a
  JSON-encoded string or a read from the protocol stream.

For example, call `parameters.get` with:

```json
{
  "options": {"project": ["=acme-staging"]}
}
```

Where valid for the tool, omitted `arguments` and `options` default to `{}`,
and omitted `stdin` defaults to `null`. Required arguments and conditional
requirements still apply: `conditions.list` needs `arguments.project`, and
`parameters.add` needs its parameter and value/type options. Explicit values,
including `null`, are validated as supplied. Inputs are limited to 16 MiB.

Completed calls return the [machine envelope](cli-contract.md#envelope) in
`structuredContent`, with the same JSON in text `content`. Inspect `outcome`,
`errors`, `warnings`, and `data` together:

- `success` maps to `isError: false`, including a changed diff with exit code
  `1`.
- `failure` and `partial_success` map to `isError: true`. Partial results
  retain completed and failed target information.

The envelope's `command`, `requested_command`, and schema IDs identify the
shared operation, not the public MCP name. Parameter tools use their unprefixed
operation IDs (`parameters.get` returns `command: "get"`); `plan.apply` uses
`apply`, and `diagnostics.doctor` uses `doctor`. Other tool names match their
operation IDs.

Confirmations and authentication can suspend a call while the host requests
user input. The host resumes the same operation through MCP; an interaction
request is not a completed machine envelope. Use the advertised MCP schemas for
validation; their generation is described in the [schema reference](../schemas/README.md#mcp-schema-adaptation).

## Authentication recovery

OAuth access tokens normally refresh silently. If Google requires browser
authorization and the host supports URL elicitation:

1. fbrcm starts a temporary callback listener on `127.0.0.1` using an available
   port, then asks the host to offer the sign-in URL.
2. The user approves navigation and signs in through the browser. fbrcm does not
   open the browser automatically.
3. Google redirects to the callback. fbrcm verifies the authorization response,
   exchanges the code, persists credentials, and closes the listener.
4. The suspended operation continues. Signing in does not grant mutation
   approval.

Tokens and authorization secrets are exchanged directly with Google, not
through tool arguments or chat. Canceling, timing out, or disconnecting cleans
up the authorization attempt.

With `--browser-auth=never` or a host without URL elicitation, follow the
structured remediation. Run `fbrcm auth login <auth-id> --profile <profile>`
in a terminal, then retry. Operations reload credentials, so reconnecting is
not normally necessary.

The browser must be able to reach localhost on the fbrcm machine. SSH,
containers, or remote hosts may require external login instead.

## Stateless mode

Launch with `--stateless` and securely supply `FBRCM_GOOGLE_ACCESS_TOKEN`
through the host environment for Firebase calls. Do not paste tokens into chat
or commit them.

Stateless execution uses no profiles, application configuration, caches, drafts,
or hooks. You can still read and write files whose paths you supply. Remote
mutations require the same write and confirmation permissions as stateful
execution.

Discovery excludes all `drafts` and `plans` tools, `draft.publish`,
`versions.restore`, and `diagnostics.doctor`. `plan.apply` is available for
stateless plans. The fbrcm execution policy is independent of MCP's
protocol-level use of the term "stateless."

Environment tokens are not refreshed. Renew them externally and restart the
subprocess with the updated environment. There is no fallback to profile or
browser authentication.

## Timeouts, lifecycle, and logs

The host manages the process lifetime. `--timeout` can impose an overall
deadline; `--request-timeout` bounds each operation, including queueing and
user interaction; `--auth-timeout` bounds each browser sign-in attempt. The
earliest applicable deadline wins. See the [launch reference](CLI.md#mcp-server)
for defaults.

Tool operations execute one at a time while discovery stays responsive.
Ordinary tool failures leave the server running. Disconnects and interrupts
cancel work and close callback listeners. Interrupted multi-target mutations
may have completed some writes; inspect their results before retrying.

Logs go to stderr as plain text without ANSI colors, decorations, or hyperlinks.
Inspector displays these in its Server Console; they are not MCP logging
notifications. `FBRCM_LOG_LEVEL` controls verbosity and
`FBRCM_LOG_NO_TIMESTAMP` controls timestamps. Plain formatting is automatic,
regardless of `NO_COLOR` or `FBRCM_LOG_PLAIN`. External hook output is not
sanitized by the logger.

## Manual smoke tests with Inspector

No agent or Firebase credentials are needed for discovery and an in-memory
`parameters.get` call. With `fbrcm` available on your `PATH` and Node.js 22.19 or
newer installed, run:

```sh
npx @modelcontextprotocol/inspector fbrcm mcp -- --stateless --toolsets inspect
```

Open the local URL printed by Inspector and connect. Inspector launches fbrcm.
Select `parameters.get`, switch to "Edit as JSON", and enter:

```json
{
  "stdin": {
    "parameters": {
      "smoke_test": {"defaultValue": {"value": "hello"}}
    }
  }
}
```

Expect `structuredContent.outcome: "success"`, `exit_code: 0`, and
`data.count: 1`. Set `options` to `{"search":"missing"}` for an empty success.
Set `options` to `{"yes":true}` to check validation failure, then remove it and
call again. These checks do not contact Firebase or use profiles and caches.

In the generated form, enter only the Remote Config document in the `stdin`
field, starting with `{"parameters": ...}`. Other fields follow their declared
JSON types: a project filter is an array such as `["=acme-staging"]`, while a
positional project argument is a string. Required fields must be filled before
execution.

Inspector's web launcher consumes the `--` separator in the example above.
When running fbrcm directly, use `fbrcm mcp --stateless --toolsets inspect`,
without that separator.

### Schema validation

Run Inspector's strict portability check:

```sh
npx @modelcontextprotocol/inspector --cli fbrcm mcp --stateless --toolsets inspect -- --method tools/list --strict --format json
```

Expect exit `0` and no `schemaFindings`. Check both errors and warnings. In
Inspector's CLI mode, arguments before `--` belong to fbrcm and options after it
belong to Inspector.

### Protocol selection

Inspector's **Legacy** mode uses the `initialize` handshake with MCP
`2025-11-25`; **Modern** uses MCP `2026-07-28` discovery. Both support the
smoke test. Select a protocol era in the server settings and reconnect.
The protocol selection does not change fbrcm's `--stateless` execution policy.
See the [Inspector configuration reference](https://github.com/modelcontextprotocol/inspector/blob/main/docs/mcp-server-configuration.md)
for host configuration details.
