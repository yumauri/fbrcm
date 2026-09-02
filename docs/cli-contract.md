# fbrcm machine contract v1

This document defines the stable machine interface used by agents, CI, and
scripts. Contract version `1.0.0` applies whenever the global `--json` flag is
present. Human output is unchanged when `--json` is absent.

For a concise integration workflow and safe usage examples, start with the
[agent quickstart](agent-quickstart.md).

The normative definition of "authoritative for agents," the finite audit
matrix, required evidence, and pass/fail criteria are in
[`cli-contract-audit.md`](cli-contract-audit.md). Contract reviews MUST use a
frozen version of that standard and MUST NOT introduce new acceptance criteria
during an audit.

## Invocation

`--json` and `--timeout <duration>` are global flags and may appear before or
after a command name:

```sh
fbrcm --json projects list
fbrcm projects list --json --timeout 30s
```

JSON mode writes exactly one JSON document followed by one newline to stdout,
including for Cobra parsing failures and startup failures. It never writes a
table, unwrapped usage text, progress UI, a prompt, or a second JSON document
to stdout. Explicit or implicit help places generated help text in `data.text`
under the `help` response schema. Root `--version` places its generated logo,
build metadata, and author contact text in `data.text` under the root response
schema without terminal styling.
Each `completion.<shell> --json` operation likewise places the generated shell
script in `data.text` under its own response schema and bypasses profile and
Firebase initialization.
Invoking a non-executable command group without a subcommand is also help and
uses the `help` response schema. An invalid subcommand below a group is an
argument failure under the published root response schema; it never advertises
a response-schema URN that does not exist.
The default log level is `silent` in JSON mode, so operational logs do not
normally appear on stderr. Setting `FBRCM_LOG_LEVEL` explicitly restores the
selected level, and trusted hook output may still use stderr; consumers must
parse stdout only. Non-fatal conditions that affect the meaning or recovery of
a machine result are also collected in `warnings`. JSON mode does not duplicate
the non-atomic batch warning as human stderr text.

`--timeout` must be a strictly positive Go duration and limits the complete
command, including project resolution, authentication, Firebase calls,
rate pacing, shared 429 cooldowns, retries, validation, hooks, and local
persistence. CLI execution has no startup
connectivity probe: every possible CLI network operation is represented by the
command's capability predicates.
An interrupt uses exit status `130`; an expired deadline uses exit status `9`.
Failure statuses are output-format independent: removing `--json` changes the
presentation, not the semantic process status.

## Envelope

### MCP transport boundary

`fbrcm mcp` is a streaming service, not a one-shot JSON command. Combining it
with `--json` returns `argument.invalid` (exit `2`) without opening a transport
or bootstrapping a profile; `context.profile` is null. It registers no successful
JSON DTO. Its machine capabilities describe the rejected JSON invocation's
absence of side effects; normal launch behavior is documented in [MCP.md](MCP.md).

MCP input reuses the normalized `arguments`, `options`, and `stdin` shape from
published invocation schemas. Server-controlled options are excluded and schema
references are bundled for clients. A supplied `stdin` document uses an explicit
in-memory reader, never the protocol stream. Fresh application invocations use
the same workflows, envelope builder, and semantic validation as CLI execution.
MCP is an independent frontend: it binds structured options directly, without
converting them to argv or executing the CLI's Cobra command pipeline. The
shared contract implementation lives in `ops/contract`; schema URNs
remain unchanged. Schema generation also publishes `schemas/capabilities.json`
for discovery without constructing a CLI tree at MCP startup.

Completed tools return the entire unchanged envelope in `structuredContent`
and identical JSON in text `content`. Success maps to `isError: false`; failure
and partial success map to `isError: true`, retaining usable data and warnings.
A successful changed diff (exit `1`) is not an MCP error. User interaction uses
MCP input-required results, not replacement fbrcm DTOs. Ordinary CLI JSON mode
remains non-interactive; only the hosted path explicitly enables OAuth through
its host observer.

### Result structure

Every result has this shape:

```json
{
  "schema": "urn:fbrcm:schema:cli:1.0.0:command:projects.list:response",
  "contract_version": "1.0.0",
  "command": "projects.list",
  "requested_command": "projects.list",
  "outcome": "success",
  "exit_code": 0,
  "producer": {
    "name": "fbrcm",
    "version": "1.8.0"
  },
  "context": {
    "profile": "personal",
    "offline": false,
    "dry_run": false,
    "draft": false
  },
  "data": {
    "count": 0,
    "items": []
  },
  "errors": [],
  "warnings": []
}
```

Fields are always present. Their meanings are:

| Field | Contract |
| --- | --- |
| `schema` | Response-schema URN for the executed operation. Implicit `--help` uses the `help` schema. |
| `contract_version` | Semantic version of this machine contract. |
| `command` | Stable dot-separated ID of the published response contract; the root contract is `root`, and implicit `--help` uses `help`. |
| `requested_command` | Dot-separated operation requested by the caller. It normally equals `command`; for an unknown nested command it preserves the attempted path while `command` remains `root`, whose failure schema is used. |
| `outcome` | `success`, `partial_success`, or `failure`. |
| `exit_code` | The process exit status, repeated inside the document. |
| `producer` | Binary identity and build version. |
| `context` | Effective profile plus offline, dry-run, and draft state. `profile` may be `null` for commands that do not require one. |
| `data` | Command DTO, artifact DTO, or `null` when a failure produced no usable result. A legacy top-level list is normalized to `{ "count", "items" }`. |
| `errors` | Ordered structured problems. Empty on success. |
| `warnings` | Ordered non-fatal structured warnings, including target, details, and safe remediation argv when known. |

An unknown nested argv path uses code `argument.unknown_command`, status `2`,
and `details.kind: "invocation"`; capability lookup of an unknown command uses
`command.not_found`, category `not_found`, and status `6`.

An exit status of `1` accompanies a successful diff whose `data.changed` is
true. It also accompanies a failure for an invalid validation report or failed
diagnostic report. Always inspect both `outcome` and `exit_code`.

## Structured problems

Each item in `errors` has stable fields:

```json
{
  "code": "project.ambiguous",
  "category": "argument",
  "message": "several projects match \"prod\"",
  "retryable": false,
  "target": null,
  "stage": null,
  "details": {
    "kind": "selection",
    "resource": "project",
    "query": "prod",
    "candidates": [
      {"name": "Production EU", "id": "acme-prod-eu"}
    ]
  },
  "remediation": []
}
```

Automation should branch on `code` or `category`, never parse `message`.
`message` is intended for logs. `details` is a typed object when the error has
additional machine data. `target` and `stage` identify a failed batch target or
pipeline stage when available. `retryable` describes whether repeating the
same operation may succeed without changing its inputs. It is false when the
caller must supply interaction or when Firebase has already accepted a publish
and only local post-publication work failed.

`remediation` contains safe suggested argument lists when fbrcm can provide
one. Each remediation has a `description`, a non-empty `argv` array, and a
`strategy` that defines how to use that array:

- `retry_with_arguments`: retry the original command with these arguments
  added or replacing the corresponding options;
- `replace_selector`: replace the original ambiguous selector with this exact
  selector;
- `run_command`: run `argv` as the complete fbrcm subcommand argument list
  (the executable name itself is omitted).

