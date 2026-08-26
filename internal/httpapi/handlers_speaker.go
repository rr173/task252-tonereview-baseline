package httpapi

import "net/http"

type speakerCreateReq struct {
	Code       string `json:"code"`
	Dialect    string `json:"dialect"`
	Gender     string `json:"gender"`
	BirthYear  int    `json:"birth_year"`
}

func (a *API) handleCreateSpeaker(w http.ResponseWriter, r *http.Request) {
	var req speakerCreateReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	sp, err := a.svc.CreateSpeaker(req.Code, req.Dialect, req.Gender, req.BirthYear)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, sp, http.StatusCreated)
}

func (a *API) handleListSpeakers(w http.ResponseWriter, r *http.Request) {
	sps, err := a.svc.ListSpeakers()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"speakers": sps, "count": len(sps)}, http.StatusOK)
}

func (a *API) handleGetSpeaker(w http.ResponseWriter, r *http.Request) {
	sp, err := a.svc.GetSpeaker(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, sp, http.StatusOK)
}

func (a *API) handleRecomputeBaseline(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.RecomputeBaseline(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	sp, err := a.svc.GetSpeaker(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, sp, http.StatusOK)
}
