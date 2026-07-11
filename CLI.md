# fbrcm CLI

`fbrcm` is a Firebase Remote Config manager. It runs as an interactive TUI when called with no arguments. Any argument switches to CLI mode.

## Command Tree

```text
fbrcm [--help] [--version]
│
├── add <parameter>
│   ├── --project, -p <query>  repeated
│   ├── --expr <expr>
│   ├── --dry-run
│   ├── --description <text>
│   ├── --group <name>
│   └── exactly one value flag:
│       ├── --boolean true|false
│       ├── --number <number>
│       ├── --string <text>
│       └── --json <json>
│
├── cache
│   ├── list [--json]
│   ├── path [--json]
│   └── purge [--yes|-y]
│
├── config
│   └── path [--json]
│
├── completion
│   ├── bash [--no-descriptions]
│   ├── fish [--no-descriptions]
│   ├── powershell [--no-descriptions]
│   └── zsh [--no-descriptions]
│
├── delete [parameter]
│   ├── --project, -p <query>  repeated
│   ├── --filter, -f <query>   repeated
│   ├── --expr <expr>
│   ├── --search <text>
│   ├── --dry-run
│   └── --yes, -y
│
├── get [parameter]
│   ├── --project, -p <query>  repeated
│   ├── --filter, -f <query>   repeated
│   ├── --expr <expr>
│   ├── --search <text>
│   ├── --json
│   ├── --all
│   └── --update
│
├── help [command]
│
├── auth
│   ├── list [--json]
│   ├── add oauth <auth-id> [--from <path>] [--label <label>]
│   ├── add service-account <auth-id> [--from <path>] [--label <label>]
│   ├── add gcloud <auth-id> [--label <label>]
│   ├── login <auth-id> [--noopen]
│   ├── path <auth-id> [--json]
│   ├── purge <auth-id> [--yes|-y]
│   └── bind <project-query> --auth <auth-id>
│
├── profile
│   ├── list [--json]
│   ├── path <profile> [--json]
│   ├── purge <profile> [--yes|-y]
│   ├── rename <old-name> <new-name>
│   └── switch <name>
│
├── project
│   ├── export <project> [--to <path>]
│   ├── import <project>
│   │   ├── --from <path>
│   │   ├── --group <name>        repeated
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --expr <expr>
│   │   ├── --search <text>
│   │   ├── --dry-run
│   │   ├── --remove-all-conditions
│   │   ├── --remove-project-specific-conditions
│   │   ├── --merge
│   │   ├── --override
│   │   └── --merge-resolve current|import
│   └── versions
│       ├── list <project>
│       │   ├── --limit <n>
│       │   ├── --all
│       │   ├── --before <version>
│       │   ├── --since <RFC3339>
│       │   ├── --until <RFC3339>
│       │   ├── --cached
│       │   └── --json
│       ├── show <project> <version>
│       │   ├── --cached
│       │   └── --json
│       ├── diff <project> <from> [<to>]
│       │   ├── --filter, -f <query>  repeated
│       │   ├── --group <name>        repeated
│       │   ├── --expr <expr>
│       │   ├── --search <text>
│       │   ├── --parameters
│       │   ├── --conditions
│       │   ├── --cached
│       │   └── --json
│       ├── export <project> <version>
│       │   ├── --to <path>
│       │   └── --cached
│       ├── rollback <project> <version>
│       │   ├── --dry-run
│       │   ├── --yes, -y
│       │   └── --json
│       └── restore <project> <version>
│           ├── --dry-run
│           ├── --yes, -y
│           └── --json
│
├── projects
│   ├── list
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --expr <expr>
│   │   ├── --json
│   │   ├── --update
│   │   └── --url
│   ├── update
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --expr <expr>
│   │   ├── --json
│   │   ├── --url
│   │   └── --auth <auth-id>
│   ├── diff <source-project> <target-project>
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --group <name>        repeated
│   │   ├── --expr <expr>
│   │   ├── --search <text>
│   │   ├── --parameters
│   │   ├── --conditions
│   │   ├── --cached
│   │   └── --json
│   ├── promote <source-project> <target-project>
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --group <name>        repeated
│   │   ├── --expr <expr>
│   │   ├── --search <text>
│   │   ├── --parameters
│   │   ├── --conditions
│   │   ├── --interactive
│   │   ├── --all
│   │   ├── --prune
│   │   ├── --dry-run
│   │   ├── --yes, -y
│   │   └── --json
│   ├── path [--json]
│   └── purge [--yes|-y]
│
└── update [parameter]
    ├── --project, -p <query>  repeated
    ├── --filter, -f <query>   repeated
    ├── --expr <expr>
    ├── --search <text>
    ├── --dry-run
    ├── --yes, -y
    ├── --description <text>
    ├── --group <name>
    ├── --no-group
    ├── --name <new-name>
    ├── --remove-all-conditional-values
    ├── --remove-conditional-value <condition>  repeated
    └── at most one value flag:
        ├── --boolean true|false
        ├── --number <number>
        ├── --string <text>
        └── --json <json>
```

