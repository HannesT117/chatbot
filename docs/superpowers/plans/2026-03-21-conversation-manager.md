# Conversation Manager Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire a real LLM into the echo REPL, adding per-session state management, turn limits, a sliding token-budget window, and canary token injection.

**Architecture:** `Session` owns all conversation state (message history, canary token, turn count) and exposes it only through methods — no public mutable fields. `ConversationManager` holds scenario config and model name, calls `Session` methods and `litellm.acompletion`, and raises typed exceptions the REPL handles. `__main__.py` is the only layer that talks to the user.

**Tech Stack:** Python 3.12, `litellm` (LLM calls + token counting), `pydantic` (config validation), `structlog` (logging), `pytest` + `pytest-asyncio` (testing), `uv` (package management)

**Spec:** `docs/superpowers/specs/2026-03-21-conversation-manager-design.md`

---

## File Map

| File | Change | Responsibility |
|------|--------|---------------|
| `scenarios/financial_advisor.yaml` | Modify | Add `max_turns: 20`, `token_budget: 4000` |
| `scenarios/brand_marketing.yaml` | Modify | Add `max_turns: 30`, `token_budget: 6000` |
| `scenarios/insurance_claims.yaml` | Modify | Add `max_turns: 20`, `token_budget: 4000` |
| `src/chatbot/config.py` | Modify | Add fields to `ScenarioConfig`, add `load_model()` |
| `src/chatbot/conversation/session.py` | Implement | `Turn`, `Session`, all session state logic |
| `src/chatbot/conversation/manager.py` | Implement | `ConversationManager`, `TurnLimitExceeded`, `LLMError` |
| `src/chatbot/conversation/tracking.py` | Implement stub | One-line docstring, filled in stage 11 |
| `src/chatbot/__main__.py` | Rewrite | Async input, session + manager wiring, error handling |
| `tests/unit/test_config.py` | Modify | Add tests for new fields and `load_model()` |
| `tests/unit/conversation/__init__.py` | Already exists | No action needed |
| `tests/unit/conversation/test_session.py` | Create | All `Session` unit tests |
| `tests/unit/conversation/test_manager.py` | Create | All `ConversationManager` unit tests |

---

## Task 1: Extend config with turn limits and model loading

**Files:**
- Modify: `src/chatbot/config.py`
- Modify: `scenarios/financial_advisor.yaml`, `scenarios/brand_marketing.yaml`, `scenarios/insurance_claims.yaml`
- Modify: `tests/unit/test_config.py`

- [ ] **Step 1.1: Add failing tests for max_turns, token_budget, and load_model**

Add these imports and tests to `tests/unit/test_config.py`. The `load_model` import belongs at module level alongside existing imports:

```python
from chatbot.config import ScenarioConfig, load_model, load_scenario
```

Then append these tests to the bottom of the file:

```python
def test_scenario_has_max_turns() -> None:
    cfg = load_scenario("financial_advisor")
    assert cfg.max_turns == 20


def test_scenario_has_token_budget() -> None:
    cfg = load_scenario("financial_advisor")
    assert cfg.token_budget == 4000


@pytest.mark.parametrize("name", ["financial_advisor", "brand_marketing", "insurance_claims"])
def test_all_scenarios_have_limits(name: str) -> None:
    cfg = load_scenario(name)
    assert cfg.max_turns > 0
    assert cfg.token_budget > 0


def test_load_model_default(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("CHATBOT_MODEL", raising=False)
    assert load_model() == "gpt-4o-mini"


def test_load_model_reads_env_var(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("CHATBOT_MODEL", "claude-haiku-4-5-20251001")
    assert load_model() == "claude-haiku-4-5-20251001"
```

- [ ] **Step 1.2: Run new tests to confirm they fail**

```bash
uv run pytest tests/unit/test_config.py -k "max_turns or token_budget or load_model" -v
```

Expected: 5 failures — `max_turns` / `token_budget` not in `ScenarioConfig`, `load_model` not defined.

