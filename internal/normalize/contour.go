// Package normalize 实现声调证据复核的核心声学算法：说话人基线归一化、
// 基频到半音轮廓的转换、轮廓等距重采样与调型分类。全部为纯函数，便于单测。
package normalize

import (
	"math"
	"sort"

	"task252-tonereview/internal/model"
)

// MinBaselineSamples 计算说话人基线所需的最少可靠样本数。
const MinBaselineSamples = 6

// Semitone 将频率（赫兹）换算为半音绝对值（以 1Hz 为参考）。
// 仅用于相对比较，不依赖绝对音高。
func Semitone(hz float64) float64 {
	if hz <= 0 {
		return 0
	}
	return 12 * math.Log2(hz)
}

// SpeakerBaseline 由若干可用人声段的可靠 F0 样本计算说话人基线：
// 所有可靠 log2(F0) 的均值。返回基线与该基线是否可信（样本数达标）。
func SpeakerBaseline(samples []model.F0Sample) (baselineLog float64, ok bool) {
	sum := 0.0
	n := 0
	for _, s := range samples {
		if !s.Reliable || s.F0Hz <= 0 {
			continue
		}
		sum += math.Log2(s.F0Hz)
		n++
	}
	if n < MinBaselineSamples {
		return 0, false
	}
	return sum / float64(n), true
}

// Point 是归一化后的 (绝对时间, 相对基线半音) 采样点。
type Point struct {
	TAbs float64 // 原始毫秒时间
	ST   float64 // 相对说话人基线的半音偏移
}

// NormalizeContour 将原始 F0 样本转为相对基线的半音轮廓，并按时间排序，
// 丢弃不可靠或非正频率点。baselineLog 应为说话人基线（log2 均值）。
func NormalizeContour(samples []model.F0Sample, baselineLog float64) []Point {
	pts := make([]Point, 0, len(samples))
	for _, s := range samples {
		if !s.Reliable || s.F0Hz <= 0 {
			continue
		}
		pts = append(pts, Point{TAbs: float64(s.TMs), ST: 12 * (math.Log2(s.F0Hz) - baselineLog)})
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].TAbs < pts[j].TAbs })
	return pts
}

// Resample 将非均匀时间轮廓等距重采样为 n 个点（时间归一化 0..1，线性插值）。
// 样本数不足或时间跨度为零时返回以均值填充的平稳轮廓。
func Resample(points []Point, n int) []float64 {
	if n <= 0 {
		return nil
	}
	if len(points) == 0 {
		return nil
	}
	if len(points) == 1 {
		out := make([]float64, n)
		for i := range out {
			out[i] = points[0].ST
		}
		return out
	}
	tmin := points[0].TAbs
	tmax := points[len(points)-1].TAbs
	span := tmax - tmin
	if span <= 0 {
		sum := 0.0
		for _, p := range points {
			sum += p.ST
		}
		avg := sum / float64(len(points))
		out := make([]float64, n)
		for i := range out {
			out[i] = avg
		}
		return out
	}
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		target := tmin + (float64(i)/float64(n-1))*span
		out[i] = interp(points, target)
	}
	return out
}

// interp 在按时间排序的点集上线性插值。
func interp(points []Point, target float64) float64 {
	if target <= points[0].TAbs {
		return points[0].ST
	}
	if target >= points[len(points)-1].TAbs {
		return points[len(points)-1].ST
	}
	for i := 1; i < len(points); i++ {
		if target <= points[i].TAbs {
			p0, p1 := points[i-1], points[i]
			if p1.TAbs == p0.TAbs {
				return p0.ST
			}
			r := (target - p0.TAbs) / (p1.TAbs - p0.TAbs)
			return p0.ST + r*(p1.ST-p0.ST)
		}
	}
	return points[len(points)-1].ST
}

// ToneThreshST 调型分类的半音幅度阈值。
const ToneThreshST = 1.5

// ClassifyTone 基于重采样半音轮廓的斜率与中点偏离判定调型：
// 平(level)/升(rising)/降(falling)/曲折(contour)。
func ClassifyTone(c []float64) string {
	if len(c) < 2 {
		return model.ToneUnknown
	}
	first, last := c[0], c[len(c)-1]
	mid := c[len(c)/2]
	lo, hi := c[0], c[0]
	for _, v := range c {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	rng := hi - lo
	if rng < ToneThreshST {
		return model.ToneLevel
	}
	slope := last - first
	if slope > ToneThreshST {
		// 整体上升；若中点明显低于两端则为升曲折。
		if mid < first-0.5 && mid < last-0.5 {
			return model.ToneContour
		}
		return model.ToneRising
	}
	if slope < -ToneThreshST {
		if mid > first+0.5 && mid > last+0.5 {
			return model.ToneContour
		}
		return model.ToneFalling
	}
	return model.ToneContour
}
