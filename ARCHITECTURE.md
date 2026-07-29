# Architecture

This codebase follows **Clean Architecture**: code is organized into
concentric layers, dependencies only point inward (toward `Domain`), and
outer layers are swappable without touching business logic.

```
                     ┌─────────────────────────┐
                     │        Delivery         │  main.go, controllers, routers
                     │  (HTTP / Gin / wiring)   │
                     └───────────┬─────────────┘
                                 │ depends on
                     ┌───────────▼─────────────┐
                     │        Usecases         │  business logic
                     │ (task_usecases.go, ...)  │
                     └───────────┬─────────────┘
                                 │ depends on
                     ┌───────────▼─────────────┐
                     │          Domain         │  entities + ALL interfaces
                     │       (domain.go)        │
                     └───────────▲─────────────┘
                                 │ implements
              ┌──────────────────┼──────────────────┐
   ┌──────────▼─────────┐                 ┌─────────▼──────────┐
   │    Repositories     │                 │   Infrastructure    │
   │  (MongoDB access)   │                 │ (JWT, bcrypt, auth  │
   │                      │                 │      middleware)    │
   └──────────────────────┘                 └─────────────────────┘
```

## Layers

### `Domain/domain.go` — the center

Contains everything the rest of the app is built around, and imports
nothing application-specific:

- **Entities**: `Task`, `User`, `Claims`.
- **DTOs**: request/response shapes (`CreateTaskRequest`, `RegisterRequest`, ...).
- **Repository interfaces**: `TaskRepository`, `UserRepository` — what
  persistence must support, not how.
- **Infrastructure service interfaces**: `PasswordService`, `JWTService` —
  what hashing/token behavior must support, not which library provides it.
- **Usecase interfaces**: `TaskUsecase`, `UserUsecase` — the application's
  public business operations, as seen by the delivery layer.
- **Sentinel errors**: `ErrTaskNotFound`, `ErrInvalidCreds`, etc., shared by
  every layer so error handling doesn't depend on string-matching.

No other package is imported here except the MongoDB `primitive.ObjectID`
type, kept only because it's the ID format used in the API's JSON contract.

### `Usecases/` — business logic

`task_usecases.go` and `user_usecases.go` implement `domain.TaskUsecase` and
`domain.UserUsecase`. They hold the actual business rules:

- The **first user ever registered becomes an admin**; everyone after that
  is a regular user (`user_usecases.go`).
- Login **never reveals whether a username exists** — both "no such user"
  and "wrong password" map to the same `ErrInvalidCreds`.
- Promotion authorization (only admins may call it) is intentionally
  enforced at the delivery layer (middleware), not here — it's a request
  authorization concern, not a data invariant.

Usecases depend only on `domain.TaskRepository`, `domain.UserRepository`,
`domain.PasswordService`, and `domain.JWTService` — all interfaces. They
have never heard of MongoDB, bcrypt, or JWT libraries.

### `Repositories/` — persistence

`task_repository.go` and `user_repository.go` implement `domain.TaskRepository`
and `domain.UserRepository` against MongoDB. This is the *only* place in the
codebase that imports `go.mongodb.org/mongo-driver`. Swapping to Postgres or
an in-memory store for tests means writing a new struct that satisfies the
same two interfaces — nothing above this layer changes.

### `Infrastructure/` — technical services

- `jwt_service.go` — `JWTService` implements `domain.JWTService` using
  `github.com/golang-jwt/jwt/v5`. The library's own claims type is kept
  private to this file; `Domain` and `Usecases` only ever see the
  plain `domain.Claims` struct.
- `password_service.go` — `PasswordService` implements
  `domain.PasswordService` using bcrypt.
- `auth_middleWare.go` — `AuthMiddleware` and `AdminOnly` are Gin
  middleware. `AuthMiddleware` depends on `domain.JWTService` (the
  interface), not the concrete `JWTService`, so it can be tested with a
  fake token service if needed.

### `Delivery/` — HTTP boundary

- `controllers/controller.go` — `TaskController` and `UserController`
  translate HTTP requests into usecase calls and usecase results into JSON
  responses. They depend on `domain.TaskUsecase` / `domain.UserUsecase`
  (interfaces), **not** on the concrete structs in `Usecases/` — so a
  controller test can inject a hand-written fake usecase with zero mocking
  frameworks.
- `routers/router.go` — pure route table. Takes an already-configured Gin
  engine, controllers, and middleware, and wires paths to handlers. Knows
  nothing about Mongo, JWT, or business rules.
- `main.go` — the **composition root**. This is the one file in the entire
  application allowed to import every layer at once: it reads
  configuration, connects to MongoDB, constructs each repository, wraps them
  in use cases, wraps those in controllers, and starts the Gin server. If
  you ever need to see "how does everything fit together," start here.

## Why this shape

- **Dependency inversion**: `Usecases` defines what it needs
  (`TaskRepository`, `JWTService`, ...) as interfaces in `Domain`;
  `Repositories`/`Infrastructure` provide concrete implementations. Business
  logic depends on abstractions it owns, not on frameworks it doesn't.
- **Testability**: every usecase can be unit-tested with a tiny hand-written
  fake repository/service (see `Usecases/*_test.go`) — no real database, no
  HTTP server, no mocking library required.
- **Backward compatibility**: the HTTP contract (routes, request/response
  JSON shapes, status codes, the first-user-is-admin rule, JWT expiry, etc.)
  is unchanged from the pre-refactor version. This refactor only moves code
  and introduces interfaces — it does not change external behavior.

## What moved from the original layout

| Before                              | After                                          |
|--------------------------------------|------------------------------------------------|
| `main.go`                            | `Delivery/main.go` (now the composition root)   |
| `router/router.go`                   | `Delivery/routers/router.go` (pure routing; DB/service wiring moved to `main.go`) |
| `controllers/*.go`                   | `Delivery/controllers/controller.go` (now depends on `domain.*Usecase` interfaces instead of concrete services) |
| `middleware/auth_middleware.go`      | `Infrastructure/auth_middleWare.go` (now depends on `domain.JWTService` interface) |
| `models/*.go`                        | `Domain/domain.go` (entities + DTOs, plus new interfaces) |
| `data/task_service.go` (Mongo + business rules mixed) | Split into `Repositories/task_repository.go` (Mongo only) and `Usecases/task_usecases.go` (business rules only) |
| `data/user_service.go` (Mongo + JWT + bcrypt + business rules mixed) | Split into `Repositories/user_repository.go` (Mongo only), `Infrastructure/jwt_service.go`, `Infrastructure/password_service.go`, and `Usecases/user_usecases.go` (business rules only) |
