# Nook — Workspace Organizer CLI

> **Name note:** Originally designed as "Cove", but that name is taken by an existing Rust CLI tool (`cove-cli` on crates.io). The name "Nook" was chosen as the alternative — it means a cozy, sheltered corner/place, fitting the workspace organizer metaphor.

## Overview

**Nook** is a cross-platform CLI tool (macOS, Windows, Linux) for developers to organize and launch their project workspace with a single command. It opens all the tools you need for a project: VS Code (in the right folder, with pre-opened terminals), DBeaver (connected to the right database), Chrome (at the right URLs), Docker Compose, and custom commands — all in parallel.

**Written in Go**, with colors, an interactive TUI, and a clean architecture designed for easy extension with new integrations.

---

## Name & Constraints

- The command is `nook`
- Verify no naming conflicts with existing tools before building
- Must compile to a **single binary** for macOS (amd64 + arm64), Linux (amd64 + arm64), Windows (amd64)
- Use **stable** versions of all dependencies only (no beta, no alpha)
- Use **Go 1.22+** (latest stable as of writing)

---

## Tech Stack (Required)

| Library | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework (commands, flags, help, autocomplete) |
| `github.com/spf13/viper` | Config management (YAML read/write, env vars) |
| `github.com/AlecAivazis/survey/v2` | Interactive prompts (menus, confirmations, inputs) |
| `github.com/fatih/color` | Colored terminal output |
| `gopkg.in/yaml.v3` | YAML parsing/serialization |
| `github.com/go-playground/validator/v10` | Struct validation via tags |
| `github.com/adrg/xdg` | Cross-platform config directory resolution |

---

## Configuration Structure

### Global config: `~/.nook/config.yaml`

```yaml
# ~/.nook/config.yaml
scan_paths:
  - ~/.nook/workspaces         # Hand-created workspaces
  # Trusted repo directories are added here automatically
```

### Workspace configs

For **hand-created workspaces** (via `nook init`):

```
~/.nook/workspaces/<workspace-name>/
├── workspace.yaml
├── .env.dev                         # optional
├── .env.staging                     # optional
└── .workspace/                      # generated files (e.g., .code-workspace)
```

For **workspaces found in repos** (committed `workspace.yaml`):

```
~/projects/some-repo/
├── workspace.yaml                   # committed to the repo
├── .env.dev                         # may or may not be committed
└── ...                              # project files
```

### `workspace.yaml` format

```yaml
name: my-project                      # Display name (folder name takes precedence)
description: "My fullstack app with React + Go"

environments:                         # At least one environment required
  dev:
    env_file: .env.dev                # optional, path relative to workspace.yaml location
    services:
      - provider: vscode
        folder: ./backend
        terminals:                    # optional — if specified, a .code-workspace is generated
          - name: "Backend"
            directory: ./backend
            command: "go run main.go"  # optional
          - name: "Frontend"
            directory: ./frontend
            command: "npm run dev"
      - provider: vscode
        folder: ./frontend            # Multiple VS Code workspaces can be opened
      - provider: dbeaver
        connection: "postgresql://${DB_USER}:${DB_PASS}@localhost:5432/mydb"
      - provider: chrome
        urls:
          - "http://localhost:3000"
          - "http://localhost:8080/api/docs"
      - provider: docker
        file: ./docker-compose.yml    # relative to workspace.yaml location
        profile: dev                  # optional docker-compose profile
      - provider: command
        cwd: ./
        cmd: "git pull --rebase"
  staging:
    env_file: .env.staging
    services:
      - provider: chrome
        urls:
          - "https://staging.myapp.com"
      - provider: command
        cwd: ./
        cmd: "git checkout staging && git pull"
```

> **If no `environments` key is defined**, a single `dev` environment is automatically created containing all top-level services.

### Naming precedence

1. Folder name containing `workspace.yaml` is the workspace name
2. If `name` field in yaml differs from folder name → print a warning, use folder name
3. The `name` field in yaml is used for display and for import scenarios (sharing files)

### Validation (using `validator.v10`)

All config structs must be validated using struct tags:

