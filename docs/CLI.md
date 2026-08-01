# fbrcm CLI

`fbrcm` is a Firebase Remote Config manager. It runs as an interactive TUI when called with no arguments. Any argument switches to CLI mode. See the [TUI guide](TUI.md) for the interactive workflow.

## Command Tree

```text
fbrcm [--help] [--version] [--profile <name>] [--no-local-config]
│
├── add <parameter>
│   ├── --project, -p <query>  repeated
│   ├── --expr <expr>
│   ├── --dry-run
│   ├── --draft
│   ├── --change-note <text>
│   ├── --yes, -y
│   ├── --json
│   ├── --description <text>
│   ├── --group <name>
│   ├── --type string|boolean|number|json
│   └── exactly one value source:
│       ├── --value <value>
│       └── --use-in-app-default
│
├── cache
│   ├── list [--json]
│   ├── path [--json]
│   └── clear [--yes|-y]
│
├── config
│   ├── path [--scope global|local] [--json]
│   ├── show [key] [--scope effective|global|local] [--json]
│   ├── set <key> <value>... [--scope global|local] [--json]
│   ├── reset [key] [--scope global|local] [--yes|-y] [--json]
│   ├── validate [--scope all|effective|global|local] [--json]
│   └── edit [--scope global|local] [--full] [--editor <command>]
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
│   │   ├── --change-note <text>
│   │   ├── --yes, -y
│   │   └── --json
│   ├── edit <project> <condition>
│   │   ├── --expression <expr>
│   │   ├── --color <color>
│   │   ├── --no-color
│   │   ├── --dry-run
│   │   ├── --draft
│   │   ├── --change-note <text>
│   │   ├── --yes, -y
│   │   └── --json
│   ├── rename <project> <condition> <new-name>
│   │   ├── --dry-run
│   │   ├── --draft
│   │   ├── --change-note <text>
│   │   ├── --yes, -y
│   │   └── --json
│   ├── move <project> <condition> <priority>
│   │   ├── --dry-run
│   │   ├── --draft
│   │   ├── --change-note <text>
│   │   ├── --yes, -y
│   │   └── --json
│   ├── delete <project> <condition>
│   │   ├── --dry-run
│   │   ├── --draft
│   │   ├── --change-note <text>
│   │   ├── --yes, -y
│   │   └── --json
│   └── validate <project> [--json]
│
├── delete [parameter]
│   ├── --project, -p <query>  repeated
│   ├── --filter, -f <query>   repeated
│   ├── --expr <expr>
│   ├── --search <text>
│   ├── --dry-run
│   ├── --draft
│   ├── --change-note <text>
│   ├── --yes, -y
│   └── --json
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
│   │   ├── --change-note <text>
│   │   ├── --yes, -y
│   │   └── --json
│   ├── change-note <project> [text]
│   │   ├── --clear
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
│   ├── --change-note <text>
│   ├── --yes, -y
│   └── --json
│
├── experiments
│   ├── list <project> [--filter|-f <query>]... [--update] [--json]
│   ├── show <project> <experiment-id> [--update] [--json]
│   └── delete <project> <experiment-id> [--yes|-y]
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
│   ├── add <name> [--project|-p <query>] [--description <text>] [--dry-run] [--draft] [--change-note <text>] [--yes|-y] [--json]
│   ├── edit <group> [--project|-p <query>] (--description <text>|--no-description) [--dry-run] [--draft] [--change-note <text>] [--yes|-y] [--json]
│   ├── rename <group> <new-name> [--project|-p <query>] [--dry-run] [--draft] [--change-note <text>] [--yes|-y] [--json]
│   └── delete <group> [--project|-p <query>] [--dry-run] [--draft] [--change-note <text>] [--yes|-y] [--json]
│
├── personalizations
│   ├── list <project> [--update] [--json]
│   └── show <project> <personalization-id> [--update] [--json]
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
│   ├── templates
│   │   ├── show <project> [--json]
│   │   └── set <project>
│   │       ├── --templates client|server|client,server
│   │       ├── --primary client|server
│   │       └── --json
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
│       ├── --change-note <text>
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
│   │   ├── --side-by-side
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
│       ├── --change-note <text>
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
│   │   ├── --change-note <text>
│   │   ├── --yes, -y
│   │   └── --json
│   ├── path [--json]
│   └── reset [--yes|-y]
│
├── rollouts
│   ├── list <project> [--update] [--json]
│   ├── show <project> <rollout-id> [--update] [--json]
│   └── delete <project> <rollout-id> [--yes|-y]
│
└── update [parameter]
    ├── --project, -p <query>  repeated
    ├── --filter, -f <query>   repeated
    ├── --expr <expr>
    ├── --search <text>
    ├── --dry-run
    ├── --draft
    ├── --change-note <text>
    ├── --yes, -y
    ├── --json
    ├── --description <text>
    ├── --group <name>
    ├── --no-group
    ├── --name <new-name>
    ├── --condition <name>
    ├── --remove-all-conditional-values
    ├── --remove-conditional-value <condition>  repeated
    ├── --type string|boolean|number|json
    └── at most one value source:
        ├── --value <value>
        └── --use-in-app-default
```

## Shared Behavior

All commands support `--help`. Root also supports `--version`.

When `FBRCM_OFFLINE` is unset, fbrcm performs a proxy-aware HTTPS connectivity probe before executing a network-capable CLI command and automatically enables offline mode if the probe fails. Help, version, and all `config` commands skip this probe. The probe and other standard HTTP requests honor `HTTPS_PROXY`, `HTTP_PROXY`, and `NO_PROXY`, including their lowercase forms. Defining `FBRCM_OFFLINE`, including with an empty value or `0`, enables offline mode without probing.

Most commands require a selected profile. `profile`, `config`, `doctor`, and `help` do not require profile initialization. Run `fbrcm profile switch <name>` to switch or create a profile. Use the root `--profile <name>` flag or `FBRCM_PROFILE` to select an existing profile for one process without changing the persisted active profile; the flag takes precedence over the environment variable.

