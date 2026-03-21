# 008 — Test strategy: mocked unit tests and gated live integration tests

**Status:** Accepted
**Date:** March 20, 2026

## Context

The pipeline makes LLM API calls, which are slow, cost money, and return
non-deterministic results. Tests that hit live APIs are problematic in CI:
they add latency, incur cost on every run, and can fail due to network issues
or rate limits unrelated to the code under test.

Three strategies were considered:

**Mocks only:** All LLM calls are mocked. Tests are fast and deterministic.
Risk: mock responses may diverge from real model behavior over time, giving
false confidence.

**Live only:** All tests hit real APIs. Tests are realistic but slow, expensive,
and non-deterministic. Not viable for CI.

**Mocked unit tests + gated live integration tests:** Unit tests use mocks and
run in CI. A separate integration test suite hits live APIs but only runs when
explicitly enabled. The two suites serve different purposes.

## Decision

Use a two-tier test strategy:

**Unit tests** (`tests/unit/`): Mock all LLM calls using recorded fixtures or
synthetic responses. These tests are fast, deterministic, and always run in CI.
Every pipeline step gets unit tests with known-good and known-bad inputs.
The unit suite must never hit a live API.

**Integration tests** (`tests/integration/`): Hit real LLM APIs. Gated behind
the `CHATBOT_LIVE_TESTS=1` environment variable — not run in CI by default.
Use a cheap, fast model (for example, `gpt-4o-mini`) to keep costs low. These
tests validate that mocked unit tests haven't diverged from real model behavior.

**Adversarial tests** (`tests/adversarial/`): The core research output of the
project. Attack payloads organized by category (injection, jailbreak,
multi-turn escalation, PII extraction, canary leak). Run against both mocked
and live backends. These tests define what the guardrail system must withstand.

## Consequences

- CI is fast and cost-free. Developers can run the full unit suite locally
  without API keys.
- The `CHATBOT_LIVE_TESTS=1` gate is a manual step, which means live tests
  can lag behind if developers forget to run them. This is an acceptable
  trade-off given the cost and speed concerns.
- Mock/real divergence is a real risk. The integration suite must be run
  periodically (for example, before any significant merge) to catch it.
- Adversarial tests are the most valuable artifact of this project. They
  must be kept up to date as new attack patterns emerge and as the pipeline
  evolves.
