package normalize

import (
	"math"
	"testing"

	"task252-tonereview/internal/model"
)

func TestSemitone(t *testing.T) {
	if got := Semitone(0); got != 0 {
		t.Fatalf("Semitone(0) = %v, want 0", got)
	}
	if Semitone(200) <= Semitone(100) {
		t.Fatalf("higher frequency must yield higher semitone value")
	}
}

func TestSpeakerBaseline(t *testing.T) {
	// 样本数不足 MinBaselineSamples 时应判定基线不可信。
	small := []model.F0Sample{}
	for i := 0; i < MinBaselineSamples-1; i++ {
		small = append(small, model.F0Sample{F0Hz: 200, Reliable: true})
	}
	if _, ok := SpeakerBaseline(small); ok {
		t.Fatalf("expected ok=false with %d samples", len(small))
	}
	// 样本达标应得到 log2 均值基线。
	big := []model.F0Sample{}
	for i := 0; i < MinBaselineSamples; i++ {
		big = append(big, model.F0Sample{F0Hz: 200, Reliable: true})
	}
	b, ok := SpeakerBaseline(big)
	if !ok {
		t.Fatalf("expected ok=true with %d samples", len(big))
	}
	if math.Abs(b-math.Log2(200)) > 1e-9 {
		t.Fatalf("baseline = %v, want log2(200)=%v", b, math.Log2(200))
	}
	// 不可靠样本须被丢弃。
	mixed := []model.F0Sample{
		{F0Hz: 100, Reliable: false},
		{F0Hz: 200, Reliable: true},
		{F0Hz: 200, Reliable: true},
		{F0Hz: 200, Reliable: true},
		{F0Hz: 200, Reliable: true},
		{F0Hz: 200, Reliable: true},
		{F0Hz: 200, Reliable: true},
	}
	if _, ok := SpeakerBaseline(mixed); !ok {
		t.Fatalf("expected ok=true when enough reliable samples present")
	}
}

func TestNormalizeContour(t *testing.T) {
	samples := []model.F0Sample{
		{TMs: 100, F0Hz: 200, Reliable: true},
		{TMs: 50, F0Hz: 100, Reliable: true}, // 时间更早，应被排序到前面
		{TMs: 150, F0Hz: 0, Reliable: false},  // 不可靠，丢弃
	}
	pts := NormalizeContour(samples, math.Log2(100))
	if len(pts) != 2 {
		t.Fatalf("expected 2 points, got %d", len(pts))
	}
	if pts[0].TAbs != 50 {
		t.Fatalf("expected sorted by time, first TAbs=50, got %v", pts[0].TAbs)
	}
	// 100Hz 相对基线 log2(100) 应得到 0 半音。
	if math.Abs(pts[0].ST) > 1e-9 {
		t.Fatalf("baseline sample should be 0 ST, got %v", pts[0].ST)
	}
	// 200Hz 相对 log2(100) 基线应为正半音。
	if pts[1].ST <= 0 {
		t.Fatalf("200Hz above baseline should be positive ST, got %v", pts[1].ST)
	}
}

func TestResample(t *testing.T) {
	// 空输入返回 nil。
	if got := Resample(nil, 10); got != nil {
		t.Fatalf("Resample(nil) should be nil, got %v", got)
	}
	// 单点输入返回以该点填充的平稳轮廓。
	single := []Point{{TAbs: 0, ST: 3.0}}
	out := Resample(single, 5)
	if len(out) != 5 {
		t.Fatalf("expected 5 points, got %d", len(out))
	}
	for _, v := range out {
		if v != 3.0 {
			t.Fatalf("single point resample should fill constant, got %v", v)
		}
	}
	// 线性插值检查。
	pts := []Point{{TAbs: 0, ST: 0}, {TAbs: 10, ST: 10}}
	out = Resample(pts, 3)
	if math.Abs(out[0]-0) > 1e-9 || math.Abs(out[2]-10) > 1e-9 {
		t.Fatalf("endpoints should be preserved: %v", out)
	}
	if math.Abs(out[1]-5) > 1e-9 {
		t.Fatalf("midpoint should be 5, got %v", out[1])
	}
}

func TestClassifyTone(t *testing.T) {
	// 平调：幅度低于阈值。
	if got := ClassifyTone([]float64{0, 0.1, 0, -0.1, 0}); got != model.ToneLevel {
		t.Fatalf("level contour -> %v, want level", got)
	}
	// 升调：整体上升且中点不凹陷。
	if got := ClassifyTone([]float64{-3, -1.5, 0, 1.5, 3}); got != model.ToneRising {
		t.Fatalf("rising contour -> %v, want rising", got)
	}
	// 降调：整体下降且中点不凸起。
	if got := ClassifyTone([]float64{3, 1.5, 0, -1.5, -3}); got != model.ToneFalling {
		t.Fatalf("falling contour -> %v, want falling", got)
	}
	// 样本不足返回 unknown。
	if got := ClassifyTone([]float64{1}); got != model.ToneUnknown {
		t.Fatalf("single point -> %v, want unknown", got)
	}
}
