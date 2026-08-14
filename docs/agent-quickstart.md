# Agent quickstart

This page is for LLM agents and scripts driving `fbrcm` non-interactively. It
is a fast path, not a replacement for the full references: see
[CLI reference](CLI.md) for every command and flag, and the
[machine contract](cli-contract.md) for the complete envelope schema, error
catalog, and exit-status table.

## The golden rule

Always pass `--json`. Every command accepts it and returns one versioned
envelope on stdout. Without it you get human tables, colors, and interactive
prompts that are not meant to be parsed. Operational logging defaults to
`silent` once `--json` is set; an explicit `FBRCM_LOG_LEVEL` overrides that
default. Parse stdout only, because explicitly enabled logs and trusted hooks
may write to stderr.

Inspect the envelope's `outcome`, `exit_code`, `errors`, and `warnings`
together. Do not treat exit status alone as the result: for example, a
successful diff that found changes uses status `1`, while partial success uses
status `12`.

## 1. Discover before you act

Don't guess flags from memory or from this doc going stale — ask the binary.

```
fbrcm capabilities --json
```

Returns every executable command with its ID, path, summary, schema URNs,
side-effect level, and destructive marker. Look up one command in detail:

```
fbrcm capabilities project import --json
```

This returns full argument/flag docs, response and error schema URNs, network
access conditions, idempotency, dry-run/draft support, and interaction rules —
enough to construct a correct call without reading source.

If you need to validate a response shape programmatically:

```
fbrcm schema list --json
fbrcm schema show <schema-id> --json
```

## 2. Read before you write

Start with a read and inspect its detailed capability record first. A command
that reads Remote Config may still access Firebase, refresh authentication,
update a cache, synchronize the project registry, or bootstrap profile state;
the capability predicates declare these effects.

```
fbrcm projects list --json
fbrcm get feature_enabled --project '^prod' --json
fbrcm conditions list my-project --json
fbrcm versions list my-project --json
```

Filter flags (`--project`, `--filter`) use mode prefixes: `~fuzzy`, `^starts`,
`/includes`, `=exact`. Use `=exact` in scripts to avoid ambiguous matches.

## 3. Preview before you mutate

For a mutation whose detailed capability reports `supports.dry_run: true`, use
`--dry-run` before authorizing publication:

```
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --dry-run --json
```

For a changed live Remote Config candidate, this performs Firebase's real
validation-only request and suppresses publication. Check `validated` and
`validation_source` in the result. A later live call can still encounter a
concurrent update or another transient failure. Trusted pre-publish hooks may
run during dry-run, so also inspect the declared side effects and warnings.

## 4. Prefer drafts for anything multi-step or reviewable

Direct Remote Config mutations publish after confirmation; in JSON mode they
require `--yes` when confirmation is needed. For anything an agent should not
publish immediately, stage it instead:

```
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --draft --json
fbrcm draft diff my-app --against current --json
fbrcm draft publish my-app --yes --json
```

A target with an unpublished draft refuses direct (non-draft) Remote Config
writes, so draft state is a safety rail, not just a convenience.

## 5. Confirmations never block JSON mode

`--json` never triggers an interactive prompt. If a command would normally
ask for confirmation and you didn't pass `--yes`, it returns a structured
`interaction.required` problem instead of hanging. Other required human input,
such as OAuth authorization or an unavailable file/editor choice, uses the
same structured problem. Retry with `--yes` only when the caller has explicitly
authorized the described write; otherwise surface the interaction to a human.

## 6. Handle errors by structure, not by text

Every failure is a typed problem with a `category`, and remediation (when
safe) carries a `strategy`:

- `retry_with_arguments` — augment your original invocation
- `replace_selector` — substitute the given exact selector
- `run_command` — a complete, ready-to-run fbrcm argv

Branch on `strategy`, not on the human-readable message. When a batch
mutation partially fails, each failed item includes a target-aware
`retry_selector` (e.g. `=my-project`) — pass it back as `--project
<retry_selector>` to retry only what failed, rather than reprocessing the
whole batch. A remediation describes a technically valid recovery; it does not
grant permission to perform a destructive or remote action. Recheck the target
scope and capability metadata before executing it.

## 7. Batch operations are per-target, not all-or-nothing

Commands touching multiple projects (`update`, `delete`, `draft publish`,
`projects promote`, ...) process independent targets and expose their results
in `data.items` when usable data is available. Inspect the envelope `outcome`,
top-level `errors`, and every `items[].status`. Do not infer the batch result
from process status alone, and do not assume that earlier successful targets
were rolled back after a later failure.

## 8. Useful environment variables for automation

| Variable | Use |
| --- | --- |
| `FBRCM_PROFILE` | pin a profile for one process without switching it persistently |
| `FBRCM_CONFIG_DIR` | isolate the configuration root for a tool runner |
| `FBRCM_CACHE_DIR` | isolate the cache root for a tool runner |
| `FBRCM_OFFLINE` | force offline mode (any defined value, including empty/`0`) |
| `FBRCM_NO_LOCAL_CONFIG` | ignore repository `.fbrcm.toml` discovery when set to a nonempty value |
| `FBRCM_LOG_LEVEL` | override the `silent` JSON-mode default if you want logs on stderr |
| `FBRCM_HOOK_TRUST` | pin the exact `fbrcm hooks fingerprint` value in CI to trust repo hooks non-interactively |
| `GOOGLE_CLOUD_QUOTA_PROJECT` | set one Google Cloud quota/billing project for Firebase and Cloud Resource Manager requests across gcloud, OAuth, and service-account identities |

`GOOGLE_CLOUD_QUOTA_PROJECT` overrides an ADC `quota_project_id` and any
gcloud target-project fallback. The caller needs `serviceusage.services.use`
on that project. Malformed values produce a typed `auth.configuration_invalid`
problem; malformed ADC quota metadata produces `auth.credentials_invalid`.
Use the standard Google variable name, not an fbrcm-specific alias:

```text
GOOGLE_CLOUD_QUOTA_PROJECT=automation-quota fbrcm projects update --json
```

## Minimal end-to-end example

```
# 1. discover
fbrcm capabilities update --json

# 2. read current state
fbrcm get feature_enabled --project '=my-app' --json

# 3. preview
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --dry-run --json

# 4. stage as draft
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --draft --json

# 5. review the effective publish diff
fbrcm draft diff my-app --against current --json

# 6. publish
fbrcm draft publish my-app --yes --json
```

## Pitfalls

- Don't parse human-readable tables — they're for terminals, not agents; use `--json` for everything.
- `--yes` authorizes the confirmation branch; it does not skip Firebase validation or make an unsafe retry idempotent.
- Positional selectors for existing resources are exact, case-sensitive, and untrimmed. Depending on the command, a mismatch produces either a typed not-found failure or a successful no-op. Use query flags such as `--filter`, `--search`, and `--project` only where the detailed command capability publishes them.
- A `published-cache-failed` result means Firebase already accepted the write. Follow its structured remediation—normally a targeted `get --update`—instead of retrying the mutation.
