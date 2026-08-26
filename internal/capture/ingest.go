// Package capture 实现发音片段（采集模块）的导入校验与基频摄入，
// 并把片段归类到待校验状态。时间单调性、字段完整性在此把关。
package capture

import (
	"fmt"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/normalize"
)

// F0Point 是导入请求中的单个基频采样（时间毫秒，频率赫兹）。
type F0Point struct {
	TMs  int64   `json:"t_ms"`
	F0Hz float64 `json:"f0_hz"`
}

// SegmentInput 是导入发音片段的请求载荷（含基频采样）。
type SegmentInput struct {
	LexicalItem string    `json:"lexical_item"`
	PhoneticSeg string    `json:"phonetic_seg"`
	SpeakerID   string    `json:"speaker_id"`
	AudioFP     string    `json:"audio_fp"`
	DurationMs  int64     `json:"duration_ms"`
	RecordedAt  int64     `json:"recorded_at"`
	F0          []F0Point `json:"f0"`
}

// Validate 校验导入载荷语义：必填字段、采样时间严格递增、频率为正。
func (in SegmentInput) Validate() error {
	if in.LexicalItem == "" {
		return fmt.Errorf("%w: empty lexical_item", model.ErrBadRequest)
	}
	if in.PhoneticSeg == "" {
		return fmt.Errorf("%w: empty phonetic_seg", model.ErrBadRequest)
	}
	if in.SpeakerID == "" {
		return fmt.Errorf("%w: empty speaker_id", model.ErrBadRequest)
	}
	if in.AudioFP == "" {
		return fmt.Errorf("%w: empty audio_fp", model.ErrBadRequest)
	}
	if len(in.F0) == 0 {
		return fmt.Errorf("%w: no f0 samples", model.ErrBadRequest)
	}
	prev := int64(-1)
	for _, p := range in.F0 {
		if p.TMs < 0 {
			return fmt.Errorf("%w: negative sample time", model.ErrBadRequest)
		}
		if p.TMs <= prev {
			return fmt.Errorf("%w: f0 sample time not strictly increasing", model.ErrBadRequest)
		}
		prev = p.TMs
		if p.F0Hz <= 0 {
			return fmt.Errorf("%w: non-positive f0 at t=%d", model.ErrBadRequest, p.TMs)
		}
	}
	return nil
}

// ToF0Samples 将导入载荷转为持久化 F0 样本（全部标为可靠）。
func (in SegmentInput) ToF0Samples(segmentID string) []model.F0Sample {
	out := make([]model.F0Sample, 0, len(in.F0))
	for _, p := range in.F0 {
		out = append(out, model.F0Sample{
			ID:        model.NewID("f0"),
			SegmentID: segmentID,
			TMs:       p.TMs,
			F0Hz:      p.F0Hz,
			Reliable:  true,
		})
	}
	return out
}

// ToneTypeOf 在已知说话人基线时给出该片段的调型分类（无基线返回 unknown）。
func ToneTypeOf(samples []model.F0Sample, hasBaseline bool, baselineLog float64) string {
	if !hasBaseline {
		return model.ToneUnknown
	}
	pts := normalize.NormalizeContour(samples, baselineLog)
	if len(pts) == 0 {
		return model.ToneUnknown
	}
	resampled := normalize.Resample(pts, normalize.MinBaselineSamples*2)
	return normalize.ClassifyTone(resampled)
}
