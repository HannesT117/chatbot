# 004 — LiteLLM for LLM abstraction

**Status:** Superseded by ADR 015
**Date:** March 20, 2026
**Superseded:** March 22, 2026

## Context

The project needs to call LLM APIs. Options include:

- **Direct provider SDKs** (`anthropic`, `openai`): Maximum control, but
  provider-specific. Switching providers requires code changes.
- **LiteLLM**: A unified abstraction over 100+ providers using the
  OpenAI-compatible API surface. Provider is selected via environment variables
  (`LITELLM_MODEL`, `ANTHROPIC_API_KEY`, and so on), not in code.
- **Custom abstraction**: Writing a thin wrapper ourselves. More control, but
  more maintenance.

The project is a testbed, not a production system for a fixed provider. Being
able to swap providers (Anthropic, OpenAI, Mistral, Bedrock) without code
changes is valuable for evaluating how guardrails behave across models.

The multi-model judge feature (step 9) requires calling a different model than
the primary LLM, potentially from a different provider. LiteLLM handles this
with the same interface.

## Decision

Use `litellm` as the sole interface for all LLM calls, including the primary
conversation call and the optional multi-model judge. The provider and model
are configured via environment variables. No provider SDK is imported directly
in application code.

## Consequences

- Switching providers requires only environment variable changes, no code edits.
- The multi-model judge can use a different model or provider than the primary
  call with no additional abstraction.
- LiteLLM adds a dependency and a thin indirection layer. If LiteLLM has a
  bug or falls behind a provider's API, it blocks the project until LiteLLM
  updates.
- Async calls use `litellm.acompletion`, consistent with the async-first
  architecture (ADR 003).

## Superseded — March 22, 2026

This ADR is superseded by ADR 015 (Go + Next.js monorepo replaces Python
monolith).

The Go server uses `openai-go` (the official OpenAI Go SDK) instead of litellm.
Provider switching is handled by configuring the base URL — `openai-go` works
with any OpenAI-compatible API (OpenAI, Anthropic, Azure, Ollama, etc.). The
multi-model judge has been cut (ADR 010), removing the need for multi-provider
calls in a single process.
