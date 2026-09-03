# CLI agent quickstart

This page covers non-interactive, one-shot CLI commands for agents and scripts.
For an AI application using MCP tools, start with the [MCP guide](MCP.md).
The [CLI reference](CLI.md) lists commands and flags; the
[machine contract](cli-contract.md) defines envelopes, errors, and exit statuses.

## Use JSON output

Pass `--json` for CLI automation to receive one versioned envelope on stdout.
Human tables, colors, and interactive prompts are not meant to be parsed.
Operational logging defaults to `silent` once `--json` is set; an explicit
`FBRCM_LOG_LEVEL` overrides that default. Parse stdout only, because explicitly
enabled logs and trusted hooks may write to stderr.

Inspect the envelope's `outcome`, `exit_code`, `errors`, and `warnings`
together. Do not treat exit status alone as the result: for example, a
successful diff that found changes uses status `1`, while partial success uses
status `12`.

## 1. Discover before you act

Don't guess flags from memory or trust this document to stay current. Ask the
binary.

```
fbrcm capabilities --json
```

Returns every executable command with its ID, path, summary, schema URNs,
side-effect level, destructive marker, and support booleans. Discover every
stateless-capable command without scraping help text:

```
fbrcm capabilities --json |
  jq -r '.data.commands[] | select(.supports.stateless) | .path | join(" ")'
```

Discover commands that can create an immutable publication plan:

```
fbrcm capabilities --json |
  jq -r '.data.commands[] | select(.supports.plan) | .path | join(" ")'
```

Look up one command in detail:

```
fbrcm capabilities project import --json
```

This returns full argument/flag docs, response and error schema URNs, network
access conditions, idempotency, dry-run/draft/plan support, and interaction rules.
That is enough to construct a correct call without reading source.

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

## 4. Use drafts while a change is evolving

Direct Remote Config mutations publish after confirmation; in JSON mode they
require `--yes` when confirmation is needed. When the desired result still
needs several edits, stage and review them as a draft:

```
fbrcm update feature_enabled --project '=my-app' --type boolean --value true --draft --json
fbrcm draft diff my-app --against current --json
fbrcm draft publish my-app --yes --json
```

A target with an unpublished draft refuses direct (non-draft) Remote Config
writes, so draft state is a safety rail, not just a convenience.

## 5. Use plans for approval and execution handoff

A plan is a private, integrity-protected file containing exact base and
candidate templates. Prefer it when the agent must prepare a validated change
now, but a human or another process will authorize or execute that exact change
later. Create it directly from a supported mutation:

```
fbrcm update feature_enabled \
  --project '=my-app' \
  --type boolean \
  --value true \
  --plan-out release.fbrcm-plan.json \
  --json
```

Or compose several changes in a draft first, then save its effective publication
candidate in a plan:

```
fbrcm draft publish my-app --plan-out release.fbrcm-plan.json --json
```

Plan creation fetches fresh bases, validates every changed candidate with
Firebase, and runs effective trusted pre-publish hooks. It does not publish or
change draft state. If any target cannot be prepared, no plan file is created.

Verify the file and inspect its non-secret machine summary before requesting
authorization:

```
fbrcm plan validate release.fbrcm-plan.json --json
fbrcm plan show release.fbrcm-plan.json --json
```

`plan validate` is an offline integrity check, not a freshness check. Apply
preflights all targets and returns `plan.stale` if Firebase matches neither
the recorded base nor the candidate. It never silently rebases or replans.
For an authorized non-interactive preview and publication:

```
fbrcm apply release.fbrcm-plan.json --dry-run --yes --json
fbrcm apply release.fbrcm-plan.json --yes --json
```

The dry run rechecks current state, Firebase validation, and effective trusted
pre-publish hooks without publishing or cleaning up a source draft. The live
apply publishes the recorded candidate. Multi-target apply becomes non-atomic
after publication starts, so inspect every item status.

A stateful plan must be applied statefully with the same effective hook
configuration. A plan created with `--stateless` must be applied with
`--stateless`. Retrying does not republish targets already at the candidate,
but is not declared safe after a trusted hook executes. Create a new plan
after any stale-base failure.

