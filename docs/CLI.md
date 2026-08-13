# fbrcm CLI

`fbrcm` is a Firebase Remote Config manager. It runs as an interactive TUI when called with no arguments. Any argument switches to CLI mode. See the [TUI guide](TUI.md) for the interactive workflow. For agents and scripts, the global `--json` flag enables the versioned [machine contract](cli-contract.md).

## Command Tree

```text
fbrcm [--help] [--version] [--profile <name>] [--no-local-config] [--json] [--timeout <duration>]
│
├── capabilities [command...]
├── schema
│   ├── list
│   └── show <schema-id>
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
│   │   └── --json
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
├── hooks
│   ├── status [--json]
│   ├── fingerprint
│   ├── trust [--yes|-y] [--json]
│   └── untrust [--json]
│
├── personalizations
│   ├── list <project> [--update] [--json]
│   └── show <project> <personalization-id> [--update] [--json]
│
├── help [command...]
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
│   │   └── --side-by-side
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
│   │   ├── --change-note <text>
│   │   ├── --yes, -y
│   │   └── --json
│   ├── aliases
│   │   ├── list [--json]
│   │   ├── set <alias> <project-id> [--yes|-y] [--json]
│   │   ├── remove <alias> [--yes|-y] [--json]
│   │   └── import --from <path>
│   │       ├── --conflict error|keep|overwrite
│   │       ├── --dry-run
│   │       ├── --yes, -y
│   │       └── --json
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

All commands support `--help`. Root also supports `--version`. With `--json`,
implicit `--help`/`-h` is represented by the separate `help` operation and its
invocation and response schemas, rather than being repeated in every command's
option schema. Version text uses the root response schema, and the root
capability publishes both `--version` and its `-v` alias.

CLI invocations do not perform a startup connectivity probe. When
`FBRCM_OFFLINE` is unset, CLI commands make only the network requests declared
by their capability metadata and report typed failures if those requests are
unavailable. Defining `FBRCM_OFFLINE`, including with an empty value or `0`,
enables CLI offline mode and suppresses network requests. Standard HTTP
requests honor `HTTPS_PROXY`, `HTTP_PROXY`, and `NO_PROXY`, including their
lowercase forms.

Every explicitly supplied positional argument must contain a non-whitespace
value. Supplied string flags and every item in a repeated string flag likewise
reject empty and whitespace-only values. Exact empty strings remain supported
only for content flags where emptiness is meaningful: `--value`,
`--description`, `--change-note`, scalar `--group`, and `--label`; a nonempty
value made only of whitespace is still invalid. Omitting an optional argument
or flag retains its documented default behavior.

A positional argument that selects an existing named resource matches only its
canonical name or ID, exactly and case-sensitively. Such selectors are not
trimmed, so surrounding whitespace is significant and normally produces no
match. Fuzzy, prefix, substring, and case-insensitive selection belongs only to
explicit query options such as `--filter`, `--search`, and `--project`.

Most commands require a selected profile before command execution. `profile`, `config`, `hooks`, `projects aliases`, `doctor`, `help`, `capabilities`, and `schema` do not. Every JSON invocation still resolves `context.profile` while building its final envelope; with no explicit or persisted effective profile, that step bootstraps the `default` profile, creates its config/cache directories, and writes the global `config.toml`. Capability metadata publishes this as `local_state_write`: commands without an unconditional local-state write use the condition `runtime_state.profile_bootstrap required`. For commands that bypass pre-execution profile initialization, this envelope-only bootstrap is best-effort: filesystem failures are logged without changing the machine outcome.

Run `fbrcm profile switch <name>` to switch or create a profile. Use the root `--profile <name>` flag or `FBRCM_PROFILE` to select an existing profile for one process without changing the persisted active profile; the flag takes precedence over the environment variable only when the command applies that option. An effective argv profile is trimmed, then must be one nonempty filesystem-safe path segment: `.`, `..`, path separators, and leading or trailing whitespace after normalization are invalid. Detailed capability flags expose `effective`; `false` means argv accepts the flag but the command does not apply it. Conditional applicability is published as `effective_when` and repeated in input schemas as `x-fbrcm-effective-when`. The input schema marks unconditional ineffectiveness as `x-fbrcm-effective: false`. `help`, `capabilities`, `schema`, `config`, `hooks`, and `projects aliases` commands currently ignore the argv `--profile` value; root version output accepts but ignores both `--profile` and `--no-local-config`. Machine-mode `auth login` ignores `--noopen`, and machine-mode `config edit` ignores `--editor`, `--full`, and `--scope` because both commands stop before reaching the corresponding human interaction. `FBRCM_PROFILE` may still supply the envelope profile because it is external process context.

At startup, fbrcm searches the current directory and every parent through the
filesystem root for `.fbrcm.toml`. The nearest match deeply overlays the global
`config.toml`: nested tables merge, while scalars and arrays replace lower-layer
values. Built-in defaults apply after both stored layers. For commands whose
`--profile` flag is effective, profile precedence is `--profile`,
`FBRCM_PROFILE`, local config, global config, then `default`; commands marked
with `effective: false` skip the argv value.
Use `--no-local-config` or set `FBRCM_NO_LOCAL_CONFIG` to ignore repository
configuration. Config commands ignore `--profile`, but honor their own explicit
configuration scope.

Native repository project aliases live only in `.fbrcm.toml` and are normally
committed with the repository:

```toml
[projects.aliases]
staging = "acme-staging-42"
prod = "acme-production-42"
```

fbrcm also discovers Firebase CLI aliases from the top-level `projects` object
in `.firebaserc`. It uses the file beside the nearest ancestor `firebase.json`,
or `.firebaserc` in the current directory when no Firebase project root exists.
Comments and trailing commas are accepted; unrelated Firebase targets and ETags
are ignored. Identical aliases in both files are deduplicated, while different
project IDs for the same alias are a configuration error. Imported Firebase
alias names must satisfy the same lowercase, selector-safe fbrcm rules.

Aliases are independent of profiles and map to physical Firebase project IDs.
Native aliases are invalid in global `config.toml`. Project synchronization,
forgetting, and registry reset never modify either repository alias file.
`--no-local-config` disables both alias sources together with the rest of the
repository overlay. Firebase CLI active-project state and its special `default`
selection behavior are not imported.

Interactive yes/no confirmations select **Yes** by default. Use the arrow keys to select No, or pass `--yes` where available to skip the prompt.

Long-running CLI work displays a gray animated progress line on stderr when
stderr is an interactive terminal. The message changes for major phases such
as project resolution, Remote Config loading, validation, publication, and
draft updates. Durable log lines temporarily replace the progress line and the
animation resumes underneath them. Progress is erased before results, diffs,
diagnostics, file pickers, editors, or confirmation prompts are displayed, and
it is never written when stderr is redirected. JSON mode defaults logging to
`silent`; an explicit `FBRCM_LOG_LEVEL` value overrides that default. Silent
logging does not suppress interactive progress in human mode.

Human-readable collection output always renders its normal table, including when there are no rows. An empty result therefore contains the table headers rather than a special empty-state message. In JSON mode, a collection is normalized under `data` as `{ "count": 0, "items": [] }`; singular resources and operation reports are objects in `data`.

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
| `FBRCM_OFFLINE` | Enable CLI offline mode whenever the variable is defined, including as an empty string or `0`. If it is unset, CLI commands perform only their declared network operations. |
| `FBRCM_LOG_LEVEL` | Set logging to `debug`, `info`, `warn`, `error`, `fatal`, or `silent`, case-insensitively. The default is `info` for human CLI/TUI use and `silent` with `--json`; an explicit value overrides either default. |
| `FBRCM_EDITOR` | Select the command used by `config edit`, after `--editor` and before `VISUAL` or `EDITOR`. Arguments are supported. |
| `FBRCM_NO_LOCAL_CONFIG` | Ignore repository `.fbrcm.toml` discovery when set to a non-empty value. The root `--no-local-config` flag provides the same behavior for one invocation. |
| `FBRCM_HOOK_TRUST` | Trust local hooks for this invocation only when the value exactly matches `fbrcm hooks fingerprint`. Intended for CI. |
| `NO_COLOR` | Disable CLI, prompt, log, and TUI colors when set to a non-empty value. |
| `COLUMNS` | Supply a positive terminal width for human-readable CLI output. Invalid values are ignored. |
| `GOOGLE_APPLICATION_CREDENTIALS` | Select an Application Default Credentials JSON file for gcloud identities and diagnostics. |
| `XDG_CONFIG_HOME` | Supply the Unix config home when `FBRCM_CONFIG_DIR` is unset; fbrcm appends `fbrcm`. |
| `XDG_CACHE_HOME` | Supply the Unix user-cache home where supported when `FBRCM_CACHE_DIR` is unset; fbrcm appends `fbrcm`. |
| `HTTPS_PROXY`, `HTTP_PROXY`, `NO_PROXY` | Configure Go's HTTP transport. Lowercase forms are also honored. |

`NO_COLOR` follows the [NO_COLOR standard](https://no-color.org/): every
non-empty value disables color, including `0`, `false`, arbitrary text, and
whitespace. It does not disable non-color terminal decoration such as bold,
faint, italic, underline, or reverse video.

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

Project filters match project display name, project ID, or any repository alias.
Alias matching follows the requested filter mode, so `=prod` is the recommended
exact selector for scripts. Parameter filters match parameter key. `--project`
and `--filter` may be repeated; repeated values are ORed and must be passed as
separate flags.
Outer Unicode whitespace is trimmed before the optional mode prefix is parsed.
The semantic schema publishes the matched resource fields, complete mode map,
and case-insensitive fuzzy, starts-with, includes, and exact matching algorithm
under `x-fbrcm-matching`. Each invocation schema also publishes selector
composition: values within a repeated selector are ORed, distinct supplied
selector sources such as `--filter`, `--group`, `--search`, and `--expr` are
ANDed, and an absent selector source matches all candidates.

### Client and Server Template Targets

Remote Config commands accept a template target wherever their command syntax shows a Remote Config `<project>` or a `--project` filter:

```text
project-id          configured template selection; implicit form
client@project-id   client template; explicit alias
server@project-id   server template in the firebase-server namespace
```

The prefix comes before a filter mode. For example, `-p 'server@=api-prod'` selects the server template of exactly `api-prod`, while `-p 'client@^mobile-'` selects client templates whose project name or ID starts with `mobile-`. Repeated flags can mix client and server targets in one invocation. Target prefixes are recognized case-insensitively and canonicalized to lowercase. Query flags trim outer whitespace and whitespace around the project query after an explicit prefix. Positional target selectors preserve the project name or ID exactly without trimming. Explicit `client@` remains distinct from an unqualified target during selection, though both canonicalize to the same client target identity.

Each cached project stores its enabled template selections and one primary template. New and existing projects default to client-only. An unqualified bulk filter, or no `--project` filter, expands every matched project to its configured enabled templates. An unqualified positional `<project>` selects that project's primary template. Explicit `client@` and `server@` prefixes always select exactly that template, independently of the saved selections. Target-aware matching annotations publish all three rules and client-target canonicalization to the unqualified project ID. Invocation schemas for bulk `--project` commands additionally publish their no-filter default over all configured projects and enabled templates.

The target syntax applies to `add`, `get`, `update`, `delete`, `duplicate`, and `groups`; all `conditions` and `versions` commands; `draft` commands; `project export`, `project import`, and `project defaults`; and the source and destination of `projects diff` and `projects promote`.

Project metadata and managed-feature commands remain project-scoped rather than template-scoped. In particular, `projects list`, `projects update`, `projects forget`, `project show`, `project templates`, `project open`, `experiments`, `rollouts`, `personalizations`, `auth bind`, and `doctor` continue to accept ordinary project IDs or names without a template prefix. `project show` and `project open` do not parse target prefixes: characters such as `client@` or `server@` are part of their literal project query. Template-preference and managed-feature commands explicitly reject recognized `client@` and `server@` target syntax. A managed-feature `server@` target reports that managed features support only the client namespace, while a `client@` target asks you to omit the unnecessary prefix.

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

1. Exact case-sensitive project ID.
2. Exact case-sensitive repository alias.
3. Exact case-sensitive project display name.

Filter mode prefixes do not apply to positional project arguments. Leading
`=`, `^`, `/`, and `~` characters are part of the literal project query. The
query is not trimmed; a case mismatch, substring, or surrounding whitespace
does not select the resource.

A single match is selected. Multiple exact display-name matches print only the
ambiguous projects and return an error. No match prints the known-project
table and returns an error. Exact ID always wins over a colliding alias, and an
alias wins over a colliding display name. Once a configured alias is recognized,
an unavailable target reports the alias, canonical ID, and selected profile
instead of falling through to another match.

Aliases name physical projects and compose with template prefixes. `prod` uses
the active profile's configured primary template, `client@prod` selects the
client template, and `server@prod` selects the server template. Operational
output, caches, drafts, retry filters, and API requests continue using canonical
target IDs.

Draft commands resolve only locally stored drafts and never synchronize projects
as a side effect. An explicit prefix selects that template kind. An unqualified
query or alias selects the configured primary template when the project is still
registered, and falls back to the client template for an unregistered project.
Aliases therefore remain usable for drafts that outlive the project registry.
This also permits `show --raw` and `discard` for drafts whose project is no
longer present in the projects cache.

### Parameter Search

Parameter-context commands also support `--search <text>`. It searches parameter name, description, default value, conditional values, condition names, and condition expressions. Name/description/condition-name matching is case-insensitive and ignores punctuation; value/expression matching is case-sensitive. `--search` is ANDed with `--filter` and parameter-context `--expr`.

### Expression Filters

`--expr` uses expr-lang and must evaluate to boolean. See [EXPR.md](EXPR.md) for full context fields and helper functions.

Expression errors are never treated as an empty successful result. A syntax or
type error fails before filtering begins. If evaluation fails for a project,
parameter, condition, diff entry, or import entry, the command reports the
affected target and returns nonzero. Multi-target `update` and `delete`
continue processing independent targets, record an evaluation failure as
`preparation-failed` for that target, and return nonzero after printing all
results; no candidate is published or drafted for the failed target.

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

As an experimental human CLI convenience on supported systems, `get` also accepts a directory passed as stdin. It reads top-level `.json` files from that directory, accepts raw Remote Config JSON or fbrcm cache JSON in each file, and treats each file stem as a canonical template target. An unqualified stem is a client target and a `server@` stem is a server target. Project name is built from the underlying project ID by splitting on `-` and `_`, then capitalizing words. This transport is intentionally outside the versioned `--json` input schemas and capability metadata.

`project import` reads JSON from `--from`, stdin, or, in human mode, an interactive `.json` file picker. JSON mode returns `interaction.required` when neither `--from` nor redirected stdin supplies the input. It accepts raw Remote Config JSON or an fbrcm parameters cache JSON file with `remote_config`, then applies the stricter local Remote Config validation required by import before selection and transformation.

In stdin transformation mode, `add`, `update`, and `delete` emit raw transformed
JSON in human mode and an artifact DTO in JSON mode. `--project`, `--dry-run`,
`--draft`, and `--change-note` are unavailable because piped input has no
persistent target project identity and is always transformed in memory.

### Draft lifecycle and write safety

Drafts are profile-scoped, target-specific, self-contained records. Each version-1 record stores the working Remote Config, its immutable base Remote Config, base version and ETag, timestamps, and an optional `change_note`. A project can therefore have independent client and server drafts. Plain Remote Config JSON is not accepted as an on-disk draft format, and no legacy draft migration or fallback is performed.

`add`, `duplicate`, `update`, `delete`, group and condition mutations, `project import`, `projects promote`, `draft publish`, and `versions restore` accept `--change-note <text>`. The note is trimmed, must be one line without control characters, and is sent as Firebase `version.description` on ordinary publication. Invalid change-note input is a typed `argument.invalid` failure with exit status 2. With `--draft`, it is stored as `change_note` and remains editable with `draft change-note`; an explicitly empty note clears it. Native `versions rollback` intentionally has no change-note flag because Firebase owns rollback metadata.

`add`, `duplicate`, `update`, `delete`, `project import`, and the condition mutation commands accept `--draft`. In draft mode they apply changes on top of an existing project draft or create a new draft from freshly revalidated Remote Config. They do not validate or publish to Firebase. Combining `--draft` with `--dry-run` previews the change without writing either draft or Firebase state.

A publication `--dry-run` performs the real Firebase validation-only `PUT`
against the current ETag, then suppresses the separate publication `PUT` and
all local state writes. It therefore requires Firebase credentials and network
access and can fail on the same template validation rules as a real publish.
Results expose `validated` and `validation_source`: `firebase` means Firebase's
validation endpoint was used, while `local` means only local candidate
validation was applicable, as with draft previews and no-op results. A failed
validation reports `validated: false` and returns nonzero.

Immediate Remote Config writes refuse to proceed when the target has an unpublished draft. This guard applies to add, duplicate, update, delete, condition mutations, project import, version rollback/restore, and project promotion. Resolve the draft with `draft publish` or `draft discard`, or add the intended mutation to it with `--draft`.

Multi-project Remote Config publication is non-atomic: Firebase accepts a separate validated write for each project. Commands process every selected project even when an independent project fails, collect one outcome per project, and print the complete `Results:` block after operation logging has finished. They return nonzero after the batch if any project failed. Successful projects are not rolled back. Conflicts are reported for a fresh explicit retry instead of silently recalculating and publishing a different candidate. Failed-project output includes exact `-p '=project-id'` filters for retrying only projects that were not published.

### Publication hooks

Global `config.toml` and repository `.fbrcm.toml` files may define commands that
run around every Remote Config publication:

```toml
[hooks]
timeout = "2m"
pre_publish = [
  "./scripts/validate-parameter-names.sh",
  "go run ./tools/validate-remote-config",
]
post_publish = ["./scripts/notify-remote-config-published.sh"]
```

Commands run sequentially through `/bin/sh -c` on Unix-like systems and
`cmd.exe /S /C` on Windows. The timeout applies to each command and defaults to
five minutes. An empty command, invalid duration, or non-positive duration
makes the configuration invalid. Hook arrays follow normal configuration merge
rules: a local array replaces the corresponding global array rather than
concatenating with it.

`pre_publish` runs after the final candidate and Firebase validation have been
prepared, immediately before the publication request. A nonzero exit, timeout,
signal, or shell error prevents publication. It also runs for publication
`--dry-run` operations, allowing CI to exercise policy checks. `post_publish`
runs only after Firebase accepts a real publication; it does not run for dry
runs or failed writes. A failed post-hook cannot undo the publication, so fbrcm
reports the operation as published with a post-hook failure and returns nonzero.
Do not blindly retry that outcome because doing so can create another Remote
Config version.

Saving or previewing a draft does not run hooks. Publishing a draft does.
Direct mutations, imports, promotions, version restores, native rollbacks, and
their TUI equivalents use the same publication lifecycle. No-op operations do
not run hooks. Multi-target commands run hooks sequentially for each target and
continue with independent targets after a target-specific failure.

For each publication fbrcm creates a private temporary directory containing
read-only `current.json` and `candidate.json` files plus `context.json`.
Post-hooks also receive `published.json`. Hooks validate or perform side
effects; editing these files never transforms the candidate sent to Firebase.
The directory is removed after the operation.

Every hook inherits the fbrcm process environment and receives:

```text
FBRCM_HOOK_EVENT       pre_publish or post_publish
FBRCM_OPERATION        originating operation
FBRCM_TARGET           canonical client or server template target
FBRCM_PROJECT_ID       physical Firebase project ID
FBRCM_TEMPLATE_KIND    client or server
FBRCM_PROFILE          selected profile
FBRCM_DRY_RUN          true or false
FBRCM_CHANGE_NOTE      publication change note, possibly empty
FBRCM_CONFIG_FILE      config file supplying this hook array
FBRCM_PROJECT_DIR      directory containing that config file
FBRCM_CURRENT_FILE     current.json path
FBRCM_CANDIDATE_FILE   candidate.json path
FBRCM_CONTEXT_FILE     context.json path
FBRCM_PUBLISHED_FILE   published.json path, empty before publication
GCLOUD_PROJECT         alias of FBRCM_PROJECT_ID
PROJECT_DIR            alias of FBRCM_PROJECT_DIR
```

The command working directory is the directory containing the config file that
supplied its effective hook array. Hook output is written to stderr in CLI mode,
preserving JSON stdout, and to the Logs panel in TUI mode.

Global hooks are trusted automatically. Repository hooks execute local code and
must be trusted explicitly with `fbrcm hooks trust`. Trust is stored outside the
repository as the canonical local config path plus a SHA-256 fingerprint of the
effective hook commands and timeout. Changing the path, commands, or timeout
invalidates trust. Noninteractive publication with untrusted hooks fails before
any Firebase write. CI can pin the exact value printed by `fbrcm hooks
fingerprint` in `FBRCM_HOOK_TRUST`; a mismatched value fails closed.

### Versioned JSON automation contract

Every command accepts the global `--json` flag. Successes and failures emit one versioned envelope; tables, prompts, file pickers, editors, browser launches, unwrapped usage text, and unwrapped raw content are excluded from JSON stdout. Requested help and root version text are returned inside `data.text` using their documented response schemas. See the [complete machine contract](cli-contract.md) for schemas, error categories, semantic exit statuses, discovery, non-interactive rules, and artifact results.

Failure exit statuses are output-format independent: the same invocation uses
the same semantic status with or without `--json`. Human mode changes only the
presentation of the failure.

If any Firebase command encounters an OAuth identity that still needs browser
authorization, JSON mode returns `interaction.required` with a safe
`auth login <auth-id>` remediation before creating a listener or launching a
browser.

Every structured error and warning remediation declares how its non-empty
`argv` must be used: `retry_with_arguments` augments the original invocation,
`replace_selector` substitutes an exact selector, and `run_command` is a
complete fbrcm subcommand argument vector. Agents should branch on that
`strategy` instead of interpreting remediation text.

Direct Remote Config mutations put an ordered target result collection in the envelope's `data` field:

```json
{
  "count": 1,
  "items": [
    {
      "target": "my-project",
      "status": "published",
      "changed_item_count": 1,
      "previous_version": "41",
      "published_version": "42",
      "draft": false,
      "dry_run": false,
      "validated": true,
      "validation_source": "firebase",
      "selection": {
        "default_scope": true,
        "resolved_target_count": 1,
        "matched_item_count": 1
      },
      "no_op_reason": null,
      "change_note": "Enable checkout v2",
      "error": null,
      "retry_selector": null
    }
  ]
}
```

`validated`, `validation_source`, and `selection` are always present. `selection.default_scope` reports whether the command used its unqualified default project scope, `resolved_target_count` reports its target breadth, and `matched_item_count` reports the selected items in this target. For `status: "unchanged"`, `changed_item_count` is zero and `no_op_reason` distinguishes `no_match` from `already_applied`; changed and failed states use their status-specific count rules and have a null no-op reason. Drafted, would-draft, and unchanged results use local validation provenance; published, would-publish, and post-publication failures use Firebase provenance. Validation, publication, conflict, preparation, and draft failures constrain `validated`, `validation_source`, and `error.stage` to the phase actually reached. `previous_version`, `published_version`, and `change_note` are `null` when unavailable or omitted. A target-level `error` is either `null` or a structured object with a stage and a redacted, bounded message. A failed target that is safe to retry includes an exact, target-aware `retry_selector`, such as `=my-project` or `server@=my-project`; pass it back as `--project <retry_selector>`. Batch envelope errors additionally retain typed per-target codes, categories, retryability, details, and remediation under `errors[].details.failures`. An all-failed batch uses its first target's category for the process status and is retryable only when every failed target is retryable. A `published-cache-failed` target has no retry selector because Firebase was already updated and the correct recovery is a cache refresh. Envelope warnings carry structured non-fatal publication/cache/hook conditions and safe remediation argv. JSON mode does not imply `--yes`: when confirmation would be required, the command returns structured `interaction.required` instead of prompting. In stdin transformation mode, `add`, `update`, and `delete` wrap the transformed Remote Config as an artifact in `data`; `--change-note` is unavailable because no publication or draft write occurs.

If Firebase accepts a publish but the returned state cannot be saved locally, the outcome is reported as `published-cache-failed`, not as an unpublished project. Refresh that project's cache instead of blindly retrying the mutation. For coordinated changes, `--draft` provides reviewable and recoverable intent, but publishing those drafts is still non-atomic across projects.

Draft publish always fetches current Firebase state, performs a three-way merge from base, draft, and current, validates using the current ETag, and publishes only the exact candidate that was previewed. Conflicts preserve the local draft. Successfully published or already-applied drafts are removed locally. A publish that succeeds remotely but cannot remove its local record reports `published-cleanup-failed`; rerunning recognizes the already-applied content and retries cleanup without creating another version. All `published-hook-failed`, `published-cache-failed`, and `published-cleanup-failed` results report `validated: true` and `validation_source: "firebase"`, because Firebase validation and publication completed before the local post-publication failure.

## Commands

### `fbrcm`

With no arguments, opens TUI. With arguments, executes CLI command.

Flags:

```text
-h, --help                  show root help
-v, --version               print version, commit, and build date
    --profile <name>        use an existing profile for this invocation without changing the active profile
    --no-local-config       ignore repository configuration
    --json                  emit one machine-contract v1 envelope
    --timeout <duration>    limit the complete command
