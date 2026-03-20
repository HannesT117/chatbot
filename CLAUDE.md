# CLAUDE.md - Project Instructions for Claude Code

## Project overview

This project is a testing ground for LLM output guardrails. It evaluates
techniques for ensuring AI-generated responses are correct, compliant, and
appropriate — targeting regulated environments (financial services) and
brand-sensitive contexts (advertising). The goal is to identify which guardrail
combinations are reliable enough for high-stakes production use, and to document
the trade-offs between safety, latency, and user experience.

## Development Workflow

1. Make changes
2. Run typecheck
3. Run tests
4. Lint before committing
5. Before creating PR: run full lint and test suite

## Step Review and Commit Protocol

After completing each implementation step, run a full review before asking the user:

1. **Verify** — run the `superpowers:verification-before-completion` skill (typecheck, tests, lint, confirm output matches plan)
2. **Code review** — dispatch the `superpowers:code-reviewer` subagent to check against the plan and coding standards
3. **QA** — dispatch the `qa` subagent to verify behaviour and edge cases
4. **Simplify** — invoke the `code-simplifier` skill to remove unnecessary complexity
5. **Present to user** — summarise what was built and the review findings, then ask for approval
6. **Commit only after explicit user approval** — never commit speculatively

## Code Style & Conventions

## Commands Reference

See @../docs/commands.md

## Self-Improvement

After every correction or mistake, update this CLAUDE.md with a rule to prevent repeating it. 

## Working with Plan Mode

- Start every complex task in plan mode (shift+tab to cycle)
- Pour energy into the plan so Claude can 1-shot the implementation
- When something goes sideways, switch back to plan mode and re-plan. Don't keep pushing.
- Use plan mode for verification steps too, not just for the build

## Parallel Work

- For tasks that need more compute, use subagents to work in parallel
- Offload individual tasks to subagents to keep the main context window clean and focused
- When working in parallel, only one agent should edit a given file at a time
- For fully parallel workstreams, use git worktrees:
  `git worktree add .claude/worktrees/<name> origin/main`

## Commit message format

- Write the subject line as a short imperative sentence (max 72 chars)
- Write the body as bullet points, not prose paragraphs
- Each bullet describes one logical change and why it was made

## Things Claude Should NOT Do

<!-- CLAUDE-SETUP Add mistakes Claude makes so it learns -->

- Don't use `Any` type hint in Python without an explanatory comment
- Don't skip error handling
- Don't commit without running tests first
- Don't make breaking API changes without discussion

## Favoured patterns

See @../docs/conventions/patterns.md
