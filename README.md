# LLM guardrails testbed

A testing ground for keeping LLM-powered chatbots on-brand, on-topic, and
compliant in regulated industries — without relying on LLM-based guardrails
that don't work.

## The problem

If you run a chatbot for a law firm, it shouldn't hand out cookie recipes. If
you run one for a bank, it shouldn't give investment advice. If you run one for
an insurance company, it shouldn't promise claim outcomes. Keeping an LLM
focused on its job and representing the brand correctly is a real problem.

The AI security industry sells LLM-based guardrails — prompt injection
detectors, LLM-as-judge validators, confidence scorers — as the solution.
These don't work. Research shows that guardrails self-reporting 99%
effectiveness are "all extremely breakable" under adaptive attack (Schulhoff,
2026). You cannot patch a brain the way you patch a software bug.

## Architecture

```
Browser → Next.js (web/) → Go server (server/) → LLM provider API
```

**Go server** (`server/`) is the security boundary. It holds LLM API keys,
manages sessions, builds system prompts, runs deterministic filters, and logs
everything. Single binary via `go build`.

**Next.js app** (`web/`) is a pure presentation layer. It renders the chat UI
and consumes the Go server's REST+SSE API. It never touches LLM keys, system
prompts, or session internals.

Scenario configs (`scenarios/`) are embedded in the Go binary via `//go:embed`.

## Our approach

Instead of building elaborate LLM-based defenses, we invest in what actually
works:

1. **A well-crafted system prompt** that tells the LLM exactly what persona to
   adopt, what topics are allowed, what constraints to follow, and what terms to
   avoid. This is brand compliance, not security — it handles the vast majority
   of normal interactions correctly.

2. **Deterministic filters** before and after the LLM call. Regex blocklists
   catch known forbidden terms. Canary tokens detect system prompt leakage.
   These are cheap, fast, and cannot be argued around.

3. **Infrastructure-level security** as the real defense boundary. The LLM has
   no tool access, no code execution, no database queries. Even a fully
   jailbroken instance can only produce text — it cannot exfiltrate data or take
   destructive actions.

4. **Observability** to catch what filters miss. Structured logging of every
   input and output, alerting on anomalies.

## What we explicitly don't do — and why

**No prompt injection detection.** All detection approaches (regex heuristics,
embedding classifiers, llm-guard) are fundamentally unreliable against adaptive
attacks. Since this chatbot has no tool access, successful injection can only
produce off-brand text, not exfiltrate data. The output blocklist and
observability catch the damage more reliably.

**No LLM-as-judge output validation.** Using a second LLM to evaluate the first
has the same failure modes. The judge can be confused by the same techniques
that confuse the primary model. It adds latency, cost, and complexity without
reliable benefit.

**No semantic rate limiting.** Embedding-based similarity detection against
denied inputs requires PyTorch (~1.5 GB) and addresses a narrow attack vector
that turn limits already handle.

**No confidence scoring or hedging detection.** The presence of "I think" in a
response does not correlate with policy compliance.

**No multi-model judge, action allowlist, or stateful attack tracking.** These
are over-engineered for a text-only chatbot without tool access. See
[ADR 010](docs/adrs/010_security_threat_model.md) for the full reasoning.

## Deferred techniques — to be revisited

These techniques were cut based on the current threat model (no tool access,
text-only chatbot). They should be revisited if the system's capabilities
change.

| Technique | Revisit when | Reference |
|-----------|-------------|-----------|
| Prompt injection detection | Tool access or database queries are added — injection becomes a real exfiltration vector | ADR 007 (superseded by 010) |
| LLM-as-judge output validation | Response quality requirements exceed what deterministic filters can catch, AND a reliable judge approach emerges | ADR 010 |
| Semantic rate limiting | Repeated rephrased attacks become a measurable problem that turn limits don't address | ADR 009 (superseded by 010) |
| Confidence / hedging scoring | A correlation between hedging language and policy violations is demonstrated empirically | ADR 010 |
| Stateful attack tracking | Multi-turn escalation attacks are observed in production logs that turn limits don't prevent | ADR 010 |
| Action allowlist | The chatbot gains the ability to trigger actions (tool calls, API writes) | ADR 010 |
| Retry loop with LLM feedback | Deterministic filter false-positive rate is high enough to justify automated retries | ADR 010 |
| HITL queue | Observability logging is insufficient and human review of flagged conversations is needed in real-time | — |
| Intent classifier / tool router | The chatbot gains tool access or needs to route requests to specialised handlers | ADR 010 |
| Output schema validation | The chatbot produces structured (JSON) output instead of free text | ADR 010 |
| PII scrubbing | Compliance requirements demand it — approach TBD (Go-native library or Presidio sidecar) | Stage 10 |

## Deployment security recommendations

The effective security boundary is the infrastructure, not the application:

- **Run in an isolated container.** No host mounts, no network access to
  internal systems.
- **No tool access.** The LLM produces text and nothing else. No function
  calling, no code execution, no database access. This is the single most
  important security property.
- **Deterministic APIs for data.** If the chatbot needs external data (product
  info, branch locations), the application layer fetches it via deterministic
  API calls. The LLM never queries systems directly.
- **Principle of least privilege.** Non-root user, minimal filesystem
  permissions, scoped API keys.
- **Rate limiting at infrastructure level.** Reverse proxy / API gateway, not
  inside the application.
- **Audit logging.** All inputs and outputs logged with session ID and
  timestamp. Immutable, retained for compliance.
- **No sensitive data in context.** The system prompt contains no customer PII
  or internal data. If it leaks, nothing sensitive is exposed.

## Scenarios

Three domain scenarios test different compliance requirements:

- **Financial advisor (Morgan)** — retail bank assistant; must never give
  investment advice, must include regulatory disclaimers
- **Brand marketing (Sage)** — outdoor apparel brand assistant; must stay
  on-brand, never mention competitors, no unverifiable claims
- **Insurance claims (Dana)** — claims assistant; must never promise outcomes,
  must escalate complex cases to human adjusters

## Setup

Install [mise](https://mise.jdx.dev/) to manage Go and Node.js versions:

```bash
curl https://mise.run | sh
mise install   # reads mise.toml and installs Go + Node.js
```

After that, `go` and `node`/`npm` are available in your shell via mise shims.

## Tech stack

- **Go server:** Go, `openai-go` (official OpenAI SDK), `slog` (structured
  logging)
- **Frontend:** Next.js (App Router), React, Tailwind CSS
- **Scenarios:** YAML, embedded via `//go:embed`
- **Testing:** `go test` (server), React Testing Library (frontend)
- **Runtime versions:** managed by [mise](https://mise.jdx.dev/) (`mise.toml`)
- **PII detection:** TBD — deferred to stage 10

## Architecture decisions

See [docs/adrs/](docs/adrs/) for the full set. Key decisions:

- [ADR 010](docs/adrs/010_security_threat_model.md) — Security threat model:
  infrastructure over LLM guardrails
- [ADR 011](docs/adrs/011_rest_sse_api_transport.md) — REST + SSE for API
  transport
- [ADR 012](docs/adrs/012_in_memory_session_storage.md) — In-memory session
  storage with pluggable interface
- [ADR 013](docs/adrs/013_embedded_scenario_configs.md) — Embed scenario
  configs via go:embed
- [ADR 014](docs/adrs/014_markdown_prompt_format.md) — Markdown headers for
  system prompt structure
- [ADR 015](docs/adrs/015_go_nextjs_monorepo.md) — Go + Next.js monorepo
  (replaces Python monolith)