```go
type WorkspaceConfig struct {
    Name         string                 `yaml:"name" validate:"required,min=1,max=100"`
    Description  string                 `yaml:"description"`
    Environments map[string]Environment `yaml:"environments" validate:"required,min=1,dive"`
}

type Environment struct {
    EnvFile  string    `yaml:"env_file"`
    Services []Service `yaml:"services" validate:"required,min=1,dive"`
}

type Service struct {
    Provider   string     `yaml:"provider" validate:"required,oneof=vscode dbeaver chrome docker command"`
    Folder     string     `yaml:"folder,omitempty"`
    Terminals  []Terminal `yaml:"terminals,omitempty"`
    Connection string     `yaml:"connection,omitempty"`
    URLs       []string   `yaml:"urls,omitempty"`
    File       string     `yaml:"file,omitempty"`
    Profile    string     `yaml:"profile,omitempty"`
    Cmd        string     `yaml:"cmd,omitempty"`
    Cwd        string     `yaml:"cwd,omitempty"`
}

type Terminal struct {
    Name      string `yaml:"name" validate:"required"`
    Directory string `yaml:"directory" validate:"required"`
    Command   string `yaml:"command"`
}
```

Validate:
- On `nook init` — before saving the file
- On `nook open` — in case someone hand-edited the yaml (print errors and abort)
- Provide **clear error messages** pointing to the exact invalid field

---

## Commands

### `nook init`

Interactive walkthrough to create a new workspace.

```
$ nook init

? Workspace name: my-fullstack-app
? Description: A fullstack app with React + Go
? Add environment? (Y/n) Yes
? Environment name: dev
? Path to .env file (optional): .env.dev

── Configure Services for "dev" ──

? Which services do you want to add? (space to select)
  [x] VS Code
  [x] DBeaver
  [x] Chrome
  [ ] Docker Compose
  [x] Custom Command

── VS Code ──
? Root folder (relative): ./
? Add pre-opened terminals? (y/N) Yes
  ? Terminal 1 name: Backend
  ? Terminal 1 directory: ./backend
  ? Terminal 1 startup command (optional): go run main.go
  ? Add another? Yes
  ? Terminal 2 name: Frontend
  ? Terminal 2 directory: ./frontend
  ? Terminal 2 startup command (optional): npm run dev

── DBeaver ──
? Connection string (use ${VAR} for env vars):
  postgresql://${DB_USER}:${DB_PASS}@localhost:5432/mydb

── Chrome ──
? URLs to open (comma separated): http://localhost:3000, http://localhost:8080

── Docker Compose ──
? Path to docker-compose file: ./docker-compose.yml
? Profile (optional):

── Custom Commands ──
? Command to run: git pull
? Working directory (relative): ./
? Add another command? (y/N)

? Add another environment? (y/N) No

✔ Workspace "my-fullstack-app" created at ~/.nook/workspaces/my-fullstack-app/
```

### `nook open [name] [--env env]`

Opens a workspace. Interactive if no name given.

```
# Direct command
$ nook open my-fullstack-app --env dev

# Interactive (name omitted)
$ nook open
? Select workspace (type to filter):
  my-fullstack-app      A fullstack app with React + Go
  another-project       Some other project

? Select environment: dev
  (only shown if multiple environments exist)

✔ Opening workspace "my-fullstack-app" (dev)...

  [VS Code]    Opening ./backend...
  [VS Code]    Opening terminal "Backend" → ./backend
  [DBeaver]    Opening connection to localhost:5432...
  [Chrome]     Opening http://localhost:3000...
  [Chrome]     Opening http://localhost:8080...
  [Docker]     Running docker compose -f ./docker-compose.yml up -d...
  [Command]    Running git pull --rebase in ./...
  [Command]    Running go mod download in ./backend...

✔ All services launched!
```

**UI rules:**
- Provider names in **cyan/bold**
- File paths in **dim/italic**
- Success marks `✔` in **green**, errors `✖` in **red**, loading `⏳` in **yellow**
- All services launch in **parallel** using goroutines
- Show incremental output as each service starts
- If a service fails, show `✖` with the error but continue launching others

### `nook list`

```
$ nook list
From ~/.nook/workspaces:
  my-fullstack-app      A fullstack app with React + Go    [dev, staging]
  another-project       Some other project                  [dev]

From ~/projects:
  side-project          A side project                       [dev]
```

### `nook edit [name]`

Opens `workspace.yaml` in `$EDITOR` (fallback: `vim` on macOS/Linux, `notepad` on Windows).

If no name given, show interactive selector.

### `nook delete [name]`

Deletes the workspace folder from `~/.nook/workspaces/`. Confirm before deleting.

Does **not** delete workspaces found in repos (only removes from scan index).