Agents must branch on `strategy`; they must not guess whether `argv` is a full
command or an option fragment from its contents or description.
When interaction is required, `details.kind: "interaction"` reports an
`interaction_type`, nullable `required_option`, and `destructive` marker. A
missing option value (for example `--from` or `--merge-resolve`) is described
there without emitting a remediation argument list. Remediation is present only when
fbrcm knows a complete, directly reusable argv fragment.

Problems are classified from typed source errors (`ArgumentError`,
`ExpressionError`, `ValidationError`, `ConflictError`, `SelectionError`,
`BatchError`, `ProfileError`, `ProjectLookupError`,
`RemoteConfigVersionLookupError`, import-group
selection errors, and managed-feature lookup/resource errors) and typed
Firebase, hook, context, network, and filesystem errors. Invalid configuration
keys, version selectors, and malformed experiment or rollout resource names are
argument failures. Unavailable requested versions use `version.not_found` with
selection details. Missing cached Remote Config used by `projects diff
--cached`, missing requested import groups, and missing published-template
personalizations use `parameters_cache.not_found`, `group.not_found`, and
`personalization.not_found`, respectively, with selection details. Invalid profile names use
`profile.invalid`; unavailable profiles use `profile.not_found` with selection
details; active, locally selected, or already-existing profile conflicts use
`profile.conflict`. Local validation failures and Firebase HTTP 400 candidate
rejections use `remote_config.validation_failed`; `details.source`
distinguishes their provenance. Other typed Firebase API failures encountered
during validation retain their specific authentication, permission, conflict,
timeout, rate-limit, unavailable, or request-failure classifications and
retryability. Message wording is never used for classification. Error messages and
captured hook output redact common credential forms and are bounded to 4,096
characters plus a truncation marker. This boundary also applies to messages in
command result DTOs, including target-level mutation errors.

Before a final rate-limit problem is emitted, authenticated API transports
coordinate concurrency, pacing, and retries through one controller keyed by
API host and quota consumer. `network.max_concurrent_requests` bounds requests
in flight across clients. A valid `Retry-After` response header sets the shared
429 cooldown; otherwise each consecutive 429 adds the effective
`network.rate_limit_cooldown` base until a non-429 response resets the sequence.
A positive
`network.requests_per_minute` evenly paces all attempts for that key, while zero
disables proactive pacing. Replayable transient failures use the bounded
attempt count, exponential base and maximum delays, and jitter percentage under
`network.retry`; a valid `Retry-After` overrides the calculated delay. These
waits remain part of the same invocation and are canceled by its context. JSON
mode does not change this behavior.

The published error schema defines every current `details.kind` object and an
enum of all codes emitted by this contract version. Each detailed capability
publishes its command-reachable top-level set as `problem_codes`, and its
response schema constrains `errors[].code` to exactly that set. Known codes
constrain their category and, where invariant, retryability and details kind.
Response schemas also constrain the first error category to its documented
exit status; a code, category, details kind, and status cannot contradict one
another. Nested `BatchError` target failures are an explicit open extension
point because they describe target operations rather than the aggregating
command; current target codes still use the shared typed problem shapes. The
envelope schema likewise enumerates warning codes and constrains their details.

`BatchError` details preserve every typed failed target under `failures`, with
its target, code, category, bounded message, retryability, stage, details, and
remediation. A partially successful batch uses category `partial_success` and
status `12`. When every target fails, the envelope uses the first target's
category for the process status while all target-specific categories remain in
`details.failures`; the batch is retryable only when every failed target is
retryable.

Warnings use the same stable `code`, `message`, `target`, `details`, and
`remediation` concepts but are non-fatal and do not change `outcome` by
themselves. Examples include non-atomic multi-target publication, stale cache
fallback, and a hook or cache failure after Firebase has accepted a publish.
Known warning codes constrain `details`: stale-cache fallback carries `source`,
non-atomic publication carries `target_count`, and post-publication warnings
carry their stable `stage`. `plan.source_draft_changed` carries stage
`source_draft` when apply preserves a draft edited after planning. Unknown
future codes may carry an object or `null`.
Known remediation vectors cover exact-ID project selection, `--yes`, failed
batch target selectors, cache refresh, and publishing or discarding an
existing draft.

Stable categories are `argument`, `configuration`, `profile`, `auth`,
`permission`, `not_found`, `conflict`, `validation`, `timeout`, `interaction`,
`unavailable`, `partial_success`, `io`, `hook`, `canceled`, and `internal`.

## Exit statuses

| Status | Meaning |
| ---: | --- |
| `0` | Successful result; no differences for diff commands. |
| `1` | Valid negative result: differences found, validation invalid, or diagnostics failed. |
| `2` | Invalid arguments, flags, or command path. |
| `3` | Configuration or profile failure. |
| `4` | Authentication failure. |
| `5` | Permission denied. |
| `6` | Project or other resource not found. |
| `7` | Conflict, including ETag/precondition failures. |
| `8` | Input or Remote Config validation failure. |
| `9` | Deadline exceeded. |
| `10` | Explicit interaction is required. |
| `11` | Network, offline, rate-limit, or remote-service unavailability. |
| `12` | Partial success in a batch operation. |
| `13` | Local file or stream I/O failure. |
| `14` | Publication hook failure. |
| `15` | Internal error or contract-encoding failure. |
| `130` | Interrupted or canceled. |

Diff commands return status `1` when differences exist and `0` when they do
not, in both human and JSON modes. Operational failures retain their semantic
status in either mode.

## Non-interactive rules

JSON mode never reads a menu choice or confirmation and never launches an
editor, file picker, or browser authorization flow.

Human-only theme palette swatches emitted by `theme` and `theme list` are
terminal presentation and are not represented in JSON DTOs or schemas.

- A mutation that needs confirmation returns `interaction.required` with
  status `10`; pass `--yes` after reviewing the command.
- Project import requires `--merge` or `--override` if the target already has
  content. Merge conflicts require `--merge-resolve=current` or
  `--merge-resolve=import`.
- Promotion requires `--all` or explicit selection filters. `--interactive`
  is unavailable in JSON mode.
- `theme import --json` accepts a positional file path, directory path, or HTTP
  URL, or a TOML document on redirected stdin. Stdin requires `--name`.
  Directory batches skip existing theme destinations and return structured
  `theme.already_exists` warnings. Without any input, the command returns
  `interaction.required` instead of opening the file selector.
- `config edit --json` always returns `interaction.required`. For `google` and
  imported `oauth`, `auth login --json` first tries a valid cached access token
  or refresh token. It returns `interaction.required` only when human
  authorization is needed. Any other JSON command that encounters either
  browser OAuth identity without a usable token also returns
  `interaction.required`, with
  `details.kind: "oauth_authorization"` and an `auth login <auth-id>`
  remediation; it never starts a listener or opens a browser. Service-account
  and gcloud `auth login --json` validate their existing credentials without
  interaction. OAuth refresh can contact Google's token endpoint and persist a
  refreshed token; gcloud ADC discovery can contact the metadata server when
  no local ADC source is available.
- `project open --json` returns the URL with `opened: false` and does not start
  a browser.
- In human mode, `FBRCM_OFFLINE` makes `project open` print that URL to stdout
  instead of launching a browser, in both stateful and stateless execution.
