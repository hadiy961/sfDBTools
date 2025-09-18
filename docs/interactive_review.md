# Review: `internal/core/mariadb/configure/interactive` package

This document reviews the `interactive` package located at `internal/core/mariadb/configure/interactive`.
It lists duplicated functions found elsewhere in the repository, functions that can be made reusable, places where code appears over-engineered, duplicated logic, and suggested refactors with priority and risks.

## Files inspected
- `collectors.go`
- `defaults.go`
- `display.go`
- `gather.go`
- `interactive.go`
- `validators.go`

## Summary of findings (high level)
- The package is well-factored into small responsibilities (defaults, collectors, gathering helpers, validators, and display).
- There are some duplicated validators and overlapping validation logic with other packages in the repository.
- Several helper functions are good candidates to be promoted to shared `utils/` packages (or existing util packages), especially path and port validation and memory-size parsing.
- A few places contain slightly redundant logic (e.g., repeated resolution of defaults, single-use wrapper functions) that can be simplified.

## Exact duplications / overlaps found
1. ValidatePortRange
   - Found in `internal/core/mariadb/configure/interactive/validators.go` and `utils/system/port.go`.
   - Differences: The `interactive` version enforces `1024 <= port <= 65535`. The `utils/system` version is more verbose and performs checks in multiple steps but effectively enforces similar constraints (it accepts >0 then separate checks). The `utils/system` function also includes user-facing messages in Indonesian and is used in networking helpers such as `FindAvailablePort`.
   - Recommendation: Consolidate to a single `utils/system` exported validator (e.g., `system.ValidatePortRange`) and have interactive call that. Keep the authoritative logic where `FindAvailablePort` and port scanning live (`utils/system`). Remove duplicate from `interactive/validators.go`.

2. Absolute path validation
   - `ValidateAbsolutePath` in `interactive/validators.go` is a very small function that wraps `filepath.IsAbs`.
   - Overlap: There are other directory/path validation utilities in `utils/fs/dir/*` and generic `dir.Validate` helpers that perform platform-specific and permission checks.
   - Recommendation: Use `utils/fs/dir.ValidatePath` or a small exported helper in `utils/fs/dir` that only checks absoluteness (if you really need that simple check). If the interactive flow only needs to ensure absolute paths syntactically, keep a tiny wrapper in a shared `utils` package; otherwise prefer calling existing `dir.Validate` which also checks accessibility and writability, giving better UX.

3. Server ID, BufferPool, Memory size validation
   - `ValidateServerIDRange`, `ValidateBufferPoolInstances`, `ValidateMemorySize` exist in `interactive/validators.go`.
   - Partial overlap: `utils/mariadb/config/validation.go` implements `validateConfigureInput` which performs similar checks (server_id range, port range, absolute paths, and additional cross-field checks like directories must differ).
   - Recommendation: Keep low-level validations (e.g., numeric ranges, memory-size parsing) as small reusable functions in a package like `utils/validation` or `utils/mariadb/validate` and have `validateConfigureInput` call them. Avoid duplicating ranges between files; centralize range constants (e.g., minPort, maxPort, minServerID, maxServerID) in a single file for consistency.

## Reusable functions / promotion opportunities
- InputCollector
  - This is a small, useful abstraction for interactive prompts and should remain. Consider moving it to `utils/interaction` or `utils/terminal` to reuse for other interactive flows (backup, restore, dbconfig), because many commands use `terminal.Ask*` directly — promoting `InputCollector` would DRY repeated prompt patterns (default resolution + validator invocation).
- ConfigDefaults
  - The `ConfigDefaults` helper resolves defaults from Template -> Current installation -> AppConfig -> hardcoded. This pattern is generic and could be reused for other configuration workflows. Consider extracting a small interface and helper to a shared package, or at least ensure the type is limited to mariadb-specific fields but keep the resolution logic reusable.
- Validators
  - Extract numeric range checks and memory-size check to `utils/validation` or `utils/mariadb/validation` and export them. This reduces duplication with `utils/mariadb/config/validation.go` and `internal/core/mariadb/configure/validation/*`.
- RequestUserConfirmation
  - `RequestUserConfirmation` is generic and can be re-used. Consider moving to `utils/terminal` as `ConfirmAction` or similar. Currently there is also a wrapper `RequestUserConfirmationForConfig` which only adds logging and context cancellation; that wrapper is fine to remain in the interactive package.

## Over-engineered code / single-use wrappers
- `RequestUserConfirmationForConfig` simply wraps `RequestUserConfirmation` adding logger/context handling. This is okay for clarity, but if many packages will need similar 