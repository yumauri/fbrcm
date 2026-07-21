# fbrcm CLI

`fbrcm` is a Firebase Remote Config manager. It runs as an interactive TUI when called with no arguments. Any argument switches to CLI mode.

## Command Tree

```text
fbrcm [--help] [--version] [--profile <name>]
│
├── add <parameter>
│   ├── --project, -p <query>  repeated
│   ├── --expr <expr>
│   ├── --dry-run
│   ├── --draft
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
│   └── clear [--yes|-y]
│
├── config
│   ├── path [--json]
│   ├── show [key] [--json]
│   ├── set <key> <value>... [--json]
│   ├── reset [key] [--yes|-y] [--json]
│   ├── validate [--json]
│   └── edit [--editor <command>]
│
├── completion
│   ├── bash [--no-descriptions]
│   ├── fish [--no-descriptions]
│   ├── powershell [--no-descriptions]
│   └── zsh [--no-descriptions]
│
├── conditions
│   ├── list <project>
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --search <text>
│   │   ├── --expr <expr>
│   │   ├── --update
│   │   └── --json
│   ├── show <project> <condition>
│   │   ├── --update
│   │   └── --json
│   ├── add <project> <name>
│   │   ├── --expression <expr>  required
│   │   ├── --color <color>
│   │   ├── --priority <n>
│   │   ├── --dry-run
│   │   ├── --draft
│   │   └── --yes, -y
│   ├── edit <project> <condition>
│   │   ├── --expression <expr>
│   │   ├── --color <color>
│   │   ├── --no-color
│   │   ├── --dry-run
│   │   ├── --draft
│   │   └── --yes, -y
│   ├── rename <project> <condition> <new-name>
│   │   ├── --dry-run
│   │   ├── --draft
│   │   └── --yes, -y
│   ├── move <project> <condition> <priority>
│   │   ├── --dry-run
│   │   ├── --draft
│   │   └── --yes, -y
│   ├── delete <project> <condition>
│   │   ├── --dry-run
│   │   ├── --draft
│   │   └── --yes, -y
│   └── validate <project> [--json]
│
├── delete [parameter]
│   ├── --project, -p <query>  repeated
│   ├── --filter, -f <query>   repeated
│   ├── --expr <expr>
│   ├── --search <text>
│   ├── --dry-run
│   ├── --draft
│   └── --yes, -y
│
├── doctor [--json] [--timeout <duration>]
│
├── draft
│   ├── list
│   │   ├── --filter, -f <query>  repeated
│   │   └── --json
│   ├── path [--json]
│   ├── show <project>
│   │   ├── --raw
│   │   └── --to <path>
│   ├── diff <project>
│   │   ├── --against base|current
│   │   ├── --cached
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --group <name>        repeated
│   │   ├── --expr <expr>
│   │   ├── --search <text>
│   │   ├── --parameters
│   │   ├── --conditions
│   │   ├── --json
│   │   └── --exit-code
│   ├── publish [project...]
│   │   ├── --all
│   │   ├── --dry-run
│   │   ├── --yes, -y
│   │   └── --json
│   └── discard [project...]
│       ├── --all
│       ├── --yes, -y
│       └── --json
│
├── duplicate <source> <target>
│   ├── --project, -p <query>  repeated
│   ├── --expr <expr>
│   ├── --dry-run
│   ├── --draft
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
├── groups
│   ├── list
│   │   ├── --project, -p <query> repeated
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --search <text>
│   │   ├── --update
│   │   └── --json
│   ├── add <name> [--project|-p <query>] [--description <text>] [--dry-run] [--draft] [--yes|-y]
│   ├── edit <group> [--project|-p <query>] (--description <text>|--no-description) [--dry-run] [--draft] [--yes|-y]
│   ├── rename <group> <new-name> [--project|-p <query>] [--dry-run] [--draft] [--yes|-y]
│   └── delete <group> [--project|-p <query>] [--dry-run] [--draft] [--yes|-y]
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
│   ├── delete <auth-id> [--yes|-y]
│   └── bind --auth <auth-id> [--project|-p <query>]...
│
├── profile
│   ├── list [--json]
│   ├── path <profile> [--json]
│   ├── delete <profile> [--yes|-y]
│   ├── rename <old-name> <new-name>
│   └── switch <name>
│
├── project
│   ├── show <project> [--update] [--json]
│   ├── open <project>
│   ├── defaults <project> [--format json|xml|plist] [--to <path>] [--yes|-y]
│   ├── export <project> [--to <path>] [--yes|-y]
│   └── import <project>
│       ├── --from <path>
│       ├── --group <name>        repeated
│       ├── --filter, -f <query>  repeated
│       ├── --expr <expr>
│       ├── --search <text>
│       ├── --dry-run
│       ├── --draft
│       ├── --remove-all-conditions
│       ├── --keep-portable-conditions-only
│       ├── --merge
│       ├── --override
│       ├── --merge-resolve current|import
│       ├── --yes, -y
│       └── --json
│
├── versions
│   ├── list <project>
│   │   ├── --limit <n>
│   │   ├── --all
│   │   ├── --before <version>
│   │   ├── --since <RFC3339>
│   │   ├── --until <RFC3339>
│   │   ├── --cached
│   │   └── --json
│   ├── show <project> <version>
│   │   ├── --cached
│   │   └── --json
│   ├── diff <project> <from> [<to>]
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --group <name>        repeated
│   │   ├── --expr <expr>
│   │   ├── --search <text>
│   │   ├── --parameters
│   │   ├── --conditions
│   │   ├── --cached
│   │   ├── --json
│   │   └── --exit-code
│   ├── export <project> <version>
│   │   ├── --to <path>
│   │   ├── --cached
│   │   └── --yes, -y
│   ├── rollback <project> <version>
│   │   ├── --dry-run
│   │   ├── --yes, -y
│   │   └── --json
│   └── restore <project> <version>
│       ├── --dry-run
│       ├── --yes, -y
│       └── --json
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
│   ├── forget
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --expr <expr>
│   │   └── --yes, -y
│   ├── diff <source-project> <target-project>
│   │   ├── --filter, -f <query>  repeated
│   │   ├── --group <name>        repeated
│   │   ├── --expr <expr>
│   │   ├── --search <text>
│   │   ├── --parameters
│   │   ├── --conditions
│   │   ├── --cached
│   │   ├── --json
│   │   └── --exit-code
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
│   └── reset [--yes|-y]
│
└── update [parameter]
    ├── --project, -p <query>  repeated
    ├── --filter, -f <query>   repeated
    ├── --expr <expr>
    ├── --search <text>
    ├── --dry-run
    ├── --draft
    ├── --yes, -y
    ├── --description <text>
    ├── --group <name>
    ├── --no-group
    ├── --name <new-name>
    ├── --condition <name>
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

When `FBRCM_OFFLINE` is unset, fbrcm performs a proxy-aware HTTPS connectivity probe before executing a network-capable CLI command and automatically enables offline mode if the probe fails. Help, version, and all `config` commands skip this probe. The probe and other standard HTTP requests honor `HTTPS_PROXY`, `HTTP_PROXY`, and `NO_PROXY`, including their lowercase forms. Defining `FBRCM_OFFLINE`, including with an empty value or `0`, enables offline mode without probing.

Most commands require a selected profile. `profile`, `config`, `doctor`, and `help` do not require profile initialization. Run `fbrcm profile switch <name>` to switch or create a profile. Use the root `--profile <name>` flag or `FBRCM_PROFILE` to select an existing profile for one process without changing the persisted active profile; the flag takes precedence over the environment variable. Config commands ignore the per-process profile override because they manage the global file, including its persisted active-profile field.

Interactive yes/no confirmations select **Yes** by default. Use the arrow keys to select No, or pass `--yes` where available to skip the prompt.

Human-readable collection output always renders its normal table, including when there are no rows. An empty result therefore contains the table headers rather than a special empty-state message. A command whose primary JSON result is a collection always emits a top-level array and uses `[]` when empty. Singular resources and single-operation reports remain JSON objects.

Auth identities, project cache, parameter cache, and drafts are profile-scoped. Project cache stores known projects plus their selected `auth_id`. Default storage lives under user config/cache directories. Override roots with:

```text
FBRCM_CONFIG_DIR
FBRCM_CACHE_DIR
FBRCM_PROFILE
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

