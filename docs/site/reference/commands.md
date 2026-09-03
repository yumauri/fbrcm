# Command index

Use this page to find a command. Run `fbrcm help <command>` for its full set of
flags, or `fbrcm capabilities <command...> --json` for a machine-readable
definition from the installed binary.
Complete [CLI-only setup](/guide/#option-2-setup-using-only-the-cli) before using
project or Remote Config commands for the first time.

## Parameters and structure

| Command | Purpose |
| --- | --- |
| `add` | Add one parameter to selected targets |
| `duplicate` | Copy a parameter under a new key |
| `get` | Read and filter parameters |
| `update` | Change selected parameters |
| `delete` | Delete selected parameters |
| `groups list` | List parameter groups |
| `groups add` | Create an empty or described group |
| `groups edit` | Change a group description |
| `groups rename` | Rename a group |
| `groups delete` | Delete a group and its parameters |

## Conditions

| Command | Purpose |
| --- | --- |
| `conditions list` | List conditions in evaluation order |
| `conditions show` | Inspect one condition and its usages |
| `conditions add` | Add a condition |
| `conditions edit` | Change the expression or display color |
| `conditions rename` | Rename a condition and its references |
| `conditions move` | Change evaluation priority |
| `conditions delete` | Remove a condition and conditional values |
| `conditions validate` | Validate condition expressions and references |

## Drafts

| Command | Purpose |
| --- | --- |
| `draft list` | List healthy and recoverable drafts |
| `draft path` | Print the draft directory |
| `draft show` | Inspect or export a draft candidate |
| `draft change-note` | Set or clear its Firebase version note |
| `draft diff` | Compare against the base or current Firebase state |
| `draft publish` | Validate and publish drafts |
| `draft discard` | Remove local draft state |

## Plans

| Command | Purpose |
| --- | --- |
| `plan show` | Verify a publication plan and show its targets and diffs |
| `plan validate` | Verify publication plan structure and integrity |
| `apply` | Preflight and apply the exact candidates in a publication plan |

## Projects and templates

| Command | Purpose |
| --- | --- |
| `projects list` / `update` | Inspect or synchronize accessible projects |
| `projects forget` / `reset` | Remove selected or all local project state |
| `projects diff` | Compare two template targets |
| `projects promote` | Transfer selected configuration to another target |
| `projects aliases ...` | List, set, remove, or import repository aliases |
| `projects path` | Print the project registry path |
| `project show` | Inspect one project registration |
| `project templates show/set` | Inspect or change enabled template types |
| `project quota-project show/set/unset` | Inspect or override one project's quota project |
| `project open` | Open or return the Firebase Console URL |
| `project export` / `import` | Move Remote Config documents |
| `project defaults` | Download application defaults |

## History and managed features

| Command | Purpose |
| --- | --- |
| `versions list/show` | Inspect Remote Config history |
| `versions diff` | Compare historical or current templates |
| `versions export` | Export one historical template |
| `versions rollback` | Ask Firebase to roll back to a retained version |
| `versions restore` | Republish a locally cached snapshot |
| `experiments list/show/delete` | Inspect or delete A/B tests |
| `rollouts list/show/delete` | Inspect or delete rollouts |
| `personalizations list/show` | Inspect personalization bindings |

## Local environment

| Command | Purpose |
| --- | --- |
| `doctor` | Diagnose credentials, APIs, connectivity, and storage |
| `cache list/path/clear` | Inspect or remove cached templates |
| `config path/show/set/reset/validate/edit` | Manage configuration |
| `hooks status/fingerprint/trust/untrust` | Manage repository hook trust |
| `auth list/add/login/path/delete` | Add, validate, inspect, or remove identities |
| `auth quota-project show/set/unset` | Manage an identity's default quota project |
| `auth bind` | Select the identity used by saved projects |
| `profile list/switch/rename/path/delete` | Manage isolated workspaces |
| `theme list/switch/reset/path/rename/delete/import` | Manage palettes |
| `completion` | Generate shell completion scripts |

## Machine discovery

| Command | Purpose |
| --- | --- |
| `mcp` | Start the [MCP server](/automation/mcp) for an AI application |
| `capabilities` | Describe commands, inputs, schemas, and side effects |
| `schema list` | List published JSON schemas |
| `schema show` | Return one JSON Schema document |
