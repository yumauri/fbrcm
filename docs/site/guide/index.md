# Getting started

fbrcm is a terminal manager for Firebase Remote Config. Use the interactive TUI
to explore and edit projects, or use the CLI for scripts, repeatable operations,
and machine-readable output.

## Install

<ContentTabs
  aria-label="Installation instructions by operating system"
  detect-os
  :tabs="[
    { id: 'macos', label: 'macOS' },
    { id: 'linux', label: 'Linux' },
    { id: 'windows', label: 'Windows' },
    { id: 'source', label: 'From source' }
  ]"
>
<template #macos>

Install with Homebrew:

```sh
brew tap yumauri/tap
brew install --cask fbrcm
```

Alternatively, run the installer directly:

```sh
curl -sSfL https://raw.githubusercontent.com/yumauri/fbrcm/main/install.sh | sh
```

</template>
<template #linux>

Run the installer to download the latest release:

```sh
curl -sSfL https://raw.githubusercontent.com/yumauri/fbrcm/main/install.sh | sh
```

If you use Homebrew on Linux, you can install from the tap instead:

```sh
brew tap yumauri/tap
brew install --cask fbrcm
```

</template>
<template #windows>

Run the installer in PowerShell:

```powershell
irm https://raw.githubusercontent.com/yumauri/fbrcm/main/install.ps1 | iex
```

Or install with Scoop:

```powershell
scoop bucket add yumauri https://github.com/yumauri/scoop-bucket
scoop install fbrcm
```

</template>
<template #source>

Install from source with Go 1.27.0 or newer:

```sh
go install github.com/yumauri/fbrcm@latest
```

</template>
</ContentTabs>

## Connect an identity

Launch fbrcm without arguments:

```sh
fbrcm
```

On a new profile, guided setup offers three authentication methods:

- an OAuth Desktop app client;
- a service-account JSON key; or
- Google Cloud CLI Application Default Credentials.

The selected identity must be able to discover the projects you need and read
their Remote Config templates. Write workflows also require permission
to validate and publish templates.

### Choose a quota project

As part of adding an identity, guided setup asks for a Google Cloud project to
use for quota and billing attribution. This value is required for the first
project discovery: before fbrcm knows which Firebase projects are accessible,
the request is targetless and has no target project ID to use as a fallback for
`X-Goog-User-Project`.

Enter a physical Google Cloud project ID on which the authenticated principal
has `serviceusage.services.use`. fbrcm saves it as the identity's default and
uses it for discovery and whenever a selected project has no more specific
override.

For CLI setup, pass `--quota-project <project-id>` to `fbrcm auth add`, or set
the value on an existing identity:

```sh
fbrcm auth quota-project set main my-quota-project
```

See [Quota and billing project](/reference/configuration#quota-and-billing-project)
for precedence, environment overrides, and per-project settings.

Run the diagnostic whenever you need to check credentials, connectivity,
permissions, or local storage:

```sh
fbrcm doctor
```

## Inspect your first parameter

The TUI opens with the project list on the left and Remote Config data in the
main workspace. Select a project, then choose a parameter to open its details.
Press `?` at any time to search the available actions and shortcuts.

The equivalent CLI workflow is:

```sh
# Discover accessible projects.
fbrcm projects list --update

# Inspect one parameter across production-like projects.
fbrcm get feature_enabled --project '^prod'
```

Project filters support fuzzy, prefix, contains, and exact matching. Use an
exact selector such as `=my-project-id` when a command must target one physical
project.

## Preview before writing

Remote Config publication replaces a complete template. In human mode, fbrcm
shows a diff and asks for confirmation before a write:

```sh
fbrcm update feature_enabled \
  --project '=my-project-id' \
  --type boolean \
  --value true \
  --dry-run
```

Use `--draft` instead of `--dry-run` to save the change locally for later
review. See [Safe changes](./safe-changes) for the complete workflow.

## Next steps

- Read the [mental model](./mental-model) before working across several
  projects or templates.
- Follow the [TUI overview](/tui/) for interactive navigation.
- Open the [CLI overview](/cli/) for command conventions and targeting.
- Use [Agent workflows](/automation/) when calling fbrcm from an LLM, CI job,
  or other automation.
