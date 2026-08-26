package httpapi

import (
	"net/http"
)

type batchCreateReq struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

func (a *API) handleCreateBatch(w http.ResponseWriter, r *http.Request) {
	var req batchCreateReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	b, err := a.svc.CreateBatch(req.Code, req.Title)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, b, http.StatusCreated)
}

func (a *API) handleListBatches(w http.ResponseWriter, r *http.Request) {
	bs, err := a.svc.ListBatches()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"batches": bs, "count": len(bs)}, http.StatusOK)
}

func (a *API) handleGetBatch(w http.ResponseWriter, r *http.Request) {
	b, err := a.svc.GetBatch(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, b, http.StatusOK)
}

func (a *API) handleBatchSummary(w http.ResponseWriter, r *http.Request) {
	sum, err := a.svc.BatchSummary(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, sum, http.StatusOK)
}

func (a *API) handleStartReview(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.StartReview(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "reviewing"}, http.StatusOK)
}

func (a *API) handlePublishBatch(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.PublishBatch(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "published"}, http.StatusOK)
}

func (a *API) handleSealBatch(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.SealBatch(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "sealed"}, http.StatusOK)
}