At startup, fbrcm searches the current directory and every parent through the
filesystem root for `.fbrcm.toml`. The nearest match deeply overlays the global
`config.toml`: nested tables merge, while scalars and arrays replace lower-layer
values. Built-in defaults apply after both stored layers. Profile precedence is
`--profile`, `FBRCM_PROFILE`, local config, global config, then `default`.
Use `--no-local-config` or set `FBRCM_NO_LOCAL_CONFIG` to ignore repository
configuration. Config commands ignore `--profile`, but honor their own explicit
configuration scope.

Interactive yes/no confirmations select **Yes** by default. Use the arrow keys to select No, or pass `--yes` where available to skip the prompt.

Long-running CLI work displays a gray animated progress line on stderr when
stderr is an interactive terminal. The message changes for major phases such
as project resolution, Remote Config loading, validation, publication, and
draft updates. Durable log lines temporarily replace the progress line and the
animation resumes underneath them. Progress is erased before results, diffs,
diagnostics, file pickers, editors, or confirmation prompts are displayed, and
it is never written when stderr is redirected. `FBRCM_LOG_LEVEL=silent`
suppresses logs without suppressing interactive progress.

Human-readable collection output always renders its normal table, including when there are no rows. An empty result therefore contains the table headers rather than a special empty-state message. A command whose primary JSON result is a collection always emits a top-level array and uses `[]` when empty. Singular resources and single-operation reports remain JSON objects.

Auth identities, project cache, parameter cache, and drafts are profile-scoped. Project cache stores known projects plus their selected `auth_id`. Default storage lives under user config/cache directories. Override roots with:

```text
FBRCM_CONFIG_DIR
FBRCM_CACHE_DIR
FBRCM_PROFILE
```

### Environment Variables

| Variable | Behavior |
| --- | --- |
| `FBRCM_PROFILE` | Select an existing profile for this process. The root `--profile` flag takes precedence. |
| `FBRCM_CONFIG_DIR` | Override the fbrcm config root. Takes precedence over `XDG_CONFIG_HOME` and the home-directory fallback. |
| `FBRCM_CACHE_DIR` | Override the fbrcm cache root. Takes precedence over the operating system's user-cache directory. |
| `FBRCM_OFFLINE` | Enable offline mode whenever the variable is defined, including as an empty string or `0`. If it is unset, network-capable commands perform a short, proxy-aware connectivity probe and may enable offline mode automatically. |
| `FBRCM_LOG_LEVEL` | Set logging to `debug`, `info`, `warn`, `error`, `fatal`, or `silent`, case-insensitively. The default is `info`. |
| `FBRCM_EDITOR` | Select the command used by `config edit`, after `--editor` and before `VISUAL` or `EDITOR`. Arguments are supported. |
| `FBRCM_NO_LOCAL_CONFIG` | Ignore repository `.fbrcm.toml` discovery when set to a non-empty value. The root `--no-local-config` flag provides the same behavior for one invocation. |
| `NO_COLOR` | Disable CLI, prompt, log, and TUI colors when set to a non-empty value. |
| `COLUMNS` | Supply a positive terminal width for human-readable CLI output. Invalid values are ignored. |
| `GOOGLE_APPLICATION_CREDENTIALS` | Select an Application Default Credentials JSON file for gcloud identities and diagnostics. |
| `XDG_CONFIG_HOME` | Supply the Unix config home when `FBRCM_CONFIG_DIR` is unset; fbrcm appends `fbrcm`. |
| `XDG_CACHE_HOME` | Supply the Unix user-cache home where supported when `FBRCM_CACHE_DIR` is unset; fbrcm appends `fbrcm`. |
| `HTTPS_PROXY`, `HTTP_PROXY`, `NO_PROXY` | Configure Go's HTTP transport and startup connectivity probe. Lowercase forms are also honored. |

`config edit` also consults `VISUAL`, `EDITOR`, and on Unix-like systems
`SHELL`. Directory discovery follows the operating system and may use `HOME`,
`USERPROFILE`, `LOCALAPPDATA`, or `APPDATA`.

### Filter Queries

Flags named `--project` or `--filter` use mode-prefixed query strings:

```text
~query   fuzzy match; default if no prefix is given
^query   starts-with match
/query   includes match
=query   exact case-insensitive match
```

Project filters match project display name or project ID. Parameter filters match parameter key. `--project` and `--filter` may be repeated; repeated values are ORed and must be passed as separate flags.

### Client and Server Template Targets

Remote Config commands accept a template target wherever their command syntax shows a Remote Config `<project>` or a `--project` filter:

```text
project-id          configured template selection; implicit form
client@project-id   client template; explicit alias
server@project-id   server template in the firebase-server namespace
```

The prefix comes before a filter mode. For example, `-p 'server@=api-prod'` selects the server template of exactly `api-prod`, while `-p 'client@^mobile-'` selects client templates whose project name or ID starts with `mobile-`. Repeated flags can mix client and server targets in one invocation.

Each cached project stores its enabled template selections and one primary template. New and existing projects default to client-only. An unqualified bulk filter, or no `--project` filter, expands every matched project to its configured enabled templates. An unqualified positional `<project>` selects that project's primary template. Explicit `client@` and `server@` prefixes always select exactly that template, independently of the saved selections.

The target syntax applies to `add`, `get`, `update`, `delete`, `duplicate`, and `groups`; all `conditions` and `versions` commands; `draft` commands; `project export`, `project import`, and `project defaults`; and the source and destination of `projects diff` and `projects promote`.

Project metadata and managed-feature commands remain project-scoped rather than template-scoped. In particular, `projects list`, `projects update`, `projects forget`, `project show`, `project templates`, `project open`, `experiments`, `rollouts`, `personalizations`, `auth bind`, and `doctor` continue to accept ordinary project IDs or names without a template prefix. Managed features belong to the client Remote Config namespace; these commands reject `client@` and `server@` target syntax. A `server@` target reports that managed features support only the client namespace, while a `client@` target asks you to omit the unnecessary prefix.

