package service

import (
	"testing"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

func TestBug04VerifySegmentClassifiesWithExistingBaseline(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	sp, _ := svc.CreateSpeaker("sp", "d", "f", 1950)
	batch, _ := svc.CreateBatch("b", "tone")
	seed, _ := svc.ImportSegment(batch.ID, segInputFor("seed", sp.ID, "seed", 200, 200))
	target, _ := svc.ImportSegment(batch.ID, segInputFor("ma-rising", sp.ID, "target", 150, 220))
	if err := svc.VerifySegment(seed.ID); err != nil { t.Fatal(err) }
	if err := svc.RecomputeBaseline(sp.ID); err != nil { t.Fatal(err) }
	if got, _ := svc.GetSegment(target.ID); got.ToneType != model.ToneUnknown { t.Fatalf("pending target changed early: %q", got.ToneType) }
	if err := svc.VerifySegment(target.ID); err != nil { t.Fatal(err) }
	got, err := svc.GetSegment(target.ID)
	if err != nil { t.Fatal(err) }
	if got.ToneType == model.ToneUnknown { t.Fatalf("verified target tone type remained unknown") }
}