- A command that would overwrite an existing file requires explicit
  confirmation. Commands with `--yes` use it as the machine bypass. `draft
  show --to`, which intentionally has no bypass option, returns
  `interaction.required` when the destination already exists.
- Experiment and rollout deletion without `--yes` returns the resolved
  resource as a `would-delete` preview together with `interaction.required`;
  rerun the same command with `--yes` to delete it.
- Commands with documented stdin support may consume redirected stdin. Stdin
  is data, never an answer to a prompt.

## Artifact results

Commands that previously emitted raw Remote Config, defaults, draft, or export
content return an artifact DTO in JSON mode:

```json
{
  "target": "my-project",
  "media_type": "application/json",
  "encoding": "json",
  "json_content": {},
  "text_content": null,
  "base64_content": null,
  "destination": null,
  "size_bytes": 1240,
  "sha256": "...",
  "overwritten": false
}
```

`encoding` is `json`, `utf-8`, `base64`, or `none`. `none` means the bytes were
written to `destination` and are not duplicated in stdout. Exactly the content
field matching the encoding is populated. For inline JSON, `sha256` and
`size_bytes` describe the contract-normalized, compact, HTML-safe UTF-8
serialization of the JSON value in `json_content`; object-member ordering is
preserved. Verify it by re-encoding that value without indentation, with HTML
escaping enabled, and without reordering object members. Envelope indentation
is not part of the digest. For other encodings they describe the exact decoded,
returned, or written bytes. `size_bytes` is the applicable byte length,
and `overwritten` is true only when an existing destination was replaced. The
response schemas require a 64-digit lowercase SHA-256 value and a nonnegative
size, reject inline content paired with a destination, and publish the
byte-length, digest, and overwrite relationships as normative
`x-fbrcm-invariants`. These invariants use the structured extension language
published by `semantic.schema.json`, rather than implementation-specific prose.
Draft 2020-12 cannot itself recompute a digest or byte length; consumers must
evaluate the extension expression against the canonical bytes described above.
Command schemas specialize this reusable DTO: stdin Remote Config transforms
and version exports constrain inline `json_content` to the published Remote
Config shape; project exports accept any syntactically valid JSON document
because that runtime path does not decode the Firebase response; and defaults
constrain the documented JSON, XML, and plist media types. They also constrain
the encodings each command can actually return: stdin Remote Config transforms
are inline JSON; Remote Config and version exports are inline JSON or destination-only;
defaults preserve downloaded bytes as inline JSON, UTF-8, or base64, or are
destination-only; and draft recovery may
use any encoding and any inline JSON value when raw stored bytes require it.
Every returned artifact
has a non-empty target. `draft show` artifacts always report
`overwritten: false` in JSON mode because an existing destination produces an
interaction result instead of authorizing replacement.

## Capability discovery

Use discovery instead of scraping help text:

```sh
fbrcm capabilities --json
fbrcm capabilities root --json
fbrcm capabilities project import --json
fbrcm schema list --json
fbrcm schema show urn:fbrcm:schema:cli:1.0.0:command:project.import:input --json
```

The no-argument JSON form returns a compact index containing each command's
stable ID, argv path, summary, schema URNs, side-effect level, and destructive
marker. It also includes the same `supports` object as the detailed record;
`supports.stateless` is the authoritative discovery signal for stateless-mode
availability. `capabilities <command...> --json` performs an exact argv-path lookup
and returns the detailed record below. Unknown paths are `not_found` failures;
non-executable command groups are argument failures. No prefix fallback is
performed.

Every executable command's detailed capability record contains:

- stable ID, argv path, summary, positional arguments, and flags, including
  whether each accepted flag is effective for that command;
- input and response schema URNs, the shared structured-error schema URN, plus
  a stdin schema when applicable;
- the closed `problem_codes` set reachable as top-level failures for that
  command;
- side-effect level (`0` read-only through `3` remote/external), concrete side
  effects, per-effect conditions, network requirements, and network conditions;
- destructive markers, typed destructive conditions, and human-readable
  destructive reasons;
- idempotency classification and typed `idempotency_when` conditions;
- dry-run, draft, publication-plan, confirmation-bypass, stdin, and stateless
  support;
- interaction requirements, typed `interaction_when` conditions, and
  JSON-mode behavior.

Side-effect, network, destructive, and interaction metadata describes the
machine invocation represented by the command schemas. For example,
`project open --json` performs a conditional remote lookup and returns a
URL without launching a browser, while every command that can construct a
Firebase client declares conditional identity-provider or metadata-server
access, possible OAuth token-file persistence, and the exact state in which
browser authorization is required. `auth login --json` exposes the same
authentication states directly. `config edit
--json` is an interaction failure and therefore declares no editor or
destination-file side effect; like every JSON command, it can still bootstrap
default profile state while constructing the envelope. Every `auth add` variant is destructive when it
replaces an identity because existing credential or token files may be
removed. `draft show --to` is not marked destructive because JSON mode cannot
overwrite an existing destination; it returns `interaction.required` and has
no confirmation-bypass flag.

`network_when` is a disjunction of clauses; each clause's `all_of` array is
a conjunction of typed predicates over an argument, option, stdin, envelope
context, or a documented runtime state. Predicate values retain their JSON
scalar type, so Boolean options compare with `true` or `false`, not strings.
When an option is omitted from the normalized invocation, evaluate it using
the typed `default` in that option's capability flag record. The stable
`--profile` schema default is empty; process-level `FBRCM_PROFILE` precedence
is external invocation context rather than a mutable schema default. Runtime-state
names and operators are enumerated by the capability schema, and its
`x-fbrcm-runtime-state-semantics` records normatively define every permitted
name/operator pair. They cover cache
usability, successful cacheable reads, trusted hooks, output-destination
conflicts or authorized writes, mutation changes and destructiveness, accepted publication,
authentication network/authorization/token state, historical-version network
resolution, project-registry synchronization requirements, profile bootstrap,
successful project-registry persistence, import strategy and merge-conflict
requirements, confirmation authorization, and actual trusted-hook execution,
Doctor's locally usable diagnostic identity and temporary cache probe,
required editors, promotion selection, and whether a confirmation
is actually required after planning. These operators use
`null` values because the state and operator fully define the test. Context
currently exposes only the Boolean `offline` value. `network_when` is populated
whenever `network_access` is `conditional`. A planner can therefore determine
that `get` with stdin cannot access the network, `draft diff` contacts Firebase
only for `--against=current` without `--cached`, and all valid `project import`
flows require Firebase even when the imported document comes from stdin.
Experiment and rollout list/show always declare required network access for
their Firebase metadata, while their template-cache write remains conditional.
Historical version show/diff/export expose project-registry synchronization
independently of whether the immutable version itself is already cached.
`side_effect_when` contains
one record for every value in `side_effects`; an empty `when` means the
effect is unconditional, while conditional effects identify their runtime
predicate. Empty condition lists are arrays, never `null`, and detailed
records conform to the standalone capability schema both directly and when
returned by `capabilities --json`. That schema enumerates the exact published
detailed records, so IDs, paths, schema URNs, flags referenced by predicates,
stdin support, effects, and conditions cannot be recombined into a different
schema-valid capability. The `x-fbrcm-side-effect-semantics` map defines every
effect value. The effect vocabulary distinguishes authentication
network access and Firebase reads, validation, writes, and managed-feature
deletion from local cache, file, draft, general state, and trusted-hook effects.
Local file, cache, and draft deletion have distinct effect names.
`local_file_move` identifies theme-file renames, while `local_cache_move`
separately identifies relocation of an existing cache tree;
destructive
commands publish the most specific applicable deletion effects rather than
requiring an agent to infer deletion from `local_state_write`.
Imported OAuth and service-account auth commands also declare
`local_file_write` when they successfully create or replace a credential file.
Network-capable commands declare `authentication_remote_access` when they may
contact an identity provider or metadata service. They declare
`local_file_write` when a Google or imported OAuth flow may persist a new or
refreshed token outside dry-run. Doctor declares remote authentication access
but not token persistence because its diagnostic client refreshes only in
memory.
Every machine invocation declares `local_state_write` coverage for possible
envelope profile bootstrap. Commands without an unconditional local-state write
use the condition `runtime_state.profile_bootstrap required`: by default, final
envelope construction resolves `context.profile` and, when no explicit or
persisted effective profile exists, creates the default profile directories and
global configuration. Profile-managed configuration and cache state live below
`profiles/<name>` in their respective application roots. An execution path explicitly marked profileless skips
profile selection and envelope bootstrap and reports `context.profile` as
`null`; no CLI argument or environment variable selects that path yet. Commands that
resolve projects through the live registry also declare a local-state write for
`runtime_state.project_registry sync_write_succeeded`, covering persistence of a
successful Firebase registry sync after a missing or empty registry. On commands
that bypass pre-execution profile initialization, envelope-only bootstrap is
best-effort: a filesystem failure is logged but does not change the command's
machine outcome.

