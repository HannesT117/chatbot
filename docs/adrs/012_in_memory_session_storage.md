# 012 — In-memory session storage with pluggable interface

**Status:** Accepted
**Date:** March 22, 2026

## Context

The Go server manages per-session conversation state: message history, turn
count, canary token, and sliding window position. This state must be stored
somewhere.

Three options were considered:

**In-memory.** Go struct behind a `sync.Mutex`. Simple, fast, zero external
dependencies. Sessions are lost on server restart.

**SQLite.** Embedded database, no external dependencies, sessions survive
restarts.

**Redis.** External dependency, standard for production session stores,
supports horizontal scaling.

## Decision

In-memory storage, behind an interface that allows swapping the backend later.

The session store is defined as a Go interface:

```go
type SessionStore interface {
    Get(id string) (*Session, error)
    Save(session *Session) error
    Delete(id string) error
}
```

The initial implementation is a `sync.Mutex`-protected `map[string]*Session`.

### Why in-memory

- This is a testbed, not a production system. Session persistence across
  restarts is not a requirement.
- Zero external dependencies. No database driver, no connection pool, no
  migration tooling.
- The simplest correct implementation. Adding complexity before it is needed
  violates YAGNI.

### Why behind an interface

- If the project moves toward production deployment, switching to SQLite or
  Redis requires only a new implementation of `SessionStore` — no changes to
  the conversation manager or HTTP handlers.
- The interface is small (3 methods) and natural. It does not add meaningful
  complexity to the initial implementation.

## Consequences

- Sessions are lost on server restart. This is acceptable for a testbed.
- No horizontal scaling — sessions are pinned to the process. A load balancer
  would need sticky sessions. Not a concern for single-instance deployment.
- The `SessionStore` interface is the extension point for future persistence
  backends. Adding SQLite or Redis later is a localised change.
