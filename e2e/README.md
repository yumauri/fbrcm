# CLI end-to-end tests

This directory is a separate Go module. The repository's normal `go test ./...`
does not discover or run these tests; run them explicitly from `e2e/`.

The harness runs a real `fbrcm` process through an HTTPS Hoverfly proxy. Replay
mode never contacts Firebase: unmatched requests fail at the proxy. The command's
stdout, stderr, exit status, HTTP method, destination, request count, and replay
mode are all checked. Each scenario declares the expected HTTP method, host,
path, and status. Order is checked by default; `"http_unordered": true` compares
the declared and observed requests as a multiset for intentionally concurrent
commands. Committed HTTP simulations preserve response bodies
exactly and are sanitized before they are written.

## Requirements

- Go 1.27.0 or newer, as declared by `e2e/go.mod`.
- Network access the first time Go downloads the E2E module dependencies.
- No manual Hoverfly installation is required. Hoverfly is pinned as a Go tool
  dependency and the harness builds it automatically when `hoverfly` is not on
  `PATH`. Set `FBRCM_E2E_HOVERFLY` to use a specific existing executable instead.
- Replay mode needs no Firebase credentials or network access to Firebase.
- Live capture modes require `FBRCM_E2E_ACCESS_TOKEN` and access to the dedicated
  Firebase project configured in `testdata/suite.json`.

## Quickstart

Run all commands below from the E2E module:

```sh
cd e2e
```

### Replay existing snapshots

```sh
go test -v -count=1 .
```

This is the default mode. It does **not** create or update files. It runs `fbrcm`,
serves the committed HTTP responses without contacting Firebase, and compares the
result with the committed snapshots. `PASS` means everything matched exactly.

Snapshots live beside each scenario:

```text
testdata/scenarios/<scenario>/scenario.json   command and expectations
testdata/scenarios/<scenario>/http.json       recorded HTTP exchange (networked scenarios only)
testdata/scenarios/<scenario>/stdout.golden   exact expected stdout
testdata/scenarios/<scenario>/stderr.golden   exact expected stderr
testdata/scenarios/<scenario>/files/...golden exact bytes of declared output files
testdata/scenarios/<scenario>/state/...golden exact bytes of declared config/cache files
```

Commands run with debug logging by default, so `stderr.golden` captures detailed
diagnostics. The harness disables log timestamps so those snapshots remain
byte-for-byte stable. Interactive progress animations are not included because
E2E commands run with redirected, non-terminal stdout and stderr.

Before comparing output, the harness canonicalizes only volatile values that it
owns. Snapshots use these placeholders:

| Placeholder | Runtime value |
| --- | --- |
| `<E2E_RUN_ROOT>` | Per-scenario temporary config, cache, home, and work root |
| `<E2E_TOOLS_ROOT>` | Temporary directory containing harness-built executables and CA files |
| `<E2E_PROXY_URL>` | Hoverfly proxy URL, including its random loopback port |
| `<E2E_PROXY_ADDRESS>` | Hoverfly loopback host and random port without the URL scheme |
| `<E2E_HOOK_FINGERPRINT>` | Hook fingerprint when its hash includes the temporary repository path |

All remaining stdout and stderr bytes are compared exactly. The unmodified raw
output is still shown when command execution or HTTP validation fails.

Use `-count=1` when you need the command to run again instead of accepting Go's
cached test result.

Replay only one scenario:

```sh
go test -v -count=1 . -run '^TestCLI$/^versions_list$'
```

### Create snapshots for a new scenario

First create `testdata/scenarios/<scenario>/scenario.json`. For a networked
scenario, set the real dedicated Firebase project in `testdata/suite.json`, then
run:

```sh
FBRCM_E2E_ACCESS_TOKEN=... go test -v -count=1 . \
  -run '^TestCLI$/^my_scenario$' \
  -args -mode=record-missing
```

`record-missing` contacts the declared upstream hosts only when `http.json` is
absent. It then creates the missing HTTP, stdout, stderr, artifact-file, and
state-file snapshots declared by that scenario. It refuses to capture HTTP
traffic without `FBRCM_E2E_ACCESS_TOKEN`. A scenario with `"expected_http": []`
needs no token and creates only stdout/stderr snapshots.