- [ ] **Step 1.3: Add fields to ScenarioConfig and add load_model()**

`os` is already imported at the top of `src/chatbot/config.py`. Add `_DEFAULT_MODEL` and `load_model()` after the `_DEFAULT_SCENARIO` line, and add the two new fields to `ScenarioConfig`. The final `ScenarioConfig` and the new constant+function look like this:

```python
_DEFAULT_MODEL = "gpt-4o-mini"


class ScenarioConfig(BaseModel):
    """Typed representation of a scenario YAML file."""

    name: str
    persona_name: str
    persona_description: str
    max_turns: int
    token_budget: int
    allowed_intents: list[str]
    blocklist_terms: list[str]
    output_constraints: list[str]


def load_model() -> str:
    """Return the LLM model name from CHATBOT_MODEL env var (default: gpt-4o-mini)."""
    return os.environ.get("CHATBOT_MODEL", _DEFAULT_MODEL)
```

- [ ] **Step 1.4: Add max_turns and token_budget to each scenario YAML**

`scenarios/financial_advisor.yaml` — add after `persona_description`:
```yaml
max_turns: 20
token_budget: 4000
```

`scenarios/brand_marketing.yaml` — add after `persona_description`:
```yaml
max_turns: 30
token_budget: 6000
```

`scenarios/insurance_claims.yaml` — add after `persona_description`:
```yaml
max_turns: 20
token_budget: 4000
```

- [ ] **Step 1.5: Run all config tests to confirm they pass**

```bash
uv run pytest tests/unit/test_config.py -v
```

Expected: all tests pass (the new required fields must be present in all 3 YAMLs, or existing scenario-loading tests will also fail with a `ValidationError`).

- [ ] **Step 1.6: Run typecheck**

```bash
uv run mypy src/
```

Expected: no errors.

- [ ] **Step 1.7: Commit**

```bash
git add src/chatbot/config.py scenarios/financial_advisor.yaml scenarios/brand_marketing.yaml scenarios/insurance_claims.yaml tests/unit/test_config.py
git commit -m "Add turn limits and model config to ScenarioConfig"
```

---

## Task 2: Session — core state and isolation

**Files:**
- Implement: `src/chatbot/conversation/session.py`
- Create: `tests/unit/conversation/test_session.py`

- [ ] **Step 2.1: Write failing tests for Session construction**

Create `tests/unit/conversation/test_session.py`:

```python
"""Unit tests for src/chatbot/conversation/session.py."""

from __future__ import annotations

from chatbot.conversation.session import Session


def test_new_produces_unique_session_ids() -> None:
    s1 = Session.new()
    s2 = Session.new()
    assert s1.session_id != s2.session_id


def test_new_produces_unique_canary_tokens() -> None:
    s1 = Session.new()
    s2 = Session.new()
    assert s1.canary_token != s2.canary_token


def test_canary_is_32_hex_chars() -> None:
    session = Session.new()
    assert len(session.canary_token) == 32
    assert all(c in "0123456789abcdef" for c in session.canary_token)


def test_new_session_has_zero_turn_count() -> None:
    session = Session.new()
    assert session.turn_count == 0
```

- [ ] **Step 2.2: Run to confirm failure**

```bash
uv run pytest tests/unit/conversation/test_session.py -v
```

Expected: import error — `Session` not yet defined.

- [ ] **Step 2.3: Implement Turn and Session core in session.py**

Replace contents of `src/chatbot/conversation/session.py`:

