# Blueprints TUI

Experimental terminal UI for browsing [Sanity Blueprints](https://www.sanity.io/docs/blueprints) stacks, resources, operations, and logs.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) v2.

> **Status:** This is an internal experiment with limited functionality. Read-only — no deploy or mutation support.

## Usage

```
go run . --project <project-id> [flags]
```

### Flags

| Flag | Env var | Description |
|---|---|---|
| `--project` | `SANITY_PROJECT_ID` | Sanity project ID (required) |
| `--token` | `SANITY_AUTH_TOKEN` | API auth token (falls back to `~/.config/sanity/config.json`) |
| `--debug` | | Write debug output to `debug.log` |
<!--| `--api-url` | `BLUEPRINTS_API_URL` | Override the API base URL |
| `--staging` | | Use the staging environment (`sanity.work`) |-->

### Navigation

| Key | Action |
|---|---|
| `enter` | Select stack / resource / operation |
| `esc` | Go back |
| `tab` / `shift+tab` | Switch tabs (detail view) |
| `/` | Filter list |
| `r` | Refresh |
| `q` | Quit |
