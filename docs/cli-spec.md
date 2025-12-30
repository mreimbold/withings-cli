# Withings CLI Spec

## Name + One-liner
- Name: `withings`
- One-liner: Interact with Withings Health Solutions data and OAuth tokens from the CLI.

## Usage
```
withings [global flags] <subcommand> [args]
```

## Subcommands
- `withings auth ...` manage OAuth tokens
- `withings user ...` user/profile lookups
- `withings measures ...` weight/BP/body metrics
- `withings activity ...` activity summaries
- `withings sleep ...` sleep summaries
- `withings heart ...` heart data
- `withings api ...` low-level action-based requests (escape hatch)

## Global flags
- `-h, --help` show help and exit
- `--version` print version to stdout
- `-v, --verbose` increase diagnostic verbosity (repeatable: `-v/-vv`)
- `-q, --quiet` suppress non-error output
- `--json` machine-readable JSON output
- `--plain` stable line-based output (no tables, no colors)
- `--no-color` disable ANSI color
- `--no-input` disable prompts; fail if required input is missing
- `--config <path>` override config file path
- `--cloud <eu|us>` select API cloud (default `eu`)
- `--base-url <url>` override API base URL (advanced)

## I/O contract
- stdout: primary results (human or `--json`/`--plain`)
- stderr: errors, warnings, progress, diagnostics
- prompts only when stdin is a TTY and `--no-input` is not set
- `--json` outputs an envelope: `{ "ok": true|false, "data": ..., "meta": ... }`

## Exit codes
- `0` success
- `1` generic failure
- `2` invalid usage/flags
- `3` auth required or refresh failed
- `4` network/connectivity error
- `5` API error (non-2xx or Withings error code)

## Config / env / precedence
- precedence: flags > env > project config > user config > system
- user config: `~/.config/withings-cli/config.toml`
- project config (optional): `./withings-cli.toml`
- env vars:
  - `WITHINGS_CLIENT_ID`
  - `WITHINGS_CLIENT_SECRET` (secret; prefer env or prompt)
  - `WITHINGS_REDIRECT_URI`
  - `WITHINGS_ACCESS_TOKEN` (optional override)
  - `WITHINGS_REFRESH_TOKEN` (optional override)
  - `WITHINGS_CLOUD` (`eu` or `us`)
  - `WITHINGS_BASE_URL` (advanced override)
- client credentials are read from env only; the CLI does not store them in config files

## Auth commands
- `withings auth login`
  - performs browser OAuth with local callback server by default
  - requires `WITHINGS_CLIENT_ID` and `WITHINGS_CLIENT_SECRET`
  - exchanges the authorization code and stores tokens automatically
  - flags: `--redirect-uri <uri>`, `--no-open`, `--listen <addr:port>`
- `withings auth status` show token age/scopes/expiry
- `withings auth logout` delete stored tokens (requires confirmation or `--force`)
- access tokens are refreshed automatically when expired (requires `WITHINGS_CLIENT_ID` and `WITHINGS_CLIENT_SECRET`)

## Data commands (common flags)
- common flags: `--start <rfc3339|epoch>`, `--end <rfc3339|epoch>`, `--last-update <epoch>`, `--limit <n>`, `--offset <n>`, `--user-id <id>`
- output: tables by default; `--json` returns raw API `body`

### measures
- `withings measures get`
  - flags: `--type <list>` (e.g., `weight,bp_sys,bp_dia,fat_mass`), `--category <real|goal>`
  - behavior: idempotent, read-only

### activity
- `withings activity get`
  - flags: `--date <YYYY-MM-DD>`, `--start/--end` for range

### sleep
- `withings sleep get`
  - flags: `--date`, `--start/--end`, `--model <1|2>` (if supported)

### heart
- `withings heart get`
  - flags: `--start/--end`, `--signal` (include signal metadata if available)

### user
- `withings user me` show current user profile
- `withings user list` list linked users (if supported)

## API escape hatch
- `withings api call --service <service> --action <action> --params <json>`
  - `--params` accepts a JSON object; use `@file.json` or `-` for stdin
  - `--dry-run` prints request URL/body without executing
  - use `--json` for raw response passthrough

## Safety rules
- `auth logout` requires confirmation unless `--force`
- prompts only when TTY and `--no-input` is not set
- `api call` supports `--dry-run` and warns on likely non-idempotent actions

## Examples
```bash
withings auth login
withings auth status
withings measures get --type weight,bp_sys,bp_dia --start 2025-12-23 --end 2025-12-30
withings activity get --date 2025-12-29 --json
withings sleep get --start 2025-12-01 --end 2025-12-31 --plain
withings api call --service measure --action getmeas --params @params.json --json
```
