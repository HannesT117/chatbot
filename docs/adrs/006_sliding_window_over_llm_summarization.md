# 006 — Sliding context window over LLM summarization

**Status:** Accepted
**Date:** March 20, 2026

## Context

The conversation manager must handle conversations that approach the LLM's
context window limit. Two strategies exist:

**LLM summarization:** When the token budget approaches the limit, call the
LLM to summarize older turns, replace them with the summary, and continue.
This preserves conversational continuity.

**Sliding window (drop oldest):** Drop the oldest turns when the budget is
approached. Simpler and requires no additional LLM calls. Continuity is
partially lost.

Summarization is appealing for user experience but introduces a security
concern specific to this project's threat model.

## Decision

Use a sliding window that drops the oldest turns when approaching the token
limit. Do not implement LLM summarization.

The core reason is security: summarization requires an unguarded LLM call.
The summary prompt is constructed from conversation history that may contain
attacker-controlled content. An attacker who has seeded the conversation
with injection payloads could cause the summarization call to produce a
manipulated summary — for example, one that overrides the persona, removes
denied topics from the "memory," or fabricates prior context. This summary
then re-enters the conversation as trusted context.

Running summarization through the full output pipeline would address some of
these concerns, but it adds complexity and latency for a feature that is not
core to the guardrail evaluation goals of this project.

For a production chatbot where continuity matters, summarization via a
guarded pipeline would be appropriate. For a guardrails testbed, the security
trade-off is not worth it.

## Consequences

- The conversation manager implementation is simpler: no secondary LLM call,
  no summarization prompt, no additional output validation path.
- Long conversations lose early context. Users in long sessions may need to
  re-state context that was dropped.
- The sliding window approach is auditable: the exact turns in context at any
  point are known and logged, with no LLM-generated content silently mixed in.
- If continuity becomes important in a future production deployment, this
  decision can be revisited with the security implications explicitly
  accounted for.
