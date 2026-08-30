# Terms of Service for fbrcm

Last updated: August 30, 2026

These terms apply to the official fbrcm application, its built-in Google OAuth
client, and the fbrcm documentation website. fbrcm is a local command-line and
terminal application for managing Firebase Remote Config. It does not provide
user accounts or a developer-operated backend.

By using the official application or website, you agree to these terms. If you
do not agree, do not use them.

## Open-source license

The fbrcm source code and software are provided under the [MIT License](LICENSE).
That license governs your rights to use, copy, modify, and distribute the
software. These terms do not limit the permissions granted by the MIT License.

## Google and Firebase services

fbrcm communicates directly from your machine with Google OAuth, Google Cloud,
and Firebase services. Your use of those services remains subject to the terms,
policies, quotas, billing rules, and access controls imposed by Google and by
the organizations that own the projects you access.

fbrcm is an independent project. It is not endorsed by or affiliated with
Google. Firebase and Google Cloud are trademarks of Google LLC.

Google services and APIs may change, become unavailable, or behave differently
across projects and accounts. The fbrcm developer does not control those
services and cannot guarantee their continued availability or compatibility.

## Your responsibilities

You are responsible for:

- using fbrcm only with accounts, credentials, quota projects, Firebase
  projects, and Remote Config data that you are authorized to access;
- complying with applicable laws, Google policies, and your organization's
  security and change-management requirements;
- protecting OAuth tokens, OAuth client files, service-account keys, local
  caches, drafts, exported data, and other credentials or data stored on your
  machine;
- checking the selected profile, authentication identity, quota project,
  Firebase project, proposed diff, and command arguments before confirming or
  automating a change; and
- maintaining any backups, audit records, or approval process required for
  your use of Remote Config.

Do not use the official fbrcm OAuth client or documentation website to violate
law, compromise another system or account, evade Google service restrictions,
or interfere with the operation of the project or third-party services.

## Remote Config changes

At your request, fbrcm can publish or restore Remote Config templates and can
delete selected experiments or rollouts. These operations can affect live
applications and their users. You are responsible for reviewing and authorizing
the operations you perform, including commands invoked through scripts, CI
systems, or agents.

fbrcm cannot guarantee that a change can be undone. Firebase version history,
retention, rollback behavior, and other recovery mechanisms are controlled by
Google and may change.

## Hooks and other programs

Publication hooks are optional and run local commands selected or trusted by
you. Hooks, shell pipelines, proxies, terminal software, CI systems, and other
programs used with fbrcm may read, store, modify, or transmit credentials and
Remote Config data. You are responsible for reviewing and trusting those
programs and for their behavior.

## Privacy

The [fbrcm Privacy Policy](PRIVACY.md) explains how the official application
accesses, uses, stores, and shares Google user data.

## Availability and changes

The official application, OAuth client, releases, and documentation website may
be changed, suspended, or discontinued. These terms may also be updated. The
current version will be published in this repository and on the fbrcm website,
with its revision date shown above.

## Disclaimer and limitation of liability

fbrcm is provided without warranties, as described in the MIT License. To the
maximum extent permitted by applicable law, the fbrcm developer is not liable
for claims, damages, data loss, configuration errors, service interruption,
security incidents, unexpected billing, or other losses arising from or related
to the use of fbrcm, the documentation website, Google services, or programs
configured to work with fbrcm.

## Contact

For questions about these terms, email
[yumaa.verdin@gmail.com](mailto:yumaa.verdin@gmail.com) or open an issue in the
[fbrcm issue tracker](https://github.com/yumauri/fbrcm/issues).
