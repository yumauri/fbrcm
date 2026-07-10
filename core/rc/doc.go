// Package rc groups Remote Config domain helpers used by core, CLI, and TUI.
//
// This tree is separate from cli/shared/rc, which implements the CLI mutation
// pipeline (stdin, ordering, publish loop). Packages here are pure RC helpers:
//
//   - value — parameter value validation and JSON number checks
//   - mutate — slot collection and in-memory RC mutation
//   - diff — colored Remote Config diff rendering
//   - display — summary/diff formatting and project/condition labels
package rc