## Shared Behavior

All commands support `--help`. Root also supports `--version`.

Most commands require an active profile. `profile` commands and `help` do not. Run `fbrcm profile switch <name>` to switch or create a profile.

Auth identities, project cache, parameter cache, and drafts are profile-scoped. Project cache stores known projects plus their selected `auth_id`. Default storage lives under user config/cache directories. Override roots with:

```text
FBRCM_CONFIG_DIR
FBRCM_CACHE_DIR
```

### Filter Queries

Flags named `--project` or `--filter` use mode-prefixed query strings:

```text
~query   fuzzy match; default if no prefix is given
^query   starts-with match
/query   includes match
=query   exact case-insensitive match
```

Project filters match project display name or project ID. Parameter filters match parameter key. `--project` and `--filter` may be repeated; repeated values are ORed and must be passed as separate flags.

### Positional Project Resolution

Commands that accept a positional `<project>` argument resolve it in this order:

1. Exact case-insensitive project ID.
2. Exact case-insensitive project display name.
3. Case-insensitive substring match against project ID or display name.

A single match is selected. Multiple exact-name or substring matches print only the ambiguous projects and return an error. No match prints the known-project table and returns an error. Exact ID always wins, including when another project's display name has the same text.

### Parameter Search

Parameter-context commands also support `--search <text>`. It searches parameter name, description, default value, conditional values, condition names, and condition expressions. Name/description/condition-name matching is case-insensitive and ignores punctuation; value/expression matching is case-sensitive. `--search` is ANDed with `--filter` and parameter-context `--expr`.

### Expression Filters

`--expr` uses expr-lang and must evaluate to boolean. See [EXPR.md](/Users/vic/Dev/pets/fbrcm/EXPR.md) for full context fields and helper functions.

Parameter-context commands:

```text
get
delete
update
project import
project versions diff
```

Project-context commands:

```text
projects list
projects update
add
```

### Stdin Remote Config Mode

`get`, `add`, `update`, and `delete` switch to stdin mode when stdin is piped. In stdin mode, command reads Firebase Remote Config JSON from stdin and writes modified JSON or query output to stdout. Remote Firebase writes are not performed. These commands also accept an fbrcm parameters cache JSON file and read its internal `remote_config` field.

`get` also accepts a directory passed as stdin. It reads top-level `.json` files from that directory, accepts raw Remote Config JSON or fbrcm cache JSON in each file, and treats each file as a project. Project ID is the file name without extension. Project name is built from that file name by splitting on `-` and `_`, then capitalizing words.

`project import` reads JSON from `--from`, stdin, or an interactive `.json` file picker. It accepts raw Remote Config JSON or an fbrcm parameters cache JSON file with `remote_config`.

## Commands

### `fbrcm`

With no arguments, opens TUI. With arguments, executes CLI command.

Flags:

```text
-h, --help      show root help
-v, --version   print version, commit, and build date
```

### `fbrcm add <parameter>`

Adds a new Remote Config parameter to every matched project. Parameter key is required and cannot be empty.

Exactly one value flag is required:

```text
--boolean true|false   value type BOOLEAN
--number <number>      value type NUMBER; must parse as float
--string <text>        value type STRING
--json <json>          value type JSON; must be valid JSON
```

Other flags:

```text
-p, --project <query>      filter target projects; may be repeated
--expr <expr>              filter target projects with project context
--dry-run                  validate/log Firebase write requests, do not publish
--description <text>       parameter description
--group <name>             add parameter inside group
```

Remote mode loads projects, filters them, adds parameter where it does not already exist, validates, and publishes. Existing parameters are skipped.

