package httpapi

import (
	"net/http"

	"task252-tonereview/internal/capture"
	"task252-tonereview/internal/store"
)

func (a *API) handleImportSegment(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	var req capture.SegmentInput
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	seg, err := a.svc.ImportSegment(batchID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, seg, http.StatusCreated)
}

func (a *API) handleListSegments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.SegmentFilter{
		BatchID:     q.Get("batch_id"),
		SpeakerID:   q.Get("speaker_id"),
		Status:      q.Get("status"),
		PhoneticSeg: q.Get("phonetic_seg"),
	}
	segs, err := a.svc.ListSegments(f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"segments": segs, "count": len(segs)}, http.StatusOK)
}

func (a *API) handleGetSegment(w http.ResponseWriter, r *http.Request) {
	seg, err := a.svc.GetSegment(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, seg, http.StatusOK)
}

func (a *API) handleVerifySegment(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.VerifySegment(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "usable"}, http.StatusOK)
}

func (a *API) handleMarkNoise(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.MarkNoise(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "noise"}, http.StatusOK)
}

func (a *API) handleExcludeSegment(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.ExcludeSegment(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "excluded"}, http.StatusOK)
}

func (a *API) handleRestoreSegment(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.RestoreSegment(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "usable"}, http.StatusOK)
}

type f0AddReq struct {
	F0 []capture.F0Point `json:"f0"`
}

func (a *API) handleAddF0(w http.ResponseWriter, r *http.Request) {
	var req f0AddReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := a.svc.AddF0(r.PathValue("id"), req.F0); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"added": len(req.F0)}, http.StatusOK)
}

func (a *API) handleListF0(w http.ResponseWriter, r *http.Request) {
	samples, err := a.svc.ListF0(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"f0": samples, "count": len(samples)}, http.StatusOK)
}
