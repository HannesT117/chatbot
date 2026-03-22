# Go + Next.js Architecture — Design Spec

## Overview

This spec describes the architectural pivot from a Python monolith to a
Go server + Next.js frontend. The Go server is the security boundary — it
holds LLM API keys, manages sessions, builds system prompts, runs
deterministic filters, and logs everything. The Next.js app is a pure
presentation layer that consumes the Go server's REST+SSE API.

This pivot was motivated by two changes:

1. **ADR 010** removed all ML-based guardrails (prompt injection detection,
   LLM judge, semantic rate limiting), eliminating the Python ML ecosystem as
   the primary reason for using Python.
2. A Go server provides a cleaner security boundary (single binary, easy
   containerisation, no tool access by design), and Next.js provides a real
   web UI instead of a terminal TUI.

## Architecture

```
Browser → Next.js (web/) → Go server (server/) → LLM provider API
                                 │
                                 ├── Session store (in-memory)
                                 ├── System prompt (Go template, markdown)
                                 ├── Input filters (blocklist)
                                 ├── Output filters (canary, blocklist)
                                 └── Observability (slog structured logging)
```

**Monorepo layout:**

```
server/           # Go module
web/              # Next.js app
scenarios/        # YAML configs (embedded into Go binary via go:embed)
docs/             # ADRs, specs, plans
```

The existing Python code (`src/`, `tests/`) is deleted. It served as a
prototype; the Go server replaces it.

## Go Server

### API (REST + SSE)

| Method | Path | Purpose |
|--------|------|---------|
| `POST /api/chat` | Send a message, receive streamed response via SSE |
| `POST /api/sessions` | Create a new session (body: `{"scenario": "financial_advisor"}`), returns session ID + scenario info |
| `DELETE /api/sessions/:id` | End a session |
| `GET /api/scenarios` | List available scenarios |
| `GET /health` | Health check |

#### `POST /api/chat`

Request:
```json
{
  "session_id": "abc123",
  "message": "What savings accounts do you offer?"
}
```

SSE response stream:
```
data: {"type": "token", "content": "We"}
data: {"type": "token", "content": " offer"}
data: {"type": "token", "content": " three"}
...
data: {"type": "done", "turn_count": 3, "turns_remaining": 17}
```

If a filter blocks the input (pre-LLM):
```
data: {"type": "blocked", "reason": "blocklist", "fallback": "I'm not able to help with that."}
```

If an output filter catches a violation after tokens have been streamed:
```
data: {"type": "blocked", "reason": "canary_leak", "fallback": "I'm not able to help with that."}
```

**SSE blocked-event behavior:** Tokens are streamed as they arrive from the
LLM. Output filters (canary, blocklist) run on the accumulated full response
after the stream completes. If an output filter triggers, the server sends a
`blocked` event as the terminal event. The client must treat all streamed
tokens as provisional until a `done` or `blocked` terminal event arrives. On
`blocked`, the client discards the streamed content and displays the fallback
message instead. This is a known UX trade-off: the user may briefly see
response text that is then replaced. Acceptable for a testbed.

#### CORS

During development, the Next.js app (port 3000) and Go server (different port)
are on different origins. The Go server must set `Access-Control-Allow-Origin`
for the Next.js origin. In production, a reverse proxy can serve both on the
same origin, eliminating CORS entirely.

### Internal package structure

```
server/
  cmd/server/main.go       # Entry point, config, wiring
  internal/
    api/                    # HTTP handlers, SSE streaming
    session/                # Session struct, SessionStore interface, in-memory impl
    prompt/                 # System prompt template rendering
    filter/                 # Input blocklist, output blocklist, canary detection
    scenario/               # ScenarioConfig, go:embed loader
    llm/                    # openai-go client wrapper, streaming adapter
    observability/          # Structured logging (slog)
  go.mod
  go.sum
```

### Session management

`Session` struct holds:
- `ID` — UUID string
- `CanaryToken` — `crypto/rand` generated hex token (32 chars)
- `Turns` — slice of `Turn{User, Assistant string}`
- `TurnCount` — historical total (never decrements after trimming)
- `ScenarioName` — which scenario this session uses

`SessionStore` interface:
```go
type SessionStore interface {
    Get(id string) (*Session, error)
    Save(session *Session) error
    Delete(id string) error
}
```

Initial implementation: `sync.Mutex`-protected `map[string]*Session`.

### System prompt

Built from a Go `text/template` using markdown headers:

```markdown
## Role

{persona_name}. {persona_description}

{persona_notes}

## Constraints

- {constraint_1}
- {constraint_2}

## Allowed Topics

- {topic_1}
- {topic_2}

## Blocklist

- {term_1}
- {term_2}

## Canary

{canary_token}
```

