# History and managed features

fbrcm exposes Remote Config version history and the Firebase-managed features
bound to published client templates.

## List and inspect versions

```sh
fbrcm versions list staging
fbrcm versions show staging 42
```

Version lists combine Firebase history with locally retained snapshots when
available. A cached historical template may outlive Firebase's retention
window, so consider exporting important versions before clearing the cache.

## Compare versions

```sh
fbrcm versions diff staging 41 42
```

Omit the second version to compare a historical version with the current
effective state. Parameter, group, condition, expression, and search filters
can narrow the diff.

## Export a version

```sh
fbrcm versions export staging 42 --to remote-config-v42.json
```

Export is useful for audit artifacts, recovery, or review outside fbrcm.

## Roll back or restore

```sh
fbrcm versions rollback staging 42 --dry-run
fbrcm versions restore staging 42 --dry-run
```

- **Rollback** asks Firebase to roll the project back to a retained version.
- **Restore** republishes a locally cached historical snapshot as a new
  version, which remains possible after Firebase no longer retains the source
  version.

Both are complete-template writes and use the standard preview, validation,
confirmation, and ETag safeguards.

## A/B tests

```sh
fbrcm experiments list staging
fbrcm experiments show staging experiment-id
fbrcm experiments delete staging experiment-id
```

List and show combine experiment metadata with bindings found in the published
client template. Delete is a remote Firebase operation and follows the normal
confirmation and machine-contract rules.

## Rollouts

```sh
fbrcm rollouts list staging
fbrcm rollouts show staging rollout-id
fbrcm rollouts delete staging rollout-id
```

Rollout views combine the public rollout API with parameter-value bindings from
the published template.

## Personalizations

```sh
fbrcm personalizations list staging
fbrcm personalizations show staging personalization-id
```

Firebase exposes personalization bindings through the template but does not
provide candidate values or result metrics through this API. Personalization
commands are therefore read-only.

Managed-feature commands apply to client templates. Server targets are omitted
because Firebase exposes these features under the client Remote Config
namespace.
