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
//
// 说话人基线为全局单值，一旦说话人的片段进入已封存批次，
// 其基线与片段调型即成为已冻结复核证据的组成部分，重算会改写这些证据。
// 因此封存批次关联的基线重算必须被拒绝，现有数据保持不变。
func (s *Service) RecomputeBaseline(speakerID string) error {
	sp, err := s.store.GetSpeaker(speakerID)
	if err != nil {
		return err
	}
	// 封存守卫：说话人若参与任何已封存批次，重算即会改写冻结证据，拒绝之。
	if sealed, err := s.store.SpeakerInSealedBatch(speakerID); err != nil {
		return err
	} else if sealed {
		return model.ErrSealed
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
