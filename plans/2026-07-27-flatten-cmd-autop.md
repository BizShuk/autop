# Flatten `cmd/autop` Implementation Plan

**Goal:** Remove the redundant `cmd/autop` directory layer by moving the CLI package to `cmd` while preserving runtime behavior.

**Architecture:** Keep `cmd` as the `autop` command package and `cmd/driver` as the independent CLI-driver package. Update imports, embedded settings, PM2 target detection, tests, and documentation to use the new paths.

## Tasks

- [x] Move `cmd/autop/*.go` and `settings.example.json` to `cmd/`; move `cmd/autop/driver/` to `cmd/driver/`.
- [x] Update Go imports, `go:embed`, package tests, and path-sensitive wizard/install logic.
- [x] Update `README.md`, `CLAUDE.md`, `docs/terminology.md`, and the active plan references.
- [x] Run formatting, package tests, vet, build, and repository-wide tests; record unrelated baseline failures if present.
