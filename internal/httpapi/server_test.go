package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"task252-tonereview/internal/service"
	"task252-tonereview/internal/store"
)

func newTestAPI(t *testing.T) *API {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return New(service.New(st))
}

func TestHealthAndSelfCheck(t *testing.T) {
	api := newTestAPI(t)
	srv := httptest.NewServer(api.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health status %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp, err = http.Get(srv.URL + "/api/selfcheck")
	if err != nil {
		t.Fatalf("selfcheck: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("selfcheck status %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["ok"] != true {
		t.Fatalf("selfcheck ok != true: %+v", body)
	}
	resp.Body.Close()
}

func TestCreateBatchViaAPI(t *testing.T) {
	api := newTestAPI(t)
	srv := httptest.NewServer(api.Handler())
	defer srv.Close()

	payload, _ := json.Marshal(map[string]string{"code": "t-batch", "title": "API 测试"})
	resp, err := http.Post(srv.URL+"/api/batches", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status %d", resp.StatusCode)
	}
	var b map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if b["id"] == nil || b["id"] == "" {
		t.Fatalf("missing id: %+v", b)
	}
	resp.Body.Close()

	// 列表应包含刚创建的批次。
	resp, err = http.Get(srv.URL + "/api/batches")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var list map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if int(list["count"].(float64)) < 1 {
		t.Fatalf("expected at least 1 batch, got %+v", list)
	}
	resp.Body.Close()
}

func TestCreateBatchBadRequest(t *testing.T) {
	api := newTestAPI(t)
	srv := httptest.NewServer(api.Handler())
	defer srv.Close()

	// 含未知字段的 JSON 体应被 DisallowUnknownFields 拒绝（400）。
	resp, err := http.Post(srv.URL+"/api/batches", "application/json", bytes.NewReader([]byte(`{"code":"x","bogus":123}`)))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown field, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
