# FLX-158 — Webhooks on Table Events

## Architecture

Row mutations (INSERT/UPDATE/DELETE) route through new Fluxend API proxy endpoints. Frontend calls Fluxend API → Fluxend proxies to PostgREST internally (`http://postgrest_{dbName}:3000/`) → captures mutated row(s) → fires webhooks synchronously → returns PostgREST response to client. READ operations remain direct PostgREST calls.

## Payload Shape

```json
{
  "event": "insert",
  "table": "orders",
  "project": "uuid-here",
  "timestamp": "2026-06-25T10:00:00Z",
  "record": { "id": 1, "name": "Example row" }
}
```

## Backend

### Migration
`internal/database/migrations/{timestamp}_add_webhook_configs_table.sql`
- Table: `fluxend.webhook_configs` (uuid, project_uuid, table_name, url, events TEXT[], is_active, created_by, updated_by, created_at, updated_at)
- FK: project_uuid → fluxend.projects(uuid) ON DELETE CASCADE

### Domain `internal/domain/webhook/`
- `entity.go` — `WebhookConfig` struct + `Response` struct
- `repository.go` — `ListForTable`, `ListActiveForTableAndEvent`, `GetByUUID`, `ExistsByUUID`, `Create`, `Delete`
- `service.go` — `List`, `Create`, `Delete`, `Fire` (fires POST to configured URLs)
- `policy.go` — authorization against project membership

### Repository
`internal/database/repositories/webhook.go` — use `pq.Array()` for events TEXT[]

### Row Proxy Handler
`internal/api/handlers/row.go` — `Insert`, `Update`, `Delete` methods
- Extracts project UUID from `X-Project` header → fetches project → gets `dbName`
- Forwards request to `http://postgrest_{dbName}:3000/{tableName}` with `Prefer: return=representation`
- Calls `webhookService.Fire()` with returned row data
- Returns PostgREST status + body to client

Routes: `POST|PATCH|DELETE /tables/:tableName/rows`

### Webhook Config Handler
`internal/api/handlers/webhook.go` — `List`, `Store`, `Delete`

Routes: `GET|POST /tables/:tableName/webhooks`, `DELETE /tables/:tableName/webhooks/:webhookUUID`

### DTO, Mapper, Routes, DI
- `internal/api/dto/webhook_dto.go` — `StoreRequest` (url, events, is_active)
- `internal/api/mapper/webhook.go`
- `internal/api/routes/webhook.go` + `internal/api/routes/row.go`
- Update `internal/app/commands/server.go` and `internal/app/container.go`

## Frontend

### Update Row Mutations `web/app/services/tables.ts`
- `updateTableRow` / `deleteTableRows` — route through Fluxend API instead of PostgREST directly
- Add `createTableRow` — `POST /tables/{tableName}/rows`

### Webhook Config UI
- `web/app/routes/tables/webhooks.tsx` — list + add webhook form
- `web/app/services/webhooks.ts` — listWebhooks, createWebhook, deleteWebhook
- `web/app/routes/tables/sidebar.tsx` — add Webhooks nav item
- `web/app/routes.ts` — add route entry
