# Nook — Workspace Organizer CLI

Cross-platform Go CLI to organize and launch project workspaces with a single command. **Fully built and shipped.** In active maintenance and feature development.

> Source of truth: `nook-prompt.md` (spec), `IMPLEMENTATION_PLAN.md` (build order). Both reference docs — not meant to be kept in sync with code changes.

## Quick Reference

```bash
go build -o nook .                                    # Build
go test ./... -v -cover -race                         # Test
go test ./... -coverprofile=cov.out && go tool cover -func=cov.out  # Coverage
go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run  # Lint
```

CI runs on every push: lint → test → coverage, across macOS, Linux, and Windows.

## Package Map

| Package | Responsibility |
|---------|---------------|
| `cmd/` | Cobra commands (init, open, list, edit, delete, scan, detect) |
| `config/` | Global config, workspace.yaml parsing/validation, env var resolver |
| `provider/` | Interface + registry + 5 implementations (vscode, dbeaver, chrome, docker, command) |
| `tui/` | Survey prompts (with filter), color helpers, formatted output |
| `detector/` | Auto-discovery + trust persistence for workspace.yaml |
| `utils/` | Path resolution, platform detection |

## Dependencies

cobra, survey/v2, fatih/color, yaml.v3, validator/v10, xdg, godotenv, testify

## Development Rules

1. **Commit as you go.** Keep commits small, focused, one feature per commit.
2. **TDD required.** Write test → fail → implement → green → commit.
3. **No comments.** Self-documenting code. No doc comments, no inline comments.
4. **No over-engineering.** Simplest solution. No new deps without asking.
5. **Never panic.** Always return errors. `fatih/color` for output, not `fmt.Print`.
6. **Mock external deps.** Never launch real apps in tests.
7. **Run lint + test before committing.** Always.

### Commit Convention

```
<package>: <imperative verb> <what>
```

## Coverage Targets

| Package | Target | Current |
|---------|--------|---------|
| `config/` | ≥ 90% | 96.7% |
| `provider/` | ≥ 85% | 84.7% |
| `detector/` | ≥ 85% | 85.9% |
| `utils/` | ≥ 90% | 92.0% |
| `cmd/` | ≥ 70% | 70.5% |
| `tui/` | ≥ 60% | 68.3% |

## Architecture Notes

### Config paths
- Global config: `xdg.ConfigHome/nook/config.yaml` (macOS: `~/Library/Application Support/nook/config.yaml`)
- Default scan_paths: `~/.nook/workspaces`
- Workspaces created via `nook init` default to `~/.nook/workspaces/<name>/`
- Workspaces can also live in repos (auto-detected) or any custom scan path

### Providers
- Register via `init()` in each provider file using `provider.Register()`
- **VS Code**: opens folder directly. Generates `.code-workspace` file only when terminals are configured
- **DBeaver**: uses direct binary path + `-con` flag with pipe-delimited `key=value|key=value` format. Init builds this interactively (driver, host, port, db, user, pass)
- **Chrome**: opens all URLs in one window with multiple tabs. Auto-prepends `http://` for bare URLs. Platform-specific launch
- **Docker**: `docker compose -f <file> up -d` with optional `--profile`
- **Command**: `sh -c` / `cmd /c`, fire-and-forget
- Services launch in parallel via goroutines. Failures don't block others
- **Launch ordering** (optional): `order`, `delay_ms`, `ready_check` fields. Docker always first. Same order = parallel. Sequential groups. Readiness checks poll a user command
- VS Code terminals get `sleep N &&` prepended if the service has `delay_ms`

### TUI
- All `Select` prompts have type-to-filter enabled (case-insensitive substring)
- Provider prefixes on `init` prompts are dimmed gray (`[VS Code]`, `[DBeaver]`, etc.)
- Color palette: cyan bold (providers), dim italic (paths), green ✔ / red ✖ / yellow ⏳

### Init flow
1. Name, description, environment name, env file
2. Multi-select services → per-service prompts with provider prefix
3. DBeaver: confirm "Build connection interactively?" → list of drivers (type to filter)
4. Validation with `validator/v10`
5. Choose save location: Default (`~/.nook/workspaces`), Current directory, or pick from scan_paths
6. If not default location, auto-trusts the directory
7. Docker health check (optional)
8. Launch ordering (optional) → position per service → delay/readiness check per transition

### Testing patterns
- `mockPrompter` defined in `cmd/delete_test.go`, reused across all cmd tests
- `testCmd()` helper in `provider/testhelpers_test.go` for platform-safe exec mocking
- `setHomeEnv()` in `cmd/init_test.go` handles HOME/USERPROFILE on both platforms
- Skip Windows-incompatible tests: `t.Skip("chmod not supported on Windows")` etc.
- Permission tests (chmod 0000) must `t.Skip` on Windows

## Release

- Pushes formula to `lorenzo-vecchio/nook` at `Formula/nook.rb` (GoReleaser)
- Pushes scoop manifest to `lorenzo-vecchio/nook` at `nook.json`
- GitHub Release with binaries for all platforms
- Install: `brew tap lorenzo-vecchio/nook https://github.com/lorenzo-vecchio/nook && brew install --formula lorenzo-vecchio/nook/nook`
- CI in `.github/workflows/ci.yaml`, release in `.github/workflows/release.yaml`
