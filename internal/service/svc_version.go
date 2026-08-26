package service

import (
	"task252-tonereview/internal/adjudicate"
	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
	"task252-tonereview/internal/version"
)

// CreateVersion 为批次创建分析版本草稿。
// 只有已发布批次才能创建分析版本，避免在尚未完成复核基础的批次上
// 产生版本链。
func (s *Service) CreateVersion(batchID, note string) (*model.AnalysisVersion, error) {
	b, err := s.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}
	if b.Status != model.BatchPublished {
		return nil, model.ErrInvalidState
	}
	v, err := s.store.NextVersion(batchID)
	if err != nil {
		return nil, err
	}
	av := &model.AnalysisVersion{
		ID:        model.NewID("ver"),
		BatchID:   batchID,
		Version:   v,
		Status:    model.VerDraft,
		Note:      note,
		CreatedAt: now(),
	}
	if err := s.store.CreateVersion(av); err != nil {
		return nil, err
	}
	return av, nil
}

// GetVersion 读取版本。
func (s *Service) GetVersion(id string) (*model.AnalysisVersion, error) {
	return s.store.GetVersion(id)
}

// ListVersions 列出批次全部版本。
func (s *Service) ListVersions(batchID string) ([]*model.AnalysisVersion, error) {
	return s.store.ListVersions(batchID)
}

// ShareVersion 草稿 → 共享。
func (s *Service) ShareVersion(id string) error {
	v, err := s.store.GetVersion(id)
	if err != nil {
		return err
	}
	if !adjudicate.CanVersionTransition(v.Status, model.VerShared) {
		return model.ErrInvalidState
	}
	b, err := s.store.GetBatch(v.BatchID)
	if err != nil {
		return err
	}
	if b.Status != model.BatchPublished {
		return model.ErrInvalidState
	}
	return s.store.UpdateVersionStatus(id, model.VerShared, v.Snapshot)
}

// FreezeVersion 共享 → 冻结：构建不可变快照，并将同批次旧冻结版本置为替代。
func (s *Service) FreezeVersion(id string) error {
	v, err := s.store.GetVersion(id)
	if err != nil {
		return err
	}
	if !adjudicate.CanVersionTransition(v.Status, model.VerFrozen) {
		return model.ErrInvalidState
	}
	b, err := s.store.GetBatch(v.BatchID)
	if err != nil {
		return err
	}
	if b.Status != model.BatchPublished {
		return model.ErrInvalidState
	}
	segs, err := s.store.ListSegments(store.SegmentFilter{BatchID: v.BatchID})
	if err != nil {
		return err
	}
	opps, err := s.store.ListOppositions(v.BatchID, "")
	if err != nil {
		return err
	}
	snap, err := version.BuildSnapshot(v.BatchID, segs, opps)
	if err != nil {
		return err
	}
	if err := s.store.UpdateVersionStatus(id, model.VerFrozen, snap); err != nil {
		return err
	}
	// 旧冻结版本退为替代。
	return s.store.SupersedeFrozen(v.BatchID, id)
}
