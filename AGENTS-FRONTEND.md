# Frontend Agent Instructions

React/TypeScript frontend for Fluxend. Read this before touching any code under `web/`.

## Directory Map

```
web/
├── package.json                 # Node 20.16+, npm
├── vite.config.ts               # Vite 5 build config
├── tsconfig.json                # TypeScript 5.7
└── app/
    ├── root.tsx                 # Root layout, global providers, error boundary
    ├── routes.ts                # Route manifest (file-based routing declaration)
    ├── globals.css              # Tailwind 4 globals + CSS custom properties
    ├── routes/                  # Page components, organised by feature
    │   ├── auth/                # login, signup, logout
    │   ├── dashboard/
    │   ├── projects/
    │   ├── tables/
    │   ├── storage/
    │   ├── forms/
    │   ├── functions/
    │   ├── logs/
    │   ├── backups/
    │   ├── docs/
    │   └── settings/
    ├── components/
    │   ├── ui/                  # Shadcn/ui + Radix primitives
    │   ├── shared/              # App-level: Layout, Logo, navigation
    │   ├── auth/                # Auth-specific components
    │   ├── tables/              # Table-related components
    │   └── storage/             # Storage-related components
    ├── services/                # API service layer (one file per resource)
    ├── hooks/                   # Custom React hooks
    ├── contexts/                # React Context (theme)
    ├── lib/                     # Utilities: auth, cookies, query, router, utils
    ├── tools/
    │   └── fetch.ts             # Universal fetch (works server + client)
    └── types/                   # Shared TypeScript types
```

## Stack

| Layer | Library | Version |
|-------|---------|---------|
| Framework | React | 19 |
| Routing | React Router | 7 (SSR-capable) |
| Server state | TanStack Query | 5 |
| Tables | TanStack React Table | 8 |
| Virtualisation | TanStack React Virtual | 3 |
| Forms | React Hook Form + Zod | 7 / 3 |
| Styling | Tailwind CSS | 4 |
| UI components | Shadcn/ui + Radix UI | — |
| Icons | Lucide React | — |
| Toasts | Sonner | — |
| Animation | Motion | — |
| HTTP | Universal fetch (`tools/fetch.ts`) | — |
| TypeScript | — | 5.7 |

## Routing

React Router v7 with **file-based routing** declared in `app/routes.ts`.

- **Loaders** run on the server before rendering — use them to fetch required data.
- **Actions** handle form submissions (POST / PUT / PATCH / DELETE).
- Nested routes use layout components (e.g., a project layout wraps tables, forms, storage).
- Always put the route component file in the matching `routes/<feature>/` directory.

## API Communication

### Fetch Utility

All HTTP calls go through `app/tools/fetch.ts`. It detects server vs. client context and uses the correct base URL:
- Client-side: `VITE_FLX_API_URL`
- Server-side (loader/action): `VITE_FLX_INTERNAL_URL`

Never use `fetch` or `axios` directly — always go through this utility.

### Service Layer

Services live in `app/services/`. Each is a factory function that takes an auth token and returns methods:

```ts
// Example shape
export const createUserService = (authToken: string) => ({
  getProfile: () => fetchUtil('/users/me', { token: authToken }),
  updateProfile: (data: UpdateUserDto) => fetchUtil('/users/me', { method: 'PATCH', body: data, token: authToken }),
})
```

- No auth token needed for `auth.ts` (login/signup).
- Add project-scoped calls with the `X-Project` header.

### Response Shape

All API responses follow the backend envelope:

```ts
interface APIResponse<T> {
  success: boolean
  errors: string[]
  content: T
  metadata: unknown
}
```

This type lives in `app/lib/types.ts`. Always type API calls against it.

## Authentication

- Session token stored in an HTTP-only cookie (`session_token`).
- Organisation UUID stored in a separate cookie (`organization_uuid`).
- Server-side: retrieve token via `getServerAuthToken()` in `app/lib/auth.ts`.
- Client-side: cookie is included automatically by the browser.
- Redirect to `/auth/login` from loaders if no token is present.

## State Management

| State type | Tool |
|------------|------|
| Server / async data | TanStack Query |
| Global UI (theme) | React Context (`contexts/theme-context.tsx`) |
| Form state | React Hook Form |
| Lightweight client state | Zustand (where needed) |

Use TanStack Query for all server state. Do not duplicate server data in local state.

## Styling

- Tailwind CSS 4 with the `@tailwindcss/vite` plugin.
- Global CSS custom properties in `globals.css` (light/dark mode via `color-scheme`).
- Primary colour: `hsl(47.9 95.8% 53.1%)` (amber/orange-gold).
- Dark mode is supported — always check both modes when adding new UI.
- Class utilities: `cn()` from `app/lib/utils.ts` (wraps `clsx` + `tailwind-merge`).
- Component variants: `class-variance-authority` (cva).
- Never write inline `style` props unless Tailwind cannot express the value.

## UI Components

- Prefer **Shadcn/ui** components from `components/ui/` before building from scratch.
- These wrap **Radix UI** primitives — they are accessible and composable.
- Icons come from **Lucide React** (`import { SomeIcon } from 'lucide-react'`).
- Toast notifications via **Sonner** (`import { toast } from 'sonner'`).

## Tables

- Headless logic via **TanStack React Table**.
- Large datasets use **TanStack React Virtual** for row virtualisation.
- Table components live in `components/tables/`.
- Do not recreate table logic inline in route components.

## Forms

- **React Hook Form** for form state.
- **Zod** schemas for validation.
- Connect them with `@hookform/resolvers/zod`.
- Use Shadcn `<Form>` components for consistent styling.

## TypeScript Conventions

- Strict mode is on — no `any` unless truly unavoidable (document why with a comment).
- Define shared types in `app/types/`.
- API-specific types should mirror the backend DTOs.
- Prefer `interface` for object shapes, `type` for unions/aliases.

## Environment Variables

| Variable | Used by |
|----------|---------|
| `VITE_FLX_API_URL` | Client-side fetch (browser) |
| `VITE_FLX_INTERNAL_URL` | Server-side fetch (loader/action) |

Never hardcode URLs. Add new env vars to `.env.example` when introducing them.

## Adding a New Feature (Checklist)

1. Add route entry in `app/routes.ts`.
2. Create route component in `app/routes/<feature>/`.
3. Add loader (server-side data fetch) and/or action (mutations) in the same file.
4. Add service methods in `app/services/<feature>.ts`.
5. Use TanStack Query for client-side data interactions.
6. Build UI using existing Shadcn/ui components before creating custom ones.
7. Type all API calls with `APIResponse<T>`.
8. Handle loading, error, and empty states.
9. Test dark mode and mobile breakpoints.
