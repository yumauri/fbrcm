# Troubleshooting

Start with the integrated diagnostic:

```sh
fbrcm doctor
fbrcm doctor --json
```

It checks credentials, connectivity, Firebase and Cloud Resource Manager API
access, permissions, quota-project configuration, and local storage.

`doctor` expects both `cloudconfig.configs.get` and
`cloudconfig.configs.update` on cached Firebase projects. A deliberately
read-only identity can still use read commands even though the update-permission
check fails.

## No projects appear

Synchronize the registry:

```sh
fbrcm projects list --update
```

If discovery still fails, verify that the selected identity can list projects
and read Remote Config. A configured quota project also requires
`serviceusage.services.use` for the authenticated principal.

## Authentication needs attention

Inspect identities and log in again:

```sh
fbrcm auth list
fbrcm auth login example-auth-name
fbrcm auth quota-project show example-auth-name
```

Replace `example-auth-name` with the auth ID shown by `fbrcm auth list`.

OAuth may require a browser in human mode. JSON mode never opens one and
returns structured interaction details instead. Service-account and gcloud
identities validate their existing credentials without an OAuth browser flow.

If project discovery reports that a quota project is required, set an
auth-level default. The two positional arguments are the local auth ID and the
physical quota project ID:

```sh
fbrcm auth quota-project set example-auth-name example-quota-project-id
```

For a single cached Firebase project, inspect or override the selection with:

```sh
fbrcm project quota-project show example-project-id
fbrcm project quota-project set example-project-id example-quota-project-id
fbrcm project quota-project unset example-project-id
```

ADC `quota_project_id` is considered only for gcloud identities. `fbrcm doctor`
verifies both resolution and `serviceusage.services.use` access.

## Reads look stale

Force a live cache revalidation:

```sh
fbrcm get --project '=example-project-id' --update
```

Check whether a local draft is active, because normal effective reads include
draft state. Use `draft diff` to inspect it, or `draft discard` only when that
local intent is no longer needed.

## Offline mode appeared unexpectedly

The TUI enters offline mode after its startup connectivity probe fails. The CLI
enters offline mode whenever `FBRCM_OFFLINE` is defined, even as an empty value
or `0`.

Check proxy variables and unset the override when live access is intended:

```sh
unset FBRCM_OFFLINE
```

## A write reports a conflict

Another client changed Firebase after fbrcm loaded the template, or a draft
rebase found overlapping local and remote edits. Refresh current state and
review again. Do not blindly retry a complete-template publication.

Draft conflicts preserve the draft. Use:

```sh
fbrcm draft diff example-project-id --against current
```

## An editor does not return

GUI editors must wait for the staged file to close:

```sh
FBRCM_EDITOR="code --wait" fbrcm
```

For CLI configuration editing, the same setting applies to `fbrcm config edit`.

## The TUI is too small

The minimum supported terminal size is 80 columns by 20 rows. Enlarge the
window or reduce terminal font size. Human CLI tables adapt to narrow widths
independently.

## JSON automation did not perform an action

Inspect `outcome`, `exit_code`, and the structured problems. Status 10 means
the operation requires explicit interaction such as `--yes`, an import merge
choice, OAuth authorization, or overwrite approval. Follow the published
remediation strategy only after checking authorization and target scope.