```python
"""Conversation session — per-session state and sliding message history."""

from __future__ import annotations

import secrets
import uuid
from dataclasses import dataclass


@dataclass(frozen=True)
class Turn:
    """One complete exchange: a user message and the assistant's reply."""

    user: str
    assistant: str


class Session:
    """Holds all state for a single conversation.

    Public attributes (session_id, canary_token) are immutable strings set at
    construction. The message history is private; callers use add_turn(),
    to_messages(), and trim_to_budget() to read and write it.
    """

    def __init__(self, session_id: str, canary_token: str) -> None:
        self.session_id = session_id
        self.canary_token = canary_token
        self._turns: list[Turn] = []
        self._turn_count: int = 0

    @classmethod
    def new(cls) -> Session:
        """Create a fresh session with a unique ID and canary token."""
        return cls(
            session_id=uuid.uuid4().hex,
            canary_token=secrets.token_hex(16),
        )

    @property
    def turn_count(self) -> int:
        """Total turns taken in this session, including trimmed ones."""
        return self._turn_count
```

- [ ] **Step 2.4: Run tests to confirm they pass**

```bash
uv run pytest tests/unit/conversation/test_session.py -v
```

Expected: 4 tests pass.

- [ ] **Step 2.5: Commit**

```bash
git add src/chatbot/conversation/session.py tests/unit/conversation/test_session.py
git commit -m "Add Session core: unique ID, canary token, turn count"
```

---

## Task 3: Session — message history (add_turn, to_messages)

**Files:**
- Modify: `src/chatbot/conversation/session.py`
- Modify: `tests/unit/conversation/test_session.py`

- [ ] **Step 3.1: Add failing tests for add_turn and to_messages**

Append to `tests/unit/conversation/test_session.py`:

```python
def test_add_turn_increments_count() -> None:
    session = Session.new()
    session.add_turn("hello", "hi there")
    assert session.turn_count == 1
    session.add_turn("how are you", "fine")
    assert session.turn_count == 2


def test_add_turn_appends_to_history() -> None:
    session = Session.new()
    session.add_turn("first", "reply one")
    session.add_turn("second", "reply two")
    msgs = session.to_messages("sys")
    contents = [m["content"] for m in msgs]
    assert "first" in contents
    assert "second" in contents
    assert "reply one" in contents
    assert "reply two" in contents


def test_to_messages_system_first() -> None:
    session = Session.new()
    msgs = session.to_messages("You are helpful.")
    assert msgs[0] == {"role": "system", "content": "You are helpful."}


def test_to_messages_empty_history() -> None:
    session = Session.new()
    msgs = session.to_messages("sys")
    assert msgs == [{"role": "system", "content": "sys"}]


def test_to_messages_with_turns() -> None:
    session = Session.new()
    session.add_turn("hello", "hi")
    msgs = session.to_messages("sys")
    assert msgs == [
        {"role": "system", "content": "sys"},
        {"role": "user", "content": "hello"},
        {"role": "assistant", "content": "hi"},
    ]


def test_to_messages_with_pending_user() -> None:
    session = Session.new()
    msgs = session.to_messages("sys", pending_user="what time is it?")
    assert msgs[-1] == {"role": "user", "content": "what time is it?"}


def test_to_messages_pending_user_after_history() -> None:
    session = Session.new()
    session.add_turn("a", "b")
    msgs = session.to_messages("sys", pending_user="c")
    # system + user + assistant + pending user
    assert len(msgs) == 4
    assert msgs[-1] == {"role": "user", "content": "c"}
```

- [ ] **Step 3.2: Run to confirm failure**

```bash
uv run pytest tests/unit/conversation/test_session.py -v
```

Expected: 7 new failures — `add_turn` and `to_messages` not yet defined.

- [ ] **Step 3.3: Implement add_turn and to_messages**

Add these methods to the `Session` class in `session.py`:

```python
    def add_turn(self, user: str, assistant: str) -> None:
        """Record a completed exchange and increment the turn counter."""
        self._turns.append(Turn(user=user, assistant=assistant))
        self._turn_count += 1

    def to_messages(
        self, system_prompt: str, pending_user: str | None = None
    ) -> list[dict[str, str]]:
        """Build an OpenAI-format message list from the current history.

        The system prompt is always first. Historical turns follow in order.
        If *pending_user* is provided it is appended as a final user message,
        allowing the caller to include the current (not yet recorded) input.
        """
        messages: list[dict[str, str]] = [{"role": "system", "content": system_prompt}]
        for turn in self._turns:
            messages.append({"role": "user", "content": turn.user})
            messages.append({"role": "assistant", "content": turn.assistant})
        if pending_user is not None:
            messages.append({"role": "user", "content": pending_user})
        return messages
```

