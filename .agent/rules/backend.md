# Backend Engineering Rules

## Core Principles

- Code must be simple, maintainable, and testable
- Prefer clarity over cleverness
- Follow SOLID, DRY, KISS, and YAGNI
- Every feature must be measurable and observable
- Security is a default requirement, not an option

---

## Architecture Standards

- Use a layered, modular architecture:
  - API / Transport Layer (HTTP/gRPC/Events)
  - Service / Application Layer (business logic)
  - Repository / Data Layer (persistence)

- Keep controllers/handlers thin: validate input, map to DTOs, call service layer, return responses.

- Business logic belongs in the Service layer; do not put domain rules in transport code.

- Data access must go through repository interfaces. Repositories are the only components that touch the database.

- Strong service boundaries:
  - No direct cross-service database access — use service APIs, gRPC, or event-driven integration.
  - Prefer explicit, versioned RPC/REST contracts or published domain events for integration.

- Transactions & consistency:
  - Keep transactional boundaries within a single service.
  - For distributed workflows use sagas/compensating actions or reliable event-driven patterns.

- Interfaces and dependency injection:
  - Depend on interfaces, not implementations; wire dependencies at construction time for testability.

- Resilience & retries:
  - Implement idempotency for retryable operations.
  - Use retries with exponential backoff, circuit breakers, and bulkheads for external calls.

- Data modeling:
  - Use internal domain models; map to transport DTOs and persistence models to avoid leaking internals.
  - Migrations should be versioned and reversible where possible.

- Observability:
  - Instrument services with tracing (OpenTelemetry), structured logging, and Prometheus-compatible metrics.
  - Include request/trace IDs in logs and propagate through calls.

- Security & configuration:
  - Enforce authN/authZ at the API boundary or middleware. Do not hardcode credentials.
  - Use environment variables or secret managers for sensitive config.

- Deployment & runtime:
  - Services must be stateless and support graceful shutdown.
  - Health checks (liveness/readiness) are required.

- Testing & contracts:
  - Provide unit tests for business logic, integration tests for persistence, and contract tests for cross-service APIs.
  - Verify protobuf/OpenAPI compatibility in CI when schemas change.

---

## API Design Rules

- APIs must be versioned
- Use REST or gRPC consistently
- Clear contracts and DTOs
- Validate all inputs
- Never expose internal models
- Standard error format
- Proper HTTP status codes

---

## Code Quality

- Unit tests required for all business logic
- Integration tests for critical flows
- Linting must pass
- No magic numbers or strings
- Consistent logging format

---

## Security

- No credentials in code
- Use environment variables or secret managers
- Validate and sanitize all inputs
- Follow least privilege principle
- Authentication and authorization mandatory

---

## Performance

- Optimize queries
- Use caching when appropriate
- Avoid premature optimization
- Measure before improving

---

## Observability

- Structured logging
- Metrics for critical paths
- Tracing for distributed systems
- Health checks required

---

## Deployment

- Service must be stateless
- Configuration via environment variables
- Graceful shutdown supported
- Backward compatibility required
