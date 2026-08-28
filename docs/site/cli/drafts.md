# Drafts

A draft stores unpublished edits in the active profile. Client and server
templates have separate drafts. Later mutation commands add their edits to the
same target's draft.

Examples use `example-project-id` as a physical Firebase project ID.

## List drafts

```sh
fbrcm draft list
fbrcm draft list --filter '^prod'
```

The list includes the canonical target, base version, update time, change
counts, status, and optional change note. fbrcm keeps invalid draft envelopes
visible so you can recover or discard them.

## Create or extend a draft

Any supported mutation can use `--draft`:

```sh
fbrcm update payments_enabled \
  --project '=example-project-id' \
  --type boolean \
  --value true \
  --draft

fbrcm groups add payments \
  --project '=example-project-id' \
  --description 'Payment configuration' \
  --draft
```

Later mutations compose onto the same target-specific draft.

## Inspect and export

```sh
fbrcm draft show example-project-id
fbrcm draft show example-project-id --to candidate.json
fbrcm draft show example-project-id --raw --to draft-envelope.json
```

Normal output is the validated working Remote Config candidate. `--raw` emits
the stored envelope including its immutable base and can recover content even
when normal draft decoding fails.

## Review changes

```sh
# Purely local: base → stored draft.
fbrcm draft diff example-project-id --against base

# Live preview: current Firebase → rebased candidate.
fbrcm draft diff example-project-id --against current
```

Add parameter, group, expression, or search filters to focus a large diff. A
diff command returns status 1 when the filtered result contains differences and
status 0 when it does not.

## Set the Firebase change note

```sh
fbrcm draft change-note example-project-id 'Prepare payments launch'
fbrcm draft change-note example-project-id --clear
```

fbrcm stores the note with the draft and sends it as the published Firebase
version description. `draft publish --change-note` can override it for one
invocation.

## Publish

```sh
fbrcm draft publish example-project-id
fbrcm draft publish --all --dry-run
```

Publication fetches current state, merges local intent, displays the effective
candidate, validates with Firebase, and publishes with ETag protection.
Conflicts preserve the draft. Batch operations continue after independent
target failures and report every outcome.

## Discard

```sh
fbrcm draft discard example-project-id
fbrcm draft discard --all
```

Discarding is a local deletion. Human mode shows the base-to-draft diff before
confirmation; `--yes` accepts the confirmation non-interactively.