Theme files affect only human terminal rendering. JSON, stateless, and
`NO_COLOR` execution skip startup theme application, so theme selection cannot
alter a machine envelope. Theme management remains available in JSON mode:
list/path/current are local reads, switch and reset write configuration, delete
removes one file, rename moves a file and may update references, and import writes one
validated file. HTTP imports alone declare conditional network access through
`runtime_state.theme_source requires_network`; they do not perform Firebase
authentication. Explicit configuration commands may still read a theme file
to validate the state they inspect or modify. See
[Theming](theming.md).

In particular, draftable remote mutations declare their unconditional or
stdin-conditional Firebase read, Firebase validation for changed live
candidates, confirmation-authorized publication condition, and separate
`local_draft_write` condition. Publication workflows distinguish pre_publish
hook execution, which can occur during dry-run, from post_publish execution after
Firebase acceptance. An explicit draft-publish change note also declares the
intermediate local draft write performed before publication.

Commands whose `supports.plan` is true publish `--plan-out` as a string path.
In that mode, their Firebase write and draft write/delete effects require the
opposite `plan-out == ""` condition, while `local_file_write` is reachable for
an authorized exclusive plan destination. Remote reads, Firebase validation,
and trusted pre-publish hooks remain reachable because they establish the
exact reviewed candidate. Destructive conditions apply only when
`plan-out == ""`.

`destructive_when` uses the same predicate clauses. `destructive_reasons`
contains explanatory text for logs and review but is not used for planning.
`interaction.json_behavior` is a stable enum rather than prose.
`interaction_when` states exactly when non-interactive execution returns an
interaction problem. Every network-capable Firebase command except Doctor
includes the `authentication requires_human_authorization` condition. Commands
with additional confirmation, input, or selection branches use
`declared_conditions_return_interaction`; `project open` separately records
both browser-launch suppression and possible OAuth authorization. A command's
explicit missing-input or missing-selection clauses are preserved alongside
confirmation clauses; `yes=false` causes an interaction only when the planned
operation actually requires confirmation.
`idempotency_when` partitions conditional retry safety. End-state local writes,
stdin-only transformations, and invocations stopped for confirmation or OAuth
authorization are
idempotent. This includes `auth login` when it persists a new or refreshed
OAuth token: repeating the command converges on the same locally cached
authentication state. Remote writes remain unsafe to retry after an authorized changed,
non-dry-run plan unless the caller can establish that no publication was
accepted. A dry-run is idempotent only when no trusted hook executed; hook
commands are arbitrary local processes and may themselves be non-idempotent.
Applying an immutable plan is retry-safe only after preflight establishes each
target's state: a candidate digest already present is `already-applied`, an
unchanged planned base may proceed, and any third state is `plan.stale`.
For a plan containing publish actions, an authorized retry is declared safe
only when no trusted hook executed; if an arbitrary trusted hook executed, the
invocation is not declared idempotent even when Firebase publication later
converges. A plan containing only `none` actions performs no Firebase request
and is retry-safe. Plan creation follows the same hook boundary: exclusive file
creation is retry-safe when no trusted hook ran, and is not declared safe after
a trusted pre-publish hook ran.

Authentication failures retain their source semantics. Missing or malformed
stored credentials are auth failures. OAuth `invalid_grant` or a missing
refresh token may require human authorization, while transient token-endpoint
network, timeout, rate-limit, and service failures remain typed retryable
failures and are never converted into `interaction.required`.

Behavior metadata comes from an exhaustive command manifest. Schema generation
fails if an executable command has no entry or the manifest names a command
that does not exist, so side effects cannot silently fall back to name-based
guesses.

