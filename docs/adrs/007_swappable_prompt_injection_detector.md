# 007 — Swappable PromptInjectionDetector protocol

**Status:** Accepted
**Date:** March 20, 2026

## Context

The input pipeline needs a prompt injection detection step. The original plan
referenced `rebuff` as the implementation, but `rebuff` was archived and
unmaintained as of mid-2023. Two maintained alternatives exist:

- **`llm-guard`** (laiyer-ai): A maintained open-source library with a prompt
  injection scanner. Faster to integrate but treats the detection as a black
  box.
- **Custom embedding-based classifier**: Build a classifier using labeled data
  and sentence embeddings. More educational for a testbed project — you
  understand exactly what it's doing and can evaluate its failure modes.

A third consideration: the testbed's purpose is to evaluate which guardrail
combinations are reliable. Using a single, hard-coded detection approach means
you can't compare approaches. Being able to swap implementations behind a
stable interface is directly useful for the project's goals.

## Decision

Define a `PromptInjectionDetector` protocol (a Python `typing.Protocol`) with
a single async method: `detect(text: str) -> DetectionResult`. The protocol is
the only interface the pipeline interacts with.

Step 7 of the implementation plan implements both:

1. A custom embedding-based classifier (approach A)
2. `llm-guard`'s injection scanner (approach B)

The active implementation is selected via the `CHATBOT_INJECTION_DETECTOR`
environment variable. Both implementations must pass the same adversarial test
suite — the shared protocol makes this straightforward.

During initial steps (before step 7), a regex stub implementation satisfies
the protocol with minimal overhead.

## Consequences

- The pipeline never depends on a specific implementation. Switching detectors
  requires only an environment variable change.
- The testbed can directly compare detection rates, false positive rates, and
  latency between the two implementations against the same attack corpus.
- Implementing both adds development effort at step 7, but this effort is
  core to the project's research purpose.
- Adding a third detector (for example, a fine-tuned classifier) in the future
  requires only implementing the protocol — no pipeline changes.
- The regex stub means early pipeline steps are testable end-to-end before the
  ML-based detector is ready.
