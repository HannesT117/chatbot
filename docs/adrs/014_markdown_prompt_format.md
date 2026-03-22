# 014 — Markdown headers for system prompt structure

**Status:** Accepted
**Date:** March 22, 2026

## Context

The system prompt is built from scenario config and session state. It needs
internal structure to separate concerns (role, constraints, allowed topics,
blocklist, canary token). Three formatting options were considered:

**XML tags** (`<role>...</role>`). Claude responds well to XML-tagged prompts.
Provides clear delimiters that are unambiguous to parse.

**Markdown headers** (`## Role`). Well-handled by all major models (GPT-4o,
Claude, Mistral, Llama, etc.). Human-readable. The most universally understood
format.

**Plain sections with delimiters** (`---`). Minimal, universally understood,
but less semantically meaningful than headers.

## Decision

Markdown headers as the default prompt format, implemented as a Go
`text/template`.

### Why markdown

- The project uses `openai-go` with configurable base URL to support any
  OpenAI-compatible API. The prompt format should not be coupled to one model
  family. Markdown headers work well across all major models.
- Human-readable without any parsing. Developers can read the assembled system
  prompt and immediately understand its structure.
- No risk of confusion with actual XML in the response. XML tags in the prompt
  can cause some models to produce XML-formatted responses.

### Why not XML

- XML tags work well with Claude specifically, but less consistently across
  other providers. Since the project is provider-agnostic via litellm, a
  universally understood format is preferred.
- If testing shows that a specific model responds significantly better to XML,
  the prompt template can be swapped per-provider without changing the prompt
  assembly logic.

### Why Go templates

- The prompt structure is mostly static text with variable interpolation
  (persona name, constraints list, canary token). Go's `text/template` handles
  this cleanly.
- Templates are easy to test — render with known inputs and assert on the
  output string.
- If per-provider prompt formats become necessary, multiple templates can
  coexist, selected by model name.

## Consequences

- The system prompt is a rendered Go template, not a format string or string
  concatenation. This makes the prompt structure explicit and testable.
- Models that respond significantly better to XML or other formats may
  underperform slightly with markdown headers. This can be addressed by adding
  per-provider template variants if testing reveals a meaningful difference.
- The prompt template is the single source of truth for prompt structure.
  Changing the format means editing one template, not hunting through code.
