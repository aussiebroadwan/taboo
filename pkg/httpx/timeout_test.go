package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeout_NormalRequest(t *testing.T) {
	handler := Timeout(100 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "OK" {
		t.Errorf("expected body %q, got %q", "OK", rec.Body.String())
	}
}

func TestTimeout_SlowRequest(t *testing.T) {
	handler := Timeout(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, rec.Code)
	}
}

func TestTimeoutWithSkip(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow"))
	})

	handler := TimeoutWithSkip(50*time.Millisecond, "/sse", "/events")(slowHandler)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "skipped path /sse",
			path:       "/sse",
			wantStatus: http.StatusOK,
		},
		{
			name:       "skipped path /events",
			path:       "/events",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-skipped path times out",
			path:       "/api",
			wantStatus: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestTimeout_ContextCancellation(t *testing.T) {
	contextCanceled := make(chan bool, 1)

	handler := Timeout(50 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			contextCanceled <- true
		case <-time.After(200 * time.Millisecond):
			contextCanceled <- false
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	select {
	case canceled := <-contextCanceled:
		if !canceled {
			t.Error("expected context to be canceled")
		}
	case <-time.After(time.Second):
		t.Error("test timed out")
	}
}
