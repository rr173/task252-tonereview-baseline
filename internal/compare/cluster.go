package compare

import "encoding/json"

// ResamplePoints 重采样固定点数（与归一化/分类保持一致）。
const ResamplePoints = 24

// ContourJSON 是证据持久化与重算的统一轮廓形态。
type ContourJSON struct {
	SegmentID string    `json:"segment_id"`
	Side      string    `json:"side"` // "a" | "b"
	ToneType  string    `json:"tone_type"`
	Points    []float64 `json:"points"`
}

// ClusterStats 汇总一个候选声调对立两侧的轮廓证据与对立得分。
type ClusterStats struct {
	SideA     []string  `json:"side_a"`     // 侧 A 证据段 ID
	SideB     []string  `json:"side_b"`     // 侧 B 证据段 ID
	WithinA   float64   `json:"within_a"`   // A 组内平均 pairwise 距离
	WithinB   float64   `json:"within_b"`   // B 组内平均 pairwise 距离
	Between   float64   `json:"between"`    // A-B 组间平均距离
	AvgWithin float64   `json:"avg_within"` // (withinA+withinB)/2
	Score     float64   `json:"score"`      // between / (avg_within + eps)
	Evidences int       `json:"evidences"`
}

// BuildCluster 由某对立的全部证据（含两侧归一化轮廓）聚合统计。
// 任一侧不足 1 条证据时返回零值统计（不可评分）。
func BuildCluster(contours []ContourJSON) *ClusterStats {
	stats := &ClusterStats{}
	var aCurves, bCurves [][]float64
	for _, c := range contours {
		if len(c.Points) == 0 {
			continue
		}
		switch c.Side {
		case "a":
			stats.SideA = append(stats.SideA, c.SegmentID)
			aCurves = append(aCurves, c.Points)
		case "b":
			stats.SideB = append(stats.SideB, c.SegmentID)
			bCurves = append(bCurves, c.Points)
		}
	}
	stats.Evidences = len(contours)
	if len(aCurves) == 0 || len(bCurves) == 0 {
		stats.Score = 0
		return stats
	}
	stats.WithinA = pairwiseAvg(aCurves)
	stats.WithinB = pairwiseAvg(bCurves)
	stats.AvgWithin = (stats.WithinA + stats.WithinB) / 2.0
	stats.Between = crossAvg(aCurves, bCurves)
	const eps = 0.25 // 防止组内零方差时除零爆炸
	if stats.Between == 0 && stats.AvgWithin == 0 {
		stats.Score = 0
		return stats
	}
	stats.Score = stats.Between / (stats.AvgWithin + eps)
	return stats
}

// MarshalContour 序列化轮廓为 JSON 字符串（供证据持久化）。
func MarshalContour(c ContourJSON) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// UnmarshalContour 从 JSON 反序列化轮廓。
func UnmarshalContour(s string) (ContourJSON, error) {
	var c ContourJSON
	if err := json.Unmarshal([]byte(s), &c); err != nil {
		return ContourJSON{}, err
	}
	return c, nil
}
