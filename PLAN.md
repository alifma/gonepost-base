# Basecode Implementation Plan

## Goal

Create a reusable production baseline with:

- Next.js dashboard, forked from [Kiranism/next-shadcn-dashboard-starter](https://github.com/Kiranism/next-shadcn-dashboard-starter), as the frontend
- Go as the sole API and domain backend
- Self-hosted PostgreSQL
- Dockerized local development
- OpenAPI as the API contract
- Built-in authentication, role-based authorization, audit logging, and observability

Supabase, resource-level ACL, Redis, workers, storage, and a project generator are intentionally out of scope for the first version.

## Architecture Decisions

| Area | Decision |
| --- | --- |
| Frontend | Fork [Kiranism/next-shadcn-dashboard-starter](https://github.com/Kiranism/next-shadcn-dashboard-starter) and place it in `apps/web` |
| Backend | Go API owns domain logic, authentication, authorization, and database access |
| Database | PostgreSQL managed locally by Docker Compose |
| API contract | OpenAPI is the source of truth; generate the TypeScript API client from it |
| Authorization | Permission-based RBAC first; resource-level ACL later when a real use case needs it |
| Auth identifier | Email is required and canonical; username is optional and unique |
| Passwords | Hash with Argon2id; never store, log, or return plaintext passwords |
| Logs | Structured application logs and immutable audit logs are separate concerns |
| Branding | Versioned configuration and CSS tokens; a script generates controlled project identity files |

## Repository Structure

```text
basecode/
├── apps/
│   ├── web/
│   │   ├── src/
│   │   │   ├── app/
│   │   │   ├── components/
│   │   │   ├── config/
│   │   │   ├── features/
│   │   │   │   ├── auth/
│   │   │   │   ├── users/
│   │   │   │   ├── roles/
│   │   │   │   └── audit-logs/
│   │   │   └── lib/
│   │   └── public/
│   │       └── brand/
│   └── api/
│       ├── cmd/api/
│       ├── internal/
│       │   ├── config/
│       │   ├── middleware/
│       │   ├── platform/
│       │   │   ├── database/
│       │   │   ├── http/
│       │   │   ├── logger/
│       │   │   ├── observability/
│       │   │   └── security/
│       │   └── modules/
│       │       ├── auth/
│       │       ├── users/
│       │       ├── roles/
│       │       ├── permissions/
│       │       └── auditlog/
│       ├── migrations/
│       ├── seeds/
│       ├── openapi/
│       ├── tests/
│       │   ├── integration/
│       │   └── testutil/
│       └── scripts/
├── packages/
│   ├── api-client/
│   ├── validation/
│   └── config/
├── infrastructure/
│   ├── compose/
│   ├── docker/
│   └── postgres/init/
├── scripts/
│   ├── branding/
│   ├── db/
│   └── generate-client/
├── docs/
│   ├── architecture/
│   ├── api/
│   ├── development/
│   └── deployment/
└── .github/workflows/
```

## Frontend Starter Baseline

`apps/web` starts from [Kiranism/next-shadcn-dashboard-starter](https://github.com/Kiranism/next-shadcn-dashboard-starter). Preserve its established Next.js, shadcn/ui, Tailwind, React Query, TanStack Form, table, layout, theme, and page-container conventions.

- [ ] Record the upstream repository URL and the exact starter commit used to initialize `apps/web`.
- [ ] Keep the starter as a frontend-only application; it must communicate with Go exclusively through `packages/api-client`.
- [ ] Replace the starter's authentication/provider integration with the Go API session strategy before enabling protected dashboard routes.
- [ ] Replace starter mock-data service implementations for users, roles, and audit logs with calls to the generated API client.
- [ ] Reuse the existing feature-based frontend structure for `auth`, `users`, `roles`, and `audit-logs`.
- [ ] Keep UI permission checks for navigation, buttons, and route experience, but treat Go API authorization as the source of truth.
- [ ] Preserve and extend CSS-variable/theme-token support so the branding system can customize the dashboard safely.

The frontend must not duplicate Go domain rules, password handling, authorization decisions, or PostgreSQL access.

## Ownership Boundaries

- `apps/api` is the only application allowed to access PostgreSQL.
- `apps/api` owns auth, sessions, permissions, business rules, audit events, migrations, and seeds.
- `apps/web` owns pages, components, client state, and UI-only permission visibility.
- `apps/web` must not import database models, migrations, or Go internals.
- `packages/api-client` is generated from OpenAPI and is the typed boundary between web and API.
- Backend authorization is authoritative. Frontend permission checks are only for user experience.

## Implementation Phases

### Phase 0 — Repository Foundation

Create the workspace layout and standard developer commands.

- [ ] Create the root workspace directories.
- [ ] Fork `Kiranism/next-shadcn-dashboard-starter` and initialize it under `apps/web`.
- [ ] Record the upstream commit in the repository documentation.
- [ ] Identify and plan the replacement of starter auth/provider-specific code with Go API integration.
- [ ] Initialize the Go module under `apps/api`.
- [ ] Add root `.env.example` with documented, non-secret example values.
- [ ] Add startup environment validation for web and API.
- [ ] Add root commands for linting, formatting, testing, builds, migrations, and code generation.
- [ ] Write the local-development setup in `README.md`.

**Done when:** a new developer can clone the repository, configure environment variables, and start both applications.

### Phase 1 — Local PostgreSQL and Database Lifecycle

Make PostgreSQL repeatable for both local development and automated tests.

- [ ] Add `infrastructure/compose/docker-compose.yml` with a PostgreSQL service.
- [ ] Add persistent development volume, health check, configurable host port, database name, user, and password.
- [ ] Add database migration tooling owned by `apps/api`.
- [ ] Add migration create, migrate, rollback, and status commands.
- [ ] Add idempotent seed support.
- [ ] Add a disposable test database strategy; never use the developer database for integration tests.

**Done when:** `docker compose up -d postgres` produces a healthy database and migrations can create a clean schema repeatedly.

### Phase 2 — API Platform Baseline

Build the shared Go API infrastructure before domain modules.

- [ ] Add router, graceful shutdown, recovery, and standard JSON response helpers.
- [ ] Define one documented error response shape.
- [ ] Add request validation and a standard validation-error response.
- [ ] Add request ID middleware; preserve a valid inbound request ID or generate one.
- [ ] Add trace ID propagation for API calls and future jobs/workers.
- [ ] Add structured JSON logger with timestamp, level, service, request ID, trace ID, message, and sanitized metadata.
- [ ] Add CORS configuration through validated environment variables.
- [ ] Add security headers appropriate for an API.
- [ ] Add health and readiness endpoints.

**Done when:** every API response and application log uses the standard conventions, including errors.

### Phase 3 — OpenAPI Contract and Generated Client

Establish the frontend-to-backend contract before implementing many endpoints.

- [ ] Add an OpenAPI specification under `apps/api/openapi`.
- [ ] Document API versioning using `/api/v1`.
- [ ] Define reusable schemas for successful responses, validation errors, and problem errors.
- [ ] Generate or validate the Go server interface/types from the specification.
- [ ] Generate a TypeScript API client in `packages/api-client`.
- [ ] Configure `apps/web` to consume only the generated client through its API wrapper.
- [ ] Add Swagger UI or a documented local endpoint for inspecting the specification.
- [ ] Make OpenAPI validation and client-generation drift fail in CI.

**Done when:** changing an API endpoint requires updating the OpenAPI contract and the web client can compile against generated types.

### Phase 4 — Authentication and User Management

Implement the identity foundation.

- [ ] Create migrations for users and sessions/tokens according to the chosen session strategy.
- [ ] Require a unique, normalized email address.
- [ ] Add optional unique username and optional full name.
- [ ] Implement password hashing with Argon2id.
- [ ] Implement login, logout, current-user, password change, and password reset flows.
- [ ] Choose and document the session strategy: secure cookie session or short-lived access token plus rotating refresh token.
- [ ] Add login rate limiting and a password policy.
- [ ] Add account status checks: active, inactive, and suspended.
- [ ] Add user CRUD endpoints with pagination, filtering, and safe response DTOs.
- [ ] Add the corresponding dashboard screens.

**Done when:** a user can securely authenticate, manage their profile, and an administrator can manage users without sensitive fields appearing in API responses or logs.

### Phase 5 — RBAC and Permission Authorization

Start with permission-based RBAC. Do not implement per-resource ACL in this phase.

- [ ] Create migrations for roles, permissions, user roles, and role permissions.
- [ ] Define permission identifiers as `resource:action`, such as `users:read` and `roles:manage`.
- [ ] Add a `SUPER_ADMIN` system role.
- [ ] Add a centralized authorization service used by handlers/services.
- [ ] Add authorization middleware and policy helpers.
- [ ] Seed system roles and initial permissions idempotently.
- [ ] Add a protected bootstrap path for the first super-admin.
- [ ] Prevent deletion or demotion of the final active super-admin.
- [ ] Add role CRUD and permission-assignment APIs.
- [ ] Add UI guards for visibility only; always enforce on Go API endpoints.

**Done when:** permission-denied requests fail consistently at the API and UI controls reflect, but do not define, backend access.

### Phase 6 — Audit Logging

Track sensitive and privileged actions separately from technical logs.

- [ ] Create the audit-log migration and retention policy.
- [ ] Define the audit event schema: actor, action, resource, resource ID, result, request ID, trace ID, timestamp, and sanitized metadata.
- [ ] Audit login/logout, password changes, user CRUD, role CRUD, permission changes, and authorization-sensitive failures.
- [ ] Ensure failed audit writes do not silently disappear; choose and document whether they block the mutation or are retried reliably.
- [ ] Add read-only, permission-protected audit-log API with pagination and filters.
- [ ] Add dashboard audit-log view.

**Done when:** privileged changes can be traced to an actor and correlated with application logs without leaking secrets.

### Phase 7 — Testing Baseline

Create fast feedback plus realistic integration coverage.

- [ ] Add Go unit tests for validation, auth helpers, authorization policies, and service rules.
- [ ] Add Go integration tests against an isolated PostgreSQL database.
- [ ] Add web unit/component tests for permission-aware UI and API-state behavior.
- [ ] Add end-to-end tests for login, logout, user CRUD, role assignment, allowed access, and denied access.
- [ ] Add reusable test fixtures and factories; avoid hand-written duplicated setup in each test.
- [ ] Require tests for security-sensitive regression fixes.

**Done when:** auth and RBAC behavior is covered at unit, integration, and end-to-end levels.

### Phase 8 — Quality Gates and CI

Make repository standards automatic and non-optional.

- [ ] Configure ESLint, Prettier, and TypeScript strict mode for web.
- [ ] Configure Go formatting, vet/static analysis, and tests for API.
- [ ] Add format-check, lint, typecheck, test, integration-test, and build commands.
- [ ] Add Husky and lint-staged for fast local checks.
- [ ] Add `.github/workflows/ci.yml` for pull requests and pushes to `main`.
- [ ] Run CI in this order: format check, lint, typecheck, unit tests, integration tests with PostgreSQL, OpenAPI validation/client drift check, then builds.
- [ ] Build API and web Docker images in CI.

**Done when:** a pull request cannot pass CI with formatting, static analysis, contract, test, or production build failures.

### Phase 9 — Branding System

Make visual identity configurable without unsafe source-wide replacement.

- [ ] Define a versioned branding configuration: project name, product name, colors, logo, favicon, metadata, and URLs.
- [ ] Drive theme values through CSS variables and existing Tailwind/shadcn tokens.
- [ ] Add a branding script with validation and dry-run support.
- [ ] Generate or copy only documented assets and configuration files.
- [ ] Update app metadata, favicon, logo locations, and environment examples through the script.

**Done when:** a project identity can be changed through configuration and a controlled script without manually editing unrelated application code.

## Deferred Work

Only implement these after the baseline has been used in at least one real project.

- [ ] Resource-level ACL: explicit grants for a user/role on an individual resource.
- [ ] Redis: caching, distributed rate limiting, queues, or session support.
- [ ] Worker: asynchronous email, exports, audit delivery, and scheduled jobs.
- [ ] Storage abstraction: local/S3-compatible provider and upload policy.
- [ ] Email abstraction: transactional provider adapter.
- [ ] Supabase: evaluate separately as a PostgreSQL/auth/storage provider profile.
- [ ] Project generator: interactive module selection and branding initialization.

## Default Commands

The root workspace should expose consistent equivalents of:

```bash
# Local dependencies
docker compose -f infrastructure/compose/docker-compose.yml up -d postgres
docker compose -f infrastructure/compose/docker-compose.yml down

# Development
bun run dev:web
make dev-api

# Quality
bun run format:check
bun run lint
bun run typecheck
make lint-api
make test-api
make test-integration-api

# Contract and database
make openapi-validate
make api-client-generate
make db-migrate
make db-rollback
make db-seed

# Production verification
bun run build
make build-api
```

The exact package manager and Go tooling can change, but the command purpose and CI coverage should remain stable.

## First Milestone

The first usable release is complete only when it contains:

```text
Next.js dashboard
Go API
PostgreSQL via Docker Compose
Migrations and seed support
Validated environment configuration
Standard HTTP errors and responses
Structured logs with request/trace IDs
OpenAPI contract and generated TypeScript client
Authentication
Permission-based RBAC
Audit logging for privileged actions
Lint, format, strict type checks, tests, and GitHub Actions CI
```

Do not add Supabase, granular ACL, Redis, workers, storage, or a project generator until this milestone is stable in a real project.