The input schema describes a normalized invocation object with `arguments`,
`options`, and `stdin`. It is intended for planners and command builders; the
CLI continues to receive ordinary argv and documented stdin bytes.
`x-fbrcm-normalization` lists the ordered operations that convert raw argv to
that object. Published operations include Unicode-whitespace trimming,
lowercase or uppercase canonicalization, and stable first-occurrence array
deduplication. Query target selectors publish case-insensitive prefix
canonicalization plus trimming around their project query; positional target
selectors publish prefix canonicalization while preserving the resource name
or ID exactly without trimming. Template arrays publish their canonical
`client`, `server` order. Nested configuration-key
normalization uses a typed conditional-prefix operation so top-level keys
retain the runtime's exact comparison behavior. `x-fbrcm-matching` declares
selection semantics that do not necessarily replace the invocation value:
mode-prefixed filters publish their resource fields, exact prefix map, and
algorithm; condition and group searches publish case-insensitive substring
fields; and parameter search publishes its simultaneous normalized-text and
raw-text variants. Invocation-level composition rules declare repeated values
as OR, distinct supplied selector sources as AND, and absent sources as
matching all candidates. Target-aware rules also declare unqualified
enabled-template expansion, explicit single-template selection, and
client-target canonicalization. Invocation schemas for bulk `--project`
commands declare the all-configured-project and enabled-template default.
Every present repeatable option is a nonempty array: omission represents no
flag occurrence, while an empty array has no argv representation and is
rejected.
Positional project selectors use a separate matching rule: after optional
template-target parsing they preserve argv, treat `=`, `^`, `/`, and `~`
literally, and compare exactly and case-sensitively by project ID, repository
alias, then display name. Target-aware positional rules declare that the
unqualified form resolves the configured primary template.
Draft positional selectors use a command-local rule instead: they resolve only
existing local draft target IDs and compare the untrimmed selector exactly and
case-sensitively by physical project ID, repository alias, then display name.
Mode-prefix characters remain literal. An unqualified selector uses the
configured primary template, or the client template when the project is no
longer registered. Draft-list filters likewise operate only on existing drafts
and apply configured enabled-template selection, with the same unregistered
client fallback. Draft publish and discard require exactly one of positional
selectors or `--all`, then deduplicate and sort canonical target IDs.
An alias whose mapped project is unavailable in the active profile returns the
same typed `project.not_found` selection problem as other unsuccessful project
resolution, while retaining the alias mapping in the safe problem message.
Explicit positional arguments and supplied string option values
must not be empty or whitespace-only. Exact empty strings remain valid only for
content options whose empty value has meaning: `--value`, `--description`,
`--change-note`, scalar `--group`, and `--label`. The
`reject_raw_whitespace_only` validation operator preserves that distinction for
content that is subsequently trimmed. `config set`, `config show`, and
`config reset` trim nested
`keys.<block>.<action>` and
`network.*` and `projects.aliases.<alias>` keys before their closed grammar is evaluated;
`config show` also trims nested `hooks.*` keys. Top-level configuration keys
are compared without trimming. The optional `get [parameter]`,
`update [parameter]`, and `delete [parameter]` arguments likewise have no argv
normalization. Each compares exactly and case-sensitively against canonical
parameter keys, participates in the command's selector composition, and is
mutually exclusive with `--filter`; a mismatch yields the documented empty or
no-op result. More generally, positional existing-resource names and IDs are
literal, untrimmed, exact, and case-sensitive. Search behavior is confined to
explicit query options such as `--filter`, `--search`, and `--project`.
An option annotated with `x-fbrcm-effective: false` is accepted by argv parsing
but not applied by that command. Conditional applicability uses
`x-fbrcm-effective-when`; detailed flag records expose the same distinction as
`effective` and `effective_when`. Root version output, for example, accepts but
does not apply `--profile` or `--no-local-config`. Unconditional
`effective: false` currently applies to `--profile` on `help`, contract
metadata, config, hooks, and project-alias commands; to `--noopen` on
machine-mode `auth login`; and to `--editor`, `--full`, and `--scope` on
machine-mode `config edit`. The latter commands return interaction metadata
before those human-interface options can affect execution.
Effective `--profile` values are trimmed and then validated against the shared
filesystem-safe `path_segment` grammar. `auth bind --auth` uses the same grammar
without trimming. Root models that profile validation conditionally because
`--version` accepts but does not apply the profile value.

Invocation schemas encode command-specific flag values and numeric bounds,
including case-insensitive runtime values without falsely rejecting accepted
capitalization, each effective config command's accepted scopes, the Firebase condition
color set, condition-add priority zero, and the exact `strconv.Atoi` grammar
and overflow behavior used for positive positional condition-move priorities.
The `condition_priority` runtime-validation annotation also publishes the
template-dependent bounds: add accepts zero as append or priorities through
the existing condition count plus one, while move accepts priorities through
the existing condition count.
Native `int` options use a portable machine-contract maximum of 2,147,483,647.
They also encode positive Go-duration syntax for the global timeout and attach
a normative `parse_duration` rule using `time.ParseDuration`, so agents can
reject overflow and values that parse or round to zero before invocation.
Cobra mutual exclusions, including options unavailable with the
implicit machine-mode `--json`, and semantic requirements such as edit
commands requiring at least one edit, draft publish/discard requiring either
projects or `--all`, template primary membership, and `--condition` requiring
a real value source (`--use-in-app-default=false` is not one). Typed Remote
Config group-description edits likewise require either `--description` or a
truthy `--no-description`: explicit false is semantically absent and may
coexist with `--description`, but does not select description removal.
Condition color edits apply the same rule to `--no-color`: explicit false is
absent, may coexist with `--color`, and cannot satisfy the required edit by
itself.
Remote Config values publish Boolean and Go-number grammars, including the underscore,
hexadecimal, and case-insensitive special-value forms accepted by
`strconv.ParseFloat`. JSON text is identified with `contentMediaType`,
`contentSchema`, and a structured `parse_json` validation rule that requires
one complete RFC 8259 value. Empty condition names,
rename targets, and conditional-value removals are rejected; applicable
parameter/group names, duplicate source/target names, and descriptions carry
their 256-code-point limits. Positional get/update/delete parameter selectors are
mutually exclusive with `--filter`; duplicate source and target must differ
by exact code-point comparison after the target's declared creation-name
normalization. `config show`, `config reset`, and `config set`
publish their distinct closed key vocabularies. `config set` also
publishes key-specific value arity and types, its terminal-key grammar,
modifier-token uniqueness rule, and the required local scope for repository
aliases. Function keys use canonical
decimal spellings from `f1` through `f63`; zero-padded and signed variants are
invalid. Version selectors publish the exact signed-parser behavior used by the
current runtime: the complete selector is untrimmed and case-sensitive,
positive numeric components may contain an optional `+` and leading zeroes,
absolute values must fit signed 64-bit parsing, and relative distances must fit
signed 32-bit parsing and remain at most 299. Version
matching metadata publishes symbolic resolution as well: `current` and
`latest` select the current publication, `previous` selects one publication
before it, relative forms select the declared distance before current, live
mode resolves Firebase history, cached mode resolves local snapshot numbers,
and an omitted `versions diff` destination defaults to current. Version
`--since` and `--until` values publish the exact Go
`time.RFC3339` parser rule used at runtime instead of relying on the broader
JSON Schema `date-time` format. Version `--before` is a canonical positive
decimal integer when supplied: empty values, signs, zero, leading zeros,
whitespace, and non-decimal forms are rejected before target resolution.
Selector and filter-query syntax is modeled
with patterns, enums, normalization, and typed matching rules. Experiment and
rollout IDs also publish a `managed_feature_id` validation rule that
accepts an untrimmed slash-free ID exactly as supplied or the exact
case-sensitive resolved-project Firebase resource name; remote selection still
succeeds only for an exact canonical existing ID. Positional condition
selectors publish untrimmed exact
case-sensitive lookup and `condition.not_found`; the non-positional update
`--condition` option separately publishes its exact-then-case-insensitive
runtime lookup. Duplicate source lookup publishes untrimmed exact
case-sensitive cross-group matching, no-match skip, and typed ambiguity when
the same exact key occurs in multiple locations. Personalization lookup
publishes untrimmed exact case-sensitive equality and its typed not-found
boundary. Named group mutations publish untrimmed exact case-sensitive map-key
lookup; zero matches skip the target and differently cased keys remain
distinct. Existing auth IDs, profile names, repository project aliases, and
embedded schema IDs likewise publish their command-specific exact
case-sensitive lookup and zero-match outcomes.
Expression values are
nonblank strings whose authoritative expr-lang version, Boolean result type,
operators, functions, context fields, and value typing are embedded in the
semantic schema; compilation and type checking remain an explicitly declared
runtime validation boundary. Stdin is either `null` or the command's concrete
payload. Remote Config stdin schemas model the local decoder, not Firebase
acceptance: they preserve the exactly-one-option parameter-value union and
reject unsupported condition fields and JSON `null` where a typed object,
string, Boolean, or value is required, while allowing condition names,
expressions, colors, parameter types, description lengths, duplicate condition
names, and forward-compatible object fields that local transformation paths
accept before any later publication validation. Firebase semantic annotations
remain on validated response DTOs and publication boundaries, not on offline
stdin decoding. `project import` uses a separate
`stdin:remote_config_import` schema because it also requires the local
`firebase.NormalizeRemoteConfigForUpdate` validation performed before import
selection and transformation. Published stdin modes contain only normalized
JSON documents. `apply`, `plan show`, and `plan validate` accept a strict
publication-plan document when their positional path is `-` and reference
`stdin:publication_plan`; invalid, tampered, and unsupported documents retain
the typed `plan.invalid`, `plan.integrity_failed`, and
`plan.unsupported_version` problems. Their `path_or_stdin_document` input rule
states that `-` consumes one stdin document, while a file path reads the file
and leaves stdin unconsumed. The plan schema publishes
`publication_plan_integrity` checks for a nonempty, canonical, unique target
set, snapshot digests, `none` equality, action/validation provenance, and the
content-derived plan ID. Embedded templates accept the forward-compatible
shape used by the runtime decoder and publish the offline
`firebase.PrepareRemoteConfigUpdate` validation boundary; plan inspection does
not claim remote Firebase validation. Imported OAuth and
service-account auth commands reference distinct
credential-object schemas whose required IDs, secrets, keys, email address,
and endpoint URIs match runtime validation. A sole OAuth `installed` or `web`
client requires the complete credential fields and at least one redirect URI.
When both objects are present, the schema models the runtime's split precedence:
Google selects `web` and requires its redirect array, while fbrcm validates the
ID, secret, and endpoints from `installed`; agents should prefer one complete
object. URI rules explicitly publish `require_absolute: true` in addition to
the exact Go parser and whitespace normalization. Service-account email fields
publish their exact Go parser rule. OAuth, service-account, project-import, and
theme-import invocation schemas publish typed `x-fbrcm-input-selection` rules:
a nonempty file, directory, or URL source is selected before `stdin.document`, a later
stdin document is not consumed when the earlier source wins, and absence of
both sources produces `interaction.required`. Theme import publishes a
`toml_document` stdin mode and requires `options.name` when stdin is selected.
Its experimental human-only directory-stdin transport remains outside schemas
and capability metadata. Capability lookup represents a
multi-component executable command path as an argument array and publishes
typed not-found and non-executable-group results. Help uses a separate
longest-existing-prefix rule: unmatched suffix components are ignored, and an
unknown first component renders root help. The root schema includes the
executable `--version` option. Schema lookup publishes exact case-sensitive
embedded-ID matching and a typed `schema.not_found` result.