Draft commands resolve only locally stored drafts and never synchronize projects as a side effect. An exact case-insensitive draft project ID wins; otherwise the query must uniquely match the locally known project ID or display name. This also permits `show --raw` and `discard` for drafts whose project is no longer present in the projects cache.

### Parameter Search

Parameter-context commands also support `--search <text>`. It searches parameter name, description, default value, conditional values, condition names, and condition expressions. Name/description/condition-name matching is case-insensitive and ignores punctuation; value/expression matching is case-sensitive. `--search` is ANDed with `--filter` and parameter-context `--expr`.

### Expression Filters

`--expr` uses expr-lang and must evaluate to boolean. See [EXPR.md](/Users/vic/Dev/pets/fbrcm/EXPR.md) for full context fields and helper functions.

Parameter-context commands:

```text
get
delete
update
draft diff
project import
versions diff
```

Condition-context commands:

```text
conditions list
```

Project-context commands:

```text
projects list
projects update
projects forget
add
duplicate
```

### Stdin Remote Config Mode

`get`, `add`, `update`, and `delete` switch to stdin mode when stdin is piped. In stdin mode, command reads Firebase Remote Config JSON from stdin and writes modified JSON or query output to stdout. Remote Firebase writes are not performed. These commands also accept an fbrcm parameters cache JSON file and read its internal `remote_config` field.

`get` also accepts a directory passed as stdin. It reads top-level `.json` files from that directory, accepts raw Remote Config JSON or fbrcm cache JSON in each file, and treats each file as a project. Project ID is the file name without extension. Project name is built from that file name by splitting on `-` and `_`, then capitalizing words.

