# Withings CLI

Pull your Withings health data (weight, blood pressure, sleep, activity) from the command line. Handles OAuth automatically—just run `auth login` once and start querying.

## Quick Start

```bash
# 1. Build
make build

# 2. Set credentials (from developer.withings.com/dashboard)
export WITHINGS_CLIENT_ID=your_client_id
export WITHINGS_CLIENT_SECRET=your_client_secret

# 3. Login (opens browser, completes OAuth)
./withings-cli auth login

# 4. Fetch your data
./withings-cli measures get --type weight --start 2025-12-01
```

## What You Get

```
$ ./withings-cli measures get --type weight,bp_sys,bp_dia --start 2025-12-01
┌────────────┬────────┬────────┬────────┐
│ Date       │ Weight │ BP Sys │ BP Dia │
├────────────┼────────┼────────┼────────┤
│ 2025-12-15 │ 72.3kg │ 120    │ 80     │
│ 2025-12-08 │ 72.1kg │ 118    │ 78     │
└────────────┴────────┴────────┴────────┘

$ ./withings-cli sleep get --start 2025-12-01 --end 2025-12-07
┌────────────┬──────────┬───────────┬─────────────┐
│ Date       │ Duration │ Deep      │ REM         │
├────────────┼──────────┼───────────┼─────────────┤
│ 2025-12-07 │ 7h 23m   │ 1h 45m    │ 1h 52m      │
│ 2025-12-06 │ 6h 58m   │ 1h 32m    │ 1h 41m      │
└────────────┴──────────┴───────────┴─────────────┘
```

Output formats: tables (default), `--json`, or `--plain`.

## Highlights

- OAuth login with local callback and automatic refresh
- Measures, activity, sleep, and heart endpoints
- Output formats: tables (default), `--json`, or `--plain`
- Low-level API escape hatch for new endpoints

## Requirements

- Go 1.25.4
- Withings developer app credentials ([create them here](https://developer.withings.com/dashboard/))

## Installation

```bash
make build
```

Build output: `./withings-cli`

If you prefer the command name `withings`, rename or symlink:

```bash
ln -s ./withings-cli ./withings
```

## Configuration

Precedence: flags > project config > user config.

Config files:
- user: `~/.config/withings-cli/config.toml`
- project: `./withings-cli.toml`

Environment:
- `WITHINGS_CLIENT_ID`
- `WITHINGS_CLIENT_SECRET`

### Callback URL

Default callback URL: <http://127.0.0.1:9876/callback>

You can override the listener with:

```bash
./withings-cli auth login --listen 127.0.0.1:9876
```

## Commands

Core commands:
- `auth` manage tokens
- `measures` weight/BP/body metrics
- `activity` activity summaries
- `sleep` sleep summaries
- `heart` heart data
- `api` low-level escape hatch

Full CLI specification: [`docs/cli-spec.md`](docs/cli-spec.md)

## Development

```bash
make fmt
make lint
make test
make build
```

## Credits

This project is 100% written by AI.

Inspired by Peter Steinberg's "Just talk to it" (section "What about MCPs").
Reference: <https://steipete.me/posts/just-talk-to-it#what-about-mcps>

## Feedback

If you have ideas or issues, please open a GitHub issue or start a discussion
in this repo.
