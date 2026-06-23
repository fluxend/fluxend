# Base Coding Instructions

General coding standards and interaction guidelines that apply across the entire Fluxend codebase.

## Communication Style

- Use a **direct, efficient tone** — no pleasantries like "Great question!" or "Perfect!".
- Be **concise and technical** — get straight to the point.
- Focus on **facts and code** — skip conversational filler.
- When referencing code, include `file_path:line_number` so the reader can jump there.
- No emojis unless explicitly asked.
- Markdown is fine for structured output (tables, code blocks).

## Response Format

- State results and decisions directly — no internal deliberation narrated aloud.
- One sentence per update is usually enough.
- End each turn with one or two sentences: what changed and what is next. Nothing else.
- No trailing summaries of what you just did — the diff speaks for itself.
- Never open with "I will now..." or "Let me...".

When blocked or uncertain:
- State the blocker in one sentence.
- Propose one concrete option to unblock.

## Branch and Commit Conventions

- Create a new branch from `origin/main` before starting work, unless instructed otherwise.
- Branch names follow the pattern: `<ticket-id>-<short-description>` or `no-ticket-<short-description>`.
  - Examples: `FLX-123-add-storage-driver`, `no-ticket-add-agentic-instruction-files`
- Use **Conventional Commits**. Include the ticket number when available, followed by a short description.

```text
feat: FLX-1234 add presigned URL download for storage files

- Introduce GetPresignedURL method on StorageService
- Add download endpoint under /storage/:id/download
- Map driver-specific errors to typed pkg/errors responses
```

- Commit body: use concise bullet points, no blank lines between bullets.
- Add a body only when extra context is useful — single-line commits are fine for small changes.
- One commit per logical change.
- Never skip hooks (`--no-verify`) without explicit user permission.
- Never force-push `main`.

**Conventional Commit types used in this project:**

| Type | When to use |
|------|------------|
| `feat` | New feature or endpoint |
| `fix` | Bug fix |
| `refactor` | Code change with no behaviour change |
| `test` | Adding or fixing tests |
| `chore` | Tooling, deps, config (no production code) |
| `docs` | Documentation only |

## General Coding Standards

### Behaviour

- Prefer editing existing files over creating new ones.
- Do not add features, abstractions, or refactors beyond what the task requires.
- Do not add error handling for scenarios that cannot happen.
- Do not add comments explaining what code does — only add a comment when the **why** is non-obvious.
- Never write multi-line comment blocks or docstrings.
- Trust framework guarantees and internal code. Only validate at system boundaries (user input, external APIs).
- No feature flags or backwards-compatibility shims — just change the code.
- No half-finished implementations.

### Design Principles

- **Composition over inheritance.**
- **Guard clauses over nested conditionals** — use early returns to reduce nesting depth.
- **"Tell, Don't Ask"** — objects/services should have methods that encapsulate behaviour, not just expose data for callers to manipulate.
- Keep functions and methods **focused and single-purpose** — if a function exceeds ~20 lines, consider splitting it.
- No god classes (services with too many responsibilities). Split when a service grows beyond its domain.
- **Readability over micro-optimisations** — but be mindful of algorithmic complexity. Avoid obvious inefficiencies like N+1 database queries or O(n²) nested loops when a linear solution exists.

### Naming

- Use **descriptive, unabbreviated names** — no `mgr`, `svc`, `util` suffixes unless they are the established codebase pattern.
- **North American spelling** throughout — applies to code, strings, comments, and documentation.
  - `canceled` not `cancelled`, `initialize` not `initialise`, `color` not `colour`.

### Security

- Never introduce command injection, XSS, SQL injection, or other OWASP top-10 vulnerabilities.
- Validate user-facing input at the handler/DTO layer, not deeper.
- Never log secrets, tokens, or passwords.
- Never hardcode credentials or API keys — use environment variables.

### What NOT to Write

- No hardcoded credentials or secrets.
- No global state or package-level mutable variables.
- No abbreviations in names.
- No business logic mixed with infrastructure concerns.

## When Explaining Code

- Reference domain concepts, not implementation details.
- Explain **architectural decisions** — why something lives where it does.
- Highlight security implications when relevant.
- Suggest refactoring opportunities when code violates the principles above, but only if asked or if the violation is severe.

## Project Overview

Fluxend is a self-hosted BaaS (Backend-as-a-Service) written in Go with a React frontend. It provides dynamic REST APIs over PostgreSQL, file storage, auth, forms, audit logs, and more — all self-contained with no external lock-in.

**Backend:** Go 1.23 · Echo v4 · sqlx · PostgreSQL · samber/do DI · JWT auth  
**Frontend:** React 19 · React Router 7 · TanStack Query · Tailwind CSS 4 · Shadcn/ui · TypeScript

See [AGENTS-BACKEND.md](./AGENTS-BACKEND.md) and [AGENTS-FRONTEND.md](./AGENTS-FRONTEND.md) for architecture detail.
