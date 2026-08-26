// Package compare 实现声调轮廓比较算法：成对轮廓距离（MAD + 相关系数）、
// 最小对立证据簇聚合与对立得分。
package compare

import "math"

// ContourDistance 在两条等长重采样轮廓上计算：
//   - mad：平均绝对偏差（半音），越小越相似；
//   - corr：皮尔逊相关系数，越大形状越一致。
// 长度不等或为空时返回 (MaxFloat64, 0)。
func ContourDistance(a, b []float64) (mad float64, corr float64) {
	n := len(a)
	if n == 0 || len(b) != n {
		return math.MaxFloat64, 0
	}
	sumAbs := 0.0
	for i := 0; i < n; i++ {
		sumAbs += math.Abs(a[i] - b[i])
	}
	mad = sumAbs / float64(n)

	ma, mb := mean(a), mean(b)
	var cov, va, vb float64
	for i := 0; i < n; i++ {
		da, db := a[i]-ma, b[i]-mb
		cov += da * db
		va += da * da
		vb += db * db
	}
	if va == 0 || vb == 0 {
		corr = 0
	} else {
		corr = cov / math.Sqrt(va*vb)
	}
	return mad, corr
}

// mean 计算切片均值，空切片返回 0。
func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

// pairwiseAvg 计算一组等长轮廓的 pairwise 平均 MAD。
func pairwiseAvg(curves [][]float64) float64 {
	total := 0.0
	pairs := 0
	for i := 0; i < len(curves); i++ {
		for j := i + 1; j < len(curves); j++ {
			mad, _ := ContourDistance(curves[i], curves[j])
			if mad == math.MaxFloat64 {
				continue
			}
			total += mad
			pairs++
		}
	}
	if pairs == 0 {
		return 0
	}
	return total / float64(pairs)
}

// crossAvg 计算两组等长轮廓之间的平均 MAD。
func crossAvg(a, b [][]float64) float64 {
	total := 0.0
	n := 0
	for _, ca := range a {
		for _, cb := range b {
			mad, _ := ContourDistance(ca, cb)
			if mad == math.MaxFloat64 {
				continue
			}
			total += mad
			n++
		}
	}
	if n == 0 {
		return math.MaxFloat64
	}
	return total / float64(n)
}
