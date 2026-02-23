package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleLivez(t *testing.T) {
	ts := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	w := httptest.NewRecorder()

	ts.handleLivez(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}

func TestHandleReadyz_AllHealthy(t *testing.T) {
	ts := newTestServer(t)

	// Start the engine to make it "running"
	ts.engine.SetRunning(true)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	ts.handleReadyz(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}

	if resp.Checks["database"] != "ok" {
		t.Errorf("expected database ok, got %s", resp.Checks["database"])
	}

	if resp.Checks["engine"] != "ok" {
		t.Errorf("expected engine ok, got %s", resp.Checks["engine"])
	}
}

func TestHandleReadyz_EngineNotRunning(t *testing.T) {
	ts := newTestServer(t)

	// Engine is not running by default

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	ts.handleReadyz(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var resp struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status degraded, got %s", resp.Status)
	}

	if resp.Checks["engine"] != "not running" {
		t.Errorf("expected engine not running, got %s", resp.Checks["engine"])
	}
}

func TestHandleReadyz_DatabaseError(t *testing.T) {
	ts := newTestServer(t)
	ts.engine.SetRunning(true)
	ts.mockStore.pingErr = errMockDB

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	ts.handleReadyz(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var resp struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status degraded, got %s", resp.Status)
	}

	if resp.Checks["database"] == "ok" {
		t.Error("expected database check to fail")
	}
}
