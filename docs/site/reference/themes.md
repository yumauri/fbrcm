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

Import one file, a directory, stdin, a raw HTTPS URL, or a repository browser URL:

```sh
fbrcm theme import ./nord.toml
fbrcm theme import ./theme-pack/
fbrcm theme import --name nord < ./nord.toml
fbrcm theme import https://example.com/themes/nord.toml
fbrcm theme import https://github.com/yumauri/fbrcm/tree/main/themes
fbrcm theme import https://github.com/yumauri/fbrcm/blob/main/themes/aurora.toml
fbrcm theme import https://gitlab.com/company/team/theme-pack/-/tree/main/themes
```

fbrcm validates imports before writing. A single-file import rejects an existing
destination. A directory import skips existing names and reports warnings.

GitHub.com and GitLab.com browser links support branches, tags, and commit SHAs.
Repository directories import top-level regular `.toml` files, with all files
read from one resolved commit. Symlinks, submodules, and nested directories are
skipped. If a ref containing slashes is ambiguous, use `--ref feature/colors`
to identify the exact revision prefix already in the URL.

Private GitHub repositories use `GH_TOKEN`, falling back to `GITHUB_TOKEN`, with
Contents read permission. Private GitLab repositories use `GITLAB_TOKEN` with
`read_api` access. Tokens are never persisted or passed to generic HTTP sources.
Repository browser imports support the two public service hosts; self-hosted
instances and automatic `gh`/`glab` authentication are not supported.

Each file is limited to 1 MiB. Repository batches select at most 256 themes and
download at most 16 MiB of theme data; existing themes are skipped before their
contents are downloaded. Download and validation failures install no themes.
Offline mode suppresses network imports. JSON results include the requested ref
and resolved commit in `repository_source`. See [CLI reference](./commands.md)
for URL syntax, source errors, and all resource limits.

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