```

`--profile` defaults from `FBRCM_PROFILE`. It applies to CLI commands that use profile state. Help, contract metadata, configuration, hooks, and project-alias commands accept but ignore it; root `--version` also ignores both `--profile` and `--no-local-config`. `FBRCM_PROFILE` selects and pins the profile when starting the TUI with no arguments; restart without it to create or switch profiles interactively.

`--json` and `--timeout` also apply to every CLI subcommand. Use `fbrcm capabilities --json` for machine-readable command discovery and `fbrcm schema list --json` to enumerate the embedded JSON Schemas.
Every JSON envelope includes both `command`, identifying the published response
contract, and `requested_command`, preserving the caller's requested operation.
They differ for an unknown nested command: the failure uses the published
`root` response schema while `requested_command` retains the attempted path and
the `argument.unknown_command` problem carries `details.kind: "invocation"`.

### `fbrcm capabilities [command...]`

Lists all executable command capabilities, or describes the command at the
exact argv path. In human mode the full listing contains stable command IDs and
summaries; a single-command lookup prints its ID. In JSON mode the full listing
is a compact index containing IDs, paths, summaries, schema URNs, side-effect
levels, and destructive markers. An exact single-command lookup returns the
detailed capability record, including arguments, flags, response and shared
error-schema URNs, side effects and their conditions, network access and its
conditions, typed destructive conditions plus explanatory reasons,
idempotency and retry-safety conditions, dry-run/draft/stdin support and stdin
transport modes, and interaction rules and conditions for JSON invocation.
Each flag record includes `effective`, distinguishing an option that changes
this command from one that argv accepts but the command ignores. In
particular, `project open --json`
returns a URL without launching a browser, OAuth login declares conditional
authentication network access and reports browser authorization only when
human action is required, and draft publication is marked destructive.
For conditional network access, `network_when` exposes OR clauses containing
typed AND predicates over options, stdin, envelope context, and a closed
runtime-state vocabulary. The capability schema's
`x-fbrcm-runtime-state-semantics` records define every allowed name/operator
pair, including cache freshness and stale-fallback behavior. Predicate values and flag defaults retain their JSON
types. `project import` is correctly marked as requiring Firebase even when its
document comes from stdin; cache-only version and project comparisons use the
`cached=false` predicate; historical version lookup also reports whether its
selector/cache state requires network resolution, and draft diff additionally
requires `against=current` before it can contact Firebase.
`side_effect_when` gives one condition record per declared side effect.
Empty side-effect, behavior-condition, destructive-condition, idempotency,
stdin-mode, and interaction-condition lists are
arrays rather than `null`; the detailed record validates against the published
capability schema. That schema enumerates the exact published records, so
cross-command combinations of IDs, paths, flags, URNs, effects, and predicates
are rejected.
The same schema defines every side-effect value under
`x-fbrcm-side-effect-semantics`. Draftable mutations separately declare
Firebase reads, Firebase validation, confirmation-authorized publication,
local draft writes, cache updates, and trusted-hook execution. Pre-publish
hooks can execute during dry-run; post-publish hooks require Firebase
acceptance. OAuth and service-account imports declare a conditional
`local_file_write` for their credential file. Every command that may construct
a Firebase client also declares conditional authentication remote access,
possible non-dry OAuth token-file persistence, and OAuth human-authorization
interaction. Doctor declares only the remote authentication effect because it
does not persist refreshed credentials or authorize interactively. Interaction
behavior is a stable enum rather than free-form wording. Command-specific
missing-input and selection clauses are preserved alongside confirmation
conditions, and confirmation is reported only when the planned operation
actually requires it. Local writes and output-file writes likewise carry
change/authorization predicates, post-publish hooks require accepted
publication, and idempotency distinguishes stdin transformations and end-state
local writes—including OAuth token persistence by `auth login`—from unsafe
remote retries. Dry-run retry safety additionally
depends on whether a trusted hook actually executed.
Every JSON command declares the conditional default-profile bootstrap write.
Commands that resolve a project through the live registry additionally declare
`runtime_state.project_registry sync_write_succeeded`, because a missing or
empty registry is synchronized from Firebase and persisted before the requested
operation. `config edit --json` declares no editor or destination-file write
because it stops with `interaction.required` before opening an editor; its only
possible write is the shared envelope profile bootstrap. Its `--editor`,
`--full`, and `--scope` flags are marked ineffective for JSON invocation. Every `auth add`
variant declares its replacement or credential-removal risk as destructive and
conditionally declares `local_file_delete`. Profile deletion separately
declares config-file, cache, and draft deletion; project registry reset
declares both the local state mutation and registry-file deletion. Draft
publication declares draft cleanup both after an accepted changed publication
and when a non-dry unchanged draft is removed.
`draft show --to --json` is not marked destructive because it cannot
overwrite: an existing destination produces `interaction.required`.
Unknown paths fail with `command.not_found`; command groups
that are not executable fail with `command.not_executable`.
The reserved lookup `capabilities root` returns the executable root operation,
whose command path is the empty array.

Use command path components, not the dot-separated ID, for a lookup:

```sh
fbrcm capabilities project import --json
```

### `fbrcm schema list`

Lists the `$id` URNs of all embedded machine-contract schemas. In JSON mode,
the identifiers are returned in `data.items` with `data.count`.

### `fbrcm schema show <schema-id>`

Prints one embedded Draft 2020-12 JSON Schema. Without `--json`, the schema
document itself is printed. With `--json`, the schema document is placed in the
envelope's `data` field. The ID is matched exactly and case-sensitively; an
unknown ID returns `schema.not_found`. Command response schemas define the exact success-data
DTO, including `{count, items}` and the complete item shape for collection
commands, and reference the published semantic schema for shared enums and
grammars. The shared error schema catalogs current problem codes,
discriminates typed `details` objects, and the envelope schema enforces
category-to-exit-status consistency. Separate stdin schemas describe OAuth and
service-account credentials, general Remote Config transformations, and the
stricter locally validated project import payload. A sole OAuth `installed` or `web` client object
requires a nonblank client ID and secret, absolute parseable endpoint URIs, and
at least one redirect URI; the schema also represents the runtime's split
selection when both objects are supplied. Service-account input publishes its
runtime email and absolute-URI parser rules. Remote Config values enforce exactly one known or
opaque future value option. Artifact schemas enforce
encoding/content/destination correlations and specialize known inline JSON:
Remote Config transformations and version exports expose the Remote Config
schema, project exports preserve any syntactically valid JSON response, and
raw draft recovery accepts any JSON value from the stored bytes.
Command schemas also limit artifacts to their reachable encodings, require a
non-empty target, and require `draft show` artifacts to be non-overwriting in
JSON mode. Response schemas admit `partial_success` only for Remote Config
publication commands and restrict warning codes to the commands that can emit
them; commands with no warning path require an empty warning array.
Response discriminators such as auth type, configuration and condition source,
draft comparison base, version operation, template kind, and managed-feature
kind use closed enums or command-specific constants. Auth path objects enforce
their per-type files and identity/type correlations. Project template results
require a nonempty unique `client`/`server` set containing the primary template.

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
--dry-run                  preview the requested mutation without applying it
--draft                    save changes to local drafts instead of publishing
--change-note <text>       set the change note for publication or draft storage
-y, --yes                  print diff and add without confirmation
--json                     print structured mutation results
--description <text>       parameter description
--group <name>             add parameter inside group
```