`project import` reads JSON from `--from`, stdin, or an interactive `.json` file picker. It accepts raw Remote Config JSON or an fbrcm parameters cache JSON file with `remote_config`.

`--draft` is unavailable in stdin transformation mode because piped input has no persistent target project identity.

### Draft lifecycle and write safety

Drafts are profile-scoped, self-contained records. Each record stores the working Remote Config, its immutable base Remote Config, base version and ETag, timestamps, and a draft format version. Plain Remote Config JSON is not accepted as an on-disk draft format, and no legacy draft migration or fallback is performed.

`add`, `duplicate`, `update`, `delete`, `project import`, and the condition mutation commands accept `--draft`. In draft mode they apply changes on top of an existing project draft or create a new draft from freshly revalidated Remote Config. They do not validate or publish to Firebase. Combining `--draft` with `--dry-run` previews the change without writing either draft or Firebase state.

Immediate Remote Config writes refuse to proceed when the target has an unpublished draft. This guard applies to add, duplicate, update, delete, condition mutations, project import, version rollback/restore, and project promotion. Resolve the draft with `draft publish` or `draft discard`, or add the intended mutation to it with `--draft`.

Draft publish always fetches current Firebase state, performs a three-way merge from base, draft, and current, validates using the current ETag, and publishes only the exact candidate that was previewed. Conflicts preserve the local draft. Successfully published or already-applied drafts are removed locally. A publish that succeeds remotely but cannot remove its local record reports `published-cleanup-failed`; rerunning recognizes the already-applied content and retries cleanup without creating another version.

## Commands

### `fbrcm`

With no arguments, opens TUI. With arguments, executes CLI command.

Flags:

```text
-h, --help      show root help
-v, --version   print version, commit, and build date
    --profile   use an existing profile for this invocation without changing the active profile
```

`--profile` defaults from `FBRCM_PROFILE`. It applies to every CLI subcommand. `FBRCM_PROFILE` also selects and pins the profile when starting the TUI with no arguments; restart without it to create or switch profiles interactively.

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
--dry-run                  preview without writing local or Firebase state
--draft                    save changes to local drafts instead of publishing
--description <text>       parameter description
--group <name>             add parameter inside group
```

Remote mode loads projects, filters them, adds parameter where it does not already exist, validates, and publishes. Existing parameters are skipped.

With `--draft`, the same mutation is stored locally on top of any existing draft. Without `--draft`, the command refuses projects that already have unpublished drafts.

Stdin mode reads Remote Config JSON from stdin, adds parameter to that JSON, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field.

### `fbrcm duplicate <source> <target>`

Duplicates one complete Remote Config parameter in every matched project. The copy preserves the source group, description, value type, default value, conditional values, and condition references. Source lookup is exact and case-insensitive. A project without the source parameter is skipped; an ambiguous source or an existing target name is an error. Target collision checks are also case-insensitive and never overwrite an existing parameter.

Flags:

```text
-p, --project <query>   filter target projects; may be repeated
--expr <expr>           filter target projects with project context
--dry-run               preview without writing local or Firebase state
--draft                 save changes to local drafts instead of publishing
-y, --yes               print diff and duplicate without confirmation
```

Remote mode prints the complete Remote Config diff and prompts before each duplicate unless `--yes` is set. It validates and publishes with ETag conflict handling. Project filters are applied before the project-context expression, matching `add` behavior.

With `--draft`, duplication composes onto each existing draft and remains local. Without `--draft`, all selected projects are checked for drafts before the first publish. The command does not use stdin transformation mode.

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
--dry-run                  preview without writing local or Firebase state
--draft                    save changes to local drafts instead of publishing
-y, --yes                  print diff and update without confirmation
--description <text>       set parameter description
--group <name>             move parameter into group
--no-group                 move parameter out of any group
--name <new-name>          rename parameter; cannot be empty
--condition <name>         assign the value flag to this condition instead of the default value
--remove-all-conditional-values
                           remove all conditional values from matched parameters
--remove-conditional-value <condition>
                           remove named conditional value from matched parameters; may be repeated
--boolean true|false       set BOOLEAN value
--number <number>          set NUMBER value
--string <text>            set STRING value
--json <json>              set JSON value
```

At most one value flag may be used. `--condition` requires a value flag and resolves the condition by exact name, then exact case-insensitive name. It preserves the default and all other conditional values while assigning the selected typed value. `--group` and `--no-group` are mutually exclusive. `--condition`, `--remove-all-conditional-values`, and `--remove-conditional-value` are mutually exclusive.

Conditional value assignment and removal edit only `conditionalValues`; they keep the parameter, default value, description, group, and all conditions themselves.

Remote mode prints diffs and prompts unless `--yes` is set. It validates and publishes with ETag conflict handling.

With `--draft`, mutations compose onto each existing draft and remain local. Without `--draft`, all selected projects are checked for drafts before the first publish, preventing a partially published batch.

Stdin mode reads Remote Config JSON from stdin, updates matching parameters, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. It does not prompt.

### `fbrcm delete [parameter]`

Deletes matched Remote Config parameters. Passing `[parameter]` is shorthand for `--filter =<parameter>`. It cannot be combined with `--filter`.

