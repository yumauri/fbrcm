# Automation and agents

Add `--json` to calls made by CI jobs, scripts, and LLM agents. The command then
returns one versioned JSON envelope instead of terminal tables, colors, diffs,
or prompts.

The stateful examples assume you have already configured a profile through the
[CLI-only setup path](/guide/#option-2-setup-using-only-the-cli).

## Discover the installed binary

Do not hard-code an assumed command set. Ask the binary what it supports:

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
fbrcm get feature_enabled --project '=example-project-id' --json
```

Use exact selectors in automation. A fuzzy name that is convenient at a
terminal can become ambiguous as projects are added.

## Preview, then stage

When a capability reports dry-run support, validate the real candidate without
publishing it:

```sh
fbrcm update feature_enabled \
  --project '=example-project-id' \
  --type boolean \
  --value true \
  --dry-run \
  --json
```

For multi-step or human-reviewed work, prefer a draft:

```sh
fbrcm update feature_enabled \
  --project '=example-project-id' \
  --type boolean \
  --value true \
  --draft \
  --json

fbrcm draft diff example-project-id --against current --json
fbrcm draft publish example-project-id --yes --json
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

JSON mode never opens a prompt, editor, file picker, or browser. If an operation
needs human input, fbrcm returns an `interaction.required` problem with status 10.
Show that request to a human, or retry only after the caller explicitly
supplies the required choice.

## Run without local state

Supported commands can operate from a short-lived access token:

```sh
GOOGLE_CLOUD_QUOTA_PROJECT=example-quota-project-id \
FBRCM_GOOGLE_ACCESS_TOKEN="$(gcloud auth application-default print-access-token)" \
  fbrcm --stateless get feature_enabled \
  --project '=example-project-id' \
  --json
```

Stateless mode skips fbrcm-managed profiles, project registrations, caches,
drafts, and hooks. It still permits explicit caller-selected input and output
files. fbrcm keeps the token in memory and never writes it. The token's
principal needs `serviceusage.services.use` on `example-quota-project-id` and Remote Config
access on `example-project-id`. Check `supports.stateless` in the capability record
before running a command this way.

Continue with the [JSON contract](./json-contract) for the envelope, exit
statuses, schemas, and artifact results.