`GOOGLE_CLOUD_QUOTA_PROJECT` is inherited process context rather than a command
argument, so it is not represented in invocation schemas or capability flags.
Persisted auth and project quota values are represented by the typed
`auth.quota-project.show|set|unset` and
`project.quota-project.show|set|unset` commands. Their responses distinguish
configured and effective values and enumerate the resolution source.

The `--quota-project` option on every `auth add` operation trims Unicode
surrounding whitespace before normalized-invocation validation. The resulting
value must satisfy the literal physical-project-ID grammar; target prefixes and
selector/query syntax are invalid arguments.

Every authenticated Firebase and Cloud Resource Manager request includes
`X-Goog-User-Project`. Stateful precedence is the environment override,
project override, auth default, ADC `quota_project_id` for gcloud auth only,
then the physical request target. Credential files for Google OAuth, imported
OAuth, and service accounts never supply the ADC credential source. A
targetless request with no resolved source fails before network access as
`auth.quota_project_required`;
its details identify `targetless: true`. An invalid nonempty environment value
fails before auth interaction or network access as
`auth.configuration_invalid`; invalid command input is `argument.invalid`, and
an invalid manually edited persisted file is `configuration.invalid`. Invalid
`quota_project_id` metadata in ADC selected for a gcloud identity fails as
`auth.credentials_invalid`. Auth-category failures use semantic exit status 4
and retain the selected auth identity as their problem target when one is
available.

`FBRCM_GOOGLE_ACCESS_TOKEN` is inherited process context. The explicit root
`--stateless` option selects profileless static-token authentication for the
supported commands published by `capabilities`: parameter reads and mutations;
condition reads, validation, and mutations; group reads and mutations;
experiment and rollout reads and deletion; personalization reads; project
defaults, export, import, open, and show; project discovery, diff, and
promotion; and version list, show, export, diff, and rollback. It works in both
human and JSON output modes and requires a nonempty token, except that `project
open` and stateless `get` with a stdin document are local-only and do not
require the token. The token itself is never
included in an invocation schema, response, log, or remediation. Except for
the project-filter behavior documented below for `get`, parameter mutations,
`groups list`, and `projects list`,
each remote target is interpreted as a literal physical Firebase project ID
with an optional `client@` or `server@` template prefix. Profile, alias, and
configured-primary resolution are skipped, `context.profile` is `null`, and the access token is neither refreshed
nor persisted. Repository configuration is disabled, `FBRCM_PROFILE` is
ignored, and an explicit `--profile` conflicts with `--stateless`. Version
commands also reject `--cached` because stateless execution cannot read local
snapshots. Commands with `--update` reject it because stateless reads are
already live. Setting the token without `--stateless` does
not alter normal profile, project-resolution, or configured-auth behavior. The
quota-project precedence is `GOOGLE_CLOUD_QUOTA_PROJECT`, then the requested
physical Firebase project ID; static-token mode has no credential file from
which to read an ADC `quota_project_id`.

Stateless `projects list` performs live paginated Cloud Resource Manager
discovery and per-project metadata enrichment. It has no requested target for
quota fallback, so only `GOOGLE_CLOUD_QUOTA_PROJECT` supplies a quota project
for the initial list request. The command skips repository aliases and all
profile-managed state. Its `filter` option matches only remote display names
and project IDs; `url` remains effective. Invocation schemas reject `update`
when true because discovery is already live. `expr` remains available and is
evaluated after `filter`; the command directly fetches the current client
Remote Config template for each remaining project and builds the standard
project expression context without reading or writing a cache.

Stateless `versions list` reads Firebase version history without consulting
the current cache pointer or immutable snapshots. Its items therefore report
`cached: false` and `current: false`; `cached` is rejected while the normal
live pagination, time, and version-number filters remain available. Stateless
`conditions list` fetches the latest Remote Config without cache or draft
overlay, then applies its name, search, and expression filters locally.
`update` is rejected for conditions because stateless reads are already live.

Condition show uses the same direct template path, and condition validate sends
the downloaded template and ETag to Firebase's validate-only endpoint without
draft inspection. Group list reuses the stateless `get` target-selection rules
and applies group filters locally. Version diff fetches both selected versions
without snapshot reads, and projects diff fetches both literal template targets
without project-registry resolution; both reject `cached`.

Managed-feature list/show commands accept one literal physical project ID and
reject client/server prefixes. They fetch current project metadata to obtain
the numeric project identifier, then read the published client template
without caching. Experiment and rollout commands also contact their live
metadata endpoints; personalization commands derive their records from the
template. Project show performs one live metadata request and leaves all
profile-derived fields empty.

Stateless experiment and rollout deletion use the same literal physical project
and metadata resolution, fetch the named resource before confirmation, and
then send its Firebase DELETE request directly. They do not fetch or rewrite
the Remote Config template. JSON mode without `yes: true` returns the normal
`would-delete` data plus `interaction.required`, and no DELETE request is sent.

