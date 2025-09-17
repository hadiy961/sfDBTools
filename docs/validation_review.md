# Validation Code Review — internal/core/mariadb/configure/validation.go

Date: 2025-09-17

This document summarizes a review and proposed refactor of `internal/core/mariadb/configure/validation.go`.

## Summary

The file implements `validateSystemRequirements` and a set of helper functions used during MariaDB configuration validation:

- validateSystemRequirements
- validateDirectories
- validatePort
- validateEncryptionKeyFile
- validateDiskSpace
- validateDirectoryPermissions

Helpers:
- ensureDirectoryExists
- checkDirectoryWritable
- isPortInUse
- getDiskFreeSpace
- checkMySQLUserAccess
- fixDirectoryPermissions

Overall assessment:
- Functionality is correct for the domain (filesystem, port and keyfile checks).
- The file is a bit large (~340 lines) and mixes higher-level orchestration with low-level filesystem and OS operations. Splitting improves readability and testability.
- There are overlaps with existing utility packages in the repo (not duplicates exactly, but similar functionality exists in `utils/disk`, `utils/system`, and `utils/fs`). Prefer reusing those shared packages where possible.
- A few helpers are reusable across the codebase (`ensureDirectoryExists`, `checkDirectoryWritable`), but many behaviors (e.g., `fixDirectoryPermissions` specific mysql UID/GID logic) are MariaDB-specific and belong in this package.

## Duplicate / Overlap findings

I searched the repo for similar functions and found overlapping functionality already provided elsewhere:

- Disk space and free bytes:
  - `utils/disk.GetFreeBytes`, `utils/disk.CheckDiskSpace` — use these instead of reimplementing disk checks.
  - `validation.go` calls `disk.GetFreeBytes` already for disk checks (good reuse).

- Port checks and validation:
  - `utils/system.CheckPortConflict`, `utils/system.IsPortAvailable`, `utils/system.ValidatePortRange` — used by `validation.go` already. `validatePort` uses `system.CheckPortConflict` and falls back to `system.IsPortAvailable` (good reuse).

- Filesystem permission utilities:
  - `utils/fs/dir/permission_unix.go` and `utils/fs/file/file_permission.go` include `Chown`/`Chmod` helpers and fallbacks. `fixDirectoryPermissions` does manual `os.Chown`/`os.Chmod` and a `/etc/passwd` parse; consider delegating to `utils/fs` helpers for portability, testing, and consistency.

- Interactive validators:
  - `internal/core/mariadb/configure/interactive/validators.go` contains `ValidatePortRange` (related but not duplicate).

No exact function duplicates were found (same name + same logic) except that `validation.go` already uses `disk.GetFreeBytes` and `system.IsPortAvailable/CheckPortConflict` from `utils` packages.

## Proposed split

Create a `validation` subpackage under `internal/core/mariadb/configure/validation/` with multiple small files. Keep the package name `configure` or choose `validation` to avoid import churn — I recommend `validation` under the same module path (`internal/core/mariadb/configure/validation`) so other files in `configure` can import `validation` while keeping functions scoped.

Suggested files and responsibilities:

- `validation.go` (high-level orchestration)
  - `ValidateSystemRequirements(ctx, cfg)` exported (rename from `validateSystemRequirements`) if used elsewhere; otherwise keep unexported.
  - calls into other sub-modules.

- `directories.go`
  - `validateDirectories(cfg)`
  - `ensureDirectoryExists(dir)` (could be exported to `utils/fs` in future)
  - `checkDirectoryWritable(dir)`

- `port.go`
  - `validatePort(port)` (wraps `utils/system` functions)
  - `isPortInUse(port)` helper (thin wrapper around `system.IsPortAvailable`)

- `encryption.go`
  - `validateEncryptionKeyFile(path)`

- `disk.go`
  - `validateDiskSpace(cfg)` (uses `utils/disk`) — already minimal

- `permissions.go`
  - `validateDirectoryPermissions(cfg)`
  - `checkMySQLUserAccess(dir)`
  - `fixDirectoryPermissions(dir)`

This split isolates OS-level operations and makes unit testing small parts easier.

## Exported vs internal

- Keep most helpers unexported (lowercase) unless other packages need them.
- Consider exporting `ValidateSystemRequirements` (capitalized) if it should be called by `RunMariaDBConfigure` (currently in same package). If kept in a new package `validation`, export the top-level function so `configure.RunMariaDBConfigure` can call `validation.ValidateSystemRequirements`.

## Reusability suggestions

- Move generic filesystem helpers (`ensureDirectoryExists`, `checkDirectoryWritable`) to `utils/fs` if used outside this package.
- Use `utils/fs` for `Chown/Chmod` with proper fallback (already at `utils/fs/dir/permission_unix.go`). This improves portability and centralizes permission handling and logging.
- `fixDirectoryPermissions` contains MariaDB-specific owner defaults (uid 992/gid 991). Keep this function within the MariaDB configure package but consider making it call a `utils/fs` helper to perform the actual chown after resolving uid/gid.

## Over-engineering notes

- The code attempts to be defensive (creates directories, tests write permissions by creating files, fallback for port checks). This is generally fine for a tool that modifies system configuration.
- Small over-engineering signs:
  - Parsing `/etc/passwd` by hand for UID/GID could be replaced with `os/user` package (though `os/user` has limitations in statically-linked or certain platforms); the manual parse is a pragmatic fallback but should be documented.
  - `checkMySQLUserAccess` logic is simplistic: it treats root-owned directories as inaccessible and flags them. In some deployments, `root` owning a directory and granting group access to `mysql` could be acceptable. Consider improving the check to look for effective permission for the `mysql` user instead of UID equality only.

## Edge cases and test considerations

- Non-Linux platforms: `os.Chown` and parsing `/etc/passwd` may behave differently. Use `utils/fs` abstraction that already has platform-specific files.
- Running as non-root: some operations (Chown) will fail — functions already return errors; ensure caller handles them gracefully.
- Symbolic links and mountpoints: `ensureDirectoryExists` and disk checks walk up the path; be careful with symlinks.

## Next steps (recommended)

1. Create new package `internal/core/mariadb/configure/validation` and split the file into the suggested smaller files. Expose only `ValidateSystemRequirements` (or keep package local and update import sites accordingly).
2. Replace manual permission changes with calls to `utils/fs` where appropriate.
3. Add unit tests for each helper and for `ValidateSystemRequirements` (happy path + 2 error cases: insufficient disk, port in use by non-mysql process).
4. Run `go build ./...` and fix any import changes.

---

If you want, I can implement the split now (create new files, adjust imports, and run a build). I will not change behavior except to move code and update package names; I can then optionally extract generic helpers to `utils/fs` in a follow-up.