Stdin mode reads Remote Config JSON from stdin, adds parameter to that JSON, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field.

### `fbrcm get [parameter]`

Prints Remote Config parameters across projects.

Passing `[parameter]` is shorthand for `--filter =<parameter>`. It cannot be combined with `--filter`.

Flags:

```text
-p, --project <query>   filter projects; may be repeated
-f, --filter <query>    filter parameters; may be repeated
--expr <expr>           filter parameters with parameter context
--search <text>         search parameter names, descriptions, values, and conditions
--json                  print JSON rows
--all                   include projects with no matching parameters in table output
--update                revalidate cached parameters before printing
```

Default output is a terminal table. JSON output includes project, project ID, group, key, description, default value, conditionals, type, version, cache time, and status.

Stdin mode reads Remote Config JSON from stdin and queries only that config. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. If stdin is a directory, `get` reads top-level `.json` files and treats them as multiple projects.

### `fbrcm update [parameter]`

Updates matched Remote Config parameters. Passing `[parameter]` is shorthand for `--filter =<parameter>`. It cannot be combined with `--filter`.

Flags:

```text
-p, --project <query>      filter projects; may be repeated
-f, --filter <query>       filter parameters; may be repeated
--expr <expr>              filter parameters with parameter context
--search <text>            search parameter names, descriptions, values, and conditions
--dry-run                  validate/log Firebase write requests, do not publish
-y, --yes                  print diff and update without confirmation
--description <text>       set parameter description
--group <name>             move parameter into group
--no-group                 move parameter out of any group
--name <new-name>          rename parameter; cannot be empty
--remove-all-conditional-values
                           remove all conditional values from matched parameters
--remove-conditional-value <condition>
                           remove named conditional value from matched parameters; may be repeated
--boolean true|false       set BOOLEAN value
--number <number>          set NUMBER value
--string <text>            set STRING value
--json <json>              set JSON value
```

At most one value flag may be used. `--group` and `--no-group` are mutually exclusive. `--remove-all-conditional-values` and `--remove-conditional-value` are mutually exclusive.

Conditional value removal edits only `conditionalValues`; it keeps the parameter, default value, description, group, and all conditions themselves.

Remote mode prints diffs and prompts unless `--yes` is set. It validates and publishes with ETag conflict handling.

Stdin mode reads Remote Config JSON from stdin, updates matching parameters, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. It does not prompt.

### `fbrcm delete [parameter]`

Deletes matched Remote Config parameters. Passing `[parameter]` is shorthand for `--filter =<parameter>`. It cannot be combined with `--filter`.

Flags:

```text
-p, --project <query>   filter projects; may be repeated
-f, --filter <query>    filter parameters; may be repeated
--expr <expr>           filter parameters with parameter context
--search <text>         search parameter names, descriptions, values, and conditions
--dry-run               validate/log Firebase write requests, do not publish
-y, --yes               print diff and delete without confirmation
```

Remote mode prints diffs and prompts unless `--yes` is set. It validates and publishes with ETag conflict handling.

Stdin mode reads Remote Config JSON from stdin, deletes matching parameters, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. It does not prompt.

### `fbrcm project export <project>`

Exports one project's Remote Config JSON. `<project>` uses the shared positional project resolution described above.

Flags:

```text
--to <path>   write JSON to file; default prints JSON to stdout
```

Export normalizes JSON by unescaping `<`, `>`, `&`, trimming trailing line breaks, and ordering numeric conditional value keys before non-numeric keys.

### `fbrcm project import <project>`

Imports Remote Config JSON into one project. `<project>` resolves like `project export`.

Input source order:

```text
--from <path>
stdin
interactive .json picker
```

Import input may be raw Remote Config JSON or an fbrcm parameters cache JSON file with `remote_config`.

Flags:

```text
--from <path>                            read Remote Config JSON from file
--group <name>                           import only named group; may be repeated
-f, --filter <query>                     import only matching parameter keys; may be repeated
--expr <expr>                            import only parameters matching parameter context expression
--search <text>                          import only parameters matching rich search text
--dry-run                                validate/log Firebase write requests, do not publish
--remove-all-conditions                  remove all conditions and conditional values
--remove-project-specific-conditions     remove project-specific conditions and their usages
--merge                                  merge import into current config
--override                               replace current config with import
--merge-resolve current|import           auto-resolve merge conflicts
```