Client targets are canonicalized to the unqualified project ID, so `project-id` and `client@project-id` share exactly the same cache, version snapshots, and draft. Server targets retain their prefix and use separate local files:

```text
$(fbrcm cache path)/project-id.3.json          client template version 3
$(fbrcm cache path)/server@project-id.3.json   server template version 3
$(fbrcm draft path)/project-id.json            client draft
$(fbrcm draft path)/server@project-id.json     server draft
```

Client and server templates have independent Firebase histories. CLI output uses the canonical target ID: unqualified for client templates and `server@project-id` for server templates.

### Positional Project Resolution

Template-aware commands first parse the optional `client@` or `server@` prefix, then resolve the remaining positional `<project>` in this order:

1. Exact case-insensitive project ID.
2. Exact case-insensitive project display name.
3. Case-insensitive substring match against project ID or display name.

A single match is selected. Multiple exact-name or substring matches print only the ambiguous projects and return an error. No match prints the known-project table and returns an error. Exact ID always wins, including when another project's display name has the same text.

Draft commands resolve only locally stored drafts and never synchronize projects as a side effect. An explicit prefix selects that template kind. An unqualified query selects the configured primary template when the project is still registered, and falls back to the client template for an unregistered project. The query must uniquely match the locally known project ID or display name. This also permits `show --raw` and `discard` for drafts whose project is no longer present in the projects cache.

### Parameter Search

Parameter-context commands also support `--search <text>`. It searches parameter name, description, default value, conditional values, condition names, and condition expressions. Name/description/condition-name matching is case-insensitive and ignores punctuation; value/expression matching is case-sensitive. `--search` is ANDed with `--filter` and parameter-context `--expr`.

### Expression Filters

`--expr` uses expr-lang and must evaluate to boolean. See [EXPR.md](EXPR.md) for full context fields and helper functions.

Parameter-context commands:

