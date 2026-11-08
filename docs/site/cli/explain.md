<Badge type="info" text="Since 1.1.0" />

# Explain a parameter value

`fbrcm explain` shows why a published Remote Config template would select a
value for one request. It takes one exact, case-sensitive parameter key,
including keys inside parameter groups. Use it with a captured fetch request or
with inputs you supply yourself. The command runs in the selected fbrcm profile.

## Start with a capture

In your network inspector, find the Firebase Remote Config `:fetch` POST. Copy
the request as cURL and save it as text, or export a HAR with the response body
if you also want to see what Firebase returned.

```sh
fbrcm explain feature_enabled --from fetch.curl
fbrcm explain feature_enabled --from session.har --entry 3
fbrcm explain feature_enabled --from - < session.har
```

fbrcm detects the format from the content and also accepts piped stdin without
`--from -`. An explicit file path takes precedence over stdin. Captures are
parsed locally and the captured request is never replayed. The capture and
decompressed bodies are limited to 32 MiB.

HAR entry numbers are one-based positions in the original file, including
unrelated traffic. Entry `3` means the third original entry. With one Firebase
fetch, selection is automatic. With several, the interactive command opens a
picker unless you supply `--entry`. Use `--entry` in scripts with multi-fetch
HARs.

The fetch identifies the Firebase project and the client or server template.
If you also supply `--project`, it must resolve to the same physical project.
An explicit `client@` or `server@` prefix must agree with the capture.

## Try a scenario without a capture

```sh
fbrcm explain feature_enabled --project '=example-project-id' \
  --set device.os=web \
  --set "app.customSignal['environment']=beta"
```

Without a capture, `--project` (or `-p`) is required. It selects one project by
ID, alias, display name, stored project number, or a query that resolves to
exactly one project. An unqualified selector uses the project's primary
template. Use `server@=example-project-id` to select its server template
explicitly.

Each `--set field=value` overrides the captured or derived input for that
field. If you repeat a field, the last assignment wins. Common inputs include
`device.os` (`web`, `android`, or `ios`), `device.country`, `device.language`,
`app.version`, `app.build`, `app.userProperty['tier']`, and
`app.customSignal['environment']`. Use `true` or `false` for an audience input
such as `app.audiences['Beta testers']`, an RFC3339 timestamp for
`device.dateTime`, and a number from 0 to 100 for a Firebase-assigned percentile
such as `percent` or `percent['checkout']`.

Use `/unknown` to leave a fact unresolved and `/unset` to state that the field
is absent. A JSON-quoted value supplies a literal string, for example
`--set 'device.language="/unset"'`. The interactive command can ask for missing
facts. Use `--no-input` to suppress those prompts while keeping the human
report, or `--json` for a noninteractive structured report.

## Read the explanation

The report identifies the template version and source, then traces the
conditions that supply the parameter in template priority order. A
higher-priority unknown can leave several outcomes possible even when a later
condition matches. The default applies when none of the parameter's conditions
match. Unsupported expressions or unavailable Firebase facts can leave an
outcome unresolved.

For example, suppose `web_beta` requires both `device.os=web` and the `beta`
environment signal, and a lower-priority `web_all` condition matches any web
request. With the platform set to web and the signal unknown, either condition
can supply the value. Set the signal to `beta` to resolve the winner to
`web_beta`.

By default, fbrcm uses a fresh cached published template and checks Firebase
when the cache is stale or missing. Use `--update` to check now or `--cached`
to accept an existing cache even if stale. `--cached` can still fetch the
template if the cache is missing. The report uses the published template from
this read, even when the capture came from another version.

A HAR response shows what Firebase returned for this parameter when that
evidence is available. A comparison can be inconclusive if the template
versions differ or if Firebase assigned an experiment, rollout, or
personalization. An in-app default identifies a fallback whose value lives in
the app. The report describes template evaluation and fetch evidence. It cannot
tell you which value the app has activated or kept in its local cache.

For flags and exact input field names, see the [`explain` command reference](/reference/commands/explain).
For scripting, see the [explanation JSON results](/automation/json-contract#parameter-explanations).
To try the same inputs across a loaded template or draft, see [TUI Context](/tui/context).