Mutual exclusions:

```text
--remove-all-conditions with --remove-project-specific-conditions
--merge with --override
```

`--merge-resolve` requires `--merge`. Valid values are `current` and `import`.

If current config is empty, import replaces it. If current config has content and neither `--merge` nor `--override` is set, command prompts for strategy. Merge adds missing conditions, groups, and parameters. Conflicting condition, group description, or parameter values prompt unless `--merge-resolve` is set.

After import transform, unused conditions, unknown condition references, empty groups, and version metadata are removed. Command validates, prints diff, asks for confirmation, then publishes.

### Remote Config version history

Version commands are scoped to one project and use the same project resolution as `project export`: project ID is matched first, followed by exact display name case-insensitively.

Firebase history and the local cache serve different purposes:

- Firebase history is authoritative for published-version metadata and native rollback availability.
- The local cache contains immutable templates that `fbrcm` has fetched or published. It may be incomplete, but it can retain a template after Firebase removes that version from its history.
- Firebase retains at most 300 versions. Inactive versions older than 90 days may be removed.
- Reading or caching a historical version does not change the current cache pointer.
- Successful publish, rollback, or restore creates and caches a new current version.

Version arguments accept a positive numeric version or a symbolic alias:

```text
142
current
latest
previous
current~2
latest~3
```

`current` and `latest` are equivalent. `previous` is shorthand for `current~1`. `current~N` and `latest~N` walk backward by `N` publications; they do not subtract `N` from the numeric version. For example, if history is `142, 140, 137`, then `current~2` resolves to version `137`.

In live mode, relative selectors walk authoritative Firebase history. With `--cached`, they walk locally cached version numbers below the cached current version; because local history may be incomplete, a cached relative selector is not guaranteed to identify the same publication as its live equivalent. Relative distance must be between 1 and 299. Commands fail clearly when the requested relative position is unavailable.

Commands always verify that an exact numeric version fetch returns the requested version; they never silently substitute another version.

### `fbrcm project versions list <project>`

Lists published Remote Config versions newest first. Live mode reads authoritative metadata from Firebase and marks locally cached versions. Cached mode performs no Firebase request and lists only local immutable snapshots.

Flags:

```text
--limit <n>          maximum versions to print; default 20; must be greater than zero
--all                retrieve every available version; mutually exclusive with an explicit --limit
--before <version>   newest version number to include
--since <RFC3339>    omit versions published before this time
--until <RFC3339>    omit versions published at or after this time
--cached             list local snapshots without contacting Firebase
--json               print structured JSON
```

Human live output includes version number, current marker, publication time, updating user, origin, update type, cached marker, and description. Cached output includes version, current marker, cache time, size, and any metadata stored in the template.

In cached mode, `--since` and `--until` apply to the local cache time because authoritative publication metadata may be unavailable.

JSON output is an object containing `project`, `versions`, and optional `next_page_token`. Each version includes Firebase metadata plus `current`, `cached`, and available local cache fields.

### `fbrcm project versions show <project> <version>`

Shows metadata for one exact version. Normal mode uses an existing immutable snapshot first and otherwise retrieves and caches the requested version from Firebase without moving the current pointer.

Flags:

```text
--cached   require the exact local snapshot and perform no Firebase request
--json     print structured metadata JSON
```

Use `project versions export` when the complete Remote Config JSON is needed.

### `fbrcm project versions diff <project> <from> [<to>]`

Compares two versions of the same project. Direction is always `<from> → <to>`. When `<to>` is omitted, it defaults to `current`.

Flags:

```text
-f, --filter <query>   include only matching parameter keys; may be repeated
--group <name>         include only parameters in named group; may be repeated
--expr <expr>          include only parameters matching parameter context expression
--search <text>        include only parameters matching rich search text
--parameters           include only parameter and group description differences
--conditions           include only condition differences
--cached               require both exact local snapshots and perform no Firebase requests
--json                 print structured diff JSON
```

`--parameters` and `--conditions` are mutually exclusive. Default output reuses the conditions, group descriptions, parameters, and summary diff format used by `projects diff`. JSON output contains `project`, `from_version`, `to_version`, and `diff`.

### `fbrcm project versions export <project> <version>`

Exports one historical Remote Config template. Retrieval is cache-first and never changes the current pointer.

