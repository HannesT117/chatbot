# Conversation Manager — Stage 2 Design

## Overview

Stage 2 wires a real LLM (via `litellm`) into the echo REPL from stage 1,
adding per-session state management, turn limits, a sliding token-budget
window, and canary token injection. The result is a working multi-round
chatbot with no safety layers yet — those are added in stages 3–13.

## Scope

Files added or modified:

| File | Status |
|------|--------|
| `scenarios/*.yaml` | Modified — add `max_turns`, `token_budget` |
| `src/chatbot/config.py` | Modified — add fields to `ScenarioConfig`, add `load_model()` |
| `src/chatbot/conversation/session.py` | New |
| `src/chatbot/conversation/manager.py` | Implemented (was stub) |
| `src/chatbot/conversation/tracking.py` | New stub (stage 11) |
| `src/chatbot/__main__.py` | Modified — async input, error handling, wire session + manager |
| `tests/unit/conversation/__init__.py` | New |
| `tests/unit/conversation/test_session.py` | New |
| `tests/unit/conversation/test_manager.py` | New |

## Config & scenario YAML

`ScenarioConfig` gains two new required fields:

```python
max_turns: int       # hard turn limit; conversation ends after this many turns
token_budget: int    # max tokens in the sliding context window
```

`load_scenario()` is wrapped in `try/except ValidationError` — a missing or
invalid field now produces a readable error message rather than a Pydantic
traceback.

Example defaults per scenario:

| Scenario | `max_turns` | `token_budget` |
|----------|-------------|----------------|
| financial_advisor | 20 | 4000 |
| brand_marketing | 30 | 6000 |
| insurance_claims | 20 | 4000 |

Model name is not part of the scenario config. It is read from the
`CHATBOT_MODEL` environment variable (default: `gpt-4o-mini`) via a
`load_model() -> str` helper in `config.py`.

## `session.py`

### Data types

```python
@dataclass(frozen=True)
class Turn:
    user: str
    assistant: str

class Session:
    session_id: str       # uuid4 hex string, read-only
    canary_token: str     # secrets.token_hex(16) — 32 hex chars, read-only
    turn_count: int       # total turns taken in this session (read-only property)
    # _turns: list[Turn]  # private; never exposed directly
```

`turns` is a private `list[Turn]`. It is never returned to callers directly,
preventing unintended mutation. All reads and writes go through the public
methods below.

`turn_count` reflects the number of turns taken in this session, *including*
turns that have since been dropped from the sliding window by
`trim_to_budget()`. It never decrements. The `_turns` list is a subset
(the sliding window); `turn_count` is the historical total.

### Factory

```python
@classmethod
def new(cls) -> Session
```

Generates a fresh `session_id` (UUID4 hex) and `canary_token`
(`secrets.token_hex(16)`). Each call produces independent values;
no shared state exists between sessions.

### Methods

**`session.add_turn(user: str, assistant: str) -> None`**
Appends a `Turn` to the internal list and increments `turn_count`.
Called after a successful LLM response.

**`session.to_messages(system_prompt: str, pending_user: str | None = None) -> list[dict[str, str]]`**
Returns an OpenAI-format message list:
`[{role: system, content: ...}, {role: user, ...}, {role: assistant, ...}, ...]`
The system message is always first. Historical turns are emitted in order.
If `pending_user` is provided, it is appended as a final `{role: user}`
message. This allows callers to build a complete message list — including
the not-yet-recorded user turn — in a single call.

**`session.trim_to_budget(model: str, max_tokens: int, system_prompt: str, pending_user: str) -> None`**
Drops `_turns[0]` repeatedly until
`litellm.token_counter(model, to_messages(system_prompt, pending_user))`
is within `max_tokens`. If the list empties and the token count still
exceeds `max_tokens` (system prompt alone is too large), logs a warning
and proceeds — this is a misconfiguration and will surface as an LLM
context-length error caught by `manager.py`'s error handling.

### Invariant

`turn_count == len(_turns) + number_of_dropped_turns`. The `_turns` list
length may be less than `turn_count` after trimming.

## `manager.py`

### Class

```python
class ConversationManager:
    def __init__(self, scenario: ScenarioConfig, model: str) -> None: ...
    async def chat(self, session: Session, user_input: str) -> str: ...
```

### `chat()` flow

1. **Turn limit check** — if `session.turn_count >= scenario.max_turns`,
   raise `TurnLimitExceeded`. No LLM call is made.
2. **Build system prompt** — minimal placeholder for stage 2:
   a bare hex string (the canary token) is appended on its own line at the
   end of the persona description, without a label identifying its purpose:
   `f"{scenario.persona_name}. {scenario.persona_description}\n{session.canary_token}"`
   Stage 3 replaces this with the full system prompt module.
