# Gonepost

**Go** + **Next.js** + **Postgres** — production baseline dengan auth, RBAC, dan audit-log built in. Frontend (`apps/web`) belum digarap.

## Prerequisites

| Tool | Versi minimal | Cek |
| --- | --- | --- |
| [Go](https://go.dev/dl/) | 1.24+ (pakai `go tool` support) | `go version` |
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | apa aja yang jalan | `docker --version` |
| [Node.js](https://nodejs.org/) | 18+ | `node --version` |
| `make` | GNU Make | `make --version` |
| [golang-migrate CLI](https://github.com/golang-migrate/migrate) | v4 | `migrate -version` |

**Windows:** `make` biasanya belum ada. Install lewat winget:
```powershell
winget install ezwinports.make
```
Restart terminal abis install biar PATH ke-refresh.

**`migrate` CLI** (semua OS):
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
Pastiin `$(go env GOPATH)/bin` ada di `PATH`.

`oapi-codegen` **gak perlu install manual** — udah didaftarin sebagai Go tool dependency (`go.mod`), otomatis kepanggil lewat `go tool oapi-codegen` / `make openapi-generate`.

## Setup (sekali doang)

```bash
git clone <repo-url>
cd gonepost
```

```bash
cp .env.example .env
```
Default-nya udah pas buat local dev (dummy creds), biasanya gak perlu diubah.

```bash
# 1. Nyalain PostgreSQL (dev + test db, port 5432 & 5433)
make db-up

# 2. Apply schema
make db-migrate

# 3. Seed data awal (permission, role SUPER_ADMIN, user admin@example.com/changeme123)
make db-seed

# 4. Install dependency Go module
cd apps/api && go build ./... && cd ../..

# 5. (opsional) Install dependency TypeScript client
cd packages/api-client && npm install && cd ../..
```

## Jalanin API

```bash
make dev-api
```
atau F5 di VS Code (`.vscode/launch.json` udah disiapin, otomatis bawa `DATABASE_URL`/`APP_ENV`/`PORT`).

Cek jalan:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Coba API

Import `docs/api/bruno/` ke [Bruno](https://www.usebruno.com/), pilih environment **Local**, login pake `admin@example.com` / `changeme123`. Detail: `docs/api/bruno/README.md`.

## Commands

```bash
# Dependencies
make db-up / make db-down / make db-logs

# Development
make dev-api
make dev-web

# Quality
make lint-api
make test-api               # unit test, gak butuh DB
make test-integration-api   # + integration test (butuh db-up)

# Database
make db-migrate / make db-rollback / make db-status
make db-migrate-create name=xxx  # bikin migration baru
make db-seed

# OpenAPI contract
make openapi-generate       # regenerate Go types dari openapi.yaml
make api-client-generate    # regenerate TypeScript client
make openapi-validate       # gagal kalau spec berubah tapi belum di-regenerate
```

## Struktur

```text
apps/
├── web/          # Next.js dashboard (belum digarap)
└── api/          # Go API — satu-satunya yang boleh akses PostgreSQL
    ├── cmd/api/           # entry point
    ├── internal/
    │   ├── platform/      # infra generik (db, http, logger, security, dst)
    │   └── modules/        # domain: auth, users, roles, permissions, auditlog
    ├── migrations/
    ├── openapi/           # kontrak API — source of truth
    └── seeds/
packages/
└── api-client/   # TypeScript client, generated dari openapi.yaml
infrastructure/
└── compose/      # docker-compose PostgreSQL
docs/
└── api/bruno/    # Bruno collection buat coba API
```

`apps/web` cuma boleh manggil API lewat `packages/api-client` — gak ada akses database langsung dari frontend. Otorisasi backend selalu jadi sumber kebenaran; permission check di frontend (kalau ada nanti) cuma buat UX.
