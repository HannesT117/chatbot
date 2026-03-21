# Conventions

## General

- Use descriptive variable names
- Keep functions small and focused
- Write tests for new functionality
- Handle errors explicitly, don't swallow them

## Python

- Use type hints everywhere — all function signatures and class attributes must be typed
- Use `pydantic` for data models; prefer `BaseModel` over dataclasses for anything crossing a boundary
- Prefer `TypeAlias` and `Protocol` over class inheritance for structural typing
- No `Any` without an explanatory comment justifying why it cannot be avoided
- Use `ruff` for linting and formatting (`uv run ruff check` / `uv run ruff format`)
- Use `mypy` in strict mode (`uv run mypy src/`)
- Use `from __future__ import annotations` at the top of every module for deferred evaluation
