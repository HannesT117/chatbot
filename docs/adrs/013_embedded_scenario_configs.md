# 013 — Embed scenario configs in the Go binary

**Status:** Accepted
**Date:** March 22, 2026

## Context

Scenario configs (financial_advisor, brand_marketing, insurance_claims) define
personas, constraints, allowed intents, blocklist terms, and other
scenario-specific settings. The Go server needs to load these at startup.

Two options were considered:

**Read YAML files at runtime.** Reuse the existing `scenarios/*.yaml` files.
The server reads them from disk at startup. Scenarios can be changed without
rebuilding.

**Embed via `//go:embed`.** Bake the YAML files into the Go binary at build
time. Single binary deployment with no file path dependencies.

## Decision

Embed via `//go:embed`.

### Why embed

- Single binary deployment. The server has no runtime dependency on file paths,
  working directories, or mounted volumes. Copy the binary anywhere and it
  runs.
- Scenario configs change rarely. They define personas and constraints that are
  part of the application's design, not runtime configuration.
- `//go:embed` is a standard library feature with zero overhead. The YAML files
  are small (< 1 KB each).

### Why not runtime file loading

- File path resolution is fragile. The server would need to know where the
  `scenarios/` directory is relative to its working directory, or accept a
  flag/env var pointing to it.
- For a testbed, the ability to change scenarios without rebuilding is not a
  meaningful benefit. Rebuilding takes seconds.

## Consequences

- Adding or modifying a scenario requires rebuilding the Go binary. This is
  acceptable for a testbed.
- If runtime-configurable scenarios become necessary (e.g., customer-specific
  deployments), a file-based loader can be added alongside the embedded
  defaults — the embedded configs serve as fallbacks.
- The `scenarios/*.yaml` files remain in the repository root and are shared
  between documentation and the Go binary.
