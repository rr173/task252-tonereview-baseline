package service

import (
	"errors"
	"testing"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

func TestBug09CannotConfirmWithoutEvidence(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	batch, _ := svc.CreateBatch("b", "tone")
	opp, _ := svc.CreateOpposition(batch.ID, "ma-high", "ma", "ma-rising")
	if err := svc.Adjudicate(opp.ID, model.OppConfirmed, "manual"); !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("confirmation without evidence error = %v, want invalid_state", err)
	}
	got, _ := svc.GetOpposition(opp.ID)
	if got.Status != model.OppCandidate { t.Fatalf("empty opposition advanced to %q", got.Status) }
}
