# 015 — Go + Next.js monorepo replaces Python monolith

**Status:** Accepted
**Date:** March 22, 2026
**Supersedes:** ADR 001

## Context

ADR 001 chose Python for its ML/NLP ecosystem: Presidio, sentence-transformers,
llm-guard, litellm. ADR 010 then removed all ML-based guardrails (prompt
injection detection, LLM judge, semantic rate limiting), eliminating the primary
reason for Python.

With the ML ecosystem no longer a factor, the language choice reopened. The
remaining requirements are:

- An HTTP server that proxies LLM calls and manages sessions (the security
  boundary)
- A web UI for the chat interface (replacing the Python TUI)
- PII scrubbing (deferred — approach TBD)

## Decision

Replace the Python monolith with two services in a monorepo:

- **`server/`** — Go HTTP server. Security boundary. Holds LLM keys, manages
  sessions, builds system prompts, runs deterministic filters, logs everything.
- **`web/`** — Next.js frontend. Pure presentation layer. Calls the Go server's
  REST+SSE API. Never touches LLM keys or session internals.

### Why Go for the server

- Single binary deployment. No runtime, no virtualenv, no dependency
  resolution at deploy time.
- Easy to containerise. Small Docker images (< 20 MB vs. hundreds of MB for
  Python + dependencies).
- Strong concurrency model for handling SSE streaming connections.
- The security boundary is clear by architecture: the Go binary is the only
  process with LLM API keys and network access to the LLM provider.
- `openai-go` (official OpenAI SDK) supports any OpenAI-compatible API via
  configurable base URL.

### Why Next.js for the frontend

- The original plan used Python's `textual` for a terminal UI. A web UI is
  easier to demo, share, and extend.
- Next.js App Router provides a solid foundation with server-side rendering,
  API routes (if needed for BFF patterns later), and a large ecosystem.
- Native SSE consumption via `fetch` + `ReadableStream` in the browser.

### Why a monorepo

- The Go server and Next.js app are tightly coupled by the API contract.
  Changes to the API often require changes to both. A single repo makes this
  atomic.
- Shared `scenarios/` YAML files and `docs/` directory.
- Simpler CI — one pipeline builds and tests both.

### What happens to the Python code

The existing `src/chatbot/`, `tests/`, and Python configuration files are
deleted. They served as a prototype for the conversation manager and session
management logic, which is now ported to Go. The design decisions and ADRs
from the Python phase remain valid and are carried forward.

## Consequences

- ADR 001 (Python as implementation language) is superseded. The reasoning
  about ML ecosystem access no longer applies.
- The Go server must reimplement session management, system prompt construction,
  and deterministic filters from scratch. The Python prototype informs the
  design but is not reused.
- PII scrubbing (Presidio) was Python-native. The Go server will need either a
  Go-native PII library, a Presidio sidecar container, or to defer PII
  scrubbing. This decision is deferred to stage 10.
- Developers now need Go and Node.js toolchains. This is a broader set of
  prerequisites than Python-only.
