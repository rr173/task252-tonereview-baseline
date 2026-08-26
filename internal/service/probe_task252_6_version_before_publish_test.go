package service

import (
	"errors"
	"testing"

	"task252-tonereview/internal/model"
	"task252-tonereview/internal/store"
)

func TestBug06CannotCreateVersionBeforePublish(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st)
	batch, _ := svc.CreateBatch("b", "tone")
	if _, err := svc.CreateVersion(batch.ID, "premature"); !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("premature version error = %v, want invalid_state", err)
	}
	versions, _ := svc.ListVersions(batch.ID)
	if len(versions) != 0 { t.Fatalf("premature version persisted: %+v", versions) }
}
