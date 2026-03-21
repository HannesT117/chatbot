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
