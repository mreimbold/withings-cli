# withings-cli

CLI for interacting with Withings Health Solutions data and OAuth tokens.

## Build

```bash
make build
```

Build output: `./withings-cli`

If you prefer the command name `withings`, rename or symlink:

```bash
ln -s ./withings-cli ./withings
```

## Quick start

```bash
./withings-cli auth login
./withings-cli measures get --type weight,bp_sys,bp_dia --start 2025-12-23 --end 2025-12-30
./withings-cli activity get --date 2025-12-29 --json
```

## Auth

Login uses OAuth with a local callback by default. It requires client
credentials from the environment:

- `WITHINGS_CLIENT_ID`
- `WITHINGS_CLIENT_SECRET`

Token refresh happens automatically when needed.

### Callback URL

Default callback URL: <http://127.0.0.1:9876/callback>

You can override the listener with:

```bash
./withings-cli auth login --listen 127.0.0.1:9876
```

### Developer Dashboard

Create your app (Client ID/Secret) in the
[Withings Developer Dashboard](https://developer.withings.com/dashboard/).


## Config and env

Precedence: flags > env > project config > user config.

Config files:
- user: `~/.config/withings-cli/config.toml`
- project: `./withings-cli.toml`

Environment:
- `WITHINGS_CLIENT_ID`
- `WITHINGS_CLIENT_SECRET`

## Commands

Core commands:
- `auth` manage tokens
- `measures` weight/BP/body metrics
- `activity` activity summaries
- `sleep` sleep summaries
- `heart` heart data
- `api` low-level escape hatch

For full CLI contract and flag details, see `docs/cli-spec.md`.

## Development

```bash
make fmt
make lint
make test
make build
```

## Credits

Inspired by Peter Steinberg's "Just talk to it" (section "What about MCPs").
Reference: `https://steipete.me/posts/just-talk-to-it#what-about-mcps`