```text
get
delete
update
draft diff
project import
versions diff
projects diff
projects promote
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

`get` also accepts a directory passed as stdin. It reads top-level `.json` files from that directory, accepts raw Remote Config JSON or fbrcm cache JSON in each file, and treats each file stem as a canonical template target. An unqualified stem is a client target and a `server@` stem is a server target. Project name is built from the underlying project ID by splitting on `-` and `_`, then capitalizing words.

`project import` reads JSON from `--from`, stdin, or an interactive `.json` file picker. It accepts raw Remote Config JSON or an fbrcm parameters cache JSON file with `remote_config`.

`--draft` is unavailable in stdin transformation mode because piped input has no persistent target project identity.

### Draft lifecycle and write safety

Drafts are profile-scoped, target-specific, self-contained records. Each version-1 record stores the working Remote Config, its immutable base Remote Config, base version and ETag, timestamps, and an optional `change_note`. A project can therefore have independent client and server drafts. Plain Remote Config JSON is not accepted as an on-disk draft format, and no legacy draft migration or fallback is performed.

`add`, `duplicate`, `update`, `delete`, group and condition mutations, `project import`, `projects promote`, `draft publish`, and `versions restore` accept `--change-note <text>`. The note is trimmed, must be one line, and is sent as Firebase `version.description` on ordinary publication. With `--draft`, it is stored as `change_note` and remains editable with `draft change-note`; an explicitly empty note clears it. Native `versions rollback` intentionally has no change-note flag because Firebase owns rollback metadata.

`add`, `duplicate`, `update`, `delete`, `project import`, and the condition mutation commands accept `--draft`. In draft mode they apply changes on top of an existing project draft or create a new draft from freshly revalidated Remote Config. They do not validate or publish to Firebase. Combining `--draft` with `--dry-run` previews the change without writing either draft or Firebase state.

Immediate Remote Config writes refuse to proceed when the target has an unpublished draft. This guard applies to add, duplicate, update, delete, condition mutations, project import, version rollback/restore, and project promotion. Resolve the draft with `draft publish` or `draft discard`, or add the intended mutation to it with `--draft`.

Multi-project Remote Config publication is non-atomic: Firebase accepts a separate validated write for each project. Commands process every selected project even when an independent project fails, collect one outcome per project, and print the complete `Results:` block after operation logging has finished. They return nonzero after the batch if any project failed. Successful projects are not rolled back. Conflicts are reported for a fresh explicit retry instead of silently recalculating and publishing a different candidate. Failed-project output includes exact `-p '=project-id'` filters for retrying only projects that were not published.

### Mutation JSON automation contract

Direct Remote Config mutations—`add`, `update`, `delete`, `duplicate`, all condition mutations, and all group mutations—accept `--json`. JSON output is an ordered array with one stable result object per selected template target:

```json
[
  {
    "target": "my-project",
    "status": "published",
    "changed_item_count": 1,
    "previous_version": "41",
    "published_version": "42",
    "draft": false,
    "dry_run": false,
    "change_note": "Enable checkout v2",
    "error": null,
    "retry_selector": null
  }
]
```

`previous_version`, `published_version`, and `change_note` are `null` when unavailable or omitted. `error` is either `null` or an object with `stage` (`preparation`, `validation`, `publication`, `draft`, or `cache`) and `message`. A failed target that is safe to retry includes an exact, target-aware `retry_selector`, such as `=my-project` or `server@=my-project`; pass it back as `--project <retry_selector>`. A `published-cache-failed` target has no retry selector because Firebase was already updated and the correct recovery is a cache refresh. JSON mode keeps stdout machine-readable but does not imply `--yes`; prompts and review output continue on stderr. In stdin transformation mode, `add`, `update`, and `delete` continue to emit the transformed Remote Config JSON itself; `--change-note` is unavailable because no publication or draft write occurs.

If Firebase accepts a publish but the returned state cannot be saved locally, the outcome is reported as `published-cache-failed`, not as an unpublished project. Refresh that project's cache instead of blindly retrying the mutation. For coordinated changes, `--draft` provides reviewable and recoverable intent, but publishing those drafts is still non-atomic across projects.

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

`--type` plus exactly one value source is required:

```text
--type <type>           STRING, BOOLEAN, NUMBER, or JSON
--value <value>         concrete value interpreted according to --type
--use-in-app-default   delegate the value to the application instead of setting --value
```

`--value` and `--use-in-app-default` are mutually exclusive. Boolean values must be `true` or `false`, number values must parse as a float, and JSON values must be valid JSON. String values may be empty when `--value ""` is passed explicitly.

Other flags:

```text
-p, --project <query>      filter template targets; may be repeated
--expr <expr>              filter target projects with project context
--dry-run                  preview without writing local or Firebase state
--draft                    save changes to local drafts instead of publishing
--change-note <text>       set the change note for publication or draft storage
-y, --yes                  print diff and add without confirmation
--json                     print structured mutation results
--description <text>       parameter description
--group <name>             add parameter inside group
```

Remote mode loads projects, filters them, and adds the parameter where it does not already exist. It prints the complete Remote Config diff and asks for confirmation for each project unless `--yes` is set, then validates and publishes each confirmed project independently. Existing parameters are skipped.

With `--draft`, each confirmed mutation is stored locally on top of any existing draft. Without `--draft`, the command refuses projects that already have unpublished drafts.

Stdin mode reads Remote Config JSON from stdin, adds parameter to that JSON, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field.

### `fbrcm duplicate <source> <target>`

Duplicates one complete Remote Config parameter in every matched project. The copy preserves the source group, description, value type, default value, conditional values, and condition references. Source lookup is exact and case-insensitive. A project without the source parameter is skipped; an ambiguous source or an existing target name is an error. Target collision checks are also case-insensitive and never overwrite an existing parameter.

Flags:

```text
-p, --project <query>   filter template targets; may be repeated
--expr <expr>           filter target projects with project context
--dry-run               preview without writing local or Firebase state
--draft                 save changes to local drafts instead of publishing
--change-note <text>    set the change note for publication or draft storage
-y, --yes               print diff and duplicate without confirmation
--json                  print structured mutation results
```

Remote mode prints the complete Remote Config diff and prompts before each duplicate unless `--yes` is set. It validates and publishes each project independently. A conflict fails that project without suppressing later projects, and the final command status is nonzero when any project fails. Project filters are applied before the project-context expression, matching `add` behavior.

With `--draft`, duplication composes onto each existing draft and remains local. Without `--draft`, a project with an unpublished draft fails independently while other selected projects continue. The command does not use stdin transformation mode.

### `fbrcm get [parameter]`

Prints Remote Config parameters across template targets.

Passing `[parameter]` is shorthand for `--filter =<parameter>`. It cannot be combined with `--filter`.

Flags:

```text
-p, --project <query>   filter template targets; may be repeated
-f, --filter <query>    filter parameters; may be repeated
--expr <expr>           filter parameters with parameter context
--search <text>         search parameter names, descriptions, values, and conditions
--json                  print JSON rows
--all                   include projects with no matching parameters in table output
--update                revalidate cached parameters before printing
```

Default output is a terminal table. Firebase-managed values use compact human summaries: personalizations render as `◈ (personalization)`, experiments as `⚗ (a/b test)`, and unknown future value options as `(optionName)`, all in muted text. Rollouts render their published value without another API request as `◐ 10% → 20 / ◑ (no change)`; the percentage uses the shared gold count color and the rollout chrome and no-change marker are muted. JSON output includes the same unstyled summaries together with project, project ID, group, key, description, default value, conditionals, type, version, cache time, and status.

Stdin mode reads Remote Config JSON from stdin and queries only that config. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. If stdin is a directory, `get` reads top-level `.json` files and treats them as multiple projects.

### `fbrcm update [parameter]`

Updates matched Remote Config parameters. Passing `[parameter]` is shorthand for `--filter =<parameter>`. It cannot be combined with `--filter`.

Flags:

```text
-p, --project <query>      filter template targets; may be repeated
-f, --filter <query>       filter parameters; may be repeated
--expr <expr>              filter parameters with parameter context
--search <text>            search parameter names, descriptions, values, and conditions
--dry-run                  preview without writing local or Firebase state
--draft                    save changes to local drafts instead of publishing
--change-note <text>       set the change note for publication or draft storage
-y, --yes                  print diff and update without confirmation
--json                     print structured mutation results
--description <text>       set parameter description
--group <name>             move parameter into group
--no-group                 move parameter out of any group
--name <new-name>          rename parameter; cannot be empty
--condition <name>         assign the selected value to this condition instead of the default value
--remove-all-conditional-values
                           remove all conditional values from matched parameters
--remove-conditional-value <condition>
                           remove named conditional value from matched parameters; may be repeated
