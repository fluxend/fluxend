<h1 align="center">⚡️ Fluxend</h1>

<p align="center">
  <strong>Open source, self-hosted Backend-as-a-Service built with Go.</strong><br>
  Instant REST APIs, auth, file storage, forms, and audit logs — all over your own PostgreSQL database.
</p>

<p align="center">
  <a href="https://github.com/fluxend/fluxend/actions"><img src="https://github.com/fluxend/fluxend/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/fluxend/fluxend/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0-blue" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.23-00ADD8?logo=go" alt="Go 1.23">
  <img src="https://img.shields.io/badge/postgres-17-336791?logo=postgresql" alt="PostgreSQL 17">
  <a href="https://hub.docker.com/u/fluxend"><img src="https://img.shields.io/badge/docker-ready-2496ED?logo=docker" alt="Docker"></a>
  <a href="https://github.com/fluxend/fluxend/stargazers"><img src="https://img.shields.io/github/stars/fluxend/fluxend?style=social" alt="Stars"></a>
</p>

<p align="center">
  <a href="https://console.fluxend.app/">🚀 Live Demo</a> •
  <a href="https://docs.fluxend.app/">📚 Docs</a> •
  <a href="https://docs.fluxend.app/quickstart">⚡ Quick Start</a> •
  <a href="https://docs.fluxend.app/faq">❓ FAQ</a> •
  <a href="https://github.com/fluxend/fluxend/issues">🐛 Issues</a>
</p>

<p align="center">
  <img src="web/public/demo.gif" alt="Fluxend demo">
</p>

---

## What is Fluxend?

Fluxend is a **self-hosted, open source Backend-as-a-Service (BaaS)**. It gives you the developer experience of Firebase or Supabase — instant APIs, authentication, file storage — while keeping your data on infrastructure you control.

You define your database tables through the UI or API. Fluxend generates fully functional REST endpoints for them automatically, backed by PostgreSQL and served through PostgREST. No code generation. No lock-in. No monthly seat fees.

**Built with:** Go 1.23 · Echo · PostgreSQL 17 · PostgREST · Docker · React 19 · TypeScript

---

## Why Fluxend?

| | Firebase | Supabase | Appwrite | PocketBase | **Fluxend** |
|---|:---:|:---:|:---:|:---:|:---:|
| Self-hosted | ✗ | ✓ | ✓ | ✓ | ✓ |
| Open source | ✗ | ✓ | ✓ | ✓ | ✓ |
| Built with Go | ✗ | ✗ | ✗ | ✓ | ✓ |
| PostgreSQL native | ✗ | ✓ | ✗ | ✗ | ✓ |
| Dynamic REST APIs | ✗ | ✓ | ✓ | ✓ | ✓ |
| CSV/XLSX import to API | ✗ | ✗ | ✗ | ✗ | ✓ |
| Multi-tenant orgs + RBAC | limited | ✓ | ✓ | ✗ | ✓ |
| Per-project JWT isolation | ✗ | ✓ | ✗ | ✗ | ✓ |
| S3-compatible storage | ✗ | ✓ | ✓ | ✗ | ✓ |
| Audit logs | ✗ | ✗ | ✓ | ✗ | ✓ |
| Smart forms | ✗ | ✗ | ✓ | ✗ | ✓ |
| License | proprietary | Apache 2 | BSD | MIT | GPL-3.0 |

---

## Features

### Authentication and Access Control
- **JWT authentication** — login, registration, token invalidation, session limits
- **Organizations and RBAC** — multi-tenant support with Owner, Admin, Developer, and Explorer roles
- **Per-project JWT secrets** — each project gets its own cryptographic signing secret; a token from one project is rejected by another
- **Row-level security** — access control enforced at the database level via PostgreSQL roles

### Database and REST APIs
- **Instant REST APIs** — create a table, get full CRUD endpoints immediately via PostgREST
- **Schema management** — create, rename, and delete tables and columns through the API or UI
- **Index management** — add and remove database indexes without writing SQL
- **Stored functions** — define and call PostgreSQL functions through a REST interface
- **CSV and XLSX import** — upload a spreadsheet, Fluxend creates the table and API for you
- **Table duplication** — clone an existing table structure in one call
- **OpenAPI export** — auto-generated OpenAPI documentation for every project

### Storage
- **Multi-driver file storage** — S3, Backblaze B2, Dropbox, or local filesystem
- **File containers** — organize files into named buckets
- **Download endpoints** — serve files through the API with access control
- **Database backups** — scheduled or on-demand PostgreSQL backups per project

### Forms
- **Smart forms** — create forms with typed fields and validation rules
- **Form responses** — collect and query submissions through the API
- **Field types** — text, number, boolean, and more

### Observability
- **Audit logs** — every API request to PostgREST endpoints is logged with user, method, status, and timestamp
- **Database statistics** — table sizes, row counts, and index usage per project
- **Health endpoint** — check container and database status across all projects

