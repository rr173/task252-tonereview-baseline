package httpapi

import "net/http"

type oppositionCreateReq struct {
	BatchID     string `json:"batch_id"`
	LexicalA    string `json:"lexical_a"`
	PhoneticSeg string `json:"phonetic_seg"`
	LexicalB    string `json:"lexical_b"`
}

type evidenceReq struct {
	SegmentID string `json:"segment_id"`
	Side      string `json:"side"`
}

type adjudicateReq struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type compareReq struct {
	SegmentIDsA []string `json:"segment_ids_a"`
	SegmentIDsB []string `json:"segment_ids_b"`
}

func (a *API) handleCreateOpposition(w http.ResponseWriter, r *http.Request) {
	var req oppositionCreateReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	o, err := a.svc.CreateOpposition(req.BatchID, req.LexicalA, req.PhoneticSeg, req.LexicalB)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, o, http.StatusCreated)
}

func (a *API) handleListOppositions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	os, err := a.svc.ListOppositions(q.Get("batch_id"), q.Get("status"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"oppositions": os, "count": len(os)}, http.StatusOK)
}

func (a *API) handleGetOpposition(w http.ResponseWriter, r *http.Request) {
	o, err := a.svc.GetOpposition(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, o, http.StatusOK)
}

func (a *API) handleAddEvidence(w http.ResponseWriter, r *http.Request) {
	var req evidenceReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := a.svc.AddEvidence(r.PathValue("id"), req.SegmentID, req.Side); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"evidence": "added", "side": req.Side}, http.StatusOK)
}

func (a *API) handleRecomputeCluster(w http.ResponseWriter, r *http.Request) {
	stats, err := a.svc.RecomputeCluster(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, stats, http.StatusOK)
}

func (a *API) handleGetCluster(w http.ResponseWriter, r *http.Request) {
	stats, err := a.svc.GetCluster(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, stats, http.StatusOK)
}

func (a *API) handleAdjudicate(w http.ResponseWriter, r *http.Request) {
	var req adjudicateReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := a.svc.Adjudicate(r.PathValue("id"), req.Decision, req.Reason); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, map[string]any{"status": req.Decision}, http.StatusOK)
}

func (a *API) handleCompare(w http.ResponseWriter, r *http.Request) {
	var req compareReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	stats, err := a.svc.CompareContours(req.SegmentIDsA, req.SegmentIDsB)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, stats, http.StatusOK)
}
