# Drafts

Drafts are local, profile-scoped, and independent for client and server
templates. They let several targeted commands compose into one reviewed
publication.

## List drafts

```sh
fbrcm draft list
fbrcm draft list --filter '^prod'
```

The list includes the canonical target, base version, update time, change
counts, status, and optional change note. Invalid draft envelopes remain
visible so they can be recovered or discarded.

## Create or extend a draft

Any supported mutation can use `--draft`:

```sh
fbrcm update payments_enabled \
  --project '=staging' \
  --type boolean \
  --value true \
  --draft

fbrcm groups add payments \
  --project '=staging' \
  --description 'Payment configuration' \
  --draft
```

Later mutations compose onto the same target-specific draft.

## Inspect and export

```sh
fbrcm draft show staging
fbrcm draft show staging --to candidate.json
fbrcm draft show staging --raw --to draft-envelope.json
```

Normal output is the validated working Remote Config candidate. `--raw` emits
the stored envelope including its immutable base and can recover content even
when normal draft decoding fails.

## Review changes

```sh
# Purely local: base → stored draft.
fbrcm draft diff staging --against base

# Live preview: current Firebase → rebased candidate.
fbrcm draft diff staging --against current
```

Add parameter, group, expression, or search filters to focus a large diff. A
diff command returns status 1 when the filtered result contains differences and
status 0 when it does not.

## Set the Firebase change note

```sh
fbrcm draft change-note staging 'Prepare payments launch'
fbrcm draft change-note staging --clear
```

The note is stored with the draft and sent as the published Firebase version
description. `draft publish --change-note` can override it for one invocation.

## Publish

```sh
fbrcm draft publish staging
fbrcm draft publish --all --dry-run
```

Publication fetches current state, merges local intent, displays the effective
candidate, validates with Firebase, and publishes with ETag protection.
Conflicts preserve the draft. Batch operations continue after independent
target failures and report every outcome.

## Discard

```sh
fbrcm draft discard staging
fbrcm draft discard --all
```

Discarding is a local deletion. Human mode shows the base-to-draft diff before
confirmation; `--yes` accepts the confirmation non-interactively.
