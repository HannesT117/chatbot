# 002 — Deterministic-first pipeline ordering

**Status:** Accepted (narrowed by ADR 010)
**Date:** March 20, 2026

## Context

The guardrails pipeline runs checks on both input (before the LLM call) and
output (after it). Each check can be broadly categorized as either
deterministic (regex, schema validation, string match) or probabilistic
(ML classifier, LLM judge).

Two orderings are possible:

**Probabilistic-first:** Run ML-based checks early to catch subtle attacks,
then run deterministic checks. This surfaces sophisticated threats before
cheap pattern matching.

**Deterministic-first:** Run cheap, exact checks first. Only invoke ML-based
checks if deterministic checks pass.

There's also a question of whether probabilistic checks act as gates (block on
low confidence) or signals (annotate and continue).

## Decision

Deterministic checks run before probabilistic checks in every pipeline stage.
Probabilistic checks act as additional signals, not sole gates.

Rationale:

- Deterministic checks cannot be argued around. A regex blocklist does not
  negotiate; a JSON schema validator does not interpret intent. This property
  is unconditional, which makes it valuable as a first line of defense.
- Probabilistic checks have false negative rates by definition. Relying on
  them as the primary gate means some attacks pass. Using them as supplementary
  signals after deterministic gates have already run limits the blast radius of
  false negatives.
- Deterministic checks are fast and cheap. Running them first means most
  benign requests clear the pipeline quickly, and ML inference runs only on
  requests that passed initial screening.
- "The LLM said it was fine" is not an auditable position in regulated
  environments. Deterministic checks produce verifiable evidence of policy
  compliance regardless of model behavior.

## Consequences

- Every pipeline step is classified at design time as deterministic or
  probabilistic. Deterministic steps are implemented first within each stage.
- Probabilistic checks annotate the request with scores and labels that are
  logged for observability, but they do not unilaterally pass or fail a
  request on their own.
- Novel attack patterns that bypass all deterministic checks may still reach
  the LLM. The output pipeline provides a second layer of defense.
- Adding a new check requires an explicit decision about where it sits in
  the ordering, which keeps the pipeline auditable.

## Update — March 22, 2026 (ADR 010)

ADR 010 removed all probabilistic/ML-based checks from the pipeline. The
ordering principle in this ADR remains valid — deterministic checks still run
in a defined order — but the "probabilistic checks as supplementary signals"
clause is now moot. The pipeline is deterministic-only.

This reinforces the original rationale: deterministic checks cannot be argued
around, produce auditable evidence, and are fast and cheap. The decision to
drop probabilistic checks entirely (rather than keep them as signals) was
driven by the finding that they create false confidence without meaningful
security benefit in a chatbot without tool access. See ADR 010 for the full
security threat model.
