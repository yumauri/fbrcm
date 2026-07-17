# Repository conventions

## Reuse and shared logic

- Before implementing parsing, filtering, sorting, terminal sizing, rendering, confirmation, or Remote Config transformations, search the repository for existing behavior and tests that already define the convention.
- Reuse existing shared and domain helpers whenever the required behavior matches. Do not create command-local copies or slightly different implementations of established behavior.
- When matching logic exists but is not accessible from the new caller, extract it to the lowest appropriate shared package and update existing callers instead of duplicating it.
- Add new local logic only when the behavior is genuinely specific to that component; keep the distinction explicit and covered by tests.

## CLI tables

- Every human-readable CLI table must use the same Lip Gloss table style as the existing `get`, `projects list`, and `cache list` commands.
- Use `lipgloss.NormalBorder()`, a header separator, padded cells, the shared colors from `cli/styles`, alternating row backgrounds, and `NoColorEnabled()` support.
- Do not implement CLI tables with `text/tabwriter`, manually padded columns, or another ad hoc renderer.
- Keep machine-readable output behind the command's JSON flag and free of terminal styling.

## CLI confirmations

- Every interactive CLI yes/no confirmation must use `cli/shared.NewConfirmation`; do not construct prompt-kit confirmations directly.
- Yes must be selected by default for every CLI confirmation. Keep that default centralized in the shared constructor.

## CLI documentation

- Update `CLI.md` in the same change whenever the CLI interface surface changes, including commands, subcommands, positional arguments, flags, defaults, output contracts, confirmations, or other user-visible behavior.
- Keep both the command tree and the detailed command sections synchronized with the implemented Cobra command structure and its tests.

## Remote Config groups

- Preserve empty and description-only parameter groups across all parameter mutations, filtering, condition cleanup, drafts, imports, merges, and promotions.
- Removing or replacing a group must be an explicit group-level operation. In the TUI, group removal must originate from the configured delete action. It opens confirmation when no draft exists and stages immediately when a draft already exists, consistent with other TUI edits.

## Validation

- After every implementation change, run the repository-wide `golangci-lint run` before handing off the result. Do not rely on tests or `go vet` alone.
