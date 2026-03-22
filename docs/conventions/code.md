# Conventions

## General

- Use descriptive variable names
- Keep functions small and focused
- Write tests for new functionality
- Handle errors explicitly, don't swallow them

## Go (server/)

- Use `go vet ./...` for static analysis
- Errors are returned, not panicked — handle every error explicitly
- Use interfaces for testability; mock via interface, not concrete types
- Use `slog` for structured logging — no `fmt.Println` in production code
- Keep packages small and focused; avoid circular imports

## TypeScript (web/)

- Use `npx tsc --noEmit` for typechecking
- Use `npm run lint` (ESLint) for linting
- No `any` without a comment justifying why it cannot be avoided
- Prefer explicit return types on exported functions