Flags:

```text
-p, --project <query>   filter projects; may be repeated
-f, --filter <query>    filter parameters; may be repeated
--expr <expr>           filter parameters with parameter context
--search <text>         search parameter names, descriptions, values, and conditions
--dry-run               preview without writing local or Firebase state
--draft                 save changes to local drafts instead of publishing
-y, --yes               print diff and delete without confirmation
```

Remote mode prints diffs and prompts unless `--yes` is set. It validates and publishes with ETag conflict handling.

With `--draft`, deletions are saved locally on top of any existing draft. Without `--draft`, all selected projects are checked for drafts before the first publish.

Stdin mode reads Remote Config JSON from stdin, deletes matching parameters, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. It does not prompt.

### `fbrcm conditions list <project>`

Lists condition definitions in Firebase evaluation-priority order. The command uses an unpublished draft when one exists; otherwise it reads the parameter cache. If a normal cache read fails but a stale cache exists, it prints that stale snapshot rather than discarding usable condition data.

Flags:

```text
-f, --filter <query>   filter condition names; may be repeated
--search <text>        case-insensitive substring search across name and expression
--expr <expr>          filter using condition expression context
--update               revalidate cached Remote Config before printing
--json                 print structured JSON
```

Condition filters use the shared mode prefixes described under Filter Queries. Repeated filters are ORed. `--filter`, `--search`, and `--expr` are ANDed together. See [EXPR.md](/Users/vic/Dev/pets/fbrcm/EXPR.md) for condition context fields and examples.

Human output prints project/version/source context followed by a terminal-width-aware table containing priority, color-styled name, usage count, and expression. Long expressions are cropped with an ellipsis. JSON output is a plain array of condition objects without repeated project/version/source context.

### `fbrcm conditions show <project> <condition>`

Shows one condition and every parameter value that uses it. Condition lookup first uses the exact name, then an exact case-insensitive name.

Flags:

```text
--update   revalidate cached Remote Config before printing
--json     print structured JSON
```

Human output includes priority, color-styled name and color, expression, a pluralized usage count, and a typed-value table. JSON output includes project/version/source context and the complete condition usage model.

### Condition mutations

The following commands edit one project's complete Remote Config:

```text
fbrcm conditions add <project> <name> --expression <expr>
fbrcm conditions edit <project> <condition>
fbrcm conditions rename <project> <condition> <new-name>
fbrcm conditions move <project> <condition> <priority>
fbrcm conditions delete <project> <condition>
```

All five commands support:

```text
--dry-run   preview without writing local or Firebase state
--draft     save changes to a local draft instead of publishing
-y, --yes   print the diff and apply without confirmation
```

Without `--draft`, mutations print the complete Remote Config diff, ask for confirmation unless `--yes` is set, validate with Firebase, and publish with ETag protection. They refuse immediate publication while the project has an unpublished draft. With `--draft`, mutations compose onto the existing draft or create one and remain local.

`add` appends the condition by default. Its additional flags are:

```text
--expression <expr>   raw Firebase condition expression; required
--color <color>       Firebase display color
--priority <n>        evaluation priority; zero/default appends last
```

`edit` requires at least one of:

```text
--expression <expr>   replace the raw Firebase condition expression
--color <color>       replace the Firebase display color
--no-color            remove the display color
```

`--color` and `--no-color` are mutually exclusive. Supported colors are `BLUE`, `BROWN`, `CYAN`, `DEEP_ORANGE`, `GREEN`, `INDIGO`, `LIME`, `ORANGE`, `PINK`, `PURPLE`, and `TEAL`; input is normalized case-insensitively. Imported condition objects accept only Firebase's `name`, `expression`, and `tagColor` fields; unsupported fields are rejected.

`rename` updates the condition definition and every conditional-value reference to it. `move` inserts the complete condition at the requested 1-based priority and reports how many conditions and parameters may be affected by the priority change. `delete` removes the condition and its conditional values; parameters left without any value may also be removed, and the command reports that impact before confirmation.

### `fbrcm conditions validate <project>`

Validates the effective condition configuration with Firebase without publishing it. If the project has a draft, validation prepares the same merged candidate used by draft publication; otherwise it revalidates the published Remote Config.

Flags:

```text
--json   print project, source, and validity as JSON
```

Human output identifies the project and whether the validated source was `draft` or `firebase`.

### `fbrcm groups list`

`groups list` lists real Firebase parameter groups across the selected projects, including intentionally empty and description-only groups. It uses an unpublished draft when present and otherwise follows the same fresh/stale cache behavior as condition reads. Human output is a naturally sized table with project ID, name, parameter count, and description; the project column is omitted for one exact `--project` filter, matching `get`. On narrow terminals, the description is cropped with an ellipsis first, followed by project ID and group name only when necessary.

List flags:

```text
-p, --project <query>  filter projects by name or ID; may be repeated
-f, --filter <query>   filter group names; may be repeated
--search <text>        search group names and descriptions
--update               revalidate cached Remote Config before printing
--json                 print structured JSON
```

