# Authentication and project discovery

fbrcm stores authentication identities in the active profile. Project discovery
then binds each Firebase project to one of those identities.

## Do not confuse these identifiers

| Term | Example | What it identifies |
| --- | --- | --- |
| fbrcm profile | `personal` | An isolated local workspace |
| Auth ID | `example-auth-name` | The local name fbrcm uses for one credential |
| Google principal | `user@example.com` | The user or service account authenticated by Google |
| Quota project ID | `example-quota-project-id` | The Google Cloud project charged for API usage |
| Firebase project ID | `example-project-id` | The project whose Remote Config is accessed |
| Repository alias | `staging` | An optional local-repository name for a Firebase project ID |

In this command:

```text
fbrcm auth quota-project set <auth-id> <quota-project-id>
```

`auth`, `quota-project`, and `set` are commands. The last two values are
positional arguments. For example:

```sh
fbrcm auth quota-project set example-auth-name example-quota-project-id
```

Here, `example-auth-name` is the auth ID and `example-quota-project-id` is the
quota project ID.

## Required permissions

The relevant permissions are:

| Permission | Resource | Used for |
| --- | --- | --- |
| `serviceusage.services.use` | Quota project | Charging API quota and billing to that project |
| `resourcemanager.projects.get` | Firebase project | Making the project visible during discovery |
| `cloudconfig.configs.get` | Firebase project | Reading Remote Config |
| `cloudconfig.configs.update` | Firebase project | Validating and publishing Remote Config |

The quota permission does not grant Remote Config access to that or any other
project. Managed-feature operations such as deleting an experiment or rollout
may require additional Firebase permissions.

Use `fbrcm doctor` after setup to test the effective credential, quota project,
Cloud Resource Manager access, and both Remote Config permissions. A deliberately
read-only identity can read successfully while `doctor` still reports the
missing `cloudconfig.configs.update` permission.

## Authentication methods

### Google Cloud CLI ADC

Create Application Default Credentials outside fbrcm:

```sh
gcloud auth application-default login
```

Then register and validate them under a local auth ID:

```sh
fbrcm auth add gcloud example-auth-name --quota-project example-quota-project-id
fbrcm auth login example-auth-name
```

If the ADC JSON already contains a `quota_project_id`, fbrcm can use it when no
environment, project, or auth-level setting takes precedence. Passing
`--quota-project` makes the intended fbrcm default explicit.

### OAuth Desktop app

Create a Desktop app OAuth client, download its JSON, and register it:

```sh
fbrcm auth add oauth example-auth-name \
  --from /path/to/client-secret.json \
  --quota-project example-quota-project-id
```

Complete authorization separately:

```sh
fbrcm auth login example-auth-name
```

fbrcm caches tokens locally for later commands. Use `fbrcm auth path example-auth-name`
to inspect the exact client-secret and token paths.

### Service-account key

Register and validate an existing JSON key:

```sh
fbrcm auth add service-account example-service-account-auth \
  --from /path/to/service-account.json \
  --quota-project example-quota-project-id

fbrcm auth login example-service-account-auth
```

fbrcm copies the key into the active profile's private configuration directory.
Protect both the original and the stored copy.

## Why discovery needs a quota project

When fbrcm already knows the target Firebase project, that project can be the
last-resort quota project. The first project-discovery request is different: it
lists accessible projects before a target exists. fbrcm therefore requires an
environment, auth-level, or gcloud ADC quota project before sending that
targetless request.

The simplest setup is to save it while adding the identity:

```sh
fbrcm auth add gcloud example-auth-name --quota-project example-quota-project-id
```

For an existing identity, inspect or change it with:

```sh
fbrcm auth quota-project show example-auth-name
fbrcm auth quota-project set example-auth-name example-quota-project-id
fbrcm auth quota-project unset example-auth-name
```

See [Quota and billing project](/reference/configuration#quota-and-billing-project)
for the complete precedence order.

## Discover and save projects

Synchronize one identity explicitly during setup:

```sh
fbrcm projects update --auth example-auth-name
```

After you configure more identities, synchronize all of them with either:

```sh
fbrcm projects update
fbrcm projects list --update
```

Synchronization keeps existing projects that are temporarily inaccessible and
marks them disabled instead of silently deleting their local state.

## Multiple identities and project bindings

List the configured identities and their saved quota projects:

```sh
fbrcm auth list
```

Discovery records which identities can see each project and chooses a binding.
To bind selected saved projects to a different identity:

```sh
fbrcm auth bind --auth example-service-account-auth --project '=example-project-id'
```

The exact selector limits the operation to one physical project. Binding does
not change IAM at Google; it only selects which configured credential fbrcm
uses for that project.

## Profiles

The default profile is sufficient for one environment. Use separate profiles
when credentials, organizations, or local caches should be isolated:

```sh
fbrcm profile switch personal
fbrcm profile switch customer-a
fbrcm profile list
```

`profile switch` changes the persisted active profile. Root
`--profile <name>` selects an existing profile for one invocation without
changing the active one.

## Remove an identity

Before deleting an identity, rebind or forget projects that depend on it. Then
run:

```sh
fbrcm auth delete example-auth-name
```

This removes the local auth entry and its fbrcm-managed credential files. It
does not revoke an OAuth grant, delete a Google service account, or remove
Google Cloud CLI ADC.
