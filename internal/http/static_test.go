package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSpaHandler_ServeIndex(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>Hello</body></html>"),
		},
	}

	handler := &spaHandler{fs: fs}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != "<!DOCTYPE html><html><body>Hello</body></html>" {
		t.Errorf("unexpected body: %s", body)
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type text/html; charset=utf-8, got %s", ct)
	}

	if cc := w.Header().Get("Cache-Control"); cc != "no-cache, no-store, must-revalidate" {
		t.Errorf("expected Cache-Control no-cache, got %s", cc)
	}
}

func TestSpaHandler_ServeStaticFile(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html></html>"),
		},
		"style.css": &fstest.MapFile{
			Data: []byte("body { color: red; }"),
		},
	}

	handler := &spaHandler{fs: fs}

	req := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != "body { color: red; }" {
		t.Errorf("unexpected body: %s", body)
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Errorf("expected Content-Type text/css; charset=utf-8, got %s", ct)
	}
}

func TestSpaHandler_SPAFallback(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>SPA</body></html>"),
		},
	}

	handler := &spaHandler{fs: fs}

	// Test SPA routes that should fall back to index.html
	routes := []string{
		"/game/123",
		"/some/deep/path",
		"/nonexistent",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			body, _ := io.ReadAll(w.Body)
			if string(body) != "<!DOCTYPE html><html><body>SPA</body></html>" {
				t.Errorf("expected SPA fallback, got: %s", body)
			}
		})
	}
}

func TestSpaHandler_HashedAssetCaching(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html>"),
		},
		"assets/index-abc12345.js": &fstest.MapFile{
			Data: []byte("console.log('hello');"),
		},
		"assets/style-xyz98765.css": &fstest.MapFile{
			Data: []byte("body {}"),
		},
	}

	handler := &spaHandler{fs: fs}

	tests := []struct {
		path        string
		expectedCC  string
		description string
	}{
		{
			path:        "/assets/index-abc12345.js",
			expectedCC:  "public, max-age=31536000, immutable",
			description: "hashed JS file",
		},
		{
			path:        "/assets/style-xyz98765.css",
			expectedCC:  "public, max-age=31536000, immutable",
			description: "hashed CSS file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			if cc := w.Header().Get("Cache-Control"); cc != tc.expectedCC {
				t.Errorf("expected Cache-Control %q, got %q", tc.expectedCC, cc)
			}
		})
	}
}

func TestSpaHandler_RegularAssetCaching(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html>"),
		},
		"logo.svg": &fstest.MapFile{
			Data: []byte("<svg></svg>"),
		},
		"favicon.ico": &fstest.MapFile{
			Data: []byte{0, 0, 0},
		},
	}

	handler := &spaHandler{fs: fs}

	tests := []struct {
		path string
	}{
		{"/logo.svg"},
		{"/favicon.ico"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			if cc := w.Header().Get("Cache-Control"); cc != "public, max-age=86400" {
				t.Errorf("expected Cache-Control public, max-age=86400, got %q", cc)
			}
		})
	}
}

func TestSpaHandler_MissingIndex(t *testing.T) {
	// Empty filesystem - no index.html
	fs := fstest.MapFS{}

	handler := &spaHandler{fs: fs}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestIsHashedAsset(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"index-abc12345.js", true},
		{"style-xyz98765.css", true},
		{"index-BDqhSP_c.js", true},
		{"main.js", false},
		{"style.css", false},
		{"image.png", false},
		{"index.html", false},
		{"-abc123.js", false}, // no name before hash
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := isHashedAsset(tc.path)
			if result != tc.expected {
				t.Errorf("isHashedAsset(%q) = %v, expected %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"index.html", "text/html; charset=utf-8"},
		{"style.css", "text/css; charset=utf-8"},
		{"main.js", "application/javascript; charset=utf-8"},
		{"data.json", "application/json; charset=utf-8"},
		{"logo.svg", "image/svg+xml"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"animation.gif", "image/gif"},
		{"favicon.ico", "image/x-icon"},
		{"font.woff", "font/woff"},
		{"font.woff2", "font/woff2"},
		{"font.ttf", "font/ttf"},
		{"unknown.xyz", ""},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := getContentType(tc.path)
			if result != tc.expected {
				t.Errorf("getContentType(%q) = %q, expected %q", tc.path, result, tc.expected)
			}
		})
	}
}
