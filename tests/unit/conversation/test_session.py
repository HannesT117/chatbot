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
