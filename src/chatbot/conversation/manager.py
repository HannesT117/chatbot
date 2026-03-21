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

        raw = response.choices[0].message.content
        if raw is None:
            raise LLMError("LLM returned no text content (content=None).")
        response_text: str = raw
        session.add_turn(user_input, response_text)
        return response_text
