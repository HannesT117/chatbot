# LLM guardrails testbed

This project is a testing ground for best practices around securing LLM
application output. It explores techniques for ensuring that AI-generated
responses are correct, compliant, and appropriate — covering regulated
industries like financial services, as well as brand-sensitive contexts
like advertising.

## Approach

The core idea is **defense in depth with deterministic guardrails first**.
The LLM is treated as an untrusted component — its output is validated by
the same rigor you'd apply to user input in a web application.

The request/response flow passes through three pipeline stages:

1. **Input pipeline**: Deterministic checks run first (schema validation,
   regex blocklists, PII scrubbing), followed by ML-based signals (prompt
   injection detection, semantic rate limiting). Intent and tool routing
   decide what the LLM is allowed to do for this request.
2. **Guarded LLM call**: A carefully designed system prompt locks the model
   into a role, forces structured output, and includes canary tokens to
   detect system prompt leakage. The conversation manager enforces turn
   limits and token budgets, summarizes old context instead of truncating
   it, and tracks multi-round attack patterns across turns.
3. **Output pipeline**: Deterministic validation runs first again — schema
   checks, action allowlist enforcement, canary leak detection. Then
   optional ML-based scoring (confidence, hallucination, multi-model
   judging). Failed outputs retry with feedback, up to a hard limit.

Everything is logged with structured labels and scores. High-risk patterns
trigger alerts. Flagged conversations are queued for human review, and 1%
of all conversations are randomly sampled for quality.

See [plans/001-guardrailed-chatbot.md](plans/001-guardrailed-chatbot.md)
for the full architecture.

## Why this approach

**Deterministic checks can't be talked around.** A regex blocklist doesn't
care how creatively the user phrases a request. A JSON schema validator
doesn't negotiate. By placing these checks before and after the LLM, the
system maintains hard boundaries even when the model is fully compromised
or cooperating with an attacker.

**Treating LLM output as untrusted input is the right threat model.** In
regulated environments, "the model said it was fine" is not a defensible
position. Structural validation (schema checks, action allowlists) provides
auditable proof that output conforms to policy, regardless of what the
model intended.

**Multi-round awareness catches what single-turn filters miss.** Many
real-world attacks work by gradually steering the conversation — first
establishing trust, then escalating. Stateful tracking of denied requests,
topic drift, and escalation patterns across turns addresses this.

**Layering is key because no single technique is sufficient.** The system
prompt will be bypassed. The input filter will miss novel attacks. The
output filter may pass edge cases. But an attacker who has to defeat all
three simultaneously, across a conversation that's being monitored for
patterns, faces a much harder problem.

## Why common alternatives fall short

**"Just use a good system prompt"** — System prompts are a single layer
of defense that lives inside the model context. They can be overridden by
sufficiently creative prompt injection, context stuffing, or multi-turn
manipulation. A system prompt is necessary but never sufficient.

**Probabilistic content filters as primary defense** — Asking an LLM to
judge whether another LLM's output is safe creates a recursive trust
problem. The judge model has the same failure modes as the primary model.
These filters are useful as supplementary signals but unreliable as sole
gates. Deterministic checks must come first.

**Blocklist-only approaches** — Keyword blocklists catch known patterns but
don't generalize. They miss rephrased attacks, encoded payloads, and novel
techniques. They're a fast first layer, not a complete solution.

**Single-turn analysis only** — Evaluating each message in isolation misses
the most effective attack class: multi-turn escalation. An attacker who is
denied on turn 1 rephrases on turn 2, builds rapport on turn 3, and
extracts on turn 4. Without cross-turn state tracking, each turn looks
benign.

**Over-relying on third-party guardrail libraries** — Many guardrail
frameworks wrap probabilistic LLM calls in a convenience API. This can
create false confidence — the guardrail "passed" so the output must be
safe. Understanding and controlling each layer explicitly produces a more
auditable and debuggable system.

## Goals

The goal is not to build a production chatbot, but to identify which
guardrail combinations are reliable enough for high-stakes environments —
and to document the trade-offs between safety, latency, and user
experience.