--type <type>              STRING, BOOLEAN, NUMBER, or JSON
--value <value>            set a concrete value interpreted according to --type
--use-in-app-default       delegate the selected value to the application instead of setting --value
```

`--value` and `--use-in-app-default` are mutually exclusive, and either one requires `--type`. `--type` cannot be passed alone. `--condition` requires one of these value sources and resolves the condition by exact name, then exact case-insensitive name. It preserves the default and all other conditional values while assigning the selected typed value. Boolean, number, and JSON values use the same validation as `add`. `--group` and `--no-group` are mutually exclusive. `--condition`, `--remove-all-conditional-values`, and `--remove-conditional-value` are mutually exclusive.

`--use-in-app-default` sets Firebase's `useInAppDefault` source on the default value, or on the conditional value selected by `--condition`. Both value sources set the parameter type from the required `--type`.

Conditional value assignment and removal edit only `conditionalValues`; they keep the parameter, default value, description, group, and all conditions themselves.

Firebase-managed personalization, A/B test, and rollout values are read-only. `update` rejects replacing them, changing their parameter type, removing their conditional slots, or relocating them through a parameter or condition rename/move. Unknown future value options receive the same protection.

Remote mode prints diffs and prompts unless `--yes` is set. It validates and publishes each project independently, reports every outcome, continues after project-scoped failures, and returns nonzero if any project failed.

With `--draft`, mutations compose onto each existing draft and remain local. Without `--draft`, publication is best-effort and non-atomic; failures do not roll back earlier projects or prevent later independent projects from being attempted.

Stdin mode reads Remote Config JSON from stdin, updates matching parameters, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. It does not prompt.

### `fbrcm delete [parameter]`

Deletes matched Remote Config parameters. Passing `[parameter]` is shorthand for `--filter =<parameter>`. It cannot be combined with `--filter`.

Flags:

```text
-p, --project <query>   filter template targets; may be repeated
-f, --filter <query>    filter parameters; may be repeated
--expr <expr>           filter parameters with parameter context
--search <text>         search parameter names, descriptions, values, and conditions
--dry-run               preview without writing local or Firebase state
--draft                 save changes to local drafts instead of publishing
--change-note <text>    set the change note for publication or draft storage
-y, --yes               print diff and delete without confirmation
--json                  print structured mutation results
```

Remote mode prints diffs and prompts unless `--yes` is set. It validates and publishes each project independently, reports every outcome, continues after project-scoped failures, and returns nonzero if any project failed.

With `--draft`, deletions are saved locally on top of any existing draft. Without `--draft`, a project with an unpublished draft fails independently while other selected projects continue.

Stdin mode reads Remote Config JSON from stdin, deletes matching parameters, and prints final JSON. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. It does not prompt.

Parameters containing Firebase-managed or unknown future values cannot be deleted. The same protection applies in remote, draft, and stdin modes.

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

Condition filters use the shared mode prefixes described under Filter Queries. Repeated filters are ORed. `--filter`, `--search`, and `--expr` are ANDed together. See [EXPR.md](EXPR.md) for condition context fields and examples.

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
--change-note <text>
            set the change note for publication or draft storage
-y, --yes   print the diff and apply without confirmation
--json      print structured mutation results
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

`groups list` lists real Firebase parameter groups across the selected template targets, including intentionally empty and description-only groups. It uses an unpublished target-specific draft when present and otherwise follows the same fresh/stale cache behavior as condition reads. Human output is a naturally sized table with canonical target ID, project name, parameter count, and description; the project column is omitted for one exact `--project` target filter, matching `get`. On narrow terminals, the description is cropped with an ellipsis first, followed by target ID and group name only when necessary.

List flags:

```text
-p, --project <query>  filter template targets by optional client@ or server@ project query; may be repeated
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

All group commands support repeatable `--project|-p` target filters with the same mode prefixes and OR behavior as `get`, `add`, `delete`, and `update`. With no project filter, they process every configured enabled template in stable project-name/target-ID order. Named mutations skip targets that do not contain the group; `add` skips targets where it already exists.

All group mutations also support `--dry-run`, `--draft`, `--change-note`, `--yes|-y`, and `--json`, with the same diff, confirmation, validation, ETag, draft-composition, draft-conflict, and structured-result behavior as condition mutations. `--description` and `--no-description` are mutually exclusive.

### `fbrcm draft list`

Lists drafts in the active profile without contacting Firebase. Invalid draft envelopes remain visible instead of failing the complete listing.

Flags:

```text
-f, --filter <query>   filter by optional client@ or server@ project query; may be repeated
--json                 print structured JSON
```

Human output includes canonical target ID, project name, base version, update time, parameter/condition change counts, status, and the optional Change Note as the final column. Status is `ready`, `unchanged`, or `invalid`.

JSON entries include `project_id`, `project`, `base_version`, `created_at`, `updated_at`, byte size, status, validity, base availability, path, change counts, and `change_note`.

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

### `fbrcm draft change-note <project> [text]`

Sets, replaces, or clears the optional note stored with one draft without changing its Remote Config candidate.

```text
--clear   clear the stored note; mutually exclusive with [text]
--json    print project_id and change_note
```

The note must be a single line. An empty `[text]` also clears it. Draft format remains version 1; the field is stored as `change_note`.

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
--change-note <text>
               override the stored note for every selected draft