### Group mutations

```text
fbrcm groups add <name> [--project|-p <query>] [--description <text>]
fbrcm groups edit <group> [--project|-p <query>] (--description <text>|--no-description)
fbrcm groups rename <group> <new-name> [--project|-p <query>]
fbrcm groups delete <group> [--project|-p <query>]
```

`add` creates a group entry even when it has no parameters or description. `edit` replaces or explicitly clears its description while preserving its parameters. `rename` preserves both parameters and description. `delete` is an explicit group-level operation and removes the group together with all parameters it contains.

All group commands support repeatable `--project|-p` filters with the same mode prefixes and OR behavior as `get`, `add`, `delete`, and `update`. With no project filter, they process every configured project in stable project-name/ID order. Named mutations skip projects that do not contain the group; `add` skips projects where it already exists.

All group mutations also support `--dry-run`, `--draft`, and `--yes|-y`, with the same diff, confirmation, validation, ETag, draft-composition, and draft-conflict behavior as condition mutations. `--description` and `--no-description` are mutually exclusive.

### `fbrcm draft list`

Lists drafts in the active profile without contacting Firebase. Invalid draft envelopes remain visible instead of failing the complete listing.

Flags:

```text
-f, --filter <query>   filter by project ID or cached display name; may be repeated
--json                 print structured JSON
```

Human output includes project ID, project name, base version, update time, parameter/condition change counts, and status. Status is `ready`, `unchanged`, or `invalid`.

JSON entries include `project_id`, `project`, `base_version`, `created_at`, `updated_at`, byte size, status, validity, base availability, path, and change counts.

### `fbrcm draft path`

Prints the directory containing Remote Config draft files for the active profile.

Flags:

```text
--json   print {"path": "..."}
```

### `fbrcm draft show <project>`

Prints one draft for recovery or export. Default output is the validated working Remote Config only, normalized like project export and without status text.

Flags:

```text
--raw         print the exact stored draft envelope, including its immutable base
--to <path>   write output to a private file instead of stdout
```

`--raw` bypasses draft decoding, so it can recover an invalid or damaged envelope. File output is forced to mode `0600`.

### `fbrcm draft diff <project>`

Shows either the local draft intent or the effective publish preview.

Flags:

```text
--against base|current   comparison target; default base
--cached                 with current, use the latest local snapshot and do not contact Firebase
-f, --filter <query>     include only matching parameter keys; may be repeated
--group <name>           include only parameters in named group; may be repeated
--expr <expr>            filter parameter changes with parameter context
--search <text>          filter changed parameters with rich search
--parameters             include only parameters and group descriptions
--conditions             include only conditions
--json                   print structured diff JSON
--exit-code              return 1 for differences and 2 for errors
```

`--against base` compares immutable base to stored draft and is entirely local. `--against current` fetches current Firebase state, performs the same three-way merge used by publish, and compares current to the effective candidate. `--cached` makes that second operation local but does not claim the cached snapshot is still current.

`--parameters` and `--conditions` are mutually exclusive. Condition ordering changes are included in human and JSON diffs.

Without `--exit-code`, both differences and no differences return success. With it, exit statuses follow diff conventions: `0` no differences, `1` differences, `2` any comparison, invocation, profile, or output error. The status describes the filtered result when selection flags are present.

### `fbrcm draft publish [project...]`

Safely rebases and publishes one or more drafts. Project arguments may be repeated. Use `--all` instead to process every draft in the active profile; `--all` and positional projects are mutually exclusive.

Flags:

```text
--all          publish every active-profile draft
--dry-run      fetch, merge, validate, and preview without publishing or deleting drafts
-y, --yes      skip publish confirmations
--json         print structured results
```

For each project, the command fetches current Firebase state, merges local intent onto it, displays `current → candidate`, and asks for confirmation. It then validates and publishes that candidate with the fetched ETag. A remote change after preview is rejected by ETag protection rather than silently producing a different candidate. Conflicts and validation failures preserve the draft.

If current Firebase state already contains the effective draft changes, no new version is created and the draft is removed as `already-applied`. Batch mode continues after independent project failures and returns nonzero if any item failed.

JSON output is an array of results. Each result includes project ID, status, base/previous/published versions, `rebased`, `changed`, `draft_deleted`, `dry_run`, and an optional error. Status values include `published`, `would-publish`, `already-applied`, `canceled`, `failed`, and `published-cleanup-failed`. Prompts and human diffs are kept off JSON stdout.

### `fbrcm draft discard [project...]`

Deletes one or more local drafts without contacting Firebase. Use `--all` instead of positional projects to process the complete active profile.

Flags:

```text
--all          discard every active-profile draft
-y, --yes      skip destructive confirmations
--json         print structured results
```

Human mode prints the local `base → draft` diff before confirmation. Invalid drafts warn that preview is unavailable but can still be explicitly discarded. Naming a nonexistent draft is an error; `--all` with no drafts is a successful no-op.

JSON output is an array containing one status result per selected project.

### `fbrcm project show <project>`