Stateless group add/edit/rename/delete reuse the same target-selection rules as
stateless group list. Each selected template is fetched directly, transformed
in memory, submitted to Firebase validation, and published with the fetched
ETag. The execution policy suppresses project-registry and cache writes,
immutable version snapshots, draft inspection, and pre/post-publish hooks.
`draft: true` is rejected by both runtime validation and invocation schemas;
dry-run, change notes, confirmation, and structured per-target outcomes remain
available. Multi-target publications remain independent and non-atomic.

Stateless condition add/edit/move/rename/delete accept exactly one literal
client or server template target. They use the same direct fetch, in-memory
transformation, Firebase validation, ETag publication, and persistence-policy
suppression as stateless group mutations. Condition ordering and parameter
references are transformed by the existing condition domain operations.
`draft: true` is rejected; dry-run, change notes, confirmation bypass, and
structured mutation outcomes remain available.

Stateless parameter add/duplicate/update/delete use the same project-selection
contract as stateless remote `get`: exact `=project-id` selectors bypass
discovery, non-exact selectors filter one live discovery result, and an omitted
selector selects all accessible projects. Optional client/server prefixes are
preserved, repeated selectors are ORed and deduplicated, and repository aliases
never participate. Each selected target is fetched, transformed, validated,
and published independently. Parameter matching for update and delete and
project-context expressions for add and duplicate are evaluated against the
directly fetched template. The persistence policy suppresses project registry,
cache, snapshot, draft, and hook access; `draft: true` is rejected while
dry-run, change notes, confirmation bypass, and structured partial outcomes
remain available.

Stateless `project import` accepts a literal client or server target and keeps
the normal explicit file-before-stdin input selection. It fetches the current
template directly, applies import selection and merge logic in memory,
validates against Firebase, and publishes with the fetched ETag. Explicit input
file access is caller-directed rather than application-managed state and
remains available. Draft reads and writes, cache and snapshot persistence, and
hooks are suppressed; `draft: true` is rejected by runtime and schema.

Stateless `projects promote` accepts two literal client or server targets. It
reads both templates directly, applies the existing local selection plan,
reloads and revalidates the target, and publishes with the latest ETag. A typed
ETag conflict reloads the target, reapplies the selected plan, and retries.
Profile aliases, drafts, caches, snapshots, and hooks do not participate;
dry-run reaches Firebase validation but omits publication.

Stateless `versions rollback` accepts one literal client or server target. It
loads source and current history directly, performs the normal pre-publication
current-version recheck, validates the candidate, and uses Firebase's native
force-publication rollback endpoint. No snapshot, cache, draft, or hook state
is accessed or updated. Dry-run performs the reads and validate-only request
without sending the rollback POST. The documented native rollback race window
after the final recheck is unchanged.

Stateless plan production records policy `stateless` and literal target IDs;
it performs the same direct reads and Firebase validation as the corresponding
mutation while suppressing publication, caches, drafts, snapshots, and hooks.
`apply --stateless` accepts only a stateless plan, preflights every target using
the one-shot token, and publishes exact candidates with their planned ETags.
A stateful plan must be applied statefully and its effective hook-definition
fingerprint must still match. Policy or hook drift is `plan.stale` before any
target write.

Stateless remote `get` discovers all token-accessible projects when `project`
is omitted. A non-exact selector uses the normal fuzzy, starts-with (`^`),
contains (`/`), or fuzzy (`~`) modes against remote project IDs and display
names only. An exact `=project-id` selector instead treats its query as a
literal physical ID and bypasses discovery. Optional `client@` and `server@`
prefixes select the requested template kind; unqualified discovered projects
default to client. Repeated selectors are ORed, mixed direct and discovered
targets are unioned and deduplicated, and discovery runs at most once. No
repository aliases or configured template preferences participate. The
command fetches every selected latest template without stale-cache fallback
and applies parameter, search, and expression filters in memory. It reports
`cached_at: null` and status `fetch`; `update` is rejected because the read is
already live. Discovery has no requested target for the initial quota fallback,
so it may require `GOOGLE_CLOUD_QUOTA_PROJECT`; exact-only selection retains
the target-project fallback. Existing stdin document mode is also
stateless-capable and tokenless, while `update` is rejected for every stdin
invocation because stdin is already the complete input.

`project open <project>` also supports `--stateless`, but performs no Firebase
API request and therefore requires no access token or quota project. Its
project argument is a literal physical project ID without `client@` or
`server@` syntax. Human mode opens the constructed Firebase Console URL; JSON
mode returns the URL without launching a browser. Human offline mode also
prints the URL without launching a browser.

After a supported stateless invocation passes setup validation, fbrcm emits an
info-level `stateless mode enabled` log record with component `cli.stateless`
and the stable command ID; credential values are never logged. JSON mode's
default silent log level suppresses this record unless logging is explicitly
enabled.

Every command execution carries an explicit persistence policy. Normal mode
enables application-managed local reads, local writes, and configured hooks.
Stateless mode disables all three controls; service resolution, parameter and
version caches, draft entry points, publication cache updates, and hook
preparation consult that policy instead of a command-specific stateless
marker. The permissive policy is also the default for internal callers that do
not attach one, preserving existing stateful behavior. Explicit caller-chosen
artifact output such as `project export --to`, `project defaults --to`, or
`versions export --to` is not application-managed local state and remains
permitted.

`--to` and `--yes` retain the normal artifact destination and overwrite
contracts. Every supported stateless invocation schema rejects `profile` and
rejects state-dependent options where they exist. Project-target schemas
conditionally narrow positional arguments to literal physical IDs with
optional template prefixes where supported. Project-scoped metadata and
managed-feature commands require a physical ID without a prefix; template-aware
commands publish client as the default when `options.stateless` is true.
Capability side-effect and
interaction conditions mark profile bootstrap, project-registry persistence,
configured-auth token persistence, identity-provider access, and browser
authorization for Google or imported OAuth identities as stateful-only. An
explicitly requested `--to` write remains available in either mode.

Direct mutation result DTOs expose selection and no-op provenance:

```json
{
  "selection": {
    "default_scope": true,
    "resolved_target_count": 4,
    "matched_item_count": 17
  },
  "changed_item_count": 0,
  "no_op_reason": "already_applied"
}
```

`no_op_reason` is `no_match` when the selector matched no item and
`already_applied` when items matched but required no change. The selection
object is present on every direct Remote Config mutation target, including
successful changes and failures.

`auth delete` success data correlates `type` with `deleted_paths`: `google` and
`service-account` return one nonempty path, imported `oauth` returns two unique
nonempty paths, and `gcloud` returns `null` because no application-managed
credential file exists.

## Published schemas and versioning

Draft 2020-12 JSON Schemas are checked in under `schemas/cli/1.0.0/` and
embedded in the binary. Each executable command has one input schema and one
response schema. Shared envelope, error, capability, semantic, general Remote
Config stdin, project-import Remote Config stdin, combined credential stdin,
OAuth credential stdin, service-account credential stdin, strict publication
plan stdin, and standalone publication-plan schemas are published beside them.

