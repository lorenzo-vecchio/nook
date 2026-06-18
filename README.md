# Nook

Workspace organizer CLI for developers. Launch all your project tools with a single command.

## Install

### macOS / Linux

```bash
brew tap lorenzo-vecchio/nook https://github.com/lorenzo-vecchio/nook
brew install --formula lorenzo-vecchio/nook/nook
```

### Windows (Scoop)

```powershell
scoop bucket add nook https://github.com/lorenzo-vecchio/nook
scoop install nook
```

### Go

```bash
go install github.com/lorenzo-vecchio/nook@latest
```

## Quick Start

```bash
nook detect          # check which tools are available
nook init            # create your first workspace
nook open            # launch it
```

## Commands

| Command | Description |
|---------|-------------|
| `nook init` | Create a workspace interactively |
| `nook open [name] [--env]` | Open a workspace and launch all its services |
| `nook list` | List all workspaces grouped by scan path |
| `nook edit [name]` | Open a workspace config in `$EDITOR` |
| `nook delete [name]` | Delete a workspace |
| `nook scan` | Rescan all scan paths for workspaces |
| `nook detect` | Show which provider tools are installed |

### `nook init`

Interactive walkthrough. Prompts for a name, description, environments (dev, staging, etc.), and which services to configure. Each service asks for provider-specific fields — VS Code folder and terminals, DBeaver connection string, Chrome URLs, Docker Compose file, or custom shell commands. Validates the config before saving.

### `nook open`

Opens a workspace. If a name is provided as an argument, opens it directly. Otherwise shows an interactive selector. If the workspace has multiple environments and no `--env` flag is passed, prompts for which environment to use. Services launch in parallel — failures in one don't block the others.

### `nook list`

Lists all workspaces found across every scan path. Grouped by source directory.

### `nook edit`

Opens the workspace's `workspace.yaml` in `$EDITOR`. Falls back to `vim` on macOS/Linux and `notepad` on Windows.

### `nook delete`

Deletes a hand-created workspace directory or removes a repo workspace from scan paths. Confirms before deleting.

### `nook scan`

Rescans every path in `scan_paths`. Removes entries for directories that no longer exist. Useful after adding or removing projects on disk.

### `nook detect`

Scans the system for installed provider tools (VS Code, DBeaver, Chrome, Docker) and prints a detection table with status for each.

## How workspaces work

### Two kinds of workspaces

**Hand-created workspaces** live under `~/.nook/workspaces/<name>/`. Created via `nook init`. The `workspace.yaml` is owned by nook and not meant to be committed.

**Repo workspaces** live in your project directories. You place a `workspace.yaml` at the root of a project repo and commit it. Teammates get the same launch config.

### Scan paths

The global config at `~/.nook/config.yaml` contains a `scan_paths` list — directories nook looks through to find workspaces. Defaults to `~/.nook/workspaces` (for hand-created ones). Paths are added automatically when you trust a detected workspace.

### Auto-detection and trust

Every time you run `nook open` or `nook list`, nook checks the current directory and its immediate subdirectories for `workspace.yaml` files. If it finds one in a directory that's not already in `scan_paths`, it prompts:

```
? Found workspace.yaml in ~/projects/team-repo/
  Workspace: "team-repo" — A fullstack app
  Do you want to trust this workspace? (Y/n)
```

Answering yes adds the directory to `scan_paths` permanently. On subsequent runs there's no prompt — it's already trusted. If the directory is later removed, `nook scan` cleans up stale entries.

## Workspace config

Workspaces are defined in `workspace.yaml`:

```yaml
name: my-app
description: A fullstack app with React + Go

environments:
  dev:
    env_file: .env.dev
    services:
      - provider: vscode
        folder: ./frontend
      - provider: vscode
        folder: ./backend
        terminals:
          - name: Backend
            directory: ./backend
            command: go run main.go
          - name: Frontend
            directory: ./frontend
            command: npm run dev
      - provider: dbeaver
        connection: "postgresql://${DB_USER}:${DB_PASS}@localhost:5432/mydb"
      - provider: chrome
        urls:
          - http://localhost:3000
          - http://localhost:8080/api/docs
      - provider: docker
        file: ./docker-compose.yml
        profile: dev
      - provider: command
        cwd: ./
        cmd: git pull --rebase
  staging:
    env_file: .env.staging
    services:
      - provider: chrome
        urls:
          - https://staging.myapp.com
      - provider: command
        cwd: ./
        cmd: git checkout staging && git pull
```

### Naming

The folder containing `workspace.yaml` is used as the workspace name. If the `name` field in the YAML differs, nook prints a warning and uses the folder name.

### Environments

Each environment maps to a set of services. At least one environment is required. An `env_file` can be specified per environment — variables from it are loaded before resolving `${VAR}` placeholders in service configs.

### Providers

| Provider | Key | Description |
|----------|-----|-------------|
| VS Code | `vscode` | Opens a folder. If `terminals` are configured, generates a `.code-workspace` file with pre-configured terminal tasks that auto-run on open. Multiple VS Code services can open different folders. |
| DBeaver | `dbeaver` | Opens a database connection via the `-con` CLI flag. Connection strings support `${VAR}` resolution from env vars and `.env` files. |
| Chrome | `chrome` | Opens each URL in a new tab. Platform-specific launch (macOS: `open -a`, Linux: direct binary, Windows: `start chrome`). |
| Docker | `docker` | Runs `docker compose -f <file> up -d`. An optional `profile` sets `--profile`. |
| Command | `command` | Runs an arbitrary shell command in a working directory. Fire-and-forget — nook doesn't wait for it to finish. |

### VS Code terminals

When `terminals` are configured for a VS Code service, nook generates a `.code-workspace` file in a `.workspace/` directory alongside the config. The generated file contains task definitions with `runOn: "folderOpen"` so terminals start automatically when VS Code opens the workspace. No extensions required.

If the VS Code service has a `delay_ms`, a sleep is prepended to each terminal command (`sleep 3 && npm run dev`).

### Launch ordering

By default all services launch in parallel. Optionally, you can control the sequence:

```yaml
environments:
  dev:
    services:
      - provider: docker
        file: ./docker-compose.yml
      - provider: vscode
        folder: ./backend
        order: 2
        delay_ms: 3000
      - provider: chrome
        urls:
          - http://localhost:3000
        order: 3
        ready_check:
          cmd: "curl -sf http://localhost:3000/health"
          interval_ms: 2000
          timeout_ms: 30000
```

| Field | Description |
|-------|-------------|
| `order` | Launch position. Docker is always first. Same order = launch together in parallel. 0 = unassigned (launches after ordered ones). |
| `delay_ms` | Sleep for N milliseconds before launching this service. |
| `ready_check.cmd` | Shell command polled until it exits 0. Blocks until success or timeout. |
| `ready_check.interval_ms` | Poll interval (default 2000). |
| `ready_check.timeout_ms` | Give up after N ms (default 30000). |
| `wait_for_compose_healthy` | Per-environment flag. Polls `docker compose ps --format json` until all containers are healthy. Timeout 120s. |

`nook init` guides you through all of this interactively — confirm ordering, assign positions, choose delays or health checks between each pair, with the last delay value remembered as default.

### Environment variables

Use `${VAR}` placeholders in connection strings, commands, and URLs. Resolution order:

1. Values from `os.Environ()`
2. Values from the `.env` file specified via `env_file` for the current environment
3. If a variable is not found, the placeholder is left unchanged

This means you can keep secrets in `.env` files (not committed) and reference them in `workspace.yaml` (which can be committed).

## Build

```bash
go build -o nook .
go test ./... -v -cover -race
```
