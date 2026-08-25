# Privacy Policy for fbrcm

Last updated: August 25, 2026

fbrcm is a local command-line and terminal application for managing Firebase
Remote Config. It has no developer-operated backend, does not include telemetry
or advertising, and does not send Google user data to the fbrcm developer.
Google API requests are made directly from the user's machine using credentials
selected by the user.

This policy applies to the official fbrcm application. Forks, modified builds,
and programs or publication hooks configured by a user may behave differently.

## Google data fbrcm accesses

fbrcm requests the following OAuth scope:

- `https://www.googleapis.com/auth/cloud-platform`

This scope is broad, but fbrcm limits its Google API use to the functionality
described below. The scope does not itself grant access to a project or action;
the authenticated identity's Google Cloud IAM permissions continue to control
which resources fbrcm can read or change.

fbrcm accesses:

- Google Cloud project metadata available to the authenticated identity,
  including project names, IDs, numbers, lifecycle state, ETags, and update
  times;
- the result of permission checks for the project operations fbrcm needs;
- Firebase Remote Config client and server templates, including parameters,
  values, descriptions, groups, conditions, ETags, and version metadata;
- Remote Config defaults and version history; and
- Remote Config managed-feature information exposed by the API, including
  experiments, rollouts, and personalization bindings.

fbrcm does not call Google profile APIs to retrieve a user's contacts, email,
calendar, Drive files, or other unrelated Google Account data, and it does not
access other Google Cloud services merely because the OAuth scope could permit
them.

## Why the broad OAuth scope is requested

Firebase Remote Config identifies
`https://www.googleapis.com/auth/firebase.remoteconfig` as a narrower scope in
some API responses and documentation, but Google's interactive OAuth consent
flow rejects that scope with `invalid_scope`. The grantable
`https://www.googleapis.com/auth/firebase` scope is rejected by the Remote
Config API with `insufficient_scope`. Google Cloud CLI Application Default
Credentials also require `cloud-platform` during user authorization. A
[public Firebase discussion](https://groups.google.com/g/firebase-talk/c/a8H9GcGiYuA)
reports the same narrower-scope problem.

As a result, `cloud-platform` is currently the only working scope for fbrcm's
interactive OAuth and user Application Default Credentials flows. fbrcm will
move to a narrower usable scope if Google makes one available for these flows.

## How fbrcm uses Google data

Google data is used only to provide user-invoked fbrcm features. These features
include discovering and displaying projects; reading, searching, comparing,
and exporting Remote Config data; creating local caches and drafts; validating
candidate templates; and applying changes selected by the user.

Depending on the command, an explicit user action can publish or restore a
Remote Config template, roll back a version, or delete a selected Remote Config
experiment or rollout. fbrcm does not perform these remote mutations for
advertising, profiling, analytics, or any purpose unrelated to its visible
Remote Config management features.

## Local storage

fbrcm stores application state only on the user's machine unless the user
explicitly sends or copies it elsewhere. The exact paths can be inspected with
`fbrcm doctor`, `fbrcm profile path <profile>`, and
`fbrcm auth path <auth-id>`.

By default, configuration is stored under `.config/fbrcm` in the user's home
directory, unless an XDG or fbrcm path override is configured. Cache storage
uses the operating system's user-cache directory, such as
`~/Library/Caches/fbrcm` on macOS, `~/.cache/fbrcm` on many Linux systems, or
the local application-data directory on Windows. The `FBRCM_CONFIG_DIR` and
`FBRCM_CACHE_DIR` environment variables can override these roots.

Stored data can include:

- OAuth client configuration and service-account key files in the profile
  configuration directory;
- OAuth access and refresh tokens in a profile cache file named `token.json`;
- the local authentication registry and project metadata;
- complete Remote Config template caches and historical version snapshots; and
- complete base and edited Remote Config templates in local drafts.

OAuth tokens are stored as plaintext JSON files, not in an operating-system
keychain. fbrcm creates credential, configuration, cache, and draft files with
private permissions where the operating system supports them. Google Cloud CLI
Application Default Credentials remain managed by Google Cloud CLI; fbrcm uses
them through Google's authentication library and does not copy them into an
fbrcm credential file. A token supplied to stateless mode through
`FBRCM_GOOGLE_ACCESS_TOKEN` is used in memory and is not persisted by fbrcm.

Command output and files written to an explicitly selected export destination
may contain Google data. Terminal software, shell history, CI systems, backups,
or other programs chosen by the user may retain that output independently of
fbrcm.

## Network transfers and sharing

For its Google integration, fbrcm sends OAuth credentials and Google data only
to Google's OAuth services, the Google Cloud Resource Manager API, and the
Firebase Remote Config API as needed for the requested operation. These
transfers are handled under Google's own terms and privacy policies. fbrcm does
not sell Google user data, share it with advertisers or data brokers, or
transfer it to the fbrcm developer or a developer-operated service.

Users can configure publication hooks. For each publication, fbrcm provides
those local commands with temporary files containing the current, candidate,
and, after publication, published Remote Config templates. A hook can transmit
that data wherever its author programmed it to do so. Repository hooks require
explicit trust, but global hooks are controlled by the user and are trusted
automatically. Users are responsible for the privacy behavior of hooks,
proxies, scripts, pipes, and other programs they configure around fbrcm.

## Security

fbrcm uses HTTPS for Google API and OAuth traffic, requests only the Google data
needed by its implemented features, applies IAM authorization at Google, uses
private local file permissions where supported, and redacts known credential
fields from application error output and HTTP logs. Because locally stored
tokens and templates are not encrypted by fbrcm, users should protect their
account, device, backups, configuration directory, and cache directory against
unauthorized access.

## Retention and deletion

The fbrcm developer does not retain Google user data because fbrcm does not send
that data to a developer-operated system. Local data remains until it is
replaced or the user deletes it. A credential may expire or be revoked without
its local file being removed. In particular, cached historical templates and
drafts do not have an automatic retention deadline and can remain after Firebase
no longer retains the corresponding version.

Users can remove local data with the following commands:

- `fbrcm auth delete <auth-id>` removes an fbrcm identity and its locally stored
  OAuth client, token, or service-account key files;
- `fbrcm cache clear` removes cached Remote Config version snapshots;
- `fbrcm draft discard --all` removes all drafts in the active profile;
- `fbrcm projects forget` removes tracked projects and their associated
  template caches, version snapshots, and drafts; and
- `fbrcm profile delete <profile>` removes a non-active profile's complete
  configuration and cache directories.

Users can also inspect the paths listed above and delete the fbrcm configuration
and cache roots manually. Removing local credentials does not revoke an OAuth
grant at Google. To revoke the grant, visit
[Google Account third-party access](https://myaccount.google.com/permissions).
Service-account keys and Google Cloud CLI credentials must also be revoked or
removed through the corresponding Google Cloud or Google Cloud CLI controls.

Deleting local fbrcm data does not delete Firebase projects or Remote Config
data already published to Google. Remote data can be changed or deleted only
through the applicable Firebase or Google Cloud controls and APIs.

## Changes to this policy

Changes to fbrcm's handling of Google user data will be reflected in this
policy. When required, fbrcm will also update its OAuth disclosures and request
new consent before using Google user data for a materially different purpose.

## Contact

For privacy questions, email [yumaa.verdin@gmail.com](mailto:yumaa.verdin@gmail.com)
or open an issue in the
[fbrcm issue tracker](https://github.com/yumauri/fbrcm/issues).
