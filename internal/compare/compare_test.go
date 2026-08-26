package compare

import (
	"math"
	"testing"
)

func TestContourDistance(t *testing.T) {
	a := []float64{0, 1, 2, 3}
	b := []float64{0, 1, 2, 3}
	mad, corr := ContourDistance(a, b)
	if math.Abs(mad) > 1e-9 {
		t.Fatalf("identical curves mad = %v, want 0", mad)
	}
	if math.Abs(corr-1) > 1e-9 {
		t.Fatalf("identical curves corr = %v, want 1", corr)
	}

	// 长度不等返回 MaxFloat64 与 0。
	mad, corr = ContourDistance(a, []float64{0, 1})
	if mad != math.MaxFloat64 || corr != 0 {
		t.Fatalf("mismatched length should be (MaxFloat64,0), got (%v,%v)", mad, corr)
	}

	// 反相曲线相关系数应接近 -1。
	c := []float64{3, 2, 1, 0}
	_, corr = ContourDistance(a, c)
	if corr > -0.99 {
		t.Fatalf("negatively correlated curves corr = %v, want ~ -1", corr)
	}
}

func TestBuildCluster(t *testing.T) {
	// A 侧两条高度相似的平调；B 侧两条相似升调 → 组间远大于组内，得分应 > 1。
	cases := []ContourJSON{
		{SegmentID: "a1", Side: "a", Points: []float64{0, 0, 0, 0}},
		{SegmentID: "a2", Side: "a", Points: []float64{0.1, -0.1, 0, 0.2}},
		{SegmentID: "b1", Side: "b", Points: []float64{-3, -1.5, 0, 1.5, 3}},
		{SegmentID: "b2", Side: "b", Points: []float64{-2.8, -1.4, 0.1, 1.6, 2.9}},
	}
	stats := BuildCluster(cases)
	if len(stats.SideA) != 2 || len(stats.SideB) != 2 {
		t.Fatalf("side counts wrong: A=%d B=%d", len(stats.SideA), len(stats.SideB))
	}
	if stats.Score <= 1.0 {
		t.Fatalf("expected high opposition score, got %v", stats.Score)
	}
	if stats.Between <= stats.AvgWithin {
		t.Fatalf("between (%v) should exceed avg_within (%v)", stats.Between, stats.AvgWithin)
	}

	// 任一侧为空 → 不可评分，得分为 0。
	oneSide := []ContourJSON{{SegmentID: "a1", Side: "a", Points: []float64{0, 0, 0}}}
	stats = BuildCluster(oneSide)
	if stats.Score != 0 {
		t.Fatalf("single-side cluster should score 0, got %v", stats.Score)
	}
}

func TestContourMarshalRoundtrip(t *testing.T) {
	c := ContourJSON{SegmentID: "s1", Side: "a", ToneType: "level", Points: []float64{1.5, 2.5, 3.5}}
	s, err := MarshalContour(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := UnmarshalContour(s)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.SegmentID != c.SegmentID || got.Side != c.Side || len(got.Points) != 3 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	if math.Abs(got.Points[0]-1.5) > 1e-9 {
		t.Fatalf("point not preserved: %v", got.Points)
	}
}
