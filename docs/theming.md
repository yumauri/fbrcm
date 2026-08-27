# Theming

fbrcm uses the same theme for human-readable CLI output, prompts, logs, and the
TUI. The built-in theme is the application's original color palette.

In the TUI, press `Ctrl+T` to open the theme picker. Moving its cursor previews
each theme immediately, `Enter` saves the selection, and `Esc` restores the
previous palette. While the picker is open, editing the highlighted custom
theme and any theme it inherits from reload the preview automatically. See
[TUI: Theme picker](TUI.md#theme-picker) for keyboard, mouse, live editing, and
configuration-layer behavior.

## Selecting a theme

Theme files live in the `themes` directory next to the global `config.toml`:

```text
<config-root>/
├── config.toml
├── profiles/
└── themes/
    ├── base.toml
    └── nord.toml
```

`fbrcm theme path <name>` prints a theme's canonical destination. Normal
startup never creates the themes directory. `fbrcm theme import` creates it
only after a supplied theme passes validation.

The filename without `.toml` is the theme name. Select it through global or
repository configuration:

```toml
theme = "nord"
```

The equivalent management command is:

```sh
fbrcm theme switch nord
```

Use `fbrcm theme switch nord --scope local` to store the selection in the nearest `.fbrcm.toml`. A
local selection still loads its file from the user-wide `themes` directory.
Local configuration overrides the global selection. Remove an override with:

```sh
fbrcm config reset theme --scope local
```

## Managing themes

The singular `theme` command mirrors profile management:

```sh
fbrcm theme                 # print the effective name, or built-in
fbrcm theme list
fbrcm theme switch built-in
fbrcm theme reset --scope local
fbrcm theme path nord
fbrcm theme rename nord nord-dark
fbrcm theme delete nord-dark
fbrcm theme import ./nord.toml
fbrcm theme import ./nord.toml --name nord-dark
fbrcm theme import ./theme-pack/
fbrcm theme import --name nord < ./nord.toml
fbrcm theme import https://example.com/themes/nord.toml
```

In color-enabled human output, `theme` prints an eight-swatch preview beside the
effective name, and `theme list` adds the same preview for every installed
theme. The swatches represent primary, selection, secondary, highlight, text,
muted text, success, and error colors. Narrow tables reduce the swatch count as
needed. `NO_COLOR` suppresses previews, and JSON output never includes them.

With no import source or redirected stdin, human mode opens an interactive
`.toml` file selector. The imported filename determines the theme name unless
`--name` overrides it. A directory source imports its top-level regular `.toml`
files as one batch; `--name` is unavailable because filenames provide the theme
names. Stream stdin requires `--name`; a positional source wins over stdin
without consuming it. Each file is limited to 1 MiB and validated before any
write. A single-theme import rejects an existing destination. A batch skips
existing themes without replacing them and logs one warning for each skip.

Inheritance within a directory batch is supported. On supported systems, human
mode also accepts a directory as stdin with the same batch and collision
behavior. Directory stdin is experimental and intentionally absent from the
versioned machine schema and capability metadata; a positional directory is a
supported input in both human and JSON modes.

Rename updates direct inheritance references and selections in global and the
currently discovered local configuration. Delete rejects selected themes and
themes that another installed theme directly inherits.

`built-in` is a reserved selector and cannot be used as a theme filename. Both
`theme switch built-in` and `theme reset` remove the selection from the chosen
configuration layer; they do not create a theme file.

An absent selection uses the built-in theme and does not read or create the
`themes` directory. `NO_COLOR`, JSON output, and `--stateless` skip startup
theme loading and application. `NO_COLOR` remains authoritative even when a
theme is configured. Explicit `config` operations may still inspect theme files
to validate the configuration they show or modify.

## Theme files

A theme may override any subset of the built-in colors:

```toml
[colors]
primary = "#88C0D0"
selection = "#4C566A"
text = "#ECEFF4"
text_muted = "#A3ABB8"
error = "#BF616A"
```

Colors accept `#RGB`, `#RRGGBB`, or a quoted ANSI index from `"0"` through
`"255"`. Keys and top-level fields are strict: unknown names invalidate the
selected theme.

Theme files must be regular files rather than symbolic links. Names are exact
and case-sensitive even on a case-insensitive filesystem. They must be one
filesystem-safe path segment without leading or trailing whitespace, and must
not use the reserved name `built-in`.

## Inheritance

One theme can inherit another theme from the same directory:

```toml
inherits = "base"

[colors]
primary = "#88C0D0"
error = "#BF616A"
```

Resolution starts with the built-in palette, applies the oldest ancestor, and
finishes with the selected child. Only the selected theme and its ancestors are
read; an unrelated invalid file does not affect the application. Missing
parents, inheritance cycles, and chains deeper than 16 files invalidate the
whole selected theme.

## Complete built-in palette

This file reproduces the built-in theme and documents every supported token:

```toml
[colors]
url = "117"

primary = "#8FA8C7"
selection = "#556B84"
secondary = "#C8A27E"
highlight = "#D8C6A0"
highlight_muted = "#BFA77A"
text = "#D7D9DE"
text_soft = "#BEC3CC"
text_muted = "#959CA8"
text_dim = "#5A6270"
error = "#C58A8A"
success = "#7FD38B"
row_stripe = "#121417"

condition_blue = "#8FA8C7"
condition_brown = "#C8A27E"
condition_cyan = "#61D6E8"
condition_deep_orange = "#FF8A5B"
condition_green = "#7FD38B"
condition_indigo = "#8AA2FF"
condition_lime = "#C1D96F"
condition_orange = "#C8A27E"
condition_pink = "#F38DB5"
condition_purple = "#B58CFF"
condition_teal = "#58D1C9"

diff_added = "42"
diff_removed = "203"
diff_changed = "221"
diff_note = "245"

history_added_background = "#315A46"
history_removed_background = "#68434A"
history_changed_background = "#665A38"
inactive_selection = "#343A43"

offline_foreground = "15"
offline_background = "196"

log_debug = "63"
log_info = "86"
log_warn = "192"
log_error = "204"
log_fatal = "134"
log_silent = "245"
log_default = "255"

logo_start = "#FFC400"
logo_middle = "#FF9100"
logo_end = "#DD2C00"
```

## Errors and validation

Normal human startup never fails because of a theme. A missing directory or
file, malformed TOML, invalid color, unknown token, unreadable file, missing
parent, or inheritance cycle produces one warning and applies the complete
built-in palette. A partially resolved theme is never used.

`fbrcm config set theme`, `fbrcm config edit`, and `fbrcm config validate` are
strict and report an invalid selected theme as a configuration error. Normal
startup remains available, while explicit configuration workflows still catch
mistakes. Fix the file or remove the selection to restore successful
validation.