- [ ] **Step 3.4: Run tests**

```bash
uv run pytest tests/unit/conversation/test_session.py -v
```

Expected: all 11 tests pass.

- [ ] **Step 3.5: Commit**

```bash
git add src/chatbot/conversation/session.py tests/unit/conversation/test_session.py
git commit -m "Add Session message history: add_turn, to_messages"
```

---

## Task 4: Session — sliding window (trim_to_budget)

**Files:**
- Modify: `src/chatbot/conversation/session.py`
- Modify: `tests/unit/conversation/test_session.py`

- [ ] **Step 4.1: Add failing tests for trim_to_budget**

Append to `tests/unit/conversation/test_session.py`:

```python
from unittest.mock import patch


def test_trim_drops_oldest_turn() -> None:
    session = Session.new()
    session.add_turn("old question", "old answer")
    session.add_turn("new question", "new answer")

    def fake_counter(model: str, messages: list[dict[str, str]]) -> int:
        user_messages = [m for m in messages if m["role"] == "user"]
        return 999 if len(user_messages) > 1 else 10

    with patch("litellm.token_counter", side_effect=fake_counter):
        session.trim_to_budget("gpt-4o-mini", 50, "sys", "pending")

    msgs = session.to_messages("sys")
    contents = [m["content"] for m in msgs]
    assert "new question" in contents
    assert "old question" not in contents


def test_trim_preserves_turn_count() -> None:
    session = Session.new()
    session.add_turn("hello", "hi")
    session.add_turn("bye", "goodbye")
    assert session.turn_count == 2

    # Always over budget — will trim everything
    with patch("litellm.token_counter", return_value=9999):
        session.trim_to_budget("gpt-4o-mini", 100, "sys", "pending")

    # turn_count must not decrement
    assert session.turn_count == 2
    # history should be empty
    assert session.to_messages("sys") == [{"role": "system", "content": "sys"}]


def test_trim_empty_turns_is_safe() -> None:
    session = Session.new()
    with patch("litellm.token_counter", return_value=9999):
        session.trim_to_budget("gpt-4o-mini", 100, "sys", "pending")
    assert session.turn_count == 0


def test_trim_does_nothing_when_within_budget() -> None:
    session = Session.new()
    session.add_turn("a", "b")
    session.add_turn("c", "d")

    with patch("litellm.token_counter", return_value=10):
        session.trim_to_budget("gpt-4o-mini", 100, "sys", "pending")

    msgs = session.to_messages("sys")
    turn_messages = [m for m in msgs if m["role"] in ("user", "assistant")]
    assert len(turn_messages) == 4  # both turns still present
```

- [ ] **Step 4.2: Run to confirm failure**

```bash
uv run pytest tests/unit/conversation/test_session.py -k "trim" -v
```

Expected: 4 failures — `trim_to_budget` not yet defined.

- [ ] **Step 4.3: Implement trim_to_budget**

Add these imports at the top of `session.py` (after existing imports):

```python
import litellm
import structlog
```

Add this method to the `Session` class:

