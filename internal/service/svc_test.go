package service

import (
	"testing"

	"task252-tonereview/internal/capture"
	"task252-tonereview/internal/store"
)

// segInput 构造一条发音片段导入载荷，基频在 [start,end] 间线性插值。
func segInput(speaker, fp string, f0Start, f0End float64) capture.SegmentInput {
	return segInputFor("ma", speaker, fp, f0Start, f0End)
}

func segInputFor(lexical, speaker, fp string, f0Start, f0End float64) capture.SegmentInput {
	pts := make([]capture.F0Point, 8)
	for i := 0; i < 8; i++ {
		pts[i] = capture.F0Point{TMs: int64(i) * 50, F0Hz: f0Start + (f0End-f0Start)*float64(i)/7.0}
	}
	return capture.SegmentInput{
		LexicalItem: lexical, PhoneticSeg: "ma", SpeakerID: speaker, AudioFP: fp,
		DurationMs: 400, RecordedAt: 1, F0: pts,
	}
}

func TestServiceFullLoop(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()
	svc := New(st)

	spA, err := svc.CreateSpeaker("spk-A", "dialect-X", "f", 1950)
	if err != nil {
		t.Fatalf("speaker A: %v", err)
	}
	spB, err := svc.CreateSpeaker("spk-B", "dialect-X", "m", 1948)
	if err != nil {
		t.Fatalf("speaker B: %v", err)
	}

	batch, err := svc.CreateBatch("smoke-batch", "高低平/升对立")
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	bid := batch.ID

	segA1, err := svc.ImportSegment(bid, segInputFor("ma-high", spA.ID, "fp-a1", 200, 200))
	if err != nil {
		t.Fatalf("import a1: %v", err)
	}
	segA2, err := svc.ImportSegment(bid, segInputFor("ma-high", spA.ID, "fp-a2", 202, 198))
	if err != nil {
		t.Fatalf("import a2: %v", err)
	}
	segB1, err := svc.ImportSegment(bid, segInputFor("ma-rising", spB.ID, "fp-b1", 150, 220))
	if err != nil {
		t.Fatalf("import b1: %v", err)
	}
	segB2, err := svc.ImportSegment(bid, segInputFor("ma-rising", spB.ID, "fp-b2", 155, 215))
	if err != nil {
		t.Fatalf("import b2: %v", err)
	}

	for _, id := range []string{segA1.ID, segA2.ID, segB1.ID, segB2.ID} {
		if err := svc.VerifySegment(id); err != nil {
			t.Fatalf("verify %s: %v", id, err)
		}
	}
	if err := svc.RecomputeBaseline(spA.ID); err != nil {
		t.Fatalf("baseline A: %v", err)
	}
	if err := svc.RecomputeBaseline(spB.ID); err != nil {
		t.Fatalf("baseline B: %v", err)
	}

	// 说话人基线建立后应回填片段调型（非 unknown）。
	ga, _ := svc.GetSegment(segA1.ID)
	if ga.ToneType == "" || ga.ToneType == "unknown" {
		t.Fatalf("expected classified tone type, got %q", ga.ToneType)
	}

	opp, err := svc.CreateOpposition(bid, "ma-high", "ma", "ma-rising")
	if err != nil {
		t.Fatalf("opp: %v", err)
	}
	for _, e := range []struct {
		seg, side string
	}{
		{segA1.ID, "a"}, {segA2.ID, "a"}, {segB1.ID, "b"}, {segB2.ID, "b"},
	} {
		if err := svc.AddEvidence(opp.ID, e.seg, e.side); err != nil {
			t.Fatalf("evidence %s/%s: %v", e.seg, e.side, err)
		}
	}
	stats, err := svc.RecomputeCluster(opp.ID)
	if err != nil {
		t.Fatalf("cluster: %v", err)
	}
	if stats.Score < 1.0 {
		t.Fatalf("expected score >= 1, got %v", stats.Score)
	}
	if err := svc.Adjudicate(opp.ID, "confirmed", "组间差异显著"); err != nil {
		t.Fatalf("adjudicate: %v", err)
	}

	if err := svc.StartReview(bid); err != nil {
		t.Fatalf("review: %v", err)
	}
	if err := svc.PublishBatch(bid); err != nil {
		t.Fatalf("publish: %v", err)
	}

	ver, err := svc.CreateVersion(bid, "首版")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if err := svc.ShareVersion(ver.ID); err != nil {
		t.Fatalf("share: %v", err)
	}
	if err := svc.FreezeVersion(ver.ID); err != nil {
		t.Fatalf("freeze: %v", err)
	}
	if err := svc.SealBatch(bid); err != nil {
		t.Fatalf("seal: %v", err)
	}

	// 终态断言。
	gb, _ := svc.GetBatch(bid)
	if gb.Status != "sealed" {
		t.Fatalf("batch should be sealed, got %s", gb.Status)
	}
	go2, _ := svc.GetOpposition(opp.ID)
	if go2.Status != "confirmed" {
		t.Fatalf("opp should be confirmed, got %s", go2.Status)
	}
	vers, _ := svc.ListVersions(bid)
	if len(vers) != 1 || vers[0].Status != "frozen" {
		t.Fatalf("version not frozen: %+v", vers)
	}
	segs, _ := svc.ListSegments(store.SegmentFilter{BatchID: bid})
	if len(segs) != 4 {
		t.Fatalf("expected 4 segs, got %d", len(segs))
	}

	// 封存批次应拒绝后续修改。
	if _, err := svc.ImportSegment(bid, segInput(spA.ID, "fp-x", 200, 200)); err == nil {
		t.Fatalf("expected sealed batch to reject import")
	}
}

func TestSegmentStateGuard(t *testing.T) {
	st, _ := store.Open(":memory:")
	defer st.Close()
	svc := New(st)
	sp, _ := svc.CreateSpeaker("sp", "d", "f", 1950)
	batch, _ := svc.CreateBatch("b", "t")
	seg, _ := svc.ImportSegment(batch.ID, segInput(sp.ID, "fp1", 200, 200))
	// pending → excluded 合法。
	if err := svc.ExcludeSegment(seg.ID); err != nil {
		t.Fatalf("exclude: %v", err)
	}
	// excluded → usable 合法（误否决后恢复）。
	if err := svc.RestoreSegment(seg.ID); err != nil {
		t.Fatalf("restore: %v", err)
	}
	// usable 态不能再被排除（仅 pending 可转 excluded）。
	if err := svc.ExcludeSegment(seg.ID); err == nil {
		t.Fatalf("expected invalid state excluding a usable segment")
	}
}

func TestOppositionRejectsEqualLexemes(t *testing.T) {
	st, _ := store.Open(":memory:")
	defer st.Close()
	svc := New(st)
	batch, _ := svc.CreateBatch("b", "t")
	if _, err := svc.CreateOpposition(batch.ID, "ma", "ma", "ma"); err == nil {
		t.Fatalf("expected rejection of identical lexemes")
	}
}
