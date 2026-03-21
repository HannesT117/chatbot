"""Unit tests for src/chatbot/config.py."""

from __future__ import annotations

import pytest

from chatbot.config import ScenarioConfig, load_model, load_scenario

# ---------------------------------------------------------------------------
# Loading each scenario by name
# ---------------------------------------------------------------------------


def test_load_financial_advisor() -> None:
    cfg = load_scenario("financial_advisor")

    assert isinstance(cfg, ScenarioConfig)
    assert cfg.name == "financial_advisor"
    assert cfg.persona_name == "Morgan"
    assert "general_financial_education" in cfg.allowed_intents
    assert "invest in" in cfg.blocklist_terms
    assert any("investment advice" in c.lower() for c in cfg.output_constraints)


def test_load_brand_marketing() -> None:
    cfg = load_scenario("brand_marketing")

    assert cfg.name == "brand_marketing"
    assert cfg.persona_name == "Sage"
    assert "social_media_post" in cfg.allowed_intents
    assert "patagonia" in cfg.blocklist_terms
    assert any("competitor" in c.lower() for c in cfg.output_constraints)


def test_load_insurance_claims() -> None:
    cfg = load_scenario("insurance_claims")

    assert cfg.name == "insurance_claims"
    assert cfg.persona_name == "Dana"
    assert "file_new_claim" in cfg.allowed_intents
    assert "guarantee" in cfg.blocklist_terms
    assert any("legal advice" in c.lower() for c in cfg.output_constraints)


# ---------------------------------------------------------------------------
# All scenarios return well-formed ScenarioConfig objects
# ---------------------------------------------------------------------------


@pytest.mark.parametrize("name", ["financial_advisor", "brand_marketing", "insurance_claims"])
def test_scenario_fields_populated(name: str) -> None:
    cfg = load_scenario(name)

    assert cfg.name == name
    assert cfg.persona_name  # non-empty string
    assert cfg.persona_description  # non-empty string
    assert len(cfg.allowed_intents) > 0
    assert len(cfg.blocklist_terms) > 0
    assert len(cfg.output_constraints) > 0


# ---------------------------------------------------------------------------
# CHATBOT_SCENARIO env var switching
# ---------------------------------------------------------------------------


def test_env_var_selects_scenario(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("CHATBOT_SCENARIO", "brand_marketing")
    cfg = load_scenario()  # no explicit name — should read env var
    assert cfg.name == "brand_marketing"


def test_env_var_overridden_by_explicit_name(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("CHATBOT_SCENARIO", "brand_marketing")
    # Explicit name takes precedence over env var
    cfg = load_scenario("insurance_claims")
    assert cfg.name == "insurance_claims"


def test_default_scenario_without_env_var(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("CHATBOT_SCENARIO", raising=False)
    cfg = load_scenario()
    assert cfg.name == "financial_advisor"


# ---------------------------------------------------------------------------
# Error on unknown scenario
# ---------------------------------------------------------------------------


def test_unknown_scenario_raises_file_not_found() -> None:
    with pytest.raises(FileNotFoundError) as exc_info:
        load_scenario("nonexistent_scenario")

    msg = str(exc_info.value)
    assert "nonexistent_scenario" in msg


def test_error_message_lists_available_scenarios() -> None:
    with pytest.raises(FileNotFoundError) as exc_info:
        load_scenario("does_not_exist")

    msg = str(exc_info.value)
    # Should mention at least one real scenario to help the user
    assert "financial_advisor" in msg or "Available scenarios" in msg


# ---------------------------------------------------------------------------
# Max turns and token budget
# ---------------------------------------------------------------------------


def test_scenario_has_max_turns() -> None:
    cfg = load_scenario("financial_advisor")
    assert cfg.max_turns == 20


def test_scenario_has_token_budget() -> None:
    cfg = load_scenario("financial_advisor")
    assert cfg.token_budget == 4000


@pytest.mark.parametrize("name", ["financial_advisor", "brand_marketing", "insurance_claims"])
def test_all_scenarios_have_limits(name: str) -> None:
    cfg = load_scenario(name)
    assert cfg.max_turns > 0
    assert cfg.token_budget > 0


# ---------------------------------------------------------------------------
# Model loading
# ---------------------------------------------------------------------------


def test_load_model_default(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("CHATBOT_MODEL", raising=False)
    assert load_model() == "gpt-4o-mini"


def test_load_model_reads_env_var(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("CHATBOT_MODEL", "claude-haiku-4-5-20251001")
    assert load_model() == "claude-haiku-4-5-20251001"
