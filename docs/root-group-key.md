# Root group key representations

Firebase Remote Config stores ungrouped parameters in the top-level
`parameters` map. Grouped parameters live in `parameterGroups`. There is no
named root group on the Firebase wire.

fbrcm uses three representations for that implicit bucket. Their constants live
in `core/rootgroup/rootgroup.go`.

## Constants

| Constant | Value | Purpose |
| --- | --- | --- |
| `rootgroup.WireKey` | `""` | Firebase JSON, draft mutations, and Remote Config mutation slots |
| `rootgroup.TreeKey` | `"__default__"` | Stable node identity in parameter and condition trees |
| `rootgroup.Label` | `"(root)"` | Human-readable UI, diff, and expression label |

`core/rootgroup` currently contains these constants only. It does not provide
generic `IsRoot` or `IsLabel` helpers.

## Wire representation

Root parameters appear directly under `parameters`:

```json
{
  "parameters": {
    "my_flag": {
      "defaultValue": { "value": "on" }
    }
  },
  "parameterGroups": {
    "experiments": {
      "parameters": {
        "exp_flag": {
          "defaultValue": { "value": "off" }
        }
      }
    }
  }
}
```

Here `my_flag` uses `WireKey`; `exp_flag` belongs to the real group
`experiments`.

Draft mutation builders and `core/rc/mutate.Slot.Group` also use the empty
string for an ungrouped parameter.

## Tree representation

The parameter and condition trees need a stable, non-empty identity for their
synthetic root node. They use `TreeKey` (`__default__`) internally while
rendering `Label` (`(root)`) to users.

TUI selection, navigation, and transient edit state may therefore carry
`TreeKey`. They must normalize it before constructing a wire-level mutation.

## Expression and display representation

The expression environment represents the root group with an internal sentinel.
Equality treats that sentinel as equal to both `nil` and `"(root)"`:

```expr
group == nil
group == "(root)"
```

Diffs and human-readable tables render `(root)`.

The current `draft diff` and `versions diff` expression paths are an exception:
they expose an ungrouped changed parameter as the literal `"default"`. This
behavior is documented in [EXPR.md](EXPR.md#parameter-context).

## Conversions

| Boundary | Current conversion |
| --- | --- |
| Firebase JSON or mutation slot | Root stays `WireKey` (`""`) |
| Wire model to parameter tree | Root becomes `TreeKey` |
| Parameter tree to UI | `TreeKey` is rendered as `Label` |
| TUI/draft mutation input to wire | `draft.NormalizeGroupKey` converts `TreeKey` to `WireKey` |
| Expression input to wire lookup | `"(root)"` maps to `WireKey` inside `core/filter` |

`draft.NormalizeGroupKey` intentionally converts only `TreeKey`; it leaves
`Label` unchanged. Callers that accept human labels must perform that
translation at their own boundary rather than assuming the draft helper accepts
all representations.

The `core.NormalizeRemoteConfigGroupKey` facade delegates to
`draft.NormalizeGroupKey` for TUI callers.

## Related code

- `core/rootgroup/rootgroup.go`: canonical constants
- `core/parameters/tree.go`: synthetic root tree node
- `core/conditions/tree.go`: root-group condition usages
- `core/draft/normalize.go`: `TreeKey` to `WireKey`
- `core/rc/mutate/slot.go`: wire-level parameter slots
- `core/filter/expr_env.go` and `expr_compare_equal.go`: expression sentinel
- `core/rc/diff/format.go`: human-facing root label
