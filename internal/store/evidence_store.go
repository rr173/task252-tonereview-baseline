package store

import (
	"task252-tonereview/internal/model"
)

// CreateEvidence 写入一条证据簇条目。
func (s *Store) CreateEvidence(e *model.Evidence) error {
	_, err := s.db.Exec(
		`INSERT INTO evidences(id, opposition_id, segment_id, side, normalized, tone_type, created_at)
		 VALUES(?,?,?,?,?,?,?)`,
		e.ID, e.OppositionID, e.SegmentID, e.Side, e.Normalized, e.ToneType, e.CreatedAt)
	return err
}

// ListEvidence 读取某对立的全部证据（按创建时间升序）。
func (s *Store) ListEvidence(oppositionID string) ([]*model.Evidence, error) {
	rows, err := s.db.Query(
		`SELECT id, opposition_id, segment_id, side, normalized, tone_type, created_at
		 FROM evidences WHERE opposition_id=? ORDER BY created_at ASC, id ASC`, oppositionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.Evidence{}
	for rows.Next() {
		e := &model.Evidence{}
		if err := rows.Scan(&e.ID, &e.OppositionID, &e.SegmentID, &e.Side, &e.Normalized, &e.ToneType, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// DeleteEvidenceForOpposition 删除某对立的全部证据（重算前清空）。
func (s *Store) DeleteEvidenceForOpposition(oppositionID string) error {
	_, err := s.db.Exec(`DELETE FROM evidences WHERE opposition_id=?`, oppositionID)
	return err
}

// DeleteEvidence 删除某对立中指定片段/侧的旧证据。
func (s *Store) DeleteEvidence(oppositionID, segmentID, side string) error {
	_, err := s.db.Exec(
		`DELETE FROM evidences WHERE opposition_id=? AND segment_id=? AND side=?`,
		oppositionID, segmentID, side,
	)
	return err
}

// EvidenceCount 返回某对立的证据条数。
func (s *Store) EvidenceCount(oppositionID string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM evidences WHERE opposition_id=?`, oppositionID).Scan(&n)
	return n, err
}