Remote mode loads projects, filters them, and adds the parameter where it does not already exist. It prints the complete Remote Config diff and asks for confirmation for each project unless `--yes` is set, then validates and publishes each confirmed project independently. Existing parameters are skipped.

With `--draft`, each confirmed mutation is stored locally on top of any existing draft. Without `--draft`, the command refuses projects that already have unpublished drafts.

Stdin mode reads Remote Config JSON from stdin and adds the parameter to that
JSON. Human mode prints the final JSON; JSON mode returns it as an artifact DTO.
It also accepts an fbrcm parameters cache JSON file and reads its internal
`remote_config` field. `--project`, `--dry-run`, `--draft`, and `--change-note`
are unavailable in this mode.

### `fbrcm duplicate <source> <target>`

Duplicates one complete Remote Config parameter in every matched project. The copy preserves the source group, description, value type, default value, conditional values, and condition references. Source and target names are nonblank, limited to 256 code points, and must differ exactly and case-sensitively. The positional source is untrimmed and matches parameter keys exactly and case-sensitively across the root and all groups. A project without that exact source is skipped; the same exact source key in more than one location is ambiguous. The new target name is trimmed, and its collision check is exact and case-sensitive, so differently cased parameter keys remain distinct; an exact collision is never overwritten.

Flags:

```text
-p, --project <query>   filter template targets; may be repeated
--expr <expr>           filter target projects with project context
--dry-run               preview the requested mutation without applying it
--draft                 save changes to local drafts instead of publishing
--change-note <text>    set the change note for publication or draft storage
-y, --yes               print diff and duplicate without confirmation
--json                  print structured mutation results
```

Remote mode prints the complete Remote Config diff and prompts before each duplicate unless `--yes` is set. It validates and publishes each project independently. A conflict fails that project without suppressing later projects, and the final command status is nonzero when any project fails. Project filters are applied before the project-context expression, matching `add` behavior.

With `--draft`, duplication composes onto each existing draft and remains local. Without `--draft`, a project with an unpublished draft fails independently while other selected projects continue. The command does not use stdin transformation mode.

### `fbrcm get [parameter]`

Prints Remote Config parameters across template targets.

Passing `[parameter]` selects that canonical parameter key exactly and case-sensitively across the root and groups. It cannot be combined with `--filter` and must be nonblank. Runtime does not trim this argv value; a case mismatch or surrounding whitespace returns no rows. Explicit `--filter` retains its documented case-insensitive query semantics.

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

Default output is a terminal table. Firebase-managed values use compact human summaries. Personalizations render as `◈ (personalization)`, and unknown future value options render as `(optionName)`. Experiments with published template variants render as `⚗ 15% : true | false | true`, or as `⚗ true | false` when the template has no exposure percentage; unavailable or incomplete variant data falls back to `⚗ (a/b test)`. Rollouts render their published value without another API request as `◐ 10% → 20 | (no change)`. The vertical bar groups variants within one managed value and remains distinct from the slash that separates collapsed conditional values. Percentages use the shared gold count color, concrete managed values use the parameter type's value color, and placeholders such as `(no change)` and `(empty string)` use the same muted gray as the surrounding chrome. JSON output includes the same unstyled summaries together with project, project ID, group, key, description, default value, conditionals, type, version, cache time, and status. Status is `fetch` for freshly fetched or just-verified data, `cached` for a usable cache, `stale` for an expired fallback, `missing` when no data is available, and `error` when cached data is shown alongside a load error.

Stdin mode reads Remote Config JSON from stdin and queries only that config. It also accepts an fbrcm parameters cache JSON file and reads its internal `remote_config` field. As an experimental human-only convenience on supported systems, a directory on stdin makes `get` read top-level `.json` files as multiple projects; this transport is not published in machine schemas or capability metadata.

### `fbrcm update [parameter]`

Updates matched Remote Config parameters. Passing `[parameter]` selects that canonical parameter key exactly and case-sensitively across the root and groups. It cannot be combined with `--filter`; when present, it must be nonblank and no longer than 256 code points. Runtime does not trim this argv value, and the input schema publishes literal positional matching with no whitespace normalization. A case mismatch or surrounding whitespace yields a no-op selection. Explicit `--filter` retains its documented case-insensitive query semantics.

Flags:

