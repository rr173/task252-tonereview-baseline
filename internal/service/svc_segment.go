package service

import (
	"fmt"

	"task252-tonereview/internal/capture"
	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

// ImportSegment 导入发音片段及其基频采样（待校验）。audio_fp 幂等：若同一指纹
// 已存在则直接返回既有片段，不重复插入。
func (s *Service) ImportSegment(batchID string, in capture.SegmentInput) (*model.Segment, error) {
	if err := s.guardBatchMutable(batchID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetBatch(batchID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetSpeaker(in.SpeakerID); err != nil {
		return nil, err
	}
	if err := in.Validate(); err != nil {
		return nil, err
	}
	// 幂等：同一录音指纹视为重复导入。
	if existing, err := s.store.GetSegmentByFP(in.AudioFP); err == nil && existing != nil {
		return existing, nil
	}
	seg := &model.Segment{
		ID:          model.NewID("seg"),
		BatchID:     batchID,
		LexicalItem: in.LexicalItem,
		PhoneticSeg: in.PhoneticSeg,
		SpeakerID:   in.SpeakerID,
		AudioFP:     in.AudioFP,
		Status:      model.SegPending,
		DurationMs:  in.DurationMs,
		RecordedAt:  in.RecordedAt,
		ToneType:    model.ToneUnknown,
		CreatedAt:   now(),
	}
	if err := s.store.CreateSegment(seg); err != nil {
		return nil, err
	}
	samples := in.ToF0Samples(seg.ID)
	if err := s.store.InsertF0(samples); err != nil {
		return nil, err
	}
	// 若说话人已有基线，立即分类调型。
	sp, err := s.store.GetSpeaker(in.SpeakerID)
	if err == nil && sp.HasBaseline {
		tt := toneTypeOfSamples(samples, sp.HasBaseline, sp.BaselineLog)
		_ = s.store.SetSegmentToneType(seg.ID, tt)
		seg.ToneType = tt
	}
	return seg, nil
}

// GetSegment 读取片段。
func (s *Service) GetSegment(id string) (*model.Segment, error) {
	return s.store.GetSegment(id)
}

// ListSegments 按过滤条件列出片段。
func (s *Service) ListSegments(f store.SegmentFilter) ([]*model.Segment, error) {
	return s.store.ListSegments(f)
}

// VerifySegment 待校验 → 可用。
func (s *Service) VerifySegment(id string) error {
	return s.transitionSegment(id, model.SegUsable)
}

// MarkNoise 待校验 → 噪声。
func (s *Service) MarkNoise(id string) error {
	return s.transitionSegment(id, model.SegNoise)
}

// ExcludeSegment 待校验 → 排除（研究者否决异常录音）。
func (s *Service) ExcludeSegment(id string) error {
	return s.transitionSegment(id, model.SegExcluded)
}

// RestoreSegment 排除 → 可用（误否决后恢复）。
func (s *Service) RestoreSegment(id string) error {
	return s.transitionSegment(id, model.SegUsable)
}

func (s *Service) transitionSegment(id, to string) error {
	seg, err := s.store.GetSegment(id)
	if err != nil {
		return err
	}
	if err := s.guardBatchMutable(seg.BatchID); err != nil {
		return err
	}
	allowed := map[string][]string{
		model.SegPending:  {model.SegUsable, model.SegNoise, model.SegExcluded},
		model.SegExcluded: {model.SegUsable},
	}
	ok := false
	for _, a := range allowed[seg.Status] {
		if a == to {
			ok = true
			break
		}
	}
	if !ok {
		return model.ErrInvalidState
	}
	return s.store.UpdateSegmentStatus(id, to)
}

// AddF0 向片段追加基频采样（时间须晚于既有最大时间，保证单调）。
func (s *Service) AddF0(segmentID string, points []capture.F0Point) error {
	seg, err := s.store.GetSegment(segmentID)
	if err != nil {
		return err
	}
	if err := s.guardBatchMutable(seg.BatchID); err != nil {
		return err
	}
	if len(points) == 0 {
		return model.ErrBadRequest
	}
	maxT, err := s.store.MaxF0Time(segmentID)
	if err != nil {
		return err
	}
	samples := make([]model.F0Sample, 0, len(points))
	prev := maxT
	for _, p := range points {
		if p.TMs <= prev {
			return fmt.Errorf("%w: appended f0 time %d not after %d", model.ErrBadRequest, p.TMs, prev)
		}
		if p.F0Hz <= 0 {
			return fmt.Errorf("%w: non-positive f0", model.ErrBadRequest)
		}
		prev = p.TMs
		samples = append(samples, model.F0Sample{
			ID: model.NewID("f0"), SegmentID: segmentID, TMs: p.TMs, F0Hz: p.F0Hz, Reliable: true,
		})
	}
	if err := s.store.InsertF0(samples); err != nil {
		return err
	}
	// 若说话人有基线，重算该片段调型。
	sp, err := s.store.GetSpeaker(seg.SpeakerID)
	if err == nil && sp.HasBaseline {
		all, err := s.store.ListF0(segmentID)
		if err == nil {
			tt := toneTypeOfSamples(all, sp.HasBaseline, sp.BaselineLog)
			_ = s.store.SetSegmentToneType(segmentID, tt)
		}
	}
	return nil
}

// ListF0 读取片段基频。
func (s *Service) ListF0(segmentID string) ([]model.F0Sample, error) {
	if _, err := s.store.GetSegment(segmentID); err != nil {
		return nil, err
	}
	return s.store.ListF0(segmentID)
}
