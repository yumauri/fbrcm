# Expression Filtering

`fbrcm` supports `--expr` on several commands for advanced filtering. Expressions are powered by [expr-lang](https://expr-lang.org/docs/language-definition), with extra `fbrcm` context and helper functions for Firebase Remote Config.

Use expressions when name filters are not enough:

```sh
fbrcm get --expr 'value == true'
fbrcm get --expr 'value > 10'
fbrcm get --expr 'value | jq(.enabled == true)'
```

## Language Overview

Expr-lang is a small expression language. `--expr` must evaluate to a boolean.

Common operators:

```expr
name == "some_flag"
project_id != "test-project"
default > 10
value == true || default == false
name in ["param_a", "param_b"]
name contains "some"
```

Boolean logic:

```expr
value == true
value >= 10 && value <= 20
value | jq(.enabled == true)
```

Built-in functions and collection helpers include `any`, `all`, `none`, `filter`, `map`, `len`, `keys`, `values`, `type`, `int`, `float`, and `string`. See full syntax and built-ins in the expr-lang documentation:

https://expr-lang.org/docs/language-definition

Expr also supports a pipe operator:

```expr
value | jq(.enabled == true)
```

This calls `jq(value, ".enabled == true")`.

## Commands With `--expr`

`--expr` is supported by these commands:

```sh
fbrcm get --expr '...'
fbrcm delete --expr '...'
fbrcm update --expr '...'
fbrcm project import --expr '...'
fbrcm projects list --expr '...'
fbrcm projects update --expr '...'
fbrcm add --expr '...'
```

Commands use one of two expression contexts: parameter context or project context.

## Parameter Context

Parameter context is used by:

```sh
fbrcm get --expr '...'
fbrcm delete --expr '...'
fbrcm update --expr '...'
fbrcm project import --expr '...'
```

In this context, the expression is evaluated once per parameter. A matching expression keeps the parameter for display, deletion, update, or import.

Available fields:

| Field | Meaning |
| --- | --- |
| `project_id` | Firebase project id. |
| `project` | Firebase project display name. |
| `conditions` | Sorted list of condition names in the config. |
| `groups` | Sorted list of parameter group names in the config. |
| `parameters` | Map of all parameters by key. Each value has `group`, `default`, `value`, and `conditionals`. |
| `name` | Current parameter key. |
| `group` | Current parameter group name, or root group. Root group compares equal to `nil` and `"(root)"`. |
| `default` | Current parameter default value only. Typed from Firebase `valueType`. |
| `value` | Any current parameter value: default OR any conditional value. Typed from Firebase `valueType`. |
| `conditionals` | Map of current parameter conditional values by condition name. Values are typed from Firebase `valueType`. |

Examples:

```sh
fbrcm get --expr 'name == "some_flag"'
fbrcm get --expr 'group == "(root)"'
fbrcm get --expr 'value == true'
fbrcm get --expr 'value > 5'
fbrcm get --expr 'is_string(value) && is_empty(value)'
fbrcm get --expr 'any(conditions, # == "cond_a")'
fbrcm get --expr 'parameters["some_flag"].value == true'
```

`default` is exact default value:

```sh
fbrcm get --expr 'default == true'
fbrcm get --expr 'default > 10'
```

`value` matches default or any conditional:

```sh
fbrcm get --expr 'value == true'
fbrcm get --expr 'value > 10'
```

Exact conditional values:

```sh
fbrcm get --expr 'conditionals["cond_a"] == true'
fbrcm get --expr 'conditionals["cond_b"] > 50'
```

## Project Context

Project context is used by:

```sh
fbrcm projects list --expr '...'
fbrcm projects update --expr '...'
fbrcm add --expr '...'
```

The expression is evaluated once per project. The command loads that project's Remote Config so project filters can inspect parameters too.

Available fields:

| Field | Meaning |
| --- | --- |
| `project_id` | Firebase project id. |
| `project` | Firebase project display name. |
| `conditions` | Sorted list of condition names in the project config. |
| `groups` | Sorted list of parameter group names in the project config. |
| `parameters` | Map of all parameters by key. Each value has `group`, `default`, `value`, and `conditionals`. |

Parameter-specific fields like `name`, `group`, `default`, `value`, and `conditionals` are empty in project context. Use the `parameters` map instead.

Examples:

```sh
fbrcm projects list --expr 'project_id startsWith "test-"'
fbrcm projects list --expr '"some_flag" in keys(parameters)'
fbrcm projects list --expr 'parameters["some_flag"].value == true'
fbrcm add new_param --boolean true --expr 'parameters["old_param"].value == true'
```

## Value Typing

Remote Config values are stored as strings by Firebase, but `fbrcm` converts expression values by `valueType`:

| Firebase `valueType` | Expr runtime type |
| --- | --- |
| `BOOLEAN` | `bool` |
| `NUMBER` | `float` |
| `STRING` | `string` |
| `JSON` | `string`, parseable by `jq(...)` |

Equality is compatibility-friendly:

```expr
value == true
value == "true"
value == 42
value == "42"
```

All of these can match the same boolean or number parameter.

Numeric comparisons ignore wrong-type values and return `false` instead of logging an evaluation error:

```expr
value > 10
default <= 42
conditionals["cond_a"] >= 25
```

String operators also ignore wrong-type values and return `false`. With `value`, they check default OR any conditional string value:

```expr
value contains "example.com"
value startsWith "https://"
value endsWith ".json"
value matches "^https://"
```

String functions such as `lower(value)` or `trim(value)` do not work with `value`, because `value` is an any-value wrapper, not a single string. Use the string operators above for simple matching, or target an exact scalar value such as `default` or `conditionals["cond_a"]` when calling string functions.

## `default`, `value`, and `conditionals`

`default` checks only the default value:

```expr
default == true
default > 10
```

`value` checks default OR any conditional value:

```expr
value == true
value > 10
is_empty(value)
value | jq(.enabled == true)
```

`conditionals` checks specific conditional values:

```expr
conditionals["cond_a"] == true
(conditionals["cond_b"] | jq(.enabled == true))
```

## Custom Helper Functions

`fbrcm` adds these helper functions:

```expr
is_number(value)
is_string(value)
is_json(value)
is_boolean(value)
is_empty(value)
jq(value, ".enabled == true")
```

Type helpers work best with `value`, because `value` carries Firebase `valueType` metadata:

```expr
is_number(value)
is_string(value)
is_json(value)
is_boolean(value)
```

They also work with scalar values like `default`, but `STRING` and `JSON` are both strings at runtime, so exact distinction is only available through `value`.

`is_empty` works with any value:

```expr
is_string(value) && is_empty(value)
is_json(value) && !is_empty(value)
```

For `value`, `is_empty(value)` is true if default or any conditional value is empty.

## JSON Filtering With `jq`

`fbrcm` embeds [gojq](https://github.com/itchyny/gojq), a pure-Go jq implementation. It does not require the external `jq` binary.

Use `jq` with quoted jq code:

```sh
fbrcm get --expr 'value | jq(".enabled == true")'
```

Or use the shorthand. `fbrcm` prepares the expression and wraps the `jq(...)` body automatically:

```sh
fbrcm get --expr 'value | jq(.enabled == true)'
```

Both forms are equivalent.

`value` means default OR any conditional JSON value:

```sh
fbrcm get --expr 'value | jq(.enabled == true)'
```

Exact default JSON:

```sh
fbrcm get --expr 'default | jq(.enabled == true)'
```

Exact conditional JSON:

```sh
fbrcm get --expr 'conditionals["cond_b"] | jq(.enabled == true)'
```

Extract and compare jq results:

```sh
fbrcm get --expr '(value | jq(.limit)) > 10'
fbrcm get --expr 'value | jq(.features.some_item == true)'
fbrcm get --expr 'value | jq(.items | length > 0)'
```

If a jq expression returns booleans, `fbrcm` treats any `true` result as a match. If it returns values, those values can be compared with expr operators.

Invalid JSON values are ignored by `jq(...)`.

## Command Examples

Show boolean params enabled anywhere:

```sh
fbrcm get --expr 'value == true'
```

Show numeric params with any value above 10:

```sh
fbrcm get --expr 'value > 10'
```

Show JSON params where any value has `enabled: true`:

```sh
fbrcm get --expr 'value | jq(.enabled == true)'
```

Delete only empty string params:

```sh
fbrcm delete --expr 'is_string(value) && is_empty(value)' --dry-run
```

Important: `value` means default OR any conditional value, so this deletes a string parameter if any value slot is empty.

Delete string params only when default and all conditional values are empty:

```sh
fbrcm delete --expr 'is_string(value) && is_empty(default) && all(values(conditionals), is_empty(#))' --dry-run
```

Update only params whose default is false:

```sh
fbrcm update --boolean true --expr 'default == false' --dry-run
```

Update a specific string parameter only where it contains a value:

```sh
fbrcm update some_text --string "new value" --expr 'value contains "old"' --dry-run
```

Import only JSON params whose imported value is enabled:

```sh
fbrcm project import test-project --from remote-config.json --expr 'value | jq(.enabled == true)' --dry-run
```

List projects where a parameter exists and is enabled:

```sh
fbrcm projects list --expr '"some_flag" in keys(parameters) && parameters["some_flag"].value == true'
```

Add a parameter only to projects with a group:

```sh
fbrcm add new_flag --boolean false --expr '"group_a" in groups' --dry-run
```

Find parameters with redundant conditional values:

```sh
fbrcm get --expr 'len(conditionals) > 0 && all(values(conditionals), # == default)'
```

This finds parameters where every conditional value is the same as the default value. In that state the conditional value entries are redundant, because no condition changes the effective value.

To inspect only one project:

```sh
fbrcm get --project '=test-project' --expr 'len(conditionals) > 0 && all(values(conditionals), # == default)'
```

To clean them up, remove all conditional values from those matched parameters:

```sh
fbrcm update --remove-all-conditional-values --expr 'len(conditionals) > 0 && all(values(conditionals), # == default)' --dry-run
```

To remove only specific conditional values, repeat `--remove-conditional-value`:

```sh
fbrcm update --remove-conditional-value cond_a --remove-conditional-value cond_b --expr 'conditionals["cond_a"] == default || conditionals["cond_b"] == default' --dry-run
```