```text
-p, --project <query>      filter template targets; may be repeated
-f, --filter <query>       filter parameters; may be repeated
--expr <expr>              filter parameters with parameter context
--search <text>            search parameter names, descriptions, values, and conditions
--dry-run                  preview the requested mutation without applying it
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

Stdin mode reads Remote Config JSON from stdin and updates matching parameters.
Human mode prints the final JSON; JSON mode returns it as an artifact DTO. It
also accepts an fbrcm parameters cache JSON file and reads its internal
`remote_config` field. It does not prompt. `--project`, `--dry-run`, `--draft`,
and `--change-note` are unavailable in this mode.

### `fbrcm delete [parameter]`

Deletes matched Remote Config parameters. Passing `[parameter]` selects that canonical parameter key exactly and case-sensitively across the root and groups. It cannot be combined with `--filter`; when present, it must be nonblank and no longer than 256 code points. The value is not trimmed; a case mismatch or surrounding whitespace yields a no-op selection. Explicit `--filter` retains its documented case-insensitive query semantics.

Flags:

```text
-p, --project <query>   filter template targets; may be repeated
-f, --filter <query>    filter parameters; may be repeated
--expr <expr>           filter parameters with parameter context
--search <text>         search parameter names, descriptions, values, and conditions
--dry-run               preview the requested mutation without applying it
--draft                 save changes to local drafts instead of publishing
--change-note <text>    set the change note for publication or draft storage
-y, --yes               print diff and delete without confirmation
--json                  print structured mutation results
```

Remote mode prints diffs and prompts unless `--yes` is set. It validates and publishes each project independently, reports every outcome, continues after project-scoped failures, and returns nonzero if any project failed.

With `--draft`, deletions are saved locally on top of any existing draft. Without `--draft`, a project with an unpublished draft fails independently while other selected projects continue.

Stdin mode reads Remote Config JSON from stdin and deletes matching parameters.
Human mode prints the final JSON; JSON mode returns it as an artifact DTO. It
also accepts an fbrcm parameters cache JSON file and reads its internal
`remote_config` field. It does not prompt. `--project`, `--dry-run`, `--draft`,
and `--change-note` are unavailable in this mode.

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

Human output prints project/version/source context followed by a terminal-width-aware table containing priority, color-styled name, usage count, and expression. Long expressions are cropped with an ellipsis. In JSON mode, `data.items` contains the condition objects without repeated project/version/source context and `data.count` contains their count.

### `fbrcm conditions show <project> <condition>`

Shows one condition and every parameter value that uses it. The positional condition name is untrimmed and must match its canonical name exactly and case-sensitively.

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

For `edit`, `rename`, `move`, and `delete`, positional `<condition>` is
untrimmed and must match the canonical condition name exactly and
case-sensitively. A mismatch returns `condition.not_found`.

All five commands support:

```text
--dry-run   preview the requested mutation without applying it
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

The portable machine input contract limits `--priority` to 2,147,483,647, then
runtime validation limits an explicit nonzero priority to the existing
condition count plus one. Zero appends the new condition.

`edit` requires at least one of:

```text
--expression <expr>   replace the raw Firebase condition expression
--color <color>       replace the Firebase display color
--no-color            remove the display color
```

A truthy `--no-color` and `--color` are mutually exclusive. Explicit
`--no-color=false` is treated as absent, so it may coexist with `--color` but
does not remove a color or satisfy `edit`'s required edit selection. Supported
colors are `BLUE`, `BROWN`, `CYAN`, `DEEP_ORANGE`, `GREEN`, `INDIGO`, `LIME`,
`ORANGE`, `PINK`, `PURPLE`, and `TEAL`; input is normalized
case-insensitively. Imported condition objects accept only Firebase's `name`,
`expression`, and `tagColor` fields; unsupported fields are rejected.

`rename` updates the condition definition and every conditional-value reference to it. `move` inserts the complete condition at the requested 1-based priority and reports how many conditions and parameters may be affected by the priority change. Its current `strconv.Atoi` parser accepts an optional leading `+` and leading zeroes, but rejects zero, negative values, non-decimal text, and machine-integer overflow; runtime validation also rejects priorities above the existing condition count. `delete` removes the condition and its conditional values; parameters left without any value may also be removed, and the command reports that impact before confirmation.

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

All group commands support repeatable `--project|-p` target filters with the same mode prefixes and OR behavior as `get`, `add`, `delete`, and `update`. With no project filter, they process every configured enabled template in stable project-name/target-ID order. The positional `<group>` used by `edit`, `rename`, and `delete` is untrimmed and must match the canonical group key exactly and case-sensitively; targets without that exact key are skipped. `add` and the rename destination create names rather than select existing groups, so those new names are trimmed and validated separately. Differently cased group keys remain distinct.

All group mutations also support `--dry-run`, `--draft`, `--change-note`, `--yes|-y`, and `--json`, with the same diff, confirmation, validation, ETag, draft-composition, draft-conflict, and structured-result behavior as condition mutations. A truthy `--no-description` and `--description` are mutually exclusive. Explicit `--no-description=false` is treated as absent, so it does not clear a description or satisfy `edit`'s required edit selection.

### `fbrcm draft list`

Lists drafts in the active profile without contacting Firebase. Invalid draft envelopes remain visible instead of failing the complete listing.

Flags:

```text
-f, --filter <query>   filter by optional client@ or server@ project query; may be repeated
--json                 print structured JSON
```

Human output includes canonical target ID, project name, base version, update time, parameter/condition change counts, status, and the optional Change Note as the final column. Status is `ready`, `unchanged`, or `invalid`.

JSON entries include `project_id`, `project`, `base_version`, `created_at`, `updated_at`, byte size, status, validity, base availability, path, change counts, and `change_note`.

All draft selectors operate only on existing local drafts. Positional draft
selectors are untrimmed and resolve exactly and case-sensitively by physical
project ID, then repository alias, then display name. For positional selectors,
`~`, `^`, `/`, and `=` are literal characters rather than mode prefixes. An
optional `client@` or `server@` prefix selects that template; an unqualified
selector uses the configured primary template, or client when the project is no
longer registered. Zero and multiple exact display-name matches return typed
`draft.not_found` and `draft.ambiguous` problems. `draft list --filter` remains
an explicit case-insensitive mode-prefixed query over the existing draft set
and only includes configured enabled templates for an unqualified match, with
the same unregistered client fallback.

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

`--raw` bypasses draft decoding, so it can recover an invalid or damaged envelope. File output is forced to mode `0600`. A new destination is created exclusively so a concurrent file cannot be overwritten. An existing destination requires the normal Yes-defaulted confirmation; JSON mode returns `interaction.required` because this command intentionally has no confirmation-bypass option. JSON artifact metadata reports `overwritten: true` only when the command actually replaced an existing destination.

### `fbrcm draft change-note <project> [text]`

Sets, replaces, or clears the optional note stored with one draft without changing its Remote Config candidate.

```text
--clear   clear the stored note; mutually exclusive with [text]
--json    print project_id and change_note
```

The note must be a single line. An empty `[text]` also clears it. Draft format remains version 1; the field is stored as `change_note`. Querying the note has no draft-write side effect. Supplying `[text]` or `--clear` updates the draft file and its `updated_at` timestamp, including when the resulting note text is unchanged; capability metadata therefore marks the update forms as non-idempotent local draft writes.

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
```

`--against base` compares immutable base to stored draft and is entirely local. `--against current` fetches current Firebase state, performs the same three-way merge used by publish, and compares current to the effective candidate. `--cached` makes that second operation local but does not claim the cached snapshot is still current.

`--parameters` and `--conditions` are mutually exclusive. Condition ordering changes are included in human and JSON diffs.

Differences return status 1 and no differences return status 0. Operational
failures use the global semantic failure status, such as 2 for invalid
arguments, 3 for configuration or profile failures, and 8 for invalid
expressions. The behavior is identical in human and JSON modes. The status
describes the filtered result when selection flags are present.

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

If current Firebase state already contains the effective draft changes, no new
version is created. A live run removes the redundant draft and reports
`already-applied`; a dry run preserves it and reports `unchanged`. Batch mode is
non-atomic, continues after independent project failures, prints its collected
results together at the end followed by a targeted retry command, and returns
nonzero if any item failed.

In JSON mode, `data.items` contains one result per target and `data.count` contains their count. Each result includes project ID, status, base/previous/published versions, `rebased`, `changed`, `draft_deleted`, `dry_run`, `validated`, `validation_source`, `change_note`, and an optional target error. Status values include `unchanged`, `published`, `would-publish`, `already-applied`, `failed`, `conflict`, `published-hook-failed`, `published-cache-failed`, and `published-cleanup-failed`. The three `published-*-failed` statuses report `validated: true` and `validation_source: "firebase"`: Firebase accepted the validated publication, while their errors identify the failed local hook, cache, or cleanup stage. Required confirmation returns an envelope-level `interaction.required` problem instead of prompting.

### `fbrcm draft discard [project...]`

Deletes one or more local drafts without contacting Firebase. Use `--all` instead of positional projects to process the complete active profile.

Flags:

```text
--all          discard every active-profile draft
-y, --yes      skip destructive confirmations
--json         print structured results
```

Human mode prints the local `base → draft` diff before confirmation. Invalid drafts warn that preview is unavailable but can still be explicitly discarded. Naming a nonexistent draft is an error; `--all` with no drafts is a successful no-op.

In JSON mode, `data.items` contains one status result per selected project and `data.count` contains their count.

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

Export normalizes JSON by unescaping `<`, `>`, `&`, trimming trailing line breaks, and ordering numeric conditional value keys before non-numeric keys. For a file destination, artifact `size_bytes` and `sha256` describe those exact written bytes. For inline JSON, they describe the contract-normalized, compact, HTML-safe, order-preserving serialization of `json_content`; envelope indentation is excluded. When `--to` names an existing file, export asks before replacing it unless `--yes` is set. A destination created after the initial check is not overwritten without authorization.

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

The machine input schema publishes the same file-before-stdin selection rule;
a later stdin document is not consumed when `--from` is present.

Import input may be raw Remote Config JSON or an fbrcm parameters cache JSON
file with `remote_config`. Before filtering or other import transformations,
the decoded template must pass the same local
`NormalizeRemoteConfigForUpdate` checks used by the runtime, including valid
condition metadata, parameter value types, unique condition names, and valid
group and parameter structure. This is published by the dedicated
`stdin:remote_config_import` schema; it is intentionally stricter than the
stdin schema for in-memory `get`, `add`, `update`, and `delete` transformations.

Flags:

```text
--from <path>                            read Remote Config JSON from file
--group <name>                           import only named group; may be repeated
-f, --filter <query>                     import only matching parameter keys; may be repeated
--expr <expr>                            import only parameters matching parameter context expression
--search <text>                          import only parameters matching rich search text
--dry-run                                preview the requested mutation without applying it
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
Every requested `--group` must exist in the imported source. A missing group is
reported as `group.not_found` with the requested names and the available source
groups in selection details.

