package httpx

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzip_CompressesResponse(t *testing.T) {
	body := strings.Repeat("Hello, World! ", 100)
	handler := Gzip()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Error("expected Content-Encoding: gzip")
	}
	if !strings.Contains(rec.Header().Get("Vary"), "Accept-Encoding") {
		t.Error("expected Vary header to contain Accept-Encoding")
	}

	// Verify the body is actually gzipped
	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if string(decompressed) != body {
		t.Errorf("decompressed body mismatch")
	}
}

func TestGzip_SkipsWhenNotAccepted(t *testing.T) {
	body := "Hello, World!"
	handler := Gzip()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Accept-Encoding header
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not compress when gzip not accepted")
	}
	if rec.Body.String() != body {
		t.Errorf("expected body %q, got %q", body, rec.Body.String())
	}
}

func TestGzip_SkipsConfiguredPaths(t *testing.T) {
	body := strings.Repeat("Hello, World! ", 100)
	handler := Gzip("/sse", "/events")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))

	tests := []struct {
		name           string
		path           string
		wantCompressed bool
	}{
		{
			name:           "skipped path /sse",
			path:           "/sse",
			wantCompressed: false,
		},
		{
			name:           "skipped path /events",
			path:           "/events",
			wantCompressed: false,
		},
		{
			name:           "regular path",
			path:           "/api",
			wantCompressed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			compressed := rec.Header().Get("Content-Encoding") == "gzip"
			if compressed != tt.wantCompressed {
				t.Errorf("compressed = %v, want %v", compressed, tt.wantCompressed)
			}
		})
	}
}

func TestGzip_Flush(t *testing.T) {
	handler := Gzip()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("part1"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		w.Write([]byte("part2"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify the response is valid gzip
	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if string(decompressed) != "part1part2" {
		t.Errorf("expected body %q, got %q", "part1part2", string(decompressed))
	}
}
