# Getting started

After these steps, fbrcm will have one working credential, a saved project
list, and a cached Remote Config template. The last step previews a change
without publishing it. Choose the TUI setup or the CLI-only setup.

## What you need

Two Google Cloud project roles appear during setup:

- **Firebase project.** This is the project whose Remote Config you want to manage.
  fbrcm discovers accessible Firebase projects, so you do not need to know all
  of their IDs in advance.
- **Quota project.** This is the project Google should charge for API quota and
  billing. You must know this physical project ID before the first discovery,
  and your Google identity needs `serviceusage.services.use` on it.

They may be the same physical Google Cloud project. The quota project does not
grant access to Remote Config; your identity still needs the appropriate
permissions on every Firebase project it manages.

If you do not have a usable quota project, ask your Google Cloud administrator
which project to use and for `serviceusage.services.use` on it. fbrcm cannot
perform the first targetless project discovery without one.

You also need one of these credential sources:

- fbrcm's built-in Google OAuth client in an official release;
- a Google OAuth Desktop app client JSON;
- a service-account key JSON; or
- Google Cloud CLI Application Default Credentials (ADC).

If these terms are unfamiliar, read [Authentication and project discovery](./authentication)
before choosing a setup path.

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

A plain source build omits fbrcm's official Google OAuth client. The other three
authentication methods remain available. Follow the repository README to build
the `google` method with a local client.

</template>
</ContentTabs>

Confirm that the binary is available:

```sh
fbrcm --version
```

## Option 1: guided setup in the TUI

Start fbrcm without arguments:

```sh
fbrcm
```

On a new profile, the setup screen asks you to:

1. Choose built-in Google sign-in, imported OAuth, service-account, or gcloud
   ADC authentication.
2. Import or validate the selected credentials.
3. Enter the physical Google Cloud project ID to use as the quota project.
4. Complete browser authorization when OAuth requires it.
5. Let fbrcm discover and save the accessible Firebase projects.

After the workspace opens, select a project in the left panel and select a
parameter in the main panel. Press `?` anywhere to find available actions
without memorizing shortcuts.

Run the diagnostic from another terminal if setup reports an error:

```sh
fbrcm doctor
```

Continue with the [TUI overview](/tui/) when you want to edit or compare data
interactively.

## Option 2: setup using only the CLI

The examples below use these concrete names:

| Example value | Meaning |
| --- | --- |
| `example-auth-name` | The local name fbrcm uses for the credential |
| `example-quota-project-id` | The Google Cloud project ID used for quota and billing |
| `example-project-id` | The Firebase project ID you want to manage |

`example-auth-name` is not a command or a Google account name. It is the name
fbrcm uses to refer to this credential.

### 1. Add one authentication identity

Choose exactly one of the following methods.

<ContentTabs
  aria-label="Authentication method"
  :tabs="[
    { id: 'google', label: 'Continue with Google' },
    { id: 'gcloud', label: 'Google Cloud CLI' },
    { id: 'oauth', label: 'OAuth Desktop app' },
    { id: 'service-account', label: 'Service account' }
  ]"
>
<template #google>

Official fbrcm releases include the application's shared OAuth client. Add an
identity without creating a Google Cloud client or downloading a client JSON:

```sh
fbrcm auth add google example-auth-name \
  --quota-project example-quota-project-id

fbrcm auth login example-auth-name
```

The final command opens Google's consent flow when authorization is needed.

</template>
<template #gcloud>

If you already use the Google Cloud CLI, run:

```sh
gcloud auth application-default login

fbrcm auth add gcloud example-auth-name \
  --quota-project example-quota-project-id

fbrcm auth login example-auth-name
```

The final command validates that fbrcm can load the ADC credentials. It does
not start another fbrcm-managed OAuth flow.

</template>
<template #oauth>

Create a Desktop app client in
[Google Cloud OAuth clients](https://console.cloud.google.com/auth/clients),
download its JSON file, and run:

```sh
fbrcm auth add oauth example-auth-name \
  --from /path/to/client-secret.json \
  --quota-project example-quota-project-id

fbrcm auth login example-auth-name
```

`auth login` opens the Google consent flow when no usable cached token exists.
Add `--noopen` if you want fbrcm to print the authorization URL without opening
the browser automatically.

</template>
<template #service-account>

Use a service-account key when your environment requires one:

```sh
fbrcm auth add service-account example-service-account-auth \
  --from /path/to/service-account.json \
  --quota-project example-quota-project-id

fbrcm auth login example-service-account-auth
```

The login command validates the stored key; it does not open a browser.

</template>
</ContentTabs>

### 2. Discover projects

Discover accessible Firebase projects and save them in the active profile:

```sh
fbrcm projects update
```

The command uses every configured identity. At this point, that is the identity
you just added. Later, `fbrcm projects list` reads the saved list and
`fbrcm projects list --update` refreshes it before printing.

### 3. Inspect Remote Config

First list the saved projects:

```sh
fbrcm projects list
```

Then replace `example-project-id` with one Firebase project ID from that output:

```sh
fbrcm get --project '=example-project-id'
```

This lists parameters from exactly that project. To inspect one parameter, add
its key:

```sh
fbrcm get feature_enabled --project '=example-project-id'
```

The leading `=` means exact project matching. It prevents a short name from
accidentally matching several projects.

### 4. Check the complete setup

```sh
fbrcm doctor
```

The report checks local files, credentials, quota-project resolution, project
discovery, and Remote Config access. It also checks
`serviceusage.services.use`, `cloudconfig.configs.get`, and
`cloudconfig.configs.update`.

## Preview your first change

Remote Config publication replaces a complete template. Start with a dry run,
which fetches and validates the candidate but does not publish or save it:

```sh
fbrcm update feature_enabled \
  --project '=example-project-id' \
  --type boolean \
  --value true \
  --dry-run
```

Inspect the matched-item count and diff. When the result is correct, either
rerun without `--dry-run` for an immediate reviewed publication or replace it
with `--draft` to save the change locally.

## Where to go next

- [Authentication and project discovery](./authentication) explains auth IDs,
  quota projects, multiple identities, and project bindings.
- [How fbrcm works](./mental-model) explains profiles, aliases, caches, drafts,
  and client/server templates.
- [Safe changes](./safe-changes) walks through dry runs, drafts, conflict
  handling, and publication.
- [CLI overview](/cli/) explains command syntax, selectors, and output modes.
- [Automation and agents](/automation/) covers stable JSON execution.
