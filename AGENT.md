# AGENT instructions for sfDBTools

Purpose: short, actionable checklist for an AI code agent making changes to this repository.

1. Read these files first (order matters):
   - `main.go` (entry + config validation)
   - `internal/config/*` (Viper loader, `Get()`, `GetBackupDefaults()`)
   - `cmd/root_cmd.go` and the `cmd/` subfolders for examples of Cobra commands
   - `utils/backup/` and `internal/logger` for helpers and logging patterns

2. Small-change contract (for each PR):
   - Inputs: which flags/config keys are added or changed; expected env overrides.
   - Outputs: CLI behavior, config files written, and log messages.
   - Error modes: missing config (`/etc/sfDBTools/config/config.yaml`), DB connect failures, missing external binary.

3. Editing rules and patterns:
   - Register flags only in `init()` of the command file under `cmd/<command>/`.
   - Use existing helpers: `backup_utils.AddCommonBackupFlags`, `backup_utils.ResolveBackupConfigWithoutDB`, `logger.Get()`.
   - Wrap errors: `fmt.Errorf("...: %w", err)`.
   - Use `config.Get()` or `config.LoadConfig()` instead of reading files directly.
   - Avoid changing system-default paths unless tests and an explicit override are provided.

4. Build & test fast checks (run after edits):
   - Build: `go build ./...` (or `go build -o sfdbtools main.go`)
   - Unit tests: `go test ./...`
   - Smoke-run an example command: `go run main.go config validate` and confirm exit code

5. Safety & review triggers (ask human):
   - Any change to `/etc/sfDBTools/*`, `/var/log/sfDBTools/*`, or install/release scripts.
   - Any change affecting encryption, backup format, or checksum computation.
   - Adding external binaries or changing their invocations (e.g., mysqldump, rsync).

6. Useful examples in repo:
   - Flag helpers and command execution: `cmd/backup/backup_cmd_user.go`
   - Encrypted DB config generation: `cmd/dbconfig/config_generate.go`
   - Config loader and env bindings: `internal/config/loader.go`, `env.go`

7. If a task is under-specified, assume:
   - Config lives at `/etc/sfDBTools/config/config.yaml` (unless a flag overrides it).
   - Environment overrides available via viper with prefix `sfDBTools` (use `SFDBTOOLS_...`-style names when automating).

8. Deliverables for non-trivial changes:
   - Source edits, a small unit test (happy path + one edge case), updated README or help text, and verification that `go build` and `go test` pass locally.

Keep changes minimal and follow existing folder patterns. Ask for human review when in doubt.
