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
Time                       Type     Value   Unit  Category
2025-12-28T07:44:12+01:00  weight   82.450  kg    real
2025-12-28T07:44:12+01:00  bp_sys   118     mmHg  real
2025-12-28T07:44:12+01:00  bp_dia   76      mmHg  real
2025-12-21T06:58:03+01:00  weight   82.120  kg    real
2025-12-21T06:58:03+01:00  bp_sys   120     mmHg  real
2025-12-21T06:58:03+01:00  bp_dia   78      mmHg  real

$ ./withings-cli sleep get --start 2025-12-01 --end 2025-12-07
Start                      End                        Duration  Score  Wakeups  Model
2025-12-06T23:14:00+01:00  2025-12-07T06:41:00+01:00  26820     84     2        2
2025-12-05T23:32:00+01:00  2025-12-06T06:28:00+01:00  24960     78     3        2
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
