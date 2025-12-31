# Withings CLI

Withings CLI for interacting with Withings Health Solutions data and OAuth tokens.

## Highlights

- OAuth login with local callback and automatic refresh
- Measures, activity, sleep, and heart endpoints
- Output formats: tables (default), `--json`, or `--plain`
- Low-level API escape hatch for new endpoints

## Overview

Withings CLI is a small command-line tool that pulls data from the
Withings API and manages OAuth tokens. The full CLI contract lives in
[`docs/cli-spec.md`](docs/cli-spec.md).

### ✍️ Author

This project is 100% written by AI.

### Requirements

- Go 1.25.4
- Withings developer app credentials (Client ID/Secret)

### Developer Dashboard

Create your app (Client ID/Secret) in the
[Withings Developer Dashboard](https://developer.withings.com/dashboard/).

## Usage instructions

```bash
./withings-cli auth login
./withings-cli measures get --type weight,bp_sys,bp_dia --start 2025-12-23 --end 2025-12-30
./withings-cli activity get --date 2025-12-29 --json
./withings-cli sleep get --start 2025-12-01 --end 2025-12-31 --plain
./withings-cli heart get --start 2025-12-23 --end 2025-12-30
```

## Installation instructions

From source:

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

## Development

```bash
make fmt
make lint
make test
make build
```

## Credits

Inspired by Peter Steinberg's "Just talk to it" (section "What about MCPs").
Reference: <https://steipete.me/posts/just-talk-to-it#what-about-mcps>

## Feedback

If you have ideas or issues, please open a GitHub issue or start a discussion
in this repo.
