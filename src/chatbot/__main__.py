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
