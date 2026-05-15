# fbrcm

`fbrcm` is a terminal tool for Firebase Remote Config management. It helps you view Firebase projects, inspect parameters across projects, export and import Remote Config JSON, and safely add, update, or delete Remote Config parameters.

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

Install with Homebrew:

```sh
brew tap yumauri/homebrew-tap
brew install --cask fbrcm
```

Install with Scoop on Windows:

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

## First Setup

`fbrcm` uses Google OAuth. You need a Desktop app OAuth client secret JSON:

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
fbrcm login import --from /path/to/client_secret.json
```

After the client secret is imported, authenticate:

```sh
fbrcm login
```

The app opens a browser authorization page and waits for the local OAuth callback. If the browser does not open, copy the printed URL into a browser.

Check current auth files:

```sh
fbrcm login whoami
fbrcm login path
```

## Where Auth Is Stored

By default, `fbrcm` stores per-profile files under your user config and cache directories.

- Client secret: `~/.config/fbrcm/<profile>/client_secret.json`
- Projects cache: `~/.config/fbrcm/<profile>/projects-config.json`
- OAuth token cache: user cache directory, under `fbrcm/<profile>/token.json`

Exact paths:

```sh
fbrcm login path
fbrcm projects path
fbrcm cache path
```

You can override root directories with environment variables:

- `FBRCM_CONFIG_DIR`
- `FBRCM_CACHE_DIR`

Delete auth files:

```sh
fbrcm login purge
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
fbrcm get --filter login
fbrcm get --json
```

Export one project Remote Config:

```sh
fbrcm project export <project-id> --to remote-config.json
```

Import Remote Config:

```sh
fbrcm project import <project-id> --from remote-config.json --dry-run
fbrcm project import <project-id> --from remote-config.json --merge
fbrcm project import <project-id> --from remote-config.json --override
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
```

Delete parameter:

```sh
fbrcm delete old_parameter --project my-project --dry-run
fbrcm delete old_parameter --project my-project --yes
```

Manage caches:

```sh
fbrcm cache list
fbrcm cache purge
fbrcm projects purge
```

## Profiles

Profiles let you keep separate OAuth clients, project caches, and token caches.

```sh
fbrcm profile list
fbrcm profile switch work
fbrcm profile switch personal
fbrcm profile rename old-name new-name
```

First run creates and uses the `default` profile.

## Filtering

Project and parameter filters support mode-prefixed queries:

- No prefix or `~`: fuzzy match.
- `^`: starts with.
- `/`: includes.
- `=`: exact match.

Examples:

```sh
fbrcm projects list --filter '^prod'
fbrcm get --filter '/checkout'
fbrcm get --project '=my-project-id'
```

Several commands also support `--expr` with [expr-lang](https://expr-lang.org/docs/language-definition) expressions for advanced filtering.

## What It Can Do

- Open a TUI for browsing projects and Remote Config parameters
- List Firebase projects available to the authenticated Google account
- Cache project and parameter data locally
- Fetch Remote Config from Firebase
- Show parameters across many projects
- Filter projects and parameters
- Export Remote Config JSON
- Import Remote Config JSON
- Merge imported config into current project config
- Override current config with imported config
- Remove project-specific conditions during import
- Add, update, rename, move, duplicate, and delete parameters
- Edit parameter values as boolean, number, string, or JSON
- Validate and publish Remote Config through Firebase APIs
- Use `--dry-run` on write commands to preview Firebase writes without sending them

## Safety Notes

Use `--dry-run` before imports, updates, adds, and deletes when you are unsure. Write commands print diffs and usually ask for confirmation unless `--yes` is used.

Keep `client_secret.json` and `token.json` private. They grant access through your Google account permissions.
