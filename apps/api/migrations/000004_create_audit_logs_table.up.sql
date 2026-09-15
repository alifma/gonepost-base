CREATE TABLE audit_logs (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Nullable: some audited events have no authenticated actor yet
    -- (e.g. a failed login attempt with an unknown/wrong email).
    actor_user_id  uuid REFERENCES users (id) ON DELETE SET NULL,
    action         text NOT NULL,
    resource       text NOT NULL,
    resource_id    text,
    result         text NOT NULL CHECK (result IN ('success', 'failure')),
    request_id     text,
    trace_id       text,
    -- Sanitized only — never passwords, tokens, or session identifiers.
    metadata       jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_created_at_idx ON audit_logs (created_at DESC);
CREATE INDEX audit_logs_actor_user_id_idx ON audit_logs (actor_user_id);
CREATE INDEX audit_logs_action_idx ON audit_logs (action);

-- Retention: no automated purge job yet (would need a worker/scheduler —
-- see PLAN.md Deferred Work). Manual purge for anyone running this before
-- that exists, e.g. keep 1 year:
--   DELETE FROM audit_logs WHERE created_at < now() - interval '1 year';
