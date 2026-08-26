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
fbrcm config set log.level warn
fbrcm config set theme nord --scope local
fbrcm config reset theme --scope local
fbrcm config edit
```

`config edit` validates the staged file before replacing the active one. The
editor comes from `--editor`, `FBRCM_EDITOR`, `VISUAL`, then `EDITOR`.

## Repository aliases

Commit stable project names with the repository:

```toml
[projects.aliases]
staging = "acme-staging-42"
prod = "acme-production-42"
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
| `FBRCM_HOOK_TRUST` | Trust the exact repository hook fingerprint in CI |
| `NO_COLOR` | Disable color for any non-empty value |
| `GOOGLE_CLOUD_QUOTA_PROJECT` | Select the Google Cloud quota/billing project |
| `HTTPS_PROXY`, `HTTP_PROXY`, `NO_PROXY` | Configure the HTTP transport |

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

## Publication hooks

Repository hooks can validate or gate a candidate before publication and react
after publication. Because they execute local commands, fbrcm requires an exact
trust fingerprint:

```sh
fbrcm hooks status
fbrcm hooks fingerprint
fbrcm hooks trust
```

For non-interactive use, set `FBRCM_HOOK_TRUST` to the exact fingerprint.
Stateless mode disables configured hooks.
