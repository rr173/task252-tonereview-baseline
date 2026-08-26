package store

import (
	"database/sql"

	"task252-tonereview/internal/model"
)

// CreateSegment 写入发音片段（待校验）。
func (s *Store) CreateSegment(seg *model.Segment) error {
	_, err := s.db.Exec(
		`INSERT INTO segments(id, batch_id, lexical_item, phonetic_seg, speaker_id, audio_fp, status, duration_ms, recorded_at, tone_type, created_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		seg.ID, seg.BatchID, seg.LexicalItem, seg.PhoneticSeg, seg.SpeakerID, seg.AudioFP,
		seg.Status, seg.DurationMs, seg.RecordedAt, seg.ToneType, seg.CreatedAt)
	return err
}

// GetSegment 按 ID 读取片段。
func (s *Store) GetSegment(id string) (*model.Segment, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, lexical_item, phonetic_seg, speaker_id, audio_fp, status, duration_ms, recorded_at, tone_type, created_at
		 FROM segments WHERE id=?`, id)
	return scanSegment(row)
}

func scanSegment(row *sql.Row) (*model.Segment, error) {
	seg := &model.Segment{}
	err := row.Scan(&seg.ID, &seg.BatchID, &seg.LexicalItem, &seg.PhoneticSeg, &seg.SpeakerID,
		&seg.AudioFP, &seg.Status, &seg.DurationMs, &seg.RecordedAt, &seg.ToneType, &seg.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return seg, nil
}

// GetSegmentByFP 按录音指纹读取片段（幂等去重用）。
func (s *Store) GetSegmentByFP(fp string) (*model.Segment, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, lexical_item, phonetic_seg, speaker_id, audio_fp, status, duration_ms, recorded_at, tone_type, created_at
		 FROM segments WHERE audio_fp=? LIMIT 1`, fp)
	return scanSegment(row)
}

// SegmentFilter 片段列表过滤条件。
type SegmentFilter struct {
	BatchID     string
	SpeakerID   string
	Status      string
	PhoneticSeg string
}

// ListSegments 按过滤条件返回片段（可组合）。
func (s *Store) ListSegments(f SegmentFilter) ([]*model.Segment, error) {
	q := `SELECT id, batch_id, lexical_item, phonetic_seg, speaker_id, audio_fp, status, duration_ms, recorded_at, tone_type, created_at FROM segments WHERE 1=1`
	args := []any{}
	if f.BatchID != "" {
		q += ` AND batch_id=?`
		args = append(args, f.BatchID)
	}
	if f.SpeakerID != "" {
		q += ` AND speaker_id=?`
		args = append(args, f.SpeakerID)
	}
	if f.Status != "" {
		q += ` AND status=?`
		args = append(args, f.Status)
	}
	if f.PhoneticSeg != "" {
		q += ` AND phonetic_seg=?`
		args = append(args, f.PhoneticSeg)
	}
	q += ` ORDER BY created_at ASC, id ASC`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.Segment{}
	for rows.Next() {
		seg := &model.Segment{}
		if err := rows.Scan(&seg.ID, &seg.BatchID, &seg.LexicalItem, &seg.PhoneticSeg, &seg.SpeakerID,
			&seg.AudioFP, &seg.Status, &seg.DurationMs, &seg.RecordedAt, &seg.ToneType, &seg.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, seg)
	}
	return out, rows.Err()
}

// UpdateSegmentStatus 更新片段状态。
func (s *Store) UpdateSegmentStatus(id, status string) error {
	res, err := s.db.Exec(`UPDATE segments SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SetSegmentToneType 写入片段调型分类（基线重算后回填）。
func (s *Store) SetSegmentToneType(id, toneType string) error {
	res, err := s.db.Exec(`UPDATE segments SET tone_type=? WHERE id=?`, toneType, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SegmentExistsForBatch 判断批次内是否已存在某词条/音段/说话人组合（避免重复导入）。
func (s *Store) SegmentWithKey(batchID, lexicalItem, phoneticSeg, speakerID string) (*model.Segment, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, lexical_item, phonetic_seg, speaker_id, audio_fp, status, duration_ms, recorded_at, tone_type, created_at
		 FROM segments WHERE batch_id=? AND lexical_item=? AND phonetic_seg=? AND speaker_id=? LIMIT 1`,
		batchID, lexicalItem, phoneticSeg, speakerID)
	return scanSegment(row)
}
