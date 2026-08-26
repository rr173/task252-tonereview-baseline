package service

import (
	"errors"
	"testing"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

func TestBug05SealedBatchRejectsBaselineRecompute(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	sp, _ := svc.CreateSpeaker("sp", "d", "f", 1950)
	batch, _ := svc.CreateBatch("b", "tone")
	seg, _ := svc.ImportSegment(batch.ID, segInputFor("ma-high", sp.ID, "fp", 200, 200))
	if err := svc.VerifySegment(seg.ID); err != nil { t.Fatal(err) }
	if err := svc.RecomputeBaseline(sp.ID); err != nil { t.Fatal(err) }
	if err := svc.StartReview(batch.ID); err != nil { t.Fatal(err) }
	if err := svc.PublishBatch(batch.ID); err != nil { t.Fatal(err) }
	if err := svc.SealBatch(batch.ID); err != nil { t.Fatal(err) }
	if err := svc.RecomputeBaseline(sp.ID); !errors.Is(err, model.ErrSealed) { t.Fatalf("sealed recompute error = %v, want sealed", err) }
}
