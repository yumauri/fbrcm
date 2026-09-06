# Firebase applications

Firebase apps are separate from Remote Config templates. They still matter to
Remote Config because a condition can target a Firebase App ID. fbrcm provides
read-only commands for finding those IDs, checking app registration details,
and downloading SDK configuration.

## List registered apps

List Android, iOS, and Web apps across every configured project:

```sh
fbrcm apps list
```

Limit the result to one project or platform:

```sh
fbrcm apps list example-project-id --platform ios
fbrcm apps list --project '=example-project-id' --filter '^checkout'
```

The app filter matches display name, namespace, or Firebase App ID. Repeat
`--filter` to combine queries with OR. Use `--show-deleted` to include apps that
Firebase has marked for permanent deletion.

## Inspect one app

```sh
fbrcm apps show com.example.checkout --project example-project-id
```

Selection checks the Firebase App ID, full resource name, namespace, and display
name in that order. These comparisons are exact and case-sensitive. A complete
Firebase App ID includes its project number, so it does not need `--project`:

```sh
fbrcm apps show '1:1234567890:android:abc123'
```

## Download SDK configuration

Write the platform configuration to stdout:

```sh
fbrcm apps config com.example.checkout --project example-project-id
```

Or save it to a private file:

```sh
fbrcm apps config com.example.checkout \
  --project example-project-id \
  --to google-services.json
```

An existing destination requires confirmation. Use `--yes` only when replacing
it is intentional. JSON mode returns a typed artifact. With `--to`, that
artifact records the destination and digest instead of embedding the content.

## Cache behavior

The app inventory, app details, and SDK configuration have separate cache
entries with a one-hour lifetime. Normal reads use fresh entries first.
`--update` forces a Firebase read.

For `apps show` and `apps config`, `--cached` requires existing cached data and
makes no Firebase request. For `apps list`, `--cached` accepts existing inventory
even when stale, but fetches and saves it when the cache is missing. Use
`FBRCM_OFFLINE=1` when the whole command must avoid network access.

Stateless mode reads Firebase directly and does not accept either cache flag.
Application data uses its own cache and never changes a Remote Config snapshot
or draft.

See the [command index](/reference/commands) for the command map and the
[complete CLI reference](https://github.com/yumauri/fbrcm/blob/main/docs/CLI.md#firebase-applications)
for every selection rule, output field, and error contract.
