"""Scenario config loading and environment variable handling.

The active scenario is selected via the ``CHATBOT_SCENARIO`` environment
variable (default: ``financial_advisor``).  Scenario definitions live in
``scenarios/<name>.yaml`` relative to the project root.
"""

from __future__ import annotations

import os
from pathlib import Path

import yaml
from pydantic import BaseModel

# ---------------------------------------------------------------------------
# Typed model
# ---------------------------------------------------------------------------


class ScenarioConfig(BaseModel):
    """Typed representation of a scenario YAML file."""

    name: str
    persona_name: str
    persona_description: str
    allowed_intents: list[str]
    blocklist_terms: list[str]
    output_constraints: list[str]


# ---------------------------------------------------------------------------
# Loader
# ---------------------------------------------------------------------------

# Project root is three levels up from this file:
# src/chatbot/config.py → src/chatbot/ → src/ → <project root>
_PROJECT_ROOT: Path = Path(__file__).resolve().parent.parent.parent
_SCENARIOS_DIR: Path = _PROJECT_ROOT / "scenarios"

_DEFAULT_SCENARIO = "financial_advisor"


def load_scenario(name: str | None = None) -> ScenarioConfig:
    """Load and return the ``ScenarioConfig`` for *name*.

    If *name* is ``None`` the value of the ``CHATBOT_SCENARIO`` environment
    variable is used, falling back to ``"financial_advisor"``.

    Raises
    ------
    FileNotFoundError
        If the scenario YAML file does not exist.
    """
    resolved_name = name or os.environ.get("CHATBOT_SCENARIO", _DEFAULT_SCENARIO)
    scenario_path = _SCENARIOS_DIR / f"{resolved_name}.yaml"

    if not scenario_path.exists():
        available = [p.stem for p in _SCENARIOS_DIR.glob("*.yaml")]
        raise FileNotFoundError(
            f"Scenario '{resolved_name}' not found at {scenario_path}. "
            f"Available scenarios: {available or ['(none)']}"
        )

    with scenario_path.open() as fh:
        raw = yaml.safe_load(fh)

    return ScenarioConfig(**raw)
