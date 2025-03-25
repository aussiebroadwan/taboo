package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"embed"
)

//go:embed all:dist/*
var FrontendFS embed.FS

func RegisterFrontend() {
	http.HandleFunc("/client-id", func(w http.ResponseWriter, r *http.Request) {
		clientId := os.Getenv("DISCORD_CLIENT_ID")

		if clientId == "" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{ "clientId": "%s" }`, clientId)))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Normalize the request path
		requestPath := strings.TrimPrefix(r.URL.Path, "/")
		if requestPath == "" {
			requestPath = "index.html" // Default to index.html
		}

		// Construct the embedded file path
		embedPath := path.Join("dist", requestPath)

		// Open the file from the embedded filesystem
		f, err := FrontendFS.Open(embedPath)
		if err != nil {
			// If the file doesn't exist, serve index.html (for Vue SPA handling)
			embedPath = "dist/index.html"
			f, err = FrontendFS.Open(embedPath)
			if err != nil {
				http.Error(w, "index.html not found", http.StatusInternalServerError)
				return
			}
		}
		defer f.Close()

		// Read file contents
		data, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		// Set Content-Type based on file extension
		contentType := mime.TypeByExtension(path.Ext(embedPath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)

		// Serve the file
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
}