If current config is empty, import replaces it. If current config has content and neither `--merge` nor `--override` is set, command prompts for strategy. Merge adds missing conditions, groups, and parameters. Conflicting condition, group description, or parameter values prompt unless `--merge-resolve` is set. `--yes` skips only the final confirmation; automated imports should also specify `--merge` or `--override` and, when needed, `--merge-resolve`.

After import transform, the CLI reports how many source conditions are kept and removed. `--keep-portable-conditions-only` removes conditions tied to destination-specific resources such as Analytics audiences or user properties, experiments, Firebase App IDs, custom signals, and installation IDs. Unused conditions and unknown condition references are also removed. Groups that become empty are preserved, including their descriptions; only an explicit group-level selection or replacement removes a group. Normal mode removes version metadata, validates, prints a diff, asks for confirmation, and publishes. Draft mode retains the working version identity, prints the same diff and confirmation, then saves locally without Firebase validation or publication.

The JSON envelope's `data` is one object containing `project_id`, `status`, `changed`, `draft`, `dry_run`, `validated`, `validation_source`, and `change_note`. Status is `imported`, `would-import`, `drafted`, `would-draft`, `unchanged`, or `validation-failed`. JSON mode suppresses human condition summaries and diffs, requires `--yes` when confirmation is needed, and requires an explicit import and merge-conflict strategy.

### Remote Config managed features

Experiments and rollouts provide read-only `list` and `show` commands plus an explicit destructive `delete` command. Personalizations remain read-only. The CLI cannot create, start, stop, or edit managed features, and none of these commands publish Remote Config. All three command groups use ordinary positional project resolution and the client Remote Config namespace.

Experiment and rollout IDs are not trimmed and accept either a slash-free ID
exactly as supplied or the exact case-sensitive
`projects/<resolved-project>/namespaces/firebase/<collection>/<id>` resource
name for that project and collection. Personalization IDs are also untrimmed
and compared exactly and case-sensitively. Every ID must
contain a non-whitespace value, so an empty or whitespace-only argument is an
argument error and never resolves to a collection resource.
An experiment or rollout value containing a slash but not matching that exact
resource-name form is `argument.invalid`. A well-formed personalization ID
which has no published-template binding is `personalization.not_found`.

Experiment and rollout metadata comes from Firebase's public Remote Config v1 managed-feature endpoints. fbrcm prefers the numeric `project_number` saved by project discovery and falls back to the Firebase project ID when that number is absent; both forms are accepted by the managed-feature resource paths.

Experiment, rollout, and personalization bindings come from `experimentValue`, `rolloutValue`, and `personalizationValue` objects in the published Remote Config template. Normal reads use the standard cache/revalidation policy. `--update` explicitly revalidates that cached template before scanning it. Drafts are intentionally excluded because Firebase managed-feature state refers to the published template.

These commands project known binding fields for display without rewriting managed values. Validate and publish preparation preserves complete opaque `experimentValue`, `rolloutValue`, and `personalizationValue` objects. Value editors treat all three as read-only rather than plain values. Template mutations also reject adding, replacing, removing, duplicating, renaming, or relocating managed values; imports, promotions, draft publication, stdin mutations, and unknown future value options use the same guard. Machine-readable experiment and rollout output also preserves unrecognized fields from Firebase's beta managed-feature responses so schema additions are not silently discarded.

### `fbrcm experiments list <project>`

Lists every experiment returned by Firebase using only the paginated list endpoint and correlates it with published-template bindings. Human output includes experiment ID, display name, parameter, condition, exposure percentage, relative last-update time, and state, in that order. Parameter names use the same blue styling as `get`, and condition names use their configured Remote Config tag colors. Missing values are shown as empty-value dashes, while an explicitly configured zero exposure is shown as `0%`. Descriptions and detail-only metadata such as variants and objectives are omitted from the human list. An experiment with no binding in the current template remains visible with empty binding columns.

Experiment metadata is always read live from Firebase. `--update` controls only
whether the separately cached published Remote Config template used for binding
correlation is explicitly revalidated.

Flags:

```text
-f, --filter <query>   filter display names locally; may be repeated
--update                revalidate cached Remote Config before reading bindings
--json                  put filtered experiment objects and references in envelope data
```

Experiment filters use the shared mode prefixes described under Filter Queries. Repeated filters are ORed. Matching is case-insensitive and applies only to the experiment display name, not its description or resource ID. Filtering is local after all list pages have been loaded; fbrcm does not send the query to Firebase.

### `fbrcm experiments show <project> <experiment-id>`

Shows one experiment's display metadata, state, timestamps, activation event, variants and weights, primary and secondary objectives, and every published parameter binding. Binding details include the experiment exposure percentage and each template variant ID with its value or no-change marker. Empty-string variant values are displayed as `""`, while absent values remain an empty-value dash. `<experiment-id>` is the final component printed in the list table, such as `2`.

The experiment metadata lookup always contacts Firebase; `--update` applies to
the published-template binding cache only.

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

Without `--yes`, the destructive confirmation selects Yes by default. Selecting No cancels successfully without deleting anything. In JSON mode, omitting `--yes` returns the resolved experiment with status `would-delete` plus an `interaction.required` problem; no DELETE request is sent. Successful output names the deleted experiment and project.

### `fbrcm rollouts list <project>`

Lists Firebase rollouts and correlates each rollout ID with its published-template parameter bindings. Human output includes ID, display name, parameter, condition, percentage, relative last-update time, and state, in that order. Parameter names use the same blue styling as `get`, and condition names use their configured Remote Config tag colors. Descriptions and enabled values are omitted from the human list but remain available through JSON and `rollouts show`. A rollout with no binding in the current template remains visible with empty binding columns.

Rollout metadata is always read live from Firebase. `--update` controls only
explicit revalidation of the published-template binding cache.

Flags:

```text
--update   revalidate cached Remote Config before reading bindings
--json     put rollout objects and references in envelope data
```

### `fbrcm rollouts show <project> <rollout-id>`

Shows rollout metadata, create/start/end/update timestamps, control and enabled variant names, and every published parameter binding. Explicit `0%` traffic and empty-string rollout values remain distinguishable from absent fields. `<rollout-id>` is the final component printed in the list table, such as `rollout_1`.

The rollout metadata lookup always contacts Firebase; `--update` applies to the
published-template binding cache only.

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

Without `--yes`, the destructive confirmation selects Yes by default. Selecting No cancels successfully without deleting anything. In JSON mode, omitting `--yes` returns the resolved rollout with status `would-delete` plus an `interaction.required` problem; no DELETE request is sent. Successful output names the deleted rollout and project.

### `fbrcm personalizations list <project>`

Scans the published template and lists every personalization ID with its group, parameter, and condition. Parameter names use the same blue styling as `get`, and condition names use their configured Remote Config tag colors. Firebase does not provide a separate public personalization resource endpoint, so the template is the authoritative API-visible source.

Flags:

```text
--update   revalidate cached Remote Config before scanning it
--json     put personalization IDs and references in envelope data
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

Version commands are scoped to one template target and use the same positional target resolution as `project export`: exact case-sensitive project ID, repository alias, then display name. Client and server histories and local snapshots are independent.

Every version command with `--cached` resolves the target and repository aliases
from the local projects registry only. It neither synchronizes projects from
Firebase nor rewrites that registry.

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

Version selectors are matched exactly and case-sensitively and are not trimmed.
For example, `CURRENT` and ` current` are invalid selectors rather than aliases
for `current`.

The current numeric parser also accepts an optional leading `+` and leading
zeroes in absolute versions and relative distances. Absolute numbers must fit
a signed 64-bit integer. Relative distances must fit signed 32-bit parsing and
remain between 1 and 299.

In live mode, relative selectors walk authoritative Firebase history. With `--cached`, they walk locally cached version numbers below the cached current version; because local history may be incomplete, a cached relative selector is not guaranteed to identify the same publication as its live equivalent. Relative distance must be between 1 and 299. Commands fail clearly when the requested relative position is unavailable.

Commands always verify that an exact numeric version fetch returns the requested version; they never silently substitute another version.

For non-cached `show`, `diff`, and `export`, project resolution may synchronize
a missing or empty project registry before inspecting local version snapshots.
Consequently a locally available immutable version does not by itself guarantee
an offline invocation; `--cached` is the complete no-network policy.

### `fbrcm versions list <project>`

Lists published Remote Config versions newest first. Live mode reads authoritative metadata from Firebase and marks locally cached versions. A version is marked `current` only when it matches the known current cache pointer; filtering the current publication out does not relabel the first remaining version as current. Cached mode performs no Firebase request and lists only local immutable snapshots.

Flags:

```text
--limit <n>          maximum versions to print; default 20; must be greater than zero
--all                retrieve every available version; mutually exclusive with an explicit --limit
--before <version>   newest version number to include; canonical positive decimal only
--since <RFC3339>    omit versions published before this time
--until <RFC3339>    omit versions published at or after this time
--cached             list local snapshots without contacting Firebase
--json               print structured JSON
```

The portable machine input contract limits `--limit` to 2,147,483,647.

Human live output keeps the existing column order: version number, current marker, publication time, updating user, origin, update type, cached marker, and Change Note. Cached output keeps version, current marker, cache time, size, and Change Note.

In cached mode, `--since` and `--until` apply to the local cache time because authoritative publication metadata may be unavailable.
`--before` accepts only an unsigned positive decimal integer in canonical form,
such as `1` or `142`; an explicitly empty value, zero, signs, leading zeros,
surrounding whitespace, and non-decimal forms are argument errors. Omitting the
option applies no version-number bound.

In JSON mode, `data.items` contains versions without project or pagination metadata and `data.count` contains their count. Each element uses fbrcm's canonical `change_note` name together with the other Firebase metadata, `current`, `cached`, and available local cache fields. Raw Firebase templates still encode this value in the API-required `version.description` field.

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
```

