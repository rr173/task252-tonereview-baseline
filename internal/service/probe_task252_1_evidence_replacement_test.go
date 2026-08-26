package service

import (
	"testing"

	"task252-tonereview/internal/store"
)

func TestBug01EvidenceReplacementKeepsOtherEvidence(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	a, _ := svc.CreateSpeaker("a", "d", "f", 1950)
	b, _ := svc.CreateSpeaker("b", "d", "m", 1950)
	batch, _ := svc.CreateBatch("b", "tone")
	a1, _ := svc.ImportSegment(batch.ID, segInputFor("ma-high", a.ID, "a1", 200, 200))
	a2, _ := svc.ImportSegment(batch.ID, segInputFor("ma-high", a.ID, "a2", 202, 198))
	b1, _ := svc.ImportSegment(batch.ID, segInputFor("ma-rising", b.ID, "b1", 150, 220))
	for _, id := range []string{a1.ID, a2.ID, b1.ID} { if err := svc.VerifySegment(id); err != nil { t.Fatal(err) } }
	if err := svc.RecomputeBaseline(a.ID); err != nil { t.Fatal(err) }
	if err := svc.RecomputeBaseline(b.ID); err != nil { t.Fatal(err) }
	opp, _ := svc.CreateOpposition(batch.ID, "ma-high", "ma", "ma-rising")
	for _, e := range []struct{ id, side string }{{a1.ID, "a"}, {a2.ID, "a"}, {b1.ID, "b"}} {
		if err := svc.AddEvidence(opp.ID, e.id, e.side); err != nil { t.Fatal(err) }
	}
	if err := svc.AddEvidence(opp.ID, a1.ID, "a"); err != nil { t.Fatal(err) }
	evs, err := st.ListEvidence(opp.ID)
	if err != nil { t.Fatal(err) }
	if len(evs) != 3 { t.Fatalf("evidence count after idempotent replacement = %d, want 3", len(evs)) }
}
