# 011 — REST + SSE for the Go server API

**Status:** Accepted
**Date:** March 22, 2026

## Context

The Go server exposes an HTTP API consumed by the Next.js frontend. Three
transport options were considered:

**REST only.** Simple `POST /api/chat` returning a complete JSON response.
Well-understood, easy to implement. But users see nothing until the full LLM
response is generated, which can take several seconds.

**REST + Server-Sent Events (SSE).** Same REST endpoints, but `POST /api/chat`
streams the LLM response token-by-token via SSE. Users see the response appear
incrementally.

**gRPC.** Typed protobuf contracts with bidirectional streaming. Strong
guarantees between Go and Next.js, cancel/backpressure built in.

## Decision

REST + SSE.

### Why not gRPC

- Next.js does not speak gRPC natively. It requires a `grpc-web` proxy or a
  gRPC-to-REST gateway (Envoy, grpc-gateway), adding deployment complexity
  that is not justified for a small API surface.
- Harder to debug — no `curl`, requires `grpcurl` or specialised tools.
- The API surface is small (2–3 endpoints). The typed-contract benefit does
  not outweigh the infrastructure cost.

### Why SSE over plain REST

- LLM responses take seconds to generate. Streaming tokens as they arrive is a
  significant UX improvement — users see the response forming immediately.
- SSE is consumed trivially in the browser via native `EventSource`. No extra
  libraries needed on the Next.js side.
- Works through any reverse proxy, CDN, or load balancer without special
  configuration.
- Unidirectional (server → client) is sufficient for chat. If cancellation is
  needed later, a separate `DELETE /api/chat/:id` endpoint handles it without
  changing the transport.

## Consequences

- The JSON response shape is defined by convention, not a typed contract.
  Changes to the API require manual coordination between Go and Next.js.
- SSE requires the Go server to hold the HTTP connection open for the duration
  of the LLM response. Connection timeouts must be configured appropriately.
- If gRPC becomes necessary in the future (e.g., multiple consuming services,
  bidirectional streaming), it can be added as a second transport without
  replacing the REST+SSE API.
