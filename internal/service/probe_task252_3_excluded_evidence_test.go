package service

import (
	"errors"
	"testing"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

func TestBug03EvidenceRejectsExcludedSegment(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := New(st)
	sp, _ := svc.CreateSpeaker("sp", "d", "f", 1950)
	batch, _ := svc.CreateBatch("b", "tone")
	baseline, _ := svc.ImportSegment(batch.ID, segInputFor("ma-high", sp.ID, "baseline", 200, 200))
	if err := svc.VerifySegment(baseline.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.RecomputeBaseline(sp.ID); err != nil {
		t.Fatal(err)
	}
	opp, _ := svc.CreateOpposition(batch.ID, "ma-high", "ma", "ma-rising")
	target, _ := svc.ImportSegment(batch.ID, segInputFor("ma-high", sp.ID, "excluded", 205, 205))
	if err := svc.ExcludeSegment(target.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddEvidence(opp.ID, target.ID, "a"); !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("excluded evidence error = %v, want invalid_state", err)
	}
}