```python
    def trim_to_budget(
        self, model: str, max_tokens: int, system_prompt: str, pending_user: str
    ) -> None:
        """Drop oldest turns until the full message list fits within max_tokens.

        Uses litellm.token_counter to measure token usage including the pending
        user message. turn_count is never decremented — it records total turns
        taken in this session, not the current window size.

        If even an empty history (system prompt + pending user alone) exceeds
        max_tokens, logs a warning and proceeds — the provider will surface a
        context-length error which manager.py catches as LLMError.
        """
        log = structlog.get_logger()

        while self._turns:
            messages = self.to_messages(system_prompt, pending_user)
            if litellm.token_counter(model=model, messages=messages) <= max_tokens:
                return
            self._turns.pop(0)

        # _turns is empty; check once whether the system prompt alone exceeds budget
        messages = self.to_messages(system_prompt, pending_user)
        token_count = litellm.token_counter(model=model, messages=messages)
        if token_count > max_tokens:
            log.warning(
                "system_prompt_exceeds_token_budget",
                token_count=token_count,
                max_tokens=max_tokens,
            )
```

- [ ] **Step 4.4: Run all session tests**

```bash
uv run pytest tests/unit/conversation/test_session.py -v
```

Expected: all 15 tests pass.

- [ ] **Step 4.5: Run typecheck**

```bash
uv run mypy src/
```

Expected: no errors.

- [ ] **Step 4.6: Commit**

```bash
git add src/chatbot/conversation/session.py tests/unit/conversation/test_session.py
git commit -m "Add Session sliding window: trim_to_budget drops oldest turns"
```

---

## Task 5: ConversationManager — turn limits and LLM integration

**Files:**
- Implement: `src/chatbot/conversation/manager.py`
- Create: `tests/unit/conversation/test_manager.py`

- [ ] **Step 5.1: Write failing tests for ConversationManager**

Create `tests/unit/conversation/test_manager.py`:

```python
"""Unit tests for src/chatbot/conversation/manager.py.

All litellm.acompletion calls are mocked — no live API calls.
"""

from __future__ import annotations

from typing import Any
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from chatbot.config import ScenarioConfig
from chatbot.conversation.manager import ConversationManager, LLMError, TurnLimitExceeded
from chatbot.conversation.session import Session


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture
def scenario() -> ScenarioConfig:
    return ScenarioConfig(
        name="test",
        persona_name="TestBot",
        persona_description="A test assistant.",
        max_turns=3,
        token_budget=4000,
        allowed_intents=[],
        blocklist_terms=[],
        output_constraints=[],
    )


@pytest.fixture
def session() -> Session:
    return Session.new()


@pytest.fixture
def manager(scenario: ScenarioConfig) -> ConversationManager:
    return ConversationManager(scenario=scenario, model="gpt-4o-mini")


def _mock_completion(content: str) -> MagicMock:
    mock = MagicMock()
    mock.choices[0].message.content = content
    return mock


# ---------------------------------------------------------------------------
# Turn limit
# ---------------------------------------------------------------------------


async def test_turn_limit_raises(manager: ConversationManager, session: Session) -> None:
    for i in range(3):  # fill to max_turns=3
        session.add_turn(f"u{i}", f"a{i}")
    with pytest.raises(TurnLimitExceeded):
        await manager.chat(session, "one more")


async def test_no_llm_call_on_turn_limit(
    manager: ConversationManager, session: Session
) -> None:
    for i in range(3):
        session.add_turn(f"u{i}", f"a{i}")
    with patch("litellm.acompletion") as mock_llm:
        with pytest.raises(TurnLimitExceeded):
            await manager.chat(session, "one more")
    mock_llm.assert_not_called()


# ---------------------------------------------------------------------------
# Successful LLM call
# ---------------------------------------------------------------------------


async def test_chat_returns_response(
    manager: ConversationManager, session: Session
) -> None:
    mock_resp = _mock_completion("Hello!")
    with patch("litellm.acompletion", new=AsyncMock(return_value=mock_resp)):
        result = await manager.chat(session, "hi")
    assert result == "Hello!"


async def test_chat_increments_turn_count(
    manager: ConversationManager, session: Session
) -> None:
    mock_resp = _mock_completion("ok")
    with patch("litellm.acompletion", new=AsyncMock(return_value=mock_resp)):
        await manager.chat(session, "hi")
    assert session.turn_count == 1


async def test_chat_records_turn(
    manager: ConversationManager, session: Session
) -> None:
    mock_resp = _mock_completion("I'm fine.")
    with patch("litellm.acompletion", new=AsyncMock(return_value=mock_resp)):
        await manager.chat(session, "how are you?")
    msgs = session.to_messages("sys")
    contents = [m["content"] for m in msgs]
    assert "how are you?" in contents
    assert "I'm fine." in contents


# ---------------------------------------------------------------------------
# Canary token
# ---------------------------------------------------------------------------


async def test_system_prompt_contains_canary(
    manager: ConversationManager, session: Session
) -> None:
    captured: list[list[dict[str, str]]] = []

    async def capture_call(**kwargs: Any) -> MagicMock:
        captured.append(kwargs["messages"])
        return _mock_completion("ok")

    with patch("litellm.acompletion", new=capture_call):
        await manager.chat(session, "test")

    system_msg = captured[0][0]
    assert system_msg["role"] == "system"
    assert session.canary_token in system_msg["content"]


async def test_canary_only_in_system_message(
    manager: ConversationManager, session: Session
) -> None:
    captured: list[list[dict[str, str]]] = []

    async def capture_call(**kwargs: Any) -> MagicMock:
        captured.append(kwargs["messages"])
        return _mock_completion("ok")

    with patch("litellm.acompletion", new=capture_call):
        await manager.chat(session, "test")

    non_system = [m for m in captured[0] if m["role"] != "system"]
    for msg in non_system:
        assert session.canary_token not in msg["content"]


# ---------------------------------------------------------------------------
# LLM error handling
# ---------------------------------------------------------------------------


async def test_llm_error_raised_on_failure(
    manager: ConversationManager, session: Session
) -> None:
    with patch(
        "litellm.acompletion", new=AsyncMock(side_effect=RuntimeError("API down"))
    ):
        with pytest.raises(LLMError):
            await manager.chat(session, "hi")
```