### `nook scan`

Rescans all paths in `scan_paths` for `workspace.yaml` files. Removes stale paths (directories that no longer exist). Prints summary of found workspaces.

### `nook detect`

Scans the system to detect which providers are available.

```
$ nook detect

  Provider       Status
  ─────────────────────────
  VS Code        ✔  detected at /usr/local/bin/code
  DBeaver        ✔  detected at /Applications/DBeaver.app
  Chrome         ✔  detected at /Applications/Google Chrome.app
  Docker         ✔  detected at /usr/local/bin/docker
  Terminal       ✔  detected (fallback for system terminal)

  Providers shown in dim/gray are not installed.
```

Detection logic per provider:
- **VS Code**: check `code` in PATH (or `code-insiders`), also check common install locations per OS
- **DBeaver**: check for `dbeaver` binary, or common app paths. On Windows, check for `dbeaver-cli.exe`
- **Chrome**: check `google-chrome`, `google-chrome-stable`, `chrome` in PATH, or common app paths
- **Docker**: check `docker` in PATH
- **Command**: always available (it's just running shell commands)

---

## Auto-detection of Workspaces in Repos

This is a key UX feature. It must work **without the user manually adding anything**.

### On every `nook` command execution:

```go
func autoDetectWorkspaces() {
    // 1. Check current working directory for workspace.yaml
    // 2. Check immediate subdirectories (./*/workspace.yaml)
    // 3. For each found file:
    //    a. Check if its parent directory is already in scan_paths
    //    b. If not → prompt user: "Found workspace.yaml in <dir>. Trust it?"
    //    c. If yes → add parent directory to scan_paths in config.yaml
}
```

### Flow:

```
$ cd ~/projects/team-repo
$ nook open

? Found workspace.yaml in /Users/me/projects/team-repo/
  Workspace: "team-repo" — A fullstack app
  Do you want to trust this workspace? (Y/n) Yes

✔ Workspace "team-repo" added to scan paths.
  (added /Users/me/projects/team-repo to scan_paths)
? Select environment: dev
✔ Opening...
```

If user is in a directory containing multiple repos:

```
$ cd ~/projects/
$ nook list

? Found workspace.yaml files in this directory:
  [x] team-repo       → Team project
  [ ] another-repo    → Side project
  [x] side-project    → Personal project

  Trust selected workspaces? (Y/n)
```

### Trust persistence

- When trusted, the **parent directory** is added to `scan_paths` in `~/.nook/config.yaml`
- On subsequent runs, no prompt is shown (already trusted)
- If a directory no longer exists → remove from `scan_paths` silently on `nook scan`

---

## VS Code + Terminals Strategy

**Rule:** Only generate a `.code-workspace` file when absolutely necessary.

| User configured | What we do |
|---|---|
| VS Code with **no terminals** | `code <folder>` — simple, no generated files |
| VS Code with **terminals** | Generate `.code-workspace` file at `<workspace-folder>/.workspace/<name>.code-workspace` |

### Generated `.code-workspace` example:

```json
{
  "folders": [
    { "path": "../../../backend" }
  ],
  "settings": {},
  "tasks": {
    "version": "2.0.0",
    "tasks": [
      {
        "label": "Terminal: Backend",
        "type": "shell",
        "command": "cd ./backend && go run main.go",
        "presentation": {
          "echo": true,
          "reveal": "always",
          "focus": false,
          "panel": "dedicated",
          "group": "nook-terminals"
        },
        "runOptions": { "runOn": "folderOpen" },
        "problemMatcher": []
      },
      {
        "label": "Terminal: Frontend",
        "type": "shell",
        "command": "cd ./frontend && npm run dev",
        "presentation": {
          "echo": true,
          "reveal": "always",
          "focus": false,
          "panel": "dedicated",
          "group": "nook-terminals"
        },
        "runOptions": { "runOn": "folderOpen" },
        "problemMatcher": []
      }
    ]
  }
}
```

Then open with: `code <workspace-folder>/.workspace/<name>.code-workspace`

> **Note:** VS Code's `runOn: "folderOpen"` will auto-start terminals when the workspace opens. This is the best we can do without extensions.

---

## DBeaver CLI Support

DBeaver **does** support CLI connection arguments. Use them.

### Launch commands by OS:

| OS | Command |
|---|---|
| macOS | `/Applications/DBeaver.app/Contents/MacOS/dbeaver -con "connection_string"` or `open -a "DBeaver.app" --args -con "connection_string"` |
| Linux | `dbeaver -con "connection_string"` |
| Windows | `dbeaver-cli.exe -con "connection_string"` (or `dbeaver.exe -con "connection_string"`) |

### Connection string format:

The `-con` argument takes a single parameter with all connection details. It opens the database connection in the DBeaver UI.

Additionally, DBeaver supports:
- `-f <file>` — Opens a SQL file in DBeaver UI, optionally connects to datasource if `-con` is also provided
- `-save` — Saves the connection for future use
- `-vars <path>` — Path to a property file with variables

**Implementation:**
- For v1, use `-con` to open the connection directly
- The connection string from `workspace.yaml` is passed as-is
- If the connection string contains `${VAR}` placeholders, they are resolved before passing to DBeaver
- On **Windows**, prefer `dbeaver-cli.exe` (does not spawn a new window, shows output in terminal)

### Credentials & Secrets

Support for resolving `${VAR}` placeholders in configuration values (especially DBeaver connection strings):

1. **Environment variables** — `${DB_PASS}` resolves from `os.Getenv("DB_PASS")`
2. **`.env` files** — If an `env_file` is specified for an environment, load it before resolving using `github.com/joho/godotenv` (or a manual parser — no external dependency needed for a simple key=value parser)
3. **System keychain** (future) — Not required for v1, but the architecture should allow adding it later

**Security warning:** If a connection string contains plaintext passwords and the user is about to commit `workspace.yaml` to a repo, show a warning.

---

## Provider Interface (Architecture for Extensibility)

All integrations must implement this interface:

```go
package provider

import "context"

type Provider interface {
    // Name returns the provider identifier (e.g., "vscode", "dbeaver")
    Name() string

    // Detect checks if this tool is installed on the system
    Detect() (bool, error)

    // Launch starts the service with the given configuration
    // ctx can carry cancellation signals
    Launch(ctx context.Context, service ServiceConfig, envVars map[string]string) error
}

type ServiceConfig struct {
    Provider   string     `yaml:"provider"`
    Folder     string     `yaml:"folder,omitempty"`
    Terminals  []Terminal `yaml:"terminals,omitempty"`
    Connection string     `yaml:"connection,omitempty"`
    URLs       []string   `yaml:"urls,omitempty"`
    File       string     `yaml:"file,omitempty"`
    Profile    string     `yaml:"profile,omitempty"`
    Cmd        string     `yaml:"cmd,omitempty"`
    Cwd        string     `yaml:"cwd,omitempty"`
    // RawConfig holds any provider-specific extra fields
    RawConfig  map[string]interface{}
}
```

Providers register themselves in `main.go` or an `init()` registry:

```go
var registry = map[string]Provider{}

func Register(p Provider) {
    registry[p.Name()] = p
}
```

**For v1, implement these providers:**
1. `vscode` — launches VS Code, optionally generates `.code-workspace`
2. `dbeaver` — launches DBeaver with connection string using the `-con` CLI flag
3. `chrome` — opens URLs in Chrome (or system default browser)
4. `docker` — runs `docker compose -f <file> up -d` (with optional profile)
5. `command` — runs arbitrary shell commands

---

## Provider-Specific Launch Behaviors

### VS Code
- **No terminals:** `exec.Command("code", folder)`
- **With terminals:** Generate `.code-workspace` file, then `exec.Command("code", workspaceFile)`
- Workspace file goes into `~/.nook/workspaces/<name>/.workspace/` or `<repo>/.workspace/`
- Generate a fresh `.code-workspace` on each `nook open` (keeps it in sync with config)

### DBeaver
- Use the `-con` flag with the connection string as described in the DBeaver CLI section above
- macOS: Prefer direct executable path over `open -a` for reliability
- Linux: `dbeaver -con "..."`
- Windows: `dbeaver-cli.exe -con "..."` or `dbeaver.exe -con "..."`
- If DBeaver is not installed / not found, print a clear error and skip

### Chrome
- macOS: `exec.Command("open", "-a", "Google Chrome", url)`
- Linux: `exec.Command("google-chrome", url)` or `xdg-open url`
- Windows: `exec.Command("cmd", "/c", "start", "chrome", url)`
- Open each URL in a new tab (use `--new-tab` or equivalent flags)

### Docker Compose
- `exec.Command("docker", "compose", "-f", file, "--profile", profile, "up", "-d")`
- If no profile specified, omit `--profile`
- Stream output to the user in real-time

### Command
- `exec.Command("sh", "-c", cmd)` on macOS/Linux
- `exec.Command("cmd", "/c", cmd)` on Windows
- Set working directory to `cwd` (resolved relative to workspace root)
- Stream stdout/stderr to the user in real-time
- Non-blocking: fire and forget (don't wait for command to finish)

---

## Output & UI Styling

### Color palette

| Element | Color |
|---|---|
| Provider names | Cyan + Bold |
| File paths | Dim + Italic |
| Success `✔` | Green |
| Error `✖` | Red |
| Loading `⏳` | Yellow |
| Headers/borders | Blue + Bold |
| Dim/less important items | Dark gray |

### Interactive prompts (survey.v2)

- **Select**: Use `survey.Select` with Vim keybindings (j/k to navigate)
- **Multi-select**: Use `survey.MultiSelect` with checkboxes
- **Input**: Use `survey.Input` with sensible defaults
- **Confirm**: Use `survey.Confirm` with Y/n defaults
- All prompts should have sensible defaults and be skippable with Enter

---

## Sample Session (Full UX Walkthrough)

```bash
# First time using nook — detect what's available
$ nook detect
  Provider       Status
  ─────────────────────────
  VS Code        ✔  detected
  DBeaver        ✔  detected
  Chrome         ✔  detected
  Docker         ✔  detected

# Create a workspace
$ nook init
  ... (interactive walkthrough as shown above)

# Open it (interactive since only one workspace)
$ nook open
✔ Opening workspace "my-fullstack-app" (dev)...
  [VS Code]    Opening ./backend...
  [VS Code]    Opening terminal "Backend" → ./backend
  [DBeaver]    Opening connection to localhost:5432...
  [Chrome]     Opening http://localhost:3000...
  [Docker]     Running docker compose -f ./docker-compose.yml up -d...
✔ All services launched!

# List all
$ nook list
  my-fullstack-app      A fullstack app with React + Go    [dev, staging]

# Open a teammate's repo with a committed workspace.yaml
$ cd ~/projects/team-project
$ nook open
? Found workspace.yaml in /Users/me/projects/team-project/
  Do you want to trust this workspace? (Y/n) Yes
? Select environment: dev
✔ Opening...
```

---

## Testing Requirements (CRITICAL)

**Test-Driven Development approach must be followed:**

For each piece of functionality implemented, the agent MUST:
1. Write the test **first** (before the implementation code)
2. Run the test — it should fail initially (red)
3. Write the implementation code to make the test pass (green)
4. Run the test again to confirm it passes
5. **Commit the passing code** (both test and implementation)
6. Move to the next piece of functionality

**Minimum test coverage targets (per package):**

| Package | Coverage Target |
|---|---|
| `config/` (parsing, validation, env resolution) | ≥ 90% |
| `provider/` (each provider + registry) | ≥ 85% |
| `detector/` (auto-detection logic) | ≥ 85% |
| `cmd/` (command integration) | ≥ 70% |
| `tui/` (prompts and output) | ≥ 60% |
| `utils/` (path resolution, platform detection) | ≥ 90% |

**What to test specifically:**
- `workspace.yaml` parsing and struct validation (valid configs, invalid configs, edge cases)
- Environment variable resolution (`${VAR}` from os.Environ, `.env` files, missing vars)
- Provider detection (tool found, tool not found, tool at custom path)
- Provider launch commands (correct command constructed, correct args)
- `.code-workspace` generation (correct JSON output)
- Auto-detection logic (workspace found in cwd, in subdirs, not found, already trusted)
- Cross-platform path resolution (relative → absolute on each OS)
- Config file save/load/update
- Error handling (malformed yaml, missing fields, invalid provider names)

**Test framework:** Use Go's built-in `testing` package + `github.com/stretchr/testify/assert` for assertions.

**Mocking:** For provider detection and launch tests, use interface-based mocking. Do not actually launch real applications in tests.

---

## Distribution & Package Management

The project must be distributable via the following package managers. Include the build and release pipeline configuration.

### macOS: Homebrew

Create a **custom Homebrew tap** repository named `homebrew-nook` (or `homebrew-tap`) containing a formula for `nook`.

Formula template:
```ruby
class Nook < Formula
  desc "Workspace organizer CLI for developers"
  homepage "https://github.com/<your-org>/nook"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/<your-org>/nook/releases/download/v0.1.0/nook-darwin-arm64.tar.gz"
      sha256 "<sha256>"
    else
      url "https://github.com/<your-org>/nook/releases/download/v0.1.0/nook-darwin-amd64.tar.gz"
      sha256 "<sha256>"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/<your-org>/nook/releases/download/v0.1.0/nook-linux-arm64.tar.gz"
      sha256 "<sha256>"
    else
      url "https://github.com/<your-org>/nook/releases/download/v0.1.0/nook-linux-amd64.tar.gz"
      sha256 "<sha256>"
    end
  end

  def install
    bin.install "nook"
  end
end
```

### Windows: Winget (built-in) + Scoop (developer-friendly)

**Winget** ships with Windows 11 and is the most accessible option. Create a winget manifest.

**Scoop** is the most popular developer-focused package manager for Windows. Create a Scoop manifest in a bucket:

```powershell
# Scoop manifest: bucket/nook.json
{
    "version": "0.1.0",
    "description": "Workspace organizer CLI for developers",
    "homepage": "https://github.com/<your-org>/nook",
    "license": "MIT",
    "architecture": {
        "64bit": {
            "url": "https://github.com/<your-org>/nook/releases/download/v0.1.0/nook-windows-amd64.zip",
            "hash": "<sha256>"
        }
    },
    "bin": "nook.exe",
    "checkver": "github",
    "autoupdate": {
        "architecture": {
            "64bit": {
                "url": "https://github.com/<your-org>/nook/releases/download/v$version/nook-windows-amd64.zip"
            }
        }
    }
}
```

### Linux: Homebrew (cross-distro)

Since Linux doesn't have a single universal package manager, **Homebrew works on Linux too** (formerly Linuxbrew) and is the most practical option for developers across distros. The same Homebrew formula above works on Linux.

For a more distro-specific approach, also provide:
- **`.deb` package** for Debian/Ubuntu (via `dpkg-deb` or `nfpm`)
- **`.rpm` package** for Fedora/RHEL (via `rpmbuild` or `nfpm`)

### Build & Release Pipeline

Use **GoReleaser** to automate the build and release process:

```yaml
# .goreleaser.yaml
version: 2
project_name: nook

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md

brews:
  - repository:
      owner: <your-org>
      name: homebrew-nook
    folder: Formula
    homepage: "https://github.com/<your-org>/nook"
    description: "Workspace organizer CLI for developers"

scoop:
  - repository:
      owner: <your-org>
      name: scoop-nook
    homepage: "https://github.com/<your-org>/nook"
    description: "Workspace organizer CLI for developers"

checksum:
  name_template: "checksums.txt"
```

---

## Project Structure (Recommended)

```
nook/
├── main.go
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go           # Root command with global flags
│   ├── init.go           # nook init
│   ├── open.go           # nook open
│   ├── list.go           # nook list
│   ├── edit.go           # nook edit
│   ├── delete.go         # nook delete
│   ├── scan.go           # nook scan
│   └── detect.go         # nook detect
├── config/
│   ├── config.go         # Global config (~/.nook/config.yaml) management
│   ├── workspace.go      # WorkspaceConfig struct, validation, load/save
│   └── resolver.go       # Env var resolver (${VAR} → value from env/.env)
├── provider/
│   ├── interface.go      # Provider interface
│   ├── registry.go       # Provider registry
│   ├── vscode.go         # VS Code provider
│   ├── dbeaver.go        # DBeaver provider
│   ├── chrome.go         # Chrome provider
│   ├── docker.go         # Docker Compose provider
│   └── command.go        # Custom command provider
├── tui/
│   ├── prompts.go        # Shared survey prompts
│   ├── styles.go         # Color/styles helpers
│   └── output.go         # Formatted output (progress, success, error messages)
├── detector/
│   └── autodetect.go     # Auto-detection of workspace.yaml in cwd/subdirs
└── utils/
    ├── path.go           # Path resolution (relative → absolute)
    └── platform.go       # OS detection, platform-specific commands
```

---

---

## GitHub Actions CI/CD

Two workflow files are required. The CI workflow should be fully active. The release workflow should be present but commented out / disabled, ready to be activated when distribution is set up.

### 1. CI Workflow — Build & Test (active on every push/PR)

```yaml
# .github/workflows/ci.yaml
name: CI

on:
  push:
    branches: ["*"]
  pull_request:
    branches: [main, master]

permissions:
  contents: read

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ["1.22", "1.23"]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
          cache: true

      - name: Lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

      - name: Test
        run: go test ./... -v -cover -coverprofile=coverage.out -race

      - name: Display coverage
        run: go tool cover -func=coverage.out

      - name: Build
        run: go build -o nook ./main.go

      - name: Upload binary (artifact)
        uses: actions/upload-artifact@v4
        with:
          name: nook-${{ matrix.os }}-go${{ matrix.go }}
          path: nook
```

### 2. Release Workflow — Tag-triggered, builds + publishes (present but DISABLED)

This workflow is **fully written but disabled** via `if: false`. The user can enable it later by removing the `if` condition and configuring the secrets.

```yaml
# .github/workflows/release.yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write
  packages: write

jobs:
  goreleaser:
    if: false  # DISABLED — remove this line to activate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
          cache: true

      - name: Run tests
        run: go test ./... -v -cover -race

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: \${{ secrets.GITHUB_TOKEN }}
          # Secrets needed when activating:
          # HOMEBREW_TAP_TOKEN: \${{ secrets.HOMEBREW_TAP_TOKEN }}
          # SCOOP_BUCKET_TOKEN: \${{ secrets.SCOOP_BUCKET_TOKEN }}
```

**What happens when activated:**
1. A developer pushes a tag `v0.1.0`
2. GoReleaser builds binaries for all platforms (macOS amd64+arm64, Linux amd64+arm64, Windows amd64)
3. Archives are created (`.tar.gz` for macOS/Linux, `.zip` for Windows)
4. A **GitHub Release** is created with the binaries attached
5. The **Homebrew formula** is updated in the `homebrew-nook` tap repository
6. The **Scoop manifest** is updated in the `scoop-nook` bucket repository
7. Checksums are generated and attached

### Required secrets (for when release workflow is activated):

| Secret | Purpose |
|---|---|
| `GITHUB_TOKEN` | Built-in, no setup needed |
| `HOMEBREW_TAP_TOKEN` | PAT with push access to `homebrew-nook` repo |
| `SCOOP_BUCKET_TOKEN` | PAT with push access to `scoop-nook` repo |

---

## What This Prompt Expects From the Agent

1. **Full working Go application** with all commands implemented
2. **All 5 providers** working (VS Code, DBeaver, Chrome, Docker, Command) — DBeaver MUST use the `-con` CLI flag
3. **Interactive TUI** using survey.v2 for all prompts
4. **Colorful output** using fatih/color
5. **Validation** using go-playground/validator on all configs
6. **Auto-detection** of workspaces in current directory and subdirectories
7. **Cross-platform** launches (macOS, Linux, Windows) using appropriate commands
8. **Clean architecture** with the Provider interface for future extensions
9. **Config persistence** in `~/.nook/config.yaml` and `~/.nook/workspaces/`
10. **Generated `.code-workspace`** files when VS Code terminals are configured
11. **Environment variable resolution** from os.Environ() and .env files
12. **Proper error handling** — never panic, always show user-friendly errors
13. **`nook detect` command** that checks for installed tools
14. **Parallel launches** using goroutines with individual progress reporting
15. **Test-Driven Development** — write tests first, then implementation, commit only after tests pass
16. **GoReleaser configuration** for automated builds and releases
17. **Homebrew formula** for macOS and Linux
18. **Scoop manifest** for Windows
19. **Winget manifest** for Windows
20. **CI workflow (`.github/workflows/ci.yaml`)** — runs on every push/PR: lint, test on all OSes, build, upload artifacts
21. **Release workflow (`.github/workflows/release.yaml`)** — present but disabled, triggers on tags, runs GoReleaser, creates GitHub Release with binaries

---

## Deliverables

- A complete Go project in a single directory ready to `go build`
- All dependencies in `go.mod` / `go.sum`
- Cross-compilation build commands for macOS, Linux, Windows (via GoReleaser)
- The binary should be named `nook`
- `.goreleaser.yaml` for automated release pipeline
- `Formula/nook.rb` for Homebrew tap
- Scoop manifest for Windows
- Winget manifest for Windows
- `.github/workflows/ci.yaml` — fully active CI workflow
- `.github/workflows/release.yaml` — written but disabled release workflow
- README.md with installation and usage instructions
- All tests passing with coverage targets met
- Clean git history with one commit per feature (test + implementation committed together)
