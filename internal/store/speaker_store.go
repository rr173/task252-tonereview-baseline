package store

import (
	"database/sql"

	"task252-tonereview/internal/model"
)

// CreateSpeaker 写入说话人。
func (s *Store) CreateSpeaker(sp *model.Speaker) error {
	_, err := s.db.Exec(
		`INSERT INTO speakers(id, code, dialect, gender, birth_year, baseline_log, has_baseline, created_at)
		 VALUES(?,?,?,?,?,?,?,?)`,
		sp.ID, sp.Code, sp.Dialect, sp.Gender, sp.BirthYear, sp.BaselineLog, boolToInt(sp.HasBaseline), sp.CreatedAt)
	return err
}

// GetSpeaker 按 ID 读取说话人。
func (s *Store) GetSpeaker(id string) (*model.Speaker, error) {
	row := s.db.QueryRow(
		`SELECT id, code, dialect, gender, birth_year, baseline_log, has_baseline, created_at
		 FROM speakers WHERE id=?`, id)
	sp := &model.Speaker{}
	var hb int
	err := row.Scan(&sp.ID, &sp.Code, &sp.Dialect, &sp.Gender, &sp.BirthYear, &sp.BaselineLog, &hb, &sp.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sp.HasBaseline = hb != 0
	return sp, nil
}

// ListSpeakers 返回全部说话人。
func (s *Store) ListSpeakers() ([]*model.Speaker, error) {
	rows, err := s.db.Query(
		`SELECT id, code, dialect, gender, birth_year, baseline_log, has_baseline, created_at
		 FROM speakers ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.Speaker{}
	for rows.Next() {
		sp := &model.Speaker{}
		var hb int
		if err := rows.Scan(&sp.ID, &sp.Code, &sp.Dialect, &sp.Gender, &sp.BirthYear, &sp.BaselineLog, &hb, &sp.CreatedAt); err != nil {
			return nil, err
		}
		sp.HasBaseline = hb != 0
		out = append(out, sp)
	}
	return out, rows.Err()
}

// SetBaseline 写入说话人基线（HasBaseline 置真）。
func (s *Store) SetBaseline(id string, baselineLog float64, hasBaseline bool, _ int64) error {
	res, err := s.db.Exec(
		`UPDATE speakers SET baseline_log=?, has_baseline=? WHERE id=?`,
		baselineLog, boolToInt(hasBaseline), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// ListUsableSegmentsBySpeaker 返回某说话人处于可用状态的片段 ID 列表（供基线计算）。
func (s *Store) ListUsableSegmentsBySpeaker(speakerID string) ([]string, error) {
	rows, err := s.db.Query(`SELECT id FROM segments WHERE speaker_id=? AND status=?`, speakerID, model.SegUsable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// SpeakerInSealedBatch 报告说话人是否有片段属于已封存的批次。
// 说话人基线为全局单值，一旦参与封存批次的证据即视为冻结证据的一部分；
// 重算会改写这些片段的调型，故须在 service 层拒绝以保持现有数据不变。
func (s *Store) SpeakerInSealedBatch(speakerID string) (bool, error) {
	var n int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM segments seg
		 JOIN batches b ON b.id = seg.batch_id
		 WHERE seg.speaker_id=? AND b.status=?`,
		speakerID, model.BatchSealed).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