### Developer Experience
- **Self-contained Docker deployment** — Traefik, PostgreSQL, API, and frontend in one `docker compose up`
- **CLI commands** — restart PostgREST instances, run migrations, seed settings from the command line
- **Sentry integration** — error tracking built in, opt-in via environment variable
- **CORS configuration** — per-deployment origin allowlist

---

## Quick Start

**Requirements:** Docker, Docker Compose

```bash
git clone https://github.com/fluxend/fluxend.git
cd fluxend
cp .env.example .env   # fill in your values
docker compose up -d
```

Open `http://console.yourdomain.com` and register the first user.

Full setup guide: [docs.fluxend.app/quickstart](https://docs.fluxend.app/quickstart)

---

## How It Works

```
Your client app
      │
      ▼
┌─────────────┐     ┌─────────────────────────────────────┐
│  Fluxend    │     │  Per-project PostgREST containers   │
│  API (Go)   │────▶│  Each signed with its own JWT secret│
│             │     │  Routed by Traefik                  │
└─────────────┘     └──────────────┬──────────────────────┘
      │                            │
      ▼                            ▼
┌─────────────────────────────────────────────────────────┐
│              PostgreSQL 17                              │
│   fluxend schema (users, orgs, projects, settings)     │
│   Per-project user databases (udb*)                    │
└─────────────────────────────────────────────────────────┘
```

1. Users authenticate against the Fluxend API and receive a JWT.
2. They create projects. Each project provisions a dedicated PostgreSQL database and a PostgREST container with its own signing secret.
3. Clients call `GET /projects/:id/token` to get a project-scoped token.
4. They use that token directly against the project's PostgREST URL for data access.
5. Traefik routes traffic and handles TLS termination.

---

## API Overview

| Resource | Endpoints |
|---|---|
| Users | register, login, logout, profile |
| Organizations | CRUD, member management |
| Projects | CRUD, token, OpenAPI, logs, stats |
| Tables | CRUD, rename, duplicate, upload |
| Columns | CRUD, rename |
| Indexes | CRUD |
| Functions | CRUD, invoke |
| Storage containers | CRUD |
| Files | upload, download, rename, delete |
| Forms | CRUD |
| Form fields | CRUD |
| Form responses | CRUD |
| Backups | create, list, delete |
| Settings | list, update, reset |
| Health | pulse |

Full API reference: [docs.fluxend.app](https://docs.fluxend.app)

---

## Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_HOST` | PostgreSQL host |
| `DATABASE_USER` | PostgreSQL user |
| `DATABASE_PASSWORD` | PostgreSQL password |
| `DATABASE_NAME` | PostgreSQL database name |
| `JWT_SECRET` | Secret for signing console/auth tokens (min 32 chars) |
| `BASE_DOMAIN` | Root domain for Traefik routing |
| `URL_SCHEME` | `http` or `https` |
| `POSTGREST_DB_USER` | PostgreSQL user for PostgREST connections |
| `POSTGREST_DB_PASSWORD` | PostgreSQL password for PostgREST connections |
| `POSTGREST_DB_HOST` | PostgreSQL host for PostgREST containers |
| `POSTGREST_DEFAULT_SCHEMA` | Default schema exposed by PostgREST |
| `POSTGREST_DEFAULT_ROLE` | Anonymous role for PostgREST |
| `STORAGE_DRIVER` | `local`, `s3`, `backblaze`, or `dropbox` |
| `CUSTOM_ORIGINS` | Comma-separated CORS allowed origins |
| `SENTRY_DSN` | Optional Sentry DSN for error tracking |

---

## Use Cases

**SaaS backends** — multi-tenant architecture with per-org projects and RBAC out of the box.

**Internal tools** — spin up a data API from an existing PostgreSQL schema in minutes, no backend code required.

**Rapid prototyping** — go from idea to working API in under five minutes. Replace Fluxend with a custom service later if you need to; your data stays in Postgres.

**Data collection** — use the forms API to collect submissions from external sources without managing a backend.

**Firebase or Supabase migration** — move off a vendor-hosted BaaS without rewriting your frontend. Fluxend speaks the same REST patterns.

---

## Tech Stack

**Backend:** Go 1.23 · Echo v4 · sqlx · PostgreSQL 17 · PostgREST · samber/do (DI) · golang-jwt · AWS SDK v2 · zerolog · Sentry

**Frontend:** React 19 · React Router 7 · TanStack Query · Tailwind CSS 4 · shadcn/ui · TypeScript

**Infrastructure:** Docker · Docker Compose · Traefik v2 · goose (migrations)

---

## Contributing

The codebase is organized as a standard Go project. Backend lives in `internal/`, frontend in `web/`.

```bash
# Run the test suite
go test ./...

# Run integration tests (requires local PostgreSQL)
go test ./tests/integration/...

# Build
go build ./cmd/...
```

See [AGENTS.md](./AGENTS.md) for architecture notes and coding standards.

- Check [open issues](https://github.com/fluxend/fluxend/issues) for things to work on
- Open a PR — reviews are fast

---

## License

GPL-3.0. See [LICENSE](./LICENSE).
