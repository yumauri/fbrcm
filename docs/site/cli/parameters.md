# Parameters, groups, and conditions

The parameter commands operate on selected client or server templates. With no
project filter, multi-project commands process every enabled configured target.

## Read parameters

Get one parameter across projects:

```sh
fbrcm get feature_enabled --project '^prod'
```

List parameters selected by search or an expression:

```sh
fbrcm get --search 'checkout'
fbrcm get --expr 'value == true'
```

Useful selectors include:

```text
--project, -p <query>  select template targets; repeatable
--filter, -f <query>   filter parameter keys; repeatable
--group <name>         include parameters in named groups; repeatable
--search <text>        search keys, descriptions, groups, and values
--expr <expression>    apply a typed expression
--update               revalidate caches before reading
```

When one exact project is selected, human output omits the redundant project
column. Machine output retains complete target context.

## Add and duplicate

```sh
fbrcm add checkout_enabled \
  --project '=staging' \
  --type boolean \
  --value false \
  --description 'Enable the new checkout' \
  --draft

fbrcm duplicate checkout_enabled checkout_enabled_v2 \
  --project '=staging' \
  --draft
```

Parameter types are string, boolean, number, and JSON. JSON values must be
valid JSON:

```sh
fbrcm add checkout_config \
  --type json \
  --value '{"enabled":false,"variant":"control"}' \
  --project '=staging' \
  --dry-run
```

## Update

Update a known parameter:

```sh
fbrcm update checkout_enabled \
  --project '=staging' \
  --type boolean \
  --value true \
  --dry-run
```

Omit the positional key only when filters deliberately select several
parameters:

```sh
fbrcm update \
  --project '^staging' \
  --filter '^checkout_' \
  --type boolean \
  --value false \
  --dry-run
```

Always inspect the matched-item count and diff before applying a broad update.

## Delete

```sh
fbrcm delete old_checkout_flag \
  --project '=staging' \
  --dry-run
```

Deleting a parameter does not implicitly remove an empty group. Empty and
description-only groups are preserved unless you use an explicit group delete.
Firebase-managed and unknown future values are protected from parameter
deletion.

## Manage groups

```sh
fbrcm groups list --project '=staging'
fbrcm groups add checkout --description 'Checkout controls' --project '=staging' --draft
fbrcm groups edit checkout --description 'Checkout and payment controls' --project '=staging' --draft
fbrcm groups rename checkout checkout_v2 --project '=staging' --draft
fbrcm groups delete checkout_v2 --project '=staging' --dry-run
```

Group names selected for edit, rename, or delete match exactly and
case-sensitively. Deleting a group removes all parameters inside it.

## Inspect conditions

```sh
fbrcm conditions list staging
fbrcm conditions show staging "iOS users"
fbrcm conditions validate staging
```

Conditions are shown in Firebase evaluation order. `show` includes the
parameters and conditional values that reference the selected condition.

## Change conditions

```sh
fbrcm conditions add staging "Beta users" \
  --expression 'percent <= 10' \
  --color ORANGE \
  --draft

fbrcm conditions edit staging "Beta users" \
  --expression 'percent <= 20' \
  --draft

fbrcm conditions move staging "Beta users" 1 --draft
fbrcm conditions rename staging "Beta users" "Early access" --draft
fbrcm conditions delete staging "Early access" --dry-run
```

Rename updates every conditional-value reference. Delete removes the condition
and its conditional values. Moving a condition can change value resolution, so
fbrcm treats it as a reviewed Remote Config mutation.

See [Filtering](/reference/filtering) for filter composition, expression
contexts, value typing, and JSON queries.
