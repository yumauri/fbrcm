# fbrcm

`fbrcm` is a terminal manager for Firebase Remote Config. Use its interactive
TUI to explore and edit projects, or use the CLI for scripts, repeatable
operations, and machine-readable output.

![fbrcm TUI demo](https://raw.githubusercontent.com/yumauri/fbrcm/main/vhs/demo.gif)

> [!CAUTION]
> This project is almost completely vibe-coded.

> [!TIP]
> Already using the official Firebase CLI? See
> [fbrcm vs. Firebase CLI](https://github.com/yumauri/fbrcm/blob/main/docs/firebase-cli-comparison.md) for feature parity,
> command equivalents, and the Remote Config workflows fbrcm adds.

It is designed for work that spans more than one Firebase project:

- inspect parameters, groups, conditions, and version history;
- compare and promote configuration between projects or template types;
- stage related changes as local drafts before publishing;
- import, export, validate, roll back, and restore Remote Config templates;
- manage client and server templates from the same workspace.

## Installation

### macOS and Linux

Install the latest release:

```sh
curl -sSfL https://raw.githubusercontent.com/yumauri/fbrcm/main/install.sh | sh
```

The installer places `fbrcm` in `/usr/local/bin` by default. To use another
directory:

```sh
curl -sSfL https://raw.githubusercontent.com/yumauri/fbrcm/main/install.sh |
  INSTALL_DIR="$HOME/.local/bin" sh
```

With [Homebrew](https://brew.sh):

```sh
brew tap yumauri/tap
brew install --cask fbrcm
```

### Windows

With [Scoop](https://scoop.sh):

```powershell
scoop bucket add yumauri https://github.com/yumauri/scoop-bucket
scoop install fbrcm
```

### Other options

Download an archive from
[GitHub Releases](https://github.com/yumauri/fbrcm/releases), or install from
source with Go 1.26.5 or newer:

```sh
go install github.com/yumauri/fbrcm@latest
```

## First run

Start the TUI:

```sh
fbrcm
```

On a new profile, fbrcm opens guided setup. It supports:

- an OAuth Desktop app client;
- a service-account JSON key;
- existing [Google Cloud CLI](https://cloud.google.com/cli) Application Default
  Credentials.

The identity needs access to the projects you want to manage. Project discovery
uses the Cloud Resource Manager API; template reads, validation, publication,
defaults, and history use the Firebase Remote Config API.

After setup, run the built-in diagnostic whenever you need to check credentials,
connectivity, API access, permissions, or local storage:

```sh
fbrcm doctor
```

See [TUI setup and workflows](https://github.com/yumauri/fbrcm/blob/main/docs/TUI.md#setup-and-authentication) for the
guided path, or [CLI authentication](https://github.com/yumauri/fbrcm/blob/main/docs/CLI.md#fbrcm-auth-list) for
non-interactive setup.

## A quick tour

Run `fbrcm` with no arguments for the TUI. Press `?` anywhere to open the
searchable action palette; it shows the shortcuts that are relevant to the
current panel.

Any argument selects CLI mode:

```sh
# Refresh and list accessible projects.
fbrcm projects list --update

# Inspect one parameter across matching projects.
fbrcm get feature_enabled --project '^prod'

# Preview a change without writing it.
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --dry-run

# Stage the same change in a local draft.
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --draft

# Review and publish the draft.
fbrcm draft diff my-app --against current
fbrcm draft publish my-app
```

Repositories can give stable environment names to physical Firebase projects:

```sh
fbrcm projects aliases import --from .firebaserc --dry-run
fbrcm projects aliases set staging acme-staging-42
fbrcm projects aliases set prod acme-production-42
fbrcm projects diff staging prod
fbrcm projects promote client@staging server@prod --dry-run
```

fbrcm reads Firebase CLI aliases from `.firebaserc` and native aliases from the
nearest `.fbrcm.toml`. Both are repository-scoped, profile-independent, and
normally committed for the team and CI. Conflicting definitions are rejected;
identical definitions are shared. Canonical project IDs remain in Firebase
requests, caches, drafts, and automation output.

Direct Remote Config mutations support a shared automation contract. Add
`--json` to `add`, `update`, `delete`, `duplicate`, condition mutations, or
group mutations to receive one structured result per target, including changed
item count, version transition, draft and dry-run state, structured errors, and
an exact retry selector when retrying is safe:

```sh
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --yes --json
```

Use `--type json --value '<json>'` when the parameter value is JSON. See the
[CLI mutation JSON contract](https://github.com/yumauri/fbrcm/blob/main/docs/CLI.md#mutation-json-automation-contract) for
the complete schema.

Project and parameter filters support fuzzy, prefix, contains, and exact modes.
Expression filters can also inspect typed values and complete Remote Config
context:

```sh
fbrcm get --expr 'value == true'
fbrcm projects list --expr '"feature_enabled" in keys(parameters)'
```

## Drafts and safe writes

Remote Config publication replaces a complete template, so fbrcm treats review
as part of the write workflow:

- write commands show a diff and normally ask for confirmation;
- `--dry-run` previews without changing Firebase or local drafts;
- `--draft` stages supported mutations in the active profile;
- draft publication rebases local intent onto current Firebase state and stops
  on conflicts;
- multi-project publication is non-atomic: each project succeeds or fails
  independently, and fbrcm reports every result.

Client and server templates have independent drafts, caches, and histories.
Use an explicit target such as `client@my-project` or
`server@my-project` when the configured default is not the one you want.

## Configuration

The TUI key map and active-profile preference can be stored in the global
`config.toml` or a repository `.fbrcm.toml`. fbrcm searches from the current
directory to the filesystem root, then deeply overlays the nearest local file
on the global configuration. Inspect the effective configuration and its
sources with:

```sh
fbrcm config show
fbrcm config path
fbrcm config path --scope local
```

For example:

```sh
fbrcm config set powerline_glyphs false
fbrcm config set keys.projects.refresh u ctrl+r
fbrcm config validate
fbrcm config edit --scope local
```

Native repository project aliases use a local-only table and can also be edited
directly:

```toml
[projects.aliases]
staging = "acme-staging-42"
prod = "acme-production-42"
```

Firebase CLI aliases under the top-level `.firebaserc` `projects` object are
also resolved automatically. Use `fbrcm projects aliases import --from
.firebaserc` to copy them into the native table.

Configuration files stay sparse: built-in keybindings are applied in memory
and are not written during startup. Use `fbrcm config show keys` as the
authoritative key-name reference, or `fbrcm config edit --full` to stage a
complete generated template.

Profiles keep authentication, project selection, drafts, and caches separate.
Use `Ctrl+P` in the TUI or `fbrcm profile --help` in the CLI.

## Documentation

| Guide | Use it for |
| --- | --- |
| [TUI guide](https://github.com/yumauri/fbrcm/blob/main/docs/TUI.md) | Setup, panels, shortcuts, editing, drafts, history, promotion, and key configuration |
| [CLI reference](https://github.com/yumauri/fbrcm/blob/main/docs/CLI.md) | Complete command tree, flags, output contracts, template targets, and write behavior |
| [Expression filters](https://github.com/yumauri/fbrcm/blob/main/docs/EXPR.md) | Expression contexts, typed values, helper functions, and `jq` queries |
| [Architecture](https://github.com/yumauri/fbrcm/blob/main/docs/architecture.md) | Package boundaries and maintainer invariants |
| [Root group keys](https://github.com/yumauri/fbrcm/blob/main/docs/root-group-key.md) | Internal root-parameter representations |

Every CLI command also has focused help:

```sh
fbrcm --help
fbrcm projects promote --help
```

## Build from source

The module currently requires Go 1.26.5 or newer:

```sh
git clone https://github.com/yumauri/fbrcm.git
cd fbrcm
go build -o fbrcm .
go test ./...
```

## Security notes

Treat OAuth client files, OAuth tokens, and service-account keys as secrets.
Use `fbrcm auth path <auth-id>` to locate identity files.

Do not store secrets in Remote Config values. Applications can receive and
inspect the parameters available to them.

Cached historical templates may outlive Firebase's retained history. Clearing
the fbrcm cache can therefore remove the only remaining local copy of an old
template. Drafts are separate and are deleted only through explicit draft or
project cleanup operations.

## Acknowledgments

fbrcm was inspired by Dmitrii Andriianov's
[RemoteConfigModifier](https://github.com/andriyanovDS/RemoteConfigModifier).
I discovered it while looking for a tool that could batch-update Firebase
Remote Config across several projects, and it gave me the motivation to build
fbrcm.

RemoteConfigModifier is a smaller Rust project with a more focused feature set.
fbrcm is an independent implementation and does not reuse any of its code.

## License

[MIT](https://github.com/yumauri/fbrcm/blob/main/LICENSE)