Flags:

```text
--to <path>   write normalized JSON to a private file; default prints JSON to stdout
--cached      require the exact local snapshot and perform no Firebase request
```

Normalization matches `project export`.

### `fbrcm project versions rollback <project> <version>`

Uses Firebase's native rollback operation. It does not reactivate the old version number: Firebase force-publishes the selected historical template as a new version whose metadata records the rollback source.

Before publishing, the command:

1. Resolves the exact source and current versions.
2. Prints the complete `current → source` diff.
3. Explains that rollback creates a new version.
4. Asks for confirmation naming the canonical project ID.
5. Rechecks the current version immediately before rollback and stops if it changed during preview.

Flags:

```text
--dry-run   show the exact recovery diff without publishing
-y, --yes   skip final publish confirmation
--json      print a structured operation result
```

Rolling back to the current version is a no-op. A successful result reports the previous version, rollback source, and newly published version. Native Firebase rollback is a force update; the final recheck narrows but cannot eliminate the race window after that check.

If Firebase no longer retains a locally cached source version, rollback reports the failure and suggests the corresponding `restore` command.

### `fbrcm project versions restore <project> <version>`

Republishes an exact locally cached immutable snapshot. Restore exists for recovery when Firebase no longer retains the historical version.

Unlike rollback, restore:

- Requires the source version to be present locally.
- Publishes through the normal validated, ETag-protected update flow.
- Creates a normal new Remote Config version rather than Firebase rollback metadata.

It otherwise uses the same complete diff preview, confirmation, dry-run, current-version recheck, JSON contract, and success fields as rollback.

Flags:

```text
--dry-run   validate and preview the cached snapshot without publishing
-y, --yes   skip final publish confirmation
--json      print a structured operation result
```

Rollback and restore JSON results include `project_id`, `operation`, `previous_version`, `source_version`, `published_version`, `dry_run`, and `changed`. Human previews are written separately from JSON data so stdout remains machine-readable.

### `fbrcm projects list`

Lists projects using cache-first loading.

Flags:

```text
-f, --filter <query>   filter projects; may be repeated
--expr <expr>          filter projects with project context
--json                 print projects as JSON
--update               sync projects from Firebase before printing
--url                  include Firebase Console Remote Config URL
```

### `fbrcm projects update`

Syncs projects from Firebase into cache, then prints them.

Flags:

```text
-f, --filter <query>   filter projects after sync; may be repeated
--expr <expr>          filter projects with project context
--json                 print projects as JSON
--url                  include Firebase Console Remote Config URL
--auth <auth-id>       sync projects for one auth identity
```

### `fbrcm projects diff <source-project> <target-project>`

Compares Remote Config between two projects. `<source-project>` is the desired config and `<target-project>` is the config being checked for drift. Both arguments use shared positional project resolution.

By default, command fetches live Remote Config for both projects. Use `--cached` to compare local parameter cache entries instead.

Flags:

```text
-f, --filter <query>   include only matching parameter keys; may be repeated
--group <name>         include only parameters in named group; may be repeated
--expr <expr>          include only parameters matching parameter context expression
--search <text>        include only parameters matching rich search text
--parameters           include only parameter and group description differences
--conditions           include only condition differences
--cached               compare cached Remote Config snapshots
--json                 print structured diff JSON
```

Default output is a terminal diff grouped by conditions, group descriptions, and parameters. JSON output includes source project, target project, summary counts, and structured change records.

### `fbrcm projects promote <source-project> <target-project>`

Promotes selected Remote Config changes from source project to target project. `<source-project>` is the desired config. `<target-project>` is the project that may be published.

By default in an interactive terminal, command reviews eligible changes item by item before publishing. V1 selection is whole-item based: parameter slots, conditions, and group descriptions. Parameter selection automatically includes required condition definitions and group descriptions when needed.

Default promotion includes source additions and source updates. Target-only removals are ignored unless `--prune` is set.

Flags:

```text
-f, --filter <query>   promote only matching parameter keys; may be repeated
--group <name>         promote only parameters in named group; may be repeated
--expr <expr>          promote only parameters matching parameter context expression
--search <text>        promote only parameters matching rich search text
--parameters           promote only parameter and group description changes
--conditions           promote only condition changes
--interactive          review each promotion item interactively
--all                  select all eligible changes without per-item prompts
--prune                include target-only removals
--dry-run              validate/log Firebase write requests, do not publish
-y, --yes              skip final publish confirmation
--json                 print promotion result JSON
```

