package store

import (
	"testing"
	"time"

	"task252-tonereview/internal/model"
)

func TestOpenAndCRUD(t *testing.T) {
	st, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()

	// 说话人读写往返。
	sp := &model.Speaker{ID: "spk-1", Code: "A", Dialect: "X", Gender: "f", BirthYear: 1950, HasBaseline: true, BaselineLog: 7.6, CreatedAt: time.Now().UnixMilli()}
	if err := st.CreateSpeaker(sp); err != nil {
		t.Fatalf("CreateSpeaker: %v", err)
	}
	got, err := st.GetSpeaker("spk-1")
	if err != nil {
		t.Fatalf("GetSpeaker: %v", err)
	}
	if got.Code != "A" || !got.HasBaseline || got.BaselineLog != 7.6 {
		t.Fatalf("speaker roundtrip wrong: %+v", got)
	}

	// 批次读写往返。
	b := &model.FieldBatch{ID: "batch-1", Code: "B1", Title: "t", Status: model.BatchCollecting, CreatedAt: 1, UpdatedAt: 1}
	if err := st.CreateBatch(b); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}
	gb, err := st.GetBatch("batch-1")
	if err != nil {
		t.Fatalf("GetBatch: %v", err)
	}
	if gb.Status != model.BatchCollecting {
		t.Fatalf("batch status wrong: %v", gb.Status)
	}

	// 片段 + 基频写入与查询。
	seg := &model.Segment{ID: "seg-1", BatchID: "batch-1", LexicalItem: "ma-high", PhoneticSeg: "ma", SpeakerID: "spk-1", AudioFP: "fp1", Status: model.SegUsable, ToneType: "level", CreatedAt: 1}
	if err := st.CreateSegment(seg); err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}
	gs, err := st.GetSegment("seg-1")
	if err != nil {
		t.Fatalf("GetSegment: %v", err)
	}
	if gs.LexicalItem != "ma-high" {
		t.Fatalf("segment wrong: %+v", gs)
	}
	f0 := []model.F0Sample{
		{ID: "f1", SegmentID: "seg-1", TMs: 0, F0Hz: 200, Reliable: true},
		{ID: "f2", SegmentID: "seg-1", TMs: 50, F0Hz: 210, Reliable: true},
	}
	if err := st.InsertF0(f0); err != nil {
		t.Fatalf("InsertF0: %v", err)
	}
	gotF0, err := st.ListF0("seg-1")
	if err != nil {
		t.Fatalf("ListF0: %v", err)
	}
	if len(gotF0) != 2 {
		t.Fatalf("expected 2 f0 samples, got %d", len(gotF0))
	}

	// 按批次过滤片段。
	segs, err := st.ListSegments(SegmentFilter{BatchID: "batch-1"})
	if err != nil {
		t.Fatalf("ListSegments: %v", err)
	}
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}

	// 对立写入与列表。
	o := &model.ToneOpposition{ID: "opp-1", BatchID: "batch-1", LexicalA: "ma-high", PhoneticSeg: "ma", LexicalB: "ma-rising", Status: model.OppCandidate, CreatedAt: 1}
	if err := st.CreateOpposition(o); err != nil {
		t.Fatalf("CreateOpposition: %v", err)
	}
	opps, err := st.ListOppositions("batch-1", "")
	if err != nil {
		t.Fatalf("ListOppositions: %v", err)
	}
	if len(opps) != 1 {
		t.Fatalf("expected 1 opposition, got %d", len(opps))
	}

	// 版本号递增与列表。
	nv, err := st.NextVersion("batch-1")
	if err != nil {
		t.Fatalf("NextVersion: %v", err)
	}
	if nv != 1 {
		t.Fatalf("first version should be 1, got %d", nv)
	}
	v := &model.AnalysisVersion{ID: "ver-1", BatchID: "batch-1", Version: 1, Status: model.VerDraft, CreatedAt: 1}
	if err := st.CreateVersion(v); err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}
	vers, err := st.ListVersions("batch-1")
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}
	if len(vers) != 1 || vers[0].Version != 1 {
		t.Fatalf("versions wrong: %+v", vers)
	}
}

func TestGetMissingReturnsNotFound(t *testing.T) {
	st, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()
	if _, err := st.GetBatch("nope"); err != model.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := st.GetSpeaker("nope"); err != model.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
