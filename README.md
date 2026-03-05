# strawpot_gemini_cli

StrawPot wrapper for [Google Gemini CLI](https://github.com/google-gemini/gemini-cli). Translates StrawPot's agent protocol into Gemini CLI flags.

## Overview

This wrapper provides two subcommands:

- **`setup`** — Runs `gemini auth login` for interactive authentication
- **`build`** — Translates StrawPot protocol args to a Gemini CLI command, returning JSON: `{"cmd": [...], "cwd": "..."}`

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_gemini_cli/main/strawpot_gemini/install.sh | sh
```

## Development

```sh
cd gemini/wrapper
go test -v ./...
go build -o strawpot_gemini .
```

## License

MIT