`--parameters` and `--conditions` are mutually exclusive. `--json` and `--side-by-side` are also mutually exclusive.

Default output reuses the conditions, group descriptions, parameters, and summary diff format used by `projects diff`. `--side-by-side` prints every changed entity as a complete, non-interactive two-column view. The command header establishes the `<from> → <to>` direction; individual changes omit repeated column headers and outer borders. Text wraps within the detected terminal width, JSON values are formatted before comparison, and long values retain contextual chunks around each difference. In JSON mode, `data` contains `project`, `from_version`, `to_version`, `changed`, and `diff`.

Differences return status 1 and no differences return status 0. Operational
failures always use their global semantic failure status in both human and JSON
modes. The status and JSON `changed` value describe the filtered result.

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

Rolling back to the current version or to an equivalent template is a no-op.
Its JSON result reports `status: "unchanged"`, `changed: false`, and a null
`published_version`. A changed successful result reports the previous version,
rollback source, and newly published version. Native Firebase rollback is a
force update; the final recheck narrows but cannot eliminate the race window
after that check.

If Firebase no longer retains a locally cached source version, rollback reports the failure and suggests the corresponding `restore` command.

### `fbrcm versions restore <project> <version>`

Republishes an exact locally cached immutable snapshot. Restore exists for recovery when Firebase no longer retains the historical version.

Restore refuses to run while the project has an unpublished draft.

Unlike rollback, restore:

- Requires the source version to be present locally.
- Publishes through the normal validated, ETag-protected update flow.
- Creates a normal new Remote Config version rather than Firebase rollback metadata.

It otherwise uses the same complete diff preview, confirmation, dry-run, current-version recheck, JSON contract, and success fields as rollback.
An already-current or equivalent cached snapshot uses the same `unchanged`
no-op result and never claims that a new version was or would be published.

Flags:

```text
--dry-run   validate and preview the cached snapshot without publishing
--change-note <text>
            set the new version's change note
-y, --yes   skip final publish confirmation
--json      print a structured operation result
```

Restore JSON includes `change_note`. Native rollback does not accept a change note and leaves its Firebase-defined rollback semantics unchanged.

Rollback and restore JSON results include `project_id`, `operation`, `previous_version`, `source_version`, `published_version`, `dry_run`, `changed`, `validated`, and `validation_source`, including no-op results where `changed` is `false`. Their command schemas fix `operation` to `rollback` or `restore` respectively and correlate post-publication failure statuses with Firebase validation provenance and the corresponding error stage. Human previews are written separately from JSON data so stdout remains machine-readable.

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

Project synchronization retains projects that are no longer accessible instead of deleting them. A project with no accessible auth identity is marked disabled. If a later update discovers it through another configured identity, the project is automatically rebound to that identity and enabled. Project JSON includes `aliases`, `disabled`, `templates`, and `primary_template`; `aliases` is always a sorted array, templates are a nonempty unique `client`/`server` set, and `primary_template` belongs to that set. Persisted `updated_at` and `synced_at` values are emitted as stored and therefore remain plain strings in the response schema. Human project listings include an Aliases column and mark disabled identities in the Auth column.

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

By default, command fetches live Remote Config for both projects. Use `--cached` to require the local projects registry and compare local parameter cache entries without contacting Firebase. Stale cache entries are compared as stored; a missing registry is `file.io_failed` with file-operation details, while an absent Remote Config entry is `parameters_cache.not_found` with selection details for the resolved target.

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

Default output is a terminal diff grouped by conditions, group descriptions, and parameters. JSON output includes source project, target project, top-level `changed`, summary counts, and structured change records.

Differences return status 1 and no differences return status 0. Operational
failures always use their global semantic failure status in both human and JSON
modes. The status and JSON `changed` value describe the filtered result.

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
--dry-run              preview the requested mutation without applying it
--change-note <text>   set the change note
-y, --yes              skip final publish confirmation
--json                 print promotion result JSON
```

Promotion JSON includes `change_note`, `changed`, `validated`, and `validation_source`; `changed` reports whether the selected result contains changes independently of whether it was a dry run or was published.

Non-interactive promote requires explicit selection intent: `--all`, `--filter`, `--group`, `--expr`, or `--search`. Command reloads the target before publishing, validates with Firebase, publishes using the latest target ETag, and retries if the target changes during promotion.

### `fbrcm projects aliases list`

Lists effective aliases from `.fbrcm.toml` and `.firebaserc`. This local
operation requires neither a profile nor network connectivity. Human output is
a naturally sized Alias/Project ID/Source table; narrow terminals crop Alias
and then Project ID while retaining Source. JSON is a sorted array containing
`alias`, `project_id`, and `source`; source is `fbrcm`, `firebase`, or `both`,
and an empty mapping is `[]`.

Flags:

```text
--json   print project aliases as JSON
```

### `fbrcm projects aliases set <alias> <project-id>`

Creates or updates one native `.fbrcm.toml` alias. Alias names use 1-63 lowercase letters,
digits, hyphens, or underscores and must start with a letter. Targets are literal
physical project IDs; template prefixes, filter prefixes, whitespace, and alias
chaining are not supported. The target need not be accessible to the current
profile. Remapping an existing native alias asks for confirmation with Yes
selected by default; setting the same effective value is an unchanged success.
An alias owned by `.firebaserc` cannot be remapped by this command; change it
with Firebase CLI or edit that file instead. Import creates a native snapshot
but does not transfer ownership while the Firebase definition remains.

Flags:

```text
-y, --yes   replace an existing alias without confirmation
    --json  print alias, previous_project_id, project_id, changed, and source
```

### `fbrcm projects aliases remove <alias>`

Removes one native alias without changing any profile registry, cache, draft,
Firebase resource, or `.firebaserc`. Removing an absent alias is an idempotent
success. A Firebase-only alias must be removed with Firebase CLI. If an
identical definition exists in both files, only the native definition is
removed and the alias remains effective from `.firebaserc`; JSON reports
`removed_native` and `remaining_source`. An invalid alias name is an
`argument.invalid` failure.

Flags:

```text
-y, --yes   remove the alias without confirmation
    --json  print alias, previous_project_id, status, changed, and source metadata
```

### `fbrcm projects aliases import --from <path>`

Reads the top-level Firebase CLI `projects` object from the exact supplied path
and snapshots its aliases into the nearest `.fbrcm.toml`. The command preserves
unrelated native configuration and never changes the source file. It renders a
preview table before confirmation. Identical mappings are unchanged; new
mappings are added.

Conflicts fail without writing by default. `--conflict keep` retains the native
value, while `--conflict overwrite` replaces it with the imported value. When
the source is the automatically discovered `.firebaserc`, keeping a different
native value leaves a live cross-source conflict that effective alias loading
will continue to reject. `--dry-run` previews without confirmation or writes.

Flags:

```text
    --from <path>                       Firebase RC file; required
    --conflict error|keep|overwrite     conflict policy; default error
    --dry-run                           preview the requested mutation without applying it
-y, --yes                              import without confirmation
    --json                              print paths, policy, dry_run, changed, and item actions
