# Backend Agent Instructions

Go backend for Fluxend. Read this before touching any code under `internal/`, `pkg/`, or `cmd/`.

## Directory Map

```
/
├── cmd/main.go                  # Entry point — loads .env, runs cobra CLI
├── internal/
│   ├── app/
│   │   ├── container.go         # DI wiring — all services/repos/handlers registered here
│   │   └── commands/            # cobra commands: server, seed, optimize, routes
│   ├── api/
│   │   ├── handlers/            # HTTP handlers, one file per resource
│   │   ├── routes/              # Route group registration
│   │   ├── dto/                 # Request/response DTOs (validation lives here)
│   │   ├── response/            # Standardised response helpers + error mapping
│   │   ├── middlewares/         # Auth, request logging, CORS, Sentry
│   │   └── mapper/              # DTO → domain entity conversion
│   ├── domain/                  # Business logic, one sub-package per feature
│   │   └── <feature>/
│   │       ├── entity.go        # Domain struct
│   │       ├── service.go       # Business logic (calls policy + repo)
│   │       ├── repository.go    # Interface definition
│   │       ├── policy.go        # Authorization rules
│   │       └── types.go         # Feature-specific types/enums
│   ├── database/
│   │   ├── db.go                # PostgreSQL connection (sqlx)
│   │   ├── repositories/        # Concrete repository implementations
│   │   ├── migrations/          # SQL migrations
│   │   ├── seeders/             # Seed data
│   │   └── factories/           # Test data factories
│   └── config/
│       └── constants/           # App-wide constants (status codes, limits, column types)
├── pkg/                         # Public reusable packages (no domain logic)
│   ├── auth/                    # Password hashing, JWT utilities
│   ├── errors/                  # Custom error types
│   └── *.go                     # Utilities: request, type_conversion, faker, dynamic, general
└── tests/                       # Integration/e2e tests
```

## Architecture: Request Lifecycle

```
HTTP Request
  → Middleware (CORS, recovery, Sentry, auth, request logger)
  → Handler      — bind DTO, validate input, call service
  → Service      — business logic, calls policy then repository
  → Policy       — authorization: returns ForbiddenError if denied
  → Repository   — database queries via sqlx
  → PostgreSQL
```

Each layer has a single responsibility. Do not skip layers (e.g., handlers must not call repositories directly).

## Dependency Injection

Uses `samber/do`. All constructors follow the pattern:

```go
func NewUserService(injector *do.Injector) (UserService, error) {
    repo := do.MustInvoke[UserRepository](injector)
    // ...
    return &userService{repo: repo}, nil
}
```

All registrations happen in `internal/app/container.go`. When adding a new service or handler, register it there.

## Response Format

All API responses use a single envelope:

```json
{
  "success": true,
  "errors": [],
  "content": { ... },
  "metadata": { ... }
}
```

Use helpers from `internal/api/response/`:
- `response.SuccessResponse(c, content)` — 200
- `response.CreatedResponse(c, content)` — 201
- `response.ErrorResponse(c, err)` — maps custom error types to HTTP status codes

Never return raw errors or custom JSON structures from handlers.

## Error Handling

Custom error types live in `pkg/errors/`. The error type determines the HTTP status code:

| Error Type | Status |
|------------|--------|
| `BadRequestError` | 400 |
| `UnauthorizedError` | 401 |
| `ForbiddenError` | 403 |
| `NotFoundError` | 404 |
| `UnprocessableError` | 422 |
| Anything else | 500 |

Always return one of these typed errors from services and policies. `response.ErrorResponse()` handles the mapping.

## Authentication & Authorization

- JWT bearer tokens in `Authorization` header.
- JWT claims contain: user UUID, role ID, JWT version.
- Max 5 concurrent sessions per user (enforced via JWT version).
- Auth middleware extracts and validates the token, sets user on context.
- Authorization is enforced in **Policy** structs — never inline in handlers or services.

## DTOs and Validation

- Request DTOs live in `internal/api/dto/` and use `ozzo-validation`.
- Bind and validate in one step: `request.BindAndValidate(c, &dto)`.
- Mapper functions in `internal/api/mapper/` convert DTOs → domain entities.
- Do not put business logic in DTOs — validation only.

## Database Conventions

- PostgreSQL with `jmoiron/sqlx` for named query support.
- Schema-based table naming: `authentication.users`, `storage.files`, etc.
- Repositories define an interface in `internal/domain/<feature>/repository.go`.
- Concrete implementations live in `internal/database/repositories/`.
- Pagination params: `Page`, `Limit`, `Sort`, `Order`.
- Use `DB.WithTransaction(fn)` for multi-step writes.
- Named parameters in SQL (`:name` style), not positional (`$1`).

## API Conventions

- RESTful endpoints, resource-plural names (`/users`, `/projects`).
- `X-Project` header carries the project UUID for project-scoped resources.
- Query parameters for filtering and pagination.
- Swagger annotations on all handlers (`@Summary`, `@Param`, `@Success`, `@Failure`).
- Run `make docs` to regenerate Swagger spec after changing annotations.

## Testing

- Integration tests in `tests/` hit a real database — do not mock the DB.
- Unit tests alongside source files.
- Test factories in `internal/database/factories/` create realistic test data.
- Run tests: `make test`.

## Key Constants

- `internal/config/constants/` — column types, user statuses, storage drivers, limits.
- Check here before hardcoding magic values anywhere.

## Adding a New Feature (Checklist)

1. Define entity in `internal/domain/<feature>/entity.go`.
2. Define repository interface in `internal/domain/<feature>/repository.go`.
3. Implement repository in `internal/database/repositories/<feature>.go`.
4. Write service in `internal/domain/<feature>/service.go`.
5. Write policy in `internal/domain/<feature>/policy.go`.
6. Write DTO(s) in `internal/api/dto/`.
7. Write mapper in `internal/api/mapper/`.
8. Write handler in `internal/api/handlers/<feature>.go`.
9. Register routes in `internal/api/routes/`.
10. Register all new types in `internal/app/container.go`.
11. Write migration in `internal/database/migrations/`.
