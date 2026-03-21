# 005 — Fail-closed pipeline error handling

**Status:** Accepted
**Date:** March 20, 2026

## Context

Pipeline steps can fail in two distinct ways:

1. **Rejection:** The step runs successfully and determines the input or output
   does not meet policy. This is expected behavior — the step returns a
   structured rejection result.

2. **Exception:** The step throws an unexpected error (for example, Presidio
   crashes on malformed Unicode, the embedding model runs out of memory, or the
   LLM call times out).

Two strategies exist for handling exceptions mid-pipeline:

**Skip and continue:** Log the exception, treat the step as if it passed, and
continue to subsequent steps. This maximizes availability — users always get a
response even if a guard crashes.

**Fail closed:** Log the exception, reject the input with a generic safe error
message, and halt the pipeline. Users see an error response instead of a
potentially unguarded one.

## Decision

When any pipeline step throws an exception, the pipeline fails closed. The
request is rejected with a generic safe error message. The exception is logged
with full context (step name, input, traceback). No subsequent steps run.

This applies to both the input pipeline and the output pipeline.

Implementation: each pipeline step implements a common `PipelineStep`
protocol with a `process()` method. The pipeline runner wraps each call in
`try/except Exception`, logs on failure, and returns a `PipelineResult.error`
that halts the pipeline.

## Consequences

- A crashing PII scrubber, injection classifier, or output validator never
  results in an unguarded response reaching the user.
- Availability takes second place to safety. If a non-critical step (for
  example, confidence scoring) crashes, the request is still rejected. This
  is intentional — in a regulated environment, "the guardrail crashed so we
  let it through" is not an acceptable outcome.
- Exception rates are observable via structured logs and alerts, so
  infrastructure failures surface quickly rather than silently degrading
  guardrail coverage.
- Steps that are truly optional (for example, the multi-model judge) must be
  wrapped at the call site to catch their own exceptions and return a
  pass-through result rather than propagating to the pipeline runner.
