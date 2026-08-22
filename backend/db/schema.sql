-- 番茄投资人 schema v1

CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  owner_id      TEXT NOT NULL DEFAULT 'me',   -- 预留多用户
  tomato_minutes INTEGER NOT NULL DEFAULT 480, -- 全局番茄时长(分钟), 默认 8h
  created_at    DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS projects (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  owner_id        TEXT NOT NULL DEFAULT 'me',
  title           TEXT NOT NULL,
  necessity      TEXT NOT NULL DEFAULT '',    -- 必要性
  mvp_scope      TEXT NOT NULL DEFAULT '',    -- 最小集
  stoploss_note  TEXT NOT NULL DEFAULT '',    -- 止损说明
  budget_tomatoes INTEGER NOT NULL DEFAULT 1, -- 预算番茄数
  tomato_minutes  INTEGER,                    -- 项目级覆盖; NULL=用全局
  status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active','completed','archived_stoploss')),
  created_at      DATETIME NOT NULL DEFAULT (datetime('now')),
  archived_at     DATETIME
);

CREATE TABLE IF NOT EXISTS milestones (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id  INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  done        INTEGER NOT NULL DEFAULT 0,
  order_idx   INTEGER NOT NULL DEFAULT 0,
  created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sessions (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  project_id      INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  status          TEXT NOT NULL DEFAULT 'running'
                    CHECK (status IN ('running','paused','ended')),
  started_at      DATETIME NOT NULL DEFAULT (datetime('now')),
  ended_at        DATETIME,
  consumed_tomato INTEGER NOT NULL DEFAULT 0,  -- 结束时置 1 表示消耗一番茄
  note            TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_id);

CREATE TABLE IF NOT EXISTS session_events (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id  INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  type        TEXT NOT NULL
                CHECK (type IN ('pause','resume','comment','voice_note')),
  payload     TEXT NOT NULL DEFAULT '{}',     -- JSON
  voice_file  TEXT,                            -- 仅 voice_note
  at          DATETIME NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_events_session ON session_events(session_id);

CREATE TABLE IF NOT EXISTS voice_notes (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  session_event_id   INTEGER NOT NULL REFERENCES session_events(id) ON DELETE CASCADE,
  file_path          TEXT NOT NULL,
  mime               TEXT NOT NULL,
  duration_ms        INTEGER NOT NULL DEFAULT 0,
  text               TEXT NOT NULL DEFAULT '',
  created_at         DATETIME NOT NULL DEFAULT (datetime('now'))
);
