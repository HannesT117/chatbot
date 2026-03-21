# Commands

## Run Dev

```sh
uv run python -m chatbot
```

## Test

```sh
uv run pytest tests/unit/
```

## Test (all, including integration)

```sh
CHATBOT_LIVE_TESTS=1 uv run pytest
```

## Typecheck

```sh
uv run mypy src/
```

## Lint

```sh
uv run ruff check src/ tests/
```

## Format

```sh
uv run ruff format src/ tests/
```

## Clean Build

```sh
uv sync --reinstall
```
