// Package service 编排领域包与持久化，对外暴露业务闭环能力。
// 所有跨实体一致性（批次封存守卫、证据簇重算、版本快照）在此实现。
package service

import (
	"time"

	"task252-tonereview/internal/compare"
	"task252-tonereview/internal/model"
	"task252-tonereview/internal/normalize"
	"task252-tonereview/internal/store"
)

// Service 封装存储与业务编排。
type Service struct {
	store *store.Store
}

// New 构造服务。
func New(s *store.Store) *Service { return &Service{store: s} }

func now() int64 { return time.Now().UnixMilli() }

// Store 暴露底层存储（测试/事务用）。
func (s *Service) Store() *store.Store { return s.store }

// contourForSegment 加载片段基频并按其说话人基线归一化、重采样为固定点轮廓。
func (s *Service) contourForSegment(segmentID string, n int) ([]float64, error) {
	seg, err := s.store.GetSegment(segmentID)
	if err != nil {
		return nil, err
	}
	if seg.Status != model.SegUsable {
		return nil, model.ErrInvalidState
	}
	sp, err := s.store.GetSpeaker(seg.SpeakerID)
	if err != nil {
		return nil, err
	}
	if !sp.HasBaseline {
		return nil, model.ErrBadRequest
	}
	samples, err := s.store.ListF0(segmentID)
	if err != nil {
		return nil, err
	}
	pts := normalize.NormalizeContour(samples, sp.BaselineLog)
	if len(pts) == 0 {
		return nil, model.ErrBadRequest
	}
	return normalize.Resample(pts, n), nil
}

// SelfCheck 返回服务健康与关键计数（自检端点使用）。
func (s *Service) SelfCheck() (map[string]any, error) {
	batches, err := s.store.ListBatches()
	if err != nil {
		return nil, err
	}
	speakers, err := s.store.ListSpeakers()
	if err != nil {
		return nil, err
	}
	segs, err := s.store.ListSegments(store.SegmentFilter{})
	if err != nil {
		return nil, err
	}
	opps, err := s.store.ListOppositions("", "")
	if err != nil {
		return nil, err
	}
	totalSegs := len(segs)
	usable := 0
	for _, sg := range segs {
		if sg.Status == model.SegUsable {
			usable++
		}
	}
	confirmed := 0
	for _, o := range opps {
		if o.Status == model.OppConfirmed {
			confirmed++
		}
	}
	return map[string]any{
		"ok":               true,
		"batch_count":      len(batches),
		"speaker_count":    len(speakers),
		"segment_count":    totalSegs,
		"usable_segments":  usable,
		"opposition_count": len(opps),
		"confirmed_opps":   confirmed,
		"timestamp":        now(),
	}, nil
}

// Stats 返回汇总统计（stats 端点使用）。
func (s *Service) Stats() (map[string]any, error) {
	allOpps, err := s.store.ListOppositions("", "")
	if err != nil {
		return nil, err
	}
	sumScore := 0.0
	for _, o := range allOpps {
		sumScore += o.OppositionScore
	}
	avg := 0.0
	if len(allOpps) > 0 {
		avg = sumScore / float64(len(allOpps))
	}
	sc, err := s.SelfCheck()
	if err != nil {
		return nil, err
	}
	sc["avg_opposition_score"] = avg
	return sc, nil
}

// resampleN 证据重采样点数（与 compare 包保持一致）。
const resampleN = compare.ResamplePoints