Per-scenario persona notes are defined as string constants or small template
fragments within the `prompt/` package.

### Scenario configs

The existing `scenarios/*.yaml` files are embedded into the Go binary via
`//go:embed`. Parsed at startup into typed `ScenarioConfig` structs.

### LLM integration

The `llm/` package wraps `openai-go` behind an interface:

```go
type Client interface {
    ChatStream(ctx context.Context, messages []Message) (<-chan StreamEvent, error)
}
```

This allows mocking in tests. The real implementation uses `openai-go` with
configurable base URL (supporting any OpenAI-compatible API: OpenAI, Anthropic,
Azure, Ollama, etc.).

### Deterministic filters

**Input blocklist:** Compiled regex patterns from scenario config. Runs before
the LLM call. On match: return a blocked response, log the event.

**Output canary detection:** Exact string match of `session.CanaryToken`
against the accumulated LLM response. On match: replace response with safe
fallback, log the event.

**Output blocklist:** Same compiled regex patterns as input, applied to the
accumulated response. On match: replace with safe fallback, log the event.

### Chat flow (one turn)

1. Parse request, look up session from `SessionStore`
2. Check turn limit → blocked response if exceeded
3. Run input blocklist → blocked response if hit
4. Build system prompt from scenario config + session canary token
5. Trim message history to token budget (drop oldest turns)
6. Call LLM via `openai-go`, stream tokens to client via SSE
7. Accumulate full response, run output filters (canary, blocklist)
8. If output filter fails → send safe fallback event, log
9. Record turn in session, log the full exchange
10. Send done event with turn metadata

### Observability

Go's `slog` for structured logging. Every request/response logged with:
session ID, turn number, scenario name, filter results, latency, model used.

Filter hits (blocklist, canary) are logged at WARN level for alerting.

## Next.js App

### Stack

- Next.js App Router
- React Server Components where applicable, client components for chat
- Tailwind CSS
- Native `fetch` with `ReadableStream` for SSE consumption
- React state for session management (no external state library)

### Pages

```
web/
  app/
    page.tsx                    # Scenario selection → create session
    chat/[sessionId]/
      page.tsx                  # Chat UI with streaming
  lib/
    api.ts                      # Typed fetch wrappers for Go server API
  components/
    ChatMessage.tsx             # Single message bubble
    ChatInput.tsx               # Input bar
    StreamingMessage.tsx        # Renders tokens as they arrive
```

### What the Next.js app does NOT know

- LLM API keys
- System prompt content or structure
- Session internals (canary token, message history)
- Filter logic or configuration
- Scenario constraints

It only knows the Go server's REST API shape.

## Testing

### Go server

- Unit tests for each `internal/` package
- HTTP handler tests using `httptest`
- LLM calls mocked via the `Client` interface
- `go test ./...` — no external dependencies

### Next.js

- Component tests with React Testing Library
- API client tests mocking Go server responses
- Lightweight — the frontend is thin

### Integration

- Gated behind `CHATBOT_LIVE_TESTS=1`
- Start Go server, hit with real requests against a real LLM
- Full flow: create session → send messages → verify streaming → verify filters

### Characterization

- Run against Go server API
- Categories: brand_compliance, prompt_leakage, regulatory
- Document known limitations, not proof of robustness

## Implementation stages

| Stage | What |
|-------|------|
| 3 | Delete Python code and config; update docs/commands.md and .gitignore for Go + Next.js |
| 4 | Go server scaffold: module, entry point, health endpoint, scenario loading via go:embed |
| 5 | Session management: struct, store interface, in-memory impl, turn limits, token budget, sliding window, canary tokens |
| 6 | System prompt: Go template, markdown headers, per-scenario persona notes, canary injection |
| 7 | LLM integration + SSE streaming + deterministic filters: openai-go wrapper, POST /api/chat with filter hooks, input blocklist, output blocklist, canary detection, CORS |
| 8 | Observability: slog structured logging, request/response logging, alerting on filter hits |
| 9 | Next.js app: chat UI, scenario selector, SSE consumption, handle provisional tokens and blocked events |
| 10 | PII scrubbing: approach TBD (Go-native, Presidio sidecar, or deferred) |
| 11 | Characterization tests: adversarial inputs against Go server API |

## ADRs

| ADR | Decision |
|-----|----------|
| 010 | Security threat model: infrastructure over LLM guardrails |
| 011 | REST + SSE for API transport |
| 012 | In-memory session storage with pluggable interface |
| 013 | Embed scenario configs via go:embed |
| 014 | Markdown headers for system prompt structure |
| 015 | Go + Next.js monorepo (replaces Python monolith) |
