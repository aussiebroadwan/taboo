package http

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/aussiebroadwan/taboo/internal/frontend"
	"github.com/aussiebroadwan/taboo/pkg/slogx"
)

// staticHandler returns an http.Handler that serves static files from the
// embedded frontend filesystem with SPA fallback support.
func (s *Server) staticHandler() http.Handler {
	frontendFS, err := frontend.GetFS()
	if err != nil {
		s.logger.Error("Failed to get frontend filesystem",
			slogx.Error(err),
			slog.String("component", "frontend"),
		)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Frontend not available", http.StatusInternalServerError)
		})
	}

	return &spaHandler{
		fs: frontendFS,
	}
}

// spaHandler serves static files with SPA fallback.
// Unknown paths that don't match a file return index.html.
type spaHandler struct {
	fs fs.FS
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path
	urlPath := path.Clean(r.URL.Path)
	if urlPath == "/" {
		urlPath = "/index.html"
	}

	// Remove leading slash for fs operations
	filePath := strings.TrimPrefix(urlPath, "/")

	// Try to open the file
	file, err := h.fs.Open(filePath)
	if err != nil {
		// File not found - serve index.html for SPA routing
		h.serveIndex(w, r)
		return
	}
	defer file.Close()

	// Check if it's a directory
	stat, err := file.Stat()
	if err != nil {
		h.serveIndex(w, r)
		return
	}

	if stat.IsDir() {
		// Try to serve index.html from the directory
		indexPath := path.Join(filePath, "index.html")
		indexFile, err := h.fs.Open(indexPath)
		if err != nil {
			h.serveIndex(w, r)
			return
		}
		defer indexFile.Close()
		file = indexFile
		filePath = indexPath
		stat, _ = indexFile.Stat()
	}

	// Set cache headers based on file type
	h.setCacheHeaders(w, filePath)

	// Set content type
	contentType := getContentType(filePath)
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Serve the file
	if seeker, ok := file.(io.ReadSeeker); ok {
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), seeker)
	} else {
		w.WriteHeader(http.StatusOK)
		if _, err := io.Copy(w, file); err != nil {
			slogx.FromContext(r.Context()).Debug("Failed to copy file to response",
				slogx.Error(err),
				slog.String("file_path", filePath),
			)
		}
	}
}

// serveIndex serves the index.html file for SPA routing.
func (h *spaHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	file, err := h.fs.Open("index.html")
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// index.html should never be cached
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if seeker, ok := file.(io.ReadSeeker); ok {
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), seeker)
	} else {
		w.WriteHeader(http.StatusOK)
		if _, err := io.Copy(w, file); err != nil {
			slogx.FromContext(r.Context()).Debug("Failed to copy index.html to response", slogx.Error(err))
		}
	}
}

// setCacheHeaders sets appropriate cache headers based on file type.
func (h *spaHandler) setCacheHeaders(w http.ResponseWriter, filePath string) {
	// index.html should never be cached
	if strings.HasSuffix(filePath, "index.html") {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		return
	}

	// Hashed assets (Vite generates files like index-abc123.js) can be cached forever
	if isHashedAsset(filePath) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}

	// Other static assets get moderate caching
	w.Header().Set("Cache-Control", "public, max-age=86400")
}

// isHashedAsset checks if a file appears to have a content hash in its name.
// Vite generates files like: index-BDqhSP_c.js, style-abc123.css
func isHashedAsset(filePath string) bool {
	ext := path.Ext(filePath)
	if ext != ".js" && ext != ".css" {
		return false
	}

	// Look for hash pattern: name-HASH.ext
	base := strings.TrimSuffix(path.Base(filePath), ext)
	if idx := strings.LastIndex(base, "-"); idx > 0 {
		hash := base[idx+1:]
		// Vite hashes are typically 8 characters
		return len(hash) >= 6 && len(hash) <= 12
	}
	return false
}

// getContentType returns the content type for a file based on its extension.
func getContentType(filePath string) string {
	ext := path.Ext(filePath)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	default:
		return ""
	}
}
