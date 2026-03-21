"""Unit tests for src/chatbot/conversation/session.py."""

from __future__ import annotations

from unittest.mock import patch

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


def test_trim_drops_oldest_turn() -> None:
    session = Session.new()
    session.add_turn("old question", "old answer")
    session.add_turn("new question", "new answer")

    def fake_counter(model: str, messages: list[dict[str, str]]) -> int:
        user_messages = [m for m in messages if m["role"] == "user"]
        return 999 if len(user_messages) > 2 else 10

    with patch("chatbot.conversation.session.litellm.token_counter", side_effect=fake_counter):
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
    with patch("chatbot.conversation.session.litellm.token_counter", return_value=9999):
        session.trim_to_budget("gpt-4o-mini", 100, "sys", "pending")

    # turn_count must not decrement
    assert session.turn_count == 2
    # history should be empty
    assert session.to_messages("sys") == [{"role": "system", "content": "sys"}]


def test_trim_empty_turns_is_safe() -> None:
    session = Session.new()
    with patch("chatbot.conversation.session.litellm.token_counter", return_value=9999):
        session.trim_to_budget("gpt-4o-mini", 100, "sys", "pending")
    assert session.turn_count == 0


def test_trim_does_nothing_when_within_budget() -> None:
    session = Session.new()
    session.add_turn("a", "b")
    session.add_turn("c", "d")

    with patch("chatbot.conversation.session.litellm.token_counter", return_value=10):
        session.trim_to_budget("gpt-4o-mini", 100, "sys", "pending")

    msgs = session.to_messages("sys")
    turn_messages = [m for m in msgs if m["role"] in ("user", "assistant")]
    assert len(turn_messages) == 4  # both turns still present


def test_trim_logs_warning_when_system_prompt_alone_exceeds_budget() -> None:
    session = Session.new()
    with (
        patch("chatbot.conversation.session.litellm.token_counter", return_value=9999),
        patch("chatbot.conversation.session.structlog.get_logger") as mock_get_logger,
    ):
        mock_log = mock_get_logger.return_value
        session.trim_to_budget("gpt-4o-mini", 100, "sys", "pending")
        mock_log.warning.assert_called_once_with(
            "system_prompt_exceeds_token_budget",
            token_count=9999,
            max_tokens=100,
        )
