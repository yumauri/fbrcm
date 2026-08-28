# Configuration

fbrcm combines built-in defaults, global user configuration, and an optional
repository configuration file.

## Configuration layers

Global settings live in `config.toml` under the operating system's user config
directory. At startup, fbrcm searches the current directory and its parents for
the nearest `.fbrcm.toml`. Nested tables merge; repository scalars and arrays
replace the global value.

Inspect the effective result and file locations:

```sh
fbrcm config show
fbrcm config path
fbrcm config path --scope local
fbrcm config validate
```

Use `--no-local-config` or `FBRCM_NO_LOCAL_CONFIG=1` to ignore `.fbrcm.toml` for
one process.

## Change settings

```sh
fbrcm config set network.max_concurrent_requests 3
fbrcm config set theme nord --scope local
fbrcm config reset theme --scope local
fbrcm config edit
```

Set the logging level with `FBRCM_LOG_LEVEL`. It is not a persisted `config set`
key.

`config edit` validates the staged file before replacing the active one. The
editor comes from `--editor`, `FBRCM_EDITOR`, `VISUAL`, then `EDITOR`.

## Repository aliases

Commit stable project names with the repository:

```toml
[projects.aliases]
staging = "example-staging-project-id"
prod = "example-production-project-id"
```

The aliases are profile-independent. fbrcm also reads Firebase CLI aliases
from the `.firebaserc` associated with the nearest Firebase project root.

## Useful environment variables

| Variable | Purpose |
| --- | --- |
| `FBRCM_PROFILE` | Select an existing profile for one process |
| `FBRCM_CONFIG_DIR` | Override the fbrcm config root |
| `FBRCM_CACHE_DIR` | Override the fbrcm cache root |
| `FBRCM_OFFLINE` | Disable network use whenever the variable is defined |
| `FBRCM_LOG_LEVEL` | Set `debug`, `info`, `warn`, `error`, `fatal`, or `silent` |
| `FBRCM_EDITOR` | Select an editor command and arguments |
| `FBRCM_NO_LOCAL_CONFIG` | Ignore repository configuration |
| `FBRCM_HOOK_TRUST` | Trust the exact [repository hook](./hooks) fingerprint in CI |
| `NO_COLOR` | Disable color for any non-empty value |
| `GOOGLE_CLOUD_QUOTA_PROJECT` | Select the Google Cloud quota/billing project |
| `HTTPS_PROXY`, `HTTP_PROXY`, `NO_PROXY` | Configure the HTTP transport |

## Quota and billing project

Every authenticated Firebase and Cloud Resource Manager request carries
`X-Goog-User-Project`. fbrcm resolves it in this order:

1. `GOOGLE_CLOUD_QUOTA_PROJECT`
2. the selected project's persisted override
3. the selected auth identity's persisted default
4. ADC `quota_project_id`, for gcloud auth only
5. the physical target Firebase project

If a targetless request cannot resolve a value, fbrcm stops before network
access. The syntax for persisted values is:

```text
fbrcm auth quota-project show <auth-id>
fbrcm auth quota-project set <auth-id> <quota-project-id>
fbrcm auth quota-project unset <auth-id>
fbrcm project quota-project set <project> <quota-project-id>
fbrcm project quota-project unset <project>
```

For example:

```sh
fbrcm auth quota-project show example-auth-name
fbrcm auth quota-project set example-auth-name example-quota-project-id
fbrcm auth quota-project unset example-auth-name
fbrcm project quota-project set example-project-id example-quota-project-id
fbrcm project quota-project unset example-project-id
```

`example-auth-name` is a local auth ID. The other values are Google Cloud
project IDs. See [Authentication and project discovery](/guide/authentication)
for the first-run setup and the distinction between these identifiers.

The authenticated principal needs `serviceusage.services.use` on the effective
quota project. The guided first-run setup also asks for an auth-level quota
project.

## Network controls

Authenticated requests share concurrency, pacing, rate-limit cooldown, and
retry policies. The main keys are:

```text
network.max_concurrent_requests
network.requests_per_minute
network.rate_limit_cooldown
network.retry.max_attempts
network.retry.base_delay
network.retry.max_delay
network.retry.jitter_percent
```

Defaults allow five concurrent requests, disable proactive per-minute pacing,
and make up to five replayable attempts with bounded exponential delay. A
Firebase `Retry-After` header takes precedence over calculated waits.
