# MCP server

Use the built-in MCP server to work with Firebase Remote Config from an
MCP-compatible AI application. The application launches `fbrcm mcp` and calls
its tools over stdio. You do not need to start a separate server before chatting.

MCP exposes Remote Config workflows, including inspection, editing, drafts,
plans, and publication. It uses the same operations and result envelopes as
the CLI, with permissions fixed when the server starts.

## Connect your AI application

Install fbrcm and configure a profile using the
[setup guide](/guide/#option-2-setup-using-only-the-cli). Authentication identities
and discovered projects belong to that profile; see
[Authentication and discovery](/guide/authentication).

Add the binary to your application's MCP server configuration. For hosts that
accept the common `mcpServers` format:

```json
{
  "mcpServers": {
    "firebase": {
      "command": "/absolute/path/to/fbrcm",
      "args": ["mcp", "--profile", "personal", "--no-local-config"]
    }
  }
}
```

Replace the binary path and profile name. The selected profile must already
exist. The host starts and stops the process; do not add `--json` to this
configuration, because stdout carries MCP protocol messages.

`--no-local-config` ignores repository `.fbrcm.toml` settings. Omit it when you
want repository configuration and aliases to apply. Tool discovery does not
require Google credentials; authentication happens when a call needs Firebase.

## Make a first tool call

Ask the agent to inspect a parameter in a specific project, or call
`parameters.get` directly from your host's tool interface:

```json
{
  "arguments": {"parameter": "feature_enabled"},
  "options": {"project": ["=example-project-id"]}
}
```

The `=` prefix selects an exact project ID. Use exact selectors for agent
workflows; see [Filtering](/reference/filtering) for other query modes.

Use each tool's discovered input schema when constructing calls:

- `arguments` contains named positional arguments.
- `options` contains long flag names without `--`. Repeatable values, such as
  `project`, are JSON arrays; booleans are JSON booleans.
- `stdin` contains a JSON document for tools that accept document input, not a
  JSON-encoded string.

Optional `arguments` and `options` default to `{}` where valid, and optional
`stdin` defaults to `null`. Required values must still be supplied. For example,
`conditions.list` requires `arguments.project`. The agent cannot pass `yes`,
switch profiles, or change server permissions through a tool call.

## Choose tools and permissions

`--toolsets` selects groups. The default is
`inspect,edit,drafts,plans,publish`; diagnostics is opt-in. Discovery lists only
tools allowed by the selected groups, execution mode, and write permission.

| Toolset | Workflows |
| --- | --- |
| `inspect` | Read parameters, projects, groups, conditions, versions, experiments, rollouts, and personalizations; compare configurations and read defaults |
| `edit` | Add, update, duplicate, and delete parameters; manage groups and conditions |
| `drafts` | List, inspect, compare, annotate, and discard local drafts |
| `plans` | Inspect and validate publication plans |
| `publish` | Apply plans, publish drafts, import/export templates, promote between projects, restore/rollback versions, and delete experiments or rollouts |
| `diagnostics` | Run `diagnostics.doctor` |

All `edit` and `publish` tools require `--allow-writes`, as do
`draft.change-note`, `draft.discard`, and explicit `to` or `plan-out` output.
Mutation tools require write permission even for a dry run. `plan.apply`
belongs to `publish`, not `plans`.

For inspection only, use these host arguments:

```json
["mcp", "--profile", "personal", "--no-local-config", "--toolsets", "inspect"]
```

For read/write workflows with user approval:

```json
["mcp", "--profile", "personal", "--no-local-config", "--allow-writes"]
```

::: warning Write authorization
The default `--confirmation=host` requires the host to present an approval form
for the exact tool and inputs. This uses MCP form elicitation. Without that
support, the call returns an interaction-required error. `--confirmation=none`
authorizes enabled writes without per-call approval. Use it only when you intend
to run unattended.
:::

Use [drafts](/cli/drafts) to compose changes and [plans](/cli/plans) to hand off
an exact publication candidate. Multi-target publication is not atomic. Inspect
every target's result before retrying a failed batch.

Hooks are disabled unless `--allow-hooks` is set, and must pass the normal
[trust checks](/reference/hooks). Inspection can refresh caches, synchronize the
project registry, and persist refreshed credentials. `diagnostics.doctor` can
create and remove local probe files.

There are no tools for managing profiles, credentials, application configuration,
aliases, hook trust, themes, or shell completion. File paths refer to the machine
running fbrcm and use its filesystem permissions. The server has no HTTP MCP
endpoint, resources, prompts, or background-task API.

## Read tool results

Completed calls return the [JSON envelope](./json-contract#envelope) in
`structuredContent`, with identical JSON in text `content`. Inspect `outcome`,
`errors`, `warnings`, and `data` together.

`success` maps to `isError: false`, including a changed diff with exit code `1`.
`failure` and `partial_success` map to `isError: true`; partial results retain
completed and failed target information. A tool failure does not stop the server.

The envelope identifies the underlying operation rather than the public MCP
name. Parameter tools use unprefixed operation IDs, so `parameters.get` returns
`command: "get"`. `plan.apply` uses `apply`, and `diagnostics.doctor` uses
`doctor`. Other tool names match their operation IDs.

## Authentication recovery

OAuth access tokens normally refresh silently. When browser authorization is
needed, a host with URL elicitation can offer sign-in:

1. fbrcm starts a temporary localhost callback listener and gives the host the
   sign-in URL.
2. The user approves navigation and signs in through the browser.
3. fbrcm receives the callback, exchanges the authorization code directly with
   Google, saves credentials, and closes the listener.
4. The suspended operation continues. Sign-in does not grant mutation approval.

Tokens and authorization secrets do not pass through chat or tool arguments.
Canceling, timing out, or disconnecting cleans up the sign-in attempt.

With `--browser-auth=never`, or a host without URL elicitation, follow the
returned remediation and sign in externally:

```sh
fbrcm auth login example-auth-name --profile personal
```

Then retry the tool call. Credentials are reloaded for each operation, so
reconnecting is not normally necessary. If fbrcm runs through SSH or in a
container, the browser must be able to reach its localhost callback; external
login may be more practical.

## Run without local state

Use `--stateless` and supply `FBRCM_GOOGLE_ACCESS_TOKEN` securely through the
host environment for Firebase calls. Do not paste the token into chat or commit
it in a configuration file.

Stateless execution uses no profiles, application configuration, caches, drafts,
or hooks. You can still read and write files whose paths you supply. Remote
mutations require the same write permission and approval as stateful execution.

Discovery excludes the `drafts` and `plans` groups, `draft.publish`,
`versions.restore`, and `diagnostics.doctor`. `plan.apply` supports stateless
plans. `--stateless` cannot be combined with `--profile` or `--allow-hooks`.

Environment tokens are not refreshed. Renew them externally and restart the
subprocess with the updated environment; there is no profile or browser fallback.

## Timeouts and logs

| Setting | Default | Scope |
| --- | --- | --- |
| `--request-timeout` | `5m` | One operation, including queueing and user interaction |
| `--auth-timeout` | `2m` | One browser sign-in attempt |
| `--timeout` | Unlimited | Entire server lifetime |

Explicit durations must be positive; the earliest applicable deadline wins.
Your host may impose a shorter timeout. Tool operations execute one at a time,
while discovery stays responsive. Disconnecting cancels work; inspect Firebase
state before retrying an interrupted mutation.

Logs go to stderr as plain text without terminal escape codes. Set
`FBRCM_LOG_LEVEL` to control verbosity and `FBRCM_LOG_NO_TIMESTAMP` to omit
timestamps. Inspector shows these logs in its Server Console; they are not MCP
logging notifications. External hook output is not sanitized by the logger.

## Test with Inspector

No agent or Firebase credentials are needed for this smoke test. With `fbrcm`
available on your `PATH` and Node.js 22.19 or newer installed, run:

```sh
npx @modelcontextprotocol/inspector fbrcm mcp -- --stateless --toolsets inspect
```

Open Inspector's local URL and connect. Select `parameters.get`, switch to
**Edit as JSON**, and execute:

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
`data.count: 1`. The call reads the supplied document without contacting Firebase.
In the generated form, enter only the document starting with
`{"parameters": ...}` inside the `stdin` field, not the enclosing tool input.

Inspector's web launcher consumes the `--` separator. When launching fbrcm
directly, use `fbrcm mcp --stateless --toolsets inspect` without that separator.
Inspector's **Legacy** and **Modern** protocol modes both support this test;
protocol selection is independent of fbrcm's `--stateless` execution policy.

For the complete tool catalog and schema-validation checks, see the
[MCP reference](https://github.com/yumauri/fbrcm/blob/main/docs/MCP.md).
