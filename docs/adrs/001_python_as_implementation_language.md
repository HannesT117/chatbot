# 001 — Python as implementation language

**Status:** Superseded by ADR 015
**Date:** March 20, 2026
**Superseded:** March 22, 2026

## Context

The project needs an implementation language for a guardrails pipeline that
includes ML-based components: PII detection, prompt injection classification,
and semantic similarity scoring. Candidates considered were Python, Go, and
Rust.

The pipeline will run as a server at some point, so server-deployment
characteristics (performance, binary size, deployment complexity) matter.

Go and Rust are attractive for server deployments: single-binary output,
low memory footprint, predictable latency, and no runtime to manage. However,
the ML/NLP tooling ecosystem is overwhelmingly Python-first.

The specific dependencies required by this project are Python-only and have no
maintained Go or Rust equivalents:

- `presidio-analyzer` / `presidio-anonymizer` — Microsoft's NLP-based PII
  detection, built on spaCy and transformers
- `sentence-transformers` — PyTorch-based semantic similarity
- `llm-guard` — prompt injection and output scanning
- `litellm` — unified LLM provider abstraction

Replacing these with cloud NLP APIs (for example, AWS Comprehend for PII)
would be possible but would change the scope of the project from a local,
auditable guardrail testbed to one dependent on third-party services.

The dominant bottleneck in this system is LLM API latency, which is measured
in seconds. Python interpreter overhead is not a meaningful factor.

## Decision

Use Python 3.12+ as the sole implementation language. Use `uv` for package
management and `ruff` / `mypy --strict` for linting and type checking.

If a server deployment eventually requires a lightweight HTTP layer for
concerns like auth, rate limiting, or TLS termination, a Go or Rust sidecar
can handle those concerns while the Python service owns all guardrail logic.
That is an ops decision and does not require rewriting the pipeline.

## Consequences

- The ML/NLP dependency ecosystem is available without any workarounds.
- Python async (`asyncio`) is sufficient for I/O-bound concurrency given that
  LLM calls dominate latency.
- Deployment requires managing a Python runtime and dependencies; Docker with
  `uv` makes this straightforward.
- A future production version that replaces ML components with hosted APIs
  could be rewritten in Go or Rust at that point, with the Python testbed
  serving as the reference implementation.

## Superseded — March 22, 2026

This ADR is superseded by ADR 015 (Go + Next.js monorepo replaces Python
monolith).

ADR 010 removed all ML-based guardrails (prompt injection detection, LLM judge,
semantic rate limiting), eliminating the ML/NLP ecosystem as the primary reason
for Python. With the remaining requirements being an HTTP server and a web UI,
Go and Next.js are a better fit: single-binary deployment, small containers,
cleaner security boundary, and a real web frontend instead of a terminal TUI.
