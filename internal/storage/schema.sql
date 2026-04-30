-- conversations: one per visitor session
CREATE TABLE IF NOT EXISTS conversations (
  id              TEXT PRIMARY KEY,
  visitor_handle  TEXT,
  surface         TEXT NOT NULL,
  started_at      INTEGER NOT NULL,
  ended_at        INTEGER,
  outcome         TEXT,
  metadata        BLOB
);

-- messages: append-only, per conversation
CREATE TABLE IF NOT EXISTS messages (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  conversation_id TEXT NOT NULL REFERENCES conversations(id),
  role            TEXT NOT NULL,
  content_enc     BLOB NOT NULL,
  ts              INTEGER NOT NULL,
  llm_provider    TEXT,
  tokens_in       INTEGER,
  tokens_out      INTEGER,
  flags           TEXT
);
CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id, ts);

-- audit: every owner-visible event, SQLite mirror (JSONL is canonical)
CREATE TABLE IF NOT EXISTS audit (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  event_id        TEXT NOT NULL UNIQUE,
  ts              INTEGER NOT NULL,
  category        TEXT NOT NULL,
  severity        TEXT NOT NULL,
  payload         BLOB NOT NULL,
  conversation_id TEXT REFERENCES conversations(id)
);
CREATE INDEX IF NOT EXISTS idx_audit_ts ON audit(ts DESC);

-- capability_grants: per-visitor / per-class data-class authorizations
CREATE TABLE IF NOT EXISTS capability_grants (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  scope           TEXT NOT NULL,
  data_class      TEXT NOT NULL,
  permission      TEXT NOT NULL,
  granted_at      INTEGER NOT NULL,
  granted_by      TEXT NOT NULL,
  expires_at      INTEGER
);
CREATE INDEX IF NOT EXISTS idx_grants_scope_class ON capability_grants(scope, data_class);

-- facts: long-term Symmate knowledge of the host
CREATE TABLE IF NOT EXISTS facts (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  fact_class      TEXT NOT NULL,
  content_enc     BLOB NOT NULL,
  authored_at     INTEGER NOT NULL,
  source          TEXT NOT NULL,
  confidence      REAL,
  retracted_at    INTEGER
);

-- fact_embeddings: optional BGE-M3 vector per fact
CREATE TABLE IF NOT EXISTS fact_embeddings (
  fact_id         INTEGER PRIMARY KEY REFERENCES facts(id),
  vec             BLOB NOT NULL
);
