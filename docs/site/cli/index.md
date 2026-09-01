# Command-line interface

Any argument switches fbrcm from its interactive TUI into CLI mode:

```sh
fbrcm projects list
fbrcm get feature_enabled
fbrcm draft list
```

If you have not configured credentials and discovered projects yet, complete
the [CLI-only setup path](/guide/#option-2-setup-using-only-the-cli) first.

Use `fbrcm help <command>` for flags accepted by the installed version and the
[command index](/reference/commands) for a map of the complete command set.

## How to read command examples

Angle-bracket names describe positional arguments; do not type the brackets.
For example, the command syntax is:

```text
fbrcm auth quota-project set <auth-id> <quota-project-id>
```

In a concrete invocation:

```sh
fbrcm auth quota-project set example-auth-name example-quota-project-id
```

- `auth`, `quota-project`, and `set` are nested commands.
- `example-auth-name` is a local auth ID chosen when the credential was added.
- `example-quota-project-id` is a physical Google Cloud project ID.

Examples throughout this site use `example-project-id` as a physical Firebase
project ID, `staging` and `prod` as optional repository aliases, and
`feature_enabled` as a parameter key. Replace them with your own values.

## Global behavior

Global options may appear before or after the command:

```sh
fbrcm --profile personal get feature_enabled
fbrcm --json projects list
fbrcm projects list --json
```

`--profile personal` requires that profile to exist already. Create or select
it persistently with `fbrcm profile switch personal`.

Important global controls include:

| Option | Purpose |
| --- | --- |
| `--profile <name>` | Select an fbrcm profile for this invocation |
| `--json` | Emit the versioned machine envelope |
| `--stateless` | Operate directly from an access token without local state |
| `--no-local-config` | Ignore repository `.fbrcm.toml` configuration |

## Select projects precisely

Most multi-project commands accept repeatable `--project` / `-p` filters.

```sh
fbrcm get feature_enabled --project prod
fbrcm get feature_enabled --project '^acme-'
fbrcm get feature_enabled --project '=example-project-id'
```

Filter prefixes are:

| Prefix | Match mode |
| --- | --- |
| none or `~` | Fuzzy |
| `^` | Starts with |
| `/` | Contains |
| `=` | Exact |

Repeated project filters are ORed. Use `client@` or `server@` before a project
query to select a specific template type:

```sh
fbrcm get --project 'server@=example-project-id'
```

## Human and machine output

Human mode may use tables, color, diffs, and confirmations. `--json` returns one
versioned envelope on success or failure and never opens prompts, editors,
browsers, or file pickers.

```sh
fbrcm --json get feature_enabled --project '=example-project-id'
```

Machine consumers should inspect structured problem codes and semantic exit
statuses rather than parsing error text. See the [JSON contract](/automation/json-contract).

## Reads and writes

Read commands use drafts when appropriate and otherwise use cached published
state. Add `--update` to revalidate a cache before reading.

Write commands normally:

1. load each selected target;
2. apply the targeted mutation in memory;
3. print the complete Remote Config diff;
4. ask for confirmation;
5. validate with Firebase; and
6. publish with ETag protection.

Common write controls are:

| Option | Behavior |
| --- | --- |
| `--dry-run` | Validate and preview without publishing or writing local state |
| `--draft` | Save the mutation into a local draft |
| `--plan-out <path>` | Write an immutable validated publication plan instead of publishing |
| `--change-note <text>` | Attach a single-line Firebase version note |
| `--yes` / `-y` | Accept the human confirmation non-interactively |

Batch mutations are per-target and non-atomic. fbrcm continues after a
project-scoped failure and reports every result.

Use a [draft](/cli/drafts) to compose and revise several edits in the active
profile. Use a [plan](/cli/plans) to preserve one exact candidate for review,
handoff, or later application.

## Stateless mode

Stateless mode runs in CI or a disposable environment without fbrcm-managed
local state:

```sh
GOOGLE_CLOUD_QUOTA_PROJECT=example-quota-project-id \
FBRCM_GOOGLE_ACCESS_TOKEN="$(gcloud auth application-default print-access-token)" \
  fbrcm --stateless --json get feature_enabled \
  --project '=example-project-id'
```

It does not read or write the project registry, caches, drafts, profiles, or
publication hooks. fbrcm keeps the short-lived token in memory. Draft
operations are unavailable because there is nowhere to persist them. See
[Automation and agents](/automation/) for stateless permissions and capability
discovery.
