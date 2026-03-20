# Local Model Setup

MacBook Air, M4, 24GB unified memory

## Quick Reference

| Need | Reach for | Why |
|---|---|---|
| Implement a feature, refactor, write tests, quick code questions | Qwen2.5-Coder 14B (local) | Free, fast, plenty of memory headroom |
| Planning, architecture, design docs, multi-file review, repo-wide reasoning | Claude Sonnet 4.6 (cloud) | Near-Opus coding quality at ~1/5 the cost |
| Novel architecture, critical migrations, second opinion | Claude Opus 4.6 (cloud) | Peak reasoning, use sparingly (~5x Sonnet cost) |

See also the [Decision Chart](./llm-pick-flow.md)

## Models

### Claude Opus 4.6 (cloud, very expensive)

Good at: Long-horizon reasoning, complex/novel architecture, multi-step agentic workflows, very long outputs and contexts.

Use case: Only for rare, high-risk "boss fights" (big migrations, tricky distributed designs, critical incident postmortems). Escalate to Opus only if Sonnet gives 1–2 clearly inadequate or contradictory passes on the same problem and the decision is high-risk. At roughly 5x Sonnet's per-token cost, any Opus call should pay for itself in reduced risk or engineering time.

### Claude Sonnet 4.6 (cloud, cheaper default)

Good at: Frontier-level coding across the whole SDLC (planning, implementation, debugging, large refactors), long-context repo reasoning, solid agentic workflows, daily professional work. Benchmarks within a couple of points of Opus on SWE-Bench Verified and coding tasks — the sweet spot for quality vs. cost.

Use case: Primary cloud brain. Handles planning, architecture, design docs, multi-file code reviews/refactors, and anything where the local model hits reasoning or context limits. With only one local model (code-focused), Sonnet picks up more of the planning and design work. Use Sonnet's long-context mode when a single prompt needs hundreds of thousands of tokens (big repo map, long spec + code) before even considering Opus.

### Qwen2.5-Coder-14B-Instruct (local, via Ollama)

Good at: Strong code completion, refactoring, test writing, and multi-language support. The 14B dense model is well-suited for IDE-integrated coding tasks and quick code questions. Excellent for local implementation — not your primary architect. For non-trivial design decisions, escalate to Sonnet even if Qwen "seems" to have an answer.

Use case: Primary local coding engine for implementing features, refactors, tests, and cross-file edits. Also handles light planning and quick questions where you want zero latency and no API cost. Tip: ask it to "think step-by-step and propose a plan, then show the patch" to offset weaker global reasoning vs. Sonnet/Opus.

Memory note: Roughly 7–9GB at Q4 (varies by specific quant), leaving plenty of headroom on 24GB for KV cache and system overhead. Model is trained for up to 128k+ context, but the Ollama runtime defaults to ~32k; you can push higher by tuning `num_ctx` at the cost of extra RAM and latency.

## Escalation Rules

When to move from Qwen (local) to Sonnet (cloud): the change is cross-cutting (multi-module, framework-level), you need a design doc / ADR / careful trade-off analysis, or Qwen produces two low-quality or obviously wrong patches in a row.

When to move from Sonnet to Opus: Sonnet gives 1–2 clearly inadequate or contradictory passes on the same problem, and the decision is high-risk (migrations, correctness-critical distributed flows, post-mortems, or novel research-y designs).

Suggested sequence for large refactors: Sonnet for the high-level design / migration plan / invariants, then Qwen locally for the mechanical implementation (functions, wiring, tests), then Sonnet again for a final multi-file "what did we miss?" review. Only consider Opus if Sonnet signals genuine uncertainty after that.