-y, --yes      skip publish confirmations
--json         print structured results
```

For each project, the command fetches current Firebase state, merges local intent onto it, displays `current → candidate`, and asks for confirmation. It then validates and publishes that exact candidate with the fetched ETag. A remote change after preview is reported as a conflict rather than silently producing a different candidate. Conflicts and validation failures preserve the draft.

If current Firebase state already contains the effective draft changes, no new version is created and the draft is removed as `already-applied`. Batch mode is non-atomic, continues after independent project failures, prints its collected results together at the end followed by a targeted retry command, and returns nonzero if any item failed.

JSON output is an array of results. Each result includes project ID, status, base/previous/published versions, `rebased`, `changed`, `draft_deleted`, `dry_run`, `change_note`, and an optional error. Status values include `published`, `would-publish`, `already-applied`, `canceled`, `failed`, `conflict`, `published-cache-failed`, and `published-cleanup-failed`. Prompts, warnings, retry hints, and human diffs are kept off JSON stdout.

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

Shows cached project metadata, enabled and primary templates, the selected auth identity used for project operations, and every configured auth identity that discovered the project during synchronization. `<project>` uses the shared positional project resolution described above.

Flags:

```text
--update   synchronize projects with every configured auth identity before printing
--json     print the project using the same JSON contract as `projects list --json`, including its Firebase Console URL
```

Without `--update`, auth access reflects the latest cached project synchronization.

The status is `disabled` when the latest project synchronization could not find the project through its assigned auth identity or any replacement identity. Disabled projects remain cached and visible, but live Firebase operations are blocked until a later update rediscovers and automatically rebinds them.

### `fbrcm project templates show <project>`

Shows the enabled templates and primary template stored for one physical project. It reads only the local projects registry and does not synchronize projects or contact Firebase. `<project>` uses normal cached project name or ID resolution; explicit `client@` and `server@` prefixes are rejected because the preferences belong to the physical project.

Flags:

```text
--json   print project, project_id, templates, and primary_template
```

### `fbrcm project templates set <project>`

Updates the local template selection for one physical project and prints the resulting state. It does not create, fetch, publish, or delete any Firebase template, cache, version, or draft.

Flags:

```text
--templates <list>          replace enabled templates with client, server, or client,server
--primary client|server     set the primary template
--json                      print the resulting state as JSON
```

At least one mutation flag is required. `--templates` accepts comma-separated values and may be repeated; duplicates are ignored and stored in canonical client-then-server order. With one enabled template and no explicit `--primary`, that template automatically becomes primary. With both enabled and no explicit `--primary`, the existing primary is preserved. `--primary` without `--templates` changes only the primary. The command rejects a primary template that is not enabled.

Examples:

```sh
fbrcm project templates set northstar-wallet --templates client
fbrcm project templates set northstar-wallet --templates server
fbrcm project templates set northstar-wallet --templates client,server
fbrcm project templates set northstar-wallet --templates client,server --primary server
fbrcm project templates set northstar-wallet --primary client
```

### `fbrcm project open <project>`

Opens the project's Remote Config page in the Firebase console. `<project>` uses the shared positional project resolution described above.

### `fbrcm project export <project>`

Exports one template target's Remote Config JSON. `<project>` accepts the target syntax and uses the shared positional resolution described above.

Flags:

```text
--to <path>   write JSON to file; default prints JSON to stdout
--yes, -y     overwrite an existing destination without confirmation
```

Export normalizes JSON by unescaping `<`, `>`, `&`, trimming trailing line breaks, and ordering numeric conditional value keys before non-numeric keys. When `--to` names an existing file, export asks before replacing it unless `--yes` is set. A destination created after the initial check is not overwritten without authorization.

### `fbrcm project defaults <project>`

Downloads the selected client or server Remote Config parameter defaults directly from Firebase. `<project>` accepts the template target syntax and uses the shared positional resolution described above. JSON is suitable for web applications, XML for Android, and plist for Apple applications.

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
--change-note <text>                     set the change note for publication or draft storage
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

JSON output is one object containing `project_id`, `status`, `changed`, `draft`, `dry_run`, and `change_note`. Status is `imported`, `would-import`, `drafted`, `would-draft`, `unchanged`, or `canceled`. JSON mode suppresses human condition summaries and diffs but does not imply `--yes` or choose an import strategy.

### Remote Config managed features

Experiments and rollouts provide read-only `list` and `show` commands plus an explicit destructive `delete` command. Personalizations remain read-only. The CLI cannot create, start, stop, or edit managed features, and none of these commands publish Remote Config. All three command groups use ordinary positional project resolution and the client Remote Config namespace.

Experiment and rollout metadata comes from Firebase's public Remote Config v1 managed-feature endpoints. fbrcm prefers the numeric `project_number` saved by project discovery and falls back to the Firebase project ID when that number is absent; both forms are accepted by the managed-feature resource paths.

Experiment, rollout, and personalization bindings come from `experimentValue`, `rolloutValue`, and `personalizationValue` objects in the published Remote Config template. Normal reads use the standard cache/revalidation policy. `--update` explicitly revalidates that cached template before scanning it. Drafts are intentionally excluded because Firebase managed-feature state refers to the published template.

These commands project known binding fields for display without rewriting managed values. Validate and publish preparation preserves complete opaque `experimentValue`, `rolloutValue`, and `personalizationValue` objects. Value editors treat all three as read-only rather than plain values. Template mutations also reject adding, replacing, removing, duplicating, renaming, or relocating managed values; imports, promotions, draft publication, stdin mutations, and unknown future value options use the same guard. Machine-readable experiment and rollout output also preserves unrecognized fields from Firebase's beta managed-feature responses so schema additions are not silently discarded.

### `fbrcm experiments list <project>`

Lists every experiment returned by Firebase using only the paginated list endpoint and correlates it with published-template bindings. Human output includes experiment ID, display name, parameter, condition, exposure percentage, relative last-update time, and state, in that order. Parameter names use the same blue styling as `get`, and condition names use their configured Remote Config tag colors. Missing values are shown as empty-value dashes, while an explicitly configured zero exposure is shown as `0%`. Descriptions and detail-only metadata such as variants and objectives are omitted from the human list. An experiment with no binding in the current template remains visible with empty binding columns.

Flags:

```text
-f, --filter <query>   filter display names locally; may be repeated
--update                revalidate cached Remote Config before reading bindings
--json                  print a top-level array of filtered experiment objects with references
```

Experiment filters use the shared mode prefixes described under Filter Queries. Repeated filters are ORed. Matching is case-insensitive and applies only to the experiment display name, not its description or resource ID. Filtering is local after all list pages have been loaded; fbrcm does not send the query to Firebase.

### `fbrcm experiments show <project> <experiment-id>`

Shows one experiment's display metadata, state, timestamps, activation event, variants and weights, primary and secondary objectives, and every published parameter binding. Binding details include the experiment exposure percentage and each template variant ID with its value or no-change marker. Empty-string variant values are displayed as `""`, while absent values remain an empty-value dash. `<experiment-id>` is the final component printed in the list table, such as `2`.

Flags:

```text
--update   revalidate cached Remote Config before reading bindings
--json     print project/template context and the complete Firebase experiment object with references
```

The metadata endpoint and template are intentionally correlated by experiment ID. Firebase's metadata remains authoritative for experiment state and definitions; the published template remains authoritative for affected parameters, per-variant values, and exposure.

### `fbrcm experiments delete <project> <experiment-id>`

Loads the experiment first so the confirmation names its display name and ID, then calls Firebase's experiment DELETE endpoint. The command does not publish or independently rewrite the Remote Config template.

Flags:

```text
-y, --yes   delete without confirmation
```

Without `--yes`, the destructive confirmation selects Yes by default. Selecting No cancels successfully without deleting anything. Successful output names the deleted experiment and project.

### `fbrcm rollouts list <project>`

Lists Firebase rollouts and correlates each rollout ID with its published-template parameter bindings. Human output includes ID, display name, parameter, condition, percentage, relative last-update time, and state, in that order. Parameter names use the same blue styling as `get`, and condition names use their configured Remote Config tag colors. Descriptions and enabled values are omitted from the human list but remain available through JSON and `rollouts show`. A rollout with no binding in the current template remains visible with empty binding columns.

Flags:

```text
--update   revalidate cached Remote Config before reading bindings
--json     print a top-level array of rollout objects with references
```

### `fbrcm rollouts show <project> <rollout-id>`

Shows rollout metadata, create/start/end/update timestamps, control and enabled variant names, and every published parameter binding. Explicit `0%` traffic and empty-string rollout values remain distinguishable from absent fields. `<rollout-id>` is the final component printed in the list table, such as `rollout_1`.

Flags:

```text
--update   revalidate cached Remote Config before reading bindings
--json     print project/template context and the rollout with references
```

### `fbrcm rollouts delete <project> <rollout-id>`

Loads the rollout first so the confirmation names its display name and ID, then calls Firebase's rollout DELETE endpoint. The command does not publish or independently rewrite the Remote Config template or local cache.

Flags:

```text
-y, --yes   delete without confirmation
```

Without `--yes`, the destructive confirmation selects Yes by default. Selecting No cancels successfully without deleting anything. Successful output names the deleted rollout and project.

### `fbrcm personalizations list <project>`

Scans the published template and lists every personalization ID with its group, parameter, and condition. Parameter names use the same blue styling as `get`, and condition names use their configured Remote Config tag colors. Firebase does not provide a separate public personalization resource endpoint, so the template is the authoritative API-visible source.

Flags:

```text
--update   revalidate cached Remote Config before scanning it
--json     print a top-level array of personalization IDs and references
```

### `fbrcm personalizations show <project> <personalization-id>`

Shows every published parameter binding for one personalization ID.

Flags:

```text
--update   revalidate cached Remote Config before scanning it
--json     print project/template context and the personalization references
```

Firebase exposes the personalization ID and binding in the template, but not the candidate values, chosen values, objectives, or result metrics through this API. Human output states that limitation explicitly.

### Remote Config version history

Version commands are scoped to one template target and use the same target resolution as `project export`: project ID is matched first, followed by exact display name case-insensitively. Client and server histories and local snapshots are independent.

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

Human live output keeps the existing column order: version number, current marker, publication time, updating user, origin, update type, cached marker, and Change Note. Cached output keeps version, current marker, cache time, size, and Change Note.

In cached mode, `--since` and `--until` apply to the local cache time because authoritative publication metadata may be unavailable.

JSON output is a plain array without project or pagination metadata. Each element uses fbrcm's canonical `change_note` name together with the other Firebase metadata, `current`, `cached`, and available local cache fields. Raw Firebase templates still encode this value in the API-required `version.description` field.

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
--side-by-side         print a static two-column terminal diff
--exit-code            return 1 for differences and 2 for errors
```

