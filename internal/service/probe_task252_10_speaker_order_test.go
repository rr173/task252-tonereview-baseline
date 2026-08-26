package service

import (
	"testing"

	"task252-tonereview/internal/store"
)

func TestBug10RecomputePreservesSpeakerOrder(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	a, _ := svc.CreateSpeaker("a", "d", "f", 1950)
	b, _ := svc.CreateSpeaker("b", "d", "m", 1950)
	batch, _ := svc.CreateBatch("batch", "tone")
	segA, _ := svc.ImportSegment(batch.ID, segInputFor("ma-a", a.ID, "fp-a", 200, 200))
	segB, _ := svc.ImportSegment(batch.ID, segInputFor("ma-b", b.ID, "fp-b", 180, 180))
	if err := svc.VerifySegment(segA.ID); err != nil { t.Fatal(err) }
	if err := svc.VerifySegment(segB.ID); err != nil { t.Fatal(err) }
	if _, err := st.DB().Exec("UPDATE speakers SET created_at=1 WHERE id=?", a.ID); err != nil { t.Fatal(err) }
	if _, err := st.DB().Exec("UPDATE speakers SET created_at=2 WHERE id=?", b.ID); err != nil { t.Fatal(err) }
	if err := svc.RecomputeBaseline(a.ID); err != nil { t.Fatal(err) }
	speakers, err := svc.ListSpeakers()
	if err != nil { t.Fatal(err) }
	if len(speakers) != 2 || speakers[0].ID != a.ID || speakers[1].ID != b.ID { t.Fatalf("speaker order after recompute = %+v", speakers) }
}
