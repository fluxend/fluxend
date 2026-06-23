# Code Review Instructions

Standards for reviewing PRs and diffs in the Fluxend codebase. Apply these regardless of whether you are a human reviewer or an AI agent.

All coding guidelines from [BASE.md](./BASE.md) apply to code reviews. This file focuses on review-specific concerns.

## CI-Owned Checks (Out of Scope)

Do not report issues already enforced by CI:
- Formatting violations caught by `gofmt`, `prettier`, or the linter.
- Type errors caught by the TypeScript compiler.
- Test failures already visible in CI output.

Focus review feedback on security, architecture, design, and correctness beyond what CI enforces.

## Priority Order

Flag findings in this order of severity:

1. **Security** — credentials, injection vectors, auth bypasses, sensitive data exposed in logs.
2. **Correctness** — logic bugs, off-by-one errors, race conditions, unhandled error paths.
3. **Architecture** — layer violations, missing abstraction, wrong placement.
4. **Simplicity** — is there a materially shorter or clearer approach?
5. **Consistency** — does this deviate from established patterns without reason?
6. **Performance** — N+1 queries, O(n²) loops where a linear solution exists, unnecessary allocations.

## Security

- Flag hardcoded secrets, API keys, or credentials anywhere in the diff.
- Verify sensitive data (tokens, passwords, PII) is not logged.
- Check that all new endpoints have auth middleware applied in routes.

## Backend (Go) Checklist

### Architecture
- [ ] Handler does not call a repository directly — must go through service.
- [ ] Business logic lives in the service, not the handler or repository.
- [ ] Authorization is enforced in a policy, not inline in handler or service.
- [ ] New service/handler/repository is registered in `internal/app/container.go`.

### Error Handling
- [ ] Returns typed errors from `pkg/errors/` (not bare `fmt.Errorf` strings).
- [ ] Handler uses `response.ErrorResponse(c, err)` — not a custom JSON shape.
- [ ] No error is silently swallowed (always returned or logged).
- [ ] Guard clauses used — early returns instead of deeply nested conditionals.

### Database
- [ ] Named SQL parameters used (`:name`), not positional (`$1`).
- [ ] Multi-step writes wrapped in `DB.WithTransaction()`.
- [ ] New tables/columns have a migration in `internal/database/migrations/`.
- [ ] Repository interface updated alongside the concrete implementation.
- [ ] No N+1 patterns — DB calls are not inside loops over result sets.

### Response
- [ ] All success responses use `response.SuccessResponse` or `response.CreatedResponse`.
- [ ] Response envelope shape preserved: `{ success, errors, content, metadata }`.

### Validation
- [ ] Input validated in the DTO with `ozzo-validation`, not in the service.
- [ ] `request.BindAndValidate(c, &dto)` used — not manual binding + separate validate call.

### Auth
- [ ] Endpoints requiring authentication have the auth middleware applied in routes.
- [ ] JWT version check is respected — no bypasses.

### Swagger
- [ ] New endpoints have Swagger annotations (`@Summary`, `@Param`, `@Success`, `@Failure`).
- [ ] Docs regenerated if annotations changed (`make docs`).

### Code Quality
- [ ] Functions stay focused — flag anything over ~20 lines that could be split cleanly.
- [ ] No god services (too many responsibilities in one struct).
- [ ] Dead code removed — unused functions, variables, or imports flagged.
- [ ] North American spelling used throughout (`canceled`, `initialize`, `color`).
- [ ] Names are descriptive and unabbreviated.

## Frontend (React/TypeScript) Checklist

### Architecture
- [ ] No direct `fetch` or `axios` calls — all HTTP goes through `app/tools/fetch.ts`.
- [ ] Server data fetching happens in loaders, not in `useEffect`.
- [ ] Service methods are in `app/services/`, not inlined in components.

### TypeScript
- [ ] No `any` types without a comment explaining why.
- [ ] API responses typed against `APIResponse<T>`.
- [ ] No implicit `any` from missing generics.

### State
- [ ] Server state managed with TanStack Query, not duplicated in local state.
- [ ] Form state managed with React Hook Form — no manual `useState` for form fields.

### UI
- [ ] Existing Shadcn/ui components used before a custom component is built.
- [ ] `cn()` used for conditional class merging — no string concatenation.
- [ ] No inline `style` props where Tailwind can express the value.
- [ ] Dark mode verified (does not break or look wrong).
- [ ] Loading, error, and empty states handled for async data.

### Auth
- [ ] Routes that require login redirect to `/auth/login` from the loader if no token.
- [ ] Auth token retrieved via `getServerAuthToken()` in loaders, not hardcoded.

### Environment
- [ ] No hardcoded URLs — use `VITE_FLX_API_URL` / `VITE_FLX_INTERNAL_URL`.
- [ ] New env vars added to `.env.example`.

### Code Quality
- [ ] No dead code — unused components, hooks, or utilities flagged.
- [ ] North American spelling used throughout.

## Performance Red Flags

- Nested loops (O(n²)) when a linear (O(n)) solution exists.
- DB or API calls inside a loop over a result set (N+1).
- Missing `useCallback` / `useMemo` on expensive computations passed as props.

## What Not to Flag

- Formatting differences caught by CI.
- Personal preference differences where both approaches are equally valid.
- Hypothetical future requirements — review the diff as shipped.
- Pre-existing technical debt that this PR did not introduce.

## Review Style

- Be **specific and actionable** — cite the file and line number for every finding.
- Explain the **why** behind each recommendation, not just what it is.
- **Acknowledge good patterns** when you see them — not just problems.
- Ask a **clarifying question** when code intent is unclear rather than assuming it is wrong.
- Group findings by severity: **blocking** → **should-fix** → **optional**.
