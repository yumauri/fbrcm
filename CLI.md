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
│   └── import <project>
│       ├── --from <path>
│       ├── --group <name>        repeated
│       ├── --filter, -f <query>  repeated
│       ├── --expr <expr>
│       ├── --search <text>
│       ├── --dry-run
│       ├── --remove-all-conditions
│       ├── --remove-project-specific-conditions
│       ├── --merge
│       ├── --override
│       └── --merge-resolve current|import
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

Exports one project's Remote Config JSON. `<project>` matches project ID first, then exact display name case-insensitively. Ambiguous or missing names print matching project table.

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

Lists cached Remote Config snapshots and draft files.

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

Deletes cached Remote Config snapshots. If drafts exist, prompts separately before deleting drafts.

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
