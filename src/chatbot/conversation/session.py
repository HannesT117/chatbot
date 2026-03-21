"""Conversation session — per-session state and sliding message history."""

from __future__ import annotations

import secrets
import uuid
from dataclasses import dataclass

import litellm
import structlog


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
        self._session_id = session_id
        self._canary_token = canary_token
        self._turns: list[Turn] = []
        # _turn_count intentionally diverges from len(_turns) after trim_to_budget()
        # drops old turns — it records the total turns taken, never decrements.
        self._turn_count: int = 0

    @classmethod
    def new(cls) -> Session:
        """Create a fresh session with a unique ID and canary token."""
        return cls(
            session_id=uuid.uuid4().hex,
            canary_token=secrets.token_hex(16),
        )

    @property
    def session_id(self) -> str:
        """Unique session identifier, set at construction."""
        return self._session_id

    @property
    def canary_token(self) -> str:
        """Security canary token, set at construction."""
        return self._canary_token

    @property
    def turn_count(self) -> int:
        """Total turns taken in this session, including trimmed ones."""
        return self._turn_count

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
