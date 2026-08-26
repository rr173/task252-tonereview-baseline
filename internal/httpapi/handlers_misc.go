package httpapi

import "net/http"

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"status": "ok"}, http.StatusOK)
}

func (a *API) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := a.svc.Stats()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, stats, http.StatusOK)
}

func (a *API) handleSelfCheck(w http.ResponseWriter, r *http.Request) {
	res, err := a.svc.SelfCheck()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, res, http.StatusOK)
}
