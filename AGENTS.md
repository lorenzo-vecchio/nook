# Nook — Workspace Organizer CLI

Cross-platform Go CLI to organize and launch project workspaces with a single command.

> **Source of truth:** `nook-prompt.md` — the full specification. This file describes *how to build it*.

## Quick Reference

```bash
go build -o nook .                                    # Build
go test ./... -v -cover -race                         # Test
go test ./... -coverprofile=cov.out && go tool cover -func=cov.out  # Coverage
golangci-lint run                                      # Lint
```

## Implementation Plan

See `IMPLEMENTATION_PLAN.md` for the ordered build sequence. The plan breaks work into phases with task-level granularity.

## Package Map

| Package | Responsibility |
|---------|---------------|
| `cmd/` | Cobra commands (init, open, list, edit, delete, scan, detect) |
| `config/` | Global config (~/.nook/config.yaml), workspace.yaml parsing/validation, env var resolution |
| `provider/` | Provider interface + registry + 5 implementations (vscode, dbeaver, chrome, docker, command) |
| `tui/` | Survey prompts, color helpers, formatted output |
| `detector/` | Auto-discovery of workspace.yaml in cwd and subdirectories |
| `utils/` | Path resolution, platform detection |

## Dependencies

cobra, viper, survey/v2, fatih/color, yaml.v3, validator/v10, xdg, testify

## Development Rules

### Work Like a Developer

1. **Commit as you go.** After completing each task from the implementation plan, commit with a descriptive message. Keep commits small and focused.
2. **TDD required.** Write the test first, watch it fail, implement, verify green, then commit test + implementation together.
3. **One commit per feature / task.** Never bundle unrelated changes.
4. **Never commit failing tests.** The repo should always be in a buildable, test-passing state.
5. **No comments.** Code should be self-documenting. No doc comments on exported symbols. No inline comments.
6. **No over-engineering.** Prefer the simplest, shortest solution. No unnecessary abstractions, no DI containers, no frameworks beyond those listed.
7. **Never panic.** Always return errors. Use `fatih/color` for user-facing output, not `fmt.Print`.
8. **Mock external deps** via interfaces. Never launch real applications in tests.

### Commit Message Convention

```
<package>: <imperative verb> <what>
```

Examples:
```
config: add workspace YAML parsing with validation
provider: implement VS Code launch with terminal support
cmd: wire nook open command
```

## Coverage Targets

| Package | Minimum |
|---------|---------|
| `config/` | 90% |
| `provider/` | 85% |
| `detector/` | 85% |
| `utils/` | 90% |
| `cmd/` | 70% |
| `tui/` | 60% |

## Architecture Notes

- Providers register via `provider.Register()` called from `init()` in each provider file.
- VS Code: open folder directly unless terminals are configured (then generate .code-workspace).
- DBeaver: use `-con` CLI flag with the connection string.
- Docker: `docker compose -f <file> up -d` with optional `--profile`.
- Command: `sh -c` (macOS/Linux) or `cmd /c` (Windows), fire-and-forget.
- Chrome: open each URL in a new tab, platform-specific commands.
- Services launch in parallel via goroutines. Failures are logged but don't block others.
- Config lives at `~/.nook/config.yaml`. Workspaces at `~/.nook/workspaces/<name>/`.
- Auto-detect workspace.yaml in cwd and immediate subdirs on every command. Prompt to trust.
