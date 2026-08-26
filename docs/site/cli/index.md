# Command-line interface

Any argument switches fbrcm from its interactive TUI into CLI mode:

```sh
fbrcm projects list
fbrcm get feature_enabled
fbrcm draft list
```

Use `fbrcm help <command>` for flags accepted by the installed version and the
[command index](/reference/commands) for a map of the complete surface.

## Global behavior

Global options may appear before or after the command:

```sh
fbrcm --profile production get feature_enabled
fbrcm --json projects list
fbrcm projects list --json
fbrcm --stateless get feature_enabled --project '=my-project'
```

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
fbrcm get feature_enabled --project '^prod'
fbrcm get feature_enabled --project '=acme-production-42'
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
fbrcm get --project 'server@=acme-production-42'
```

## Human and machine output

Human mode may use tables, color, diffs, and confirmations. `--json` returns one
versioned envelope on success or failure and never opens prompts, editors,
browsers, or file pickers.

```sh
fbrcm --json get feature_enabled --project '=prod'
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
| `--change-note <text>` | Attach a single-line Firebase version note |
| `--yes` / `-y` | Accept the human confirmation non-interactively |

Batch mutations are per-target and non-atomic. fbrcm continues after a
project-scoped failure and reports every result.

## Stateless mode

Stateless mode is intended for CI or disposable environments:

```sh
FBRCM_GOOGLE_ACCESS_TOKEN="$TOKEN" \
  fbrcm --stateless --json get feature_enabled \
  --project '=acme-production-42'
```

It does not read or write the project registry, caches, drafts, profiles, or
publication hooks. Draft operations are unavailable because there is nowhere
to persist them.
