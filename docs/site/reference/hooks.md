# Publication hooks

Hooks run local commands before or after fbrcm publishes Remote Config. A
`pre_publish` hook can reject a candidate. A `post_publish` hook can notify
another system after Firebase accepts the publication. fbrcm does not run hooks
unless you configure them.

## Configure hooks

Add hooks to the global `config.toml` or a repository `.fbrcm.toml`. Use
`fbrcm config edit` for the global file or this command for the repository file:

```sh
fbrcm config edit --scope local
```

```toml
[hooks]
timeout = "2m"
pre_publish = [
  "./scripts/validate-parameter-names.sh",
  "go run ./tools/validate-remote-config",
]
post_publish = ["./scripts/notify-remote-config-published.sh"]
```

Commands run in order through `/bin/sh -c` on Unix-like systems and
`cmd.exe /S /C` on Windows. The timeout applies separately to each command and
defaults to five minutes. Repository hook arrays replace the corresponding
global arrays instead of extending them.

Run `fbrcm config validate` after editing the file. Empty commands, invalid
durations, and non-positive timeouts make the configuration invalid.

## When hooks run

| Hook | Runs | Failure result |
| --- | --- | --- |
| `pre_publish` | After Firebase validates the final candidate and before publication | Stops publication |
| `post_publish` | After Firebase accepts a real publication | Reports failure, but cannot undo publication |

`pre_publish` also runs during a publication dry run. This lets CI check the
same policy without writing to Firebase. `post_publish` does not run for dry
runs or failed publications.

Saving or previewing a draft does not run hooks. Publishing a draft does.
Direct changes, imports, promotions, restores, rollbacks, and their TUI
equivalents use the same lifecycle. No-op operations do not run hooks.

For commands that select several targets, fbrcm runs each target's hooks in
order. A failure does not prevent it from processing independent targets.

## Files available to a hook

For each publication, fbrcm creates a private temporary directory containing
JSON files. It passes each file path to the hook in an environment variable:

| Environment variable | Value |
| --- | --- |
| `FBRCM_CURRENT_FILE` | Path to `current.json`, containing Remote Config before the change |
| `FBRCM_CANDIDATE_FILE` | Path to `candidate.json`, containing the validated candidate |
| `FBRCM_CONTEXT_FILE` | Path to `context.json`, containing hook and publication metadata |
| `FBRCM_PUBLISHED_FILE` | Path to `published.json` after publication; empty before publication |

For example, a Unix shell hook can read the candidate with:

```sh
jq . "$FBRCM_CANDIDATE_FILE"
```

The template files are read-only. Changing them does not change the candidate
sent to Firebase. fbrcm removes the temporary directory after the operation.

fbrcm also adds these metadata environment variables to the hook process:

```text
FBRCM_HOOK_EVENT       pre_publish or post_publish
FBRCM_OPERATION        originating operation
FBRCM_TARGET           canonical client or server template target
FBRCM_PROJECT_ID       Firebase project ID
FBRCM_TEMPLATE_KIND    client or server
FBRCM_PROFILE          selected profile
FBRCM_DRY_RUN          true or false
FBRCM_CHANGE_NOTE      publication change note, possibly empty
FBRCM_CONFIG_FILE      config file that supplied the hook
FBRCM_PROJECT_DIR      directory containing that config file
```

`GCLOUD_PROJECT` contains the same value as `FBRCM_PROJECT_ID`, and
`PROJECT_DIR` contains the same value as `FBRCM_PROJECT_DIR`. The command runs
with the directory containing its source configuration file as the working
directory.

## Trust repository hooks

Global hooks are trusted automatically because they come from your user
configuration. Repository hooks execute code from the checked-out project, so
fbrcm requires explicit trust before running them.

Review the effective commands first:

```sh
fbrcm hooks status
fbrcm hooks fingerprint
```

Then trust the current repository definition:

```sh
fbrcm hooks trust
```

fbrcm stores trust outside the repository. The record includes the canonical
configuration path and a SHA-256 fingerprint of the effective commands and
timeout. Changing the path, commands, or timeout invalidates trust.

Remove stored trust with:

```sh
fbrcm hooks untrust
```

For CI, set `FBRCM_HOOK_TRUST` to the exact value printed by
`fbrcm hooks fingerprint`. A missing or mismatched value stops the operation
before any Firebase write.

## Failures and output

A nonzero exit, timeout, signal, or shell error fails the current hook stage.
CLI mode writes hook output to stderr, which keeps JSON stdout valid. The TUI
writes it to the Logs panel.

A `pre_publish` failure means Firebase was not changed. A `post_publish` failure
means Firebase accepted the publication before the hook failed. Inspect the
hook and the published version before retrying, because another publication may
create another Remote Config version.

JSON mode reports hook failures with status 14 and structured error details.
After a successful Firebase write, a post-hook failure appears as a
post-publication warning or a result such as `published-hook-failed`.

## Stateless mode

`--stateless` disables configured hooks. Stateless commands do not read global
or repository hook configuration and do not use stored hook trust.