Plan files contain complete Remote Config templates and metadata. Keep them out
of logs and source control, transfer them only through an approved secure
channel, and apply your normal secret-retention policy. See the
[Plans website guide](site/cli/plans.md) for the full workflow.

## 6. Confirmations never block JSON mode

`--json` never triggers an interactive prompt. If a command would normally
ask for confirmation and you didn't pass `--yes`, it returns a structured
`interaction.required` problem instead of hanging. Other required human input,
such as OAuth authorization or an unavailable file/editor choice, uses the
same structured problem. Retry with `--yes` only when the caller has explicitly
authorized the described write. Otherwise, show the interaction to a human.

## 7. Handle errors by structure, not by text

Every failure is a typed problem with a `category`, and remediation (when
safe) carries a `strategy`:

- `retry_with_arguments`: add the arguments to your original invocation
- `replace_selector`: replace the selector with the given exact value
- `run_command`: use the complete, ready-to-run fbrcm argument list

Branch on `strategy`, not on the human-readable message. When a batch
mutation partially fails, each failed item includes a target-aware
`retry_selector`, such as `=my-project`. Pass it back as `--project
<retry_selector>` to retry only what failed, rather than reprocessing the
whole batch. A remediation describes a technically valid recovery; it does not
grant permission to perform a destructive or remote action. Recheck the target
scope and capability metadata before executing it.

## 8. Batch operations are per-target, not all-or-nothing

Commands touching multiple projects (`update`, `delete`, `draft publish`,
`projects promote`, ...) process independent targets and expose their results
in `data.items` when usable data is available. Inspect the envelope `outcome`,
top-level `errors`, and every `items[].status`. Do not infer the batch result
from process status alone, and do not assume that earlier successful targets
were rolled back after a later failure.

## 9. Useful environment variables for automation

| Variable | Use |
| --- | --- |
| `FBRCM_GOOGLE_ACCESS_TOKEN` | supply the static Google OAuth access token required by supported stateless Firebase API commands; `project open` and stateless `get` stdin mode are tokenless |
| `FBRCM_PROFILE` | pin a profile for one process without switching it persistently |
| `FBRCM_CONFIG_DIR` | isolate the configuration root for a tool runner |
| `FBRCM_CACHE_DIR` | isolate the cache root for a tool runner |
| `FBRCM_OFFLINE` | force offline mode (any defined value, including empty/`0`); human `project open` prints its URL instead of launching a browser |
| `FBRCM_NO_LOCAL_CONFIG` | ignore repository `.fbrcm.toml` discovery when set to a nonempty value |
| `FBRCM_LOG_LEVEL` | override the `silent` JSON-mode default if you want logs on stderr |
| `FBRCM_HOOK_TRUST` | pin the exact `fbrcm hooks fingerprint` value in CI to trust repo hooks non-interactively |
| `GOOGLE_CLOUD_QUOTA_PROJECT` | override the Google Cloud quota/billing project for every Firebase and Cloud Resource Manager request |

Quota-project precedence is `GOOGLE_CLOUD_QUOTA_PROJECT`, the selected
project's persisted override, the auth identity's persisted default, ADC
`quota_project_id` for gcloud auth only, then the physical target project. The
caller needs `serviceusage.services.use` on the selected project. Targetless
requests fail before network access if no source resolves; fbrcm never omits
`X-Goog-User-Project`. Malformed values produce a typed
`auth.configuration_invalid` problem; malformed ADC quota metadata produces
`auth.credentials_invalid`.
Use the standard Google variable name, not an fbrcm-specific alias:

```text
GOOGLE_CLOUD_QUOTA_PROJECT=automation-quota fbrcm projects update --json
```

For persisted profile configuration, use `auth quota-project show|set|unset`
and `project quota-project show|set|unset`; do not edit profile JSON directly.

