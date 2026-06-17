# Nook

Workspace organizer CLI for developers. Launch all your project tools with a single command.

## Install

### macOS / Linux

```bash
brew install lorenzo-vecchio/nook/nook
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

## Usage

```
nook init                  Create a workspace interactively
nook open [name] [--env]   Open a workspace
nook list                  List all workspaces
nook edit [name]           Edit a workspace config in $EDITOR
nook delete [name]         Delete a workspace
nook scan                  Rescan paths for workspaces
nook detect                Show detected provider tools
```

## Workspace config

Workspaces are defined in `workspace.yaml`:

```yaml
name: my-app
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
      - provider: dbeaver
        connection: "postgresql://${DB_USER}:${DB_PASS}@localhost:5432/mydb"
      - provider: chrome
        urls:
          - http://localhost:3000
      - provider: docker
        file: ./docker-compose.yml
        profile: dev
      - provider: command
        cwd: ./
        cmd: git pull --rebase
  staging:
    services:
      - provider: chrome
        urls:
          - https://staging.myapp.com
```

### Providers

| Provider | Description |
|----------|-------------|
| `vscode` | Opens VS Code in a folder, with optional pre-configured terminals |
| `dbeaver` | Opens DBeaver with a database connection |
| `chrome` | Opens URLs in Chrome tabs |
| `docker` | Runs `docker compose up -d` |
| `command` | Runs arbitrary shell commands |

### Environment variables

Use `${VAR}` in connection strings, commands, and URLs. Nook resolves from `os.Environ()` and from `.env` files specified via `env_file`.

## Auto-detection

Place a `workspace.yaml` in any project repo. When you run `nook open` from that directory, Nook detects it and prompts you to trust it once.

## Build

```bash
go build -o nook .
go test ./... -v -cover -race
```
