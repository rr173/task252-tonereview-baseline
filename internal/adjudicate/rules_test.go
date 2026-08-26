package adjudicate

import (
	"testing"

	"task252-tonereview/internal/model"
)

func TestSuggestStatus(t *testing.T) {
	if got := SuggestStatus(2.0); got != model.OppConfirmed {
		t.Fatalf("score 2.0 -> %v, want confirmed", got)
	}
	if got := SuggestStatus(1.2); got != model.OppCandidate {
		t.Fatalf("score 1.2 -> %v, want candidate", got)
	}
	if got := SuggestStatus(0.5); got != model.OppInsufficient {
		t.Fatalf("score 0.5 -> %v, want insufficient", got)
	}
}

func TestCanTransition(t *testing.T) {
	// 非终态可流转到任意合法状态。
	if !CanTransition(model.OppCandidate, model.OppConfirmed) {
		t.Fatalf("candidate -> confirmed should be allowed")
	}
	if !CanTransition(model.OppInsufficient, model.OppCandidate) {
		t.Fatalf("insufficient -> candidate should be allowed")
	}
	// 终态不可再改。
	if CanTransition(model.OppConfirmed, model.OppRejected) {
		t.Fatalf("confirmed -> rejected should be forbidden")
	}
	if CanTransition(model.OppRejected, model.OppCandidate) {
		t.Fatalf("rejected -> candidate should be forbidden")
	}
}

func TestCanVersionTransition(t *testing.T) {
	if !CanVersionTransition(model.VerDraft, model.VerShared) {
		t.Fatalf("draft -> shared should be allowed")
	}
	if !CanVersionTransition(model.VerShared, model.VerFrozen) {
		t.Fatalf("shared -> frozen should be allowed")
	}
	if CanVersionTransition(model.VerFrozen, model.VerShared) {
		t.Fatalf("frozen -> shared should be forbidden")
	}
	if CanVersionTransition(model.VerSuperseded, model.VerDraft) {
		t.Fatalf("superseded -> draft should be forbidden")
	}
}

func TestCanBatchTransition(t *testing.T) {
	if !CanBatchTransition(model.BatchCollecting, model.BatchReviewing) {
		t.Fatalf("collecting -> reviewing should be allowed")
	}
	if !CanBatchTransition(model.BatchPublished, model.BatchSealed) {
		t.Fatalf("published -> sealed should be allowed")
	}
	if CanBatchTransition(model.BatchSealed, model.BatchPublished) {
		t.Fatalf("sealed (terminal) -> published should be forbidden")
	}
}
