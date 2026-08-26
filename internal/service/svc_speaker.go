package service

import (
	"task252-tonereview/internal/model"
)

// CreateSpeaker 新建说话人（尚未建立基线）。
func (s *Service) CreateSpeaker(code, dialect, gender string, birthYear int) (*model.Speaker, error) {
	if code == "" {
		code = model.NewID("spk")
	}
	sp := &model.Speaker{
		ID:          model.NewID("spk"),
		Code:        code,
		Dialect:     dialect,
		Gender:      gender,
		BirthYear:   birthYear,
		HasBaseline: false,
		CreatedAt:   now(),
	}
	if err := s.store.CreateSpeaker(sp); err != nil {
		return nil, err
	}
	return sp, nil
}

// GetSpeaker 读取说话人。
func (s *Service) GetSpeaker(id string) (*model.Speaker, error) {
	return s.store.GetSpeaker(id)
}

// ListSpeakers 列出全部说话人。
func (s *Service) ListSpeakers() ([]*model.Speaker, error) {
	return s.store.ListSpeakers()
}

// RecomputeBaseline 由说话人全部可用片段的可靠基频重算基线，
// 并回填空基线片段的调型分类。
func (s *Service) RecomputeBaseline(speakerID string) error {
	sp, err := s.store.GetSpeaker(speakerID)
	if err != nil {
		return err
	}
	segIDs, err := s.store.ListUsableSegmentsBySpeaker(speakerID)
	if err != nil {
		return err
	}
	var all []model.F0Sample
	for _, sid := range segIDs {
		samples, err := s.store.ListF0(sid)
		if err != nil {
			return err
		}
		all = append(all, samples...)
	}
	baseline, ok := speakerBaselineOf(all)
	if !ok {
		// 样本不足：清除基线，相关片段调型置 unknown。
		if err := s.store.SetBaseline(speakerID, 0, false, now()); err != nil {
			return err
		}
		for _, sid := range segIDs {
			if err := s.store.SetSegmentToneType(sid, model.ToneUnknown); err != nil {
				return err
			}
		}
		return nil
	}
	if err := s.store.SetBaseline(speakerID, baseline, true, now()); err != nil {
		return err
	}
	sp.BaselineLog = baseline
	sp.HasBaseline = true
	// 回填空基线片段调型。
	for _, sid := range segIDs {
		samples, err := s.store.ListF0(sid)
		if err != nil {
			return err
		}
		tt := toneTypeOfSamples(samples, sp.HasBaseline, sp.BaselineLog)
		if err := s.store.SetSegmentToneType(sid, tt); err != nil {
			return err
		}
	}
	return nil
}
