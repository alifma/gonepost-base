-include .env
export

COMPOSE_FILE := infrastructure/compose/docker-compose.yml
MIGRATIONS_DIR := apps/api/migrations

.PHONY: db-up db-down db-logs \
	dev-api build-api lint-api test-api test-integration-api \
	db-migrate db-rollback db-status db-migrate-create db-seed \
	openapi-generate api-client-generate openapi-validate

## --- Local dependencies ---

db-up:
	docker compose -f $(COMPOSE_FILE) up -d

db-down:
	docker compose -f $(COMPOSE_FILE) down

db-logs:
	docker compose -f $(COMPOSE_FILE) logs -f postgres

## --- API development ---

dev-api:
	cd apps/api && go run ./cmd/api

build-api:
	cd apps/api && go build -o bin/api ./cmd/api

lint-api:
	cd apps/api && go vet ./... && gofmt -l .

test-api:
	cd apps/api && go test ./...

test-integration-api:
	cd apps/api && go test -tags=integration ./...

## --- Database migrations (golang-migrate, against DATABASE_URL) ---

db-migrate:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

db-rollback:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

db-status:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

# usage: make db-migrate-create name=create_users_table
db-migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

db-seed:
	cd apps/api && go run ./seeds

## --- OpenAPI contract and generated clients ---

# Regenerate Go types from apps/api/openapi/openapi.yaml
openapi-generate:
	cd apps/api && go generate ./...

# Regenerate the TypeScript client in packages/api-client
api-client-generate:
	cd packages/api-client && npm run generate

# Fail if the spec changed but generated code wasn't regenerated (CI check).
# Regenerates into a temp copy and diffs against what's committed.
openapi-validate:
	@cp apps/api/internal/platform/openapigen/types.gen.go /tmp/types.gen.go.before
	@cp packages/api-client/src/types.gen.ts /tmp/types.gen.ts.before
	@$(MAKE) openapi-generate api-client-generate >/dev/null
	@diff -u /tmp/types.gen.go.before apps/api/internal/platform/openapigen/types.gen.go || (echo "Go types are out of date — run 'make openapi-generate' and commit the result." && exit 1)
	@diff -u /tmp/types.gen.ts.before packages/api-client/src/types.gen.ts || (echo "TS types are out of date — run 'make api-client-generate' and commit the result." && exit 1)
	@echo "openapi-validate: no drift"
