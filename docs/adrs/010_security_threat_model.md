# 010 — Security threat model: infrastructure over LLM guardrails

**Status:** Accepted
**Date:** March 22, 2026
**Supersedes:** ADR 007, ADR 009

## Context

The original architecture invested heavily in LLM-based guardrails: prompt
injection detection (custom embedding classifier + llm-guard), LLM-as-judge
output validation, confidence scoring, and semantic rate limiting via
sentence-transformers. These components required significant ML dependencies
(PyTorch ~1.5 GB, sentence-transformers, llm-guard) and engineering effort.

A review of the AI security literature — particularly Schulhoff's analysis of
guardrail effectiveness ("The AI Security Industry is Bullshit", 2026) and the
adaptive attack research it cites — exposed fundamental problems with this
approach.

### Why LLM-based guardrails don't work

**The input space is astronomical.** There are effectively infinite possible
prompts. Testing even a fraction is computationally infeasible. Guardrail
evaluations use static prompt sets that don't reflect real-world adaptive
attacks, producing inflated success metrics that collapse under adversarial
pressure.

**All LLM-based classifiers are breakable.** Research shows that guardrails
self-reporting 99% effectiveness are "all extremely breakable" under adaptive
attack. An LLM judging whether another LLM's output is safe has the same
fundamental failure modes as the primary model — it's turtles all the way down.

**Guardrails create false confidence.** When a guardrail "passes" a response,
operators assume the output is safe. This is worse than having no guardrail at
all, because it discourages the manual review and observability that would
actually catch problems.

### Why this chatbot's threat model is different

**This chatbot has no tool access.** It cannot execute code, query databases,
send emails, or take any action beyond producing text. Even a fully jailbroken
instance can only generate inappropriate text — it cannot exfiltrate data or
take destructive actions. This places it in the lowest-risk category for AI
security.

**The real risks are brand damage and regulatory non-compliance, not data
exfiltration.** A financial advisor chatbot giving investment advice, or a brand
chatbot disparaging competitors, causes real business harm. But this is a brand
compliance problem, not a cybersecurity problem. It is better addressed by
system prompt design and deterministic output filters than by ML classifiers.

## Decision

Drop all LLM-based and ML-based guardrails. Invest in infrastructure-level
security, deterministic filters, and observability.

### What we keep and why

**System prompt with full scenario policy.** Explicit persona, constraints,
allowed topics, and blocklist terms injected via XML-tagged prompt. This is
brand compliance: telling the LLM what to be and what to talk about. It is not
a security boundary — it will be bypassed by sufficiently creative prompting —
but it handles the vast majority of normal user interactions correctly.

**Deterministic input/output blocklist.** Regex/keyword filters for known
forbidden terms. Cheap, fast, cannot be argued around. Catches obvious
violations before and after the LLM call. Not a complete defense, but a
reliable first and last layer.

**Canary token detection.** Per-session random hex token in the system prompt,
exact string match on output. Deterministic, zero false positives. Detects
verbatim system prompt leakage only — paraphrased or partial leakage is not
caught. This is an observability signal, not a security boundary.

**PII scrubbing.** Presidio-based detection and redaction. Framed as regulatory
compliance (data minimisation), not as a security gate. Prevents accidental PII
exposure in LLM responses.

**Observability.** Structured logging of all inputs/outputs, alerting on
anomalies, HITL queue for flagged conversations, random sampling for quality
audit. This is explicitly supported by the security literature as the right
investment for AI systems.

**Turn limits and token budget.** Prevents resource abuse and limits
conversation length. Classical rate limiting.

### What we cut and why

**Prompt injection detection** (was ADR 007). All detection approaches — regex
heuristics, embedding classifiers, llm-guard — are fundamentally unreliable
against adaptive attacks. Since this chatbot has no tool access, successful
prompt injection has no blast radius beyond brand damage, which is better
caught by output blocklist and observed via logging. Building and maintaining a
prompt injection classifier creates false confidence without meaningful security
improvement.

**LLM-as-judge output validation.** Using a second LLM to evaluate the first
LLM's output has the same failure modes as the primary model. The judge can be
confused by the same techniques that confuse the primary model. It adds latency,
cost, and complexity without reliable security benefit.

**Confidence/hedging scoring.** Heuristic regex patterns for "I think" or
"probably" are not actionable as security signals. The presence or absence of
hedging language does not correlate with policy compliance.

**Semantic rate limiting** (was ADR 009). Embedding-based cosine similarity
against denied inputs requires PyTorch (~1.5 GB), adds deployment complexity,
and addresses a narrow attack vector (rephrased re-asks) that turn limits
already handle adequately.

**Multi-model judge.** Same fundamental problem as single-model judge, with
doubled latency and cost.

**Action allowlist.** This chatbot has no actions to allowlist. No tool calls,
no code execution, no database writes.

**Stateful attack tracking / circuit breakers.** Over-engineered for a
text-only chatbot. Turn limits and infrastructure-level rate limiting provide
sufficient conversation-level control.

### Deployment security recommendations

The effective security boundary is the infrastructure, not the application
layer:

1. **Run in an isolated container.** Docker with no host mounts, no network
   access to internal systems. The chatbot process should have access to nothing
   it does not need.

2. **No tool access.** The LLM has no function calling, no code execution, no
   database access. It produces text and nothing else. This is the single most
   important security property of the system.

3. **Deterministic APIs for data.** If the chatbot needs external data (product
   information, branch locations), fetch it via deterministic API calls made by
   the application layer, not by the LLM. The LLM receives pre-fetched data in
   its context; it never queries systems directly.

4. **Principle of least privilege.** The application process runs as a non-root
   user with minimal filesystem permissions. Environment variables containing
   API keys are scoped to the specific LLM provider and nothing else.

5. **Rate limiting at infrastructure level.** Rate limit requests at the
   reverse proxy or API gateway layer, not inside the application. This is
   cheaper, more reliable, and handles volumetric abuse.

6. **Audit logging.** All inputs and outputs are logged with session ID,
   timestamp, and scenario metadata. Logs are immutable and retained for
   compliance review.

7. **No sensitive data in context.** The system prompt and conversation history
   contain no customer PII, financial data, or internal business data. If the
   entire system prompt leaks, no sensitive information is exposed.

## Consequences

- The dependency footprint drops significantly: no PyTorch, no
  sentence-transformers, no llm-guard.
- The pipeline is simpler: deterministic input checks → system prompt → LLM →
  deterministic output checks → observability.
- False confidence from "guardrails passed" is eliminated. The system is honest
  about what it can and cannot prevent.
- Novel attacks that bypass the system prompt and deterministic filters will
  produce inappropriate responses. These are caught by observability and human
  review, not by automated guardrails that claim to be reliable.
- Adding tool access or database queries in the future would fundamentally
  change the threat model and require re-evaluating this decision.

## References

- Schulhoff, "The AI Security Industry is Bullshit" (2026): analysis of
  guardrail limitations and adaptive attack research
- ADR 007 (superseded): original prompt injection detector design
- ADR 009 (superseded): original sentence-transformers dependency decision