3. **Trim to budget** — call
   `session.trim_to_budget(model, scenario.token_budget, system_prompt, user_input)`
   before recording the turn.
4. **LLM call** — call
   `session.to_messages(system_prompt, pending_user=user_input)` and pass
   the result to `litellm.acompletion(model=self.model, messages=messages)`.
   Wrap in `try/except Exception` — any failure raises `LLMError` with the
   original exception attached.
5. **Record turn** — call `session.add_turn(user_input, response_text)`.
6. **Return** the assistant response string.

### Exceptions

```python
class TurnLimitExceeded(Exception): ...
class LLMError(Exception): ...
```

`TurnLimitExceeded` — raised at step 1. The REPL catches it, prints
"Maximum turns reached — starting a new session." and rotates to a fresh
`Session.new()`.

`LLMError` — raised at step 4 on any `litellm.acompletion` exception.
The REPL catches it and prints a user-friendly error message without
exposing internal details.

## `tracking.py`

Stub only:
```
"""Stateful attack tracking and circuit breakers — implemented in stage 11."""
```

## `__main__.py` changes

- Replace blocking `input()` with
  `await asyncio.get_event_loop().run_in_executor(None, input, "You: ")`
  so the event loop is not blocked during user input.
- Wrap `load_scenario()` call in `try/except (FileNotFoundError, ValidationError)`
  with a readable error message.
- Instantiate `Session.new()` and `ConversationManager` at startup.
- In the loop, catch `TurnLimitExceeded` and `LLMError` separately.

```python
session = Session.new()
manager = ConversationManager(scenario=scenario, model=load_model())

while True:
    user_input = await asyncio.get_event_loop().run_in_executor(None, input, "You: ")
    user_input = user_input.strip()
    if not user_input:
        continue
    if user_input.lower() == "quit":
        break
    try:
        response = await manager.chat(session, user_input)
    except TurnLimitExceeded:
        print("Maximum turns reached — starting a new session.")
        session = Session.new()
        continue
    except LLMError as e:
        print(f"Error: could not get a response. ({e})")
        continue
    print(f"{scenario.persona_name}: {response}")
```

## Testing

### `tests/unit/conversation/test_session.py`

LLM token counting is mocked via `unittest.mock.patch` on
`litellm.token_counter`.

| Test | What it verifies |
|------|-----------------|
| `test_add_turn_increments_count` | `turn_count` goes 0 → 1 → 2 |
| `test_add_turn_appends` | `to_messages()` grows with each `add_turn()` call |
| `test_to_messages_format` | System message first; alternating user/assistant; correct roles |
| `test_to_messages_with_pending_user` | `pending_user` appended as final user message |
| `test_trim_drops_oldest` | Oldest turn removed first when over budget |
| `test_trim_empty_is_safe` | No error when `_turns` is already empty |
| `test_trim_preserves_turn_count` | `turn_count` does not decrement after trim |
| `test_new_unique_ids` | Two `Session.new()` calls → different `session_id` |
| `test_new_unique_canaries` | Two `Session.new()` calls → different `canary_token` |
| `test_canary_is_32_hex_chars` | `len(canary_token) == 32`; all chars in `0-9a-f` |

### `tests/unit/conversation/test_manager.py`

`litellm.acompletion` is mocked via `unittest.mock.AsyncMock`. No live API
calls in unit tests.

| Test | What it verifies |
|------|-----------------|
| `test_turn_limit_raises` | `TurnLimitExceeded` raised when `turn_count >= max_turns` |
| `test_no_llm_call_on_turn_limit` | `acompletion` not called when limit exceeded |
| `test_chat_increments_turn_count` | `session.turn_count` is 1 after one `chat()` call |
| `test_chat_records_turn` | Correct user/assistant pair in session after `chat()` |
| `test_system_prompt_contains_canary` | Canary token appears in system message sent to LLM |
| `test_canary_only_in_system_message` | Canary token does not appear in any non-system message |
| `test_llm_error_raised_on_failure` | `LLMError` raised when `acompletion` throws |

## Security properties delivered

- **Canary token** — `secrets.token_hex(16)` generated per session, embedded
  in the system prompt as a bare hex string (no identifying label). The
  output pipeline (stage 5) does exact string match. If the token appears in
  output, the system prompt has leaked.
- **Turn limit** — prevents indefinite multi-round escalation attacks.
- **Sliding window** — drops context rather than summarising, per ADR 006.
  No unguarded LLM summarisation call is introduced.
- **Session isolation** — structural: no shared mutable state across
  `Session` instances.
- **Fail closed on LLM error** — exceptions from `litellm` are caught and
  wrapped; no raw provider error reaches the user.
