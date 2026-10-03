# Adding a feature or a route

`items` is the reference feature. It has everything a real feature needs: table, repository, handler with validation, per-user data, audit log, RBAC permissions, OpenAPI contract, tests, and a web page. **Copy it, do not start from nothing.**

| Layer | Files of the `items` feature |
| --- | --- |
| Database | `apps/api/migrations/000005_create_items_table.{up,down}.sql` |
| Permissions | `apps/api/internal/modules/permissions/permissions.go` (`ItemsRead`, `ItemsWrite`) |
| Audit | `apps/api/internal/modules/auditlog/model.go` (`ActionItem*`) |
| Backend module | `apps/api/internal/modules/items/` (`model`, `repository`, `handler`, `routes` + tests) |
| Contract | `apps/api/openapi/openapi.yaml` (paths `/api/v1/items…`, schemas `Item*`) |
| Wiring | `apps/api/cmd/api/main.go`, `apps/api/seeds/main.go` |
| Web | `apps/web/src/features/items/`, `apps/web/src/app/dashboard/items/page.tsx`, `apps/web/src/config/nav-config.ts` |

The rules that make it work:

- **The contract comes first.** Edit `openapi.yaml`, then generate. Go handlers and the web both use the generated types.
- **Every permission is a pair** `<feature>:read` / `<feature>:write`. GET routes need read, everything else needs write.
- **Every query filters by the caller** (`owner_id`). Another user's row looks like "not found".
- **The API enforces access; the web only hides things.** The nav config is the single place the web decides who sees a page.

## A. New feature (full steps)

Example: a feature called `notes`. Replace `note`/`notes`/`Notes` everywhere.

### 1. Permissions

`apps/api/internal/modules/permissions/permissions.go`: add the constants and add both to `All`.

```go
NotesRead  = "notes:read"
NotesWrite = "notes:write"
```

`apps/api/seeds/main.go`: add both to `memberPermissions` so the MEMBER role can use the feature. `SUPER_ADMIN` gets everything from `All` automatically.

### 2. Migration

```bash
make db-migrate-create name=create_notes_table
```

Fill the `.up.sql` and `.down.sql` (copy `000005_create_items_table.*`). Keep `owner_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE` and an index starting with `owner_id`. Never edit a migration that is already applied: add a new one.

```bash
make db-migrate
```

### 3. OpenAPI contract

In `apps/api/openapi/openapi.yaml` copy the `/api/v1/items` and `/api/v1/items/{id}` paths and the `Item`, `ItemInput`, `ItemListResponse`, `ItemStatus` schemas. Rename, then change the fields. Use a new `operationId` per operation (`listNotes`, `createNote`, …).

```bash
make openapi-generate        # Go types -> internal/platform/openapigen
make api-client-generate     # TypeScript types + build packages/api-client/dist
```

### 4. Backend module

```bash
cp -r apps/api/internal/modules/items apps/api/internal/modules/notes
```

In the new folder: rename the package and types, then change the fields in `model.go` (struct, `Input`, `Public()`), the SQL in `repository.go` (`itemCols`, queries, `scanItem`), and `itemRequest` plus `parse` in `handler.go` (validate tags). Update the tests too. Add audit actions (`ActionNoteCreate/Update/Delete`) to `auditlog/model.go` and use them in the handler.

### 5. Wire it in `cmd/api/main.go`

Three small pieces, next to the items ones:

```go
requireNotesRead := roles.RequirePermission(authz, auditService, permissions.NotesRead)
requireNotesWrite := roles.RequirePermission(authz, auditService, permissions.NotesWrite)

notesHandler := notes.NewHandler(notes.NewRepository(pool), auditService)

notesHandler.Register(mux, requireAuth, requireNotesRead, requireNotesWrite)
```

### 6. Check the backend

```bash
make test-api
make db-migrate-test          # once per new migration: migrate the test DB (port 5433)
make test-integration-api
make db-seed                  # creates the new permissions and grants them
make dev-api                  # then try it with Bruno or curl
```

### 7. Web page

1. `cp -r apps/web/src/features/items apps/web/src/features/notes`. Rename, then edit `api.ts` (types come from `@gonepost/api-client`, endpoints must match the contract) and `components/notes-page.tsx` (columns and form fields).
2. Create `apps/web/src/app/dashboard/notes/page.tsx` (copy `items/page.tsx`).
3. Add the nav entry in `apps/web/src/config/nav-config.ts` with `access: { permission: 'notes:read' }`. This single entry hides the menu item, hides it in Cmd+K, and makes `RouteGuard` block the page. Do not check permissions inside the page. Only hide write buttons with `useCan().can('notes:write')`.
4. Add the two codes to the `permissions` array in `apps/web/src/features/roles/components/api-roles-panel.tsx` so admins can grant them from the Roles page.
5. Check: `cd apps/web && npx tsc --noEmit && npx oxlint src`.

### 8. Give someone access

Sign in as `admin@example.com`, open **Users**, assign the **MEMBER** role (or any role holding `notes:*`) to the user, and have them reload. A user without the role gets 403 from the API and the "No access" screen in the web.

## B. New route in an existing module

1. Handler method in `handler.go` (decode, validate, call the repository, audit if it changes data, respond).
2. Repository method if it needs SQL. Always filter by owner or scope.
3. Route line in the module's `routes.go`. Pick `read(...)` or `write(...)`.
4. Path and schema in `openapi.yaml`, then `make openapi-generate api-client-generate`.
5. Test: handler test with the fake store; repository test (`//go:build integration`) when SQL is new.
6. Web: add a hook in `features/<name>/api.ts` and use it in the component.

For a module without a `routes.go` (`users`, `roles`), add the route line in `main.go` instead.

## C. New permission or role only

- **Permission:** constant plus entry in `All` in `permissions.go`, `make db-seed`, then use `roles.RequirePermission(authz, auditService, permissions.X)` on the route.
- **Role:** create it from the Roles page or the API. A role every user should have by default belongs in `seeds/main.go` next to `MEMBER`.

## D. Pitfalls

- A permission missing from `All` is never seeded, so nobody holds it, not even `SUPER_ADMIN` after a fresh seed.
- Forgetting `make db-seed` after adding a permission gives 403 for everyone except users whose role was granted by hand.
- `apps/web` imports the **built** `packages/api-client/dist`. Run `make api-client-generate` (it builds) or the web will not see new types.
- `make openapi-validate` fails when generated files are stale. Run both generate targets before committing.
- Invalid UUIDs in a path must answer 404, not 500. The `notFound` helper in `items/repository.go` handles that.
- Audit writes are fail-closed: if the audit insert fails, the request returns 500 even though the data changed.
