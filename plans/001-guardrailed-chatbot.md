# Plan: Guardrailed chatbot

> **Revised March 22, 2026 — Architecture pivot.**
>
> Two major changes from the original plan:
>
> 1. **Security model.** ADR 010 removed all LLM-based guardrails (prompt
>    injection detection, LLM judge, semantic rate limiting, confidence scoring).
>    Research on adaptive attacks showed these are fundamentally unreliable for a
>    chatbot without tool access. The focus is now brand compliance, deterministic
>    filters, infrastructure isolation, and observability.
>
> 2. **Tech stack.** ADR 015 replaced the Python monolith with a Go server
>    (security boundary) + Next.js frontend (presentation layer). The ML
>    ecosystem that motivated Python is no longer needed.

## Architecture overview

```
Browser → Next.js (web/) → Go server (server/) → LLM provider API
                                 │
                                 ├── Session store (in-memory)
                                 ├── System prompt (Go template, markdown)
                                 ├── Input filters (blocklist)
                                 ├── Output filters (canary, blocklist)
                                 └── Observability (slog structured logging)
```

The Go server is the security boundary. It holds LLM API keys, manages
sessions, builds system prompts, runs deterministic filters, and logs
everything. The Next.js app is a pure presentation layer that consumes the
Go server's REST+SSE API.

## What was removed and why

Each cut component is documented here so the reasoning is explicit and
traceable. See the README's "Deferred techniques" table for conditions under
which each should be revisited.

**Prompt injection detection** — All detection approaches are fundamentally
unreliable against adaptive attacks. No tool access means no blast radius
beyond brand damage. See ADR 010.

**LLM-as-judge output validation** — A second LLM has the same failure modes
as the primary model. See ADR 010.

**Confidence/hedging scoring** — Hedging language does not correlate with
policy compliance. See ADR 010.

**Semantic rate limiting** — Required PyTorch (~1.5 GB) for a narrow use case.
Turn limits handle conversation-length abuse. See ADR 010.

**Multi-model judge** — Same problem as single-model judge, doubled cost. See
ADR 010.

**Action allowlist** — No actions to allowlist. No tool calls, no code
execution. See ADR 010.

**Stateful attack tracking / circuit breakers** — Over-engineered for a
text-only chatbot. Turn limits suffice. See ADR 010.

**Retry loop** — Assumes guardrails are reliable enough to retry against. With
deterministic-only filters, a blocklist hit replaces the response with a safe
fallback. See ADR 010.

**Intent classifier and tool router** — No tools to route to. See ADR 010.

**Output schema validation** — Free-text responses have no schema to validate.

**Python codebase** — Replaced by Go + Next.js. See ADR 015.

## Monorepo structure

```
server/                          # Go module — security boundary
  cmd/server/main.go             # Entry point, config, wiring
  internal/
    api/                         # HTTP handlers, SSE streaming
    session/                     # Session struct, SessionStore interface, in-memory impl
    prompt/                      # System prompt template rendering
    filter/                      # Input blocklist, output blocklist, canary detection
    scenario/                    # ScenarioConfig, go:embed loader
    llm/                         # openai-go client wrapper, streaming adapter
    observability/               # Structured logging (slog)
  go.mod
  go.sum
web/                             # Next.js app — presentation layer
  app/
    page.tsx                     # Scenario selection
    chat/[sessionId]/page.tsx    # Chat UI with streaming
  lib/api.ts                     # Typed fetch wrappers for Go server API
  components/                    # ChatMessage, ChatInput, StreamingMessage
scenarios/                       # YAML configs (embedded into Go binary)
  financial_advisor.yaml
  brand_marketing.yaml
  insurance_claims.yaml
docs/                            # ADRs, specs, plans
  adrs/
  superpowers/specs/
  superpowers/plans/
```

## Go server API

| Method | Path | Purpose |
|--------|------|---------|
| `POST /api/chat` | Send a message, receive streamed response via SSE |
| `POST /api/sessions` | Create a new session |
| `DELETE /api/sessions/:id` | End a session |
| `GET /api/scenarios` | List available scenarios |
| `GET /health` | Health check |

See the [architecture spec](docs/superpowers/specs/2026-03-22-go-nextjs-architecture-design.md)
for request/response formats and the full chat flow.

## Components

### 1. Input pipeline (deterministic)

| Step | What it does |
|------|-------------|
| Schema validation | Reject malformed input, enforce max length |
| Keyword/regex blocklist | Catch known forbidden terms and patterns |

PII scrubbing is deferred to stage 10 (approach TBD).

### 2. Conversation manager

- **Turn limits** — end conversation after N turns
- **Token budget** — sliding window, drop oldest turns (no summarisation per
  ADR 006)
