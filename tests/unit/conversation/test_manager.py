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
    with patch("litellm.acompletion", new=AsyncMock()) as mock_llm, pytest.raises(
        TurnLimitExceeded
    ):
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

    async def capture_call(**kwargs: Any) -> MagicMock:  # Any: litellm kwargs are untyped
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

    async def capture_call(**kwargs: Any) -> MagicMock:  # Any: litellm kwargs are untyped
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
    ), pytest.raises(LLMError):
        await manager.chat(session, "hi")