Every command response schema references the shared published envelope schema
and defines the command's exact successful `data` DTO. Enforceable correlations
such as config-reset status/changed pairs, cache-clear status/count states,
alias-removal states, auth-bind item status/reason pairs,
import/promotion/publication/version states, config-validation diagnostic
severity, and direct mutation status fields are expressed with Draft 2020-12
conditions. Count and size fields are nonnegative. Closed response
vocabularies, including configuration scopes and sources, diagnostic codes,
alias-import conflict policies, diff change kinds, auth types, template kinds,
draft comparison bases, version operations, and managed-feature kinds, are enums
or command-specific constants rather than open strings. Auth results correlate
the outer identity and type with their path object and enforce the credential
paths applicable to each auth type. Project template results require a unique,
nonempty subset of `client` and `server`, with the primary template included in
that subset. Persisted project `updated_at` and `synced_at` values remain plain
strings because the accepted local projects registry does not validate their
format. Human confirmation-decline states are not included in JSON
success DTOs: machine mode fails with `interaction.required` before prompting.
Version restore/rollback no-ops and unchanged draft-publish dry runs use the
explicit `unchanged` status; `published` and `would-publish` therefore always
mean that the represented operation changed or would change Firebase state.
Live publication of an unchanged draft remains `already-applied` because that
operation removes the redundant local draft.
Plan-creation variants return `plan_id`, private destination `path`,
`created_at`, `command_id`, total/publish/unchanged target counts, and
non-secret metadata for the exact file artifact: fixed publication-plan media
type, `none` encoding, destination, byte size, lowercase SHA-256, and
`overwritten: false`. The destination equals `path`, and the total count is the
sum of the publish and unchanged counts. `plan show` returns a non-secret
summary whose counts correlate with its target array, while `plan validate`
returns `valid: true` and the verified nonzero count. `apply` returns its plan
ID, dry-run marker, accepted publication count, and typed per-target statuses;
the count includes `published`, `published-hook-failed`, and
`published-cache-failed`. Status-specific schema branches constrain dry-run
reachability, validation provenance, publication-version presence, and error
stage. The sensitive publication-plan document remains a private file rather
than response content.
Experiment and rollout deletion previews correlate `status: "would-delete"`
with a failure envelope containing `interaction.required`; `status: "deleted"`
requires a success envelope. Deterministic machine fields such as
`project open.opened`, successful condition validity, profile-rename changes,
hook trust state, profile-switch state, alias-import actions, and version-show
cache provenance are constants or conditionally constrained. Draft, version,
and promotion post-publication statuses constrain validation provenance to
`validated: true` with `validation_source: "firebase"` and constrain their
status-specific error stage. Direct-mutation statuses similarly correlate the
change count, no-op reason, validation provenance, publication version, retry
selector, dry-run/draft mode, and error stage with the phase actually reached.
Relationships that Draft 2020-12 cannot calculate, such as collection counts,
the unique default auth/active profile markers, auth-bind aggregate counts,
artifact digests, and byte lengths, are published as structured normative
`x-fbrcm-invariants` on the affected DTO. The shared envelope
contains the problem definition. Reusable enums and grammars are referenced
from the published semantic schema, including
mutation statuses, no-op reasons, validation sources, artifact encodings,
target selectors, literal physical project IDs, alias and path-segment names,
version selectors, filter modes and queries, and expression contexts.
Collection commands describe the normalized `{count, items}` data object and
the complete item shape; mutation and lookup commands describe their named
result DTOs. For example, `groups.list` publishes the group identity and count
fields (`name`, optional `description`, and `parameter_count`) alongside its
project, template version, source, and draft provenance. `null` is also
accepted for failures without usable data. Only
`schema show`, whose result is the requested schema document itself,
intentionally permits an arbitrary JSON object.

Schema validators should register the published envelope, semantic, and
capability resources by their `$id` URNs before compiling a command schema.
Command responses can reference all three transitively. The resources are
available through `schema list` and `schema show` and are embedded in the
binary.

The shared envelope schema also enforces the relationship between outcome,
status, data, and errors: success has status `0` or the documented valid
negative status `1` and no errors; partial success has status `12`, usable
non-null data, and at least one error; failure has a nonzero status and at
least one error. Failure status must also match the first problem's category.
Each command response schema narrows that shared outcome set: only commands
that can publish Remote Config admit `partial_success`, and commands with no
successful DTO admit only `failure`. It similarly narrows warnings to the
codes reachable from that command; commands with no warning path require an
empty warning array.

Executable commands must register every successful response DTO with the
contract package. Schema generation fails when a command has no registration,
and contract tests compare every checked-in response schema with its registered
DTO so stale or generic data schemas cannot be published.
At runtime, fbrcm also validates a registered command's completed envelope
against the standard Draft 2020-12 keywords in the response schema it
advertises. A mismatched DTO, invalid enum, enforceable field correlation, or
other JSON Schema violation is replaced with a conforming
`internal.contract_violation` failure at status `15`; arbitrary captured text
cannot silently become a successful `data.text` payload for another command.
The semantic schema publishes the versioned extension language for
`x-fbrcm-validation`, `x-fbrcm-normalization`, `x-fbrcm-matching`, and
`x-fbrcm-invariants`,
including its operators, term forms, evaluation order, pass condition, and
Draft 2020-12 schemas for each rule AST. All four annotations use arrays of
structured rule objects; Remote Config validation uses the same typed
`remote_validate` and `unique_by` operators instead of prose-only annotations.
These extensions are normative: an agent claiming complete conformance must
evaluate them. They remain custom annotations to a standard Draft 2020-12
validator, so Draft validation alone is intentionally only the structural
subset of the fbrcm contract. Schema generation validates every occurrence
against the published rule schema and fails on an unknown operator, a missing
operand, an extra operand, or the wrong annotation wire type.

Schema generation stages the complete v1 schema directory, capability goldens,
the command-by-test-class audit-evidence golden, and contract lock, then
publishes those assets as one rollback-capable rename transaction. A failed
publication restores every previous asset instead of leaving a partially
replaced schema directory.

Regenerate the registry after any command-contract change:

```sh
go run ./cmd/schemagen
```

Contract tests require every executable command to have both schemas,
resolvable local or semantic references, and a complete capability record in
the reviewed golden file. They also expand the audit-evidence golden into all
15 section-6 test classes for every executable command, require either known
test/scenario IDs or a justified `N/A` in every cell, and reject stale evidence
references or applicability decisions. Runtime conformance tests execute a failure envelope
for every executable command and validate it with a Draft 2020-12 validator;
focused scenarios cover success, empty/no-op data, invalid input and
expressions, interaction, validation sources, timeout/cancellation, mixed
batches with typed target failures, help/version response modes,
post-publication failures, warning remediation, and bounded secret redaction
in envelope and result messages. A changed command ID,
removed or renamed field, changed field type, changed meaning, or changed exit
status requires a new major contract version. Additive optional fields and new
commands require a minor version. Documentation-only corrections may use a
patch version.

`schemas/cli/contract.lock.json` fingerprints the current generated schemas,
capability goldens, and audit-evidence golden and records whether that contract
has been released.
While `released` is `false`, the unreleased v1 contract may be corrected without
inventing a version bump. Once it is set to `true`, the generator rejects a
different fingerprint at the same version. Generation happens in a staging
directory and validates the lock before replacing checked-in files, so a lock
rejection leaves the working tree untouched. Previous version directories are
not removed by generation.