```

### `fbrcm projects path`

Prints projects config file path.

Flags:

```text
--json   print {"path": "..."}
```

### `fbrcm projects reset`

Resets the locally cached projects registry by deleting its rebuildable config file. Project Remote Config snapshots, cached versions, and drafts are not deleted.
In JSON mode, `status` is `reset`; `changed` is `true` when the registry file
was removed and `false` when it was already absent.

Flags:

```text
-y, --yes   skip confirmation
```

### `fbrcm doctor`

Runs a complete, non-interactive application health check. It verifies the selected profile and profile directories, auth registry, credential files, OAuth token presence and expiry, network/offline state, Cloud Resource Manager API access, Remote Config API reads, required Firebase read/update IAM permissions for cached projects, and profile cache writability. The writability check creates and removes a temporary probe file; capability metadata publishes those transient local file effects separately.

Doctor never opens OAuth login and never persists a refreshed token. In offline mode it reports the state and skips live API and permission checks. Online mode accesses Firebase only when at least one locally usable authentication identity is available. It prints every check even when some fail, and exits with status 1 when any check has `fail` status; warnings alone do not fail the command. The diagnostic run has no overall time limit by default. A deadline or `Ctrl+C` still prints the partial table or JSON report, then exits with the semantic timeout status 9 or canceled status 130 respectively; a failed check does not mask that context error.

An expired cached OAuth access token is normal when its refresh token still works. Online diagnostics report that token as `pass` after a successful in-memory refresh, `fail` when refresh fails, and `warn` only when refresh cannot be tested in offline mode. Doctor does not persist the refreshed access token.

Human-readable output uses the narrowest table and column widths that fit all content. When the natural table exceeds the detected terminal width, only Detail shrinks; long paths, permission lists, and API errors wrap onto additional lines inside that cell. Status and Check remain single-line and content-width.

Flags:

```text
--json                 print diagnostic checks as JSON
--timeout <duration>   optional positive time limit for the complete diagnostic run
```

In JSON mode, `data.items` contains checks and `data.count` contains their count. Every check includes the report-level profile, config directory, cache directory, and offline state.

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

All `config` subcommands are local operations and perform no network access.
They do not initialize a profile before executing their own operation. In JSON
mode, final envelope construction can still perform the shared conditional
default-profile bootstrap described under Shared Behavior.

### `fbrcm config show [key]`

Shows the effective layered configuration after applying keybinding migration
and built-in defaults. With no key, human output is TOML and includes the
complete effective key map. `--scope global` or `--scope local` instead shows
only values physically stored in that layer. Supported keys are `profile`,
`powerline_glyphs`, `keys`, `keys.<block>`, `keys.<block>.<action>`, `hooks`,
`hooks.timeout`, `hooks.pre_publish`, `hooks.post_publish`, `projects`,
`projects.aliases`, and `projects.aliases.<alias>`. A selected scalar prints as plain text; a selected
keybinding list or map prints scoped TOML. JSON is emitted only with `--json`.
Outer Unicode whitespace is removed for nested `keys.*`, `hooks.*`, and
`projects.aliases.<alias>` lookups. Top-level keys are compared exactly and
should be emitted in canonical form without surrounding whitespace.

The `projects.aliases` config keys describe native `.fbrcm.toml` state only.
Use `fbrcm projects aliases list` for the effective union with `.firebaserc` and
per-alias source metadata.

Flags:

```text
--scope effective|global|local   select the view; default effective
--json                           print structured JSON
```

Full JSON output includes `scope`, `path`, `exists`, both stored source paths,
and `config`. Selected-key JSON has `key`, `value`, and `source`; effective
sources are `local`, `global`, `default`, `mixed`, or `migrated`. A missing
config file is not created. A selected lookup of a physically stored global or
local layer uses `source: "absent"` and `value: null` when that layer has no
override for the requested key. Use `fbrcm config show keys` as the authoritative
reference for every configurable keybinding block and action.

### `fbrcm config set <key> <value>...`

Sets a typed preference. It atomically replaces the global config file with
private permissions by default. `--scope local` explicitly targets the nearest
repository config, or creates `.fbrcm.toml` in the current directory when none
is found. Supported forms are:

```text
powerline_glyphs true|false
keys.<block>.<action> <key>...
projects.aliases.<alias> <project-id>   requires --scope local
```

The active `profile` is read-only here; use `fbrcm profile switch <name>` or edit
the local TOML. Only explicit overrides are stored: inherited values and
built-in defaults are never copied into the target layer. The complete effective
candidate configuration is validated before writing, including unknown
blocks/actions, empty or duplicate bindings, unsupported key names, and
conflicts with configured or default actions. Failed validation leaves the file
unchanged.

Leading and trailing Unicode whitespace around nested
`keys.<block>.<action>` and `projects.aliases.<alias>` keys is removed before
lookup. The top-level `powerline_glyphs` key is compared exactly. The normalized
machine invocation schema publishes this conditional trimming.
Keybinding values accept a printable single character, `f1` through `f63`, the
documented terminal key names, or a unique sequence of supported modifiers
followed by one of those keys. Function-key names are canonical decimal names:
zero-padded or signed spellings such as `f01`, `f001`, `f+1`, and `ctrl+f01`
are rejected. A dangling modifier such as `ctrl+` is also invalid.

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
preserving that layer's `profile`. The optional key also accepts
`projects.aliases` or `projects.aliases.<alias>` in local scope. Reset can repair an invalid key map by
discarding the requested obsolete subtree. A changed reset asks for
confirmation; Yes is selected by default. Writes are validated and atomic.
Outer Unicode whitespace is removed for nested `keys.*` and
`projects.aliases.<alias>` keys; top-level reset keys are compared exactly.

Flags:

```text
-y, --yes             reset without confirmation
    --scope global|local   select the stored layer; default global
    --json             print the scoped reset result
```

### `fbrcm config validate`

Strictly validates TOML structure, profile references, project aliases,
keybindings, hook commands, and hook timeout syntax. Project aliases are
rejected in global scope. Effective and all-scope validation also checks the
discovered `.firebaserc` and cross-source alias conflicts. By
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
`--full` instead stages a complete generated keybinding template and commented
hook examples. Saving that
template makes every retained entry an explicit override, so remove entries
that should continue following future built-in defaults. Normal startup never
materializes defaults into either config file.

Editor resolution order is `--editor`, `FBRCM_EDITOR`, `VISUAL`, `EDITOR`, then `vi` on Unix-like systems or `notepad.exe` on Windows. Commands may include arguments; GUI editors generally need their wait flag, for example `--editor "code --wait"`.

In JSON mode the command returns `interaction.required` before reading
`--scope`, `--full`, or `--editor`; detailed capabilities and the invocation
schema therefore mark those accepted flags ineffective. Their documented
values and behavior above apply to interactive human execution.

Flags:

```text
--scope global|local   select the stored layer; default global
--full                 stage a complete generated keybinding template
--editor <command>     override the editor command
```

Hook arrays are authored through `config edit`; `config set` intentionally does
not split command strings into array elements.

### `fbrcm hooks status`

Shows the effective pre- and post-publication commands, timeout, local config
path, fingerprint, and whether local hooks are currently trusted.

Flags:

```text
--json   print structured hook status
```

### `fbrcm hooks fingerprint`

Prints the SHA-256 fingerprint of the current effective local hook definition.
Returns nonzero when the repository config does not define a hook array. Pin
this value in `FBRCM_HOOK_TRUST` for noninteractive CI execution.

### `fbrcm hooks trust`

Displays the local config path, hook working directory, fingerprint, and every
effective command, then asks whether to trust it. Trust is stored in the global
private `hook-trust.json` file and becomes invalid when the hook definition or
canonical local config path changes. Yes is selected by default.

Flags:

```text
-y, --yes   trust without confirmation
    --json  print structured hook status
```

### `fbrcm hooks untrust`

Removes stored trust for the currently discovered local config. It does not
edit either config layer.

Flags:

```text
--json   print the local path and whether a record changed
```

### `fbrcm auth list`

Lists configured auth identities.

When a command needs Firebase access but the active profile has no configured auth identity, the error includes setup guidance. Run `fbrcm` for guided setup, or use `fbrcm auth add --help` to see the CLI authentication options.

Flags:

```text
--json   print auth identities as JSON
```

In JSON mode, `data.items` contains identities and `data.count` contains their count. Every identity includes a `default` boolean; exactly the configured default identity has `default: true`.

### `fbrcm auth add oauth <auth-id>`

Adds or replaces an OAuth identity and imports its desktop client secret JSON.
The input must contain `installed`, `web`, or both. With only one object, it
requires nonblank `client_id` and `client_secret` values, absolute parseable
`auth_uri` and `token_uri` endpoints, and a nonempty `redirect_uris` array.
When both are present, the existing runtime uses `web` for Google's redirect
selection and `installed` for fbrcm's field validation: `web.redirect_uris`
must be nonempty, while `installed` must supply the nonblank ID/secret and
absolute endpoint fields; `installed.redirect_uris` is optional. The schema
models that precedence exactly, although agents should prefer one complete
client object.

Input source order:

```text
--from <path>
stdin
interactive .json file picker
```

The machine input schema publishes this as a typed first-available rule. When
both `--from` and redirected stdin are present, the file wins and stdin is not
consumed.

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

The machine input schema uses the same file-before-stdin selection rule.

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

Authenticates or validates an auth identity. OAuth uses a valid cached token,
refreshes it when possible, and starts browser login only when needed;
service-account validates the key; gcloud validates ADC discovery. In JSON
mode, OAuth returns `interaction.required` only when human authorization is
needed. OAuth refresh may contact Google's token endpoint and persist a token;
gcloud ADC discovery may contact the metadata server when no local ADC source
is available. Service-account and gcloud validation remain non-interactive.
Malformed stored credentials return typed auth failures. A missing or revoked
OAuth grant can require human authorization, while transient token-endpoint
network, timeout, rate-limit, and service errors retain typed retryable failures
instead of being reported as authorization interaction.

Because JSON mode blocks human OAuth authorization before any browser-opening
step, `--noopen` is accepted but marked ineffective in the machine contract.

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
--auth <auth-id>          required auth identity to bind
-p, --project <query>     filter projects; may be repeated
```

A project is rebound only when the target identity discovered it during project synchronization. Inaccessible projects are skipped, logged individually as errors, and counted in the final bound/skipped summary; they do not fail the rest of the batch. When no cached project matches the supplied filters, JSON mode returns `project.not_found` with typed selection details.

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
session; repository selection applies again on the next launch. The global
profile selection is persisted on every successful invocation, including when
the requested profile was already selected; JSON `changed` describes the
selected name, not whether the configuration file was rewritten.

### `fbrcm profile rename <old-name> <new-name>`

Renames an existing profile. fbrcm refuses to rename a profile selected by the
nearest `.fbrcm.toml`, because it never rewrites repository configuration as a
side effect of profile management. Supplying the same valid profile name twice
is an unchanged success and reports `changed: false` in JSON. A changed rename
moves the profile configuration directory and, when it exists and the
destination does not, its complete cache directory. Capability metadata exposes
the latter as conditional `local_cache_move`.

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

Shows help for the longest exact existing command-path prefix. Navigational
command groups are valid. Any unmatched suffix components are ignored; an
unknown first component, or no components, shows root help rather than
returning a not-found failure.

Examples:

```sh
fbrcm help project import
fbrcm get --help
```
