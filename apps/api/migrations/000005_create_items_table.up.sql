-- Reference CRUD table (see docs/guides/adding-a-feature.md). Every row is
-- owned by one user; queries always filter on owner_id.
CREATE TABLE items (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        text NOT NULL,
    description text,
    status      text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX items_owner_id_created_at_idx ON items (owner_id, created_at DESC);
