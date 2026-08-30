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

### Google sign-in

<Badge type="warning" text="Unavailable — Google verification pending" />

Official fbrcm release binaries include fbrcm's OAuth Desktop client. You do
not need to create a Google Cloud OAuth client or download a client JSON:

This method will remain unavailable until Google completes verification of
fbrcm's OAuth application. Use Google Cloud CLI ADC, your own OAuth Desktop app,
or a service-account key in the meantime.

```sh
fbrcm auth add google example-auth-name \
  --quota-project example-quota-project-id
fbrcm auth login example-auth-name
```

The login command opens Google's authorization flow and caches the resulting
token locally.

A plain source build omits the built-in client. Use one of the other three
methods, or follow the repository README to build with a local client.

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

::: details Where to get client-secret.json

1. Open [Google Auth Platform → Clients](https://console.cloud.google.com/auth/clients)
   and select the Google Cloud project that will own the OAuth client.
2. If Google asks you to register the application first, complete the displayed
   setup for the app name, user-support email, audience, and developer contact.
   This is the current Google Auth Platform version of configuring an OAuth
   consent screen.
3. If you select an **External** audience and the publishing status is
   **Testing**, open **Google Auth Platform → Audience** and add the Google
   account that you will use with `fbrcm auth login` under **Test users**.
   fbrcm requests the Cloud Platform scope, so authorization is not covered by
   Google's basic-profile exception.
4. Click **Create client**, choose **Desktop app**, give the client a recognizable
   name, and click **Create**.
5. Download the JSON when the client is created. You may rename the downloaded
   file to `client-secret.json`; fbrcm does not require that exact filename.

For an External app in Testing, Google expires the authorization after seven
days, including an offline refresh token. You must then run `fbrcm auth login`
again. See Google's [app registration](https://support.google.com/cloud/answer/15544987)
and [audience and test-user](https://support.google.com/cloud/answer/15549945)
documentation for the current rules.

Google only exposes newly created client secrets in full at creation time, so
keep the downloaded JSON somewhere private. If it is lost, create or rotate the
client credentials. Never commit this file to a repository. See Google's
[OAuth client instructions](https://support.google.com/cloud/answer/15549257)
for the current console workflow.

:::

```sh
fbrcm auth add oauth example-auth-name \
  --from /path/to/client-secret.json \
  --quota-project example-quota-project-id
```

Complete authorization separately:

```sh
fbrcm auth login example-auth-name
```

fbrcm caches tokens locally for later commands. Use
`fbrcm auth path example-auth-name` to inspect the exact client-secret and token
paths. This bring-your-own method stores its imported client separately from
the built-in `google` method <Badge type="warning" text="Unavailable — Google verification pending" />.

### Service-account key

Register and validate an existing JSON key:

::: details Where to get service-account.json

1. Open [IAM & Admin → Service Accounts](https://console.cloud.google.com/iam-admin/serviceaccounts)
   and select the project that owns the service account.
2. Create a service account, or open the email address of an existing one.
3. Grant the service account roles containing the permissions listed in
   [Required permissions](#required-permissions) on the relevant quota and
   Firebase projects.
4. Open **Keys**, choose **Add key → Create new key**, select **JSON**, and click
   **Create**.
5. The browser downloads the key. You may rename it to `service-account.json`;
   fbrcm does not require that exact filename.

The private key can be downloaded only when it is created. Store it securely
and never commit it to a repository. If key creation is unavailable, an
organization policy may prohibit service-account keys; ask the Google Cloud
administrator rather than weakening that policy without review. See Google's
[service-account key instructions](https://cloud.google.com/iam/docs/keys-create-delete)
for details.

:::

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
fbrcm auth add google example-auth-name --quota-project example-quota-project-id
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
