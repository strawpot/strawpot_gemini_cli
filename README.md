# StrawPot Gemini CLI

A Go wrapper that translates [StrawPot](https://github.com/strawpot) protocol arguments into [Google Gemini CLI](https://github.com/google-gemini/gemini-cli) flags. It acts as a pure translation layer — process management, sessions, and infrastructure are handled by StrawPot core.

## Prerequisites

- [Gemini CLI](https://www.npmjs.com/package/@google/gemini-cli) (`npm install -g @google/gemini-cli`)
- A Gemini API key (or a Google account for browser-based login)

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_gemini_cli/main/strawpot_gemini/install.sh | sh
```

This downloads a pre-built binary for your platform (macOS/Linux, amd64/arm64) to `/usr/local/bin`. Override the install directory with `INSTALL_DIR`:

```sh
INSTALL_DIR=~/.local/bin curl -fsSL ... | sh
```

## Usage

The wrapper exposes two subcommands:

### `setup`

Runs `gemini auth login` to authenticate with Google.

```sh
strawpot_gemini setup
```

### `build`

Translates StrawPot protocol flags into a Gemini CLI command and outputs it as JSON.

```sh
strawpot_gemini build \
  --agent-workspace-dir /path/to/workspace \
  --working-dir /path/to/project \
  --task "fix the bug" \
  --config '{"model":"gemini-2.5-pro"}'
```

Output:

```json
{
  "cmd": ["gemini", "-p", "fix the bug", "-m", "gemini-2.5-pro", "--yolo", "--include-directories", "/path/to/workspace"],
  "cwd": "/path/to/project"
}
```

#### Build flags

| Flag | Required | Description |
|---|---|---|
| `--agent-workspace-dir` | Yes | Workspace directory (used as `--include-directories`) |
| `--working-dir` | No | Working directory for the command (`cwd` in output) |
| `--task` | No | Task prompt (passed as `gemini -p`) |
| `--config` | No | JSON config object (default: `{}`) |
| `--role-prompt` | No | Role prompt text (written to `GEMINI.md`) |
| `--memory-prompt` | No | Memory/context prompt (appended to `GEMINI.md`) |
| `--skills-dir` | No | Directory with skill subdirectories (symlinked to `skills/`) |
| `--roles-dir` | No | Directory with role subdirectories (repeatable, symlinked to `roles/`) |
| `--agent-id` | No | Agent identifier |

## Configuration

### Config JSON

Pass via `--config`:

| Key | Type | Default | Description |
|---|---|---|---|
| `model` | string | `gemini-2.5-pro` | Model to use |
| `dangerously_skip_permissions` | boolean | `true` | Auto-approve all tool calls (`--yolo`). Set to `false` to require approval. |

### Environment variables

| Variable | Description |
|---|---|
| `GEMINI_API_KEY` | Gemini API key (optional if logged in with Google account) |
| `SANDBOX_MODE` | When set, enables `--sandbox` for sandboxed execution |

## Development

```sh
cd gemini/wrapper
go test -v ./...
```

Releases are built with [GoReleaser](https://goreleaser.com/) and published automatically via GitHub Actions.

## License

See [LICENSE](LICENSE) for details.