Shows cached project metadata, the selected auth identity used for project operations, and every configured auth identity that discovered the project during synchronization. `<project>` uses the shared positional project resolution described above.

Flags:

```text
--update   synchronize projects with every configured auth identity before printing
--json     print the project using the same JSON contract as `projects list --json`, including its Firebase Console URL
```

Without `--update`, auth access reflects the latest cached project synchronization.

The status is `disabled` when the latest project synchronization could not find the project through its assigned auth identity or any replacement identity. Disabled projects remain cached and visible, but live Firebase operations are blocked until a later update rediscovers and automatically rebinds them.

### `fbrcm project open <project>`

Opens the project's Remote Config page in the Firebase console. `<project>` uses the shared positional project resolution described above.

### `fbrcm project export <project>`

Exports one project's Remote Config JSON. `<project>` uses the shared positional project resolution described above.

Flags:

```text
--to <path>   write JSON to file; default prints JSON to stdout
--yes, -y     overwrite an existing destination without confirmation
```

Export normalizes JSON by unescaping `<`, `>`, `&`, trimming trailing line breaks, and ordering numeric conditional value keys before non-numeric keys. When `--to` names an existing file, export asks before replacing it unless `--yes` is set. A destination created after the initial check is not overwritten without authorization.

### `fbrcm project defaults <project>`

Downloads the active client-side Remote Config parameter defaults directly from Firebase. `<project>` uses the shared positional project resolution described above. JSON is suitable for web applications, XML for Android, and plist for Apple applications.

Flags:

```text
--format json|xml|plist   output format; default json; case-insensitive
--to <path>               write defaults to a private file; default prints the raw response to stdout
--yes, -y                 overwrite an existing destination without confirmation
```

The downloaded payload contains parameter keys and their backend default values rather than the complete Remote Config template. Output bytes are preserved as returned by Firebase. When `--to` names an existing file, the command asks before downloading or replacing it unless `--yes` is set. A destination created after the initial check is not overwritten without authorization.

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
--dry-run                                preview without writing local or Firebase state
--draft                                  save the import as a local draft
--remove-all-conditions                  remove all conditions and conditional values
--keep-portable-conditions-only          keep portable conditions and remove destination-specific usages
--merge                                  merge import into current config
--override                               replace current config with import
--merge-resolve current|import           auto-resolve merge conflicts
--yes, -y                                skip final import confirmation
--json                                   print the import result as JSON
```

Mutual exclusions:

```text
--remove-all-conditions with --keep-portable-conditions-only
--merge with --override
```

`--merge-resolve` requires `--merge`. Valid values are `current` and `import`.

If current config is empty, import replaces it. If current config has content and neither `--merge` nor `--override` is set, command prompts for strategy. Merge adds missing conditions, groups, and parameters. Conflicting condition, group description, or parameter values prompt unless `--merge-resolve` is set. `--yes` skips only the final confirmation; automated imports should also specify `--merge` or `--override` and, when needed, `--merge-resolve`.

After import transform, the CLI reports how many source conditions are kept and removed. `--keep-portable-conditions-only` removes conditions tied to destination-specific resources such as Analytics audiences or user properties, experiments, Firebase App IDs, custom signals, and installation IDs. Unused conditions and unknown condition references are also removed. Groups that become empty are preserved, including their descriptions; only an explicit group-level selection or replacement removes a group. Normal mode removes version metadata, validates, prints a diff, asks for confirmation, and publishes. Draft mode retains the working version identity, prints the same diff and confirmation, then saves locally without Firebase validation or publication.

JSON output is one object containing `project_id`, `status`, `changed`, `draft`, and `dry_run`. Status is `imported`, `would-import`, `drafted`, `would-draft`, `unchanged`, or `canceled`. JSON mode suppresses human condition summaries and diffs but does not imply `--yes` or choose an import strategy.

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

### `fbrcm versions list <project>`

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

JSON output is a plain array of version objects without project or pagination metadata. Each element includes Firebase metadata, `current`, `cached`, and available local cache fields.

### `fbrcm versions show <project> <version>`

Shows metadata for one exact version. Normal mode uses an existing immutable snapshot first and otherwise retrieves and caches the requested version from Firebase without moving the current pointer.

Flags:

```text
--cached   require the exact local snapshot and perform no Firebase request
--json     print structured metadata JSON
```

Use `versions export` when the complete Remote Config JSON is needed.

### `fbrcm versions diff <project> <from> [<to>]`

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
--exit-code            return 1 for differences and 2 for errors
```

`--parameters` and `--conditions` are mutually exclusive. Default output reuses the conditions, group descriptions, parameters, and summary diff format used by `projects diff`. JSON output contains `project`, `from_version`, `to_version`, `changed`, and `diff`.

Without `--exit-code`, both differences and no differences return success. With it, exit statuses are `0` for no differences, `1` for differences, and `2` for any error. The status and JSON `changed` value describe the filtered result.

### `fbrcm versions export <project> <version>`

Exports one historical Remote Config template. Retrieval is cache-first and never changes the current pointer.

Flags:

```text
--to <path>   write normalized JSON to a private file; default prints JSON to stdout
--cached      require the exact local snapshot and perform no Firebase request
--yes, -y     overwrite an existing destination without confirmation
```

Normalization and overwrite protection match `project export`.

### `fbrcm versions rollback <project> <version>`

Uses Firebase's native rollback operation. It does not reactivate the old version number: Firebase force-publishes the selected historical template as a new version whose metadata records the rollback source.

Rollback refuses to run while the project has an unpublished draft.

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

### `fbrcm versions restore <project> <version>`

Republishes an exact locally cached immutable snapshot. Restore exists for recovery when Firebase no longer retains the historical version.

Restore refuses to run while the project has an unpublished draft.

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

Rollback and restore JSON results include `project_id`, `operation`, `previous_version`, `source_version`, `published_version`, `dry_run`, and `changed`, including no-op results where `changed` is `false`. Human previews are written separately from JSON data so stdout remains machine-readable.

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

Project synchronization retains projects that are no longer accessible instead of deleting them. A project with no accessible auth identity is marked disabled. If a later update discovers it through another configured identity, the project is automatically rebound to that identity and enabled. Project JSON includes a `disabled` boolean; human project listings mark disabled identities in the Auth column.

### `fbrcm projects forget`

Forgets matching locally tracked projects and deletes their cached Remote Config snapshots, cached versions, and drafts. It never deletes Firebase projects or otherwise reads from or writes to Firebase. With no filter or expression, every configured project is selected. The expression uses the same project context as `projects list`, but evaluates against local Remote Config cache only; project-only expressions work even when that cache is missing.

Flags:

```text
-f, --filter <query>   filter projects; may be repeated
--expr <expr>          filter projects with project context using local cache only
-y, --yes              skip confirmation
```

### `fbrcm projects diff <source-project> <target-project>`

Compares Remote Config between two projects. `<source-project>` is the desired config and `<target-project>` is the config being checked for drift. Both arguments use shared positional project resolution.

By default, command fetches live Remote Config for both projects. Use `--cached` to require the local projects registry and compare local parameter cache entries without contacting Firebase. Stale cache entries are compared as stored; a missing registry or Remote Config entry is an error.

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
--exit-code            return 1 for differences and 2 for errors
```

Default output is a terminal diff grouped by conditions, group descriptions, and parameters. JSON output includes source project, target project, top-level `changed`, summary counts, and structured change records.

Without `--exit-code`, both differences and no differences return success. With it, exit statuses are `0` for no differences, `1` for differences, and `2` for any error. The status and JSON `changed` value describe the filtered result.

### `fbrcm projects promote <source-project> <target-project>`

Promotes selected Remote Config changes from source project to target project. `<source-project>` is the desired config. `<target-project>` is the project that may be published.

Promotion refuses to publish when the target project has an unpublished draft.

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
--dry-run              preview without writing local or Firebase state
-y, --yes              skip final publish confirmation
--json                 print promotion result JSON
```

Promotion JSON includes `changed`, which reports whether the selected result contains changes independently of whether it was a dry run or was published.

Non-interactive promote requires explicit selection intent: `--all`, `--filter`, `--group`, `--expr`, or `--search`. Command reloads the target before publishing, validates with Firebase, publishes using the latest target ETag, and retries if the target changes during promotion.

### `fbrcm projects path`

Prints projects config file path.

Flags:

```text
--json   print {"path": "..."}
```

### `fbrcm projects reset`

Resets the locally cached projects registry by deleting its rebuildable config file. Project Remote Config snapshots, cached versions, and drafts are not deleted.

Flags:

```text
-y, --yes   skip confirmation
```

### `fbrcm doctor`

Runs a complete, non-interactive application health check. It verifies the selected profile and profile directories, auth registry, credential files, OAuth token presence and expiry, network/offline state, Cloud Resource Manager API access, Remote Config API reads, required Firebase read/update IAM permissions for cached projects, and profile cache writability.

Doctor never opens OAuth login and never persists a refreshed token. In offline mode it reports the state and skips live API and permission checks. It prints every check even when some fail, and exits with status 1 when any check has `fail` status; warnings alone do not fail the command. The diagnostic run has no overall time limit by default. Pressing `Ctrl+C` cancels the current check, prints the partial table or JSON report, and then exits nonzero.

An expired cached OAuth access token is normal when its refresh token still works. Online diagnostics report that token as `pass` after a successful in-memory refresh, `fail` when refresh fails, and `warn` only when refresh cannot be tested in offline mode. Doctor does not persist the refreshed access token.

Human-readable output uses the narrowest table and column widths that fit all content. When the natural table exceeds the detected terminal width, only Detail shrinks; long paths, permission lists, and API errors wrap onto additional lines inside that cell. Status and Check remain single-line and content-width.

Flags:

```text
--json                 print diagnostic checks as JSON
--timeout <duration>   optional positive time limit for the complete diagnostic run
```

JSON output is an array of checks. Every element includes the report-level profile, config directory, cache directory, and offline state.

### `fbrcm cache list`

