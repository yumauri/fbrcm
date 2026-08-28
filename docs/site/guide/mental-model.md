# How fbrcm works

fbrcm calls the Firebase Remote Config API from your machine. Published
templates stay in Firebase. fbrcm stores credentials, project registrations,
caches, and drafts locally. It has no intermediary backend.

## Projects and template targets

A Firebase project can expose two independent Remote Config templates:

- `client@project-id` is the template used by Firebase client SDKs;
- `server@project-id` is the server template.

An unqualified project name uses the project's configured primary template,
which defaults to the client template. Use a qualified target whenever the
distinction matters:

```sh
fbrcm get --project 'client@=example-project-id'
fbrcm get --project 'server@=example-project-id'
```

Client and server targets have separate caches, drafts, and version histories.

## Profiles

A profile is an isolated fbrcm workspace. It contains:

- registered authentication identities;
- discovered and selected projects;
- cached templates and version snapshots;
- active drafts.

fbrcm stores the active-profile preference in global configuration. The
preference points to one of these workspaces.

Use separate profiles for unrelated organizations, credentials, or project
sets. Switch interactively with `Ctrl+P`, or use:

```sh
fbrcm profile list
fbrcm profile switch personal
```

## Repository aliases

Aliases give stable environment names to physical Firebase project IDs. Native
aliases live in the nearest `.fbrcm.toml`:

```toml
[projects.aliases]
staging = "example-staging-project-id"
prod = "example-production-project-id"
```

fbrcm also resolves Firebase CLI aliases from `.firebaserc`. Aliases belong to
the repository instead of a profile, so teams can commit them and use the same
names locally and in CI.

```sh
fbrcm projects aliases import --from .firebaserc --dry-run
fbrcm projects diff staging prod
```

Firebase requests, caches, and drafts continue to use canonical project IDs.

## Published state, cache, and drafts

fbrcm distinguishes three states:

| State | Meaning |
| --- | --- |
| Published | The template currently stored by Firebase |
| Cached | A local snapshot used for fast reads and offline work |
| Draft | Local edit intent that has not been published |

Normal reads use a valid draft when one exists; otherwise they use cached
published state and refresh it when requested. A draft contains the edited
template and the base used to calculate local intent and detect conflicts.

## TUI and CLI

The TUI and CLI read the same profiles, projects, caches, drafts, themes, and
configuration.

- Use the TUI when discovery, comparison, or visual review is the main task.
- Use human-readable CLI output for focused terminal operations.
- Add `--json` for scripts and agents that need a stable machine contract.
- Add root `--stateless` when automation must operate directly from an access
  token without reading or writing application-managed local state.

## Remote Config writes are complete-template writes

Firebase validates and publishes a complete Remote Config template even when
you intend to change one parameter. fbrcm therefore turns targeted operations
into a complete candidate template, shows the diff, validates the candidate,
and publishes it with ETag protection.

Even a one-parameter change produces a full candidate template. Use a
[dry run or draft](./safe-changes) to inspect that candidate before publication.
