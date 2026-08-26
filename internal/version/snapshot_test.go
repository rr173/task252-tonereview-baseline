package version

import (
	"encoding/json"
	"testing"

	"task252-tonereview/internal/model"
)

func TestBuildSnapshot(t *testing.T) {
	segs := []*model.Segment{
		{ID: "s1", LexicalItem: "ma-high", PhoneticSeg: "ma", SpeakerID: "spk-A", Status: "usable", ToneType: "level"},
		{ID: "s2", LexicalItem: "ma-rising", PhoneticSeg: "ma", SpeakerID: "spk-B", Status: "usable", ToneType: "rising"},
	}
	opps := []*model.ToneOpposition{
		{ID: "o1", LexicalA: "ma-high", LexicalB: "ma-rising", Status: "confirmed", OppositionScore: 3.9},
	}
	snap, err := BuildSnapshot("b1", segs, opps)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}
	var s Snapshot
	if err := json.Unmarshal([]byte(snap), &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if s.BatchID != "b1" || s.SegmentN != 2 {
		t.Fatalf("snapshot header wrong: %+v", s)
	}
	if len(s.Segments) != 2 || len(s.Oppositions) != 1 {
		t.Fatalf("snapshot body wrong: segs=%d opps=%d", len(s.Segments), len(s.Oppositions))
	}
	if s.Segments[0].ID != "s1" || s.Oppositions[0].Score != 3.9 {
		t.Fatalf("snapshot content wrong: %+v", s)
	}
}