- [ ] **Step 5.2: Run to confirm failure**

```bash
uv run pytest tests/unit/conversation/test_manager.py -v
```

Expected: import error — `ConversationManager`, `TurnLimitExceeded`, `LLMError` not yet defined.

- [ ] **Step 5.3: Implement ConversationManager**

Replace contents of `src/chatbot/conversation/manager.py`:

```python
"""Conversation manager — enforces turn limits, injects canary, calls the LLM."""

from __future__ import annotations

import litellm

from chatbot.config import ScenarioConfig
from chatbot.conversation.session import Session


class TurnLimitExceeded(Exception):
    """Raised when a session has reached its maximum number of turns."""


class LLMError(Exception):
    """Raised when the LLM call fails for any reason."""


class ConversationManager:
    """Orchestrates one turn of conversation.

    Checks limits, builds the system prompt with canary injection, trims the
    context window, calls the LLM, and records the result — in that order.
    """

    def __init__(self, scenario: ScenarioConfig, model: str) -> None:
        self._scenario = scenario
        self._model = model

    async def chat(self, session: Session, user_input: str) -> str:
        """Advance the conversation by one turn and return the assistant reply.

        Raises
        ------
        TurnLimitExceeded
            If the session has already reached max_turns. No LLM call is made.
        LLMError
            If litellm.acompletion raises for any reason.
        """
        if session.turn_count >= self._scenario.max_turns:
            raise TurnLimitExceeded(
                f"Session reached the {self._scenario.max_turns}-turn limit."
            )

        system_prompt = (
            f"{self._scenario.persona_name}. {self._scenario.persona_description}\n"
            f"{session.canary_token}"
        )

        session.trim_to_budget(
            self._model,
            self._scenario.token_budget,
            system_prompt,
            user_input,
        )

        messages = session.to_messages(system_prompt, pending_user=user_input)

        try:
            response = await litellm.acompletion(model=self._model, messages=messages)
        except Exception as exc:
            raise LLMError(str(exc)) from exc

        response_text: str = response.choices[0].message.content
        session.add_turn(user_input, response_text)
        return response_text
```

