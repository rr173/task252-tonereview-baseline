package store

import (
	"database/sql"

	"task252-tonereview/internal/model"
)

// CreateBatch 写入一个新的田野批次（整理中）。
func (s *Store) CreateBatch(b *model.FieldBatch) error {
	_, err := s.db.Exec(
		`INSERT INTO batches(id, code, title, status, created_at, updated_at)
		 VALUES(?,?,?,?,?,?)`,
		b.ID, b.Code, b.Title, b.Status, b.CreatedAt, b.UpdatedAt)
	return err
}

// GetBatch 按 ID 读取批次。
func (s *Store) GetBatch(id string) (*model.FieldBatch, error) {
	row := s.db.QueryRow(
		`SELECT id, code, title, status, created_at, updated_at FROM batches WHERE id=?`, id)
	b := &model.FieldBatch{}
	err := row.Scan(&b.ID, &b.Code, &b.Title, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ListBatches 返回全部批次，按创建时间升序。
func (s *Store) ListBatches() ([]*model.FieldBatch, error) {
	rows, err := s.db.Query(
		`SELECT id, code, title, status, created_at, updated_at FROM batches ORDER BY created_at ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.FieldBatch{}
	for rows.Next() {
		b := &model.FieldBatch{}
		if err := rows.Scan(&b.ID, &b.Code, &b.Title, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// UpdateBatchStatus 更新批次状态与时间戳。
func (s *Store) UpdateBatchStatus(id, status string, updatedAt int64) error {
	res, err := s.db.Exec(`UPDATE batches SET status=?, updated_at=? WHERE id=?`, status, updatedAt, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// BatchSummary 汇总批次内的片段/对立/版本计数与状态分布。
type BatchSummary struct {
	Batch          *model.FieldBatch
	SegmentTotal   int
	SegByStatus    map[string]int
	OppTotal       int
	OppConfirmed   int
	VersionTotal   int
	VersionFrozen  int
}

// SummaryForBatch 读取批次及其统计。
func (s *Store) SummaryForBatch(id string) (*BatchSummary, error) {
	b, err := s.GetBatch(id)
	if err != nil {
		return nil, err
	}
	sum := &BatchSummary{Batch: b, SegByStatus: map[string]int{}}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM segments WHERE batch_id=?`, id).Scan(&sum.SegmentTotal); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT status, COUNT(*) FROM segments WHERE batch_id=? GROUP BY status`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var st string
		var n int
		if err := rows.Scan(&st, &n); err != nil {
			rows.Close()
			return nil, err
		}
		sum.SegByStatus[st] = n
	}
	rows.Close()
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM oppositions WHERE batch_id=?`, id).Scan(&sum.OppTotal); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM oppositions WHERE batch_id=? AND status=?`, id, model.OppConfirmed).Scan(&sum.OppConfirmed); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM versions WHERE batch_id=?`, id).Scan(&sum.VersionTotal); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM versions WHERE batch_id=? AND status=?`, id, model.VerFrozen).Scan(&sum.VersionFrozen); err != nil {
		return nil, err
	}
	return sum, nil
}
