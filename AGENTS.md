# Repository conventions

## CLI tables

- Every human-readable CLI table must use the same Lip Gloss table style as the existing `get`, `projects list`, and `cache list` commands.
- Use `lipgloss.NormalBorder()`, a header separator, padded cells, the shared colors from `cli/styles`, alternating row backgrounds, and `NoColorEnabled()` support.
- Do not implement CLI tables with `text/tabwriter`, manually padded columns, or another ad hoc renderer.
- Keep machine-readable output behind the command's JSON flag and free of terminal styling.

## CLI confirmations

- Every interactive CLI yes/no confirmation must use `cli/shared.NewConfirmation`; do not construct prompt-kit confirmations directly.
- Yes must be selected by default for every CLI confirmation. Keep that default centralized in the shared constructor.

## Remote Config groups

- Preserve empty and description-only parameter groups across all parameter mutations, filtering, condition cleanup, drafts, imports, merges, and promotions.
- Removing or replacing a group must be an explicit group-level operation. In the TUI, group removal must originate from the configured delete action. It opens confirmation when no draft exists and stages immediately when a draft already exists, consistent with other TUI edits.
