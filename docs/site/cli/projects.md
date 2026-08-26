# Projects and templates

Projects are local registrations of physical Firebase projects and the
identity used to access them. Forgetting one removes local fbrcm state; it does
not delete or change the Firebase project.

## Discover and inspect projects

```sh
fbrcm projects list --update
fbrcm project show staging
```

`projects list` reads the local registry and optionally synchronizes accessible
projects. `project show` displays one resolved project and its local metadata.

Use `fbrcm doctor` when discovery fails. It checks credentials, connectivity,
API access, permissions, and local storage in one report.

## Use repository aliases

```sh
fbrcm projects aliases set staging acme-staging-42
fbrcm projects aliases set prod acme-production-42
fbrcm projects aliases list
```

Import aliases from a Firebase CLI project file:

```sh
fbrcm projects aliases import --from .firebaserc --dry-run
fbrcm projects aliases import --from .firebaserc
```

Native aliases are stored in the nearest `.fbrcm.toml`; Firebase aliases remain
in `.firebaserc`. Conflicting definitions are rejected.

## Choose client or server templates

Inspect available template types and the current primary:

```sh
fbrcm project templates show staging
```

Set the primary template used by unqualified project selectors:

```sh
fbrcm project templates set staging --primary server
```

An explicit target always wins over the saved preference:

```sh
fbrcm get --project 'client@=staging'
fbrcm get --project 'server@=staging'
```

## Compare environments

```sh
fbrcm projects diff staging prod
```

The comparison includes parameter values, groups, descriptions, conditions,
and condition ordering. Use filters to narrow a large diff to the configuration
you are reviewing.

## Promote selected configuration

Preview a promotion before writing:

```sh
fbrcm projects promote client@staging client@prod --all --dry-run
```

Promotion is selective: it transfers chosen parameters, groups, and conditions
from the source candidate to the target candidate. The source is never
modified. Use `--draft` to stage the target-side result for a later review.

Explicitly qualify targets when promoting between client and server templates:

```sh
fbrcm projects promote client@staging server@prod --all --dry-run
```

## Import a template

```sh
fbrcm project import staging --from remote-config.json --dry-run
```

Import supports merge and replacement workflows, parameter or group selection,
filters, condition cleanup, and conflict choices. The final candidate always
goes through the normal diff and validation path.

Use stdin when another command produces the template:

```sh
generate-config | fbrcm project import staging --from - --dry-run
```

## Export and application defaults

```sh
fbrcm project export staging --to remote-config.json
fbrcm project defaults staging --format plist --to RemoteConfigDefaults.plist
```

Export writes the effective Remote Config template. Defaults asks Firebase for
the selected platform format. Existing destinations require an explicit human
confirmation; JSON mode reports that interaction requirement instead of
prompting.

## Open and forget

```sh
fbrcm project open staging
fbrcm projects forget --filter '=staging'
```

`project open` launches the Remote Config page in Firebase Console. JSON mode
returns the URL without opening a browser. `projects forget` removes
registrations, caches, snapshots, and drafts for the selected local projects
without touching Firebase.
