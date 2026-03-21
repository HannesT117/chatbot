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