- [ ] **Step 5.4: Run all manager tests**

```bash
uv run pytest tests/unit/conversation/test_manager.py -v
```

Expected: all 9 tests pass.

- [ ] **Step 5.5: Run full unit suite**

```bash
uv run pytest tests/unit/ -v
```

Expected: all tests pass.

- [ ] **Step 5.6: Run typecheck**

```bash
uv run mypy src/
```

Expected: no errors.

- [ ] **Step 5.7: Commit**

```bash
git add src/chatbot/conversation/manager.py tests/unit/conversation/test_manager.py
git commit -m "Implement ConversationManager with turn limits and LLM integration"
```

---

## Task 6: tracking.py stub and __main__.py wiring

**Files:**
- Implement stub: `src/chatbot/conversation/tracking.py`
- Rewrite: `src/chatbot/__main__.py`

- [ ] **Step 6.1: Write tracking.py stub**

Replace contents of `src/chatbot/conversation/tracking.py`:

```python
"""Stateful attack tracking and circuit breakers — implemented in stage 11."""
```

- [ ] **Step 6.2: Rewrite __main__.py**

Replace contents of `src/chatbot/__main__.py`:

```python
"""CLI entry point — run with ``python -m chatbot``."""

from __future__ import annotations

import asyncio

from pydantic import ValidationError

from chatbot.config import load_model, load_scenario
from chatbot.conversation.manager import ConversationManager, LLMError, TurnLimitExceeded
from chatbot.conversation.session import Session


async def main() -> None:
    try:
        scenario = load_scenario()
    except (FileNotFoundError, ValidationError) as exc:
        print(f"Error: could not load scenario — {exc}")
        return

    model = load_model()
    session = Session.new()
    manager = ConversationManager(scenario=scenario, model=model)

    print(f"Chatbot active: {scenario.persona_name} (scenario: {scenario.name})")
    print(
        f"Model: {model} | Max turns: {scenario.max_turns} "
        f"| Token budget: {scenario.token_budget}"
    )
    print("Type 'quit' or press Ctrl-C to exit.\n")

    loop = asyncio.get_running_loop()

    while True:
        try:
            user_input = await loop.run_in_executor(None, input, "You: ")
        except (EOFError, KeyboardInterrupt):
            print("\nGoodbye.")
            break

        user_input = user_input.strip()

        if not user_input:
            continue

        if user_input.lower() == "quit":
            print("Goodbye.")
            break

        try:
            response = await manager.chat(session, user_input)
        except TurnLimitExceeded:
            print("\nMaximum turns reached — starting a new session.\n")
            session = Session.new()
            continue
        except LLMError as exc:
            print(f"\nError: could not get a response. ({exc})\n")
            continue

        print(f"\n{scenario.persona_name}: {response}\n")


if __name__ == "__main__":
    asyncio.run(main())
```

- [ ] **Step 6.3: Run the full unit test suite**

```bash
uv run pytest tests/unit/ -v
```

Expected: all tests pass.

- [ ] **Step 6.4: Run typecheck on all source files**

```bash
uv run mypy src/
```

Expected: no errors.

- [ ] **Step 6.5: Run linter**

```bash
uv run ruff check src/ tests/
```

Expected: no issues.

- [ ] **Step 6.6: Commit**

```bash
git add src/chatbot/conversation/tracking.py src/chatbot/__main__.py
git commit -m "Wire Session and ConversationManager into async REPL"
```

---

## Done

At this point `uv run python -m chatbot` starts a real multi-round conversation
with the configured LLM, enforces turn limits, uses a sliding token-budget
window, and injects a canary token into every system prompt.

All unit tests pass with no live API calls. Integration testing (with a real
model) is gated behind `CHATBOT_LIVE_TESTS=1` as per the project test strategy.
