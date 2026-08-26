# Safe changes

The safest fbrcm workflow separates intent, review, and publication. Use a dry
run for a disposable preview or a draft when several edits belong together.

## Choose dry run or draft

Use `--dry-run` when you want to answer “what would this command do?” without
persisting anything:

```sh
fbrcm update checkout_enabled \
  --project '=staging' \
  --type boolean \
  --value true \
  --dry-run
```

The command prepares the candidate, validates it with Firebase, prints the
diff, and suppresses publication and local writes.

Use `--draft` when you want to collect related edits:

```sh
fbrcm update checkout_enabled \
  --project '=staging' \
  --type boolean \
  --value true \
  --draft

fbrcm update checkout_title \
  --project '=staging' \
  --type string \
  --value 'Try the new checkout' \
  --draft
```

Once a target has a draft, immediate publication commands refuse to bypass it.
Continue editing the draft or publish or discard it explicitly.

## Review local intent

Compare the draft with the immutable base it started from:

```sh
fbrcm draft diff staging --against base
```

This operation is local. It answers “what did I change?”

To preview what publication would do against the latest Firebase state:

```sh
fbrcm draft diff staging --against current
```

This fetches the current template, performs the same three-way merge used by
publication, and shows `current → candidate`. Add `--cached` only when you
explicitly want to avoid the network and accept that the snapshot may be stale.

## Add a change note

Firebase versions can carry a single-line change note:

```sh
fbrcm draft change-note staging 'Enable checkout experiment'
```

You can also use `--change-note` on a mutation or override the stored note at
publish time.

## Publish

```sh
fbrcm draft publish staging
```

For each draft, fbrcm:

1. fetches current Firebase state;
2. rebases local intent with a three-way merge;
3. stops if local and remote changes conflict;
4. shows the exact candidate diff;
5. validates the candidate with Firebase; and
6. publishes it with ETag protection.

Conflicts and failed validation preserve the draft. If Firebase already
contains the intended result, fbrcm avoids creating a redundant version and
cleans up the draft after a live publish.

## Work across projects

Batch publication is intentionally non-atomic. Each target is fetched,
reviewed, validated, and attempted independently:

```sh
fbrcm draft publish --all
```

A failure in one project does not hide the results for the others. Treat the
final report as a set of per-target outcomes and retry only failed targets.

## Discard intentionally

Discarding removes local draft state and never contacts Firebase:

```sh
fbrcm draft discard staging
```

Human mode shows the local base-to-draft diff before confirmation. Use
`fbrcm draft show staging --raw` first when you need to recover or archive a
damaged draft envelope.

## A practical default

For one obvious edit, start with `--dry-run`, inspect the diff, then rerun the
command normally. For multiple edits, use a draft from the start and publish it
only after reviewing `--against current`.
