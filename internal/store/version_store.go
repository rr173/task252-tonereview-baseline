package store

import (
	"database/sql"

	"task252-tonereview/internal/model"
)

// NextVersion 返回某批次下一个版本号（当前最大 +1，无则为 1）。
func (s *Store) NextVersion(batchID string) (int, error) {
	var mx sql.NullInt64
	err := s.db.QueryRow(`SELECT MAX(version) FROM versions WHERE batch_id=?`, batchID).Scan(&mx)
	if err != nil {
		return 0, err
	}
	if !mx.Valid {
		return 1, nil
	}
	return int(mx.Int64) + 1, nil
}

// CreateVersion 写入分析版本（草稿）。
func (s *Store) CreateVersion(v *model.AnalysisVersion) error {
	_, err := s.db.Exec(
		`INSERT INTO versions(id, batch_id, version, status, snapshot, note, created_at)
		 VALUES(?,?,?,?,?,?,?)`,
		v.ID, v.BatchID, v.Version, v.Status, v.Snapshot, v.Note, v.CreatedAt)
	return err
}

// GetVersion 按 ID 读取版本。
func (s *Store) GetVersion(id string) (*model.AnalysisVersion, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, version, status, snapshot, note, created_at FROM versions WHERE id=?`, id)
	v := &model.AnalysisVersion{}
	err := row.Scan(&v.ID, &v.BatchID, &v.Version, &v.Status, &v.Snapshot, &v.Note, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

// ListVersions 返回某批次的全部版本（按版本号升序）。
func (s *Store) ListVersions(batchID string) ([]*model.AnalysisVersion, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, version, status, snapshot, note, created_at FROM versions WHERE batch_id=? ORDER BY version ASC`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.AnalysisVersion{}
	for rows.Next() {
		v := &model.AnalysisVersion{}
		if err := rows.Scan(&v.ID, &v.BatchID, &v.Version, &v.Status, &v.Snapshot, &v.Note, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// UpdateVersionStatus 更新版本状态与快照。
func (s *Store) UpdateVersionStatus(id, status, snapshot string) error {
	res, err := s.db.Exec(`UPDATE versions SET status=?, snapshot=? WHERE id=?`, status, snapshot, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SupersedeFrozen 将某批次除 exceptID 外的全部冻结版本置为替代（superseded）。
func (s *Store) SupersedeFrozen(batchID, exceptID string) error {
	_, err := s.db.Exec(
		`UPDATE versions SET status=? WHERE batch_id=? AND status=? AND id<>?`,
		model.VerSuperseded, batchID, model.VerFrozen, exceptID)
	return err
}
