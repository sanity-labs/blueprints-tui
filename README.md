# Blueprints TUI

Experimental terminal UI for browsing [Sanity Blueprints](https://www.sanity.io/docs/blueprints) stacks, resources, operations, and logs.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) v2.

> **Status:** This is an experiment with limited functionality. Read-only — no deploy or mutation support.

## Usage

```
go run . [flags]
```

When no scope flag is provided, an interactive picker lists your organizations and their projects. Select an organization or project to set the session scope, then browse stacks within that scope.

### Flags

| Flag | Env var | Description |
|---|---|---|
| `--org` | `SANITY_ORG_ID` | Sanity organization ID |
| `--project` | `SANITY_PROJECT_ID` | Sanity project ID |
| `--token` | `SANITY_AUTH_TOKEN` | API auth token (falls back to `~/.config/sanity/config.json`) |
| `--debug` | | Write debug output to `debug.log` |
<!--
| `--api-url` | `BLUEPRINTS_API_URL` | Override the API base URL |
| `--staging` | | Use the staging environment (`sanity.work`) |
-->

`--org` and `--project` are mutually exclusive. If either is provided the scope picker is skipped. If neither is set, the picker is shown on startup.

### Navigation

| Key | Action |
|---|---|
| `enter` | Select scope / stack / resource / operation |
| `esc` | Go back (exit scope, return to parent view) |
| `tab` / `shift+tab` | Switch tabs (detail view) |
| `/` | Filter list |
| `r` | Refresh |
| `q` | Quit |
