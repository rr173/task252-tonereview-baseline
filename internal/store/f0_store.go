package store

import (
	"database/sql"

	"task252-tonereview/internal/model"
)

// InsertF0 批量写入基频采样（事务内）。t_ms 调用方需保证单调。
func (s *Store) InsertF0(samples []model.F0Sample) error {
	if len(samples) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO f0_samples(id, segment_id, t_ms, f0_hz, reliable) VALUES(?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, sm := range samples {
		if _, err := stmt.Exec(sm.ID, sm.SegmentID, sm.TMs, sm.F0Hz, boolToInt(sm.Reliable)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ListF0 按时间升序返回某片段的基频采样。
func (s *Store) ListF0(segmentID string) ([]model.F0Sample, error) {
	rows, err := s.db.Query(
		`SELECT id, segment_id, t_ms, f0_hz, reliable FROM f0_samples WHERE segment_id=? ORDER BY t_ms ASC`, segmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.F0Sample{}
	for rows.Next() {
		sm := model.F0Sample{}
		var rel int
		if err := rows.Scan(&sm.ID, &sm.SegmentID, &sm.TMs, &sm.F0Hz, &rel); err != nil {
			return nil, err
		}
		sm.Reliable = rel != 0
		out = append(out, sm)
	}
	return out, rows.Err()
}

// MaxF0Time 返回某片段已存在的最大采样时间（用于追加时的单调性校验）。
func (s *Store) MaxF0Time(segmentID string) (int64, error) {
	var mx sql.NullInt64
	err := s.db.QueryRow(`SELECT MAX(t_ms) FROM f0_samples WHERE segment_id=?`, segmentID).Scan(&mx)
	if err != nil {
		return 0, err
	}
	if !mx.Valid {
		return 0, nil
	}
	return mx.Int64, nil
}

// F0CountForSegment 返回片段可靠样本数（基线计算用）。
func (s *Store) F0CountForSegment(segmentID string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM f0_samples WHERE segment_id=? AND reliable=1`, segmentID).Scan(&n)
	return n, err
}
