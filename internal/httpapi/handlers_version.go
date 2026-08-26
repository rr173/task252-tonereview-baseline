package httpapi

import "net/http"

type versionCreateReq struct {
	BatchID string `json:"batch_id"`
	Note    string `json:"note"`
}

func (a *API) handleCreateVersion(w http.ResponseWriter, r *http.Request) {
	var req versionCreateReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	v, err := a.svc.CreateVersion(req.BatchID, req.Note)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, v, http.StatusCreated)
}

func (a *API) handleShareVersion(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.ShareVersion(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "shared"}, http.StatusOK)
}

func (a *API) handleFreezeVersion(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.FreezeVersion(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": "frozen"}, http.StatusOK)
}

func (a *API) handleListVersions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	vs, err := a.svc.ListVersions(q.Get("batch_id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"versions": vs, "count": len(vs)}, http.StatusOK)
}

func (a *API) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	v, err := a.svc.GetVersion(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, v, http.StatusOK)
}
