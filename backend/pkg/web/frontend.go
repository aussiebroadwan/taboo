package web

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/lcox74/tabo/backend/assets"
)

// serveEmbeddedFile opens a file from the embedded filesystem,
// determines its content type, and writes its contents to the response.
func serveEmbeddedFile(w http.ResponseWriter, embedPath string) error {
	// Open file from the embedded filesystem.
	f, err := assets.FrontendFS.Open(embedPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the file content.
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	// Set the proper Content-Type header based on the file extension.
	contentType := mime.TypeByExtension(path.Ext(embedPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)

	// Write the file data to the response.
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	return err
}

// RegisterFrontend registers the HTTP handlers for serving the frontend assets.
func RegisterFrontend() {
	// Handler for serving the Discord client ID.
	http.HandleFunc("/client-id", func(w http.ResponseWriter, r *http.Request) {
		clientId := os.Getenv("DISCORD_CLIENT_ID")
		if clientId == "" {
			// Return an empty JSON object if the client ID is not set.
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
			return
		}

		// Return the client ID as a JSON payload.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fmt.Appendf(nil, `{ "clientId": "%s" }`, clientId))
	})

	// Handler for game routes (e.g. /game/123). This always serves index.html
	// to allow client-side routing in the SPA.
	http.HandleFunc("/game/{gameid}", func(w http.ResponseWriter, r *http.Request) {
		if err := serveEmbeddedFile(w, "dist/index.html"); err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
		}
	})

	// Catch-all handler for serving static assets.
	// If a file is not found, it falls back to serving index.html
	// to support client-side routing (e.g. for a Vue or React SPA).
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Normalize the request path. Default to index.html if empty.
		requestPath := strings.TrimPrefix(r.URL.Path, "/")
		if requestPath == "" {
			requestPath = "index.html"
		}

		// Build the path for the embedded asset.
		embedPath := path.Join("dist", requestPath)

		// Try to serve the requested file.
		if err := serveEmbeddedFile(w, embedPath); err != nil {
			// Fallback: if the file is not found, serve index.html for SPA routing.
			if err := serveEmbeddedFile(w, "dist/index.html"); err != nil {
				http.Error(w, "index.html not found", http.StatusInternalServerError)
			}
		}
	})
}
