# Themes

One theme controls human-readable CLI output, prompts, logs, and the TUI. Press
`Ctrl+T` in the TUI to preview installed themes live.

## Select and manage themes

```sh
fbrcm theme
fbrcm theme list
fbrcm theme switch nord
fbrcm theme switch nord --scope local
fbrcm theme reset --scope local
fbrcm theme path nord
```

fbrcm stores a local selection in `.fbrcm.toml`. Custom theme files live in the
user-wide `themes` directory next to global `config.toml`.

Import one file, a directory, stdin, or an HTTPS URL:

```sh
fbrcm theme import ./nord.toml
fbrcm theme import ./theme-pack/
fbrcm theme import --name nord < ./nord.toml
fbrcm theme import https://example.com/themes/nord.toml
```

fbrcm validates imports before writing. A single-file import rejects an existing
destination. A directory import skips existing names and reports warnings.

## Theme file

A file may override any subset of the built-in palette:

```toml
[colors]
primary = "#88C0D0"
selection = "#4C566A"
text = "#ECEFF4"
text_muted = "#A3ABB8"
error = "#BF616A"
```

Colors accept `#RGB`, `#RRGGBB`, or a quoted ANSI index from `"0"` through
`"255"`. Unknown top-level fields or color tokens invalidate the theme.

## Inheritance

Themes can inherit another installed theme:

```toml
inherits = "base"

[colors]
primary = "#88C0D0"
```

Resolution starts from the built-in palette, applies the oldest ancestor, then
the selected child. Missing parents, cycles, and chains deeper than 16 files
invalidate the selected theme.

## Failure behavior

Normal startup warns and falls back to the complete built-in palette when a
selected theme is invalid. Explicit configuration commands such as
`config validate` remain strict so configuration problems can be caught in CI.

`NO_COLOR` disables color and skips theme application. JSON and stateless modes
also skip startup theme loading.
