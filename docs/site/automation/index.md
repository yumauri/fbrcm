# Automation and agents

fbrcm publishes a versioned machine interface for CI jobs, scripts, and LLM
agents. Always use `--json`; human tables, colors, diffs, and prompts are not a
stable parsing surface.

## Discover the installed binary

Do not hard-code an assumed command surface. Ask the binary what it supports:

```sh
fbrcm capabilities --json
fbrcm capabilities project import --json
fbrcm schema list --json
```

Capability records describe arguments, flags, schemas, side effects,
interaction, and support for stateless, dry-run, and draft execution.

## Read before writing

Resolve the exact target and inspect current state first:

```sh
fbrcm projects list --json
fbrcm get feature_enabled --project '=my-project-id' --json
```

Use exact selectors in automation. A fuzzy name that is convenient at a
terminal can become ambiguous as projects are added.

## Preview, then stage

When a capability reports dry-run support, validate the real candidate without
publishing it:

```sh
fbrcm update feature_enabled \
  --project '=my-project-id' \
  --type boolean \
  --value true \
  --dry-run \
  --json
```

For multi-step or human-reviewed work, prefer a draft:

```sh
fbrcm update feature_enabled \
  --project '=my-project-id' \
  --type boolean \
  --value true \
  --draft \
  --json

fbrcm draft diff my-project-id --against current --json
fbrcm draft publish my-project-id --yes --json
```

`--yes` authorizes a documented confirmation. It does not disable Firebase
validation, make a retry safe, or turn a destructive action into an authorized
one.

## Handle results structurally

Inspect all of these together:

- envelope `outcome` and `exit_code`;
- top-level `errors` and `warnings`;
- each `data.items[].status` in a batch result; and
- structured remediation `strategy` and `argv`.

Never branch on human-readable `message` wording. A successful diff with
changes uses status 1, and a batch can partially succeed with status 12.
Earlier successful targets are not rolled back after another target fails.

## Interaction never blocks JSON mode

JSON mode never opens a prompt, editor, file picker, or browser. If human input
is required, fbrcm returns an `interaction.required` problem with status 10.
Surface that request or retry only after the caller explicitly supplies the
required choice.

## Run without local state

Supported commands can operate from a short-lived access token:

```sh
FBRCM_GOOGLE_ACCESS_TOKEN="$TOKEN" \
  fbrcm --stateless get feature_enabled \
  --project '=my-project-id' \
  --json
```

Stateless mode skips fbrcm-managed profiles, project registrations, caches,
drafts, and hooks. It still permits explicit caller-selected input and output
files. Discover current coverage from `supports.stateless` in capabilities.

Continue with the [JSON contract](./json-contract) for the envelope, exit
statuses, schemas, and artifact results.
