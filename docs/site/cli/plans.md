# Plans

A publication plan is a private, integrity-protected file containing the exact
Remote Config candidates that fbrcm prepared and validated. Create one when a
change must be reviewed, approved, handed off, or applied later without
recalculating it.

Examples use `example-project-id` as a physical Firebase project ID and
`release.fbrcm-plan.json` as a local plan file.

## Plans and drafts solve different problems

| | Draft | Plan |
| --- | --- | --- |
| Purpose | Compose and revise several edits | Preserve one exact publication decision |
| Storage | Active fbrcm profile | Caller-selected private file |
| Later changes | Mutations extend the draft | Create a new plan |
| Publication | Rebase local intent with `draft publish` | Require the recorded base with `apply` |
| Portability | Tied to local profile state | Can be handed off securely |
| Stateless mode | Unavailable | Supported when the producing command supports it |

Use a draft while the desired change is still evolving. Use a plan once the
candidate and target selection are ready for an approval or execution
boundary. A draft can become a plan instead of being published immediately.

## Create a plan

Commands whose capability reports `supports.plan: true` accept
`--plan-out <path>`:

```sh
fbrcm update feature_enabled \
  --project '=example-project-id' \
  --type boolean \
  --value true \
  --plan-out release.fbrcm-plan.json
```

Planning performs the normal target selection, fetches fresh base templates,
builds the candidates, validates changed candidates with Firebase, and runs
effective trusted pre-publish hooks. It does not publish Remote Config or
change drafts. If any selected target cannot be prepared or validated, fbrcm
does not create the plan.

The destination is created as a private file and is never overwritten.
`--plan-out` cannot be combined with `--dry-run`, `--draft`, or `--yes`, and
`-` is not accepted as an output path.

Supported producers include parameter, group, and condition mutations,
`project import`, `projects promote`, `draft publish`, and `versions restore`.
Discover the installed binary's exact set instead of hard-coding it:

```sh
fbrcm capabilities --json |
  jq -r '.data.commands[] | select(.supports.plan) | .path | join(" ")'
```

## Seal a draft for approval

Compose a multi-step change as a draft, then capture its effective publication
candidate in a plan:

```sh
fbrcm draft diff example-project-id --against current
fbrcm draft publish example-project-id \
  --plan-out release.fbrcm-plan.json
```

Creating the plan does not remove the source draft. A live apply removes it
only if it still matches the draft recorded by the plan. If the draft changed
after planning, fbrcm preserves it and reports a warning.

## Inspect and validate

```sh
fbrcm plan show release.fbrcm-plan.json
fbrcm plan validate release.fbrcm-plan.json
```

`plan show` verifies the file and prints its identity, execution policy,
targets, actions, validation provenance, and complete base-to-candidate diffs.
With `--json`, it returns a non-secret typed summary instead of embedding the
templates.

`plan validate` verifies the document shape, snapshot digests, invariants, and
content-derived plan ID. It is an offline integrity check; it does not check
whether Firebase still matches the recorded base.

Both commands accept `-` to read one plan document from stdin.

## Apply the exact candidates

Preview execution with a full preflight, then apply after authorization:

```sh
fbrcm apply release.fbrcm-plan.json --dry-run
fbrcm apply release.fbrcm-plan.json
```

Apply verifies the plan and execution environment, reads every changed target,
and preflights the entire plan before the first publication. A target must
still equal the recorded base, or already equal the candidate. Any other state
fails with `plan.stale`; fbrcm never silently rebases or replans.

After aggregate confirmation, fbrcm validates every changed candidate and
trusted pre-publish hook before writing. `--dry-run` performs those checks
without publication or draft cleanup. Use `--yes` for an explicitly authorized
non-interactive invocation.

Plans with several targets are non-atomic once publication starts. Successful
targets are not rolled back if a later target fails. Inspect every result item;
statuses distinguish unchanged, would-publish, published, already-applied,
conflict, and post-publication failures.

## Execution environment and retries

A stateful plan must be applied statefully with the same effective hook
configuration. A plan created with `--stateless` must be applied with
`--stateless` and a suitable access token. A plan containing only unchanged
targets can be applied entirely offline.

An already-applied candidate converges without another publication. Retrying a
plan is otherwise declared safe only when no trusted hook ran, because hook
commands may have non-idempotent effects. A stale plan must be replaced by a
newly created plan after the current state is reviewed.

## Protect plan files

A plan contains complete base and candidate Remote Config templates, ETags,
digests, change notes, and provenance. Treat it as sensitive configuration:
keep it out of logs and source control, restrict who can read it, transfer it
through an approved secure channel, and remove it according to your retention
policy after use.

In JSON mode, plan creation returns only the plan ID, counts, and artifact
metadata including the output path, byte size, and SHA-256 digest. The plan
contents are never copied into stdout.
