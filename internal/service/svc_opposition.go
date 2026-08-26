package service

import (
	"fmt"

	"task252-tonereview/internal/adjudicate"
	"task252-tonereview/internal/compare"
	"task252-tonereview/internal/model"
)

// CreateOpposition 新建候选声调对立（同一音段上的两个词条）。
func (s *Service) CreateOpposition(batchID, lexicalA, phoneticSeg, lexicalB string) (*model.ToneOpposition, error) {
	if err := s.guardBatchMutable(batchID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetBatch(batchID); err != nil {
		return nil, err
	}
	if lexicalA == "" || lexicalB == "" || phoneticSeg == "" {
		return nil, model.ErrBadRequest
	}
	if lexicalA == lexicalB {
		return nil, model.ErrBadRequest
	}
	o := &model.ToneOpposition{
		ID:          model.NewID("opp"),
		BatchID:     batchID,
		LexicalA:    lexicalA,
		PhoneticSeg: phoneticSeg,
		LexicalB:    lexicalB,
		Status:      model.OppCandidate,
		CreatedAt:   now(),
	}
	if err := s.store.CreateOpposition(o); err != nil {
		return nil, err
	}
	return o, nil
}

// GetOpposition 读取对立。
func (s *Service) GetOpposition(id string) (*model.ToneOpposition, error) {
	return s.store.GetOpposition(id)
}

// ListOppositions 列出对立（按批次/状态过滤）。
func (s *Service) ListOppositions(batchID, status string) ([]*model.ToneOpposition, error) {
	return s.store.ListOppositions(batchID, status)
}

// AddEvidence 将某片段的归一化轮廓作为证据加入对立的某一侧（a/b）。
// 同一片段同一侧重复加入会先删后插（基线变化后可刷新）。
func (s *Service) AddEvidence(oppID, segmentID, side string) error {
	if side != "a" && side != "b" {
		return model.ErrBadRequest
	}
	o, err := s.store.GetOpposition(oppID)
	if err != nil {
		return err
	}
	if err := s.guardBatchMutable(o.BatchID); err != nil {
		return err
	}
	seg, err := s.store.GetSegment(segmentID)
	if err != nil {
		return err
	}
	if seg.BatchID != o.BatchID || seg.PhoneticSeg != o.PhoneticSeg {
		return model.ErrBadRequest
	}
	if side == "a" && seg.LexicalItem != o.LexicalA {
		return model.ErrBadRequest
	}
	if side == "b" && seg.LexicalItem != o.LexicalB {
		return model.ErrBadRequest
	}
	// 只有可用录音可生成对立证据：噪声/排除（已否决）或待校验（未完成校验）
	// 的片段一律拒绝。校验须先于任何证据改动，使失败时不删旧证据、不新增证据。
	if seg.Status != model.SegUsable {
		return fmt.Errorf("%w: segment %s not usable (status=%s)", model.ErrInvalidState, seg.ID, seg.Status)
	}
	curve, err := s.contourForSegment(segmentID, resampleN)
	if err != nil {
		return err
	}
	// 去重：删除同一对立/片段/侧的旧证据。
	evs, err := s.store.ListEvidence(oppID)
	if err != nil {
		return err
	}
	for _, e := range evs {
		if e.SegmentID == segmentID && e.Side == side {
			if err := s.store.DeleteEvidence(oppID, segmentID, side); err != nil {
				return err
			}
			break
		}
	}
	cj := compare.ContourJSON{
		SegmentID: segmentID,
		Side:      side,
		ToneType:  seg.ToneType,
		Points:    curve,
	}
	norm, err := compare.MarshalContour(cj)
	if err != nil {
		return err
	}
	e := &model.Evidence{
		ID:           model.NewID("evi"),
		OppositionID: oppID,
		SegmentID:    segmentID,
		Side:         side,
		Normalized:   norm,
		ToneType:     seg.ToneType,
		CreatedAt:    now(),
	}
	return s.store.CreateEvidence(e)
}

// GetCluster 返回某对立的当前证据簇统计（不修改）。
func (s *Service) GetCluster(oppID string) (*compare.ClusterStats, error) {
	if _, err := s.store.GetOpposition(oppID); err != nil {
		return nil, err
	}
	evs, err := s.store.ListEvidence(oppID)
	if err != nil {
		return nil, err
	}
	return s.buildClusterFromEvidence(evs), nil
}

// RecomputeCluster 由对立全部证据重算并写回对立得分。
func (s *Service) RecomputeCluster(oppID string) (*compare.ClusterStats, error) {
	o, err := s.store.GetOpposition(oppID)
	if err != nil {
		return nil, err
	}
	if err := s.guardBatchMutable(o.BatchID); err != nil {
		return nil, err
	}
	evs, err := s.store.ListEvidence(oppID)
	if err != nil {
		return nil, err
	}
	stats := s.buildClusterFromEvidence(evs)
	o.OppositionScore = stats.Score
	if err := s.store.UpdateOpposition(o); err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *Service) buildClusterFromEvidence(evs []*model.Evidence) *compare.ClusterStats {
	contours := make([]compare.ContourJSON, 0, len(evs))
	for _, e := range evs {
		cj, err := compare.UnmarshalContour(e.Normalized)
		if err != nil {
			continue
		}
		contours = append(contours, cj)
	}
	return compare.BuildCluster(contours)
}

// Adjudicate 裁决声调对立：confirm/reject/insufficient。终态不可改。
func (s *Service) Adjudicate(oppID, decision, reason string) error {
	o, err := s.store.GetOpposition(oppID)
	if err != nil {
		return err
	}
	if err := s.guardBatchMutable(o.BatchID); err != nil {
		return err
	}
	target := decision
	switch decision {
	case model.OppConfirmed, model.OppRejected, model.OppInsufficient:
		target = decision
	default:
		return model.ErrBadRequest
	}
	if !adjudicate.CanTransition(o.Status, target) {
		return model.ErrInvalidState
	}
	if target == model.OppConfirmed {
		stats, err := s.GetCluster(oppID)
		if err != nil {
			return err
		}
		if adjudicate.SuggestStatus(stats.Score) != model.OppConfirmed {
			return model.ErrInvalidState
		}
	}
	o.Status = target
	o.DecisionReason = reason
	o.DecidedAt = now()
	return s.store.UpdateOpposition(o)
}

// CompareContours 直接对两组实时片段计算最小对立轮廓距离（快速比较，不落证据）。
func (s *Service) CompareContours(segIDsA, segIDsB []string) (*compare.ClusterStats, error) {
	if len(segIDsA) == 0 || len(segIDsB) == 0 {
		return nil, model.ErrBadRequest
	}
	contours := make([]compare.ContourJSON, 0, len(segIDsA)+len(segIDsB))
	for _, id := range segIDsA {
		c, err := s.contourForSegment(id, resampleN)
		if err != nil {
			return nil, err
		}
		contours = append(contours, compare.ContourJSON{SegmentID: id, Side: "a", Points: c})
	}
	for _, id := range segIDsB {
		c, err := s.contourForSegment(id, resampleN)
		if err != nil {
			return nil, err
		}
		contours = append(contours, compare.ContourJSON{SegmentID: id, Side: "b", Points: c})
	}
	return compare.BuildCluster(contours), nil
}