`--parameters` and `--conditions` are mutually exclusive. `--json` and `--side-by-side` are also mutually exclusive.

Default output reuses the conditions, group descriptions, parameters, and summary diff format used by `projects diff`. `--side-by-side` prints every changed entity as a complete, non-interactive two-column view. The command header establishes the `<from> → <to>` direction; individual changes omit repeated column headers and outer borders. Text wraps within the detected terminal width, JSON values are formatted before comparison, and long values retain contextual chunks around each difference. JSON output contains `project`, `from_version`, `to_version`, `changed`, and `diff`.

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
--change-note <text>
            set the new version's change note
-y, --yes   skip final publish confirmation
--json      print a structured operation result
```

Restore JSON includes `change_note`. Native rollback does not accept a change note and leaves its Firebase-defined rollback semantics unchanged.

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

Project synchronization retains projects that are no longer accessible instead of deleting them. A project with no accessible auth identity is marked disabled. If a later update discovers it through another configured identity, the project is automatically rebound to that identity and enabled. Project JSON includes `disabled`, `templates`, and `primary_template`; human project listings mark disabled identities in the Auth column.

### `fbrcm projects forget`

Forgets matching locally tracked projects and deletes both their client and server cached Remote Config snapshots, cached versions, and drafts. It never deletes Firebase projects or otherwise reads from or writes to Firebase. With no filter or expression, every configured project is selected. The expression uses the same project context as `projects list`, but evaluates against local client Remote Config cache only; project-only expressions work even when that cache is missing.

Flags:

```text
-f, --filter <query>   filter projects; may be repeated
--expr <expr>          filter projects with project context using local cache only
-y, --yes              skip confirmation
```

### `fbrcm projects diff <source-project> <target-project>`

Compares Remote Config between two template targets. `<source-project>` is the desired config and `<target-project>` is the config being checked for drift. Each argument independently accepts an implicit client, explicit `client@`, or `server@` target, so comparison can cross both projects and template kinds.

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

Promotes selected Remote Config changes from one template target to another. `<source-project>` is the desired config. `<target-project>` is the target that may be published. Each argument independently accepts an implicit client, explicit `client@`, or `server@` target.

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
--change-note <text>   set the change note
-y, --yes              skip final publish confirmation
--json                 print promotion result JSON
```

Promotion JSON includes `change_note` and `changed`; `changed` reports whether the selected result contains changes independently of whether it was a dry run or was published.

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

Lists immutable cached Remote Config versions for client and server template targets. Client entries use the unqualified project ID; server entries use `server@project-id`. Drafts have a separate lifecycle under `fbrcm draft` and are not included.

Flags:

```text
--json   print cache entries as JSON
```

JSON entries include canonical target ID in `project_id`, underlying project name, version, file size, cached time, and path.

### `fbrcm cache path`

Prints the directory containing immutable cached Remote Config snapshots for the active profile. It does not return the profile-wide cache root used by drafts and OAuth token caches.

Flags:

```text
--json   print {"path": "..."}
```

