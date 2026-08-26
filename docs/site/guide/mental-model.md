# Mental model

fbrcm sits between your terminal and the Firebase Remote Config API. It does
not introduce a backend or an additional source of truth: published templates
remain in Firebase, while fbrcm keeps only the local state needed for faster
work and safer changes.

## Projects and template targets

A Firebase project can expose two independent Remote Config templates:

- `client@project-id` is the template used by Firebase client SDKs;
- `server@project-id` is the server template.

An unqualified project name uses the project's configured primary template,
which defaults to the client template. Use a qualified target whenever the
distinction matters:

```sh
fbrcm get --project 'client@=acme-production'
fbrcm get --project 'server@=acme-production'
```

Client and server targets have separate caches, drafts, and version histories.

## Profiles

A profile is an isolated fbrcm workspace. It contains:

- registered authentication identities;
- discovered and selected projects;
- cached templates and version snapshots;
- active drafts; and
- the active-profile preference.

Profiles are useful when you work with unrelated organizations, credentials,
or project sets. Switch interactively with `Ctrl+P`, or use:

```sh
fbrcm profile list
fbrcm profile switch work
```

## Repository aliases

Aliases give stable environment names to physical Firebase project IDs. Native
aliases live in the nearest `.fbrcm.toml`:

```toml
[projects.aliases]
staging = "acme-staging-42"
prod = "acme-production-42"
```

Firebase CLI aliases from `.firebaserc` are resolved too. Because aliases are
repository-scoped and profile-independent, teams can commit them and use the
same names locally and in CI.

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
published state and refresh it when requested. A draft is not just a replacement
template: it retains the base used to calculate local intent and later detect
conflicts.

## TUI and CLI

The TUI and CLI operate on the same profiles, projects, caches, drafts, themes,
and configuration.

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

That is why [dry runs and drafts](./safe-changes) are central rather than
optional polish.
