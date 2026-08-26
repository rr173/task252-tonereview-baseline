package store

import (
	"database/sql"

	"task252-tonereview/internal/model"
)

// CreateOpposition 写入候选声调对立。
func (s *Store) CreateOpposition(o *model.ToneOpposition) error {
	_, err := s.db.Exec(
		`INSERT INTO oppositions(id, batch_id, lexical_a, phonetic_seg, lexical_b, status, opposition_score, decision_reason, decided_at, created_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		o.ID, o.BatchID, o.LexicalA, o.PhoneticSeg, o.LexicalB, o.Status, o.OppositionScore, o.DecisionReason, o.DecidedAt, o.CreatedAt)
	return err
}

// GetOpposition 按 ID 读取对立。
func (s *Store) GetOpposition(id string) (*model.ToneOpposition, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, lexical_a, phonetic_seg, lexical_b, status, opposition_score, decision_reason, decided_at, created_at
		 FROM oppositions WHERE id=?`, id)
	o := &model.ToneOpposition{}
	err := row.Scan(&o.ID, &o.BatchID, &o.LexicalA, &o.PhoneticSeg, &o.LexicalB, &o.Status,
		&o.OppositionScore, &o.DecisionReason, &o.DecidedAt, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

// ListOppositions 按批次/状态过滤返回对立。
func (s *Store) ListOppositions(batchID, status string) ([]*model.ToneOpposition, error) {
	q := `SELECT id, batch_id, lexical_a, phonetic_seg, lexical_b, status, opposition_score, decision_reason, decided_at, created_at FROM oppositions WHERE 1=1`
	args := []any{}
	if batchID != "" {
		q += ` AND batch_id=?`
		args = append(args, batchID)
	}
	if status != "" {
		q += ` AND status=?`
		args = append(args, status)
	}
	q += ` ORDER BY created_at ASC, id ASC`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.ToneOpposition{}
	for rows.Next() {
		o := &model.ToneOpposition{}
		if err := rows.Scan(&o.ID, &o.BatchID, &o.LexicalA, &o.PhoneticSeg, &o.LexicalB, &o.Status,
			&o.OppositionScore, &o.DecisionReason, &o.DecidedAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// UpdateOpposition 更新对立的状态、得分与裁决信息。
func (s *Store) UpdateOpposition(o *model.ToneOpposition) error {
	res, err := s.db.Exec(
		`UPDATE oppositions SET status=?, opposition_score=?, decision_reason=?, decided_at=? WHERE id=?`,
		o.Status, o.OppositionScore, o.DecisionReason, o.DecidedAt, o.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}
