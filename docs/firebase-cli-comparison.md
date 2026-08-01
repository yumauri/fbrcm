# fbrcm vs. Firebase CLI

Both [fbrcm](../README.md) and the
[official Firebase CLI](https://firebase.google.com/docs/cli) can read and
publish Firebase Remote Config. They are optimized for different workflows:

- **Firebase CLI** is a general Firebase deployment tool. It works especially
  well when a repository already uses `firebase.json`, `.firebaserc`, project
  aliases, and deployment hooks for several Firebase products.
- **fbrcm** is a Remote Config workbench. It is designed for interactive
  inspection, targeted edits, local drafts, comparisons, and promotion across
  many projects and both client and server templates. Repository hooks can
  enforce policy across every fbrcm publication path.

This comparison covers the documented Firebase CLI surface and
`firebase-tools` 15.25.1 as of August 1, 2026. Consult the
[current Firebase CLI reference](https://firebase.google.com/docs/cli) when
using a later release.

## Which tool should I use?

Use Firebase CLI when:

- Remote Config is one part of a larger Firebase deployment;
- the complete client template is maintained in source control;
- the team relies on Firebase's active-project workflow and deploy hooks;
- hooks must coordinate Remote Config with deployments of other Firebase
  products;
- one command should deploy Hosting, Functions, Rules, Remote Config, and
  other Firebase resources.

Use fbrcm when:

- you need to inspect or change the same parameter across many projects;
- you want to review typed, item-level changes without editing template JSON;
- changes should be staged in local drafts and rebased safely before publish;
- you compare or promote Remote Config between environments;
- you want repository aliases without introducing an implicit active project;
- you manage client and server templates from the same terminal workspace;
- repository hooks should validate the exact generated Remote Config candidate
  before any direct, draft, import, promotion, restore, or rollback publication;
- you want an interactive TUI as well as automation-friendly CLI commands.

The tools can be used together. A team can keep its canonical template in a
Firebase CLI repository while using fbrcm for investigation, comparison,
careful production edits, exports, and migration work.

## Feature parity

**✅ Yes** means the workflow has a dedicated command or interface.
**⚠️ Partial** means it is possible, but requires editing a complete template,
scripting separate invocations, or accepting a narrower workflow. **❌ No**
means the workflow is unavailable, while **➖ No direct equivalent** indicates
that the other tool approaches the same concern differently.

| Capability | Firebase CLI | fbrcm | Notes |
| --- | --- | --- | --- |
| Read current client template | ✅ | ✅ | Firebase CLI uses `remoteconfig:get`; fbrcm uses `project export` for the complete JSON or `get` for parameter views. |
| Publish complete client template | ✅ | ✅ | Firebase CLI deploys the file configured in `firebase.json`; fbrcm imports a selected file or stdin. |
| Read and publish server templates | ❌ | ✅ | fbrcm accepts explicit `server@project-id` targets and keeps their state separate from client templates. |
| List template versions | ✅ | ✅ | Both read Firebase version history. |
| Retrieve or export a historical version | ✅ | ✅ | Both can write a selected version to a file. |
| Roll back to a retained Firebase version | ✅ | ✅ | fbrcm adds a reviewed diff and dry-run support. |
| Restore a locally cached expired version | ❌ | ✅ | fbrcm can republish an immutable local snapshot after Firebase no longer returns it. |
| Download application defaults | ❌ | ✅ | fbrcm downloads JSON, Android XML, or Apple plist defaults. |
| List, inspect, and delete A/B tests | ✅ | ✅ | fbrcm also correlates experiments with their published parameter bindings and variant values. |
| List, inspect, and delete rollouts | ✅ | ✅ | fbrcm also correlates rollout metadata with published parameter bindings. |
| Inspect personalizations | ❌ | ✅ | fbrcm lists personalization IDs and bindings visible in the published template. |
| Interactive terminal UI | ❌ | ✅ | Running `fbrcm` without arguments opens the TUI. |
| Multi-project parameter table | ⚠️ | ✅ | Firebase CLI can be scripted one project at a time; fbrcm resolves and combines project selections directly. |
| Batch selection by project name or ID | ⚠️ | ✅ | fbrcm supports repeated fuzzy, prefix, contains, exact, and expression selectors. |
| Parameter search and typed expression filters | ❌ | ✅ | fbrcm filters on keys, descriptions, values, conditions, types, and project context. |
| Add, update, delete, or duplicate individual parameters | ⚠️ | ✅ | Firebase CLI publishes an edited complete template; fbrcm exposes typed item-level mutations. |
| Manage parameter groups | ⚠️ | ✅ | fbrcm adds, edits, renames, and explicitly removes groups while preserving empty groups elsewhere. |
| Manage conditions and evaluation priority | ⚠️ | ✅ | fbrcm edits definitions, colors, ordering, references, and conditional values. |
| Transform Remote Config through stdin | ❌ | ✅ | `get`, `add`, `update`, and `delete` can operate as JSON pipeline stages. |
| Preview a structured Remote Config diff | ❌ | ✅ | fbrcm diffs drafts, versions, projects, imports, mutations, and promotions. |
| Compare two projects or template types | ❌ | ✅ | `fbrcm projects diff` supports filters and CI exit codes. |
| Promote selected changes between projects | ❌ | ✅ | fbrcm handles required conditions, group descriptions, optional pruning, and target revalidation. |
| Local drafts | ❌ | ✅ | Drafts retain an immutable base and compose multiple related edits. |
| Three-way rebase before publish | ❌ | ✅ | fbrcm rebases draft intent over current Firebase state and preserves conflicts for review. |
| Item-level dry run | ⚠️ | ✅ | Firebase CLI has a generic deploy dry run; fbrcm previews individual mutations, imports, promotions, drafts, and version changes. |
| ETag-protected publication | ✅ | ✅ | Both use Firebase concurrency protection. |
| Machine-readable JSON | ✅ | ✅ | Firebase CLI has a global `--json`; fbrcm defines stable output contracts per operation. |
| Cache and offline inspection | ❌ | ✅ | fbrcm keeps project, template, and immutable version snapshots for offline work. |
| Credential health and permission diagnostics | ⚠️ | ✅ | `fbrcm doctor` checks identities, storage, connectivity, APIs, and Remote Config read/update permissions. |
| Multiple local authentication identities | ✅ | ✅ | Firebase CLI selects accounts globally; fbrcm binds cached projects to identities inside isolated profiles. |
| Repository project aliases | ✅ | ✅ | fbrcm reads Firebase CLI aliases from `.firebaserc`, merges native `.fbrcm.toml` aliases with conflict detection, supports explicit import, and composes aliases with client/server template targets. |
| Repository-local configuration | ✅ | ✅ | Firebase CLI uses `firebase.json` and `.firebaserc`; fbrcm discovers the nearest `.fbrcm.toml` and deeply overlays it on user-wide `config.toml`. |
| Pre- and post-publication hooks | ✅ | ✅ | Firebase CLI defines `predeploy` and `postdeploy` hooks for deployable resources in `firebase.json`. fbrcm defines Remote Config `pre_publish` and `post_publish` hooks, supplies current/candidate/context JSON files, and requires explicit trust for repository hooks. |
| User-authored Remote Config change note | ❌ | ✅ | fbrcm exposes `--change-note`, stores `change_note` with drafts, prompts in TUI publication reviews, and maps it to the REST API's writable `version.description`. Firebase CLI's `--message` is used for Hosting release comments. |
| Deploy other Firebase products | ✅ | ❌ | fbrcm intentionally focuses on Remote Config. |

## Command equivalents

Firebase CLI normally selects a project with `--project <id-or-alias>` or the
active `.firebaserc` alias. fbrcm commands below use an explicit project or
template target. An unqualified fbrcm project selects its configured primary
template; use `client@project-id` or `server@project-id` when needed.

| Goal | Firebase CLI | fbrcm |
| --- | --- | --- |
| Initialize the local workflow | `firebase init remoteconfig` | Run `fbrcm` for guided setup, or use `fbrcm auth add ...` followed by `fbrcm projects update` |
| Export current template | `firebase --project PROJECT remoteconfig:get -o FILE` | `fbrcm project export PROJECT --to FILE` |
| Print current template | `firebase --project PROJECT remoteconfig:get` | `fbrcm project export PROJECT` |
| Inspect current parameters | Edit or inspect exported JSON | `fbrcm get --project '=PROJECT'` |
| Export historical version | `firebase --project PROJECT remoteconfig:get -v VERSION -o FILE` | `fbrcm versions export PROJECT VERSION --to FILE` |
| Show version metadata | Included in `remoteconfig:get` output | `fbrcm versions show PROJECT VERSION` |
| List recent versions | `firebase --project PROJECT remoteconfig:versions:list --limit N` | `fbrcm versions list PROJECT --limit N` |
| List every retained version | `firebase --project PROJECT remoteconfig:versions:list --limit 0` | `fbrcm versions list PROJECT --all` |
| Publish template file | `firebase --project PROJECT deploy --only remoteconfig` | `fbrcm project import PROJECT --from FILE [--change-note TEXT]` |
| Preview template-file deployment | `firebase --project PROJECT deploy --only remoteconfig --dry-run` | `fbrcm project import PROJECT --from FILE --dry-run` |
| Stage template file without publishing | No equivalent | `fbrcm project import PROJECT --from FILE --draft` |
| Configure publication hooks | Add `predeploy` or `postdeploy` under the resource in `firebase.json` | Add `hooks.pre_publish` or `hooks.post_publish` to `.fbrcm.toml` or global `config.toml` |
| Inspect repository hook trust | No explicit trust workflow | `fbrcm hooks status` or `fbrcm hooks fingerprint` |
| Trust repository hooks | Repository configuration executes as part of deployment | `fbrcm hooks trust`, or pin `FBRCM_HOOK_TRUST` to the expected fingerprint in CI |
| Roll back to version | `firebase --project PROJECT remoteconfig:rollback -v VERSION` | `fbrcm versions rollback PROJECT VERSION` |
| Restore cached version | No equivalent | `fbrcm versions restore PROJECT VERSION` |
| List experiments | `firebase --project PROJECT remoteconfig:experiments:list` | `fbrcm experiments list PROJECT` |
| Show experiment | `firebase --project PROJECT remoteconfig:experiments:get ID` | `fbrcm experiments show PROJECT ID` |
| Delete experiment | `firebase --project PROJECT remoteconfig:experiments:delete ID` | `fbrcm experiments delete PROJECT ID` |
| List rollouts | `firebase --project PROJECT remoteconfig:rollouts:list` | `fbrcm rollouts list PROJECT` |
| Show rollout | `firebase --project PROJECT remoteconfig:rollouts:get ID` | `fbrcm rollouts show PROJECT ID` |
| Delete rollout | `firebase --project PROJECT remoteconfig:rollouts:delete ID` | `fbrcm rollouts delete PROJECT ID` |

The two filtering syntaxes for managed features are not interchangeable.
Firebase CLI's experiment and rollout `--filter` accepts a full resource-name
filter. fbrcm filters experiment display names locally and otherwise uses its
shared project and parameter filtering conventions.

## fbrcm workflows without a Firebase CLI equivalent

### Change one parameter across projects

```sh
fbrcm update feature_enabled \
  --project '^mobile-prod-' \
  --type boolean \
  --value true \
  --dry-run
```

Remove `--dry-run` to review and publish each project, or replace it with
`--draft` to stage the changes locally.

### Compare environments

```sh
fbrcm projects diff staging-project production-project --exit-code
```

The exit status is `0` for no differences, `1` for differences, and `2` for an
error, making the command suitable for CI drift checks.

### Promote selected changes

```sh
fbrcm projects promote staging-project production-project --interactive
```

Promotion can select parameters, conditions, and group descriptions, include
their dependencies, and optionally prune target-only items.

### Stage, review, and publish a draft

```sh
fbrcm update checkout_enabled \
  --project '=production-project' \
  --type boolean \
  --value true \
  --draft

fbrcm draft diff production-project --against current
fbrcm draft publish production-project \
  --change-note "Enable checkout v2 for production"
```

Draft publication fetches current Firebase state, performs a three-way merge,
validates the exact reviewed candidate, and publishes it with ETag protection.

### Work with a server template

```sh
fbrcm get --project 'server@=backend-project'
fbrcm projects diff client@backend-project server@backend-project
```

Client and server targets have independent template caches, drafts, and version
histories.

### Enforce repository publication policy

Commit a local `.fbrcm.toml`:

```toml
[hooks]
timeout = "2m"
pre_publish = ["./scripts/validate-remote-config.sh"]
post_publish = ["./scripts/notify-remote-config-published.sh"]
```

The validation script reads the exact candidate through
`FBRCM_CANDIDATE_FILE`. After reviewing the commands, each developer explicitly
trusts that definition:

```sh
fbrcm hooks status
fbrcm hooks trust
```

`pre_publish` also runs for publication dry runs, so the same policy can be
exercised in CI without writing Firebase state. Saving a local draft does not
run hooks; publishing it does.

## Important behavioral differences

### Complete-template deployment vs. targeted operations

Firebase CLI treats Remote Config primarily as a complete repository artifact:
edit the template referenced by `firebase.json`, then deploy it. fbrcm can
publish a complete imported template, but it also performs parameter-, group-,
and condition-level operations with a generated diff before publication.

### Project selection

Firebase CLI is directory-oriented and uses an active project, `--project`, and
shared aliases in `.firebaserc`. fbrcm reads those aliases and merges them with
native aliases in `.fbrcm.toml`, rejecting conflicting targets. It discovers
accessible projects, caches them per profile, and supports selecting multiple
projects or template targets within one invocation. An fbrcm alias does not
create an implicit active project: unqualified aliases use the active profile's
configured primary template, while `client@alias` and `server@alias` are
deterministic.

### Safety and local state

Both tools use Firebase ETags. fbrcm additionally keeps local drafts and
historical snapshots. Immediate fbrcm writes refuse to bypass an existing
draft, and draft publication rebases local intent rather than silently
overwriting unrelated remote changes.

### Hook scope and trust

Firebase CLI deployment hooks belong to deployable resources in
`firebase.json` and can coordinate several Firebase products. fbrcm hooks are
intentionally narrower: they run around every Remote Config publication made by
fbrcm, including the TUI, direct mutations, draft publication, imports,
promotions, restores, and native rollbacks.

Because fbrcm builds candidates in memory, it gives hooks private temporary
`current.json`, `candidate.json`, and `context.json` files; post-hooks also
receive `published.json`. Hooks validate or perform side effects and cannot
transform the published candidate. A failed pre-hook prevents publication. A
failed post-hook is reported as a published partial success and is not safe to
retry blindly.

Global fbrcm hooks are user-owned and trusted automatically. Hooks introduced
by `.fbrcm.toml` require an explicit path-and-content fingerprint trust record,
which is invalidated when the local hook definition changes. CI can pin that
fingerprint with `FBRCM_HOOK_TRUST`.

### Scope

Firebase CLI is the appropriate umbrella tool for a complete Firebase
application deployment. fbrcm deliberately does not deploy Hosting, Functions,
Rules, Data Connect, or other Firebase products; its narrower scope enables a
deeper Remote Config interface.
