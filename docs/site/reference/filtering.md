# Filtering and expressions

Use `--project` and `--filter` to match project names and parameter keys. Use
`--search` to search text fields, or `--expr` to select typed Remote Config
data.

## Query modes

Flags named `--project` or `--filter` accept a mode prefix:

| Prefix | Mode |
| --- | --- |
| none or `~` | Fuzzy match |
| `^` | Starts with |
| `/` | Contains |
| `=` | Exact, case-insensitive |

```sh
fbrcm get --project '^acme-prod' --filter '^checkout_'
fbrcm get --project '=example-project-id' --filter '=feature_enabled'
```

Repeated values of the same selector are ORed. Different selector sources are
ANDed: for example, a parameter must satisfy `--filter`, `--group`, `--search`,
and `--expr` when all four are supplied.

Project queries match display names, project IDs, and repository aliases.
Parameter filters match keys. Positional resource names are different: they
match canonical names or IDs exactly and case-sensitively.

## Search

Parameter search covers keys, descriptions, groups, and values. Condition
search covers names, expressions, colors, and usage data. Search and
expressions compose with the other filters rather than replacing them.

## Expression language

`--expr` uses expr-lang and must evaluate to a boolean:

```sh
fbrcm get --expr 'value == true'
fbrcm get --expr 'default > 10'
fbrcm get --expr 'name startsWith "checkout_"'
fbrcm get --expr 'conditionals["Beta users"] == true'
```

Invalid expressions are errors. They never silently become an empty result.

## Parameter context

Parameter expressions expose `project_id`, `project`, `conditions`, `groups`,
`parameters`, `name`, `group`, `default`, `value`, and `conditionals`.

- `default` is only the default slot.
- `value` matches the default or any conditional slot.
- `conditionals` selects a value for one named condition.
- root-group parameters compare as `group == "(root)"` in normal parameter
  contexts.

Firebase values are typed from `valueType`: BOOLEAN becomes boolean, NUMBER
becomes numeric, and STRING and JSON remain strings. Compatibility comparisons
allow a stored boolean or number to match its string representation.

## Condition context

`conditions list --expr` exposes `name`, `priority`, `expression`, `color`,
`usage_count`, and `usages`, plus the project fields:

```sh
fbrcm conditions list example-project-id --expr 'usage_count == 0'
fbrcm conditions list example-project-id --expr 'priority <= 5'
fbrcm conditions list example-project-id \
  --expr 'any(usages, #.parameter startsWith "legacy_")'
```

## Project context

Project expressions expose `project_id`, `project`, `conditions`, `groups`,
and the complete `parameters` map:

```sh
fbrcm projects list --expr 'project_id startsWith "test-"'
fbrcm projects list --expr '"feature_enabled" in keys(parameters)'
fbrcm projects list --expr 'parameters["feature_enabled"].value == true'
```

## Helpers and JSON values

fbrcm adds `is_number`, `is_string`, `is_json`, `is_boolean`, `is_empty`, and
an embedded gojq implementation:

```sh
fbrcm get --expr 'is_string(value) && is_empty(value)'
fbrcm get --expr 'value | jq(.enabled == true)'
fbrcm get --expr '(default | jq(.limit)) > 10'
```

No external `jq` executable is required. Invalid JSON values simply do not
match a `jq(...)` predicate.
