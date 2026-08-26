// Package version 实现分析版本的快照构建与状态流转守卫。
package version

import (
	"encoding/json"

	"task252-tonereview/internal/model"
)

// OppSnapshot 是冻结快照中单个声调对立的精简记录。
type OppSnapshot struct {
	ID     string  `json:"id"`
	LexA   string  `json:"lexical_a"`
	LexB   string  `json:"lexical_b"`
	Status string  `json:"status"`
	Score  float64 `json:"score"`
}

// SegmentSnapshot 是冻结快照中片段的精简记录。
type SegmentSnapshot struct {
	ID          string `json:"id"`
	LexicalItem string `json:"lexical_item"`
	PhoneticSeg string `json:"phonetic_seg"`
	SpeakerID   string `json:"speaker_id"`
	Status      string `json:"status"`
	ToneType    string `json:"tone_type"`
}

// Snapshot 是分析版本冻结时写出的不可变证据集合。
type Snapshot struct {
	BatchID    string            `json:"batch_id"`
	SegmentN   int               `json:"segment_n"`
	Segments   []SegmentSnapshot `json:"segments"`
	Oppositions []OppSnapshot    `json:"oppositions"`
}

// BuildSnapshot 由批次当前片段与对立构建快照 JSON。
func BuildSnapshot(batchID string, segments []*model.Segment, oppositions []*model.ToneOpposition) (string, error) {
	snap := Snapshot{BatchID: batchID, SegmentN: len(segments)}
	for _, s := range segments {
		snap.Segments = append(snap.Segments, SegmentSnapshot{
			ID: s.ID, LexicalItem: s.LexicalItem, PhoneticSeg: s.PhoneticSeg,
			SpeakerID: s.SpeakerID, Status: s.Status, ToneType: s.ToneType,
		})
	}
	for _, o := range oppositions {
		snap.Oppositions = append(snap.Oppositions, OppSnapshot{
			ID: o.ID, LexA: o.LexicalA, LexB: o.LexicalB, Status: o.Status, Score: o.OppositionScore,
		})
	}
	b, err := json.Marshal(snap)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