Review the new files and `git diff` before committing them. Do not use a production
project for capture, even for commands expected to be read-only.

#### Use an Application Default Credentials token

You can obtain `FBRCM_E2E_ACCESS_TOKEN` from gcloud Application Default
Credentials. Configure ADC first if necessary:

```sh
gcloud auth application-default login
```

Then generate a short-lived token and run the capture in one command:

```sh
FBRCM_E2E_ACCESS_TOKEN="$(gcloud auth application-default print-access-token)" \
go test -v -count=1 . \
  -run '^TestCLI$/^versions_list$' \
  -args -mode=refresh-all
```

The ADC identity must have read access to the Firebase project. The
[Firebase Remote Config Viewer role](https://docs.cloud.google.com/iam/docs/roles-permissions/cloudconfig)
(`roles/cloudconfig.viewer`) is sufficient for the initial read-only scenarios.
Run the test promptly because the
[access token](https://docs.cloud.google.com/sdk/gcloud/reference/auth/application-default/print-access-token)
normally expires after one hour.

If ADC requires a quota project, configure it alongside the Firebase target in
`testdata/suite.json`:

```json
{
  "project_id": "firebase-test-project",
  "project_name": "fbrcm E2E",
  "quota_project_id": "google-cloud-quota-project",
  "default_terminal_width": 200,
  "default_log_level": "debug"
}
```

The harness maps `quota_project_id` to `GOOGLE_CLOUD_QUOTA_PROJECT`, causing fbrcm
to send `X-Goog-User-Project`. The field is optional and deliberately does not
default to `project_id`, because the Firebase target and quota project can differ.
The ADC identity also needs `serviceusage.services.use` on the quota project.

### Update existing snapshots

Update stdout, stderr, and declared file snapshots while replaying the existing
HTTP snapshot:

```sh
go test -v -count=1 . -run '^TestCLI$/^versions_list$' \
  -args -mode=update-output
```

Re-record HTTP while keeping stdout and stderr as comparisons:

```sh
FBRCM_E2E_ACCESS_TOKEN=... go test -v -count=1 . \
  -run '^TestCLI$/^versions_list$' \
  -args -mode=refresh-http
```

Re-record HTTP and update stdout and stderr together:

```sh
FBRCM_E2E_ACCESS_TOKEN=... go test -v -count=1 . \
  -run '^TestCLI$/^versions_list$' \
  -args -mode=refresh-all
```

Remove `-run ...` from an update command to apply that mode to every scenario.
Scenarios with `"http_replay_only": true` keep their committed synthetic HTTP
fixture in every mode; `refresh-all` still refreshes their stdout, stderr, and
declared file snapshots by replaying that fixture. This prevents account-wide or
purpose-built synthetic responses from being replaced with live account data.

### Record a mutating lifecycle

Keep each mutating command as an independent scenario with its own HTTP and
output snapshots. Declare their live-recording order in `testdata/suite.json`:

```json
{
  "project_id": "firebase-test-project",
  "recording_sequences": [
    {
      "name": "parameter-lifecycle",
      "scenarios": [
        "parameter_add_live_json",
        "parameter_update_live_json",
        "parameter_delete_live_json"
      ]
    }
  ]
}
```

Then run HTTP recording as usual against the dedicated Firebase project:

```sh
FBRCM_E2E_ACCESS_TOKEN=... go test -v -count=1 . \
  -run '^TestCLI$' \
  -args -mode=refresh-all
```

In `record-missing`, `refresh-http`, and `refresh-all`, the harness automatically
runs declared sequences first, in suite order. Every sequence is a contiguous
block whose scenarios run in their listed order. Each scenario still gets
isolated local config, cache, home, and work directories, while its live Firebase
mutations establish the remote preconditions for the next scenario. If a member
fails, Go continues with the remaining subtests, allowing a final delete scenario
to clean up the dedicated resource.

The order is mandatory rather than selected by a flag. A capture-mode `-run`
filter may select every member of a recording sequence or none of them; selecting
only part of a sequence fails before any CLI process runs. `record-missing` also
fails when only some members already have HTTP cassettes, preventing replayed
prerequisites from being mixed with live mutations. Use `refresh-all` to replace
the complete sequence in that case.

Replay and `update-output` retain normal alphabetical order, and every scenario
and cassette remains independently runnable. A sequence cannot list the same
scenario more than once; unknown or duplicate entries fail during suite loading.
Any scenario declaring mutating live HTTP traffic must belong to a recording
sequence; validate-only requests and IAM permission checks remain independent.
Use a resource name reserved for E2E and place its delete scenario last.

## Choose the fbrcm executable

By default, the harness builds the current checkout. It can instead execute an
installed binary or use `go run`:

```sh
go test -v . -args -binary=/absolute/path/to/fbrcm
go test -v . -args -go-run=..
```

The harness generates a temporary CA for each suite run. For a historical binary
that does not honor `SSL_CERT_FILE` on its platform, provide a CA certificate and
key that the binary already trusts:

```sh
go test -v . -args -binary=/absolute/path/to/fbrcm-old \
  -ca-cert=/absolute/path/to/cert.pem -ca-key=/absolute/path/to/key.pem
```

## Snapshot modes

The default `replay` mode is read-only and requires every HTTP and output snapshot
to exist.

| Mode | HTTP simulation | stdout/stderr/files/state |
| --- | --- | --- |
| `replay` | Replay existing | Compare existing |
| `record-missing` | Capture only when missing | Create only when missing |
| `refresh-http` | Capture and replace | Compare existing |
| `update-output` | Replay existing | Create or replace |
| `refresh-all` | Capture and replace | Create or replace |

Live capture requires `FBRCM_E2E_ACCESS_TOKEN` and a dedicated test project
configured in `testdata/suite.json`. Grant the capture identity only the
permissions required by the selected scenarios. A middleware guard derives its
method, host, path, and optional exact-query allowlist from each scenario's
`expected_http` entries; anything else is blocked before reaching an upstream
service. This supports multi-host read flows and explicitly declared
methods—including mutations in a recording sequence—without broadly allowing
access to other endpoints. Captured fixtures are checked for supplied tokens and
common credential fields before being saved.

The harness writes the value supplied through `FBRCM_E2E_ACCESS_TOKEN` to the
temporary OAuth profile used by stateful scenarios. Replay mode substitutes its
fixed non-secret replay token when no live token was supplied. Scenario `envs`
may map the exact value `${FBRCM_E2E_ACCESS_TOKEN}` to an application variable;
the harness replaces it with that effective live-or-replay token. Unlisted
variables are not inferred from argv, and inherited `FBRCM_GOOGLE_ACCESS_TOKEN`
is removed. The private `FBRCM_E2E_ACCESS_TOKEN` variable itself is never
forwarded to fbrcm.

Use `"http_replay_only": true` for a scenario whose HTTP response is
intentionally synthetic or whose live upstream operation observes account-wide
resources. Its `http.json` must already exist and is never captured or replaced
automatically.

## Add a scenario

Create `testdata/scenarios/<name>/scenario.json` with the command capability ID,
exact argv, expected exit code, and ordered HTTP expectations. `${PROJECT_ID}`
expands in argv and HTTP paths from `testdata/suite.json`. An expectation may set
`query` to require the exact sorted query string. The proxy guard checks method,
host, and path for every request, plus this query when declared; write-shaped
read operations such as Firebase validation must always declare their
non-mutating query. Then run that subtest with `-mode=record-missing` and, when it
declares HTTP traffic, a live access token. Review `http.json`, `stdout.golden`,
`stderr.golden`, and any `files/` snapshots before committing them.

Suite-wide output defaults are stored in `testdata/suite.json`:

```json
{
  "default_terminal_width": 200,
  "default_log_level": "debug"
}
```

Both fields are optional; omitted fields retain those same built-in defaults for
compatibility. Set `terminal_width` or `log_level` on an individual scenario to
override the suite value and exercise different output:

```json
{
  "name": "versions_list_narrow",
  "command_id": "versions.list",
  "args": ["versions", "list", "${PROJECT_ID}", "--limit", "3"],
  "expected_exit_code": 0,
  "expected_http": [
    {
      "method": "GET",
      "host": "firebaseremoteconfig.googleapis.com",
      "path": "/v1/projects/${PROJECT_ID}/remoteConfig:listVersions",
      "status": 200
    }
  ],
  "json_output": false,
  "terminal_width": 72,
  "log_level": "info"
}
```

Supported scenario log levels are `debug`, `info`, `warn`, `error`, `fatal`, and
`silent`. Log timestamps remain disabled at every level. Set `"offline": true`
to run a scenario with `FBRCM_OFFLINE=1`; this is useful for deterministic stale
cache and fallback coverage and still enforces a zero-request HTTP contract.
Set `json_output` when stdout uses the versioned CLI JSON envelope. The harness
then validates stdout against
`schemas/cli/1.0.0/<command>.response.schema.json` before snapshot comparison.
Use `envs` for explicit command environment setup. An exact
`${FBRCM_E2E_ACCESS_TOKEN}` value selects the harness's effective live-or-replay
token; any other value is passed literally. For example:

```json
{
  "envs": {
    "FBRCM_GOOGLE_ACCESS_TOKEN": "${FBRCM_E2E_ACCESS_TOKEN}"
  }
}
```

Omit the entry to test a missing variable, or use a literal such as
`"incorrect"` to test malformed credentials. The harness rejects overrides of
its private token, isolated state roots, proxy configuration, and trusted CA.

For a command that must not use the network, declare an empty contract:

```json
{
  "name": "auth_list_json",
  "command_id": "auth.list",
  "args": ["auth", "list", "--json"],
  "expected_exit_code": 0,
  "expected_http": [],
  "json_output": true
}
```

It still runs through an offline Hoverfly instance with an empty simulation.
Any accidental request therefore fails and is reported from the proxy journal.

To verify a command's file output byte-for-byte, list its working-directory
relative paths in `expected_files`:

```json
{
  "name": "versions_export_file_json",
  "command_id": "versions.export",
  "args": [
    "versions", "export", "${PROJECT_ID}", "11",
    "--to", "version-11.json", "--json"
  ],
  "expected_exit_code": 0,
  "expected_http": [
    {
      "method": "GET",
      "host": "firebaseremoteconfig.googleapis.com",
      "path": "/v1/projects/${PROJECT_ID}/remoteConfig",
      "query": "versionNumber=11",
      "status": 200
    }
  ],
  "expected_files": ["version-11.json"],
  "json_output": true
}
```

The harness rejects absolute paths and `..` traversal, requires each result to
be a regular file, and stores its exact bytes as
`files/version-11.json.golden`. File snapshots follow the same output-update
modes as stdout and stderr.

To verify persisted harness state, use `expected_state_files`. Its `root` is
restricted to the isolated `config` or `cache` directory, and `path` must be
relative:

```json
{
  "expected_state_files": [
    {
      "root": "config",
      "path": "default/projects-config.json",
      "json_replacements": {
        "/synced_at": "<E2E_SYNCED_AT>"
      }
    }
  ]
}
```

State snapshots are stored under `state/<root>/<path>.golden`. They preserve
the file bytes exactly except for runtime path replacements and explicitly
declared JSON Pointer replacements. Each declared pointer must exist; this keeps
volatile fields visible and prevents broad, accidental normalization. Declared
values are canonicalized in stdout and stderr too, so logs containing the same
generated value remain stable.

To verify deletion, use `expected_absent_state_paths`. It accepts the same
isolated `config` and `cache` roots, but the path may name either a file or a
directory:

```json
{
  "expected_absent_state_paths": [
    {"root": "cache", "path": "default/remote-config"},
    {"root": "config", "path": "default/auth/old/client-secret.json"}
  ]
}
```

The scenario fails if any filesystem object still exists at a declared path. A
path cannot be declared as both an expected state file and an expected absent
path.

### Scenario state fixtures

The harness always creates deterministic base auth and project state. A scenario
may additionally set `"fixture": "fixture-name"`. Set `"local_config": true`
when the command must discover repository files such as `.fbrcm.toml` or
`.firebaserc`; local configuration remains disabled by default to prevent state
outside the fixture from leaking into a test. The corresponding directory under
`testdata/state/fixture-name/` can contain any of these overlay roots:

```text
config/   merged into FBRCM_CONFIG_DIR
cache/    merged into FBRCM_CACHE_DIR
home/     merged into HOME
work/     merged into the scenario working directory
```

Files replace base fixture files at the same relative path. Symbolic links are
rejected. Fixture contents may use `<E2E_CONFIG_DIR>`, `<E2E_CACHE_DIR>`,
`<E2E_HOME_DIR>`, and `<E2E_WORK_DIR>` when the staged file must refer to an
isolated runtime path. `<E2E_CANONICAL_WORK_DIR>` resolves filesystem symlinks
too, which is useful for path-keyed trust state on macOS. The harness expands
only those explicit tokens. This is intended for cached templates, drafts,
aliases, hooks, repository-local configuration, and deterministic local-mutation
preconditions.

## Read-command coverage gate

`TestReadCommandCoverage` reads the generated CLI capability inventory and
requires every non-destructive command at side-effect level 1 or 2 to have either
an E2E scenario or a reasoned entry in `testdata/read_coverage.json`. Read commands
whose destructive metadata only describes optional artifact-file overwrite are
listed explicitly there as additional targets. Stale exclusions fail once a
scenario covers the command, so adding coverage also requires removing its
temporary exclusion.

The exclusions file is the visible backlog for complete read-command coverage.
Current offline coverage includes capability/schema discovery; isolated auth,
profile, config, and project reads (including every path command and browser-safe
`project open --json`); valid and invalid configuration validation; merged
native/Firebase project aliases; a synthetic cached client/server `projects diff`;
populated client/server caches; draft list/show/diff/path; and local-hook
status/fingerprint. Root version and command help are covered in both human and
JSON forms. The draft and alias table scenarios also exercise narrow terminal
rendering. Cobra-generated completion scripts are intentionally omitted because
their source command tree is already covered.

The recorded Firebase coverage currently uses non-mutating, idempotent commands:

- `project defaults <project> --format json|xml|plist`, including JSON `--to`
- `project export <project> --json`, both inline and with `--to`
- `versions list <project> --limit 3`
- `versions show <project> <version> --json`
- `versions diff <project> <from> <to> --json`, including combined filters
- `versions export <project> <version> --to <path> --json`
- `conditions list` and `conditions show` with JSON filtering
- `conditions validate <project> --json` with an exact `validateOnly=true` guard
- `groups list` with JSON filtering
- `get --project =<project> --update`

Managed-feature coverage uses a deliberately synthetic, non-empty cached
template plus scenario-local Firebase response simulations. It covers filtered
experiment list/show, personalization list/show, and rollout list/show without
depending on an active production experiment, personalization, or rollout.

`doctor --json` uses one strict multi-host simulation containing Cloud Resource
Manager project list/details, a Firebase Remote Config read, and the IAM
permissions check. All eleven local and live diagnostics pass in the snapshot.

`projects update --json` uses synthetic Cloud Resource Manager list/detail
responses, combines auth selection, filtering, expression filtering, and URL
output, and snapshots the resulting projects registry under the isolated config
root. A fresh synthetic Remote Config cache supplies the expression context
without adding undeclared Firebase requests.

Offline local-mutation coverage snapshots both the JSON result and debug stderr,
then verifies the resulting config/cache bytes or the required deletion. It
covers config set/reset; draft change-note/discard; auth bind; project template
selection; profile switch/rename/delete; hook trust/untrust; project alias
set/remove/import; cache clear; projects forget/reset; and OAuth,
service-account, and gcloud auth add/delete variants. These scenarios all
declare an empty HTTP contract and run with `FBRCM_OFFLINE=1`.

Add flag combinations as separate scenarios when their output or HTTP behavior is
meaningfully distinct. Every scenario runs in isolated config, cache, home, and
working directories, with color disabled and a fixed terminal width and locale.
