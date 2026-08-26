// Package store 负责领域实体的 SQLite 持久化。使用纯 Go 驱动 modernc.org/sqlite，
// 无需 CGO，可在 GOTOOLCHAIN=local + CGO_ENABLED=0 环境下离线构建。
// 全部写操作基于 database/sql，并对状态字段做约束，保证重启后可恢复。
package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Store 封装数据库连接与全部实体的读写方法。
type Store struct {
	db *sql.DB
}

// schema 为建表语句，全部幂等（IF NOT EXISTS）。
const schema = `
CREATE TABLE IF NOT EXISTS speakers (
  id           TEXT PRIMARY KEY,
  code         TEXT NOT NULL,
  dialect      TEXT NOT NULL DEFAULT '',
  gender       TEXT NOT NULL DEFAULT '',
  birth_year   INTEGER NOT NULL DEFAULT 0,
  baseline_log REAL NOT NULL DEFAULT 0,
  has_baseline INTEGER NOT NULL DEFAULT 0,
  created_at   INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_speakers_code ON speakers(code);

CREATE TABLE IF NOT EXISTS batches (
  id         TEXT PRIMARY KEY,
  code       TEXT NOT NULL,
  title      TEXT NOT NULL DEFAULT '',
  status     TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_batches_status ON batches(status);

CREATE TABLE IF NOT EXISTS segments (
  id            TEXT PRIMARY KEY,
  batch_id      TEXT NOT NULL,
  lexical_item  TEXT NOT NULL,
  phonetic_seg  TEXT NOT NULL,
  speaker_id    TEXT NOT NULL,
  audio_fp      TEXT NOT NULL DEFAULT '',
  status        TEXT NOT NULL,
  duration_ms   INTEGER NOT NULL DEFAULT 0,
  recorded_at   INTEGER NOT NULL DEFAULT 0,
  tone_type     TEXT NOT NULL DEFAULT 'unknown',
  created_at    INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_segments_batch ON segments(batch_id);
CREATE INDEX IF NOT EXISTS idx_segments_speaker ON segments(speaker_id);
CREATE INDEX IF NOT EXISTS idx_segments_fp ON segments(audio_fp);
CREATE INDEX IF NOT EXISTS idx_segments_status ON segments(status);

CREATE TABLE IF NOT EXISTS f0_samples (
  id         TEXT PRIMARY KEY,
  segment_id TEXT NOT NULL,
  t_ms       INTEGER NOT NULL,
  f0_hz      REAL NOT NULL,
  reliable   INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_f0_segment ON f0_samples(segment_id, t_ms);

CREATE TABLE IF NOT EXISTS oppositions (
  id               TEXT PRIMARY KEY,
  batch_id         TEXT NOT NULL,
  lexical_a        TEXT NOT NULL,
  phonetic_seg     TEXT NOT NULL,
  lexical_b        TEXT NOT NULL,
  status           TEXT NOT NULL,
  opposition_score REAL NOT NULL DEFAULT 0,
  decision_reason  TEXT NOT NULL DEFAULT '',
  decided_at       INTEGER NOT NULL DEFAULT 0,
  created_at       INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_opp_batch ON oppositions(batch_id);
CREATE INDEX IF NOT EXISTS idx_opp_status ON oppositions(status);

CREATE TABLE IF NOT EXISTS evidences (
  id            TEXT PRIMARY KEY,
  opposition_id TEXT NOT NULL,
  segment_id    TEXT NOT NULL,
  side          TEXT NOT NULL,
  normalized    TEXT NOT NULL DEFAULT '',
  tone_type     TEXT NOT NULL DEFAULT 'unknown',
  created_at    INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_evi_opp ON evidences(opposition_id);

CREATE TABLE IF NOT EXISTS versions (
  id         TEXT PRIMARY KEY,
  batch_id   TEXT NOT NULL,
  version    INTEGER NOT NULL,
  status     TEXT NOT NULL,
  snapshot   TEXT NOT NULL DEFAULT '',
  note       TEXT NOT NULL DEFAULT '',
  created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_versions_batch ON versions(batch_id);
CREATE INDEX IF NOT EXISTS idx_versions_status ON versions(status);
`

// Open 打开（或创建）位于 path 的 SQLite 数据库，并应用 schema。
// path 可为 ":memory:"（用于测试），或磁盘文件路径（支持重启恢复）。
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// 开启外键与 WAL，提升一致性与并发写入稳定性。
	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// DB 暴露底层连接，供需要在事务中组合多实体操作的调用方使用。
func (s *Store) DB() *sql.DB { return s.db }

// Close 关闭数据库连接。
func (s *Store) Close() error { return s.db.Close() }
