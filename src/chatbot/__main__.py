"""CLI entry point — run with ``python -m chatbot``.

Starts a simple async REPL that echoes user input.
The full pipeline is not wired yet; this establishes the conversation loop
that later steps will plug into.
"""

from __future__ import annotations

import asyncio

from chatbot.config import load_scenario


async def main() -> None:
    scenario = load_scenario()
    print(f"Chatbot active scenario: {scenario.name} (persona: {scenario.persona_name})")
    print("Type 'quit' or press Ctrl-C to exit.\n")

    while True:
        try:
            user_input = input("You: ").strip()
        except (EOFError, KeyboardInterrupt):
            print("\nGoodbye.")
            break

        if user_input.lower() == "quit":
            print("Goodbye.")
            break

        if not user_input:
            continue

        print(f"Echo: {user_input}\n")


if __name__ == "__main__":
    asyncio.run(main())