For a one-shot operation without profile files, supply a short-lived access
token. Do not hard-code a stateless command list in automation: use
`supports.stateless` from `fbrcm capabilities --json`, as shown above. Current
coverage includes parameter reads and mutations; condition and group reads,
validation, and mutations; experiment, rollout, and personalization reads;
experiment and rollout deletion; project defaults, export, import, open, and
show; project discovery, diff, and promotion; and version list, show, export,
diff, and rollback.

Commands that take one direct target require a literal Firebase project ID,
optionally qualified with `client@` or `server@` when that command supports
template selection. Parameter mutations, `get`, and `groups list` also support
live project discovery and filtering. Managed-feature commands require
an unqualified physical project ID:

```text
FBRCM_GOOGLE_ACCESS_TOKEN="$(gcloud auth application-default print-access-token)" \
  fbrcm --stateless project export my-project-id --json --to remote-config.json

FBRCM_GOOGLE_ACCESS_TOKEN="$(gcloud auth application-default print-access-token)" \
  fbrcm --stateless get --project =my-project-id --json
```

Stateless execution disables application-managed local reads, local writes,
and configured hooks. Explicit caller-selected input and output files remain
allowed, including plan files. Where present, `--update` and `--cached` are
rejected because stateless reads are already live and cannot use snapshots;
mutations that offer `--draft` also reject it. A plan created statelessly must
be applied statelessly. Omitting `--stateless` retains normal cache, draft,
profile, and hook behavior.

`fbrcm --stateless projects list --json` discovers projects live without the
profile project registry. Filters match remote project names and IDs only;
`--update` is rejected because discovery is already live. `--expr` remains
available: after ordinary project filtering, fbrcm directly fetches the current
client Remote Config template for each remaining project and evaluates the
expression without reading or writing a cache. Set `GOOGLE_CLOUD_QUOTA_PROJECT`
when the access-token identity requires a quota project for account-wide
listing.

`versions list` reads live history and rejects `--cached`. `conditions list`
fetches the latest template without applying local drafts or caches; its
filter, search, and expression options remain available, while `--update` is
rejected because the read is already live.

Remote `get` discovers all accessible projects when `--project` is omitted.
Non-exact selectors discover once and filter remote project IDs and display
names; exact `=project-id` selectors bypass discovery and fetch that physical
project directly. Repeated selectors are ORed, optional `client@` and `server@`
prefixes select a template kind, and repository aliases are never consulted.
The command preserves its parameter, search, and expression filters and
rejects `--update`. Stateless `get` can also process a stdin Remote Config
document without an access token.

`fbrcm --stateless project open my-project-id` is also available without an
access token because it only constructs and opens the Firebase Console URL.

## Minimal end-to-end example

```
# 1. discover
fbrcm capabilities update --json

# 2. read current state
fbrcm get feature_enabled --project '=my-app' --json

# 3. create an exact validated plan
fbrcm update feature_enabled --project '=my-app' --type boolean --value true \
  --plan-out release.fbrcm-plan.json --json

# 4. verify its integrity and machine summary
fbrcm plan validate release.fbrcm-plan.json --json
fbrcm plan show release.fbrcm-plan.json --json

# 5. preflight without publication after authorization
fbrcm apply release.fbrcm-plan.json --dry-run --yes --json

# 6. apply the exact reviewed candidate after authorization
fbrcm apply release.fbrcm-plan.json --yes --json
```

## Pitfalls

- Don't parse human-readable tables. They're for terminals, not agents. Use `--json` for everything.
- `--yes` authorizes the confirmation branch; it does not skip Firebase validation or make an unsafe retry idempotent.
- A valid plan is not necessarily current. Only `apply` preflight establishes that every target still matches its recorded base or candidate.
- Do not print, commit, or attach publication plans casually. Their JSON contains complete base and candidate Remote Config templates.
- Positional selectors for existing resources are exact, case-sensitive, and untrimmed. Depending on the command, a mismatch produces either a typed not-found failure or a successful no-op. Use query flags such as `--filter`, `--search`, and `--project` only where the detailed command capability publishes them.
- A `published-cache-failed` result means Firebase already accepted the write. Follow its structured remediation, normally a targeted `get --update`, instead of retrying the mutation.
