# 003 — Async-first architecture

**Status:** Superseded by ADR 015
**Date:** March 20, 2026
**Superseded:** March 22, 2026

## Context

The pipeline is I/O-bound. LLM API calls take seconds. Presidio runs local
NLP inference. The semantic rate limiter runs embedding inference. The
conversation loop waits for user input.

Two implementation strategies were considered:

**Synchronous:** Simpler code, no async/await syntax. Straightforward to
reason about. Requires threads or multiprocessing for concurrency.

**Async (asyncio):** Native Python async/await. No threads needed for I/O
concurrency. Compatible with Textual (the TUI framework, which is async-native)
and with `litellm`'s async interface (`acompletion`).

The project also uses `pytest-asyncio`, which supports async test functions
directly.

## Decision

The entire pipeline is async. Each stage runs its steps sequentially —
order matters for security, and a later step must not run if an earlier step
rejects. LLM calls use `litellm.acompletion`. The conversation loop is an
async REPL. The TUI uses Textual's native async support.

Sync `input()` calls inside async functions are a known pitfall; they block
the event loop. Any user input in the CLI must use
`asyncio.get_event_loop().run_in_executor(None, input, prompt)` or an
equivalent non-blocking approach.

## Consequences

- All pipeline step interfaces use `async def`. Sync implementations must be
  wrapped with `run_in_executor` if they perform blocking I/O or heavy CPU
  work.
- `pytest-asyncio` with `asyncio_mode = "auto"` handles async test functions
  without boilerplate.
- Textual integration is natural; no threading bridge is needed.
- Blocking calls inside async functions (for example, `input()`, or
  synchronous ML inference) will stall the event loop. Each such call must be
  explicitly wrapped or moved to an executor.
- CPU-bound steps (ML inference) may benefit from `ProcessPoolExecutor` in a
  production deployment to avoid blocking the event loop, but this is deferred
  until profiling shows it's necessary.

## Superseded — March 22, 2026

This ADR is superseded by ADR 015 (Go + Next.js monorepo replaces Python
monolith).

The Python asyncio architecture is no longer applicable. The Go server uses
goroutines for concurrency (native to the language, no async/await syntax
needed). The principle — sequential pipeline execution where order matters —
carries forward, but the implementation is entirely different.