Non-interactive promote requires explicit selection intent: `--all`, `--filter`, `--group`, `--expr`, or `--search`. Command reloads the target before publishing, validates with Firebase, publishes using the latest target ETag, and retries if the target changes during promotion.

### `fbrcm projects path`

Prints projects config file path.

Flags:

```text
--json   print {"path": "..."}
```

### `fbrcm projects purge`

Deletes cached projects config file.

Flags:

```text
-y, --yes   skip confirmation
```

### `fbrcm cache list`

Lists immutable cached Remote Config versions and mutable draft files.

Flags:

```text
--json   print cache entries as JSON
```

JSON entries include project ID, project name, version, file size, cached time, draft flag, and path.

### `fbrcm cache path`

Prints Remote Config cache directory path.

Flags:

```text
--json   print {"path": "..."}
```

### `fbrcm cache purge`

Deletes all locally cached immutable Remote Config versions. The confirmation reports snapshot count, total size, and project count, and warns that versions no longer retained by Firebase may be permanently lost. If drafts exist, prompts separately before deleting drafts.

Flags:

```text
-y, --yes   skip confirmations and delete both caches and drafts
```

### `fbrcm config path`

Prints global config file path.

Flags:

```text
--json   print {"path": "..."}
```

### `fbrcm auth list`

Lists configured auth identities.

Flags:

```text
--json   print auth identities as JSON
```

### `fbrcm auth add oauth <auth-id>`

Adds or replaces an OAuth identity and imports its desktop client secret JSON.

Input source order:

```text
--from <path>
stdin
interactive .json file picker
```

Flags:

```text
--from <path>      import client secret from file
--label <label>    auth identity label
```

### `fbrcm auth add service-account <auth-id>`

Adds or replaces a service account identity and imports its JSON key.

Input source order:

```text
--from <path>
stdin
interactive .json file picker
```

Flags:

```text
--from <path>      import service account key from file
--label <label>    auth identity label
```

### `fbrcm auth add gcloud <auth-id>`

Adds or replaces a gcloud ADC identity. Run `gcloud auth application-default login` first so ADC discovery can find credentials.

Flags:

```text
--label <label>    auth identity label
```

### `fbrcm auth login <auth-id>`

Authenticates or validates an auth identity. OAuth starts browser login when needed; service-account validates the key; gcloud validates ADC discovery.

Flags:

```text
--noopen   do not open browser automatically; print URL instead
```

### `fbrcm auth path <auth-id>`

Prints auth file paths.

Flags:

```text
--json   print paths as JSON
```

### `fbrcm auth purge <auth-id>`

Deletes an auth identity and its client secret/token files.

Flags:

```text
-y, --yes   skip confirmation dialogs
```

### `fbrcm auth bind <project-query>`

Binds matching cached projects to an auth identity.

Flags:

```text
--auth <auth-id>   auth identity to bind
```

### `fbrcm profile`

Prints active profile name.

### `fbrcm profile list`

Lists profiles and marks active profile.

Flags:

```text
--json   print [{"profile": "...", "active": true|false}, ...]
```

### `fbrcm profile switch <name>`

Switches to profile, creating it if needed.

### `fbrcm profile rename <old-name> <new-name>`

Renames existing profile.

### `fbrcm profile path <profile>`

Prints profile config and cache directory paths.

Flags:

```text
--json   print [{"path": "..."}, {"path": "..."}]
```

### `fbrcm profile purge <profile>`

Deletes profile config and cache directories. Confirmation defaults to yes. Active profile cannot be purged.

Flags:

```text
-y, --yes   skip confirmation
```

### `fbrcm completion`

Generates shell completion scripts.

Commands:

```text
fbrcm completion bash
fbrcm completion fish
fbrcm completion powershell
fbrcm completion zsh
```

Each shell command supports:

```text
--no-descriptions   disable completion descriptions
```

Examples:

```sh
source <(fbrcm completion bash)
source <(fbrcm completion zsh)
fbrcm completion fish | source
fbrcm completion powershell | Out-String | Invoke-Expression
```

### `fbrcm help [command]`

Shows help for command path.

Examples:

```sh
fbrcm help project import
fbrcm get --help
```
