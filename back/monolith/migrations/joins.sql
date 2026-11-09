CREATE TABLE IF NOT EXISTS join_children (
    parent_id  TEXT    NOT NULL,
    child_id   TEXT    NOT NULL,
    payload    TEXT    NOT NULL,
    expires_at INTEGER NOT NULL,
    PRIMARY KEY (parent_id, child_id)
);

CREATE INDEX IF NOT EXISTS idx_join_children_expires_at ON join_children (expires_at);

CREATE TABLE IF NOT EXISTS join_claims (
    parent_id  TEXT    PRIMARY KEY,
    expires_at INTEGER NOT NULL
);
