# System Prompt Module — Stage 3 Design

## Overview

Stage 3 replaces the 2-line inline system prompt in `ConversationManager` with a
structured XML prompt built by a dedicated `src/chatbot/prompts/system.py` module.
The prompt embeds the full scenario policy (persona, output constraints, allowed
topics, blocklist, canary token) so the LLM has explicit behavioural instructions
from the start of every conversation.

## Scope

Files added or modified:

| File | Status |
|------|--------|
| `src/chatbot/prompts/system.py` | Implement — `build_system_prompt` + scenario registry |
| `src/chatbot/prompts/scenarios/financial_advisor.py` | Implement — `build_persona_notes` |
| `src/chatbot/prompts/scenarios/brand_marketing.py` | Implement — `build_persona_notes` |
| `src/chatbot/prompts/scenarios/insurance_claims.py` | Implement — `build_persona_notes` |
| `src/chatbot/conversation/manager.py` | Modify — replace inline prompt with `build_system_prompt` call |
| `tests/unit/prompts/__init__.py` | New |
| `tests/unit/prompts/test_system.py` | New — unit tests |

## Public Interface

`manager.py` imports one function and calls it:

```python
from chatbot.prompts.system import build_system_prompt

system_prompt = build_system_prompt(self._scenario, session)
```

```python
# src/chatbot/prompts/system.py
def build_system_prompt(scenario: ScenarioConfig, session: Session) -> str:
    """Build the XML system prompt for the given scenario and session.

    Raises ValueError for unknown scenario names.
    """
```

Each scenario file exports one function:

```python
# src/chatbot/prompts/scenarios/<name>.py
def build_persona_notes(config: ScenarioConfig) -> str:
    """Return persona-specific instructions to inject into <role>."""
```

## Prompt Structure

```xml
<role>
{persona_name}. {persona_description}

{build_persona_notes(config)}
</role>
<constraints>
- {output_constraint_1}
- {output_constraint_2}
</constraints>
<allowed_topics>
- {allowed_intent_1}
- {allowed_intent_2}
</allowed_topics>
<blocklist>
- {blocklist_term_1}
- {blocklist_term_2}
</blocklist>
<canary>{session.canary_token}</canary>
```

Each section maps directly from `ScenarioConfig` fields:

| Section | Source |
|---------|--------|
| `<role>` | `persona_name`, `persona_description`, `build_persona_notes(config)` |
| `<constraints>` | `output_constraints` list |
| `<allowed_topics>` | `allowed_intents` list |
| `<blocklist>` | `blocklist_terms` list |
| `<canary>` | `session.canary_token` (bare hex, no identifying label) |

## `system.py` Implementation

```python
from chatbot.prompts.scenarios import financial_advisor, brand_marketing, insurance_claims

_PERSONA_BUILDERS: dict[str, Callable[[ScenarioConfig], str]] = {
    "financial_advisor": financial_advisor.build_persona_notes,
    "brand_marketing": brand_marketing.build_persona_notes,
    "insurance_claims": insurance_claims.build_persona_notes,
}

def build_system_prompt(scenario: ScenarioConfig, session: Session) -> str:
    if scenario.name not in _PERSONA_BUILDERS:
        raise ValueError(f"No persona builder registered for scenario '{scenario.name}'.")
    persona_notes = _PERSONA_BUILDERS[scenario.name](scenario)
    # assemble XML prompt
    ...
```

## Scenario-Specific `build_persona_notes` Content

### `financial_advisor` (Morgan)

- Morgan always recommends consulting a CFP or qualified financial professional for personalised advice.
- Every response touching regulated topics must include the standard disclaimer: "This is general information only and does not constitute financial advice."
- When declining a request (e.g. investment advice, trade execution), Morgan is warm but unequivocal — no hedging.

### `brand_marketing` (Sage)

- Sage uses a positive, aspirational tone consistent with the brand voice guidelines.
- Never disparages competitors by name or implication.
- Escalates customer complaints or negative sentiment to a human representative rather than attempting resolution autonomously.

### `insurance_claims` (Dana)

- Dana uses empathetic, plain-English language — no insurance jargon without explanation.
- Never makes coverage determinations or promises about claim outcomes.
- Complex cases (disputed claims, legal questions, large settlements) are referred to a human adjuster.

## Testing

All tests in `tests/unit/prompts/test_system.py`. `Session.new()` is used directly — no mocking needed.

| Test | What it verifies |
|------|-----------------|
| `test_contains_persona_name` | `persona_name` appears in the prompt |
| `test_contains_persona_description` | `persona_description` appears in the prompt |
| `test_contains_canary_token` | `session.canary_token` appears in the prompt |
| `test_canary_in_canary_tag` | canary is wrapped in `<canary>…</canary>` |
| `test_constraints_present` | each `output_constraints` item appears in the prompt |
| `test_allowed_topics_present` | each `allowed_intents` item appears in the prompt |
| `test_blocklist_present` | each `blocklist_terms` item appears in the prompt |
| `test_persona_notes_present` | `build_persona_notes` output appears in the prompt |
| `test_unknown_scenario_raises` | `ValueError` raised for an unregistered scenario name |
| `test_two_sessions_different_canary` | same scenario, two sessions → different `<canary>` content |

Existing `test_manager.py` tests (`test_system_prompt_contains_canary`,
`test_canary_only_in_system_message`) continue to pass unchanged — they assert
`session.canary_token` is present in the system message, which remains true.

## Security Properties Delivered

- **Explicit policy in context** — the LLM receives its constraints and allowed topics
  verbatim at the start of every turn, reducing reliance on implicit training.
- **Canary token** — bare hex string in `<canary>`, no identifying label. The output
  pipeline (stage 5) matches it exactly; its presence in output signals system prompt leakage.
- **Separation of concerns** — prompt construction is fully isolated from conversation
  orchestration. `ConversationManager` has no knowledge of prompt structure.
