# fbrcm

`fbrcm` is a terminal manager for Firebase Remote Config. Use its interactive
TUI to explore and edit projects, or use the CLI for scripts, repeatable
operations, and machine-readable output.

> [!TIP]
> Using `fbrcm` from an LLM agent, script, or CI runner? Start with the
> [agent quickstart](https://github.com/yumauri/fbrcm/blob/main/docs/agent-quickstart.md).
> Machine-readable command discovery is also available through
> `fbrcm capabilities --json`.

![fbrcm TUI demo](https://raw.githubusercontent.com/yumauri/fbrcm/main/vhs/demo.gif)

> [!CAUTION]
> This project is almost completely vibe-coded.

> [!TIP]
> Already using the official Firebase CLI? See
> [fbrcm vs. Firebase CLI](https://github.com/yumauri/fbrcm/blob/main/docs/firebase-cli-comparison.md) for feature parity,
> command equivalents, and the Remote Config workflows fbrcm adds.

fbrcm is most useful when work spans more than one Firebase project:

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

Install the latest release from PowerShell:

```powershell
irm https://raw.githubusercontent.com/yumauri/fbrcm/main/install.ps1 | iex
```

The installer places `fbrcm.exe` in
`%LOCALAPPDATA%\Programs\fbrcm\bin` by default and adds that directory to your
user `PATH`. To use another directory:

```powershell
$env:INSTALL_DIR = 'C:\Tools\fbrcm'
irm https://raw.githubusercontent.com/yumauri/fbrcm/main/install.ps1 | iex
```

With [Scoop](https://scoop.sh):

```powershell
scoop bucket add yumauri https://github.com/yumauri/scoop-bucket
scoop install fbrcm
```

### Other options

Download an archive from
[GitHub Releases](https://github.com/yumauri/fbrcm/releases), or install from
source with Go 1.27.0 or newer:

```sh
go install github.com/yumauri/fbrcm@latest
```

Official release binaries include fbrcm's Google OAuth Desktop client. A plain
source build leaves it out. In that build, the `google` authentication method
reports that the built-in client is unavailable. The `oauth`,
`service-account`, and `gcloud` methods continue to work.

For local development, build with a Desktop client from values in the current
shell. This does not add the client JSON to the repository or runtime
configuration:

```sh
export FBRCM_GOOGLE_OAUTH_CLIENT_ID="$(jq -r '.installed.client_id' /path/to/client-secret.json)"
export FBRCM_GOOGLE_OAUTH_CLIENT_SECRET="$(jq -r '.installed.client_secret' /path/to/client-secret.json)"
go run ./cmd/genoauthclient
go build -tags=fbrcm_google_auth .
go run ./cmd/genoauthclient -clean
```

Use a local development client unless you have access to the official release
credentials.

## First run

Start the TUI:

```sh
fbrcm
```

On a new profile, fbrcm opens guided setup. It supports:

- Google sign-in using fbrcm's built-in shared OAuth client;
- an OAuth Desktop app client;
- a service-account JSON key;
- existing [Google Cloud CLI](https://cloud.google.com/cli) Application Default
  Credentials.

The identity needs access to the projects you want to manage. Project discovery
uses the Cloud Resource Manager API; template reads, validation, publication,
defaults, and history use the Firebase Remote Config API.

fbrcm requests the broad `cloud-platform` OAuth scope because Google's
interactive consent flow rejects the narrower `firebase.remoteconfig` scope,
while the grantable `firebase` scope is insufficient for Remote Config. This
[Firebase discussion](https://groups.google.com/g/firebase-talk/c/a8H9GcGiYuA)
reports the same limitation. fbrcm limits its Google API calls to Cloud Resource
Manager and Firebase Remote Config; see the [privacy policy](PRIVACY.md) for the
data-access, local-storage, and deletion details.

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

Every CLI command supports a shared versioned automation contract. Add
`--json` to receive one envelope on success or failure, with typed data,
structured errors, semantic exit status, and execution context:

```sh
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --yes --json
```

Use `--type json --value '<json>'` when the parameter value is JSON. Agents can
discover commands with `fbrcm capabilities --json` and embedded schemas with
`fbrcm schema list --json`. See the [CLI machine contract](https://github.com/yumauri/fbrcm/blob/main/docs/cli-contract.md)
for the envelope, errors, exit statuses, non-interactive rules, and schemas.

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
- publication `--dry-run` uses Firebase's validation endpoint, then suppresses
  publication and local writes;
- `--draft` stages supported mutations in the active profile;
- draft publication rebases local intent onto current Firebase state and stops
  on conflicts;
- multi-project publication is non-atomic: each project succeeds or fails
  independently, and fbrcm reports every result.

Client and server templates have independent drafts, caches, and histories.
Use an explicit target such as `client@my-project` or
`server@my-project` when the configured default is not the one you want.

## Configuration

The TUI key map, theme, and active-profile preference can be stored in the global
`config.toml` or a repository `.fbrcm.toml`. fbrcm searches from the current
directory to the filesystem root, then deeply overlays the nearest local file
on the global configuration. Inspect the effective configuration and its
sources with:

Profile-specific configuration and cache state is stored under
`profiles/<name>` within the corresponding application root.

```sh
fbrcm config show
fbrcm config path
fbrcm config path --scope local
```

For example:

```sh
fbrcm config set powerline_glyphs false
fbrcm config set theme nord
fbrcm config set keys.projects.refresh u ctrl+r
fbrcm config set network.max_concurrent_requests 3
fbrcm config set network.requests_per_minute 30
fbrcm config set network.rate_limit_cooldown 30s
fbrcm config set network.retry.max_attempts 5
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

Network requests share a configurable concurrency limit and Firebase 429
cooldowns across workers. A positive
`network.requests_per_minute` paces requests by API and quota consumer when set
to a positive value. Its default of zero disables proactive pacing. Retry
attempts, exponential delays, and jitter are configurable under
`network.retry`. Configuration files stay sparse. fbrcm applies built-in values
in memory and does not write them during startup. Use `fbrcm config show keys`
as the authoritative key-name reference, or `fbrcm config edit --full` to stage
a complete generated template.

Themes are shareable TOML files stored under the user-wide `themes` directory.
See [Theming](docs/theming.md) for installation, inheritance, fallback rules,
and the complete color-token reference.

Profiles keep authentication, project selection, drafts, and caches separate.
Use `Ctrl+P` in the TUI or `fbrcm profile --help` in the CLI.

## Documentation

| Guide | Use it for |
| --- | --- |
| [Agent quickstart](https://github.com/yumauri/fbrcm/blob/main/docs/agent-quickstart.md) | Noninteractive discovery, safe mutation, structured recovery, and automation pitfalls |
| [TUI guide](https://github.com/yumauri/fbrcm/blob/main/docs/TUI.md) | Setup, panels, shortcuts, editing, drafts, history, promotion, and key configuration |
| [Theming](https://github.com/yumauri/fbrcm/blob/main/docs/theming.md) | Shared CLI/TUI themes, inheritance, color tokens, and fallback behavior |
| [CLI reference](https://github.com/yumauri/fbrcm/blob/main/docs/CLI.md) | Complete command tree, flags, output contracts, template targets, and write behavior |
| [Expression filters](https://github.com/yumauri/fbrcm/blob/main/docs/EXPR.md) | Expression contexts, typed values, helper functions, and `jq` queries |
| [Architecture](https://github.com/yumauri/fbrcm/blob/main/docs/architecture.md) | Package boundaries and maintainer invariants |
| [Root group keys](https://github.com/yumauri/fbrcm/blob/main/docs/root-group-key.md) | Internal root-parameter representations |
| [Privacy policy](PRIVACY.md) | Google API data access, OAuth scope rationale, local storage, sharing, and deletion |

Every CLI command also has focused help:

```sh
fbrcm --help
fbrcm projects promote --help
```

## Build from source

The module currently requires Go 1.27.0 or newer:

```sh
git clone https://github.com/yumauri/fbrcm.git
cd fbrcm
go build -o fbrcm .
go test ./...
```

### Documentation site

The VitePress site uses the curated Markdown files under `docs/site/`. The
long-form project documents in `docs/` remain separate from the website:

```sh
cd docs
npm ci
npm run dev
```

Use `npm run build` to create a production build locally.

The development and production commands generate the human-readable
`/privacy-policy` page from the repository-root `PRIVACY.md` and copy the
repository-root `llms.txt` and `LICENSE` byte-for-byte to `/llms.txt` and
`/LICENSE.txt`. Edit only the root source files; their generated website copies
are ignored by Git.

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
