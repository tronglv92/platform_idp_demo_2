# Architecture Overview

## 1. Purpose

This document describes the **overall architecture, layering, and dependency wiring strategy** of the system. It serves as a reference for:

* Developers joining the project
* Code generation tools (pmctl, LLMs)
* Architecture & design reviews

The system follows **Clean Architecture + Dependency Composition** principles with a strong focus on **runtime-specific contexts** and **explicit dependency wiring**.

---

## 2. High-level Principles

### 2.1 Dependency Rules

* Business logic **must not** depend on infrastructure
* Infrastructure dependencies are **wired at runtime**
* Contexts expose **interfaces only**, implementations remain private

```
Domain / Service Logic
        ↑
   Context Interfaces
        ↑
   Runtime Contexts
```

---

### 2.2 Composition over Inheritance

* Dependencies are composed via contexts
* Contexts are small, focused, and runtime-specific
* Interfaces are composed to form higher-level contracts

---

## 3. Package Structure

```
registry/
├── base_context.go
├── repository_context.go
│
├── service_context.go
├── cron_context.go
├── consumer_context.go
│
├── security_authentication.go
├── security_authorization.go
├── security_http.go
└── security_context.go
```

---

## 4. Registry Package

### 4.1 Purpose of `registry`

The `registry` package is the **composition root** of the application. It is responsible for:

* Wiring dependencies
* Managing lifecycle of shared resources
* Providing runtime-specific contexts

⚠️ **No business logic is allowed in this package.**

---

## 5. Base Context

### File: `base_context.go`

**Responsibility:**

* Application-wide infrastructure
* Configuration access
* Logger / tracer (if applicable)

**Example responsibilities:**

* Config loading
* Environment metadata
* Shared utilities

---

## 6. Repository Context

### File: `repository_context.go`

**Responsibility:**

* Database connections
* Repository initialization

**Rules:**

* Only data-access related dependencies
* No service, security, or transport logic

**Example:**

```go
type RepositoryContext interface {
GetProductRepo() ProductRepository
}
```

---

## 7. Service Context (HTTP / gRPC)

### File: `service_context.go`

**Responsibility:**

* Wiring dependencies for synchronous services
* Used by HTTP / gRPC servers

**Typical dependencies:**

* RepositoryContext
* SecurityContext
* Cache
* ServiceFactory

**Example:**

```go
type ServiceContext interface {
RepositoryContext
SecurityContext
}
```

---

## 8. Cron Context

### File: `cron_context.go`

**Responsibility:**

* Wiring dependencies for scheduled jobs

**Characteristics:**

* No HTTP middleware
* No request-scoped auth
* Uses system/service identity

---

## 9. Consumer Context

### File: `consumer_context.go`

**Responsibility:**

* Wiring dependencies for async consumers

**Characteristics:**

* Message-based execution
* Explicit ack / retry policies
* No HTTP transport

---

## 10. Security Architecture

Security is split into **capability-based contexts** and composed at a higher level.

---

### 10.1 Authentication Context

**File:** `security_authentication.go`

**Responsibility:**

* Identity verification
* Token lifecycle management

```go
type AuthenticationContext interface {
GetAuthenticator() Authenticator
GetTokenManager() TokenManager
}
```

---

### 10.2 Authorization Context

**File:** `security_authorization.go`

**Responsibility:**

* Permission evaluation
* Role / policy resolution

```go
type AuthorizationContext interface {
GetAuthorizer() Authorizer
GetPermissionProvider() PermissionProvider
}
```

---

### 10.3 HTTP Security Context

**File:** `security_http.go`

**Responsibility:**

* HTTP-specific security middleware

```go
type HttpSecurityContext interface {
GetAuthMiddleware() Middleware
GetAuthzMiddleware() Middleware
}
```

---

### 10.4 Security Context (Composition)

**File:** `security_context.go`

**Responsibility:**

* Composes all security capabilities
* Exposed to service contexts

```go
type SecurityContext interface {
AuthenticationContext
AuthorizationContext
HttpSecurityContext
}
```

---

## 11. Context Construction Strategy

Each context:

* Has its own `NewXContext(...)` constructor
* Accepts only required dependencies
* Lazily initializes heavy components when needed

**Example:**

```go
func NewServiceContext(cfg Config) ServiceContext {
return &serviceContext{...}
}
```

---

## 12. Code Generation Compatibility

The architecture is designed to be **codegen-friendly**:

* Clear naming conventions
* Predictable constructor patterns
* One responsibility per file

This allows tools like **pmctl** or LLM-based generators to:

* Scan contexts
* Generate factories
* Safely extend the system

---

## 13. Design Rules (Strict)

✅ Allowed:

* Interface composition
* Lazy initialization
* Runtime-specific contexts

❌ Not allowed:

* Business logic in registry
* Cross-context coupling
* Hidden dependencies

---

## 14. Summary

This architecture provides:

* Clear separation of concerns
* Scalable dependency wiring
* Strong runtime isolation
* Excellent support for automation and code generation

It is suitable for **microservices**, **platform teams**, and **long-lived systems**.

---

*Last updated: 2025*
