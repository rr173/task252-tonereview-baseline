package service

import (
	"task252-tonereview/internal/adjudicate"
	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

// CreateBatch 新建田野批次（整理中）。code 为空时自动生成。
func (s *Service) CreateBatch(code, title string) (*model.FieldBatch, error) {
	if code == "" {
		code = model.NewID("batch")
	}
	b := &model.FieldBatch{
		ID:        model.NewID("batch"),
		Code:      code,
		Title:     title,
		Status:    model.BatchCollecting,
		CreatedAt: now(),
		UpdatedAt: now(),
	}
	if err := s.store.CreateBatch(b); err != nil {
		return nil, err
	}
	return b, nil
}

// GetBatch 读取批次。
func (s *Service) GetBatch(id string) (*model.FieldBatch, error) {
	return s.store.GetBatch(id)
}

// ListBatches 列出全部批次。
func (s *Service) ListBatches() ([]*model.FieldBatch, error) {
	return s.store.ListBatches()
}

// BatchSummary 返回批次统计。
func (s *Service) BatchSummary(id string) (*store.BatchSummary, error) {
	return s.store.SummaryForBatch(id)
}

// StartReview 整理中 → 待复核。
func (s *Service) StartReview(id string) error {
	b, err := s.store.GetBatch(id)
	if err != nil {
		return err
	}
	if !adjudicate.CanBatchTransition(b.Status, model.BatchReviewing) {
		return model.ErrInvalidState
	}
	return s.store.UpdateBatchStatus(id, model.BatchReviewing, now())
}

// PublishBatch 待复核 → 已发布。
func (s *Service) PublishBatch(id string) error {
	b, err := s.store.GetBatch(id)
	if err != nil {
		return err
	}
	if !adjudicate.CanBatchTransition(b.Status, model.BatchPublished) {
		return model.ErrInvalidState
	}
	return s.store.UpdateBatchStatus(id, model.BatchPublished, now())
}

// SealBatch 已发布 → 封存（终态，拒绝后续修改）。
func (s *Service) SealBatch(id string) error {
	b, err := s.store.GetBatch(id)
	if err != nil {
		return err
	}
	if !adjudicate.CanBatchTransition(b.Status, model.BatchSealed) {
		return model.ErrInvalidState
	}
	return s.store.UpdateBatchStatus(id, model.BatchSealed, now())
}

// guardBatchMutable 封存批次拒绝一切修改。
func (s *Service) guardBatchMutable(batchID string) error {
	b, err := s.store.GetBatch(batchID)
	if err != nil {
		return err
	}
	if b.Status == model.BatchSealed {
		return model.ErrSealed
	}
	return nil
}
