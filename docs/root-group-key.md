# Root group key representations

Firebase Remote Config stores default (ungrouped) parameters at the top level of
the `parameters` map. Grouped parameters live under `parameterGroups`. There is
no named "root group" on the Firebase wire — the default bucket is implicit.

fbrcm uses three distinct representations depending on the layer. They are defined
in `core/rootgroup/rootgroup.go` and must not be conflated during refactors.

## Constants

| Constant | Value | Used where |
| --- | --- | --- |
| `WireKey` | `""` (empty string) | Firebase JSON, draft slot keys, `rcmutate` group field for root params |
| `TreeKey` | `"__default__"` | Parameters tree node identity, TUI navigation, internal tree maps |
| `Label` | `"(root)"` | Human-facing label in UI and filter expressions |

## Wire format (`WireKey`)

On the wire and in cached/draft JSON:

- Root parameters appear in `parameters`, not inside `parameterGroups`.
- Code passing a group key for a root parameter uses `""`.
- Draft mutations and `rcmutate.Slot.Group` use empty string for the root bucket.

Example (simplified):

```json
{
  "parameters": {
    "my_flag": { "defaultValue": { "value": "on" } }
  },
  "parameterGroups": {
    "experiments": {
      "parameters": { "exp_flag": { "defaultValue": { "value": "off" } } }
    }
  }
}
```

Here `my_flag` is in the root group (`WireKey`); `exp_flag` belongs to group
`experiments`.

## Tree and TUI (`TreeKey`)

The parameters tree (`core` `ParametersTree`, TUI parameters panel) assigns every
group a stable node key. Real groups use their Firebase group name; the root bucket
uses `TreeKey` (`__default__`) so it can be distinguished from an empty or missing
name in UI state machines.

Navigation, selection, and move/rename targets in the TUI refer to `TreeKey` for
the default bucket.

## Display and filters (`Label`)

Users see `(root)` as the group name for ungrouped parameters. Filter expressions
that match on group name accept `(root)` as the root group value (see `IsLabel`).

## Conversion rules

When translating between layers:

| From | To | Rule |
| --- | --- | --- |
| Wire / draft | Tree / TUI | Empty group → `TreeKey` |
| Tree / TUI | Wire / draft | `TreeKey` or `Label` → `WireKey` |
| Filter input | Internal | `(root)` → `WireKey` via `NormalizeGroupKey` in `core/draft` |

`IsRoot(value)` returns true for any of `WireKey`, `TreeKey`, or `Label`. Use it
when accepting user or UI input that may use any representation.

## Related code

- `core/rootgroup/rootgroup.go` — canonical constants and helpers
- `core/draft/normalize.go` — `NormalizeGroupKey` for mutation paths
- `core/rc/mutate` — slot collection keyed by group + parameter
- TUI parameters panel — renders root group with `Label`, navigates with `TreeKey`
