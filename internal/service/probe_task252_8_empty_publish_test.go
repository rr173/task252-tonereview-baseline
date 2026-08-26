package service

import (
	"errors"
	"testing"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

func TestBug08EmptyBatchCannotPublish(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	batch, _ := svc.CreateBatch("b", "empty")
	if err := svc.StartReview(batch.ID); err != nil { t.Fatal(err) }
	if err := svc.PublishBatch(batch.ID); !errors.Is(err, model.ErrInvalidState) { t.Fatalf("empty publish error = %v, want invalid_state", err) }
	got, _ := svc.GetBatch(batch.ID)
	if got.Status != model.BatchReviewing { t.Fatalf("empty batch advanced to %q", got.Status) }
}
