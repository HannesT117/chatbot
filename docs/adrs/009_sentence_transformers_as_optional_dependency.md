# 009 — sentence-transformers as an optional dependency

**Status:** Accepted
**Date:** March 20, 2026

## Context

The semantic rate limiter (step 12) uses `sentence-transformers` with the
`all-MiniLM-L6-v2` model to compute cosine similarity between denied inputs
across turns. `sentence-transformers` depends on PyTorch, which adds
approximately 1.5 GB to the install footprint. The model itself adds another
~80 MB.

For a developer who only wants to work on deterministic pipeline steps (schema
validation, blocklists, canary detection), requiring PyTorch as a core
dependency creates significant friction: slow installs, large Docker images,
and CI overhead — all for a feature that is step 12 of 13.

## Decision

Declare `sentence-transformers` in an optional dependency group (`ml`) rather
than as a core dependency. Install it with `uv sync --extra ml` when needed.

The semantic rate limiter degrades gracefully when the dependency is missing:
it disables itself and emits a structured log warning at startup. The pipeline
continues to function without semantic rate limiting.

## Consequences

- Default installs are lightweight. Developers working on steps 1–11 don't
  need PyTorch.
- The `ml` group must be explicitly installed for step 12 development and for
  any deployment where semantic rate limiting is required.
- CI runs without the `ml` group by default. A separate CI job or manual step
  is needed to test the semantic rate limiter.
- The graceful degradation means a misconfigured deployment (missing `ml`
  group) silently loses semantic rate limiting rather than crashing. The log
  warning at startup is the only signal. Monitoring must check for this warning
  in production deployments.
