# fbrcm

`fbrcm` is a terminal Firebase Remote Config manager. It helps you manage Remote Config across Firebase projects, inspect parameters and condition priority/usage, export and import Remote Config JSON, and safely add, update, or delete Remote Config parameters.

Run `fbrcm` without arguments to open the interactive TUI. Run `fbrcm <command>` to use the CLI.

> [!CAUTION]
> This project is almost completely vibe-coded

## Requirements

- Go 1.26 or newer
- Google account with access to the Firebase or Google Cloud projects you want to manage
- OAuth Desktop Client JSON from Google Cloud Console
- Access to these Google APIs for the target projects:
  - Cloud Resource Manager API, used to list projects
  - Firebase Remote Config API, used to read, validate, publish, and list Remote Config versions

## Install or Run

Install latest release with the shell installer on macOS or Linux:

```sh
curl -sSfL https://raw.githubusercontent.com/yumauri/fbrcm/main/install.sh | sh
```

Install to a custom directory:

```sh
curl -sSfL https://raw.githubusercontent.com/yumauri/fbrcm/main/install.sh | INSTALL_DIR="$HOME/.local/bin" sh
```

Install with [Homebrew](https://brew.sh):

```sh
brew tap yumauri/tap
brew install --cask fbrcm
```

Install with [Scoop](https://scoop.sh) on Windows:

```powershell
scoop bucket add yumauri https://github.com/yumauri/scoop-bucket
scoop install fbrcm
```

Install with Go:

```sh
go install github.com/yumauri/fbrcm@latest
```

Download a release archive manually from:

```text
https://github.com/yumauri/fbrcm/releases
```

From the repository root:

```sh
go run .
```

Build a local binary:

```sh
go build -o fbrcm .
./fbrcm --help
```

## TUI Configuration

The TUI stores its global settings in `config.toml` under the fbrcm config directory. Powerline separators are enabled by default; disable them to use standard Unicode arrows when the terminal font does not include Powerline glyphs:

```toml
powerline_glyphs = false
```

History and version-chooser keys are configurable like all other TUI bindings:

```toml
[keys.global]
focus_conditions = ["3"]
focus_history = ["4"]
focus_details = ["5"]

[keys.conditions]
rename = ["r"]
edit = ["e"]
color = ["c"]
new = ["a"]
move = ["m"]
delete = ["x"]
publish = ["p"]
publish_all = ["P"]
discard = ["d"]
discard_all = ["D"]

[keys.history]
pair_older = [","]
pair_newer = ["."]
choose_versions = ["v"]
toggle_changes = ["c"]

[keys.history_picker]
cancel = ["esc"]
toggle = ["tab", "shift+tab"]
left = ["left"]
right = ["right"]
pair_older = [","]
pair_newer = ["."]
rollback = ["R"]
reset = ["r"]
up = ["up", "k"]
down = ["down", "j"]
page_up = ["pgup"]
page_down = ["pgdown"]
home = ["home"]
end = ["end"]
submit = ["enter"]
```

## First Setup

`fbrcm` supports three Google auth methods: OAuth desktop login, service account keys, and gcloud Application Default Credentials (ADC).

### OAuth Desktop Login

You need a Desktop app OAuth client secret JSON:

> [!NOTE]
> APIs & Services -> Credentials -> Create Credentials -> OAuth client ID -> Desktop app

1. Open Google Cloud Console OAuth clients:
   `https://console.cloud.google.com/auth/clients`
2. Select or create a Google Cloud project.
3. Create an OAuth client.
4. Choose application type `Desktop app`.
5. Create it and download the JSON file.
6. Import that JSON into `fbrcm`.

Import downloaded client secret file:

```sh
fbrcm auth add oauth default --from /path/to/client-secret.json
```

If `--from` is omitted, the command reads piped stdin; without stdin it opens an interactive `.json` file picker.

After the client secret is imported, authenticate:

```sh
fbrcm auth login default
```

The app opens a browser authorization page and waits for the local OAuth callback. If the browser does not open, copy the printed URL into a browser.

Check current auth files:

```sh
fbrcm auth path default
```

### Service Account

Import a service account JSON key:

```sh
fbrcm auth add service-account prod --from /path/to/service-account.json
```

If `--from` is omitted, the command reads piped stdin; without stdin it opens an interactive `.json` file picker.

### gcloud ADC

Create Application Default Credentials with gcloud, then add an auth identity that uses ADC discovery:

```sh
gcloud auth application-default login
fbrcm auth add gcloud default
```

## Where Auth Is Stored

By default, `fbrcm` stores per-profile files under your user config and cache directories.

- Auth config: `~/.config/fbrcm/<profile>/auth-config.json`
- OAuth client secrets: `~/.config/fbrcm/<profile>/auth/<auth-id>/client-secret.json`
- Service account keys: `~/.config/fbrcm/<profile>/auth/<auth-id>/service-account.json`
- Projects cache: `~/.config/fbrcm/<profile>/projects-config.json`
- OAuth token cache: user cache directory, under `fbrcm/<profile>/auth/<auth-id>/token.json`

Project cache is a known-project registry. Each project stores its selected `auth_id`, so different projects can use different auth identities.

Exact paths:

```sh
fbrcm auth path default
fbrcm projects path
fbrcm cache path
fbrcm draft path
```

You can override root directories with environment variables:

- `FBRCM_CONFIG_DIR`
- `FBRCM_CACHE_DIR`

Delete auth files:

```sh
fbrcm auth purge default
```

## Basic Usage

Open interactive UI:

```sh
fbrcm
```

Show CLI help:

```sh
fbrcm --help
fbrcm <command> --help
```

List projects:

```sh
fbrcm projects list
fbrcm projects list --update
fbrcm projects list --json
```

Get Remote Config parameters across projects:

```sh
fbrcm get
fbrcm get some_parameter
fbrcm get --project my-project
fbrcm get --project proj1 --project proj2
fbrcm get --filter login
fbrcm get --filter login --filter checkout
fbrcm get --search rollout
fbrcm get --json
```

Inspect conditions in their Firebase evaluation order and see which parameters use them:

```sh
fbrcm conditions list <project-id>
fbrcm conditions list <project-id> --filter beta
fbrcm conditions list <project-id> --search platform
fbrcm conditions show <project-id> <condition-name>
fbrcm conditions show <project-id> <condition-name> --json
```

Manage condition definitions from the CLI:

```sh
fbrcm conditions add <project-id> beta_users --expression "percent <= 10" --color BLUE
fbrcm conditions edit <project-id> beta_users --expression "percent <= 20"
fbrcm conditions edit <project-id> beta_users --color GREEN
fbrcm conditions rename <project-id> beta_users expanded_beta
fbrcm conditions move <project-id> expanded_beta 1
fbrcm conditions delete <project-id> expanded_beta
fbrcm conditions validate <project-id>
```

Definition mutations print a Remote Config diff and offer publication or can be staged with `--draft`. Use `--dry-run` to preview without persisting state and `--yes` to skip confirmation. `conditions validate` validates the current draft, if present, or the published template with Firebase's validate-only API.

In the TUI, press `3` by default to open the Conditions tab. The default actions are `a` add, `r` rename, `e` edit the raw expression, `c` change color, `m` move priority, and `x` delete. Mutations show a diff with Publish, Draft, and Cancel choices; once a project has a draft, subsequent edits stage into it immediately. Use `p`/`P` to publish and `d`/`D` to discard project/all drafts. Press Enter on a condition to see its expression, priority, color, and parameter usages; the same edit actions work from Details.

Export one project Remote Config:

```sh
fbrcm project export <project-id> --to remote-config.json
```

Inspect and recover Remote Config version history:

```sh
fbrcm versions list <project-id>
fbrcm versions show <project-id> 142
fbrcm versions diff <project-id> 138 current
fbrcm versions rollback <project-id> 138 --dry-run
fbrcm versions rollback <project-id> 138
```

Firebase version history is authoritative, but Firebase retains at most 300 versions and may remove inactive versions older than 90 days. `fbrcm` keeps immutable templates it has encountered until the cache is purged. A cached version that Firebase no longer retains can be republished with:

```sh
fbrcm versions restore <project-id> 37 --dry-run
fbrcm versions restore <project-id> 37
```

`rollback` uses Firebase's native rollback operation and creates a new version with rollback metadata. `restore` republishes a local snapshot as a normal new version. Both commands print a full diff and ask for confirmation unless `--yes` is used.

Import Remote Config:

```sh
fbrcm project import <project-id> --from remote-config.json --dry-run
fbrcm project import <project-id> --from remote-config.json --merge
fbrcm project import <project-id> --from remote-config.json --override
fbrcm project import <project-id> --from remote-config.json --search rollout --dry-run
```

Add parameter:

```sh
fbrcm add new_parameter --project my-project --string "value" --description "Used by app startup"
fbrcm add feature_enabled --project my-project --boolean true --dry-run
```

Update parameter:

```sh
fbrcm update existing_parameter --project my-project --string "new value" --dry-run
fbrcm update existing_parameter --project my-project --name renamed_parameter --yes
fbrcm update --filter feature --search rollout --boolean true --dry-run
```

Delete parameter:

```sh
fbrcm delete old_parameter --project my-project --dry-run
fbrcm delete old_parameter --project my-project --yes
fbrcm delete --filter old --search rollout --dry-run
```

Stage changes in profile-scoped local drafts instead of publishing immediately:

```sh
fbrcm add new_flag --project my-project --boolean true --draft
fbrcm update existing_parameter --project my-project --string "new value" --draft --yes
fbrcm delete old_parameter --project my-project --draft --yes
fbrcm project import <project-id> --from remote-config.json --merge --draft
```

Inspect and resolve drafts:

```sh
fbrcm draft list
fbrcm draft show <project-id> --to recovered-draft.json
fbrcm draft diff <project-id>
fbrcm draft diff <project-id> --against current
fbrcm draft publish <project-id> --dry-run
fbrcm draft publish <project-id>
fbrcm draft discard <project-id>
```

Publishing rebases local draft changes onto the latest Firebase Remote Config and refuses conflicting changes. `draft publish --all` and `draft discard --all` process every draft in the active profile. Other CLI write commands refuse to publish over a project with an unresolved draft.

Manage caches:

```sh
fbrcm cache list
fbrcm cache purge
fbrcm projects purge
```

## Profiles

Profiles let you keep separate OAuth clients, project caches, and token caches.

```sh
fbrcm profile
fbrcm profile list
fbrcm profile switch work
fbrcm profile switch personal
fbrcm profile rename old-name new-name
```

First run creates and uses the `default` profile.

## Filtering

Project, parameter, and condition filters support mode-prefixed queries:

- No prefix or `~`: fuzzy match.
- `^`: starts with.
- `/`: includes.
- `=`: exact match.

Examples:

```sh
fbrcm projects list --filter '^prod'
fbrcm get --filter '/checkout'
fbrcm get --project '=my-project-id'
fbrcm conditions list my-project --filter '~bt'
```

Several commands also support `--expr` with [expr-lang](https://expr-lang.org/docs/language-definition) expressions for advanced filtering.

Parameter commands support `--search` for matching names, descriptions, values, condition names, and condition expressions. `--filter`, parameter-context `--expr`, and `--search` are ANDed.

## What It Can Do

- Open a TUI for managing Firebase projects, Remote Config parameters, and conditions
- List Firebase projects available to the authenticated Google account
- Cache project metadata and Remote Config snapshots locally
- List, inspect, compare, export, roll back, and restore Remote Config versions
- Fetch Remote Config from Firebase
- Show parameters across many projects
- Add, edit, rename, reorder, delete, validate, list, and inspect conditions and their parameter/value usage
- Filter projects and parameters
- Export Remote Config JSON
- Import Remote Config JSON
- Merge imported config into current project config
- Override current config with imported config
- Remove project-specific conditions during import
- Add, update, rename, move, duplicate, and delete parameters
- Display empty parameter groups and remove groups explicitly with the TUI delete action
- Stage, inspect, diff, safely publish, recover, and discard local drafts
- Edit parameter values as boolean, number, string, or JSON
- Validate and publish Remote Config through Firebase APIs
- Use `--dry-run` on write commands to preview Firebase writes without sending them

## Safety Notes

Use `--dry-run` before imports, updates, adds, deletes, draft publishes, rollbacks, and restores when you are unsure. Write commands print diffs and usually ask for confirmation unless `--yes` is used.

Purging the Remote Config cache deletes every locally retained immutable version. Versions no longer retained by Firebase may then be permanently unavailable.

Drafts are managed separately from cached snapshots. `fbrcm cache purge` does not delete drafts; use `fbrcm draft discard` explicitly.

Keep `client-secret.json`, `token.json`, and service-account key files private. They grant access through Google account or service account permissions.
