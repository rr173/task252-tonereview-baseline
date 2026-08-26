package service

import (
	"errors"
	"testing"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

func TestBug02EvidenceRejectsForeignBatchSegment(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	sp, _ := svc.CreateSpeaker("sp", "d", "f", 1950)
	b1, _ := svc.CreateBatch("b1", "first")
	b2, _ := svc.CreateBatch("b2", "second")
	seg1, _ := svc.ImportSegment(b1.ID, segInputFor("ma-high", sp.ID, "fp1", 200, 200))
	seg2, _ := svc.ImportSegment(b2.ID, segInputFor("ma-high", sp.ID, "fp2", 202, 198))
	for _, id := range []string{seg1.ID, seg2.ID} { if err := svc.VerifySegment(id); err != nil { t.Fatal(err) } }
	if err := svc.RecomputeBaseline(sp.ID); err != nil { t.Fatal(err) }
	opp, _ := svc.CreateOpposition(b1.ID, "ma-high", "ma", "ma-rising")
	if err := svc.AddEvidence(opp.ID, seg2.ID, "a"); !errors.Is(err, model.ErrBadRequest) {
		t.Fatalf("foreign batch evidence error = %v, want bad_request", err)
	}
	count, _ := st.EvidenceCount(opp.ID)
	if count != 0 { t.Fatalf("foreign evidence persisted, count=%d", count) }
}