### `fbrcm cache clear`

Deletes all locally cached immutable Remote Config versions for both template kinds. The confirmation reports snapshot count, total size, and template-target count, and warns that versions no longer retained by Firebase may be permanently lost. Drafts are never deleted by this command.

Flags:

```text
-y, --yes   skip cache confirmation
```

Use `fbrcm draft discard` or `fbrcm draft discard --all` for explicit draft deletion.

### `fbrcm config path`

Prints the global config file path by default. `--scope local` prints the nearest
discovered `.fbrcm.toml`. When none exists, it returns a nonzero error with the
current-directory creation candidate and suggests `config edit --scope local`.

Flags:

```text
--scope global|local   select the stored layer; default global
--json                 print {"scope": "...", "path": "...", "exists": true|false}
```

All `config` subcommands are local operations. They neither initialize a profile nor run the startup connectivity probe.

### `fbrcm config show [key]`

Shows the effective layered configuration after applying keybinding migration
and built-in defaults. With no key, human output is TOML and includes the
complete effective key map. `--scope global` or `--scope local` instead shows
only values physically stored in that layer. Supported keys are `profile`,
`powerline_glyphs`, `keys`, `keys.<block>`, and
`keys.<block>.<action>`. A selected scalar prints as plain text; a selected
keybinding list or map prints scoped TOML. JSON is emitted only with `--json`.

Flags:

```text
--scope effective|global|local   select the view; default effective
--json                           print structured JSON
```

Full JSON output includes `scope`, `path`, `exists`, both stored source paths,
and `config`. Selected-key JSON has `key`, `value`, and `source`; effective
sources are `local`, `global`, `default`, `mixed`, or `migrated`. A missing
config file is not created. Use `fbrcm config show keys` as the authoritative
reference for every configurable keybinding block and action.

### `fbrcm config set <key> <value>...`

Sets a typed preference. It atomically replaces the global config file with
private permissions by default. `--scope local` explicitly targets the nearest
repository config, or creates `.fbrcm.toml` in the current directory when none
is found. Supported forms are:

```text
powerline_glyphs true|false
keys.<block>.<action> <key>...
```

The active `profile` is read-only here; use `fbrcm profile switch <name>` or edit
the local TOML. Only explicit overrides are stored: inherited values and
built-in defaults are never copied into the target layer. The complete effective
candidate configuration is validated before writing, including unknown
blocks/actions, empty or duplicate bindings, unsupported key names, and
conflicts with configured or default actions. Failed validation leaves the file
unchanged.

Flags:

```text
--scope global|local   select the stored layer; default global
--json                 print the scoped update result
```

### `fbrcm config reset [key]`

Removes a stored override from the selected layer. Removing a local override
reveals the global value; removing a global override reveals the built-in
default. The optional key may be `powerline_glyphs`, `keys`, `keys.<block>`, or
`keys.<block>.<action>`. With no key, it removes all stored preferences while
preserving that layer's `profile`. Reset can repair an invalid key map by
discarding the requested obsolete subtree. A changed reset asks for
confirmation; Yes is selected by default. Writes are validated and atomic.

Flags:

```text
-y, --yes             reset without confirmation
    --scope global|local   select the stored layer; default global
    --json             print the scoped reset result
```

### `fbrcm config validate`

Strictly validates TOML structure, profile references, and keybindings. By
default, it validates both stored layers and their effective merged result.
Use `--scope` to isolate a stored layer or the effective configuration. It
reports all keybinding diagnostics in stable order. Missing files are valid
because lower layers and built-in defaults apply. Invalid configuration returns
exit status 1; operational failures also return nonzero.

Flags:

```text
--scope all|effective|global|local   select validation scope; default all
--json                              print the validation report
```

Each diagnostic contains `severity`, `code`, `key`, and `message`.

### `fbrcm config edit`

Opens a staged copy of the global config in an editor by default. `--scope
local` explicitly edits the nearest repository file or creates `.fbrcm.toml` in
the current directory. After the editor exits, fbrcm strictly validates the
staged file and effective result, then atomically replaces the original only
when valid. If editing or validation fails, the original remains unchanged and
the staged path is preserved for recovery.

When the target does not exist, the editor starts with a sparse commented file.
`--full` instead stages a complete generated keybinding template. Saving that
template makes every retained entry an explicit override, so remove entries
that should continue following future built-in defaults. Normal startup never
materializes defaults into either config file.

Editor resolution order is `--editor`, `FBRCM_EDITOR`, `VISUAL`, `EDITOR`, then `vi` on Unix-like systems or `notepad.exe` on Windows. Commands may include arguments; GUI editors generally need their wait flag, for example `--editor "code --wait"`.

Flags:

```text
--scope global|local   select the stored layer; default global
--full                 stage a complete generated keybinding template
--editor <command>     override the editor command
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

Adds or replaces a [Google Cloud CLI](https://cloud.google.com/cli) ADC
identity. Run `gcloud auth application-default login` first so ADC discovery
can find credentials.

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

Binds cached projects to an auth identity. Without `--project`, every cached project is selected. Repeated project filters are ORed and use the same mode-prefixed name/project-ID matching as `projects list --filter`; auth binding is project-scoped and does not accept template prefixes.

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

Switches the global profile, creating it if needed. If repository configuration
selects another profile, the command reports that the local selection remains
effective. A profile switch performed inside the TUI remains active for that
session; repository selection applies again on the next launch.

### `fbrcm profile rename <old-name> <new-name>`

Renames an existing profile. fbrcm refuses to rename a profile selected by the
nearest `.fbrcm.toml`, because it never rewrites repository configuration as a
side effect of profile management.

### `fbrcm profile path <profile>`

Prints profile config and cache directory paths.

Flags:

```text
--json   print [{"path": "..."}, {"path": "..."}]
```

### `fbrcm profile delete <profile>`

Deletes profile config and cache directories. Confirmation defaults to yes.
Neither the global persisted profile nor the currently effective repository or
session profile can be deleted.

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