Lists immutable cached Remote Config versions. Drafts have a separate lifecycle under `fbrcm draft` and are not included.

Flags:

```text
--json   print cache entries as JSON
```

JSON entries include project ID, project name, version, file size, cached time, and path.

### `fbrcm cache path`

Prints the directory containing immutable cached Remote Config snapshots for the active profile. It does not return the profile-wide cache root used by drafts and OAuth token caches.

Flags:

```text
--json   print {"path": "..."}
```

### `fbrcm cache clear`

Deletes all locally cached immutable Remote Config versions. The confirmation reports snapshot count, total size, and project count, and warns that versions no longer retained by Firebase may be permanently lost. Drafts are never deleted by this command.

Flags:

```text
-y, --yes   skip cache confirmation
```

Use `fbrcm draft discard` or `fbrcm draft discard --all` for explicit draft deletion.

### `fbrcm config path`

Prints global config file path.

Flags:

```text
--json   print {"path": "..."}
```

All `config` subcommands are local operations. They neither initialize a profile nor run the startup connectivity probe.

### `fbrcm config show [key]`

Shows the effective global configuration after applying built-in defaults and keybinding migration. With no key, human output is TOML and includes the complete effective key map. Supported keys are `profile`, `powerline_glyphs`, `keys`, `keys.<block>`, and `keys.<block>.<action>`. A selected scalar prints as plain text; a selected keybinding list or map prints scoped TOML. JSON is emitted only with `--json`.

Flags:

```text
--json   print structured JSON
```

Full JSON output has `path`, `exists`, and `config` fields. Selected-key JSON has `key`, `value`, and `source`; `source` is `configured`, `default`, or `migrated`. A missing config file is not created.

### `fbrcm config set <key> <value>...`

Sets a typed preference and atomically replaces the global config file with private permissions. Supported forms are:

```text
powerline_glyphs true|false
keys.<block>.<action> <key>...
```

The active `profile` is read-only here; use `fbrcm profile switch <name>`. The complete candidate configuration is validated before writing, including unknown blocks/actions, empty or duplicate bindings, unsupported key names, and conflicts with configured or default actions. Failed validation leaves the file unchanged.

Flags:

```text
--json   print {"key": "...", "previous": ..., "value": ..., "changed": true|false}
```

### `fbrcm config reset [key]`

Resets a preference to its built-in default. The optional key may be `powerline_glyphs`, `keys`, `keys.<block>`, or `keys.<block>.<action>`. With no key, it resets all preferences while preserving the persisted active profile. A changed reset asks for confirmation; Yes is selected by default. Writes are validated and atomic.

Flags:

```text
-y, --yes   reset without confirmation
    --json  print {"key": "...", "status": "reset|unchanged|canceled", "changed": true|false}
```

### `fbrcm config validate`

Strictly validates TOML structure, profile references, and the complete effective key map. It reports all keybinding diagnostics in stable order. A missing file is valid because built-in defaults apply. Invalid configuration returns exit status 1; operational failures also return nonzero.

Flags:

```text
--json   print {"path": "...", "exists": true|false, "valid": true|false, "errors": [...], "warnings": [...]}
```

Each diagnostic contains `severity`, `code`, `key`, and `message`.

### `fbrcm config edit`

Opens a staged copy of the global config in an editor. After the editor exits, fbrcm strictly validates the staged file and atomically replaces the original only when it is valid. If editing or validation fails, the original remains unchanged and the staged file path is preserved for recovery. When the config file does not exist, the editor starts with the complete default configuration.

Editor resolution order is `--editor`, `FBRCM_EDITOR`, `VISUAL`, `EDITOR`, then `vi` on Unix-like systems or `notepad.exe` on Windows. Commands may include arguments; GUI editors generally need their wait flag, for example `--editor "code --wait"`.

Flags:

```text
--editor <command>   override the editor command
```

### `fbrcm auth list`

Lists configured auth identities.

When a command needs Firebase access but the active profile has no configured auth identity, the error includes setup guidance. Run `fbrcm` for guided setup, or use `fbrcm auth add --help` to see the CLI authentication options.

Flags:

```text
--json   print auth identities as JSON
```

JSON output is an array. Every identity includes a `default` boolean; exactly the configured default identity has `default: true`.

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

### `fbrcm auth delete <auth-id>`

Deletes an auth identity and its client secret/token files.

Flags:

```text
-y, --yes   skip confirmation dialogs
```

### `fbrcm auth bind`

Binds cached projects to an auth identity. Without `--project`, every cached project is selected. Repeated project filters are ORed and use the same mode-prefixed name/project-ID matching as `get --project` and `projects list --filter`.

Flags:

```text
--auth <auth-id>          auth identity to bind
-p, --project <query>     filter projects; may be repeated
```

A project is rebound only when the target identity discovered it during project synchronization. Inaccessible projects are skipped, logged individually as errors, and counted in the final bound/skipped summary; they do not fail the rest of the batch.

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

### `fbrcm profile delete <profile>`

Deletes profile config and cache directories. Confirmation defaults to yes. The active profile cannot be deleted.

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