- **Session isolation** — per-session state, no cross-session leakage
- **Canary tokens** — `crypto/rand` hex token per session, embedded in system
  prompt

### 3. System prompt

Go `text/template` with markdown headers. Sections: Role (persona + notes),
Constraints, Allowed Topics, Blocklist, Canary. Per-scenario persona notes
provide tone and escalation guidance. See ADR 014.

### 4. Output pipeline (deterministic)

| Step | What it does |
|------|-------------|
| Canary leak detection | Exact string match of session canary in output |
| Keyword/regex blocklist | Same patterns as input, applied to output |

On failure: replace response with safe fallback, log the event.

### 5. Observability

Go `slog` structured logging. Every request/response logged with: session ID,
turn number, scenario name, filter results, latency, model used. Filter hits
logged at WARN level.

### 6. Next.js frontend

Chat UI with SSE streaming, scenario selector. Pure presentation layer — no
knowledge of LLM keys, system prompts, or filter logic.

## Test strategy

**Go server:** Unit tests per `internal/` package. HTTP handler tests via
`httptest`. LLM calls mocked via interface. `go test ./...` — no external
dependencies.

**Next.js:** Component tests with React Testing Library. API client tests
mocking Go server responses.

**Integration:** Gated behind `CHATBOT_LIVE_TESTS=1`. Full flow against a
real LLM.

**Characterization:** Adversarial inputs against Go server API. Documents
known limitations, not proof of robustness. Categories: brand_compliance,
prompt_leakage, regulatory.

## Security principles

1. **Infrastructure is the security boundary.** No tool access, no code
   execution, no database access. Text only.
2. **Deterministic over probabilistic.** Regex blocklists and string matching
   are the only automated filters. No LLM-based classifiers.
3. **The system prompt is brand compliance, not security.** It handles normal
   interactions but will be bypassed by creative prompting.
4. **Observability over prevention.** Logging and alerting catch what filters
   miss. No false confidence from automated guardrails.
5. **Fail closed.** Pipeline errors reject the request with a safe fallback.
6. **No sensitive data in context.** Nothing harmful if the system prompt leaks.
7. **Principle of least privilege.** Isolated container, non-root user, scoped
   API keys.

## Tech stack

- **Go server:** Go, `openai-go`, `slog`
- **Frontend:** Next.js (App Router), React, Tailwind CSS
- **Scenarios:** YAML, embedded via `//go:embed`
- **Testing:** `go test`, React Testing Library
- **PII detection:** TBD (stage 10)

## Implementation stages

| Stage | What | Status |
|-------|------|--------|
| 1 | Python scaffold | Done (to be deleted in stage 3) |
| 2 | Python conversation manager | Done (to be deleted in stage 3) |
| 3 | Delete Python code and config: remove `src/`, `tests/`, `pyproject.toml`, `.python-version`, `uv.lock`; update `docs/commands.md` and `.gitignore` for Go + Next.js | Done |
| 4 | Go server scaffold: module, entry point, health endpoint, scenario loading via `go:embed` | Done |
| 5 | Session management: struct, store interface, in-memory impl, turn limits, token budget, sliding window, canary tokens | Done |
| 6 | System prompt: Go template, markdown headers, per-scenario persona notes, canary injection | Done |
| 7 | LLM integration + SSE streaming + deterministic filters: openai-go wrapper, `POST /api/chat` with filter hooks, input blocklist, output blocklist, canary detection, CORS | Done |
| 8 | Observability: slog structured logging, request/response logging, alerting on filter hits | Done |
| 9 | Next.js app: chat UI, scenario selector, SSE consumption, handle provisional tokens and `blocked` events | Done |
| 10 | PII scrubbing: approach TBD (Go-native, Presidio sidecar, or deferred) | Pending |
| 11 | Characterization tests: adversarial inputs against Go server API | Pending |

## ADRs

| ADR | Decision |
|-----|----------|
| 001 | Python as implementation language (superseded by 015) |
| 002 | Deterministic-first pipeline ordering (narrowed by 010) |
| 003 | Async-first architecture |
| 004 | litellm for LLM abstraction (superseded by 015 — now openai-go) |
| 005 | Fail-closed pipeline error handling |
| 006 | Sliding window over LLM summarisation |
| 007 | Swappable prompt injection detector (superseded by 010) |
| 008 | Test strategy: mocked unit, gated integration |
| 009 | sentence-transformers as optional dependency (superseded by 010) |
| 010 | Security threat model: infrastructure over LLM guardrails |
| 011 | REST + SSE for API transport |
| 012 | In-memory session storage with pluggable interface |
| 013 | Embed scenario configs via go:embed |
| 014 | Markdown headers for system prompt structure |
| 015 | Go + Next.js monorepo (replaces Python monolith) |
