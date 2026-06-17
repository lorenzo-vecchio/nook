# Implementation Plan

Each task: write tests → implement → verify → commit. See `AGENTS.md` for rules.

---

## Phase 1: Foundation

1. `go mod init github.com/<org>/nook`, add all deps
2. Create directory structure (`cmd/`, `config/`, `provider/`, `tui/`, `detector/`, `utils/`)
3. Write `main.go` skeleton (import cmd package, call Execute)

## Phase 2: Config Package

4. Define config structs: `GlobalConfig`, `WorkspaceConfig`, `Environment`, `Service`, `Terminal`
5. Global config load/save (`~/.nook/config.yaml`) via viper
6. Workspace YAML parser (load/save `workspace.yaml`)
7. Struct validation with `validator/v10` — all edge cases
8. Env var resolver (`${VAR}` from os.Environ + `.env` files)

## Phase 3: Utils Package

9. Path resolution (relative → absolute, tilde expansion)
10. Platform detection helpers (isMacOS, isLinux, isWindows, editor fallback)

## Phase 4: Provider Package

11. `Provider` interface + registry (`Register`, `Get`, `List`)
12. VS Code provider (detect, launch plain, launch with terminals → generate `.code-workspace`)
13. DBeaver provider (detect, launch with `-con` flag, platform-specific paths)
14. Chrome provider (detect, open URLs, platform-specific commands)
15. Docker provider (detect, `docker compose up -d` with optional profile)
16. Command provider (always available, `sh -c` / `cmd /c`, fire-and-forget)

## Phase 5: Detector Package

17. Auto-detect `workspace.yaml` in cwd
18. Auto-detect `workspace.yaml` in immediate subdirectories
19. Trust prompt + `scan_paths` persistence

## Phase 6: TUI Package

20. Color/style helpers (cyan bold, dim italic, green/red/yellow marks)
21. Survey prompts (select, multi-select, input, confirm) with defaults
22. Formatted output helpers (progress lines, success/error summaries)

## Phase 7: Commands

23. Root command (`nook`) — cobra root, global flags, scan_paths loading
24. `nook detect` — iterate providers, show detection table
25. `nook init` — interactive workspace creation walkthrough
26. `nook open` — workspace selection, env selection, parallel provider launch
27. `nook list` — tabular listing grouped by scan_path
28. `nook edit` — open workspace.yaml in $EDITOR
29. `nook delete` — confirm, delete from ~/.nook/workspaces/
30. `nook scan` — rescan all paths, remove stale entries

## Phase 8: CI/CD & Distribution

31. `.goreleaser.yaml`
32. Homebrew formula (`Formula/nook.rb`)
33. Scoop manifest (`scoop/nook.json`)
34. Winget manifest (`winget/nook.yaml`)
35. `.github/workflows/ci.yaml` (active)
36. `.github/workflows/release.yaml` (disabled, ready to activate)

## Phase 9: Final Integration

37. Wire all commands in `main.go`
38. Cross-platform smoke test (build on each OS)
39. `README.md` with install + usage instructions
